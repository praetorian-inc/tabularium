package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsSeedPromotion(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		other    string
		expected bool
	}{
		{"non-seed to seed", SelfSource, SeedSource, true},
		{"seed to seed", SeedSource, SeedSource, false},
		{"self to self", SelfSource, SelfSource, false},
		{"seed to self", SeedSource, SelfSource, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := &BaseAsset{Source: tt.current}
			other := &BaseAsset{Source: tt.other}
			assert.Equal(t, tt.expected, IsSeedPromotion(current, other))
		})
	}
}

func TestApplySeedLabels(t *testing.T) {
	base := &BaseAsset{Source: SelfSource}
	ls := &LabelSettableEmbed{}

	ApplySeedLabels(base, ls)

	assert.Equal(t, SeedLabel, ls.PendingLabelAddition)
	assert.Equal(t, SeedSource, base.Source)
	assert.Empty(t, base.History.History, "ApplySeedLabels should NOT create history records")
}

func TestPromoteToSeed(t *testing.T) {
	base := &BaseAsset{Source: SelfSource, Status: Active}
	ls := &LabelSettableEmbed{}

	PromoteToSeed(base, ls, Pending)

	assert.Equal(t, SeedLabel, ls.PendingLabelAddition)
	assert.Equal(t, SeedSource, base.Source)
	require.Len(t, base.History.History, 1)
	assert.Equal(t, "", base.History.History[0].From)
	assert.Equal(t, Pending, base.History.History[0].To)
}

func TestMergeWithPromotionCheck_Promotion(t *testing.T) {
	// Active to Pending promotion
	base := &BaseAsset{Source: SelfSource, Status: Active}
	ls := &LabelSettableEmbed{}
	other := &Asset{BaseAsset: BaseAsset{Source: SeedSource, Status: Pending}}

	MergeWithPromotionCheck(base, ls, other)

	assert.Equal(t, SeedLabel, ls.PendingLabelAddition)
	assert.Equal(t, SeedSource, base.Source)
	assert.Equal(t, Pending, base.Status)
	require.Len(t, base.History.History, 1)
	assert.Equal(t, "", base.History.History[0].From)
	assert.Equal(t, Pending, base.History.History[0].To)
}

func TestMergeWithPromotionCheck_PromotionSameStatus(t *testing.T) {
	// Active to Active promotion (status preserved)
	base := &BaseAsset{Source: SelfSource, Status: Active}
	ls := &LabelSettableEmbed{}
	other := &Asset{BaseAsset: BaseAsset{Source: SeedSource, Status: Active}}

	MergeWithPromotionCheck(base, ls, other)

	assert.Equal(t, SeedLabel, ls.PendingLabelAddition)
	assert.Equal(t, SeedSource, base.Source)
	assert.Equal(t, Active, base.Status)
	require.Len(t, base.History.History, 1)
	assert.Equal(t, "", base.History.History[0].From)
	assert.Equal(t, Active, base.History.History[0].To)
}

func TestMergeWithPromotionCheck_NonPromotion(t *testing.T) {
	base := &BaseAsset{Source: SelfSource, Status: Active, History: History{History: []HistoryRecord{}}}
	ls := &LabelSettableEmbed{}
	other := &Asset{BaseAsset: BaseAsset{Source: SelfSource}}

	MergeWithPromotionCheck(base, ls, other)

	assert.Equal(t, NO_PENDING_LABEL_ADDITION, ls.PendingLabelAddition)
	assert.Equal(t, SelfSource, base.Source)
	assert.Empty(t, base.History.History)
}
