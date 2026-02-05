// Package external provides simplified Go types for external tool writers
// and mechanisms to convert them to full Tabularium types.
package external

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	_ "github.com/praetorian-inc/tabularium/pkg/model/model" // Ensure init() functions run
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// FieldInfo holds information about a struct field for code generation.
type FieldInfo struct {
	Name        string // Go field name
	JSONName    string // JSON tag name
	Type        string // Go type as string
	Description string // desc tag value
	Example     string // example tag value
	Omitempty   bool   // Whether field has omitempty
}

// ModelInfo holds information about a model for code generation.
type ModelInfo struct {
	Name        string      // Lowercase registry name (e.g., "asset")
	TypeName    string      // Go type name (e.g., "Asset")
	Description string      // Model description
	Fields      []FieldInfo // Fields with external:"true" tag
}

// Generator generates simplified external Go types from registered models.
type Generator struct {
	// Models to include (empty means all with external fields)
	IncludeModels []string
}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates simplified Go types for all registered models that have
// fields tagged with external:"true" and writes them to the output directory.
func (g *Generator) Generate(outputDir string) error {
	models := g.extractModels()

	if len(models) == 0 {
		return fmt.Errorf("no models with external fields found")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the types file
	outputPath := filepath.Join(outputDir, "types_generated.go")
	return g.generateTypesFile(models, outputPath)
}

// extractModels extracts all registered models with external fields.
func (g *Generator) extractModels() []*ModelInfo {
	types := registry.Registry.GetAllTypes()
	var models []*ModelInfo

	for name, typ := range types {
		// Skip if we have a filter and this model isn't in it
		if len(g.IncludeModels) > 0 && !contains(g.IncludeModels, name) {
			continue
		}

		// Handle pointer types
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			continue
		}

		model := g.extractModelInfo(name, typ)
		if model != nil && len(model.Fields) > 0 {
			models = append(models, model)
		}
	}

	// Sort by name for consistent output
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models
}

// extractModelInfo extracts information about a single model.
func (g *Generator) extractModelInfo(name string, typ reflect.Type) *ModelInfo {
	model := &ModelInfo{
		Name:     name,
		TypeName: capitalize(name),
	}

	// Get description from the Model interface
	modelPtrInstance := reflect.New(typ).Interface()
	if m, ok := modelPtrInstance.(registry.Model); ok {
		model.Description = m.GetDescription()
	}

	// Extract fields with external:"true" tag, including from embedded types
	model.Fields = g.extractExternalFields(typ, make(map[reflect.Type]bool))

	return model
}

// extractExternalFields recursively extracts fields with external:"true" tag.
func (g *Generator) extractExternalFields(typ reflect.Type, visited map[reflect.Type]bool) []FieldInfo {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	// Prevent infinite recursion
	if visited[typ] {
		return nil
	}
	visited[typ] = true

	var fields []FieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded/anonymous fields by recursively extracting
		if field.Anonymous {
			embeddedFields := g.extractExternalFields(field.Type, visited)
			fields = append(fields, embeddedFields...)
			continue
		}

		// Check for external:"true" tag
		externalTag := field.Tag.Get("external")
		if externalTag != "true" {
			continue
		}

		// Get JSON info
		jsonName, omitempty, include := getJSONInfo(field)
		if !include {
			continue
		}

		fieldInfo := FieldInfo{
			Name:        field.Name,
			JSONName:    jsonName,
			Type:        goTypeString(field.Type),
			Description: field.Tag.Get("desc"),
			Example:     field.Tag.Get("example"),
			Omitempty:   omitempty,
		}

		fields = append(fields, fieldInfo)
	}

	return fields
}

// generateTypesFile generates the Go source file with simplified types.
func (g *Generator) generateTypesFile(models []*ModelInfo, outputPath string) error {
	tmpl, err := template.New("types").Funcs(template.FuncMap{
		"capitalize": capitalize,
	}).Parse(typesTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	data := struct {
		Models []*ModelInfo
	}{
		Models: models,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// goTypeString returns the Go type representation as a string.
func goTypeString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + goTypeString(t.Elem())
	case reflect.Slice:
		return "[]" + goTypeString(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), goTypeString(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", goTypeString(t.Key()), goTypeString(t.Elem()))
	default:
		// Handle qualified types (e.g., time.Time)
		if t.PkgPath() != "" {
			pkgParts := strings.Split(t.PkgPath(), "/")
			pkgName := pkgParts[len(pkgParts)-1]
			return pkgName + "." + t.Name()
		}
		return t.Name()
	}
}

// getJSONInfo extracts JSON tag information from a struct field.
func getJSONInfo(field reflect.StructField) (name string, omitempty bool, include bool) {
	jsonTag := field.Tag.Get("json")
	name = field.Name
	include = true

	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] == "-" {
			include = false
			return
		}
		if parts[0] != "" {
			name = parts[0]
		}
		for _, part := range parts[1:] {
			if part == "omitempty" {
				omitempty = true
				break
			}
		}
	}
	return
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

const typesTemplate = `// Code generated by externalgen. DO NOT EDIT.
// This file contains simplified types for external tool writers.
// Use the Convert() function to transform these into full Tabularium types.

package external

{{range .Models}}
// {{.TypeName}} is a simplified version of the Tabularium {{.Name}} type.
// {{.Description}}
type {{.TypeName}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} ` + "`" + `json:"{{.JSONName}}{{if .Omitempty}},omitempty{{end}}"` + "`" + `{{if .Description}} // {{.Description}}{{end}}
{{- end}}
}

{{end}}
// ModelType returns the tabularium model type name for the given external type.
func ModelType(v any) string {
	switch v.(type) {
{{- range .Models}}
	case {{.TypeName}}, *{{.TypeName}}:
		return "{{.Name}}"
{{- end}}
	default:
		return ""
	}
}
`
