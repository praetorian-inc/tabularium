package convert

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/external"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset(t *testing.T) {
	t.Run("ToModel creates valid asset", func(t *testing.T) {
		ext := external.Asset{DNS: "example.com", Name: "192.168.1.1"}

		asset, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, asset)

		assert.Equal(t, "example.com", asset.DNS)
		assert.Equal(t, "192.168.1.1", asset.Name)
		assert.Contains(t, asset.Key, "#asset#")
		assert.Equal(t, model.Active, asset.Status)
	})

	t.Run("implements Target interface", func(t *testing.T) {
		ext := external.Asset{DNS: "example.com", Name: "192.168.1.1"}

		assert.Equal(t, "example.com", ext.Group())
		assert.Equal(t, "192.168.1.1", ext.Identifier())

		target, err := ext.ToTarget()
		require.NoError(t, err)
		assert.IsType(t, &model.Asset{}, target)
	})

	t.Run("empty asset returns error", func(t *testing.T) {
		ext := external.Asset{}
		_, err := ext.ToModel()
		require.Error(t, err)
	})
}

func TestRisk(t *testing.T) {
	t.Run("ToModel with asset target", func(t *testing.T) {
		ext := external.Risk{
			Name:   "CVE-2023-12345",
			Status: "TH",
			Target: external.Asset{DNS: "example.com", Name: "192.168.1.1"},
		}

		risk, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, risk)

		assert.Equal(t, "CVE-2023-12345", risk.Name)
		assert.Equal(t, "TH", risk.Status)
		assert.Equal(t, "example.com", risk.DNS)
		assert.Contains(t, risk.Key, "#risk#")
	})

	t.Run("default status when empty", func(t *testing.T) {
		ext := external.Risk{
			Name:   "CVE-2023-12345",
			Target: external.Asset{DNS: "example.com", Name: "example.com"},
		}

		risk, err := ext.ToModel()
		require.NoError(t, err)
		assert.Equal(t, "TH", risk.Status) // Default to Triage High
	})

	t.Run("missing name returns error", func(t *testing.T) {
		ext := external.Risk{
			Status: "TH",
			Target: external.Asset{DNS: "example.com", Name: "192.168.1.1"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})

	t.Run("missing target returns error", func(t *testing.T) {
		ext := external.Risk{
			Name:   "CVE-2023-12345",
			Status: "TH",
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})
}

func TestIPToAsset(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		domain      string
		wantDNS     string
		wantName    string
		wantClass   string
		wantPrivate bool
		wantErr     bool
	}{
		{
			name:        "public IPv4",
			address:     "8.8.8.8",
			domain:      "example.com",
			wantDNS:     "example.com",
			wantName:    "8.8.8.8",
			wantClass:   "ipv4",
			wantPrivate: false,
		},
		{
			name:        "private IPv4",
			address:     "192.168.1.1",
			domain:      "internal.local",
			wantDNS:     "internal.local",
			wantName:    "192.168.1.1",
			wantClass:   "ipv4",
			wantPrivate: true,
		},
		{
			name:        "IPv4 without domain",
			address:     "10.0.0.1",
			wantDNS:     "10.0.0.1",
			wantName:    "10.0.0.1",
			wantClass:   "ipv4",
			wantPrivate: true,
		},
		{
			name:        "IPv6",
			address:     "2001:4860:4860::8888",
			domain:      "google.com",
			wantDNS:     "google.com",
			wantName:    "2001:4860:4860::8888",
			wantClass:   "ipv6",
			wantPrivate: false,
		},
		{
			name:    "invalid IP",
			address: "not-an-ip",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var asset *model.Asset
			var err error

			if tt.domain != "" {
				asset, err = IPToAsset(tt.address, tt.domain)
			} else {
				asset, err = IPToAsset(tt.address)
			}

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, asset)

			assert.Equal(t, tt.wantDNS, asset.DNS)
			assert.Equal(t, tt.wantName, asset.Name)
			assert.Equal(t, tt.wantClass, asset.Class)
			assert.Equal(t, tt.wantPrivate, asset.Private)
			assert.Contains(t, asset.Key, "#asset#")
		})
	}
}

