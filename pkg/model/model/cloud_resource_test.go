package model

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudResource_WithStatus_TypePreservation(t *testing.T) {
	t.Run("AWSResource WithStatus preserves type", func(t *testing.T) {
		original, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"InstanceId": "i-1234567890abcdef0",
				"State":      "running",
			},
		)
		require.NoError(t, err)

		original.BaseAsset.Status = Active

		result := original.WithStatus("AH")

		awsResult, ok := result.(*AWSResource)
		require.True(t, ok, "WithStatus returned %T, expected *AWSResource", result)
		assert.Equal(t, "AH", awsResult.BaseAsset.Status)
		assert.Equal(t, original.Name, awsResult.Name)
		assert.Equal(t, Active, original.BaseAsset.Status, "Original status should not be modified")
	})

	t.Run("AzureResource WithStatus preserves type", func(t *testing.T) {
		sub := "e7c75ba8-b0ef-4ef8-bad2-fc8c30a92c70"
		name := fmt.Sprintf("/subscriptions/%s/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/test-vm", sub)
		original, err := NewAzureResource(
			name,
			sub,
			AzureVM,
			map[string]any{
				"location": "eastus",
				"vmSize":   "Standard_B1s",
			},
		)
		require.NoError(t, err)

		original.BaseAsset.Status = Active

		result := original.WithStatus("AL")

		azureResult, ok := result.(*AzureResource)
		require.True(t, ok, "WithStatus returned %T, expected *AzureResource", result)
		assert.Equal(t, "AL", azureResult.BaseAsset.Status)
		assert.Equal(t, original.Name, azureResult.Name)
		assert.Equal(t, Active, original.BaseAsset.Status, "Original status should not be modified")
	})

	t.Run("GCPResource WithStatus preserves type", func(t *testing.T) {
		original, err := NewGCPResource(
			"projects/test-project/zones/us-central1-a/instances/test-instance",
			"test-project",
			GCPResourceInstance,
			map[string]any{
				"machineType": "e2-micro",
				"zone":        "us-central1-a",
			},
		)
		require.NoError(t, err)

		original.BaseAsset.Status = Active

		result := original.WithStatus(ActivePassive)

		gcpResult, ok := result.(*GCPResource)
		require.True(t, ok, "WithStatus returned %T, expected *GCPResource", result)
		assert.Equal(t, ActivePassive, gcpResult.BaseAsset.Status)
		assert.Equal(t, original.Name, gcpResult.Name)
		assert.Equal(t, Active, original.BaseAsset.Status, "Original status should not be modified")
	})

	t.Run("All cloud resources maintain interface compliance", func(t *testing.T) {
		awsResource, _ := NewAWSResource("arn:aws:s3:::test-bucket", "123456789012", AWSS3Bucket, nil)
		azureResource, _ := NewAzureResource("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/test", "sub", AzureVM, nil)
		gcpResource, _ := NewGCPResource("projects/test/buckets/test-bucket", "test", GCPResourceBucket, nil)

		testCases := []struct {
			name     string
			resource Target
		}{
			{"AWS", &awsResource},
			{"Azure", &azureResource},
			{"GCP", &gcpResource},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := tc.resource.WithStatus("TEST")

				require.NotNil(t, result, "WithStatus returned nil for %s resource", tc.name)
				assert.Equal(t, "TEST", result.GetStatus())
				assert.NotEmpty(t, result.Group())
				assert.True(t, result.IsStatus("T"))
			})
		}
	})
}

func TestCloudResource_WithStatus_PreventTypeErasure(t *testing.T) {
	t.Run("regression test for type erasure bug", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{"InstanceId": "i-1234567890abcdef0"},
		)
		require.NoError(t, err)

		result := awsResource.WithStatus("AH")

		awsResult, ok := result.(*AWSResource)
		require.True(t, ok, "Expected *AWSResource, got %T - type erasure occurred!", result)

		// Test AWS-specific functionality - should not panic
		_ = awsResult.GetIPs()
		_ = awsResult.GetDNS()
		_ = awsResult.NewAssets()
	})
}

