package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

type Assetlike interface {
	Target
	Merge(Assetlike)
	Visit(Assetlike)
	GetMetadata() *Metadata
	Attribute(string, string) Attribute
	GetBase() *BaseAsset
	GetClass() string
	GetSource() string
	SetSource(string)
	SetStatus(string)
}

func useGroupAndIdentifier(asset Assetlike, groupDst, identifierDst *string) registry.Hook {
	return registry.Hook{
		Call: func() error {
			copyGroupAndIdentifier(asset, groupDst, identifierDst)
			reuseGroupOrIdentifierIfSeed(asset, groupDst, identifierDst)
			copyGroupAndIdentifier(asset, groupDst, identifierDst) // in case the reuse function changed the values
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

func copyGroupAndIdentifier(asset Assetlike, groupDst, identifierDst *string) {
	group := asset.GetBase().Group
	if group != "" {
		*groupDst = group
	}

	identifier := asset.GetBase().Identifier
	if identifier != "" {
		*identifierDst = identifier
	}
}

// to maintain backwards compatibility with the old seeding logic, we need
// to accept seeds that only specify the group (dns) or identifier (IP).
// we do not want to do this for non-seeds
func reuseGroupOrIdentifierIfSeed(a Assetlike, groupDst, identifierDst *string) {
	if a.GetSource() != SeedSource {
		return
	}
	if *groupDst == "" && *identifierDst != "" {
		*groupDst = *identifierDst
	}
	if *identifierDst == "" && *groupDst != "" {
		*identifierDst = *groupDst
	}
}
