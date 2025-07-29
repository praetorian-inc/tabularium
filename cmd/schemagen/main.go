package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/praetorian-inc/tabularium/pkg/schema"
	"gopkg.in/yaml.v2"
)

func main() {
	outputFile := flag.String("output", "", "Output file path for the OpenAPI schema (if not specified, prints to stdout)")
	flag.Parse()

	// Generate the OpenAPI document using the library
	doc, err := schema.GenerateOpenAPISchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating OpenAPI schema: %v\n", err)
		os.Exit(1)
	}

	// Convert to YAML
	bytes, err := yaml.Marshal(doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema to YAML: %v\n", err)
		os.Exit(1)
	}

	// Handle output
	if *outputFile == "" {
		fmt.Print(string(bytes))
		return
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// Write to file
	if err := os.WriteFile(*outputFile, bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema to %s: %v\n", *outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("Generated OpenAPI schema at %s\n", *outputFile)
}
