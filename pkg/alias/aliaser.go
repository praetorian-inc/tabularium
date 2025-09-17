package alias

import "github.com/praetorian-inc/tabularium/pkg/model/filters"

// Aliaser is an interface for objects that can provide alternative filters to fetch themselves
// from the database when the primary key information is incomplete.
//
// Objects implementing this interface can return a set of filters that can be used to
// query for the correct object using alternative identifiers. This is useful when tools
// provide partial information that doesn't match the object's key construction pattern.
//
// For example, an ADObject uses domain + ObjectID for its key, but many tools only
// provide the DistinguishedName. The ADObject can implement FromAlias to return a
// filter for DistinguishedName, allowing the system to fetch the correct object.
type Aliaser interface {
	// FromAlias returns one or more filters that can be used to fetch the correct
	// object from the database when the primary key cannot be determined.
	// The caller should iterate through the filters until a match is found.
	// If no filters are applicable, return nil or an empty slice.
	FromAlias() []filters.Filter
}