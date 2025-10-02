package model

import (
	"fmt"
	"net"
	"slices"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper structs to override methods for complete coverage testing
type testAWSResourceWithURL struct {
	*AWSResource
	testURL string
}

func (t *testAWSResourceWithURL) GetURL() string {
	return t.testURL
}

func (t *testAWSResourceWithURL) IsPrivate() bool {
	// Use the same logic as AWSResource.IsPrivate() but with our overridden methods
	// Check if resource has any public IP addresses
	if ips := t.AWSResource.GetIPs(); len(ips) > 0 {
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

func TestAWSResource_IsPrivate(t *testing.T) {
	tests := []struct {
		name        string
		resource    interface{ IsPrivate() bool }
		want        bool
		description string
	}{
		{
			name: "EC2 with public IP should be public",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "203.0.113.1", // Public IP
					},
				},
			},
			want:        false,
			description: "Resource with public IP should not be private",
		},
		{
			name: "EC2 with private IP should be private",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp":  "",
						"PrivateIp": "10.0.1.100", // Private IP
					},
				},
			},
			want:        true,
			description: "Resource with only private IP should be private",
		},
		{
			name: "EC2 with both public and private IP should be public",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp":  "203.0.113.1", // Public IP
						"PrivateIp": "10.0.1.100",  // Private IP
					},
				},
			},
			want:        false,
			description: "Resource with at least one public IP should not be private",
		},
		{
			name: "Resource with empty public IP should be private",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "",
					},
				},
			},
			want:        true,
			description: "Resource with empty public IP should be private",
		},
		{
			name: "Resource with no IPs or URLs should be private",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Resource with no public endpoints should be private",
		},
		{
			name: "Resource with empty IP strings should be private",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp":  "",
						"PrivateIp": "",
					},
				},
			},
			want:        true,
			description: "Resource with empty IP strings should be private",
		},
		{
			name: "Resource with invalid IP should be private",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "invalid-ip",
					},
				},
			},
			want:        true,
			description: "Resource with invalid IP should be private",
		},
		{
			name: "Resource with localhost IP should be public",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "127.0.0.1", // Localhost
					},
				},
			},
			want:        false,
			description: "Resource with localhost IP should be public (Go's IsPrivate() returns false for localhost)",
		},
		{
			name: "Resource with link-local IP should be public",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "169.254.1.1", // Link-local
					},
				},
			},
			want:        false,
			description: "Resource with link-local IP should be public (Go's IsPrivate() returns false for link-local)",
		},
		{
			name: "Non-EC2 resource should be private by default",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSS3Bucket,
					Properties:   map[string]any{},
				},
			},
			want:        true,
			description: "Non-EC2 resources without IPs should be private by default",
		},
		{
			name: "Resource with public URL should be public",
			resource: &testAWSResourceWithURL{
				AWSResource: &AWSResource{
					CloudResource: CloudResource{
						ResourceType: AWSS3Bucket,
						Properties:   map[string]any{},
					},
				},
				testURL: "https://my-bucket.s3.amazonaws.com",
			},
			want:        false,
			description: "Resource with public URL should not be private",
		},
		{
			name: "Resource with URL but also private IP should check IP first",
			resource: &testAWSResourceWithURL{
				AWSResource: &AWSResource{
					CloudResource: CloudResource{
						ResourceType: AWSEC2Instance,
						Properties: map[string]any{
							"PrivateIp": "10.0.1.100", // Private IP only
						},
					},
				},
				testURL: "https://internal.example.com",
			},
			want:        false,
			description: "Resource with URL should not be private even if only has private IPs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.IsPrivate()
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

func TestAWSResource_GetIPs(t *testing.T) {
	tests := []struct {
		name     string
		resource *AWSResource
		want     []string
	}{
		{
			name: "EC2 with both public and private IPs",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp":  "203.0.113.1",
						"PrivateIp": "10.0.1.100",
					},
				},
			},
			want: []string{"203.0.113.1", "10.0.1.100"},
		},
		{
			name: "EC2 with only public IP",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp": "203.0.113.1",
					},
				},
			},
			want: []string{"203.0.113.1"},
		},
		{
			name: "EC2 with only private IP",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PrivateIp": "10.0.1.100",
					},
				},
			},
			want: []string{"10.0.1.100"},
		},
		{
			name: "EC2 with empty IP strings",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSEC2Instance,
					Properties: map[string]any{
						"PublicIp":  "",
						"PrivateIp": "",
					},
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
		{
			name: "Non-EC2 resource",
			resource: &AWSResource{
				CloudResource: CloudResource{
					ResourceType: AWSS3Bucket,
					Properties:   map[string]any{},
				},
			},
			want: make([]string, 0), // Empty slice, not nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.CallHooks(tt.resource)
			got := tt.resource.GetIPs()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, got, tt.resource.IPs)
		})
	}
}

