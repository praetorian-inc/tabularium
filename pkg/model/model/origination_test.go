package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveAttackSurfaceFlags_External(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"external"},
	}

	DeriveAttackSurfaceFlags(base)

	assert.True(t, base.IsExternal, "IsExternal should be true")
	assert.False(t, base.IsInternal, "IsInternal should be false")
	assert.False(t, base.IsCloud, "IsCloud should be false")
	assert.False(t, base.IsApplication, "IsApplication should be false")
	assert.False(t, base.IsRepository, "IsRepository should be false")
}

func TestDeriveAttackSurfaceFlags_Internal(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"internal"},
	}

	DeriveAttackSurfaceFlags(base)

	assert.False(t, base.IsExternal, "IsExternal should be false")
	assert.True(t, base.IsInternal, "IsInternal should be true")
	assert.False(t, base.IsCloud, "IsCloud should be false")
	assert.False(t, base.IsApplication, "IsApplication should be false")
	assert.False(t, base.IsRepository, "IsRepository should be false")
}

func TestDeriveAttackSurfaceFlags_Application(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"application"},
	}

	DeriveAttackSurfaceFlags(base)

	assert.True(t, base.IsExternal, "IsExternal should be true (application means external)")
	assert.False(t, base.IsInternal, "IsInternal should be false")
	assert.False(t, base.IsCloud, "IsCloud should be false")
	assert.True(t, base.IsApplication, "IsApplication should be true")
	assert.False(t, base.IsRepository, "IsRepository should be false")
}

func TestDeriveAttackSurfaceFlags_Multiple(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"external", "cloud"},
	}

	DeriveAttackSurfaceFlags(base)

	assert.True(t, base.IsExternal, "IsExternal should be true")
	assert.False(t, base.IsInternal, "IsInternal should be false")
	assert.True(t, base.IsCloud, "IsCloud should be true")
	assert.False(t, base.IsApplication, "IsApplication should be false")
	assert.False(t, base.IsRepository, "IsRepository should be false")
}

func TestDeriveAttackSurfaceFlags_Empty(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{},
	}

	DeriveAttackSurfaceFlags(base)

	assert.False(t, base.IsExternal, "IsExternal should be false")
	assert.False(t, base.IsInternal, "IsInternal should be false")
	assert.False(t, base.IsCloud, "IsCloud should be false")
	assert.False(t, base.IsApplication, "IsApplication should be false")
	assert.False(t, base.IsRepository, "IsRepository should be false")
}

func TestDeriveAttackSurfaceFlags_Idempotent(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"external", "cloud"},
		IsExternal:    false, // Existing values should be overwritten
		IsCloud:       false,
	}

	DeriveAttackSurfaceFlags(base)

	assert.True(t, base.IsExternal, "IsExternal should be true after derivation")
	assert.True(t, base.IsCloud, "IsCloud should be true after derivation")
}

func TestDeriveAttackSurfaceFlags_Repository(t *testing.T) {
	base := &OriginationData{
		AttackSurface: []string{"repository"},
	}

	DeriveAttackSurfaceFlags(base)

	assert.False(t, base.IsExternal, "IsExternal should be false")
	assert.False(t, base.IsInternal, "IsInternal should be false")
	assert.False(t, base.IsCloud, "IsCloud should be false")
	assert.False(t, base.IsApplication, "IsApplication should be false")
	assert.True(t, base.IsRepository, "IsRepository should be true")
}

// BLOCKER 1 TEST: Verify false values are explicit (not omitted)
func TestDeriveAttackSurfaceFlags_FalseValuesExplicit(t *testing.T) {
	data := &OriginationData{AttackSurface: []string{"internal"}}
	DeriveAttackSurfaceFlags(data)

	// Explicitly verify false is set (not just zero value)
	assert.True(t, data.IsInternal, "IsInternal must be true for internal surface")
	assert.False(t, data.IsExternal, "IsExternal must be explicitly false, not omitted")
	assert.False(t, data.IsCloud, "IsCloud must be explicitly false, not omitted")
	assert.False(t, data.IsApplication, "IsApplication must be explicitly false, not omitted")
	assert.False(t, data.IsRepository, "IsRepository must be explicitly false, not omitted")
}

// BLOCKER 2 TEST 1: Merge() auto-derives flags after merge
func TestOriginationData_Merge_DerivesFlagsAfterMerge(t *testing.T) {
	// Start with internal asset
	data := &OriginationData{
		AttackSurface: []string{"internal"},
		IsInternal:    true,
		IsExternal:    false,
	}

	// Merge with update that changes to external
	update := OriginationData{AttackSurface: []string{"external"}}
	data.Merge(update)

	// Flags must be re-derived after merge
	assert.True(t, data.IsExternal, "IsExternal must be true after merging external surface")
	assert.False(t, data.IsInternal, "IsInternal must be false after changing surface to external-only")
}

// BLOCKER 2 TEST 2: Visit() auto-derives flags after visit
func TestOriginationData_Visit_DerivesFlagsAfterVisit(t *testing.T) {
	// Asset currently internal
	data := &OriginationData{
		AttackSurface: []string{"internal"},
		IsInternal:    true,
		IsExternal:    false,
	}

	// Visit with capability result that adds external
	other := OriginationData{AttackSurface: []string{"external"}}
	data.Visit(other)

	// After union: ["internal", "external"] → both flags should be true
	assert.True(t, data.IsInternal, "IsInternal must remain true after union")
	assert.True(t, data.IsExternal, "IsExternal must be true after Visit adds external surface")
}