func TestCloudResource_NilPropertiesHandling(t *testing.T) {
	t.Run("AWS Merge with nil properties should not panic", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: "active",
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   map[string]any{"key": "value"},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: "updated",
				},
			},
		}

		resource1.Merge(resource2)

		assert.Equal(t, "updated", resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties)
		assert.Equal(t, "value", resource1.Properties["key"])
	})

	t.Run("AWS Visit with nil properties should not panic", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   map[string]any{"visited": true},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
				},
			},
		}

		resource1.Visit(resource2)

		assert.Equal(t, Active, resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties)
		assert.Equal(t, true, resource1.Properties["visited"])
	})

	t.Run("Azure Merge with nil properties should not panic", func(t *testing.T) {
		resource1 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: "active",
				},
			},
		}

		resource2 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   map[string]any{"location": "eastus"},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: "updated",
				},
			},
		}

		resource1.Merge(resource2)

		assert.Equal(t, "updated", resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties)
		assert.Equal(t, "eastus", resource1.Properties["location"])
	})

	t.Run("GCP Visit with nil properties should not panic", func(t *testing.T) {
		resource1 := &GCPResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "gcp",
				ResourceType: GCPResourceInstance,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
				},
			},
		}

		resource2 := &GCPResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "gcp",
				ResourceType: GCPResourceInstance,
				Properties:   map[string]any{"zone": "us-central1-a"},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
				},
			},
		}

		resource1.Visit(resource2)

		assert.Equal(t, Active, resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties)
		assert.Equal(t, "us-central1-a", resource1.Properties["zone"])
	})

	t.Run("Merge with both nil properties should not panic", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: "active",
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: "updated",
				},
			},
		}

		resource1.Merge(resource2)

		assert.Equal(t, "updated", resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties, "Properties should be initialized even when source is nil")
	})

	t.Run("Visit with source nil properties should not panic", func(t *testing.T) {
		resource1 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   map[string]any{"existing": "value"},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
				},
			},
		}

		resource2 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   nil,
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
				},
			},
		}

		resource1.Visit(resource2)

		assert.Equal(t, Active, resource1.BaseAsset.Status)
		require.NotNil(t, resource1.Properties)
		assert.Equal(t, "value", resource1.Properties["existing"], "Existing properties should be preserved when source is nil")
	})
}

func TestCloudResource_TTLUpdateLogic(t *testing.T) {
	t.Run("AWS TTL update logic should work correctly", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
					TTL:    0,
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
					TTL:    12345,
				},
			},
		}

		resource1.Visit(resource2)
		assert.Equal(t, int64(12345), resource1.BaseAsset.TTL)

		resource1.BaseAsset.TTL = 9999
		resource2.BaseAsset.TTL = 54321
		resource1.Visit(resource2)
		assert.Equal(t, int64(54321), resource1.BaseAsset.TTL)

		resource1.BaseAsset.TTL = 7777
		resource2.BaseAsset.TTL = 0
		resource1.Visit(resource2)
		assert.Equal(t, int64(0), resource1.BaseAsset.TTL)
	})

	t.Run("Azure TTL update logic should work correctly", func(t *testing.T) {
		resource1 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
					TTL:    0,
				},
			},
		}

		resource2 := &AzureResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "azure",
				ResourceType: AzureVM,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
					TTL:    98765,
				},
			},
		}

		resource1.Visit(resource2)
		assert.Equal(t, int64(98765), resource1.BaseAsset.TTL)
	})

	t.Run("GCP TTL update logic should work correctly", func(t *testing.T) {
		resource1 := &GCPResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "gcp",
				ResourceType: GCPResourceInstance,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
					TTL:    0,
				},
			},
		}

		resource2 := &GCPResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "gcp",
				ResourceType: GCPResourceInstance,
				Properties:   map[string]any{},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
					TTL:    11111,
				},
			},
		}

		resource1.Visit(resource2)
		assert.Equal(t, int64(11111), resource1.BaseAsset.TTL)
	})
}

