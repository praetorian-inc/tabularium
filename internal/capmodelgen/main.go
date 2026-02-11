// capmodelgen generates simplified "capability model" types from the Tabularium model
// registry. It reads capmodel struct tags on registered model fields and emits one Go
// source file per type, each containing a plain struct. Optionally generates a
// convert_gen.go with registry converters into a separate directory.
//
// Usage:
//
//	go run ./internal/capmodelgen -output internal/capmodel/
//	go run ./internal/capmodelgen -output internal/capmodel/ -converters pkg/capmodel/
package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func main() {
	output := flag.String("output", "", "output directory for generated model files")
	converters := flag.String("converters", "", "output directory for generated converter file (defaults to -output)")
	flag.Parse()

	if *output == "" {
		fmt.Fprintln(os.Stderr, "usage: capmodelgen -output <dir> [-converters <dir>]")
		os.Exit(1)
	}

	converterDir := *output
	if *converters != "" {
		converterDir = *converters
	}

	types := parseCapmodelTags(registry.Registry)

	if err := generate(types, *output, converterDir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating code: %v\n", err)
		os.Exit(1)
	}
}
