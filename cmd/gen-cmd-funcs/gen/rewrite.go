package gen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-astvisit"
)

func Rewrite(path string, verbose bool, printOnly io.Writer) (err error) {
	recursive := strings.HasSuffix(path, "...")
	if recursive {
		path = strings.TrimSuffix(path, "...")
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return RewriteFile(path, verbose, printOnly)
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
		err = RewriteAstFile(fset, pkg, file, fileName, verbose, printOnly)
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
		return Rewrite(filepath.Join(path, d.Name())+"...", verbose, printOnly)
	})
}

func RewriteFile(filePath string, verbose bool, printOnly io.Writer) (err error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("file path is a directory: %s", filePath)
	}
	fset := token.NewFileSet()
	pkg, err := astvisit.ParsePackage(fset, filepath.Dir(filePath), filterOutTests)
	if err != nil {
		return err
	}
	return RewriteAstFile(fset, pkg, pkg.Files[filePath], filePath, verbose, printOnly)
}

func RewriteAstFile(fset *token.FileSet, filePkg *ast.Package, file *ast.File, filePath string, verbose bool, printTo io.Writer) (err error) {
	// ast.Print(fset, file)

	funcImpls := findFuncImpls(fset, file)
	if len(funcImpls) == 0 {
		if verbose {
			fmt.Println("nothing found to rewrite in", filePath)
		}
		return nil
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

	var replacements astvisit.NodeReplacements
	for _, impl := range funcImpls {
		importName, funcName := impl.WrappedFuncPkgAndFuncName()
		referencedPkg, ok := functions[importName]
		if !ok {
			return fmt.Errorf("can't find package %s in imports of file %s", importName, filePath)
		}
		wrappedFunc, ok := referencedPkg.Funcs[funcName]
		if !ok {
			return fmt.Errorf("can't find function %s in package %s", funcName, importName)
		}

		var repl strings.Builder
		// fmt.Fprintf(&newSrc, "////////////////////////////////////////\n")
		// fmt.Fprintf(&newSrc, "// %s\n\n", impl.WrappedFunc)
		fmt.Fprintf(&repl, "// %s wraps %s as %s (generated code)\n", impl.VarName, impl.WrappedFunc, impl.Implements)
		fmt.Fprintf(&repl, "var %[1]s %[1]sT\n\n", impl.VarName)
		err = impl.Implements.WriteFunction(&repl, file, wrappedFunc.Decl, impl.VarName+"T", importName)
		if err != nil {
			return err
		}

		var implReplacements astvisit.NodeReplacements
		for i, node := range impl.Nodes {
			if i == 0 {
				implReplacements.AddReplacement(node, repl.String())
			} else {
				implReplacements.AddRemoval(node)
			}
		}
		replacements.Add(implReplacements)
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	rewritten, err := replacements.Apply(fset, source)
	if err != nil {
		return err
	}
	source, err = format.Source(source)
	if err != nil {
		return err
	}

	if printTo != nil {
		_, err = printTo.Write(rewritten)
		return err
	}
	if verbose {
		fmt.Println("rewriting", filePath)
	}
	return ioutil.WriteFile(filePath, rewritten, 0660)
}

type funcImpl struct {
	VarName     string
	WrappedFunc string
	Type        string
	Nodes       []ast.Node
	Implements  Impl
}

func (impl *funcImpl) WrappedFuncPkgAndFuncName() (pkgName, funcName string) {
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

	for _, decl := range file.Decls {
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
					//   // documentCanUserRead wraps document.CanUserRead as function.Wrapper (generated code)
					//   var documentCanUserRead documentCanUserReadT
					wrappedFunc, implements, err := parseImplementsComment(implVarName, decl.Doc.Text())
					if err != nil {
						continue
					}
					impl := named[implVarName]
					if impl == nil {
						impl = new(funcImpl)
						ordered = append(ordered, impl)
						named[implVarName] = impl
					}
					impl.VarName = implVarName
					impl.WrappedFunc = wrappedFunc
					impl.Implements |= implements
					impl.Type = astvisit.ExprString(valueSpec.Type)
					if decl.Doc != nil {
						impl.Nodes = append(impl.Nodes, decl.Doc)
					}
					impl.Nodes = append(impl.Nodes, decl)

					typed[impl.Type] = impl
					continue
				}

				if len(valueSpec.Values) != 1 {
					continue
				}
				callExpr, ok := valueSpec.Values[0].(*ast.CallExpr)
				if !ok || len(callExpr.Args) != 1 {
					continue
				}
				todoFunc := astvisit.ExprString(callExpr.Fun)
				if !strings.HasSuffix(todoFunc, "TODO") {
					continue
				}
				implements, err := ImplFromString(strings.TrimSuffix(todoFunc, "TODO"))
				if err != nil {
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
				impl.Implements |= implements
				if decl.Doc != nil {
					impl.Nodes = append(impl.Nodes, decl.Doc)
				}
				impl.Nodes = append(impl.Nodes, decl)

			case token.TYPE:
				// ast.Print(fset, decl)
				typeSpec, ok := decl.Specs[0].(*ast.TypeSpec)
				if !ok || astvisit.ExprString(typeSpec.Type) != "struct{}" {
					continue
				}
				implTypeName := typeSpec.Name.Name
				// Example:
				//   // documentCanUserReadT wraps document.CanUserRead as function.Wrapper (generated code)
				//   type documentCanUserReadT struct{}
				wrappedFunc, implements, err := parseImplementsComment(implTypeName, decl.Doc.Text())
				if err != nil {
					continue
				}
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
				impl.Implements |= implements
				if decl.Doc != nil {
					impl.Nodes = append(impl.Nodes, decl.Doc)
				}
				impl.Nodes = append(impl.Nodes, decl)
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
			if decl.Doc != nil {
				impl.Nodes = append(impl.Nodes, decl.Doc)
			}
			impl.Nodes = append(impl.Nodes, decl)
		}
	}

	return ordered
}

// parseImplementsComment parses a comment that indicates the wrapped function
// and what interface is implemented
//
// Example:
//   // documentCanUserRead wraps document.CanUserRead as function.Wrapper (generated code)
//   var documentCanUserRead documentCanUserReadT
// or:
//   // documentCanUserReadT wraps document.CanUserRead as function.Wrapper (generated code)
//   type documentCanUserReadT struct{}
func parseImplementsComment(implementor, comment string) (wrappedFunc string, implements Impl, err error) {
	comment = strings.TrimSuffix(strings.TrimSpace(comment), " (generated code)")
	prefix := implementor + " wraps "
	asPos := strings.Index(comment, " as ")
	if !strings.HasPrefix(comment, prefix) || asPos <= len(prefix) {
		return "", 0, errors.New("no implementation comment")
	}
	wrappedFunc = comment[len(prefix):asPos]
	implements, err = ImplFromString(comment[asPos+len(" as "):])
	if err != nil {
		return "", 0, err
	}
	return wrappedFunc, implements, nil
}
