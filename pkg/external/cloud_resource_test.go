package external

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudResource_ToModel(t *testing.T) {
	tests := []struct {
		name    string
		cr      CloudResource
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cloud resource",
			cr: CloudResource{
				Name:         "test-resource",
				Provider:     "aws",
				ResourceType: "AWS::EC2::Instance",
				DisplayName:  "Test Instance",
				Region:       "us-east-1",
				AccountRef:   "123456789012",
				Properties:   map[string]any{"key": "value"},
				IPs:          []string{"1.2.3.4"},
				URLs:         []string{"https://example.com"},
				Labels:       []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cr: CloudResource{
				Provider:   "aws",
				AccountRef: "123456789012",
			},
			wantErr: true,
			errMsg:  "cloud resource requires name",
		},
		{
			name: "missing provider",
			cr: CloudResource{
				Name:       "test-resource",
				AccountRef: "123456789012",
			},
			wantErr: true,
			errMsg:  "cloud resource requires provider",
		},
		{
			name: "missing accountRef",
			cr: CloudResource{
				Name:     "test-resource",
				Provider: "aws",
			},
			wantErr: true,
			errMsg:  "cloud resource requires accountRef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.cr.ToModel()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.cr.Name, result.Name)
				assert.Equal(t, tt.cr.Provider, result.Provider)
				assert.Equal(t, tt.cr.ResourceType, result.ResourceType.String())
				assert.Equal(t, tt.cr.DisplayName, result.DisplayName)
				assert.Equal(t, tt.cr.Region, result.Region)
				assert.Equal(t, tt.cr.AccountRef, result.AccountRef)
				assert.Equal(t, tt.cr.Properties, result.Properties)
				assert.Equal(t, tt.cr.IPs, result.IPs)
				assert.Equal(t, tt.cr.URLs, result.URLs)
				assert.Equal(t, tt.cr.Labels, result.Labels)
			}
		})
	}
}

func TestCloudResource_GroupAndIdentifier(t *testing.T) {
	cr := CloudResource{
		Name:       "test-resource",
		AccountRef: "123456789012",
	}

	assert.Equal(t, "123456789012", cr.Group())
	assert.Equal(t, "test-resource", cr.Identifier())
}

func TestCloudResourceFromModel(t *testing.T) {
	modelCR := &model.CloudResource{
		Name:         "test-resource",
		Provider:     "aws",
		ResourceType: model.AWSEC2Instance,
		DisplayName:  "Test Instance",
		Region:       "us-east-1",
		AccountRef:   "123456789012",
		Properties:   map[string]any{"key": "value"},
		IPs:          []string{"1.2.3.4"},
		URLs:         []string{"https://example.com"},
		Labels:       []string{"test"},
	}

	externalCR := CloudResourceFromModel(modelCR)

	assert.Equal(t, modelCR.Name, externalCR.Name)
	assert.Equal(t, modelCR.Provider, externalCR.Provider)
	assert.Equal(t, modelCR.ResourceType.String(), externalCR.ResourceType)
	assert.Equal(t, modelCR.DisplayName, externalCR.DisplayName)
	assert.Equal(t, modelCR.Region, externalCR.Region)
	assert.Equal(t, modelCR.AccountRef, externalCR.AccountRef)
	assert.Equal(t, modelCR.Properties, externalCR.Properties)
	assert.Equal(t, modelCR.IPs, externalCR.IPs)
	assert.Equal(t, modelCR.URLs, externalCR.URLs)
	assert.Equal(t, modelCR.Labels, externalCR.Labels)
}

func TestCloudResource_RoundTrip(t *testing.T) {
	// Test that we can convert to model and back
	original := CloudResource{
		Name:         "test-resource",
		Provider:     "aws",
		ResourceType: "AWS::EC2::Instance",
		DisplayName:  "Test Instance",
		Region:       "us-east-1",
		AccountRef:   "123456789012",
		Properties:   map[string]any{"key": "value"},
		IPs:          []string{"1.2.3.4"},
		URLs:         []string{"https://example.com"},
		Labels:       []string{"test"},
	}

	// Convert to model
	modelCR, err := original.ToModel()
	require.NoError(t, err)

	// Convert back
	result := CloudResourceFromModel(modelCR)

	// Verify essential fields match
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Provider, result.Provider)
	assert.Equal(t, original.ResourceType, result.ResourceType)
	assert.Equal(t, original.DisplayName, result.DisplayName)
	assert.Equal(t, original.Region, result.Region)
	assert.Equal(t, original.AccountRef, result.AccountRef)
	assert.Equal(t, original.Properties, result.Properties)
	assert.Equal(t, original.IPs, result.IPs)
	assert.Equal(t, original.URLs, result.URLs)
}