func TestNewAWSResource_Fields(t *testing.T) {
	name := "arn:aws:lambda:us-east-2:123456789012:function:test"
	rtype := AWSLambdaFunction
	accountRef := "123456789012"
	props := map[string]any{
		"runtime": "python3.9",
	}
	awsRes, err := NewAWSResource(name, accountRef, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKey := "#awsresource#" + accountRef + "#" + name
	if awsRes.Key != expectedKey {
		t.Errorf("expected Key '%s', got '%s'", expectedKey, awsRes.Key)
	}
	if awsRes.Name != name {
		t.Errorf("expected Name '%s', got '%s'", name, awsRes.Name)
	}
	if awsRes.Provider != "aws" {
		t.Errorf("expected Provider 'aws', got '%s'", awsRes.Provider)
	}
	if awsRes.ResourceType != rtype {
		t.Errorf("expected ResourceType '%s', got '%s'", rtype, awsRes.ResourceType)
	}
	if awsRes.Region != "us-east-2" {
		t.Errorf("expected Region 'us-east-2', got '%s'", awsRes.Region)
	}
	if awsRes.AccountRef != accountRef {
		t.Errorf("expected AccountRef '%s', got '%s'", accountRef, awsRes.AccountRef)
	}

	// Test defaulted origination data fields
	expectedOrigins := []string{"amazon"}
	if !slices.Equal(awsRes.Origins, expectedOrigins) {
		t.Errorf("expected Origins %v, got %v", expectedOrigins, awsRes.Origins)
	}
	expectedAttackSurface := []string{"cloud"}
	if !slices.Equal(awsRes.AttackSurface, expectedAttackSurface) {
		t.Errorf("expected AttackSurface %v, got %v", expectedAttackSurface, awsRes.AttackSurface)
	}
}

func TestNewAWSResource_Labels(t *testing.T) {
	name := "arn:aws:iam::123456789012:role/acme-admin-access"
	rtype := AWSRole
	accountRef := "123456789012"
	props := map[string]any{}

	awsRes, err := NewAWSResource(name, accountRef, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLabels := []string{"Role", "Principal", "AWS_IAM_Role", "Asset", "AWSResource", "TTL", "CloudResource"}
	actualLabels := slices.Clone(awsRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}
}

func TestNewAWSResource(t *testing.T) {
	t.Run("successful creation with valid ARN", func(t *testing.T) {
		name := "arn:aws:lambda:us-east-2:123456789012:function:test-function"
		rtype := AWSLambdaFunction
		accountRef := "123456789012"
		props := map[string]any{
			"runtime": "python3.9",
		}

		awsRes, err := NewAWSResource(name, accountRef, rtype, props)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Validate fields
		expectedKey := "#awsresource#" + accountRef + "#" + name
		assert.Equal(t, expectedKey, awsRes.Key)
		assert.Equal(t, name, awsRes.Name)
		assert.Equal(t, "function:test-function", awsRes.DisplayName)
		assert.Equal(t, "aws", awsRes.Provider)
		assert.Equal(t, rtype, awsRes.ResourceType)
		assert.Equal(t, "us-east-2", awsRes.Region)
		assert.Equal(t, accountRef, awsRes.AccountRef)

		// Validate labels
		expectedLabels := []string{"AWS_Lambda_Function", "Asset", "AWSResource", "TTL", "CloudResource"}
		actualLabels := slices.Clone(awsRes.GetLabels())
		slices.Sort(actualLabels)
		slices.Sort(expectedLabels)
		if !slices.Equal(actualLabels, expectedLabels) {
			t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
		}

		// Validate properties
		require.Contains(t, awsRes.Properties, "runtime")
		assert.Equal(t, awsRes.Properties["runtime"], "python3.9")
	})

	t.Run("error on invalid ARN", func(t *testing.T) {
		invalidARNs := []string{
			"invalid-arn",
			"arn:aws:invalid",
			"not-an-arn-at-all",
			"arn:aws",
			"arn:aws:ec2",
			"",
		}

		for _, invalidARN := range invalidARNs {
			t.Run("ARN: "+invalidARN, func(t *testing.T) {
				_, err := NewAWSResource(invalidARN, "123456789012", AWSEC2Instance, nil)
				if err == nil {
					t.Errorf("expected error for invalid ARN '%s', but got none", invalidARN)
				}

				expectedErrMsg := fmt.Sprintf("invalid ARN: %s", invalidARN)
				if err.Error() != expectedErrMsg {
					t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
				}
			})
		}
	})

	t.Run("successful creation with different resource types", func(t *testing.T) {
		testCases := []struct {
			name           string
			arn            string
			rtype          CloudResourceType
			expectedRegion string
		}{
			{
				name:           "EC2 Instance",
				arn:            "arn:aws:ec2:us-west-1:123456789012:instance/i-1234567890abcdef0",
				rtype:          AWSEC2Instance,
				expectedRegion: "us-west-1",
			},
			{
				name:           "S3 Bucket",
				arn:            "arn:aws:s3:::my-test-bucket",
				rtype:          AWSS3Bucket,
				expectedRegion: "",
			},
			{
				name:           "IAM Role",
				arn:            "arn:aws:iam::123456789012:role/test-role",
				rtype:          AWSRole,
				expectedRegion: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				awsRes, err := NewAWSResource(tc.arn, "123456789012", tc.rtype, nil)
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", tc.name, err)
				}

				if awsRes.ResourceType != tc.rtype {
					t.Errorf("expected ResourceType '%s', got '%s'", tc.rtype, awsRes.ResourceType)
				}
				if awsRes.Region != tc.expectedRegion {
					t.Errorf("expected Region '%s', got '%s'", tc.expectedRegion, awsRes.Region)
				}
				if awsRes.Provider != "aws" {
					t.Errorf("expected Provider 'aws', got '%s'", awsRes.Provider)
				}
			})
		}
	})

	t.Run("successful creation with nil properties", func(t *testing.T) {
		name := "arn:aws:s3:::test-bucket"
		awsRes, err := NewAWSResource(name, "123456789012", AWSS3Bucket, nil)
		if err != nil {
			t.Fatalf("unexpected error with nil properties: %v", err)
		}

		if awsRes.Properties == nil {
			t.Errorf("expected Properties to be initialized even when nil is passed")
		}
	})
}

