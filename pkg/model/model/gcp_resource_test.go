package model

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCPResource_IsPrivate(t *testing.T) {
	tests := []struct {
		name        string
		resource    interface{ IsPrivate() bool }
		want        bool
		description string
	}{
		{
			name: "Resource with no IPs or URLs should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Resource with no public endpoints should be private by default",
		},
		{
			name: "Resource with empty properties should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Resource with empty properties should be private",
		},
		{
			name: "Resource with nil properties should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   nil,
				},
			},
			want:        true,
			description: "Resource with nil properties should be private",
		},
		{
			name: "Different resource types should be private by default",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceBucket,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Different GCP resource types should be private by default",
		},
		{
			name: "Service account should be private by default",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceServiceAccount,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "GCP service accounts should be private by default",
		},
		{
			name: "Project should be private by default",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceProject,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "GCP projects should be private by default",
		},
		{
			name: "Resource with public IP should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"203.0.113.1"}, // Public IP
					},
				},
			},
			want:        false,
			description: "Resource with public IP should not be private",
		},
		{
			name: "Resource with private IP should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIP": "10.0.1.100", // Private IP
					},
				},
			},
			want:        true,
			description: "Resource with only private IP should be private",
		},
		{
			name: "Resource with mixed IPs should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"10.0.1.100", "203.0.113.1"}, // Private and public IPs
					},
				},
			},
			want:        false,
			description: "Resource with at least one public IP should not be private",
		},
		{
			name: "Resource with empty IP strings should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"", ""}, // Empty IP strings
					},
				},
			},
			want:        true,
			description: "Resource with empty IP strings should be private",
		},
		{
			name: "Resource with invalid IP should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"invalid-ip"}, // ideally blocked in capabilities
					},
				},
			},
			want:        true,
			description: "Resource with invalid IP should be private",
		},
		{
			name: "Resource with localhost IP should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"127.0.0.1"},
					},
				},
			},
			want:        false,
			description: "Resource with localhost IP should be public (Go's IsPrivate() returns false for localhost)",
		},
		{
			name: "Resource with link-local IP should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"169.254.1.1"},
					},
				},
			},
			want:        false,
			description: "Resource with link-local IP should be public (Go's IsPrivate() returns false for link-local)",
		},
		{
			name: "Resource with public URL should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicURL": "https://my-app.run.app",
					},
				},
			},
			want:        false,
			description: "Resource with public URL should not be private",
		},
		{
			name: "Resource with URL and private IP should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs":  []string{"10.0.1.100"},
						"publicURLs": []string{"https://internal.example.com"},
					},
				},
			},
			want:        false,
			description: "Resource with URL should not be private even if only has private IPs",
		},
		{
			name: "Resource with empty IPs but URL should be public",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs":  []string{},
						"publicURLs": []string{"https://cloud-run.app"},
					},
				},
			},
			want:        false,
			description: "Resource with URL should not be private even with no IPs",
		},
		{
			name: "Resource with multiple valid private IPs should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicIPs": []string{"10.0.1.100", "192.168.1.1", "172.16.0.1"}, // All private IPs
					},
				},
			},
			want:        true,
			description: "Resource with multiple private IPs should be private",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.IsPrivate()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

// Test original GCP IsPrivate method directly for maximum coverage
func TestGCPResource_IsPrivate_OriginalMethod(t *testing.T) {
	tests := []struct {
		name        string
		resource    *GCPResource
		want        bool
		description string
	}{
		{
			name: "Original method: Resource with no IPs should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Original GCP IsPrivate should return true for resources with no IPs",
		},
		{
			name: "Original method: Service account should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceServiceAccount,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Original GCP IsPrivate should return true for service accounts",
		},
		{
			name: "Original method: Project should be private",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceProject,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Original GCP IsPrivate should return true for projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.IsPrivate() // Call original method directly
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

func TestGCPResource_GetIPs(t *testing.T) {
	tests := []struct {
		name     string
		resource *GCPResource
		want     []string
	}{
		{
			name: "Resource with no properties",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
		{
			name: "Resource with nil properties",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   nil,
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
		{
			name: "Different resource types",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceBucket,
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

func TestGCPResource_GetDNS(t *testing.T) {
	tests := []struct {
		name     string
		resource *GCPResource
		want     []string
	}{
		{
			name: "Resource should return empty DNS list",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want: []string{},
		},
		{
			name: "Resource with public DNS should return public DNS",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicDomain": "my-app.run.app",
					},
				},
			},
			want: []string{"my-app.run.app"},
		},
		{
			name: "Resource with multiple public DNS should return multiple public DNS",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties: map[string]any{
						"publicDomains": []string{"my-app.run.app", "my-app.run.app2"},
					},
				},
			},
			want: []string{"my-app.run.app", "my-app.run.app2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.GetDNS()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGCPResource_GetURL(t *testing.T) {
	tests := []struct {
		name     string
		resource *GCPResource
		want     []string
	}{
		{
			name: "Resource should return empty URL list",
			resource: &GCPResource{
				CloudResource: CloudResource{
					ResourceType: GCPResourceInstance,
					Properties:   map[string]any{},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.GetURLs()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewGcpResource(t *testing.T) {
	name := "projects/acme-project/zones/us-central1-a/instances/test-instance"
	rtype := GCPResourceInstance
	accountRef := "acme-project"
	props := map[string]any{
		"location": "us-central1-a",
	}

	gcpRes, err := NewGCPResource(name, accountRef, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate fields
	expectedKey := "#gcpresource#" + accountRef + "#" + name
	if gcpRes.Key != expectedKey {
		t.Errorf("expected Key '%s', got '%s'", expectedKey, gcpRes.Key)
	}
	if gcpRes.Name != name {
		t.Errorf("expected Name '%s', got '%s'", name, gcpRes.Name)
	}
	if gcpRes.DisplayName != "test-instance" {
		t.Errorf("expected DisplayName 'test-instance', got '%s'", gcpRes.DisplayName)
	}
	if gcpRes.Provider != "gcp" {
		t.Errorf("expected Provider 'gcp', got '%s'", gcpRes.Provider)
	}
	if gcpRes.ResourceType != rtype {
		t.Errorf("expected ResourceType '%s', got '%s'", rtype, gcpRes.ResourceType)
	}
	if gcpRes.AccountRef != accountRef {
		t.Errorf("expected AccountRef '%s', got '%s'", accountRef, gcpRes.AccountRef)
	}
	if gcpRes.Region != "us-central1" {
		t.Errorf("expected Region 'us-central1', got '%s'", gcpRes.Region)
	}

	// Validate labels
	expectedLabels := []string{"compute_googleapis_com_Instance", "GCPResource", "TTL", "Cloud"}
	actualLabels := slices.Clone(gcpRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}
}
