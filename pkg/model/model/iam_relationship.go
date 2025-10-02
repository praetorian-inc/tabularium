package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
	"strings"
)

func init() {
	registry.Registry.MustRegisterModel(&IamRelationship{})
}

type IamRelationship struct {
	*BaseRelationship
	Permission string `neo4j:"permission" json:"permission"`
}

func NewIamRelationship(source, target GraphModel, label string) *IamRelationship {
	return &IamRelationship{
		BaseRelationship: NewBaseRelationship(source, target, label),
		Permission:       label,
	}
}

func (ir *IamRelationship) Label() string {
	sanitized := specialCharRegex.ReplaceAllString(ir.Permission, "_")
	return strings.ToUpper(sanitized)
}
