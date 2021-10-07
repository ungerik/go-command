package gen

import (
	"fmt"
	"go/ast"

	"github.com/ungerik/go-astvisit"
)

func funcDeclArgNames(funcDecl *ast.FuncDecl) (names []string) {
	for _, field := range funcDecl.Type.Params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

func funcDeclArgTypes(funcDecl *ast.FuncDecl) (types []string) {
	for _, field := range funcDecl.Type.Params.List {
		for range field.Names {
			types = append(types, astvisit.ExprString(field.Type))
		}
	}
	return types
}

func funcDeclResultTypes(funcDecl *ast.FuncDecl) (types []string) {
	for _, field := range funcDecl.Type.Results.List {
		types = append(types, astvisit.ExprString(field.Type))
		for i := 1; i < len(field.Names); i++ {
			types = append(types, astvisit.ExprString(field.Type))
		}
	}
	return types
}

func recursiveExprSelectors(expr ast.Expr, selectors map[string]struct{}) {
	switch e := expr.(type) {
	case *ast.Ident:
		// Name without selector
	case *ast.SelectorExpr:
		selectors[e.X.(*ast.Ident).Name] = struct{}{}
	case *ast.StarExpr:
		recursiveExprSelectors(e.X, selectors)
	case *ast.Ellipsis:
		recursiveExprSelectors(e.Elt, selectors)
	case *ast.ArrayType:
		recursiveExprSelectors(e.Elt, selectors)
	case *ast.StructType:
		for _, f := range e.Fields.List {
			recursiveExprSelectors(f.Type, selectors)
		}
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			recursiveExprSelectors(elt, selectors)
		}
	case *ast.MapType:
		recursiveExprSelectors(e.Key, selectors)
		recursiveExprSelectors(e.Value, selectors)
	case *ast.ChanType:
		recursiveExprSelectors(e.Value, selectors)
	case *ast.FuncType:
		for _, p := range e.Params.List {
			recursiveExprSelectors(p.Type, selectors)
		}
		for _, r := range e.Results.List {
			recursiveExprSelectors(r.Type, selectors)
		}
	default:
		panic(fmt.Sprintf("UNSUPPORTED: %#v", expr))
	}
}
