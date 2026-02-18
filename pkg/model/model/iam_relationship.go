package model

import (
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&IAMRelationship{})
}

// IAMRelationship is the legacy per-action IAM relationship type.
// Deprecated: Use IAMAWSRelationship which consolidates all actions into a single relationship.
type IAMRelationship struct {
	*BaseRelationship
	Permission string `neo4j:"permission" json:"permission"`
}

// Deprecated: Use NewIAMAWSRelationship which consolidates all actions into a single relationship.
func NewIAMRelationship(source, target GraphModel, label string) *IAMRelationship {
	return &IAMRelationship{
		BaseRelationship: NewBaseRelationship(source, target, label),
		Permission:       label,
	}
}

func (ir *IAMRelationship) Label() string {
	sanitized := specialCharRegex.ReplaceAllString(ir.Permission, "_")
	return strings.ToUpper(sanitized)
}