func TestAWSResource_GetLabels(t *testing.T) {
	name := "arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef0"
	rtype := AWSEC2Instance
	accountRef := "123456789012"
	props := map[string]any{
		"region": "us-east-1",
	}

	awsRes, err := NewAWSResource(name, accountRef, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLabels := []string{"AWS_EC2_Instance", "AWSResource", "Asset", "TTL", "CloudResource"}
	actualLabels := slices.Clone(awsRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}
}

func TestAWSResource_NewAssets(t *testing.T) {
	t.Run("EC2 instance with DNS and IP", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"PublicIp":      "203.0.113.1",
				"PrivateIp":     "10.0.1.100",
				"PublicDnsName": "ec2-203-0-113-1.compute-1.amazonaws.com",
			},
		)
		require.NoError(t, err)

		assets := awsResource.NewAssets()

		if len(assets) != 4 {
			t.Errorf("Expected 4 assets got %d", len(assets))
		}

		// Check the first asset (public IP)
		asset := assets[0]

		assert.Equal(t, "ec2", asset.CloudService)
		assert.Equal(t, "ec2-203-0-113-1.compute-1.amazonaws.com", asset.DNS)
		assert.Equal(t, "203.0.113.1", asset.Name)
		assert.Equal(t, awsResource.Name, asset.CloudId)
		assert.Equal(t, "123456789012", asset.CloudAccount)
		assert.True(t, asset.Valid())
	})

	t.Run("EC2 instance with IP but no DNS", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"PublicIp":      "203.0.113.1",
				"PublicDnsName": "", // Empty DNS
			},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		assets := awsResource.NewAssets()

		// Should create 1 asset for the IP
		if len(assets) != 1 {
			t.Errorf("Expected 1 asset, got %d", len(assets))
		}

		asset := assets[0]

		// Verify service extraction
		assert.Equal(t, "ec2", asset.CloudService)

		// Verify IP is used as identifier when DNS is empty
		assert.Equal(t, "203.0.113.1", asset.Name)

		// Verify valid key
		assert.True(t, asset.Valid())
	})

	t.Run("resource with multiple IPs creates multiple assets", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"PublicIp":      "203.0.113.1",
				"PrivateIp":     "10.0.1.100",
				"PublicDnsName": "ec2-203-0-113-1.compute-1.amazonaws.com",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		assets := awsResource.NewAssets()

		if len(assets) != 4 {
			t.Errorf("Expected 4 assets (two for each IP), got %d", len(assets))
		}

		// Check first asset uses DNS+IP format
		asset1 := assets[0]
		if asset1.DNS != "ec2-203-0-113-1.compute-1.amazonaws.com" {
			t.Errorf("Expected first asset DNS to be DNS name, got '%s'", asset1.DNS)
		}

		// All assets should have the same cloud metadata
		for i, asset := range assets {
			if asset.CloudService != "ec2" {
				t.Errorf("Asset %d: Expected CloudService 'ec2', got '%s'", i, asset.CloudService)
			}
			if asset.CloudId != awsResource.Name {
				t.Errorf("Asset %d: Expected CloudId '%s', got '%s'", i, awsResource.Name, asset.CloudId)
			}
			if asset.CloudAccount != "123456789012" {
				t.Errorf("Asset %d: Expected CloudAccount '123456789012', got '%s'", i, asset.CloudAccount)
			}
			if !asset.Valid() {
				t.Errorf("Asset %d should be valid, but got invalid asset with key: %s", i, asset.Key)
			}
		}
	})

	t.Run("resource with multiple IPs creates multiple assets", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"PublicIp":      "203.0.113.1",
				"PrivateIp":     "10.0.1.100",
				"PublicDnsName": "ec2-203-0-113-1.compute-1.amazonaws.com",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		assets := awsResource.NewAssets()

		if len(assets) != 4 {
			t.Errorf("Expected 4 assets (two for each IP), got %d", len(assets))
		}

		// Check first asset uses DNS+IP format
		asset1 := assets[0]
		if asset1.DNS != "ec2-203-0-113-1.compute-1.amazonaws.com" {
			t.Errorf("Expected first asset DNS to be DNS name, got '%s'", asset1.DNS)
		}

		// All assets should have the same cloud metadata
		for i, asset := range assets {
			if asset.CloudService != "ec2" {
				t.Errorf("Asset %d: Expected CloudService 'ec2', got '%s'", i, asset.CloudService)
			}
			if asset.CloudId != awsResource.Name {
				t.Errorf("Asset %d: Expected CloudId '%s', got '%s'", i, awsResource.Name, asset.CloudId)
			}
			if asset.CloudAccount != "123456789012" {
				t.Errorf("Asset %d: Expected CloudAccount '123456789012', got '%s'", i, asset.CloudAccount)
			}
			if !asset.Valid() {
				t.Errorf("Asset %d should be valid, but got invalid asset with key: %s", i, asset.Key)
			}
		}
	})
}

