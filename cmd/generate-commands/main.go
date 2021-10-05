package main

import (
	"os"

	"github.com/ungerik/go-command"
)

var pkgDir string

func main() {
	if len(os.Args) < 2 {
		panic("need pkgDir argument")
	}
	pkgDir = os.Args[1]
	err := command.GeneratePackageFunctions(pkgDir, "pkgfuncs.go")
	if err != nil {
		panic(err)
	}
}
