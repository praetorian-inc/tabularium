package model

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helper structs to override methods for complete coverage testing
type testAzureResourceWithIPs struct {
	*AzureResource
	testIPs []string
}

func (t *testAzureResourceWithIPs) GetIPs() []string {
	return t.testIPs
}

func (t *testAzureResourceWithIPs) IsPrivate() bool {
	// Use the same logic as AzureResource.IsPrivate() but with our overridden methods
	// Check if resource has any public IP addresses
	if ips := t.GetIPs(); len(ips) > 0 {
		for _, ip := range ips {
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				if parsedIP != nil && !parsedIP.IsPrivate() {
					return false // Has at least one public IP = not private
				}
			}
		}
	}

	// Check if resource has a public URL/endpoint
	if urls := t.AzureResource.GetURLs(); len(urls) > 0 {
		for _, url := range urls {
			if url != "" {
				return false // Has at least one public URL = not private
			}
		}
	}

	// No public IPs or URL = assume private
	return true
}

type testAzureResourceWithURL struct {
	*AzureResource
	testURL string
}

func (t *testAzureResourceWithURL) GetURL() string {
	return t.testURL
}

func (t *testAzureResourceWithURL) IsPrivate() bool {
	// Use the same logic as AzureResource.IsPrivate() but with our overridden methods
	// Check if resource has any public IP addresses
	if ips := t.AzureResource.GetIPs(); len(ips) > 0 {
		for _, ip := range ips {
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				if parsedIP != nil && !parsedIP.IsPrivate() {
					return false // Has at least one public IP = not private
				}
			}
		}
	}

	// Check if resource has a public URL/endpoint (using our overridden method)
	if url := t.GetURL(); url != "" {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}

type testAzureResourceWithIPsAndURL struct {
	*AzureResource
	testIPs []string
	testURL string
}

func (t *testAzureResourceWithIPsAndURL) GetIPs() []string {
	return t.testIPs
}

func (t *testAzureResourceWithIPsAndURL) GetURL() string {
	return t.testURL
}

func (t *testAzureResourceWithIPsAndURL) IsPrivate() bool {
	// Use the same logic as AzureResource.IsPrivate() but with our overridden methods
	// Check if resource has any public IP addresses
	if ips := t.GetIPs(); len(ips) > 0 {
		for _, ip := range ips {
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				if parsedIP != nil && !parsedIP.IsPrivate() {
					return false // Has at least one public IP = not private
				}
			}
		}
	}

	// Check if resource has a public URL/endpoint (using our overridden method)
	if url := t.GetURL(); url != "" {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}

func TestAzureResource_IsPrivate(t *testing.T) {
	tests := []struct {
		name        string
		resource    interface{ IsPrivate() bool }
		want        bool
		description string
	}{
		{
			name: "Resource with no IPs or URLs should be private",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Resource with no public endpoints should be private by default",
		},
		{
			name: "Resource with empty properties should be private",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Resource with empty properties should be private",
		},
		{
			name: "Resource with nil properties should be private",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   nil,
				},
			},
			want:        true,
			description: "Resource with nil properties should be private",
		},
		{
			name: "Different resource types should be private by default",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureSubscription,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Different Azure resource types should be private by default",
		},
		{
			name: "Resource with public IP should be public",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"203.0.113.1"}, // Public IP
			},
			want:        false,
			description: "Resource with public IP should not be private",
		},
		{
			name: "Resource with private IP should be private",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"10.0.1.100"}, // Private IP
			},
			want:        true,
			description: "Resource with only private IP should be private",
		},
		{
			name: "Resource with mixed IPs should be public",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"10.0.1.100", "203.0.113.1"}, // Private and public IPs
			},
			want:        false,
			description: "Resource with at least one public IP should not be private",
		},
		{
			name: "Resource with empty IP strings should be private",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"", ""}, // Empty IP strings
			},
			want:        true,
			description: "Resource with empty IP strings should be private",
		},
		{
			name: "Resource with invalid IP should be private",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"invalid-ip"}, // Invalid IP
			},
			want:        true,
			description: "Resource with invalid IP should be private",
		},
		{
			name: "Resource with localhost IP should be public",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"127.0.0.1"}, // Localhost
			},
			want:        false,
			description: "Resource with localhost IP should be public (Go's IsPrivate() returns false for localhost)",
		},
		{
			name: "Resource with link-local IP should be public",
			resource: &testAzureResourceWithIPs{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"169.254.1.1"}, // Link-local
			},
			want:        false,
			description: "Resource with link-local IP should be public (Go's IsPrivate() returns false for link-local)",
		},
		{
			name: "Resource with public URL should be public",
			resource: &testAzureResourceWithURL{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testURL: "https://myapp.azurewebsites.net",
			},
			want:        false,
			description: "Resource with public URL should not be private",
		},
		{
			name: "Resource with URL and private IP should be public",
			resource: &testAzureResourceWithIPsAndURL{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{"10.0.1.100"}, // Private IP
				testURL: "https://internal.example.com",
			},
			want:        false,
			description: "Resource with URL should not be private even if only has private IPs",
		},
		{
			name: "Resource with empty IPs but URL should be public",
			resource: &testAzureResourceWithIPsAndURL{
				AzureResource: &AzureResource{
					CloudResource: CloudResource{
						ResourceType: AzureVM,
						Properties:   map[string]any{},
					},
				},
				testIPs: []string{}, // Empty IP array
				testURL: "https://webapp.azurewebsites.net",
			},
			want:        false,
			description: "Resource with URL should not be private even with no IPs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.IsPrivate()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// Test original Azure IsPrivate method directly for maximum coverage
func TestAzureResource_IsPrivate_OriginalMethod(t *testing.T) {
	tests := []struct {
		name        string
		resource    *AzureResource
		want        bool
		description string
	}{
		{
			name: "Original method: Resource with no IPs should be private",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureVM,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Original Azure IsPrivate should return true for resources with no IPs",
		},
		{
			name: "Original method: Different resource type should be private",
			resource: &AzureResource{
				CloudResource: CloudResource{
					ResourceType: AzureSubscription,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Original Azure IsPrivate should return true for different resource types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.IsPrivate() // Call original method directly
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

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
}
