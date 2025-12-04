package model

import (
	"maps"
	"slices"
)

type OriginationData struct {
	Capability    []string `neo4j:"capability,omitempty" json:"capability,omitempty" desc:"List of all capabilities that have discovered this asset." example:"[\"amazon\", \"portscan\"]"`
	AttackSurface []string `neo4j:"attackSurface,omitempty" json:"attackSurface,omitempty" desc:"List of attack surface identifiers related to the asset." example:"[\"internal\", \"external\"]"`
	Origins       []string `neo4j:"origins,omitempty" json:"origins,omitempty" desc:"List of originating asset classes for this entity" example:"[\"amazon\", \"ipv4\"]"`
}

func (o *OriginationData) Merge(other OriginationData) {
	if other.Origins != nil {
		o.Origins = other.Origins
	}
	if other.AttackSurface != nil {
		o.AttackSurface = other.AttackSurface
	}
	if other.Capability != nil {
		o.Capability = other.Capability
	}
}

func (o *OriginationData) Visit(other OriginationData) {
	seen := make(map[string]bool)
	for _, s := range append(o.Origins, other.Origins...) {
		seen[s] = true
	}
	o.Origins = slices.Collect(maps.Keys(seen))

	seen = make(map[string]bool)
	for _, s := range append(o.AttackSurface, other.AttackSurface...) {
		seen[s] = true
	}
	o.AttackSurface = slices.Collect(maps.Keys(seen))

	seen = make(map[string]bool)
	for _, s := range append(o.Capability, other.Capability...) {
		seen[s] = true
	}
	o.Capability = slices.Collect(maps.Keys(seen))
}
