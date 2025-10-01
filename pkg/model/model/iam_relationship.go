package model

import "strings"

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
	return strings.ToUpper(neo4jNegateLabelRegex.ReplaceAllString(ir.Permission, "_"))
}
