package main

import (
	"fmt"
	"reflect"
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
	SourceFieldName  string   // Go field name in the source model (first field if merged)
	SourceJSONNames  []string // json tag names in the source model (may be multiple if merged)
	JSONName         string   // json tag name for the slim type (may differ if renamed)
	GoType           string   // Go type string for the slim type
	EmbedSlimType    string   // if non-empty, this field is an embedded slim type
	IsPointer        bool     // whether the source field is a pointer type
}

// slimType describes a generated slim type.
type slimType struct {
	Name            string
	SourceModelName string // registered name in the registry (lowercase)
	SourceTypeName  string // Go type name (e.g., "Asset", "Risk")
	Fields          []fieldSpec
}

// parentKind classifies how the parent field is typed.
type parentKind int

const (
	parentInject   parentKind = iota // GraphModelWrapper — inject via JSON
	parentNoInject                   // *WebApplication — direct pointer, no injection
)

// parentInfo holds metadata about the parent field for a slim type.
type parentInfo struct {
	Kind           parentKind
	SourceType     string // e.g., "GraphModelWrapper" or "*WebApplication"
	EmbedSlimType  string // e.g., "Asset" or "WebApplication"
	SourceField    string // e.g., "Parent" or "Target"
	InterfaceField bool   // true if the source field is an interface (e.g., Target)
}

// slimTypeSourceOverrides maps slim type names to their desired source model.
// This handles cases where tagged fields live in an embedded struct (e.g., CloudResource)
// but the slim type should convert to a specific subtype.
var slimTypeSourceOverrides = map[string]string{
	"AWSResource":   "AWSResource",
	"AzureResource": "AzureResource",
	"GCPResource":   "GCPResource",
	"CloudResource": "CloudResource",
}

// parseSlimTags walks all registered types and extracts slim tag information.
func parseSlimTags(reg *registry.TypeRegistry) ([]slimType, error) {
	builders := map[string]*slimType{}
	// Track seen fields per slim type to avoid duplicates from aliases/embedded structs
	seen := map[string]map[string]bool{} // slimTypeName → set of fieldNames

	// Collect the canonical names (skip aliases)
	canonicalTypes := map[string]reflect.Type{}
	for name, typ := range reg.GetAllTypes() {
		aliases := reg.GetAliases(name)
		// If this name is an alias (not the primary), skip it
		if len(aliases) > 1 && aliases[0] != name {
			continue
		}
		canonicalTypes[name] = typ
	}

	for _, typ := range canonicalTypes {
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

			entries := strings.Split(slimTag, ",")
			for _, entry := range entries {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}

				typeName, jsonName, embedType := parseEntry(entry)

				// Deduplicate: skip if we've already added this field for this slim type
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

				goType := resolveGoType(sf.Type)

				b, ok := builders[typeName]
				if !ok {
					sourceName := goTypeName
					sourceModelName := strings.ToLower(goTypeName)
					if override, hasOverride := slimTypeSourceOverrides[typeName]; hasOverride {
						sourceName = override
						sourceModelName = strings.ToLower(override)
					}
					b = &slimType{
						Name:            typeName,
						SourceModelName: sourceModelName,
						SourceTypeName:  sourceName,
					}
					builders[typeName] = b
				}

				// Check if there's already a field with the same jsonName (merge case)
				merged := false
				if embedType == "" {
					for i, existing := range b.Fields {
						if existing.JSONName == jsonName && existing.EmbedSlimType == "" {
							// Merge: add source json name to the list
							b.Fields[i].SourceJSONNames = append(b.Fields[i].SourceJSONNames, sourceJSONName)
							merged = true
							break
						}
					}
				}

				if !merged {
					b.Fields = append(b.Fields, fieldSpec{
						SourceFieldName: sf.Name,
						SourceJSONNames: []string{sourceJSONName},
						JSONName:        jsonName,
						GoType:          goType,
						EmbedSlimType:   embedType,
						IsPointer:       sf.Type.Kind() == reflect.Ptr,
					})
				}
			}
		}
	}

	result := make([]slimType, 0, len(builders))
	for _, b := range builders {
		result = append(result, *b)
	}

	return result, nil
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

// getParentInfo returns metadata about the parent/target field of a slim type.
func getParentInfo(st slimType) *parentInfo {
	for _, f := range st.Fields {
		if f.EmbedSlimType == "" {
			continue
		}
		switch f.GoType {
		case "GraphModelWrapper":
			return &parentInfo{
				Kind:          parentInject,
				SourceType:    f.GoType,
				EmbedSlimType: f.EmbedSlimType,
				SourceField:   f.SourceFieldName,
			}
		default:
			info := &parentInfo{
				Kind:          parentNoInject,
				SourceType:    f.GoType,
				EmbedSlimType: f.EmbedSlimType,
				SourceField:   f.SourceFieldName,
			}
			if f.GoType == "any" { // interface field (e.g., Target)
				info.InterfaceField = true
			}
			return info
		}
	}
	return nil
}
