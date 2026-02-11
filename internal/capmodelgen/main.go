// capmodelgen generates simplified "capability model" types from the Tabularium model
// registry. It reads capmodel struct tags on registered model fields and emits:
//   - one model file per type (plain struct with JSON tags)
//   - convert_gen.go (capmodel JSON → chariot model)
//   - extract_gen.go (chariot model → capmodel struct)
//
// Usage:
//
//	go run ./internal/capmodelgen -output pkg/capmodel/
package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func main() {
	output := flag.String("output", "", "output directory for all generated files")
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
