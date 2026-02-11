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

// templateData is the data passed to convert_gen.go and extract_gen.go templates.
type templateData struct {
	Package   string
	TypeSpecs []typeSpec
}

// generate renders all generated files into outputDir:
//   - one model file per typeSpec
//   - convert_gen.go with registry converters
//   - extract_gen.go with registry extractors
func generate(typeSpecs []typeSpec, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, st := range typeSpecs {
		if err := writeFormatted(modelTmpl, st, outputDir, strings.ToLower(st.Name)+"_model.go"); err != nil {
			return err
		}
	}

	data := templateData{
		Package:   filepath.Base(outputDir),
		TypeSpecs: typeSpecs,
	}

	if err := writeFormatted(convertTmpl, data, outputDir, "convert_gen.go"); err != nil {
		return err
	}
	return writeFormatted(extractTmpl, data, outputDir, "extract_gen.go")
}

// writeFormatted executes a template, formats the output as Go source, and writes it.
func writeFormatted(tmpl *template.Template, data any, dir, filename string) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("generating %s: %w", filename, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting %s: %w", filename, err)
	}

	return os.WriteFile(filepath.Join(dir, filename), formatted, 0644)
}

//go:embed model.go.tmpl
var modelTmplStr string

var modelTmpl = template.Must(template.New("model").Parse(modelTmplStr))

//go:embed convert.go.tmpl
var convertTmplStr string

var convertTmpl = template.Must(template.New("convert").Funcs(funcMap).Parse(convertTmplStr))

//go:embed extract.go.tmpl
var extractTmplStr string

var extractTmpl = template.Must(template.New("extract").Parse(extractTmplStr))
