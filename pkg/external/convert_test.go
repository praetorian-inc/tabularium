package external

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset(t *testing.T) {
	t.Run("ToModel creates valid asset", func(t *testing.T) {
		ext := Asset{DNS: "example.com", Name: "192.168.1.1"}

		asset, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, asset)

		assert.Equal(t, "example.com", asset.DNS)
		assert.Equal(t, "192.168.1.1", asset.Name)
		assert.Contains(t, asset.Key, "#asset#")
		assert.Equal(t, model.Active, asset.Status)
	})

	t.Run("implements Target interface", func(t *testing.T) {
		ext := Asset{DNS: "example.com", Name: "192.168.1.1"}

		assert.Equal(t, "example.com", ext.Group())
		assert.Equal(t, "192.168.1.1", ext.Identifier())

		target, err := ext.ToTarget()
		require.NoError(t, err)
		assert.IsType(t, &model.Asset{}, target)
	})

	t.Run("empty asset returns error", func(t *testing.T) {
		ext := Asset{}
		_, err := ext.ToModel()
		require.Error(t, err)
	})
}

func TestPort(t *testing.T) {
	t.Run("ToModel creates valid port", func(t *testing.T) {
		ext := Port{
			Protocol: "tcp",
			Port:     443,
			Service:  "https",
			Parent:   Asset{DNS: "example.com", Name: "192.168.1.1"},
		}

		port, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, port)

		assert.Equal(t, "tcp", port.Protocol)
		assert.Equal(t, 443, port.Port)
		assert.Equal(t, "https", port.Service)
		assert.Contains(t, port.Key, "#port#")
	})

	t.Run("implements Target interface", func(t *testing.T) {
		ext := Port{
			Protocol: "tcp",
			Port:     443,
			Parent:   Asset{DNS: "example.com", Name: "192.168.1.1"},
		}

		assert.Equal(t, "example.com", ext.Group())
		assert.Equal(t, "192.168.1.1:443", ext.Identifier())

		target, err := ext.ToTarget()
		require.NoError(t, err)
		assert.IsType(t, &model.Port{}, target)
	})

	t.Run("invalid port number returns error", func(t *testing.T) {
		ext := Port{
			Protocol: "tcp",
			Port:     0,
			Parent:   Asset{DNS: "example.com", Name: "192.168.1.1"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})

	t.Run("missing protocol returns error", func(t *testing.T) {
		ext := Port{
			Port:   443,
			Parent: Asset{DNS: "example.com", Name: "192.168.1.1"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})
}

func TestRisk(t *testing.T) {
	t.Run("ToModel with asset target", func(t *testing.T) {
		ext := Risk{
			Name:   "CVE-2023-12345",
			Status: "TH",
			Target: Asset{DNS: "example.com", Name: "192.168.1.1"},
		}

		risk, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, risk)

		assert.Equal(t, "CVE-2023-12345", risk.Name)
		assert.Equal(t, "TH", risk.Status)
		assert.Equal(t, "example.com", risk.DNS)
		assert.Contains(t, risk.Key, "#risk#")
	})

	t.Run("ToModel with port target", func(t *testing.T) {
		ext := Risk{
			Name:   "ssl-weak-cipher",
			Status: "OH",
			Target: Port{
				Protocol: "tcp",
				Port:     443,
				Parent:   Asset{DNS: "example.com", Name: "192.168.1.1"},
			},
		}

		risk, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, risk)

		assert.Equal(t, "ssl-weak-cipher", risk.Name)
		assert.Equal(t, "OH", risk.Status)
		assert.Contains(t, risk.Key, "#risk#")
	})

	t.Run("default status when empty", func(t *testing.T) {
		ext := Risk{
			Name:   "CVE-2023-12345",
			Target: Asset{DNS: "example.com", Name: "example.com"},
		}

		risk, err := ext.ToModel()
		require.NoError(t, err)
		assert.Equal(t, "TH", risk.Status) // Default to Triage High
	})

	t.Run("missing name returns error", func(t *testing.T) {
		ext := Risk{
			Status: "TH",
			Target: Asset{DNS: "example.com", Name: "192.168.1.1"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
	})

	t.Run("missing target returns error", func(t *testing.T) {
		ext := Risk{
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

func TestPortFromModel(t *testing.T) {
	asset := model.NewAsset("example.com", "192.168.1.1")
	fullPort := model.NewPort("tcp", 443, &asset)
	fullPort.Service = "https"

	ext := PortFromModel(&fullPort)

	assert.Equal(t, "tcp", ext.Protocol)
	assert.Equal(t, 443, ext.Port)
	assert.Equal(t, "https", ext.Service)
	assert.Equal(t, "example.com", ext.Parent.DNS)
	assert.Equal(t, "192.168.1.1", ext.Parent.Name)
}

func TestAccount(t *testing.T) {
	t.Run("ToModel creates valid account", func(t *testing.T) {
		ext := Account{
			Name:   "customer@example.com",
			Member: "amazon",
			Value:  "123456789012",
			Secret: map[string]string{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
		}

		account, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, account)

		assert.Equal(t, "customer@example.com", account.Name)
		assert.Equal(t, "amazon", account.Member)
		assert.Equal(t, "123456789012", account.Value)
		assert.Equal(t, ext.Secret, account.Secret)
		assert.Contains(t, account.Key, "#account#")
		assert.NotEmpty(t, account.Updated)
	})

	t.Run("ToModel with settings", func(t *testing.T) {
		settings := []byte(`{"region": "us-east-1", "notifications": true}`)
		ext := Account{
			Name:     "customer@example.com",
			Member:   "azure",
			Value:    "subscription-id-12345",
			Secret:   map[string]string{"client_secret": "secret123"},
			Settings: settings,
		}

		account, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, account)

		assert.Equal(t, "customer@example.com", account.Name)
		assert.Equal(t, "azure", account.Member)
		assert.Equal(t, "subscription-id-12345", account.Value)
		assert.Equal(t, settings, []byte(account.Settings))
	})

	t.Run("missing name returns error", func(t *testing.T) {
		ext := Account{
			Member: "amazon",
			Value:  "123456789012",
			Secret: map[string]string{"key": "value"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("missing member returns error", func(t *testing.T) {
		ext := Account{
			Name:   "customer@example.com",
			Value:  "123456789012",
			Secret: map[string]string{"key": "value"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "member")
	})

	t.Run("missing value returns error", func(t *testing.T) {
		ext := Account{
			Name:   "customer@example.com",
			Member: "amazon",
			Secret: map[string]string{"key": "value"},
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value")
	})

	t.Run("empty secret is allowed", func(t *testing.T) {
		ext := Account{
			Name:   "customer@example.com",
			Member: "amazon",
			Value:  "123456789012",
		}
		account, err := ext.ToModel()
		require.NoError(t, err)
		assert.NotNil(t, account)
	})
}

func TestAccountFromModel(t *testing.T) {
	settings := []byte(`{"region": "us-west-2"}`)
	fullAccount := model.NewAccountWithSettings(
		"customer@example.com",
		"amazon",
		"123456789012",
		map[string]string{"access_key": "AKIATEST"},
		settings,
	)

	ext := AccountFromModel(&fullAccount)

	assert.Equal(t, "customer@example.com", ext.Name)
	assert.Equal(t, "amazon", ext.Member)
	assert.Equal(t, "123456789012", ext.Value)
	assert.Equal(t, map[string]string{"access_key": "AKIATEST"}, ext.Secret)
	assert.Equal(t, settings, ext.Settings)
}

func TestAWSResource(t *testing.T) {
	t.Run("ToModel creates valid AWS resource", func(t *testing.T) {
		ext := AWSResource{
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
		ext := AWSResource{
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
		ext := AWSResource{
			AccountRef:   "123456789012",
			ResourceType: model.AWSEC2Instance,
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "arn")
	})

	t.Run("missing AccountRef returns error", func(t *testing.T) {
		ext := AWSResource{
			ARN:          "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			ResourceType: model.AWSEC2Instance,
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "accountRef")
	})

	t.Run("preserves OrgPolicyFilename", func(t *testing.T) {
		ext := AWSResource{
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
		ext := AWSResource{
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

func TestWebApplication(t *testing.T) {
	t.Run("ToModel creates valid web application", func(t *testing.T) {
		ext := WebApplication{
			PrimaryURL: "https://app.example.com",
			Name:       "Example App",
			URLs:       []string{"https://api.example.com", "https://admin.example.com"},
			Status:     model.Active,
		}

		webapp, err := ext.ToModel()
		require.NoError(t, err)
		require.NotNil(t, webapp)

		// URLs are normalized with trailing slash
		assert.Equal(t, "https://app.example.com/", webapp.PrimaryURL)
		assert.Equal(t, "Example App", webapp.Name)
		assert.Equal(t, 2, len(webapp.URLs))
		assert.Contains(t, webapp.URLs, "https://api.example.com/")
		assert.Contains(t, webapp.URLs, "https://admin.example.com/")
		assert.Equal(t, model.Active, webapp.Status)
		assert.Contains(t, webapp.Key, "#webapplication#")
	})

	t.Run("implements Target interface", func(t *testing.T) {
		ext := WebApplication{
			PrimaryURL: "https://app.example.com",
			Name:       "Example App",
		}

		assert.Equal(t, "Example App", ext.Group())
		assert.Equal(t, "https://app.example.com", ext.Identifier())

		target, err := ext.ToTarget()
		require.NoError(t, err)
		assert.IsType(t, &model.WebApplication{}, target)
	})

	t.Run("defaults name to primary URL when name is empty", func(t *testing.T) {
		ext := WebApplication{
			PrimaryURL: "https://app.example.com",
		}

		webapp, err := ext.ToModel()
		require.NoError(t, err)
		// Name defaults to the input URL (before normalization)
		assert.Equal(t, "https://app.example.com", webapp.Name)
	})

	t.Run("missing primary URL returns error", func(t *testing.T) {
		ext := WebApplication{
			Name: "Example App",
		}
		_, err := ext.ToModel()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "primary_url")
	})

	t.Run("normalizes URLs", func(t *testing.T) {
		ext := WebApplication{
			PrimaryURL: "http://example.com:80",
			Name:       "Example",
		}

		webapp, err := ext.ToModel()
		require.NoError(t, err)
		// NewWebApplication will normalize the URL
		assert.NotEmpty(t, webapp.PrimaryURL)
	})
}

func TestWebApplicationFromModel(t *testing.T) {
	modelWebApp := model.NewWebApplication("https://app.example.com", "Example App")
	modelWebApp.URLs = []string{"https://api.example.com/"}
	modelWebApp.Status = model.Active

	ext := WebApplicationFromModel(&modelWebApp)

	// URLs are normalized in the model, so they'll have trailing slash
	assert.Equal(t, "https://app.example.com/", ext.PrimaryURL)
	assert.Equal(t, "Example App", ext.Name)
	assert.Equal(t, 1, len(ext.URLs))
	assert.Contains(t, ext.URLs, "https://api.example.com/")
	assert.Equal(t, model.Active, ext.Status)
}