func TestAWSResource_Defaulted(t *testing.T) {
	t.Run("Defaulted sets correct Origins and AttackSurface values", func(t *testing.T) {
		awsRes := &AWSResource{
			CloudResource: CloudResource{
				Name:         "arn:aws:s3:::test-bucket",
				Provider:     "aws",
				ResourceType: AWSS3Bucket,
				AccountRef:   "123456789012",
			},
		}

		// Call Defaulted method directly
		awsRes.Defaulted()

		// Check that Origins is set to ["amazon"]
		expectedOrigins := []string{"amazon"}
		assert.Equal(t, expectedOrigins, awsRes.Origins, "Origins should be set to ['amazon']")

		// Check that AttackSurface is set to ["cloud"]
		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, awsRes.AttackSurface, "AttackSurface should be set to ['cloud']")
	})

	t.Run("NewAWSResource calls Defaulted automatically", func(t *testing.T) {
		awsRes, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{"region": "us-east-1"},
		)
		require.NoError(t, err)

		// Verify that Origins and AttackSurface were set by NewAWSResource calling Defaulted()
		expectedOrigins := []string{"amazon"}
		assert.Equal(t, expectedOrigins, awsRes.Origins, "NewAWSResource should call Defaulted() which sets Origins to ['amazon']")

		expectedAttackSurface := []string{"cloud"}
		assert.Equal(t, expectedAttackSurface, awsRes.AttackSurface, "NewAWSResource should call Defaulted() which sets AttackSurface to ['cloud']")
	})
}

