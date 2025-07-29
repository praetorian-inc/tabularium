package model

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"testing"

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
			got := tt.resource.GetIPs()
			assert.Equal(t, tt.want, got)
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
}

func TestNewAwsResource_Labels(t *testing.T) {
	name := "arn:aws:iam::123456789012:role/acme-admin-access"
	rtype := AWSRole
	accountRef := "123456789012"
	props := map[string]any{}

	awsRes, err := NewAWSResource(name, accountRef, rtype, props)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLabels := []string{"Role", "Principal", "AWS_IAM_Role", "AWSResource", "TTL"}
	actualLabels := slices.Clone(awsRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}
}

func TestNewAwsResource(t *testing.T) {
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
		if awsRes.Key != expectedKey {
			t.Errorf("expected Key '%s', got '%s'", expectedKey, awsRes.Key)
		}
		if awsRes.Name != name {
			t.Errorf("expected Name '%s', got '%s'", name, awsRes.Name)
		}
		if awsRes.DisplayName != "function:test-function" {
			t.Errorf("expected DisplayName 'function:test-function', got '%s'", awsRes.DisplayName)
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

		// Validate labels
		expectedLabels := []string{"AWS_Lambda_Function", "AWSResource", "TTL"}
		actualLabels := slices.Clone(awsRes.GetLabels())
		slices.Sort(actualLabels)
		slices.Sort(expectedLabels)
		if !slices.Equal(actualLabels, expectedLabels) {
			t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
		}

		// Validate properties
		if runtime, ok := awsRes.Properties["runtime"].(string); !ok || runtime != "python3.9" {
			t.Errorf("expected Properties[runtime] 'python3.9', got '%v'", awsRes.Properties["runtime"])
		}
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

func TestAwsResource_GetLabels(t *testing.T) {
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

	expectedLabels := []string{"AWS_EC2_Instance", "AWSResource", "TTL"}
	actualLabels := slices.Clone(awsRes.GetLabels())
	slices.Sort(actualLabels)
	slices.Sort(expectedLabels)
	if !slices.Equal(actualLabels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, actualLabels)
	}
}

func TestAWSResource_NewAsset(t *testing.T) {
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

		asset := awsResource.NewAsset()

		assert.Equal(t, "ec2", asset.CloudService)
		assert.Equal(t, "ec2-203-0-113-1.compute-1.amazonaws.com", asset.DNS)
		assert.Equal(t, "ec2-203-0-113-1.compute-1.amazonaws.com", asset.Name)
		assert.Equal(t, awsResource.Name, asset.CloudId)
		assert.Equal(t, "123456789012", asset.CloudAccount)
		assert.True(t, asset.Valid())
	})

	t.Run("Lambda function without DNS or IP", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:lambda:us-west-2:123456789012:function:my-function",
			"123456789012",
			AWSLambdaFunction,
			map[string]any{
				"Runtime": "python3.9",
			},
		)
		require.NoError(t, err)

		asset := awsResource.NewAsset()

		assert.Equal(t, "lambda", asset.CloudService)
		assert.Equal(t, awsResource.Name, asset.Name)
		assert.Equal(t, awsResource.Name, asset.DNS)
		assert.Equal(t, awsResource.Name, asset.CloudId)
		assert.True(t, asset.Valid(), asset.Key)
	})

	t.Run("S3 bucket without DNS or IP", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:s3:::my-test-bucket",
			"123456789012",
			AWSS3Bucket,
			map[string]any{
				"BucketName": "my-test-bucket",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		asset := awsResource.NewAsset()

		// Verify service extraction (should be "s3", not "aws")
		if asset.CloudService != "s3" {
			t.Errorf("Expected CloudService 's3', got '%s'", asset.CloudService)
		}

		// Verify ARN is used as fallback identifier
		if asset.Name != awsResource.Name {
			t.Errorf("Expected Name to be ARN '%s', got '%s'", awsResource.Name, asset.Name)
		}

		// DNS should be set to ARN for S3 (to create valid Asset key)
		if asset.DNS != awsResource.Name {
			t.Errorf("Expected DNS to be ARN '%s', got '%s'", awsResource.Name, asset.DNS)
		}

		// Verify valid key (should not be "#asset##")
		if !asset.Valid() {
			t.Errorf("Asset should be valid, but got invalid asset with key: %s", asset.Key)
		}
	})

	t.Run("IAM role without DNS or IP", func(t *testing.T) {
		awsResource, err := NewAWSResource(
			"arn:aws:iam::123456789012:role/MyRole",
			"123456789012",
			AWSRole,
			map[string]any{
				"RoleName": "MyRole",
			},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		asset := awsResource.NewAsset()

		// Verify service extraction (should be "iam", not "aws")
		if asset.CloudService != "iam" {
			t.Errorf("Expected CloudService 'iam', got '%s'", asset.CloudService)
		}

		// Verify ARN is used as fallback identifier
		if asset.Name != awsResource.Name {
			t.Errorf("Expected Name to be ARN '%s', got '%s'", awsResource.Name, asset.Name)
		}

		// Verify valid key (should not be "#asset##")
		if !asset.Valid() {
			t.Errorf("Asset should be valid, but got invalid asset with key: %s", asset.Key)
		}
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

		asset := awsResource.NewAsset()

		// Verify service extraction
		assert.Equal(t, "ec2", asset.CloudService)

		// Verify IP is used as identifier when DNS is empty
		assert.Equal(t, "203.0.113.1", asset.Name)

		// Verify valid key
		assert.True(t, asset.Valid())
	})

	t.Run("malformed ARN falls back gracefully", func(t *testing.T) {
		// This creates a resource with malformed ARN (bypassing constructor validation)
		awsResource := &AWSResource{
			CloudResource: CloudResource{
				Name:         "invalid-arn",
				Provider:     "aws",
				ResourceType: AWSEC2Instance,
				AccountRef:   "123456789012",
				Properties:   map[string]any{},
			},
		}

		asset := awsResource.NewAsset()

		// Should fall back to "Unknown Service"
		assert.Equal(t, "Unknown Service", asset.CloudService)

		// Should use ARN as fallback identifier
		assert.Equal(t, "invalid-arn", asset.Name)

		// Should still be valid
		assert.True(t, asset.Valid())
	})
}

