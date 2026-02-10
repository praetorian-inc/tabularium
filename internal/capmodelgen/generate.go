package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// generate renders each typeSpec through the model template (one file per type),
// then renders all typeSpecs through the convert template (single file).
func generate(typeSpecs []typeSpec, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, st := range typeSpecs {
		var buf bytes.Buffer

		if err := modelTmpl.Execute(&buf, st); err != nil {
			return fmt.Errorf("generating %s: %w", st.Name, err)
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("formatting %s: %w", st.Name, err)
		}

		outPath := filepath.Join(outputDir, strings.ToLower(st.Name)+"_model.go")
		if err := os.WriteFile(outPath, formatted, 0644); err != nil {
			return err
		}
	}

	// Generate single convert_gen.go with all converters
	var buf bytes.Buffer
	if err := convertTmpl.Execute(&buf, typeSpecs); err != nil {
		return fmt.Errorf("generating convert_gen.go: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting convert_gen.go: %w", err)
	}

	outPath := filepath.Join(outputDir, "convert_gen.go")
	if err := os.WriteFile(outPath, formatted, 0644); err != nil {
		return err
	}

	return nil
}

//go:embed model.go.tmpl
var modelTmplStr string

var modelTmpl = template.Must(template.New("model").Parse(modelTmplStr))

//go:embed convert.go.tmpl
var convertTmplStr string

var convertTmpl = template.Must(template.New("convert").Parse(convertTmplStr))
