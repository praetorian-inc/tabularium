package model

import (
	"testing"
)

func TestEntitlement_Valid(t *testing.T) {
	tests := []struct {
		ent  Entitlement
		want bool
	}{
		{EntitlementRead, true},
		{EntitlementWriteAssets, true},
		{EntitlementManageAccounts, true},
		{EntitlementManageSettings, true},
		{EntitlementManageIntegrations, true},
		{EntitlementManageAgents, true},
		{EntitlementManageRedteam, true},
		{EntitlementManageOSINT, true},
		{EntitlementManageModels, true},
		{Entitlement("bogus"), false},
		{Entitlement(""), false},
	}
	for _, tt := range tests {
		if got := tt.ent.Valid(); got != tt.want {
			t.Errorf("Entitlement(%q).Valid() = %v, want %v", tt.ent, got, tt.want)
		}
	}
}

func TestEntitlementsForRole(t *testing.T) {
	tests := []struct {
		role Role
		want int // expected count
	}{
		{RoleReadOnly, 1},
		{RoleAnalyst, 2},
		{RoleAdmin, 9},
		{Role("bogus"), 0},
	}
	for _, tt := range tests {
		got := EntitlementsForRole(tt.role)
		if len(got) != tt.want {
			t.Errorf("EntitlementsForRole(%q) returned %d entitlements, want %d", tt.role, len(got), tt.want)
		}
	}
}

func TestRoleHasEntitlement(t *testing.T) {
	tests := []struct {
		role Role
		ent  Entitlement
		want bool
	}{
		// ReadOnly
		{RoleReadOnly, EntitlementRead, true},
		{RoleReadOnly, EntitlementWriteAssets, false},
		{RoleReadOnly, EntitlementManageAccounts, false},
		// Analyst
		{RoleAnalyst, EntitlementRead, true},
		{RoleAnalyst, EntitlementWriteAssets, true},
		{RoleAnalyst, EntitlementManageAccounts, false},
		{RoleAnalyst, EntitlementManageSettings, false},
		// Admin
		{RoleAdmin, EntitlementRead, true},
		{RoleAdmin, EntitlementWriteAssets, true},
		{RoleAdmin, EntitlementManageAccounts, true},
		{RoleAdmin, EntitlementManageSettings, true},
		{RoleAdmin, EntitlementManageIntegrations, true},
		{RoleAdmin, EntitlementManageAgents, true},
		{RoleAdmin, EntitlementManageRedteam, true},
		{RoleAdmin, EntitlementManageOSINT, true},
		{RoleAdmin, EntitlementManageModels, true},
		// Unknown role
		{Role("bogus"), EntitlementRead, false},
	}
	for _, tt := range tests {
		if got := RoleHasEntitlement(tt.role, tt.ent); got != tt.want {
			t.Errorf("RoleHasEntitlement(%q, %q) = %v, want %v", tt.role, tt.ent, got, tt.want)
		}
	}
}

func TestAllEntitlements(t *testing.T) {
	all := AllEntitlements()
	if len(all) != 9 {
		t.Errorf("AllEntitlements() returned %d, want 9", len(all))
	}

	// Verify it's a copy (modifying shouldn't affect the original)
	all[0] = "modified"
	if allEntitlements[0] == "modified" {
		t.Error("AllEntitlements() returned a reference, not a copy")
	}
}
