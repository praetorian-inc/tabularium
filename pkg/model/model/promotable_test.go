package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementation for testing
type mockPromotableModel struct {
	BaseAsset
	PromotableEmbed
}

func (m *mockPromotableModel) GetLabels() []string {
	return []string{AssetLabel}
}

func (m *mockPromotableModel) Valid() bool {
	return true
}

func TestPromotableEmbed_GetPendingPromotion(t *testing.T) {
	tests := []struct {
		name           string
		pendingLabel   string
		expectedResult string
	}{
		{
			name:           "No pending promotion",
			pendingLabel:   "",
			expectedResult: NO_PENDING_PROMOTION,
		},
		{
			name:           "Seed label promotion pending",
			pendingLabel:   SeedLabel,
			expectedResult: SeedLabel,
		},
		{
			name:           "Verified label promotion pending",
			pendingLabel:   "Verified",
			expectedResult: "Verified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embed := PromotableEmbed{
				pendingPromotion: tt.pendingLabel,
			}
			assert.Equal(t, tt.expectedResult, embed.GetPendingPromotion())
		})
	}
}

func TestHasPendingPromotion(t *testing.T) {
	tests := []struct {
		name             string
		model            GraphModel
		expectedResult   string
		expectedRequired bool
	}{
		{
			name: "Model with pending promotion",
			model: &mockPromotableModel{
				BaseAsset: BaseAsset{Key: "test-key"},
				PromotableEmbed: PromotableEmbed{
					pendingPromotion: SeedLabel,
				},
			},
			expectedResult:   SeedLabel,
			expectedRequired: true,
		},
		{
			name: "Model without pending promotion",
			model: &mockPromotableModel{
				BaseAsset: BaseAsset{Key: "test-key"},
				PromotableEmbed: PromotableEmbed{
					pendingPromotion: NO_PENDING_PROMOTION,
				},
			},
			expectedResult:   NO_PENDING_PROMOTION,
			expectedRequired: false,
		},
		{
			name: "Non-promotable model (Asset without promotion)",
			model: &Asset{
				BaseAsset: BaseAsset{
					Key: "test-key",
				},
			},
			expectedResult:   NO_PENDING_PROMOTION,
			expectedRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, required := PendingPromotion(tt.model)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedRequired, required)
		})
	}
}

func TestPromotableInterface(t *testing.T) {
	var _ Promotable = &mockPromotableModel{}

	model := &mockPromotableModel{
		BaseAsset: BaseAsset{Key: "test-key"},
		PromotableEmbed: PromotableEmbed{
			pendingPromotion: SeedLabel,
		},
	}

	assert.Equal(t, SeedLabel, model.GetPendingPromotion())
	assert.Equal(t, "test-key", model.GetKey())
	assert.True(t, model.Valid())
}
