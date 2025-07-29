package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSeed_GetClassType(t *testing.T) {
	tests := []struct {
		name      string
		dns       string
		wantClass string
		wantType  string
	}{
		{
			name:      "ipv4",
			dns:       "192.168.1.1",
			wantClass: "ip",
			wantType:  "ip",
		},
		{
			name:      "ipv6",
			dns:       "2001:db8::1",
			wantClass: "ip",
			wantType:  "ip",
		},
		{
			name:      "domain",
			dns:       "sub.example.com",
			wantClass: "domain",
			wantType:  "domain",
		},
		{
			name:      "tld",
			dns:       "example.com",
			wantClass: "tld",
			wantType:  "domain",
		},
		{
			name:      "cidr",
			dns:       "192.168.1.0/24",
			wantClass: "cidr",
			wantType:  "ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := NewSeed(tt.dns)
			assert.Equal(t, tt.wantClass, seed.GetClass())
			assert.Equal(t, tt.wantType, seed.GetType())
		})
	}
}

func TestSeed_StatusHandling(t *testing.T) {
	tests := []struct {
		name           string
		dns            string // Add DNS to control the seed type
		status         string
		expectedStatus string
		expectedGet    string
	}{
		{
			name:           "ip status",
			dns:            "192.168.1.1", // Use IP address to get IP type
			status:         "P",
			expectedStatus: "ip#P",
			expectedGet:    "P",
		},
		{
			name:           "domain status",
			dns:            "example.com", // Use domain to get domain type
			status:         "A",
			expectedStatus: "domain#A",
			expectedGet:    "A",
		},
		{
			name:           "empty status",
			dns:            "example.com",
			status:         "",
			expectedStatus: "domain#",
			expectedGet:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := NewSeed(tt.dns)
			// Don't manually override the type - let it be determined by the DNS

			// Test SetStatus
			seed.SetStatus(tt.status)
			if seed.Status != tt.expectedStatus {
				t.Errorf("SetStatus() got = %v, want %v", seed.Status, tt.expectedStatus)
			}

			// Test GetStatus
			got := seed.GetStatus()
			if got != tt.expectedGet {
				t.Errorf("GetStatus() got = %v, want %v", got, tt.expectedGet)
			}
		})
	}
}

func TestSeed_Merge(t *testing.T) {
	tests := []struct {
		name           string
		original       Seed
		update         Seed
		expectedStatus string
	}{
		{
			name: "Update status when provided",
			original: Seed{
				DNS:    "test.com",
				Key:    "#seed#domain#test.com",
				Status: "domain#pending",
			},
			update: Seed{
				DNS:    "test.com",
				Key:    "#seed#domain#test.com",
				Status: "domain#active",
			},
			expectedStatus: "domain#active",
		},
		{
			name: "No status update when empty",
			original: Seed{
				DNS:    "test.com",
				Key:    "#seed#domain#test.com",
				Status: "domain#pending",
			},
			update: Seed{
				DNS:    "test.com",
				Key:    "#seed#domain#test.com",
				Status: "",
			},
			expectedStatus: "domain#pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.original.Merge(tt.update)
			if tt.original.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", tt.original.Status, tt.expectedStatus)
			}
		})
	}
}