func TestDomainToAsset(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		wantDNS   string
		wantName  string
		wantClass string
		wantErr   bool
	}{
		{
			name:      "simple domain",
			domain:    "example.com",
			wantDNS:   "example.com",
			wantName:  "example.com",
			wantClass: "tld",
		},
		{
			name:      "subdomain",
			domain:    "api.example.com",
			wantDNS:   "api.example.com",
			wantName:  "api.example.com",
			wantClass: "domain",
		},
		{
			name:      "domain with https prefix",
			domain:    "https://example.com",
			wantDNS:   "example.com",
			wantName:  "example.com",
			wantClass: "tld",
		},
		{
			name:      "domain with http prefix and trailing slash",
			domain:    "http://example.com/",
			wantDNS:   "example.com",
			wantName:  "example.com",
			wantClass: "tld",
		},
		{
			name:    "empty domain",
			domain:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := DomainToAsset(tt.domain)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, asset)

			assert.Equal(t, tt.wantDNS, asset.DNS)
			assert.Equal(t, tt.wantName, asset.Name)
			assert.Equal(t, tt.wantClass, asset.Class)
			assert.Contains(t, asset.Key, "#asset#")
		})
	}
}

func TestCIDRToAsset(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		wantDNS     string
		wantName    string
		wantClass   string
		wantPrivate bool
		wantErr     bool
	}{
		{
			name:        "public CIDR",
			cidr:        "8.8.8.0/24",
			wantDNS:     "8.8.8.0/24",
			wantName:    "8.8.8.0/24",
			wantClass:   "cidr",
			wantPrivate: false,
		},
		{
			name:        "private CIDR",
			cidr:        "10.0.0.0/8",
			wantDNS:     "10.0.0.0/8",
			wantName:    "10.0.0.0/8",
			wantClass:   "cidr",
			wantPrivate: true,
		},
		{
			name:        "private CIDR 192.168",
			cidr:        "192.168.0.0/16",
			wantDNS:     "192.168.0.0/16",
			wantName:    "192.168.0.0/16",
			wantClass:   "cidr",
			wantPrivate: true,
		},
		{
			name:    "invalid CIDR",
			cidr:    "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := CIDRToAsset(tt.cidr)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, asset)

			assert.Equal(t, tt.wantDNS, asset.DNS)
			assert.Equal(t, tt.wantName, asset.Name)
			assert.Equal(t, tt.wantClass, asset.Class)
			assert.Equal(t, tt.wantPrivate, asset.Private)
			assert.Contains(t, asset.Key, "#asset#")
		})
	}
}

func TestAssetFromModel(t *testing.T) {
	fullAsset := model.NewAsset("example.com", "192.168.1.1")

	ext := AssetFromModel(&fullAsset)

	assert.Equal(t, "example.com", ext.DNS)
	assert.Equal(t, "192.168.1.1", ext.Name)
}

