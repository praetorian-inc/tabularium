package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GeneratorFunc defines the signature for functions that generate code for a specific language.
type GeneratorFunc func(inputFile, outputDir string) error

// generatorRegistry maps language identifiers (e.g., "py") to their corresponding GeneratorFunc.
var generatorRegistry = map[string]GeneratorFunc{
	"py": generatePythonPydantic,
}

// generationTarget stores the language and output directory for a single generation request.
type generationTarget struct {
	Lang      string
	OutputDir string
}

// generationTargets implements the flag.Value interface
// for parsing multiple "-gen lang:output_dir" flags.
type generationTargets []generationTarget

func (g *generationTargets) String() string {
	// Required by flag.Value interface.
	var targets []string
	for _, t := range *g {
		targets = append(targets, fmt.Sprintf("%s:%s", t.Lang, t.OutputDir))
	}
	return strings.Join(targets, ", ")
}

func (g *generationTargets) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for generation target: %q. Expected lang:output_dir", value)
	}
	lang := parts[0]
	outputDir := parts[1]
	if _, ok := generatorRegistry[lang]; !ok {
		return fmt.Errorf("unsupported language: %q", lang)
	}
	*g = append(*g, generationTarget{Lang: lang, OutputDir: outputDir})
	return nil
}

func main() {
	var inputFile string
	var targets generationTargets

	flag.StringVar(&inputFile, "input", "", "Path to the input OpenAPI schema file (required)")
	flag.Var(&targets, "gen", "Generation target in the format lang:output_dir (can be specified multiple times)")
	flag.Parse()

	if inputFile == "" {
		log.Fatal("-input flag is required")
	}
	if len(targets) == 0 {
		log.Fatal("-gen flag must be specified at least once")
	}

	absInputFile, err := filepath.Abs(inputFile)
	if err != nil {
		log.Fatalf("Error getting absolute path for input file %s: %v", inputFile, err)
	}

	for _, target := range targets {
		generatorFunc := generatorRegistry[target.Lang]
		absOutputDir, err := filepath.Abs(target.OutputDir)
		if err != nil {
			log.Fatalf("Error getting absolute path for output directory %s: %v", target.OutputDir, err)
		}

		fmt.Printf("Generating %s code from %s to %s...\n", target.Lang, absInputFile, absOutputDir)

		if err := os.MkdirAll(absOutputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory %s: %v", absOutputDir, err)
		}

		if err := generatorFunc(absInputFile, absOutputDir); err != nil {
			log.Fatalf("Error generating %s code: %v", target.Lang, err)
		}
		fmt.Printf("Successfully generated %s code in %s\n", target.Lang, absOutputDir)
	}
}

func generatePythonPydantic(inputFile, outputDir string) error {
	// Assumes datamodel-codegen is installed and in PATH.
	outputFile := filepath.Join(outputDir, "models.py")
	cmd := exec.Command("datamodel-codegen",
		"--input", inputFile,
		"--input-file-type", "openapi",
		"--output", outputFile,
		"--output-model-type", "pydantic_v2.BaseModel",
		"--use-annotated",
		"--target-python-version", "3.11",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s\n", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("datamodel-codegen failed: %w", err)
	}
	return nil
}
