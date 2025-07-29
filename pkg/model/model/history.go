package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type HistoryRecord struct {
	registry.BaseModel
	From               string   `json:"from,omitempty" neo4j:"from,omitempty" desc:"The previous state or value." example:"TL"`
	To                 string   `json:"to,omitempty" neo4j:"to,omitempty" desc:"The new state or value." example:"OL"`
	By                 string   `json:"by,omitempty" neo4j:"by,omitempty" desc:"Identifier of the user or system that made the change." example:"user@example.com"`
	Comment            string   `json:"comment,omitempty" neo4j:"comment,omitempty" desc:"Comment associated with the history event." example:"Asset confirmed via scan."`
	Updated            string   `json:"updated,omitempty" neo4j:"updated,omitempty" desc:"Timestamp of the history event (RFC3339)." example:"2023-10-27T11:05:00Z"`
	Model              *string  `json:"base,omitempty" neo4j:"base,omitempty" desc:"Identifier of the ML model used for the prediction, if applicable." example:"affiliation-model-v1.2"`
	Logit              *float32 `json:"logit,omitempty" neo4j:"logit,omitempty" desc:"Logit value from an ML model prediction, if applicable." example:"0.85"`
	AffiliationVerdict string   `json:"affiliationVerdict,omitempty" neo4j:"affiliationVerdict,omitempty" desc:"Affiliation verdict from an ML model, if applicable." example:"Affiliated"`
	FilePath           string   `json:"filePath,omitempty" neo4j:"filePath,omitempty" desc:"Path to a related file, if applicable." example:"proofs/evidence.png"`
}

func init() {
	registry.Registry.MustRegisterModel(&HistoryRecord{})
	registry.Registry.MustRegisterModel(&History{})
}

type History struct {
	registry.BaseModel
	History []HistoryRecord `neo4j:"history" json:"history,omitempty" desc:"List of history records detailing changes."`
	// Remove is used internally for managing history edits via API/UI, not persisted.
	Remove *int `neo4j:"-" json:"remove,omitempty" desc:"Index of the history record to remove (used for updates, not stored)." example:"0"`
}

func (h *History) Update(from, to, by, comment string, other History) bool {
	if other.Remove != nil && *other.Remove < len(h.History) {
		i := *other.Remove
		h.History[i].Comment = ""
		if h.History[i].To == "" {
			h.History = append(h.History[:i], h.History[i+1:]...)
		}
		return false
	}
	if to != "" && from != to {
		event := HistoryRecord{
			From:    from,
			To:      to,
			By:      by,
			Comment: comment,
			Updated: Now(),
		}
		h.History = append(h.History, event)
		return true
	} else if comment != "" {
		h.History = append(h.History, HistoryRecord{
			By:      by,
			Comment: comment,
			Updated: Now(),
		})
	}
	return false
}

func (h *History) AddAutoTriageEntry(recommendation string, logit *float32, model *string) {
	h.History = append(h.History, HistoryRecord{
		By:      "Praetorian AI",
		Comment: fmt.Sprintf("ML Model Prediction: %s", recommendation),
		Logit:   logit,
		Model:   model,
		Updated: Now(),
	})
}

func (h *History) AddAffiliationEntry(verdict string, filePath string) {
	h.History = append(h.History, HistoryRecord{
		By:                 "Praetorian AI",
		Comment:            fmt.Sprintf("ML Model Prediction: %s", verdict),
		AffiliationVerdict: verdict,
		FilePath:           filePath,
		Updated:            Now(),
	})
}

func CreateTestHistory(by string, comment string, updated string, verdict string) History {
	return History{
		History: []HistoryRecord{
			{
				By:                 by,
				Comment:            comment,
				Updated:            updated,
				AffiliationVerdict: verdict,
			},
		},
	}
}

// GetDescription returns a description for the History model.
func (h *History) GetDescription() string {
	return "Represents a container for a list of historical event records."
}

// GetDescription returns a description for the HistoryRecord model.
func (hr *HistoryRecord) GetDescription() string {
	return "Represents a single event or change in the history of an entity."
}
