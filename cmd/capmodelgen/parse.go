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

// parentKind describes how a parent field is wired to the child model.
type parentKind string

const (
	parentInject    parentKind = "inject"    // GraphModelWrapper: wrap with NewGraphModelWrapper, set before hooks
	parentConcrete  parentKind = "concrete"  // concrete pointer: set directly, before hooks
	parentInterface parentKind = "interface" // interface field: set after hooks
)

// field describes a regular data field in a slim type.
type field struct {
	SourceFieldName string   // Go field name in the source model
	SourceJSONNames []string // json names in the source model (multiple when merged)
	JSONName        string   // json name for the slim type field
	GoType          string   // Go type string for generated struct
}

// parentField describes a parent/embed relationship field.
type parentField struct {
	SourceFieldName string     // Go field name in the source model (e.g., "Parent", "Target")
	JSONName        string     // json name for the slim struct
	SlimType        string     // slim type name (e.g., "Asset", "WebApplication")
	Kind            parentKind // how the parent is wired
}

// slimType describes a generated slim type.
type slimType struct {
	Name           string
	SourceTypeName string // Go type name (e.g., "Asset", "Risk")
	Fields         []field
	Parent         *parentField
}

// parseSlimTags walks all registered types and extracts slim tag information.
func parseSlimTags(reg *registry.TypeRegistry) []slimType {
	builders := map[string]*slimType{}
	// Tracks which (slimType, fieldIndex) pairs we've seen to avoid
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

				// Dedup: skip fields seen from a different registered type
				// (same embedded field reached via different type walks).
				if seen[typeName] == nil {
					seen[typeName] = map[string]bool{}
				}
				if seen[typeName][sf.Name] {
					continue
				}
				seen[typeName][sf.Name] = true

				sourceJSONName := jsonTagName(sf)
				if jsonName == "" {
					jsonName = sourceJSONName
				}

				b, ok := builders[typeName]
				if !ok {
					b = &slimType{
						Name:           typeName,
						SourceTypeName: resolveSourceTypeName(reg, typeName, goTypeName),
					}
					builders[typeName] = b
				}

				if embedType != "" {
					b.Parent = &parentField{
						SourceFieldName: sf.Name,
						JSONName:        jsonName,
						SlimType:        embedType,
						Kind:            resolveParentKind(sf.Type),
					}
					continue
				}

				// If another source field maps to the same slim JSON name,
				// merge it (both source fields get set from one slim field).
				if mergeField(b, jsonName, sourceJSONName) {
					continue
				}

				b.Fields = append(b.Fields, field{
					SourceFieldName: sf.Name,
					SourceJSONNames: []string{sourceJSONName},
					JSONName:        jsonName,
					GoType:          resolveGoType(sf.Type),
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

// resolveParentKind determines the parent wiring strategy from the field's reflect.Type.
func resolveParentKind(t reflect.Type) parentKind {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch {
	case t.Kind() == reflect.Interface:
		return parentInterface
	case t.Name() == "GraphModelWrapper":
		return parentInject
	default:
		return parentConcrete
	}
}

// resolveSourceTypeName determines the Go type name for a slim type.
// If the slim type name matches a registered model, use that model's Go type name;
// otherwise fall back to the struct where the tags were found.
func resolveSourceTypeName(reg *registry.TypeRegistry, slimTypeName, fallback string) string {
	if typ, ok := reg.GetType(strings.ToLower(slimTypeName)); ok {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		return typ.Name()
	}
	return fallback
}

// mergeField checks if a field with the same jsonName already exists in the builder.
// If so, appends the sourceJSONName to that field's list (if not already present) and returns true.
func mergeField(b *slimType, jsonName, sourceJSONName string) bool {
	for i, existing := range b.Fields {
		if existing.JSONName == jsonName {
			for _, name := range b.Fields[i].SourceJSONNames {
				if name == sourceJSONName {
					return true
				}
			}
			b.Fields[i].SourceJSONNames = append(b.Fields[i].SourceJSONNames, sourceJSONName)
			return true
		}
	}
	return false
}

// sortedRegistryNames returns canonical (non-alias) registry names in sorted order.
// A name is canonical if it equals strings.ToLower of the Go type name.
func sortedRegistryNames(reg *registry.TypeRegistry) []string {
	var names []string
	for name, typ := range reg.GetAllTypes() {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if name == strings.ToLower(typ.Name()) {
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
