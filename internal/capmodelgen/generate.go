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

// needsRemap reports whether a field's capmodel JSON name differs from its
// source JSON name(s), meaning the converter must remap the key.
func needsRemap(f field) bool {
	return len(f.SourceJSONNames) > 1 || f.SourceJSONNames[0] != f.JSONName
}

var funcMap = template.FuncMap{
	"hasRemap": func(ts typeSpec) bool {
		for _, f := range ts.Fields {
			if needsRemap(f) {
				return true
			}
		}
		return false
	},
	"remapFields": func(ts typeSpec) []field {
		var result []field
		for _, f := range ts.Fields {
			if needsRemap(f) {
				result = append(result, f)
			}
		}
		return result
	},
	"needsJSON": func(specs []typeSpec) bool {
		for _, ts := range specs {
			if ts.Parent != nil {
				return true
			}
			for _, f := range ts.Fields {
				if needsRemap(f) {
					return true
				}
			}
		}
		return false
	},
}

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

var convertTmpl = template.Must(template.New("convert").Funcs(funcMap).Parse(convertTmplStr))
