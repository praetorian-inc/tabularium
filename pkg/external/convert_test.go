package external

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			// Key should be auto-generated
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
			// Key should be auto-generated
			assert.Contains(t, asset.Key, "#asset#")
		})
	}
}

func TestCIDRToAsset(t *testing.T) {
	// Note: CIDR assets require the CIDR to be in the DNS field for class detection.
	// The domain parameter is ignored for CIDR assets - both DNS and Name are set to the CIDR.
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
			// Key should be auto-generated
			assert.Contains(t, asset.Key, "#asset#")
		})
	}
}

func TestConvert(t *testing.T) {
	t.Run("convert external asset to tabularium asset", func(t *testing.T) {
		// Simulate an external Asset-like struct
		type ExternalAsset struct {
			DNS     string `json:"dns"`
			Name    string `json:"name"`
			Private bool   `json:"private"`
		}

		external := ExternalAsset{
			DNS:     "example.com",
			Name:    "192.168.1.1",
			Private: true,
		}

		asset, err := Convert[*model.Asset](external)
		require.NoError(t, err)
		require.NotNil(t, asset)

		assert.Equal(t, "example.com", asset.DNS)
		assert.Equal(t, "192.168.1.1", asset.Name)
		assert.Equal(t, true, asset.Private)
		// Key should be generated via hooks
		assert.Contains(t, asset.Key, "#asset#")
		// Status should be defaulted
		assert.Equal(t, model.Active, asset.Status)
	})

	t.Run("convert external risk to tabularium risk", func(t *testing.T) {
		type ExternalRisk struct {
			DNS    string `json:"dns"`
			Name   string `json:"name"`
			Status string `json:"status"`
		}

		external := ExternalRisk{
			DNS:    "example.com",
			Name:   "CVE-2023-12345",
			Status: "TH",
		}

		risk, err := Convert[*model.Risk](external)
		require.NoError(t, err)
		require.NotNil(t, risk)

		assert.Equal(t, "example.com", risk.DNS)
		assert.Equal(t, "CVE-2023-12345", risk.Name)
		assert.Equal(t, "TH", risk.Status)
		// Key should be generated via hooks
		assert.Contains(t, risk.Key, "#risk#")
	})

	t.Run("convert external job to tabularium job", func(t *testing.T) {
		type ExternalJob struct {
			Config       map[string]string `json:"config"`
			Capabilities []string          `json:"capabilities"`
			Full         bool              `json:"full"`
		}

		external := ExternalJob{
			Config:       map[string]string{"target": "192.168.1.1"},
			Capabilities: []string{"portscan", "nuclei"},
			Full:         true,
		}

		job, err := Convert[*model.Job](external)
		require.NoError(t, err)
		require.NotNil(t, job)

		assert.Equal(t, map[string]string{"target": "192.168.1.1"}, job.Config)
		assert.Equal(t, []string{"portscan", "nuclei"}, job.Capabilities)
		assert.Equal(t, true, job.Full)
		// Status should be defaulted to Queued
		assert.Contains(t, job.Status, model.Queued)
	})
}

func TestConvertToModelByName(t *testing.T) {
	t.Run("convert to asset by name", func(t *testing.T) {
		external := map[string]any{
			"dns":     "example.com",
			"name":    "test.example.com",
			"private": false,
		}

		result, err := ConvertToModelByName("asset", external)
		require.NoError(t, err)
		require.NotNil(t, result)

		asset, ok := result.(*model.Asset)
		require.True(t, ok)
		assert.Equal(t, "example.com", asset.DNS)
		assert.Equal(t, "test.example.com", asset.Name)
	})

	t.Run("unknown model name", func(t *testing.T) {
		external := map[string]any{}

		_, err := ConvertToModelByName("nonexistent", external)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not registered")
	})
}

func TestRegisteredTransformerPriority(t *testing.T) {
	// IP transformer should be called instead of JSON conversion
	ip := IP{
		Address: "8.8.8.8",
		Domain:  "google.com",
	}

	result, err := Convert[*model.Asset](ip)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "google.com", result.DNS)
	assert.Equal(t, "8.8.8.8", result.Name)
	assert.Equal(t, "ipv4", result.Class)
}
