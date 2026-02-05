package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// ADObject is a simplified Active Directory object for external tool writers.
// It contains only the essential fields needed to identify an AD object.
type ADObject struct {
	Label             string `json:"label"`             // Primary label (ADUser, ADComputer, ADGroup, etc.)
	Domain            string `json:"domain"`            // AD domain
	ObjectID          string `json:"objectid"`          // Object identifier (SID or GUID)
	DistinguishedName string `json:"distinguishedname"` // DN path
}

// Group implements Target interface.
func (a ADObject) Group() string { return a.Domain }

// Identifier implements Target interface.
func (a ADObject) Identifier() string { return a.ObjectID }

// ToTarget converts to a full Tabularium ADObject.
func (a ADObject) ToTarget() (model.Target, error) {
	if a.Domain == "" {
		return nil, fmt.Errorf("adobject requires domain")
	}
	if a.ObjectID == "" {
		return nil, fmt.Errorf("adobject requires objectid")
	}

	label := a.Label
	if label == "" {
		label = model.ADObjectLabel
	}

	adObject := model.NewADObject(a.Domain, a.ObjectID, a.DistinguishedName, label)
	return &adObject, nil
}

// ToModel converts to a full Tabularium ADObject (convenience method).
func (a ADObject) ToModel() (*model.ADObject, error) {
	target, err := a.ToTarget()
	if err != nil {
		return nil, err
	}
	return target.(*model.ADObject), nil
}

// ADObjectFromModel converts a Tabularium ADObject to an external ADObject.
func ADObjectFromModel(m *model.ADObject) ADObject {
	return ADObject{
		Label:             m.Label,
		Domain:            m.Domain,
		ObjectID:          m.ObjectID,
		DistinguishedName: m.DistinguishedName,
	}
}
