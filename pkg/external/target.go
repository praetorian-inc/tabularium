package external

import (
	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Target is the minimal interface for external targets.
// External types that can be targets (Asset, Port, etc.) implement this interface.
type Target interface {
	// Group returns the grouping identifier (typically DNS/domain).
	Group() string
	// Identifier returns the unique identifier within the group.
	Identifier() string
	// ToTarget converts this external type to a full Tabularium Target.
	ToTarget() (model.Target, error)
}
