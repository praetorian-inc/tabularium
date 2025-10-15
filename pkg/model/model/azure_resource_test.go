package model

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureResource_GetIPs(t *testing.T) {
	tests := []struct {
		name     string
		resource *AzureResource
		want     []string
	}{
		{
			name: "Resource with no properties",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   map[string]any{},
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
		{
			name: "Resource with nil properties",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   nil,
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
		{
			name: "Different resource types",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureSubscription,
					Properties:   map[string]any{},
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.GetIPs()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAzureResource_GetURL(t *testing.T) {
	tests := []struct {
		name     string
		resource *AzureResource
		want     []string
	}{
		{
			name: "Resource should return empty URL",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   map[string]any{},
				},
			},
			want: make([]string, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.GetURLs()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAzureResource_GetRegion(t *testing.T) {
	sub := "e7c75ba8-b0ef-4ef8-bad2-fc8c30a92c70"
	rg := "TEST_GROUP"
	name := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/test", sub, rg)

	t.Run("returns location when set", func(t *testing.T) {
		az, err := NewAzureResource(name, sub, AzureVM, map[string]any{"location": "eastus"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := az.GetRegion(); got != "eastus" {
			t.Errorf("expected 'eastus', got '%s'", got)
		}
	})

	t.Run("returns empty string when location not set", func(t *testing.T) {
		az, err := NewAzureResource(name, sub, AzureVM, map[string]any{"other": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := az.GetRegion(); got != "" {
			t.Errorf("expected empty string, got '%s'", got)
		}
	})

	t.Run("returns empty string when Properties is nil", func(t *testing.T) {
		az, err := NewAzureResource(name, sub, AzureVM, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := az.GetRegion(); got != "" {
			t.Errorf("expected empty string, got '%s'", got)
		}
	})
}

func TestNewAzureResource_Fields(t *testing.T) {
	sub := "e7c75ba8-b0ef-4ef8-bad2-fc8c30a92c70"
	rg := "TEST_GROUP"
	name := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/test", sub, rg)
	rtype := AzureVM
	props := map[string]any{
		"location":      "eastus",
		"resourceGroup": "TEST_GROUP",
	}
	az, err := NewAzureResource(name, sub, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKey := "#azureresource#" + sub + "#" + name
	if az.Key != expectedKey {
		t.Errorf("expected Key '%s', got '%s'", expectedKey, az.Key)
	}
	if az.Name != name {
		t.Errorf("expected Name '%s', got '%s'", name, az.Name)
	}
	if az.DisplayName != "test" {
		t.Errorf("expected DisplayName 'test', got '%s'", az.DisplayName)
	}
	if az.Provider != "azure" {
		t.Errorf("expected Provider 'azure', got '%s'", az.Provider)
	}
	if az.ResourceType != rtype {
		t.Errorf("expected ResourceType '%s', got '%s'", rtype, az.ResourceType)
	}
	if az.Region != "eastus" {
		t.Errorf("expected Region 'eastus', got '%s'", az.Region)
	}
	if az.AccountRef != sub {
		t.Errorf("expected AccountRef '%s', got '%s'", sub, az.AccountRef)
	}
	if az.ResourceGroup != rg {
		t.Errorf("expected ResourceGroup '%s', got '%s'", rg, az.ResourceGroup)
	}

	// Test defaulted origination data fields
	expectedOrigins := []string{"azure"}
	if !slices.Equal(az.Origins, expectedOrigins) {
		t.Errorf("expected Origins %v, got %v", expectedOrigins, az.Origins)
	}
	expectedAttackSurface := []string{"cloud"}
	if !slices.Equal(az.AttackSurface, expectedAttackSurface) {
		t.Errorf("expected AttackSurface %v, got %v", expectedAttackSurface, az.AttackSurface)
	}
}

func TestAzureResource_Defaulted(t *testing.T) {
	t.Run("Defaulted sets correct Origins and AttackSurface values", func(t *testing.T) {
		azureRes := &AzureResource{
			CloudResource: CloudResource{
				Name:         "/subscriptions/123/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/vm-test",
				Provider:     "azure",
				ResourceType: AzureVM,
				AccountRef:   "123",
			},
		}

		// Call Defaulted method directly
		azureRes.Defaulted()

		// Check that Origins is set to ["azure"]
		expectedOrigins := []string{"azure"}
		assert.Equal(t, expectedOrigins, azureRes.Origins, "Origins should be set to ['azure']")

		// Check that AttackSurface is set to ["cloud"]
		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, azureRes.AttackSurface, "AttackSurface should be set to ['cloud']")
	})

	t.Run("NewAzureResource calls Defaulted automatically", func(t *testing.T) {
		sub := "e7c75ba8-b0ef-4ef8-bad2-fc8c30a92c70"
		name := fmt.Sprintf("/subscriptions/%s/resourceGroups/test/providers/Microsoft.Compute/virtualMachines/test-vm", sub)
		azureRes, err := NewAzureResource(
			name,
			sub,
			AzureVM,
			map[string]any{"location": "eastus"},
		)
		require.NoError(t, err)

		// Verify that Origins and AttackSurface were set by NewAzureResource calling Defaulted()
		expectedOrigins := []string{"azure"}
		assert.Equal(t, expectedOrigins, azureRes.Origins, "NewAzureResource should call Defaulted() which sets Origins to ['azure']")

		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, azureRes.AttackSurface, "NewAzureResource should call Defaulted() which sets AttackSurface to ['cloud']")
	})
}
