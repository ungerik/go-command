package gen

import (
	"fmt"
	"go/ast"
	"strings"
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
			types = append(types, astExprString(field.Type))
		}
	}
	return types
}

func funcDeclResultTypes(funcDecl *ast.FuncDecl) (types []string) {
	for _, field := range funcDecl.Type.Results.List {
		types = append(types, astExprString(field.Type))
		for i := 1; i < len(field.Names); i++ {
			types = append(types, astExprString(field.Type))
		}
	}
	return types
}

func funcTypeString(functype *ast.FuncType) string {
	var b strings.Builder
	b.WriteByte('(')
	for fieldIndex, field := range functype.Params.List {
		if fieldIndex > 0 {
			b.WriteString(", ")
		}
		for i, name := range field.Names {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(name.Name)
		}
		b.WriteByte(' ')
		b.WriteString(astExprString(field.Type))
	}
	b.WriteByte(')')
	if len(functype.Results.List) == 0 {
		return b.String()
	}
	b.WriteByte(' ')
	if len(functype.Results.List) == 1 && len(functype.Results.List[0].Names) == 0 {
		b.WriteString(astExprString(functype.Results.List[0].Type))
		return b.String()
	}
	b.WriteByte('(')
	for fieldIndex, field := range functype.Results.List {
		if fieldIndex > 0 {
			b.WriteString(", ")
		}
		for i, name := range field.Names {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(name.Name)
		}
		b.WriteByte(' ')
		b.WriteString(astExprString(field.Type))
	}
	b.WriteByte(')')
	return b.String()
}

func astExprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return astExprString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + astExprString(e.X)
	case *ast.Ellipsis:
		return "..." + astExprString(e.Elt)
	case *ast.ArrayType:
		return "[" + astExprString(e.Len) + "]" + astExprString(e.Elt)
	case *ast.MapType:
		return "map[" + astExprString(e.Key) + "]" + astExprString(e.Value)
	case *ast.FuncType:
		return "func" + funcTypeString(e)
	default:
		panic(fmt.Sprintf("UNKNOWN: %#v", expr))
	}
}

func astExprSelectors(expr ast.Expr, selectors map[string]struct{}) {
	switch e := expr.(type) {
	case *ast.Ident:
		// Name without selector
	case *ast.SelectorExpr:
		selectors[e.X.(*ast.Ident).Name] = struct{}{}
	case *ast.StarExpr:
		astExprSelectors(e.X, selectors)
	case *ast.Ellipsis:
		astExprSelectors(e.Elt, selectors)
	case *ast.ArrayType:
		astExprSelectors(e.Elt, selectors)
	case *ast.StructType:
		for _, f := range e.Fields.List {
			astExprSelectors(f.Type, selectors)
		}
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			astExprSelectors(elt, selectors)
		}
	case *ast.MapType:
		astExprSelectors(e.Key, selectors)
		astExprSelectors(e.Value, selectors)
	case *ast.ChanType:
		astExprSelectors(e.Value, selectors)
	case *ast.FuncType:
		for _, p := range e.Params.List {
			astExprSelectors(p.Type, selectors)
		}
		for _, r := range e.Results.List {
			astExprSelectors(r.Type, selectors)
		}
	default:
		panic(fmt.Sprintf("UNSUPPORTED: %#v", expr))
	}
}
