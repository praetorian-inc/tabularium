package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const IAMAWSPermissionLabel = "IAM_AWS_PERMISSION"

func init() {
	registry.Registry.MustRegisterModel(&IAMAWSRelationship{}, IAMAWSPermissionLabel)
}

type IAMAWSRelationship struct {
	*BaseRelationship
	Actions []string `neo4j:"actions" json:"actions"`
}

func NewIAMAWSRelationship(source, target GraphModel, actions []string) *IAMAWSRelationship {
	return &IAMAWSRelationship{
		BaseRelationship: NewBaseRelationship(source, target, IAMAWSPermissionLabel),
		Actions:          actions,
	}
}

func (r *IAMAWSRelationship) Label() string {
	return IAMAWSPermissionLabel
}

func (r *IAMAWSRelationship) GetDescription() string {
	return "Represents an aggregated AWS IAM permission relationship between a principal and a resource, containing all allowed actions."
}

func (r *IAMAWSRelationship) Visit(o GraphRelationship) {
	other, ok := o.(*IAMAWSRelationship)
	if !ok {
		return
	}

	r.Actions = mergeStringSlices(r.Actions, other.Actions)
	r.BaseRelationship.Visit(other)
}

// mergeStringSlices returns the union of two string slices, preserving order.
func mergeStringSlices(existing, incoming []string) []string {
	seen := make(map[string]struct{}, len(existing))
	for _, s := range existing {
		seen[s] = struct{}{}
	}
	for _, s := range incoming {
		if _, ok := seen[s]; !ok {
			existing = append(existing, s)
			seen[s] = struct{}{}
		}
	}
	return existing
}
