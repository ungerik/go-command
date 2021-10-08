package gen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ungerik/go-astvisit"
)

func Rewrite(path string, printOnly io.Writer) (err error) {
	recursive := strings.HasSuffix(path, "...")
	if recursive {
		path = strings.TrimSuffix(path, "...")
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return RewriteFile(path, printOnly)
	}

	fset := token.NewFileSet()
	pkg, err := astvisit.ParsePackage(fset, path, filterOutTests)
	if err != nil {
		if errors.Is(err, astvisit.ErrPackageNotFound) {
			return nil
		}
		return err
	}
	for fileName, file := range pkg.Files {
		err = RewriteAstFile(fset, pkg, file, fileName, printOnly)
		if err != nil {
			return err
		}
	}
	if !recursive {
		return nil
	}

	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		return Rewrite(filepath.Join(path, d.Name())+"...", printOnly)
	})
}

func RewriteFile(filePath string, printOnly io.Writer) (err error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("file path is a directory: %s", filePath)
	}
	fileDir, fileName := filepath.Split(filePath)
	fset := token.NewFileSet()
	pkg, err := astvisit.ParsePackage(fset, fileDir, filterOutTests)
	if err != nil {
		return err
	}
	return RewriteAstFile(fset, pkg, pkg.Files[fileName], filePath, printOnly)
}

func RewriteAstFile(fset *token.FileSet, filePkg *ast.Package, file *ast.File, filePath string, printOnly io.Writer) (err error) {
	// ast.Print(fset, file)

	funcImpls := findFuncImpls(fset, file)
	if len(funcImpls) == 0 {
		return nil // nothing to rewrite
	}

	fileDir := filepath.Dir(filePath)

	// Gather imported packages of file
	// and parse packages for function declarations
	// that could be referenced by command.Function implementations
	type importedPkg struct {
		Location *astvisit.PackageLocation
		Funcs    map[string]funcInfo
	}
	functions := make(map[string]importedPkg)
	for _, imp := range file.Imports {
		importName, pkgLocation, err := astvisit.ImportSpecInfo(fileDir, imp)
		if err != nil {
			return err
		}
		if pkgLocation.Std {
			continue
		}
		impPkg, err := astvisit.ParsePackage(fset, pkgLocation.SourcePath, filterOutTests)
		if err != nil {
			return err
		}
		exportedFuncs := make(map[string]funcInfo)
		for _, f := range impPkg.Files {
			for _, decl := range f.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if ok && funcDecl.Name.IsExported() {
					exportedFuncs[funcDecl.Name.Name] = funcInfo{
						Decl: funcDecl,
						File: f,
					}
				}
			}
		}
		functions[importName] = importedPkg{
			Location: pkgLocation,
			Funcs:    exportedFuncs,
		}
	}
	// Also parse all functions of the file's package
	// because they could als be referenced with an empty import name
	pkgFuncs := make(map[string]funcInfo)
	for _, f := range filePkg.Files {
		for _, decl := range f.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				pkgFuncs[funcDecl.Name.Name] = funcInfo{
					Decl: funcDecl,
					File: f,
				}
			}
		}
	}
	functions[""] = importedPkg{
		Location: &astvisit.PackageLocation{
			PkgName:    filePkg.Name,
			SourcePath: fileDir,
		},
		Funcs: pkgFuncs,
	}

	// Modify file.Decls
	for i := len(funcImpls) - 1; i >= 0; i-- {
		impl := funcImpls[i]

		// Remove previous declarations
		for j := len(impl.DeclIndices) - 1; j >= 0; j-- {
			declIndex := impl.DeclIndices[j]
			file.Decls = append(file.Decls[:declIndex], file.Decls[declIndex+1:]...)
		}
		// Remove previous comments
		for _, comment := range impl.Comments {
			for j := len(file.Comments) - 1; j >= 0; j-- {
				if file.Comments[j] == comment {
					file.Comments = append(file.Comments[:j], file.Comments[j+1:]...)
				}
			}
		}

		importName, funcName := impl.WrappedFuncPkgFuncName()
		referencedPkg, ok := functions[importName]
		if !ok {
			return fmt.Errorf("can't find package %s in imports of file %s", importName, filePath)
		}
		wrappedFunc, ok := referencedPkg.Funcs[funcName]
		if !ok {
			return fmt.Errorf("can't find function %s in package %s", funcName, importName)
		}

		var newSrc strings.Builder
		// fmt.Fprintf(&newSrc, "////////////////////////////////////////\n")
		// fmt.Fprintf(&newSrc, "// %s\n\n", impl.WrappedFunc)
		// fmt.Fprintf(&newSrc, "// XXX %s wraps %s as command.Function (generated code)\n", impl.VarName, impl.WrappedFunc)
		fmt.Fprintf(&newSrc, "var %[1]s %[1]sT\n\n", impl.VarName)
		err = WriteFunctionImpl(&newSrc, file, wrappedFunc.Decl, impl.VarName+"T", importName)
		if err != nil {
			return err
		}
		newDecls, newComments, err := astvisit.ParseDeclarations(fset, newSrc.String())
		if err != nil {
			return err
		}
		// Insert rewritten declarations at position of first old declaration
		insertIndex := impl.DeclIndices[0]
		file.Decls = append(file.Decls[:insertIndex], append(newDecls, file.Decls[:insertIndex]...)...)
		// file.Comments = append(file.Comments, newComments...)
		var _ = newDecls
		var _ = newComments
	}

	buf := bytes.NewBuffer(nil)
	const printerNormalizeNumbers = 1 << 30
	config := printer.Config{Mode: printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers, Tabwidth: 8}
	err = config.Fprint(buf, fset, file)
	// err = format.Node(buf, fset, file)
	if err != nil {
		return err
	}
	if printOnly != nil {
		_, err = printOnly.Write(buf.Bytes())
		return err
	}
	fmt.Println("Writing file", filePath)
	return ioutil.WriteFile(filePath, buf.Bytes(), 0660)
}

