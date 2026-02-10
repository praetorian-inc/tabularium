package main

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func derefPtr(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// typeMap maps named Go types to the type string that should appear in generated code.
// This is needed for types whose reflect.Kind doesn't match the desired output
// (e.g., SmartBytes is a named type over []byte, but reflect reports []uint8).
var typeMap = map[string]string{
	"SmartBytes":        "[]byte",
	"CloudResourceType": "string",
}

// field represents a regular (non-parent) field in a generated capmodel type.
// SourceJSONNames may contain multiple entries when several source model fields
// map to the same capmodel field (e.g., DNS and Name both map to "ip" for the IP type).
type field struct {
	SourceFieldName string
	SourceJSONNames []string
	JSONName        string
	GoType          string
}

// parentField represents a parent/embed relationship in a generated capmodel type.
// Wrap is true when the source field is a GraphModelWrapper, which requires
// NewGraphModelWrapper() in the generated Convert method.
type parentField struct {
	SourceFieldName string
	JSONName        string
	EmbedType       string
	Wrap            bool
}

// typeSpec holds all the information needed to generate a single capmodel type file.
// SourceTypeName is the internal model type that Convert() produces (e.g., "Asset").
// It may differ from Name when the capmodel type doesn't correspond to a registered
// model (e.g., capmodel "IP" maps to source type "Asset").
type typeSpec struct {
	Name           string
	SourceTypeName string
	Fields         []field
	Parent         *parentField
	fieldIdx       map[string]int // JSONName → index in Fields
}

// parseCapmodelTags walks every registered model type, extracts capmodel struct tags,
// and builds a typeSpec for each distinct capmodel type name found. A single source
// field may contribute to multiple capmodel types via comma-separated tag entries
// (e.g., `capmodel:"Asset,IP=ip,Domain=domain"`). The returned slice is sorted by name.
func parseCapmodelTags(reg *registry.TypeRegistry) []typeSpec {
	builders := map[string]*typeSpec{}
	// Tracks which (typeName, fieldName) pairs have been processed.
	// Embedded fields appear in multiple registered types that share
	// the same embed, so we skip duplicates.
	visited := map[string]map[string]bool{}

	for _, name := range sortedRegistryNames(reg) {
		typ := derefPtr(mustGetType(reg, name))
		goTypeName := typ.Name()

		for _, sf := range reflect.VisibleFields(typ) {
			if !sf.IsExported() {
				continue
			}
			tag := sf.Tag.Get("capmodel")
			if tag == "" {
				continue
			}

			for _, entry := range strings.Split(tag, ",") {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}

				typeName, jsonName, embedType := parseEntry(entry)

				if !markVisited(visited, typeName, sf.Name) {
					continue
				}

				sourceJSONName := jsonTagName(sf)
				if jsonName == "" {
					jsonName = sourceJSONName
				}

				b := getOrCreateBuilder(builders, reg, typeName, goTypeName)

				if embedType != "" {
					t := derefPtr(sf.Type)
					b.Parent = &parentField{
						SourceFieldName: sf.Name,
						JSONName:        jsonName,
						EmbedType:       embedType,
						Wrap:            t == reflect.TypeFor[model.GraphModelWrapper](),
					}
					continue
				}

				// Multiple source fields can map to the same JSON name
				// (e.g., DNS and Name both map to "ip" in the IP type).
				if idx, ok := b.fieldIdx[jsonName]; ok {
					if !slices.Contains(b.Fields[idx].SourceJSONNames, sourceJSONName) {
						b.Fields[idx].SourceJSONNames = append(b.Fields[idx].SourceJSONNames, sourceJSONName)
					}
					continue
				}

				b.fieldIdx[jsonName] = len(b.Fields)
				b.Fields = append(b.Fields, field{
					SourceFieldName: sf.Name,
					SourceJSONNames: []string{sourceJSONName},
					JSONName:        jsonName,
					GoType:          resolveGoType(sf.Type),
				})
			}
		}
	}

	result := make([]typeSpec, 0, len(builders))
	for _, b := range builders {
		result = append(result, *b)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func markVisited(visited map[string]map[string]bool, typeName, fieldName string) bool {
	if visited[typeName] == nil {
		visited[typeName] = map[string]bool{}
	}
	if visited[typeName][fieldName] {
		return false
	}
	visited[typeName][fieldName] = true
	return true
}

func getOrCreateBuilder(builders map[string]*typeSpec, reg *registry.TypeRegistry, typeName, goTypeName string) *typeSpec {
	if b, ok := builders[typeName]; ok {
		return b
	}
	b := &typeSpec{
		Name:           typeName,
		SourceTypeName: resolveSourceTypeName(reg, typeName, goTypeName),
		fieldIdx:       map[string]int{},
	}
	builders[typeName] = b
	return b
}

// resolveSourceTypeName looks up the Go type name for a capmodel type. Falls back to
// the declaring struct when the type isn't a registered model (e.g., IP → Asset).
func resolveSourceTypeName(reg *registry.TypeRegistry, name, fallback string) string {
	if typ, ok := reg.GetType(strings.ToLower(name)); ok {
		return derefPtr(typ).Name()
	}
	return fallback
}

// sortedRegistryNames returns canonical (non-alias) names in sorted order.
// A name is canonical if it equals strings.ToLower of the Go type name.
func sortedRegistryNames(reg *registry.TypeRegistry) []string {
	var names []string
	for name, typ := range reg.GetAllTypes() {
		if name == strings.ToLower(derefPtr(typ).Name()) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func mustGetType(reg *registry.TypeRegistry, name string) reflect.Type {
	typ, ok := reg.GetType(name)
	if !ok {
		panic("registered type not found: " + name)
	}
	return typ
}

// parseEntry parses a capmodel tag entry: "TypeName[=jsonname[(EmbedType)]]"
func parseEntry(entry string) (typeName, jsonName, embedType string) {
	parts := strings.SplitN(entry, "=", 2)
	typeName = parts[0]
	if len(parts) == 2 {
		rest := parts[1]
		if idx := strings.Index(rest, "("); idx != -1 {
			jsonName = rest[:idx]
			embedType = strings.TrimSuffix(rest[idx+1:], ")")
		} else {
			jsonName = rest
		}
	}
	return
}

// jsonTagName extracts the JSON field name from a struct field's json tag.
// Falls back to the lowercased Go field name when the tag is absent or "-".
func jsonTagName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(sf.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// resolveGoType converts a reflect.Type into the Go source string for the generated code.
// Named types are checked against typeMap before structural resolution so that types like
// SmartBytes emit as "[]byte" rather than "[]uint8". Pointers are excluded from the
// typeMap check so that *SmartBytes correctly recurses to "*[]byte" via the Ptr case.
func resolveGoType(t reflect.Type) string {
	if t.Kind() != reflect.Ptr {
		if mapped, ok := typeMap[t.Name()]; ok {
			return mapped
		}
	}
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + resolveGoType(t.Elem())
	case reflect.Slice:
		return "[]" + resolveGoType(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", resolveGoType(t.Key()), resolveGoType(t.Elem()))
	case reflect.Interface:
		return "any"
	default:
		return t.Name()
	}
}
