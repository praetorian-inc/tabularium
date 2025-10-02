package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementation for testing
type mockRelabelableModel struct {
	BaseAsset
	LabelSettableEmbed
}

func (m *mockRelabelableModel) GetLabels() []string {
	return []string{AssetLabel}
}

func (m *mockRelabelableModel) Valid() bool {
	return true
}

func TestLabelSettableEmbed_GetPendingLabelAddition(t *testing.T) {
	tests := []struct {
		name           string
		pendingLabel   string
		expectedResult string
	}{
		{
			name:           "No pending label addition",
			pendingLabel:   "",
			expectedResult: NO_PENDING_LABEL_ADDITION,
		},
		{
			name:           "Seed label addition pending",
			pendingLabel:   SeedLabel,
			expectedResult: SeedLabel,
		},
		{
			name:           "Verified label addition pending",
			pendingLabel:   "Verified",
			expectedResult: "Verified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embed := LabelSettableEmbed{
				PendingLabelAddition: tt.pendingLabel,
			}
			assert.Equal(t, tt.expectedResult, embed.GetPendingLabelAddition())
		})
	}
}

func TestHasPendingLabelAddition(t *testing.T) {
	tests := []struct {
		name             string
		model            GraphModel
		expectedResult   string
		expectedRequired bool
	}{
		{
			name: "Model with pending label addition",
			model: &mockRelabelableModel{
				BaseAsset: BaseAsset{Key: "test-key"},
				LabelSettableEmbed: LabelSettableEmbed{
					PendingLabelAddition: SeedLabel,
				},
			},
			expectedResult:   SeedLabel,
			expectedRequired: true,
		},
		{
			name: "Model without pending label addition",
			model: &mockRelabelableModel{
				BaseAsset: BaseAsset{Key: "test-key"},
				LabelSettableEmbed: LabelSettableEmbed{
					PendingLabelAddition: NO_PENDING_LABEL_ADDITION,
				},
			},
			expectedResult:   NO_PENDING_LABEL_ADDITION,
			expectedRequired: false,
		},
		{
			name: "Non-relabelable model (Asset without pending label addition)",
			model: &Asset{
				BaseAsset: BaseAsset{
					Key: "test-key",
				},
			},
			expectedResult:   NO_PENDING_LABEL_ADDITION,
			expectedRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, required := PendingLabelAddition(tt.model)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedRequired, required)
		})
	}
}

func TestLabelSettableInterface(t *testing.T) {
	var _ LabelSettable = &mockRelabelableModel{}

	model := &mockRelabelableModel{
		BaseAsset: BaseAsset{Key: "test-key"},
		LabelSettableEmbed: LabelSettableEmbed{
			PendingLabelAddition: SeedLabel,
		},
	}

	assert.Equal(t, SeedLabel, model.GetPendingLabelAddition())
	assert.Equal(t, "test-key", model.GetKey())
	assert.True(t, model.Valid())
}
