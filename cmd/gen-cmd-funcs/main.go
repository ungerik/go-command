package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ungerik/go-command/cmd/gen-cmd-funcs/gen"
)

var (
	// genFilename   string
	// namePrefix    string
	// exportedFuncs bool
	verbose   bool
	printOnly bool
	printHelp bool
)

func main() {
	// flag.BoolVar(&exportedFuncs, "exported", false, "generate function.Wrapper implementation types exported package functions")
	// flag.StringVar(&genFilename, "genfile", "generated.go", "name of the file to be generated")
	// flag.StringVar(&namePrefix, "prefix", "Func", "prefix for function type names in the same package")
	flag.BoolVar(&verbose, "verbose", false, "prints information of what's happening")
	flag.BoolVar(&printOnly, "print", false, "prints to stdout instead of writing files")
	flag.BoolVar(&printHelp, "help", false, "prints this help output")
	flag.Parse()
	if printHelp {
		flag.PrintDefaults()
		os.Exit(2)
	}

	args := flag.Args()

	// if exportedFuncs {
	// 	if len(args) < 1 {
	// 		fmt.Fprintln(os.Stderr, "gen-cmd-funcs needs package path argument")
	// 		os.Exit(1)
	// 	}
	// 	pkgDir, onlyFuncs := filepath.Clean(args[0]), args[1:]
	// 	err := gen.PackageFunctions(pkgDir, genFilename, namePrefix, printOnly, onlyFuncs...)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, "gen-cmd-funcs error:", err)
	// 		os.Exit(2)
	// 	}
	// 	return
	// }

	filePath, _ := os.Getwd()
	if len(args) > 0 && args[0] != "." {
		filePath = filepath.Clean(args[0])
	}
	var printOnlyWriter io.Writer
	if printOnly {
		printOnlyWriter = os.Stdout
	}
	err := gen.Rewrite(filePath, verbose, printOnlyWriter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen-cmd-funcs error:", err)
		os.Exit(2)
	}
}