func TestCloudResource_OriginationDataMerge(t *testing.T) {
	t.Run("CloudResource should merge OriginationData correctly", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: "instance",
				Properties:   map[string]any{"key1": "value1"},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: "active",
				},
				OriginationData: OriginationData{
					Capability:    []string{"dns"},
					AttackSurface: []string{"internal"},
					Origins:       []string{"dns"},
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: "instance",
				Properties:   map[string]any{"key2": "value2"},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: "updated",
				},
				OriginationData: OriginationData{
					Capability:    []string{"amazon", "portscan"},
					AttackSurface: []string{"external"},
					Origins:       []string{"amazon", "ipv4"},
				},
			},
		}

		resource1.Merge(resource2)

		assert.Equal(t, "updated", resource1.BaseAsset.Status)
		assert.Len(t, resource1.Properties, 2)
		assert.Equal(t, []string{"amazon", "portscan"}, resource1.Capability)
		assert.Equal(t, []string{"external"}, resource1.AttackSurface)
		assert.Equal(t, []string{"amazon", "ipv4"}, resource1.Origins)
	})
}

func TestCloudResource_OriginationDataVisit(t *testing.T) {
	t.Run("CloudResource should visit OriginationData correctly", func(t *testing.T) {
		resource1 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test1",
				Provider:     "aws",
				ResourceType: "instance",
				Properties:   map[string]any{"key1": "value1"},
				BaseAsset: BaseAsset{
					Key:    "test1",
					Status: Pending,
					Source: SelfSource,
					TTL:    0,
				},
				OriginationData: OriginationData{
					Capability:    []string{"dns"},
					AttackSurface: []string{"internal"},
					Origins:       []string{"dns"},
				},
			},
		}

		resource2 := &AWSResource{
			CloudResource: CloudResource{
				Name:         "test2",
				Provider:     "aws",
				ResourceType: "instance",
				Properties:   map[string]any{"key2": "value2"},
				BaseAsset: BaseAsset{
					Key:    "test2",
					Status: Active,
					Source: SelfSource,
					TTL:    12345,
				},
				OriginationData: OriginationData{
					Capability:    []string{"amazon", "portscan"},
					AttackSurface: []string{"external"},
					Origins:       []string{"amazon", "ipv4"},
				},
			},
		}

		resource1.Visit(resource2)

		assert.Equal(t, Active, resource1.BaseAsset.Status)
		assert.Equal(t, int64(12345), resource1.BaseAsset.TTL)
		assert.Equal(t, "value1", resource1.Properties["key1"])
		assert.Equal(t, "value2", resource1.Properties["key2"])

		sort.Strings(resource1.Capability)
		sort.Strings(resource1.AttackSurface)
		sort.Strings(resource1.Origins)

		assert.Equal(t, []string{"amazon", "dns", "portscan"}, resource1.Capability)
		assert.Equal(t, []string{"external", "internal"}, resource1.AttackSurface)
		assert.Equal(t, []string{"amazon", "dns", "ipv4"}, resource1.Origins)
	})
}

func TestAWSResource_OriginationDataIntegration(t *testing.T) {
	t.Run("AWSResource should use CloudResource OriginationData merge/visit", func(t *testing.T) {
		resource1, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{"InstanceId": "i-1234567890abcdef0"},
		)
		require.NoError(t, err)

		resource1.OriginationData = OriginationData{
			Capability: []string{"amazon"},
			Origins:    []string{"aws-account"},
		}

		resource2, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-abcdef1234567890",
			"123456789012",
			AWSEC2Instance,
			map[string]any{"InstanceId": "i-abcdef1234567890"},
		)
		require.NoError(t, err)

		resource2.OriginationData = OriginationData{
			Capability: []string{"portscan"},
			Origins:    []string{"discovery"},
		}

		resource1.Merge(&resource2)

		assert.Equal(t, []string{"portscan"}, resource1.Capability)

		resource1.OriginationData.Capability = []string{"amazon"}
		resource1.Visit(&resource2)

		sort.Strings(resource1.Capability)
		assert.Equal(t, []string{"amazon", "portscan"}, resource1.Capability)
	})
}

func TestCloudResource_VisitFields(t *testing.T) {
	original, err := NewAWSResource("arn:aws:ec2:us-east-1:123456789012:instance/i-abcdef1234567890", "123456789012", AWSEC2Instance, nil)
	require.NoError(t, err)

	updated, err := NewAWSResource("arn:aws:ec2:us-east-1:123456789012:instance/i-abcdef1234567890", "123456789012", AWSEC2Instance, nil)
	require.NoError(t, err)

	updated.DisplayName = "new-name"

	original.Visit(&updated)

	assert.Equal(t, "new-name", original.DisplayName)
}
