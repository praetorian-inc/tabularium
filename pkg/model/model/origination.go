package model

import (
	"maps"
	"slices"
)

type OriginationData struct {
	Capability    []string `neo4j:"capability,omitempty" json:"capability,omitempty" desc:"List of all capabilities that have discovered this asset." example:"[\"amazon\", \"portscan\"]"`
	AttackSurface []string `neo4j:"attackSurface,omitempty" json:"attackSurface,omitempty" desc:"List of attack surface identifiers related to the asset." example:"[\"internal\", \"external\"]"`
	Origins       []string `neo4j:"origins,omitempty" json:"origins,omitempty" desc:"List of originating asset classes for this entity" example:"[\"amazon\", \"ipv4\"]"`
	IsExternal    bool     `neo4j:"isExternal" json:"isExternal" desc:"Boolean flag indicating if asset is external"`
	IsInternal    bool     `neo4j:"isInternal" json:"isInternal" desc:"Boolean flag indicating if asset is internal"`
	IsCloud       bool     `neo4j:"isCloud" json:"isCloud" desc:"Boolean flag indicating if asset is cloud"`
	IsApplication bool     `neo4j:"isApplication" json:"isApplication" desc:"Boolean flag indicating if asset is application"`
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
	// Re-derive boolean flags after merge
	DeriveAttackSurfaceFlags(o)
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

	// Re-derive boolean flags after visit
	DeriveAttackSurfaceFlags(o)
}

// DeriveAttackSurfaceFlags populates boolean index fields from the AttackSurface slice.
// Must be called whenever AttackSurface is set.
func DeriveAttackSurfaceFlags(base *OriginationData) {
	// Reset all flags first for idempotency
	base.IsExternal = false
	base.IsInternal = false
	base.IsCloud = false
	base.IsApplication = false

	for _, s := range base.AttackSurface {
		switch s {
		case "external":
			base.IsExternal = true
		case "internal":
			base.IsInternal = true
		case "cloud":
			base.IsCloud = true
		case "application":
			base.IsExternal = true // application also means external
			base.IsApplication = true
		}
	}
}
