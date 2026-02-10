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

func generate(typeSpecs []typeSpec, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, st := range typeSpecs {
		var buf bytes.Buffer

		if err := typeTmpl.Execute(&buf, st); err != nil {
			return fmt.Errorf("generating %s: %w", st.Name, err)
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			debugPath := filepath.Join(outputDir, "generated_debug.go")
			if writeErr := os.WriteFile(debugPath, buf.Bytes(), 0644); writeErr != nil {
				return fmt.Errorf("formatting %s: %w (also failed to write debug: %v)", st.Name, err, writeErr)
			}
			return fmt.Errorf("formatting %s: %w (debug written to %s)", st.Name, err, debugPath)
		}

		outPath := filepath.Join(outputDir, strings.ToLower(st.Name)+".go")
		if err := os.WriteFile(outPath, formatted, 0644); err != nil {
			return err
		}
	}

	return nil
}

// manualUnmarshal is true when the parent must be set before hooks run,
// requiring manual Defaulted + Unmarshal + CallHooks instead of UnmarshalModel.
func manualUnmarshal(p *parentField) bool {
	return p != nil && p.Kind != parentInterface
}

//go:embed type.go.tmpl
var typeTmplStr string

var typeTmpl = template.Must(template.New("type").Funcs(template.FuncMap{
	"manualUnmarshal": manualUnmarshal,
}).Parse(typeTmplStr))