type funcImpl struct {
	VarName     string
	WrappedFunc string
	Type        string
	DeclIndices []int
	Comments    []*ast.CommentGroup
}

func (impl *funcImpl) WrappedFuncPkgFuncName() (pkgName, funcName string) {
	dot := strings.IndexByte(impl.WrappedFunc, '.')
	if dot == -1 {
		return "", impl.WrappedFunc
	}
	return impl.WrappedFunc[:dot], impl.WrappedFunc[dot+1:]
}

func findFuncImpls(fset *token.FileSet, file *ast.File) []*funcImpl {
	ordered := make([]*funcImpl, 0)
	named := make(map[string]*funcImpl)
	typed := make(map[string]*funcImpl)

	for i, decl := range file.Decls {
		// ast.Print(fset, decl)
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if len(decl.Specs) != 1 {
				continue
			}
			switch decl.Tok {
			case token.VAR:
				valueSpec, ok := decl.Specs[0].(*ast.ValueSpec)
				if !ok || len(valueSpec.Names) != 1 {
					continue
				}
				implVarName := valueSpec.Names[0].Name

				if len(valueSpec.Values) == 0 {
					// Example:
					//   // documentCanUserRead wraps document.CanUserRead as command.Function
					//   var documentCanUserRead documentCanUserReadT
					comment := strings.TrimSpace(decl.Doc.Text())
					prefix := implVarName + " wraps "
					suffix := " as command.Function"
					if !strings.HasPrefix(comment, prefix) || !strings.HasSuffix(comment, suffix) {
						continue
					}
					wrappedFunc := comment[len(prefix) : len(comment)-len(suffix)]
					impl := named[implVarName]
					if impl == nil {
						impl = new(funcImpl)
						ordered = append(ordered, impl)
						named[implVarName] = impl
					}
					impl.VarName = implVarName
					impl.WrappedFunc = wrappedFunc
					impl.DeclIndices = append(impl.DeclIndices, i)
					impl.Type = astvisit.ExprString(valueSpec.Type)
					impl.Comments = append(impl.Comments, decl.Doc)
					typed[impl.Type] = impl
					continue
				}

				if len(valueSpec.Values) != 1 {
					continue
				}
				callExpr, ok := valueSpec.Values[0].(*ast.CallExpr)
				if !ok || len(callExpr.Args) != 1 || astvisit.ExprString(callExpr.Fun) != "command.GenerateFunctionTODO" {
					continue
				}
				impl := named[implVarName]
				if impl == nil {
					impl = new(funcImpl)
					ordered = append(ordered, impl)
					named[implVarName] = impl
				}
				impl.VarName = implVarName
				impl.WrappedFunc = astvisit.ExprString(callExpr.Args[0])
				impl.DeclIndices = append(impl.DeclIndices, i)
				if decl.Doc != nil {
					impl.Comments = append(impl.Comments, decl.Doc)
				}

			case token.TYPE:
				// ast.Print(fset, decl)
				typeSpec, ok := decl.Specs[0].(*ast.TypeSpec)
				if !ok || astvisit.ExprString(typeSpec.Type) != "struct{}" {
					continue
				}
				implTypeName := typeSpec.Name.Name
				// Example:
				//   // documentCanUserReadT wraps document.CanUserRead as command.Function
				//   type documentCanUserReadT struct{}
				comment := strings.TrimSpace(decl.Doc.Text())
				prefix := implTypeName + " wraps "
				suffix := " as command.Function"
				if !strings.HasPrefix(comment, prefix) || !strings.HasSuffix(comment, suffix) {
					continue
				}
				wrappedFunc := comment[len(prefix) : len(comment)-len(suffix)]
				impl := typed[implTypeName]
				if impl == nil {
					impl = new(funcImpl)
					ordered = append(ordered, impl)
					typed[implTypeName] = impl
					impl.Type = implTypeName
					// No var with that type declared
					// so also use the type like a var
					// and let the user instanciate the type with {}
					named[implTypeName] = impl
					impl.VarName = implTypeName
				}
				impl.WrappedFunc = wrappedFunc
				impl.DeclIndices = append(impl.DeclIndices, i)
				impl.Comments = append(impl.Comments, decl.Doc)
			}

		case *ast.FuncDecl:
			if decl.Recv.NumFields() != 1 {
				continue
			}
			recvType := astvisit.ExprString(decl.Recv.List[0].Type)
			impl := typed[recvType]
			if impl == nil {
				continue
			}
			impl.DeclIndices = append(impl.DeclIndices, i)
			if decl.Doc != nil {
				impl.Comments = append(impl.Comments, decl.Doc)
			}
		}
	}

	for _, impl := range ordered {
		sort.Ints(impl.DeclIndices)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].DeclIndices[0] < ordered[j].DeclIndices[0] })
	return ordered
}
