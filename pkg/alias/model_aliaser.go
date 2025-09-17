package alias

import "github.com/praetorian-inc/tabularium/pkg/model/filters"

// ModelAliaser enables canonical objects to be resolved from partial information
type ModelAliaser interface {
	// FromAlias returns a single filter to find the canonical object
	//
	// Returns nil if no alias resolution possible
	//
	// The filter MUST return exactly one match - multiple matches indicate non-canonical match
	FromAlias() *filters.Filter
}
