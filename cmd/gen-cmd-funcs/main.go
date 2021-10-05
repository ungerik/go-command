package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ungerik/go-command/cmd/gen-cmd-funcs/gen"
)

var (
	genFilename  string
	namePrefix   string
	replaceTODOs bool
	printOnly    bool
)

func main() {
	flag.StringVar(&genFilename, "gen", "generated.go", "name of the file to be generated")
	flag.StringVar(&namePrefix, "prefix", "Func", "prefix for function type names in the same package")
	flag.BoolVar(&replaceTODOs, "todo", false, "replaces command.GenerateFunctionTODO with generated types")
	flag.BoolVar(&printOnly, "print", false, "prints to stdout instead of writing files")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "gen-cmd-funcs needs package path argument")
		os.Exit(1)
	}

	if replaceTODOs {
		filePath := args[0]
		err := gen.RewriteGenerateFunctionTODOs(filePath, printOnly)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gen-cmd-funcs error:", err)
			os.Exit(1)
		}
	} else {
		pkgDir, onlyFuncs := args[0], args[1:]
		err := gen.PackageFunctions(pkgDir, genFilename, namePrefix, printOnly, onlyFuncs...)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gen-cmd-funcs error:", err)
			os.Exit(1)
		}
	}
}
