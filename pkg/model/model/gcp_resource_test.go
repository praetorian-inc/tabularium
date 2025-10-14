package model

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNewGCPResource(t *testing.T) {
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
	assert.Equal(t, expectedKey, gcpRes.Key)
	assert.Equal(t, name, gcpRes.Name)
	assert.Equal(t, "test-instance", gcpRes.DisplayName)
	assert.Equal(t, "gcp", gcpRes.Provider)
	assert.Equal(t, rtype, gcpRes.ResourceType)
	assert.Equal(t, accountRef, gcpRes.AccountRef)
	assert.Equal(t, "us-central1", gcpRes.Region)

	// Validate labels
	expectedLabels := []string{"compute_googleapis_com_Instance", "GCPResource", "Asset", "TTL", "CloudResource"}
	actualLabels := slices.Clone(gcpRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}

	// Test defaulted origination data fields
	expectedOrigins := []string{"gcp"}
	if !slices.Equal(gcpRes.Origins, expectedOrigins) {
		t.Errorf("expected Origins %v, got %v", expectedOrigins, gcpRes.Origins)
	}
	expectedAttackSurface := []string{"cloud"}
	if !slices.Equal(gcpRes.AttackSurface, expectedAttackSurface) {
		t.Errorf("expected AttackSurface %v, got %v", expectedAttackSurface, gcpRes.AttackSurface)
	}
}

func TestGCPResource_Defaulted(t *testing.T) {
	t.Run("Defaulted sets correct Origins and AttackSurface values", func(t *testing.T) {
		gcpRes := &GCPResource{
			CloudResource: CloudResource{
				Name:         "projects/test-project/zones/us-central1-a/instances/test-vm",
				Provider:     "gcp",
				ResourceType: GCPResourceInstance,
				AccountRef:   "test-project",
			},
		}

		// Call Defaulted method directly
		gcpRes.Defaulted()

		// Check that Origins is set to ["gcp"]
		expectedOrigins := []string{"gcp"}
		assert.Equal(t, expectedOrigins, gcpRes.Origins, "Origins should be set to ['gcp']")

		// Check that AttackSurface is set to ["cloud"]
		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, gcpRes.AttackSurface, "AttackSurface should be set to ['cloud']")
	})

	t.Run("NewGCPResource calls Defaulted automatically", func(t *testing.T) {
		name := "projects/test-project/zones/us-central1-a/instances/test-instance"
		gcpRes, err := NewGCPResource(
			name,
			"test-project",
			GCPResourceInstance,
			map[string]any{"location": "us-central1-a"},
		)
		require.NoError(t, err)

		// Verify that Origins and AttackSurface were set by NewGCPResource calling Defaulted()
		expectedOrigins := []string{"gcp"}
		assert.Equal(t, expectedOrigins, gcpRes.Origins, "NewGCPResource should call Defaulted() which sets Origins to ['gcp']")

		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, gcpRes.AttackSurface, "NewGCPResource should call Defaulted() which sets AttackSurface to ['cloud']")
	})
}