func TestAWSResource_NewAsset_IssuesFixed(t *testing.T) {
	t.Run("Before fix: Service extraction and invalid Asset issues", func(t *testing.T) {
		// Test case 1: Lambda function - tests both issues
		lambdaARN := "arn:aws:lambda:us-west-2:123456789012:function:my-function"
		awsResource, err := NewAWSResource(
			lambdaARN,
			"123456789012",
			AWSLambdaFunction,
			map[string]any{"Runtime": "python3.9"},
		)
		if err != nil {
			t.Fatalf("Failed to create AWSResource: %v", err)
		}

		asset := awsResource.NewAsset()

		// ✅ ISSUE 1 FIXED: Service extraction now correctly uses index 2
		// Before fix: parts[1] would extract "aws" (partition)
		// After fix: parts[2] correctly extracts "lambda" (service)
		if asset.CloudService != "lambda" {
			t.Errorf("❌ Service extraction still broken: expected 'lambda', got '%s'", asset.CloudService)
		} else {
			t.Logf("✅ Service extraction fixed: correctly extracted 'lambda' from ARN")
		}

		// ✅ ISSUE 2 FIXED: Asset validation now passes
		// Before fix: NewAsset("", "") → Key: "#asset##" (invalid)
		// After fix: Uses ARN as identifier creating valid key "#asset#arn#arn"
		if !asset.Valid() {
			t.Errorf("❌ Asset validation still broken: invalid key '%s'", asset.Key)
		} else {
			t.Logf("✅ Asset validation fixed: valid key '%s'", asset.Key)
		}

		// Verify the fix details
		expectedKey := strings.ToLower(fmt.Sprintf("#asset#%s#%s", lambdaARN, lambdaARN))
		if asset.Key != expectedKey {
			t.Errorf("Unexpected key format: got '%s', expected '%s'", asset.Key, expectedKey)
		}

		t.Logf("📊 Lambda Asset Summary:")
		t.Logf("   CloudService: %s (correctly extracted from ARN)", asset.CloudService)
		t.Logf("   DNS: %s (set to ARN for valid key)", asset.DNS)
		t.Logf("   Name: %s (set to ARN)", asset.Name)
		t.Logf("   Key: %s (valid format)", asset.Key)
		t.Logf("   Valid: %t", asset.Valid())

		// Test case 2: S3 bucket - similar issues
		s3ARN := "arn:aws:s3:::my-test-bucket"
		s3Resource, err := NewAWSResource(
			s3ARN,
			"123456789012",
			AWSS3Bucket,
			map[string]any{"BucketName": "my-test-bucket"},
		)
		if err != nil {
			t.Fatalf("Failed to create S3Resource: %v", err)
		}

		s3Asset := s3Resource.NewAsset()

		// Verify S3 service extraction
		if s3Asset.CloudService != "s3" {
			t.Errorf("❌ S3 service extraction failed: expected 's3', got '%s'", s3Asset.CloudService)
		} else {
			t.Logf("✅ S3 service extraction fixed: correctly extracted 's3' from ARN")
		}

		// Verify S3 asset validation
		if !s3Asset.Valid() {
			t.Errorf("❌ S3 Asset validation failed: invalid key '%s'", s3Asset.Key)
		} else {
			t.Logf("✅ S3 Asset validation fixed: valid key '%s'", s3Asset.Key)
		}

		// Test case 3: EC2 with IP but no DNS - should use IP as identifier
		ec2Resource, err := NewAWSResource(
			"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			"123456789012",
			AWSEC2Instance,
			map[string]any{
				"PublicIp":      "203.0.113.1",
				"PublicDnsName": "", // No DNS
			},
		)
		if err != nil {
			t.Fatalf("Failed to create EC2Resource: %v", err)
		}

		ec2Asset := ec2Resource.NewAsset()

		// Should use IP when DNS is empty
		if ec2Asset.Name != "203.0.113.1" {
			t.Errorf("Expected EC2 asset to use IP as identifier, got '%s'", ec2Asset.Name)
		} else {
			t.Logf("✅ EC2 without DNS: correctly uses IP as identifier")
		}

		if !ec2Asset.Valid() {
			t.Errorf("❌ EC2 Asset validation failed: invalid key '%s'", ec2Asset.Key)
		} else {
			t.Logf("✅ EC2 Asset validation: valid key '%s'", ec2Asset.Key)
		}
	})
}

/*
Summary of fixes made to AWSResource.NewAsset():

ISSUE 1 - Incorrect AWS Service Extraction:
❌ Before: service = parts[1]  // extracted "aws" (partition)
✅ After:  service = parts[2]  // extracts actual service (lambda, s3, ec2, etc.)

ISSUE 2 - Invalid Asset Creation:
❌ Before: NewAsset("", "") → Key: "#asset##" (invalid)
✅ After:  Uses ARN as fallback → Key: "#asset#arn#arn" (valid)

The fix ensures:
1. Correct service extraction from ARN parts[2]
2. Valid Asset keys by using ARN as fallback when DNS/IP are empty
3. Maintains backward compatibility for EC2 instances with DNS/IP
*/
