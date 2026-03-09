package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Generic struct {
	BaseAsset
	LabelSettableEmbed
}

const GenericLabel = "Generic"

var genericKey = regexp.MustCompile(`^#generic#[^#]+#[^#]+$`)

func init() {
	MustRegisterLabel(GenericLabel)
	registry.Registry.MustRegisterModel(&Generic{})
}

func (g *Generic) GetLabels() []string {
	labels := []string{GenericLabel, AssetLabel, TTLLabel}
	if g.Source == SeedSource {
		labels = append(labels, SeedLabel)
	}
	return labels
}

func (g *Generic) GetClass() string {
	return "generic"
}

func (g *Generic) IsPrivate() bool {
	return false
}

func (g *Generic) Valid() bool {
	return genericKey.MatchString(g.Key)
}

func (g *Generic) GetPartitionKey() string {
	return g.BaseAsset.Identifier
}

func (g *Generic) Merge(o Assetlike) {
	other, ok := o.(*Generic)
	if !ok {
		return
	}
	MergeWithPromotionCheck(&g.BaseAsset, &g.LabelSettableEmbed, other)
}

func (g *Generic) Visit(o Assetlike) {
	other, ok := o.(*Generic)
	if !ok {
		return
	}
	if IsSeedPromotion(&g.BaseAsset, &other.BaseAsset) {
		ApplySeedLabels(&g.BaseAsset, &g.LabelSettableEmbed)
	}
	g.BaseAsset.Visit(other)
}

func (g *Generic) WithStatus(status string) Target {
	ret := *g
	ret.Status = status
	return &ret
}

func (g *Generic) Group() string {
	return g.BaseAsset.Group
}

func (g *Generic) Identifier() string {
	return g.BaseAsset.Identifier
}

func (g *Generic) SetSource(source string) {
	g.BaseAsset.SetSource(source)
	g.Class = g.GetClass()
}

func (g *Generic) SeedModels() []Seedable {
	copy := *g
	return []Seedable{&copy}
}

func (g *Generic) GetDescription() string {
	return "Represents a generic discoverable entity with arbitrary group and identifier strings."
}

func (g *Generic) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if g.Group() == "" {
					return fmt.Errorf("generic asset requires non-empty group")
				}
				if g.Identifier() == "" {
					return fmt.Errorf("generic asset requires non-empty identifier")
				}
				if strings.Contains(g.Group(), "#") {
					return fmt.Errorf("generic asset group must not contain '#'")
				}
				if strings.Contains(g.Identifier(), "#") {
					return fmt.Errorf("generic asset identifier must not contain '#'")
				}
				g.Key = fmt.Sprintf("#generic#%s#%s", g.Group(), g.Identifier())
				g.Class = g.GetClass()
				return nil
			},
		},
	}
}

func NewGeneric(group, identifier string) Generic {
	g := Generic{}
	g.BaseAsset.Group = group
	g.BaseAsset.Identifier = identifier

	g.Defaulted()
	registry.CallHooks(&g)

	return g
}

func NewGenericSeed(name string) Generic {
	g := NewGeneric(name, name)
	g.Source = SeedSource
	g.Status = Pending
	g.TTL = 0
	return g
}