func TestSeed_Asset(t *testing.T) {
	tests := []struct {
		name          string
		seed          Seed
		expectedAsset Asset
	}{
		{
			name: "Convert seed to asset with status",
			seed: func() Seed {
				s := NewSeed("example.com")
				s.SetStatus(Active)
				return s
			}(),
			expectedAsset: Asset{
				DNS:  "example.com",
				Name: "example.com",
				BaseAsset: BaseAsset{
					Class:  "tld",
					Status: Active,
					Source: SeedSource,
				},
			},
		},
		{
			name: "Convert seed to asset with pending status",
			seed: NewSeed("example.com"),
			expectedAsset: Asset{
				DNS:  "example.com",
				Name: "example.com",
				BaseAsset: BaseAsset{
					Class:  "tld",
					Status: Pending,
					Source: SeedSource,
				},
			},
		},
		{
			name: "subdomain seed",
			seed: NewSeed("sub.example.com"),
			expectedAsset: Asset{
				DNS:  "sub.example.com",
				Name: "sub.example.com",
				BaseAsset: BaseAsset{
					Class:  "domain",
					Status: Pending,
					Source: SeedSource,
				},
			},
		},
		{
			name: "ipv4 seed",
			seed: NewSeed("192.168.1.1"),
			expectedAsset: Asset{
				DNS:  "192.168.1.1",
				Name: "192.168.1.1",
				BaseAsset: BaseAsset{
					Class:  "ipv4",
					Status: Pending,
					Source: SeedSource,
				},
			},
		},
		{
			name: "ipv6 seed",
			seed: NewSeed("2001:db8::1"),
			expectedAsset: Asset{
				DNS:  "2001:db8::1",
				Name: "2001:db8::1",
				BaseAsset: BaseAsset{
					Class:  "ipv6",
					Status: Pending,
					Source: SeedSource,
				},
			},
		},
		{
			name: "cidr seed",
			seed: NewSeed("192.168.1.0/24"),
			expectedAsset: Asset{
				DNS:  "192.168.1.0/24",
				Name: "192.168.1.0/24",
				BaseAsset: BaseAsset{
					Class:  "cidr",
					Status: Pending,
					Source: SeedSource,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := tt.seed.Asset()

			// Check core fields
			assert.Equal(t, tt.expectedAsset.DNS, asset.DNS)
			assert.Equal(t, tt.expectedAsset.Name, asset.Name)
			assert.Equal(t, tt.expectedAsset.Status, asset.Status)
			assert.Equal(t, tt.expectedAsset.Source, asset.Source)
			assert.Zero(t, asset.TTL)
			assert.Equal(t, tt.expectedAsset.DNS, asset.GetBase().Group)
			assert.Equal(t, tt.expectedAsset.Name, asset.GetBase().Identifier)
			assert.Equal(t, tt.expectedAsset.Class, asset.GetBase().Class)

			// Verify timestamps exist and are valid
			_, err := time.Parse(time.RFC3339, asset.Created)
			assert.NoError(t, err)

			_, err = time.Parse(time.RFC3339, asset.Visited)
			assert.NoError(t, err)
		})
	}
}

func TestSeed_DomainVerificationJob(t *testing.T) {
	tests := []struct {
		name   string
		seed   Seed
		config []string
		want   map[string]string
	}{
		{
			name: "Basic domain verification job",
			seed: func() Seed {
				s := NewSeed("example.com")
				s.SetStatus(Pending)
				return s
			}(),
			config: []string{"source", "test-source"},
			want: map[string]string{
				"source": "test-source",
			},
		},
		{
			name: "Domain verification job with multiple config pairs",
			seed: func() Seed {
				s := NewSeed("example.com")
				s.SetStatus(Active)
				return s
			}(),
			config: []string{"source", "test-source", "key", "value"},
			want: map[string]string{
				"source": "test-source",
				"key":    "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := tt.seed.Asset()
			dummy := NewJob("whois", &asset)
			job := tt.seed.DomainVerificationJob(&dummy, tt.config...)

			assert.Equal(t, job.Config, tt.want)

			asset = tt.seed.Asset()
			assert.Equal(t, job.Target.Model.Group(), asset.DNS)
			assert.Equal(t, job.Target.Model.GetStatus(), asset.Status)
			assert.Equal(t, job.Source, "whois")
			assert.True(t, job.Full)
		})
	}
}

func TestAssetSeedStatusConversions(t *testing.T) {
	tests := []struct {
		name          string
		assetStatus   string
		seedStatus    string
		dns           string
		startWithSeed bool
	}{
		{
			name:          "asset->seed->asset: pending",
			assetStatus:   Pending,
			seedStatus:    "domain#P",
			dns:           "example.com",
			startWithSeed: false,
		},
		{
			name:          "asset->seed->asset: active",
			assetStatus:   Active,
			seedStatus:    "domain#A",
			dns:           "example.com",
			startWithSeed: false,
		},
		{
			name:          "seed->asset->seed: pending domain",
			assetStatus:   Pending,
			seedStatus:    "domain#P",
			dns:           "example.com",
			startWithSeed: true,
		},
		{
			name:          "seed->asset->seed: active IP",
			assetStatus:   Active,
			seedStatus:    "ip#A",
			dns:           "192.168.1.1",
			startWithSeed: true,
		},
		{
			name:          "seed->asset->seed: active CIDR",
			assetStatus:   Active,
			seedStatus:    "ip#A",
			dns:           "192.168.1.0/24",
			startWithSeed: true,
		},
		{
			name:          "asset->seed->asset: active high",
			assetStatus:   ActiveHigh,
			seedStatus:    "domain#AH",
			dns:           "example.com",
			startWithSeed: false,
		},
		{
			name:          "asset->seed->asset: frozen",
			assetStatus:   Frozen,
			seedStatus:    "domain#F",
			dns:           "example.com",
			startWithSeed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.startWithSeed {
				// Test seed->asset->seed conversion
				seed := NewSeed(tt.dns)
				seed.SetStatus(tt.assetStatus)

				// Convert seed to asset
				asset := seed.Asset()
				if asset.Status != tt.assetStatus {
					t.Errorf("seed.Asset().Status = %v, want %v", asset.Status, tt.assetStatus)
				}

				// Convert back to seed
				newSeed := asset.Seed()
				if newSeed.Status != tt.seedStatus {
					t.Errorf("asset.Seed().Status = %v, want %v", newSeed.Status, tt.seedStatus)
				}
			} else {
				// Test asset->seed->asset conversion
				asset := NewAsset(tt.dns, tt.dns)
				asset.Status = tt.assetStatus

				// Convert asset to seed
				seed := asset.Seed()
				if seed.Status != tt.seedStatus {
					t.Errorf("asset.Seed().Status = %v, want %v", seed.Status, tt.seedStatus)
				}

				// Convert back to asset
				newAsset := seed.Asset()
				if newAsset.Status != tt.assetStatus {
					t.Errorf("seed.Asset().Status = %v, want %v", newAsset.Status, tt.assetStatus)
				}
			}
		})
	}
}

