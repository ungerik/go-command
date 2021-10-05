package gen

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

func WriteFunctionImpl(w io.Writer, file *ast.File, funcDecl *ast.FuncDecl, implType, funcPackageSel string) error {
	argNames := funcDeclArgNames(funcDecl)
	argTypes := funcDeclArgTypes(funcDecl)
	if len(argNames) != len(argTypes) {
		panic("len(argNames) != len(argTypes)")
	}
	resultTypes := funcDeclResultTypes(funcDecl)
	hasContextArg := len(argTypes) > 0 && argTypes[0] == "context.Context"
	hasErrorResult := len(resultTypes) > 0 && resultTypes[len(resultTypes)-1] == "error"
	if funcPackageSel != "" && !strings.HasSuffix(funcPackageSel, ".") {
		funcPackageSel += "."
	}

	writeFuncCall := func() {
		numResultsWithoutErr := len(resultTypes)
		if hasErrorResult {
			numResultsWithoutErr--
		}
		if numResultsWithoutErr > 0 {
			fmt.Fprintf(w, "\tresults = make([]interface{}, %d)\n", numResultsWithoutErr)
		}
		fmt.Fprintf(w, "\t")
		for i := 0; i < numResultsWithoutErr; i++ {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "results[%d]", i)
		}
		if hasErrorResult {
			if numResultsWithoutErr == 0 {
				fmt.Fprintf(w, "err")
			} else {
				fmt.Fprintf(w, ", err")
			}
		}
		if len(resultTypes) > 0 {
			fmt.Fprintf(w, " = ")
		}
		ellipsis := ""
		if len(argTypes) > 0 && strings.HasPrefix(argTypes[len(argTypes)-1], "...") {
			ellipsis = "..."
		}
		fmt.Fprintf(w, "%s%s(%s%s)\n", funcPackageSel, funcDecl.Name.Name, strings.Join(argNames, ", "), ellipsis)
		fmt.Fprintf(w, "\treturn results, err\n")
	}

	fmt.Fprintf(w, "type %s struct{}\n\n", implType)

	fmt.Fprintf(w, "func (%s) Name() string {\n", implType)
	fmt.Fprintf(w, "\treturn \"%s\"\n", funcDecl.Name.Name)
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (%s) String() string {\n", implType)
	fmt.Fprintf(w, "\treturn \"%s%s\"\n", funcDecl.Name.Name, funcTypeString(funcDecl.Type))
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (%s) NumArgs() int      { return %d }\n", implType, len(argTypes))
	fmt.Fprintf(w, "func (%s) ContextArg() bool  { return %t }\n", implType, hasContextArg)
	fmt.Fprintf(w, "func (%s) NumResults() int   { return %d }\n", implType, len(resultTypes))
	fmt.Fprintf(w, "func (%s) ErrorResult() bool { return %t }\n\n", implType, hasErrorResult)

	fmt.Fprintf(w, "func (%s) ArgNames() []string {\n", implType)
	{
		fmt.Fprintf(w, "\treturn %#v\n", argNames)
	}
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (%s) ArgTypes() []reflect.Type {\n", implType)
	if len(argTypes) == 0 {
		fmt.Fprintf(w, "\treturn nil\n")
	} else {
		fmt.Fprintf(w, "\treturn []reflect.Type{\n")
		for _, t := range argTypes {
			fmt.Fprintf(w, "\t\treflect.TypeOf((*%s)(nil)).Elem(),\n", strings.Replace(t, "...", "[]", 1))
		}
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (%s) ResultTypes() []reflect.Type {\n", implType)
	if len(resultTypes) == 0 {
		fmt.Fprintf(w, "\treturn nil\n")
	} else {
		fmt.Fprintf(w, "\treturn []reflect.Type{\n")
		for _, t := range resultTypes {
			fmt.Fprintf(w, "\t\treflect.TypeOf((*%s)(nil)).Elem(),\n", t)
		}
		fmt.Fprintf(w, "\t}\n")
	}
	fmt.Fprintf(w, "}\n\n")

	ctxArgName := "ctx"
	if !hasContextArg {
		ctxArgName = "_"
	}
	strsArgName := "strs"
	if len(argNames) == 0 || hasContextArg && len(argNames) == 1 {
		strsArgName = "_"
	}

	fmt.Fprintf(w, "func (f %s) CallWithStrings(%s context.Context, %s ...string) (results []interface{}, err error) {\n", implType, ctxArgName, strsArgName)
	{
		for i, argName := range argNames {
			if i == 0 && hasContextArg {
				if argName != "ctx" {
					fmt.Fprintf(w, "\t%s := ctx\n", argName)
				}
				continue
			}
			strsIndex := i
			if hasContextArg {
				strsIndex--
			}
			fmt.Fprintf(w, "\tvar %s %s\n", argName, strings.Replace(argTypes[i], "...", "[]", 1))
			fmt.Fprintf(w, "\tif len(strs) > %d {\n", strsIndex)
			if argTypes[i] == "string" {
				fmt.Fprintf(w, "\t\t%s = strs[%d]\n", argName, strsIndex)
			} else {
				fmt.Fprintf(w, "\t\terr = command.AssignFromString(&%s, strs[%d])\n", argName, strsIndex)
				fmt.Fprintf(w, "\t\tif err != nil {\n")
				{
					fmt.Fprintf(w, "\t\t\treturn nil, command.NewErrArgFromString(err, f, %q)\n", argName)
				}
				fmt.Fprintf(w, "\t\t}\n")
			}
			fmt.Fprintf(w, "\t}\n")
		}
		writeFuncCall()
	}
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (f %s) CallWithNamedStrings(%s context.Context, %s map[string]string) (results []interface{}, err error) {\n", implType, ctxArgName, strsArgName)
	{
		for i, argName := range argNames {
			if i == 0 && hasContextArg {
				if argName != "ctx" {
					fmt.Fprintf(w, "\t%s := ctx\n", argName)
				}
				continue
			}
			fmt.Fprintf(w, "\tvar %s %s\n", argName, strings.Replace(argTypes[i], "...", "[]", 1))
			fmt.Fprintf(w, "\tif str, ok := strs[%q]; ok {\n", argName)
			if argTypes[i] == "string" {
				fmt.Fprintf(w, "\t\t%s = str\n", argName)
			} else {
				fmt.Fprintf(w, "\t\terr = command.AssignFromString(&%s, str)\n", argName)
				fmt.Fprintf(w, "\t\tif err != nil {\n")
				{
					fmt.Fprintf(w, "\t\t\treturn nil, command.NewErrArgFromString(err, f, %q)\n", argName)
				}
				fmt.Fprintf(w, "\t\t}\n")
			}
			fmt.Fprintf(w, "\t}\n")
		}
		writeFuncCall()
	}
	fmt.Fprintf(w, "}\n\n")

	return nil
}

func FunctionImplString(file *ast.File, funcDecl *ast.FuncDecl, implType, funcPackageSel string) (implSource string, err error) {
	b := new(strings.Builder)
	err = WriteFunctionImpl(b, file, funcDecl, implType, funcPackageSel)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func GetFunctionImports(outImportLines map[string]bool, file *ast.File, funcDecl *ast.FuncDecl) error {
	funcSelectors := make(map[string]struct{})
	astExprSelectors(funcDecl.Type, funcSelectors)
	// fmt.Println(funcSelectors)
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if _, ok := funcSelectors[imp.Name.Name]; ok {
				delete(outImportLines, imp.Path.Value)
				outImportLines[imp.Name.Name+" "+imp.Path.Value] = true
			}
			continue
		}
		guessedName, err := guessPackageNameFromPath(imp.Path.Value)
		if err != nil {
			return err
		}
		if _, ok := funcSelectors[guessedName]; ok && !outImportLines[guessedName+" "+imp.Path.Value] {
			outImportLines[imp.Path.Value] = true
		}
	}
	return nil
}

func guessPackageNameFromPath(path string) (string, error) {
	pkg := path
	if len(pkg) >= 2 && pkg[0] == '"' && pkg[len(pkg)-1] == '"' {
		pkg = pkg[1 : len(pkg)-1]
	}
	pkg = pkg[strings.LastIndex(pkg, "/")+1:]
	pkg = strings.TrimPrefix(pkg, "go-")
	pkg = strings.TrimSuffix(pkg, ".go")
	if pkg == "" || strings.ContainsAny(pkg, ".-") {
		return "", fmt.Errorf("could not guess package name from import path %s", path)
	}
	return pkg, nil
}