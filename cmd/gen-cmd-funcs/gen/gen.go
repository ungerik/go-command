package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ungerik/go-astvisit"
	"golang.org/x/tools/imports"
)

func PackageFunctions(pkgDir, genFilename, namePrefix string, printOnly bool, onlyFuncs ...string) error {
	pkgName, funcs, err := parsePackage(pkgDir, genFilename, onlyFuncs...)
	if err != nil {
		return err
	}

	importLines := map[string]bool{
		`"reflect"`: true,
		`command "github.com/ungerik/go-command"`: true,
	}
	for _, fun := range funcs {
		err = GetFunctionImports(importLines, fun.File, fun.Decl)
		if err != nil {
			return err
		}
	}
	var sortedImportLines []string
	for l := range importLines {
		sortedImportLines = append(sortedImportLines, l)
	}
	sort.Strings(sortedImportLines)

	b := bytes.NewBuffer(nil)

	fmt.Fprintf(b, "// This file has been AUTOGENERATED!\n\n")
	fmt.Fprintf(b, "package %s\n\n", pkgName)
	if len(sortedImportLines) > 0 {
		fmt.Fprintf(b, "import (\n")
		for _, importLine := range sortedImportLines {
			fmt.Fprintf(b, "\t%s\n", importLine)
		}
		fmt.Fprintf(b, ")\n\n")
	}

	for funName, fun := range funcs {
		err = WriteFunctionImpl(b, fun.File, fun.Decl, namePrefix+funName, "")
		if err != nil {
			return err
		}
	}

	genFileData := b.Bytes()
	genFilePath := filepath.Join(pkgDir, genFilename)

	imports.LocalPrefix = "github.com/ungerik/"
	genFileData, err = imports.Process(genFilePath, genFileData, &imports.Options{Comments: true, FormatOnly: true})
	if err != nil {
		return err
	}

	if printOnly {
		fmt.Println(genFileData)
	} else {
		fmt.Println("Writing file", genFilePath)
		err = ioutil.WriteFile(genFilePath, genFileData, 0660)
		if err != nil {
			return err
		}
	}
	// err = exec.Command("gofmt", "-s", "-w", genFile).Run()
	// if err != nil {
	// 	return err
	// }

	return nil
}

func filterGoFiles(excludeFilenames ...string) func(info os.FileInfo) bool {
	return func(info os.FileInfo) bool {
		name := info.Name()
		for _, exclude := range excludeFilenames {
			if name == exclude {
				return false
			}
		}
		if strings.HasSuffix(name, "_test.go") {
			return false
		}
		return true
	}
}

func RewriteGenerateFunctionTODOs(filePath string, printOnly bool) (err error) {
	fileDir := filepath.Dir(filePath)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, fileData, parser.DeclarationErrors)
	if err != nil {
		return err
	}

	imports := make(map[string]string)
	for _, imp := range file.Imports {
		var pkgName string
		if imp.Name != nil {
			pkgName = imp.Name.Name
		} else {
			pkgName, err = guessPackageNameFromPath(imp.Path.Value)
			if err != nil {
				return err
			}
		}
		pkgPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return err
		}
		imports[pkgName] = pkgPath
	}

	importFuncs := make(map[string]map[string]funcInfo)
	for pkgName, pkgImportPath := range imports {
		loc, err := astvisit.LocatePackage(fileDir, pkgImportPath)
		if err != nil {
			return err
		}
		if loc.Std {
			continue
		}
		_, funcs, err := parsePackage(loc.SourcePath, "")
		if err != nil {
			return err
		}
		importFuncs[pkgName] = funcs
	}

	buf := bytes.NewBuffer(nil)
	nextSourceOffset := 0
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		if len(genDecl.Specs) != 1 {
			continue
		}
		// ast.Print(fileSet, decl)
		valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
			continue
		}
		implName := valueSpec.Names[0].Name
		callExpr, ok := valueSpec.Values[0].(*ast.CallExpr)
		if !ok || len(callExpr.Args) != 1 || astvisit.ExprString(callExpr.Fun) != "command.GenerateFunctionTODO" {
			continue
		}
		sel, ok := callExpr.Args[0].(*ast.SelectorExpr)
		if !ok {
			continue
		}
		pkgName := astvisit.ExprString(sel.X)
		funcName := astvisit.ExprString(sel.Sel)
		pkgFuncs, ok := importFuncs[pkgName]
		if !ok {
			continue
		}
		fun, ok := pkgFuncs[funcName]
		if !ok {
			continue
		}
		// Found a function to replace the placeholder variable,
		// write all unwritten source bytes until variable declaraton
		endOffset := fset.Position(decl.Pos()).Offset
		buf.Write(fileData[nextSourceOffset:endOffset])
		nextSourceOffset = fset.Position(decl.End()).Offset

		fmt.Fprintf(buf, "////////////////////////////////////////\n")
		fmt.Fprintf(buf, "// %s.%s\n\n", pkgName, fun.Decl.Name.Name)
		fmt.Fprintf(buf, "// %s wraps %s.%s as command.Function (generated code)\n", implName, pkgName, fun.Decl.Name.Name)
		fmt.Fprintf(buf, "var %[1]s %[1]sT\n\n", implName)
		err = WriteFunctionImpl(buf, fun.File, fun.Decl, implName+"T", pkgName)
		if err != nil {
			return err
		}

		// format.Node(buf)
	}
	// Write unwritten rest of original source bytes
	buf.Write(fileData[nextSourceOffset:])

	if printOnly {
		fmt.Println(buf.String())
	} else {
		fmt.Println("Writing file", filePath)
		err = ioutil.WriteFile(filePath, buf.Bytes(), 0660)
		if err != nil {
			return err
		}
	}
	return nil
}
