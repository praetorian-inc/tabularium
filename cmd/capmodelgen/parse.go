package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// typeMap maps named types to their simplified Go representations.
var typeMap = map[string]string{
	"SmartBytes":        "[]byte",
	"CloudResourceType": "string",
	"GobSafeBool":       "bool",
}

// fieldSpec describes a single field in a slim type.
type fieldSpec struct {
	SourceFieldName string   // Go field name in the source model
	SourceJSONNames []string // json names in the source model (multiple when merged)
	JSONName        string   // json name for the slim type field
	GoType          string   // Go type string
	EmbedSlimType   string   // if non-empty, this is an embedded slim type
}

// slimType describes a generated slim type.
type slimType struct {
	Name            string
	SourceModelName string // registered name in the registry (lowercase)
	SourceTypeName  string // Go type name (e.g., "Asset", "Risk")
	Fields          []fieldSpec
}

// parseSlimTags walks all registered types and extracts slim tag information.
func parseSlimTags(reg *registry.TypeRegistry) []slimType {
	builders := map[string]*slimType{}
	// Tracks which (slimType, field:jsonName) pairs we've seen to avoid
	// duplicates from embedded struct traversal.
	seen := map[string]map[string]bool{}

	for _, name := range sortedRegistryNames(reg) {
		typ := mustGetType(reg, name)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		goTypeName := typ.Name()

		for _, sf := range reflect.VisibleFields(typ) {
			if !sf.IsExported() {
				continue
			}
			slimTag := sf.Tag.Get("slim")
			if slimTag == "" {
				continue
			}

			for _, entry := range strings.Split(slimTag, ",") {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}

				typeName, jsonName, embedType := parseEntry(entry)

				// Dedup: skip fields already seen for this slim type
				if seen[typeName] == nil {
					seen[typeName] = map[string]bool{}
				}
				fieldKey := sf.Name + ":" + jsonName
				if seen[typeName][fieldKey] {
					continue
				}
				seen[typeName][fieldKey] = true

				sourceJSONName := jsonTagName(sf)
				if jsonName == "" {
					jsonName = sourceJSONName
				}

				b, ok := builders[typeName]
				if !ok {
					srcType, srcModel := resolveSourceModel(reg, typeName, goTypeName)
					b = &slimType{
						Name:            typeName,
						SourceModelName: srcModel,
						SourceTypeName:  srcType,
					}
					builders[typeName] = b
				}

				// If another source field maps to the same slim JSON name,
				// merge it (both source fields get set from one slim field).
				if embedType == "" {
					if merged := mergeField(b, jsonName, sourceJSONName); merged {
						continue
					}
				}

				b.Fields = append(b.Fields, fieldSpec{
					SourceFieldName: sf.Name,
					SourceJSONNames: []string{sourceJSONName},
					JSONName:        jsonName,
					GoType:          resolveGoType(sf.Type),
					EmbedSlimType:   embedType,
				})
			}
		}
	}

	result := make([]slimType, 0, len(builders))
	for _, b := range builders {
		result = append(result, *b)
	}
	return result
}

// resolveSourceModel determines the Go type name and registry name for a slim type.
// If the slim type name is itself a registered model, use that; otherwise fall back
// to the struct where the tags were found.
func resolveSourceModel(reg *registry.TypeRegistry, slimTypeName, fallbackGoType string) (typeName, modelName string) {
	lower := strings.ToLower(slimTypeName)
	if typ, ok := reg.GetType(lower); ok {
		t := typ
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t.Name(), lower
	}
	return fallbackGoType, strings.ToLower(fallbackGoType)
}

// mergeField checks if a field with the same jsonName already exists in the builder.
// If so, appends the sourceJSONName to that field's list and returns true.
func mergeField(b *slimType, jsonName, sourceJSONName string) bool {
	for i, existing := range b.Fields {
		if existing.JSONName == jsonName && existing.EmbedSlimType == "" {
			b.Fields[i].SourceJSONNames = append(b.Fields[i].SourceJSONNames, sourceJSONName)
			return true
		}
	}
	return false
}

// sortedRegistryNames returns canonical (non-alias) registry names in sorted order.
// Aliases are detected by checking if another name resolves to the same reflect.Type
// with a different primary name.
func sortedRegistryNames(reg *registry.TypeRegistry) []string {
	// Build set of primary names: a name is primary if it equals strings.ToLower(Type.Name())
	allTypes := reg.GetAllTypes()
	var names []string
	seen := map[reflect.Type]bool{}
	for name, typ := range allTypes {
		elemType := typ
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		// A name is canonical if it matches the lowercased Go type name
		if name == strings.ToLower(elemType.Name()) {
			if !seen[typ] {
				seen[typ] = true
				names = append(names, name)
			}
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

// parseEntry parses a single slim tag entry: "TypeName[=jsonname[(EmbedType)]]"
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

// jsonTagName extracts the json name from a struct field tag.
func jsonTagName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(sf.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// resolveGoType converts a reflect.Type to a Go type string, applying typeMap substitutions.
func resolveGoType(t reflect.Type) string {
	if t.Kind() != reflect.Ptr && t.Name() != "" {
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
