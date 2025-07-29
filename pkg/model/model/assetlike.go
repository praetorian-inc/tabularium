package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

type Assetlike interface {
	GraphModel
	Target
	Merge(Assetlike)
	Visit(Assetlike)
	GetMetadata() *Metadata
	Attribute(string, string) Attribute
	GetBase() *BaseAsset
	SetSource(string)
	SetStatus(string)
	Seed() Seed
}

func useGroupAndIdentifier(asset Assetlike, groupDst, identifierDst *string) registry.Hook {
	return registry.Hook{
		Call: func() error {
			group := asset.GetBase().Group
			if group != "" {
				*groupDst = group
			}

			identifier := asset.GetBase().Identifier
			if identifier != "" {
				*identifierDst = identifier
			}
			return nil
		},
	}
}

func setGroupAndIdentifier(asset Assetlike, groupDst, identifierDst *string) registry.Hook {
	return registry.Hook{
		Call: func() error {
			asset.GetBase().Group = *groupDst
			asset.GetBase().Identifier = *identifierDst
			return nil
		},
	}
}
