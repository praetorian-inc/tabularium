package external

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
)

// Risk is a simplified risk/vulnerability for external tool writers.
type Risk struct {
	Name   string `json:"name"`   // Vulnerability name (e.g., "CVE-2023-1234")
	Status string `json:"status"` // Status code (e.g., "TH", "OH", "OC")
	Target Target `json:"target"` // The target this risk is associated with
}

// ToModel converts to a full Tabularium Risk.
func (r Risk) ToModel() (*model.Risk, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("risk requires name")
	}
	if r.Target == nil {
		return nil, fmt.Errorf("risk requires target")
	}

	target, err := r.Target.ToTarget()
	if err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	status := r.Status
	if status == "" {
		status = model.TriageHigh // Default to "TH" (Triage High)
	}

	risk := model.NewRisk(target, r.Name, status)
	return &risk, nil
}