func TestAWSResource_HydrateDehydrate(t *testing.T) {
	resource, err := NewAWSResource("arn:aws:organizations::992382775570:account/o-a6zw2rb1jz/992382775570", "992382775570", AWSAccount, nil)
	require.NoError(t, err)

	gotFilepath := resource.HydratableFilepath()
	assert.Equal(t, gotFilepath, "")

	err = resource.Hydrate([]byte(`{"dummy": "test policy"}`))
	require.NoError(t, err)

	gotFilepath = resource.HydratableFilepath()
	expectedFilepath := "awsresource/992382775570/arn_aws_organizations__992382775570_account_o-a6zw2rb1jz_992382775570/org-policies.json"
	assert.Equal(t, gotFilepath, expectedFilepath)

	expectedFile := NewFile(expectedFilepath)
	expectedFile.Bytes = []byte(`{"dummy": "test policy"}`)
	gotFile := resource.HydratedFile()
	assert.Equal(t, expectedFile.Key, gotFile.Key)
	assert.Equal(t, expectedFile.Name, gotFile.Name)
	assert.Equal(t, expectedFile.Bytes, gotFile.Bytes)

	dehydrated, ok := resource.Dehydrate().(*AWSResource)
	require.True(t, ok, "object is not *AWSResource: %T", resource)
	assert.Nil(t, dehydrated.OrgPolicy)
}

func TestAWSResource_Visit(t *testing.T) {
	existing, err := NewAWSResource("arn:aws:organizations::992382775570:account/o-a6zw2rb1jz/992382775570", "992382775570", AWSAccount, nil)
	require.NoError(t, err)

	other, err := NewAWSResource("arn:aws:organizations::992382775570:account/o-a6zw2rb1jz/992382775570", "992382775570", AWSAccount, nil)
	other.OrgPolicyName = "other-file"
	require.NoError(t, err)

	existing.Merge(&other)

	assert.Equal(t, existing.OrgPolicyName, "other-file")
}
