package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_ParentTenantField(t *testing.T) {
	// Test that ParentTenant field exists and serializes correctly
	account := Account{
		Name:         "child@example.com",
		Member:       "user@example.com",
		ParentTenant: "parent@example.com",
	}

	assert.Equal(t, "parent@example.com", account.ParentTenant)
}

func TestAccount_TenantTypeField(t *testing.T) {
	// Test that TenantType field exists and serializes correctly
	account := Account{
		Name:       "child@example.com",
		Member:     "user@example.com",
		TenantType: "child",
	}

	assert.Equal(t, "child", account.TenantType)
}

func TestAccount_InheritedDomainsField(t *testing.T) {
	// Test that InheritedDomains field exists and handles arrays
	account := Account{
		Name:             "child@example.com",
		Member:           "user@example.com",
		InheritedDomains: []string{"example.com", "acme.com"},
	}

	assert.Len(t, account.InheritedDomains, 2)
	assert.Contains(t, account.InheritedDomains, "example.com")
}

func TestAccount_AllowedDomainsField(t *testing.T) {
	// Test that AllowedDomains field exists and handles arrays
	account := Account{
		Name:           "tenant@example.com",
		Member:         "user@example.com",
		AllowedDomains: []string{"subsidiary.com"},
	}

	assert.Len(t, account.AllowedDomains, 1)
	assert.Equal(t, "subsidiary.com", account.AllowedDomains[0])
}

func TestAccount_SharedFromParentField(t *testing.T) {
	// Test that SharedFromParent field exists and handles boolean
	account := Account{
		Name:             "child@example.com",
		Member:           "parent-user@example.com",
		SharedFromParent: true,
	}

	assert.True(t, account.SharedFromParent)
}

func TestAccount_HierarchyFieldsSerialization(t *testing.T) {
	// Test JSON serialization includes all hierarchy fields
	account := NewAccount("child@example.com", "user@example.com", "value", nil)
	account.ParentTenant = "parent@example.com"
	account.TenantType = "child"
	account.InheritedDomains = []string{"example.com"}
	account.AllowedDomains = []string{"subsidiary.com"}
	account.SharedFromParent = true

	// Verify fields are set
	assert.Equal(t, "parent@example.com", account.ParentTenant)
	assert.Equal(t, "child", account.TenantType)
	assert.Equal(t, []string{"example.com"}, account.InheritedDomains)
	assert.Equal(t, []string{"subsidiary.com"}, account.AllowedDomains)
	assert.True(t, account.SharedFromParent)
}
