package alias

import "github.com/praetorian-inc/tabularium/pkg/model/filters"

// Aliaser enables objects to be resolved from partial information
type Aliaser interface {
	// FromAlias returns a single filter to find the canonical object
	// Returns nil if no alias resolution possible
	// MUST return exactly one match - multiple matches indicate error
	FromAlias() *filters.Filter
}
