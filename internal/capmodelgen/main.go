// capmodelgen generates simplified "capability model" types from the Tabularium model
// registry. It reads capmodel struct tags on registered model fields and emits one Go
// source file per type into pkg/capmodel/, each containing a plain struct and a Convert
// method that produces the corresponding internal model.
//
// Usage:
//
//	go run ./cmd/capmodelgen -output pkg/capmodel/
package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func main() {
	output := flag.String("output", "", "output directory for generated files")
	flag.Parse()

	if *output == "" {
		fmt.Fprintln(os.Stderr, "usage: capmodelgen -output <dir>")
		os.Exit(1)
	}

	types := parseCapmodelTags(registry.Registry)

	if err := generate(types, *output); err != nil {
		fmt.Fprintf(os.Stderr, "error generating code: %v\n", err)
		os.Exit(1)
	}
}