func TestAWSResource(t *testing.T) {
	t.Run("ToModel creates valid AWS resource", func(t *testing.T) {
		ext := external.AWSResource{
			ARN:          "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			AccountRef:   "123456789012",
			ResourceType: model.AWSEC2Instance,
			Properties: map[string]any{
				"PublicIp": "54.123.45.67",
			},
		}

		resource, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, resource)

		assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0", resource.Name)
		assert.Equal(t, "123456789012", resource.AccountRef)
		assert.Equal(t, model.AWSEC2Instance, resource.ResourceType)
		assert.Equal(t, model.AWSProvider, resource.Provider)
		assert.Equal(t, "us-east-1", resource.Region)
		assert.Contains(t, resource.Key, "#awsresource#")
		assert.Equal(t, model.Active, resource.Status)
	})

	t.Run("implements Target interface", func(t *testing.T) {
		ext := external.AWSResource{
			ARN:          "arn:aws:s3:::my-bucket",
			AccountRef:   "123456789012",
			ResourceType: model.AWSS3Bucket,
		}

		assert.Equal(t, "123456789012", ext.Group())
		assert.Equal(t, "arn:aws:s3:::my-bucket", ext.Identifier())

		target, err := ext.ToTarget()
		require.NoError(t, err)
		assert.IsType(t, &model.AWSResource{}, target)
	})

	t.Run("missing ARN returns error", func(t *testing.T) {
		ext := external.AWSResource{
			AccountRef:   "123456789012",
			ResourceType: model.AWSEC2Instance,
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "arn")
	})

	t.Run("missing AccountRef returns error", func(t *testing.T) {
		ext := external.AWSResource{
			ARN:          "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			ResourceType: model.AWSEC2Instance,
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "accountRef")
	})

	t.Run("preserves OrgPolicyFilename", func(t *testing.T) {
		ext := external.AWSResource{
			ARN:               "arn:aws:lambda:us-west-2:123456789012:function:my-function",
			AccountRef:        "123456789012",
			ResourceType:      model.AWSLambdaFunction,
			OrgPolicyFilename: "awsresource/123456789012/my-function/org-policies.json",
		}

		resource, err := ext.ToModel()
		require.NoError(t, err)
		assert.Equal(t, "awsresource/123456789012/my-function/org-policies.json", resource.OrgPolicyFilename)
	})

	t.Run("invalid ARN returns error", func(t *testing.T) {
		ext := external.AWSResource{
			ARN:          "not-an-arn",
			AccountRef:   "123456789012",
			ResourceType: model.AWSEC2Instance,
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})
}

func TestAWSResourceFromModel(t *testing.T) {
	fullResource, err := model.NewAWSResource(
		"arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
		"123456789012",
		model.AWSEC2Instance,
		map[string]any{
			"PublicIp": "54.123.45.67",
		},
	)
	require.NoError(t, err)
	fullResource.OrgPolicyFilename = "test-policy.json"

	ext := AWSResourceFromModel(&fullResource)

	assert.Equal(t, "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0", ext.ARN)
	assert.Equal(t, "123456789012", ext.AccountRef)
	assert.Equal(t, model.AWSEC2Instance, ext.ResourceType)
	assert.Equal(t, "test-policy.json", ext.OrgPolicyFilename)
	assert.NotNil(t, ext.Properties)
	assert.Equal(t, "54.123.45.67", ext.Properties["PublicIp"])
}

func TestPreseedFromModel(t *testing.T) {
	modelPreseed := model.NewPreseed("whois", "registrant_email", "test@example.com")
	modelPreseed.Display = "text"
	modelPreseed.Status = "A"
	modelPreseed.Capability = "whois-lookup"
	modelPreseed.Metadata = map[string]string{"source": "manual"}

	extPreseed := PreseedFromModel(&modelPreseed)

	if extPreseed.Type != modelPreseed.Type {
		t.Errorf("Type = %v, want %v", extPreseed.Type, modelPreseed.Type)
	}
	if extPreseed.Title != modelPreseed.Title {
		t.Errorf("Title = %v, want %v", extPreseed.Title, modelPreseed.Title)
	}
	if extPreseed.Value != modelPreseed.Value {
		t.Errorf("Value = %v, want %v", extPreseed.Value, modelPreseed.Value)
	}
	if extPreseed.Display != modelPreseed.Display {
		t.Errorf("Display = %v, want %v", extPreseed.Display, modelPreseed.Display)
	}
	if extPreseed.Status != modelPreseed.Status {
		t.Errorf("Status = %v, want %v", extPreseed.Status, modelPreseed.Status)
	}
	if extPreseed.Capability != modelPreseed.Capability {
		t.Errorf("Capability = %v, want %v", extPreseed.Capability, modelPreseed.Capability)
	}
	if len(extPreseed.Metadata) != len(modelPreseed.Metadata) {
		t.Errorf("Metadata length = %v, want %v", len(extPreseed.Metadata), len(modelPreseed.Metadata))
	}
}
