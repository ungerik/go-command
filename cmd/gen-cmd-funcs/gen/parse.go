package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type funcInfo struct {
	Decl *ast.FuncDecl
	File *ast.File
}

func parsePackage(pkgDir, excludeFilename string, onlyFuncs ...string) (pkgName string, funcs map[string]funcInfo, err error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgDir, filterGoFiles(excludeFilename), 0)
	if err != nil {
		return "", nil, err
	}
	if len(pkgs) != 1 {
		return "", nil, fmt.Errorf("%d packages found in %s", len(pkgs), pkgDir)
	}
	var files []*ast.File
	for _, p := range pkgs {
		pkgName = p.Name
		for _, file := range p.Files {
			files = append(files, file)
		}
	}

	// // typesInfo.Uses allows to lookup import paths for identifiers.
	// typesInfo := &types.Info{Uses: make(map[*ast.Ident]types.Object)}
	// // Type check the parsed code using the default importer.
	// // Use golang.org/x/tools/go/loader to check a program
	// // consisting of multiple packages.
	// conf := types.Config{Importer: importer.Default()}
	// _, err = conf.Check(pkgDir, fileSet, files, typesInfo)
	// if err != nil {
	// 	return nil, err
	// }

	funcs = make(map[string]funcInfo)
	for _, file := range files {
		// ast.Print(fileSet, file.Imports)
		for _, obj := range file.Scope.Objects {
			if obj.Kind != ast.Fun {
				// ast.Print(fileSet, obj)
				continue
			}
			funcDecl := obj.Decl.(*ast.FuncDecl)
			if len(onlyFuncs) > 0 {
				for _, name := range onlyFuncs {
					if funcDecl.Name.Name == name {
						funcs[name] = funcInfo{Decl: funcDecl, File: file}
						break
					}
				}
			} else if funcDecl.Name.IsExported() {
				funcs[funcDecl.Name.Name] = funcInfo{Decl: funcDecl, File: file}
			}
		}
	}
	return pkgName, funcs, nil
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

func filterOutTests(info os.FileInfo) bool {
	if strings.HasSuffix(info.Name(), "_test.go") {
		return false
	}
	return true
}

// func parsePackage2(pkgDir, genFilename string, onlyFuncs ...string) (pkgName string, funcs map[*ast.FuncDecl]*ast.File, err error) {
// 	config := &packages.Config{
// 		Mode: packages.NeedName + packages.NeedImports + packages.NeedTypes + packages.NeedSyntax + packages.NeedTypesInfo,
// 		Dir:  pkgDir,
// 	}
// 	pkgs, err := packages.Load(config, pkgDir)
// 	if err != nil {
// 		return "", nil, err
// 	}
// 	if err != nil {
// 		return "", nil, err
// 	}
// 	if len(pkgs) != 1 {
// 		return "", nil, fmt.Errorf("%d packages found in %s", len(pkgs), pkgDir)
// 	}
// 	pkgName = pkgs[0].Name
// 	files := pkgs[0].Syntax

// 	funcs = make(map[*ast.FuncDecl]*ast.File)
// 	for _, file := range files {
// 		// ast.Print(fileSet, file.Imports)
// 		for _, obj := range file.Scope.Objects {
// 			if obj.Kind != ast.Fun {
// 				// ast.Print(fileSet, obj)
// 				continue
// 			}
// 			funcDecl := obj.Decl.(*ast.FuncDecl)
// 			if len(onlyFuncs) > 0 {
// 				for _, name := range onlyFuncs {
// 					if funcDecl.Name.Name == name {
// 						funcs[funcDecl] = file
// 						break
// 					}
// 				}
// 			} else if funcDecl.Name.IsExported() {
// 				funcs[funcDecl] = file
// 			}
// 		}
// 	}
// 	return pkgName, funcs, nil
// }
