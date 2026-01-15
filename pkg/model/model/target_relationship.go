package model

import (
	"fmt"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// TargetRelationship wraps a GraphRelationship to implement the Target interface,
// enabling capabilities to match on relationships (e.g., HasTechnology).
type TargetRelationship struct {
	registry.BaseModel
	Relationship   GraphRelationshipWrapper `json:"relationship"`
	StatusOverride string                   `json:"status_override,omitempty"`
}

func init() {
	registry.Registry.MustRegisterModel(&TargetRelationship{})
}

// NewTargetRelationship creates a new TargetRelationship wrapping the given relationship.
func NewTargetRelationship(rel GraphRelationship) *TargetRelationship {
	return &TargetRelationship{
		Relationship: NewGraphRelationshipWrapper(rel),
	}
}

// Nodes returns the source and target models for the relationship.
func (tr *TargetRelationship) Nodes() (GraphModel, GraphModel) {
	if tr.Relationship.Model == nil {
		return nil, nil
	}
	return tr.Relationship.Model.Nodes()
}

// Label returns the relationship label (e.g., "HAS_TECHNOLOGY").
func (tr *TargetRelationship) Label() string {
	if tr.Relationship.Model == nil {
		return ""
	}
	return tr.Relationship.Model.Label()
}

// --- GraphModel interface ---

func (tr *TargetRelationship) GetKey() string {
	if tr.Relationship.Model == nil {
		return ""
	}
	return tr.Relationship.Model.GetKey()
}

func (tr *TargetRelationship) GetLabels() []string {
	return []string{"TargetRelationship"}
}

func (tr *TargetRelationship) Valid() bool {
	return tr.Relationship.Model != nil && tr.Relationship.Model.Valid()
}

// --- Target interface ---

func (tr *TargetRelationship) GetStatus() string {
	if tr.StatusOverride != "" {
		return tr.StatusOverride
	}
	if tr.Relationship.Model == nil {
		return Active
	}
	source, _ := tr.Relationship.Model.Nodes()
	if t, ok := source.(Target); ok {
		return t.GetStatus()
	}
	return Active
}

func (tr *TargetRelationship) WithStatus(status string) Target {
	ret := *tr
	ret.StatusOverride = status
	return &ret
}

func (tr *TargetRelationship) Group() string {
	if tr.Relationship.Model == nil {
		return ""
	}
	source, target := tr.Relationship.Model.Nodes()
	sourceGroup, targetGroup := "", ""
	if t, ok := source.(Target); ok {
		sourceGroup = t.Group()
	}
	if t, ok := target.(Target); ok {
		targetGroup = t.Group()
	}
	return fmt.Sprintf("%s#%s", sourceGroup, targetGroup)
}

func (tr *TargetRelationship) Identifier() string {
	if tr.Relationship.Model == nil {
		return ""
	}
	return tr.Relationship.Model.GetKey()
}

func (tr *TargetRelationship) IsStatus(status string) bool {
	return tr.GetStatus() == status
}

func (tr *TargetRelationship) IsClass(class string) bool {
	if tr.Relationship.Model == nil {
		return false
	}
	return strings.EqualFold(tr.Relationship.Model.Label(), class)
}

func (tr *TargetRelationship) IsPrivate() bool {
	if tr.Relationship.Model == nil {
		return false
	}
	source, target := tr.Relationship.Model.Nodes()

	if t, ok := source.(Target); ok && t.IsPrivate() {
		return true
	}
	if t, ok := target.(Target); ok && t.IsPrivate() {
		return true
	}
	return false
}

func (tr *TargetRelationship) GetDescription() string {
	return "Wraps a graph relationship to enable capability matching on relationships."
}