func TestSeed_Valid(t *testing.T) {
	tests := []struct {
		name string
		dns  string
		want bool
	}{
		// Domain and TLD+1 cases
		{
			name: "valid TLD+1",
			dns:  "example.com",
			want: true,
		},
		{
			name: "valid subdomain",
			dns:  "sub.example.com",
			want: true,
		},
		{
			name: "invalid domain - no TLD",
			dns:  "Domain",
			want: false,
		},
		{
			name: "invalid domain - trailing dot",
			dns:  "sub.example.com.",
			want: false,
		},

		// IP cases
		{
			name: "valid IPv4",
			dns:  "192.168.1.1",
			want: true,
		},
		{
			name: "valid IPv6",
			dns:  "2001:db8::1",
			want: true,
		},

		// CIDR cases
		{
			name: "valid IPv4 CIDR",
			dns:  "192.168.1.0/24",
			want: true,
		},
		{
			name: "valid IPv6 CIDR",
			dns:  "2001:db8::/32",
			want: true,
		},
		{
			name: "invalid IPv4 CIDR - missing prefix",
			dns:  "192.168.1.0/",
			want: false,
		},
		{
			name: "invalid IPv4 CIDR - prefix too large",
			dns:  "192.168.1.0/33",
			want: false,
		},
		{
			name: "invalid IPv4 CIDR - prefix negative",
			dns:  "192.168.1.0/-1",
			want: false,
		},
		{
			name: "invalid IPv4 CIDR - malformed IP",
			dns:  "256.168.1.0/24",
			want: false,
		},
		{
			name: "invalid IPv6 CIDR - missing prefix",
			dns:  "2001:db8::/",
			want: false,
		},
		{
			name: "invalid IPv6 CIDR - prefix too large",
			dns:  "2001:db8::/129",
			want: false,
		},
		{
			name: "invalid IPv6 CIDR - malformed IP",
			dns:  "2001:zz8::/32",
			want: false,
		},
		{
			name: "invalid CIDR - random text with slash",
			dns:  "not-a-cidr/24",
			want: false,
		},

		// Edge cases
		{
			name: "empty string",
			dns:  "",
			want: false,
		},
		{
			name: "single comma",
			dns:  ",",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := NewSeed(tt.dns)
			if got := seed.Valid(); got != tt.want {
				t.Errorf("Seed.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeed_Case(t *testing.T) {
	seed1 := NewSeed("2001:db8::1/128")
	seed2 := NewSeed("2001:DB8::1/128")
	seed3 := NewSeed("CASEsensitive.com")

	assert.Equal(t, seed1.Key, seed2.Key, "Seeds should be case insensitive")
	assert.Equal(t, seed1.DNS, seed2.DNS, "Seeds DNS should be case insensitive")
	assert.Equal(t, "casesensitive.com", seed3.DNS, "Domain seeds should be case sensitive")
}
