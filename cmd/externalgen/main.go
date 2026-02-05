// Command externalgen generates simplified Go types from Tabularium models.
//
// It scans all registered models for fields tagged with external:"true" and
// generates simplified struct definitions that external tool writers can use.
//
// Usage:
//
//	go run ./cmd/externalgen -output pkg/external/generated
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/external"
)

func main() {
	outputDir := flag.String("output", "", "Output directory for generated code (required)")
	models := flag.String("models", "", "Comma-separated list of models to include (empty = all)")
	flag.Parse()

	if *outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: -output flag is required")
		flag.Usage()
		os.Exit(1)
	}

	gen := external.NewGenerator()

	if *models != "" {
		gen.IncludeModels = strings.Split(*models, ",")
	}

	if err := gen.Generate(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating external types: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated external types in %s\n", *outputDir)
}
