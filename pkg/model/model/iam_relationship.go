package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
	"strings"
)

func init() {
	registry.Registry.MustRegisterModel(&IAMRelationship{})
}

type IAMRelationship struct {
	*BaseRelationship
	Permission string `neo4j:"permission" json:"permission"`
}

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
