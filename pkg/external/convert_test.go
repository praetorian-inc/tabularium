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
