package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset_Class(t *testing.T) {
	tests := []struct {
		dns         string
		name        string
		source      string
		want        string
		wantPrivate bool
	}{
		// SeedSource
		{"example.com", "example.com", SeedSource, "tld", false},
		{"0.0.0.0/8", "0.0.0.0/8", SeedSource, "cidr", false},

		// AccountSource
		{"github", "github.com/example-inc", AccountSource, "github", false},
		{"gitlab", "gitlab.com/example-inc", AccountSource, "gitlab", false},
		{"burp-enterprise", "7a27e7a8.portswigger.cloud", AccountSource, "burp-enterprise", false},
		{"tenablevm", "https://cloud.tenable.com", AccountSource, "tenablevm", false},

		// SelfSource
		{"subdomain.example.com", "subdomain.example.com", SelfSource, "domain", false},
		{"example.com", "0.0.0.0", SelfSource, "ipv4", false},
		{"example.com", "2001::0000", SelfSource, "ipv6", false},
		{"example.com", "192.168.0.1", SelfSource, "ipv4", true},
		// {"https://github.com/example-inc/example-repo", "example-repo", SelfSource, "repository", false},
		// {"https://gitlab.com/example-inc/example-repo", "example-repo", SelfSource, "repository", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAsset(tt.dns, tt.name)
			a.Source = tt.source
			if got := a.GetClass(); got != tt.want {
				t.Errorf("key: %s, GetClass(): %s, want: %s", a.Key, got, tt.want)
			}
			if gotPrivate := a.Private; gotPrivate != tt.wantPrivate {
				t.Errorf("key: %s, Private: %t, want: %t", a.Key, gotPrivate, tt.wantPrivate)
			}
		})
	}
}

func TestAsset_Valid(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "valid asset key with domain and IP",
			key:  "#asset#registry.prod01.example.infra-host.com#172.16.254.1",
			want: true,
		},
		{
			name: "valid asset key with IP",
			key:  "#asset#203.0.113.42#203.0.113.42",
			want: true,
		},
		{
			name: "invalid - missing asset prefix",
			key:  "#registry#db.example.com#10.0.0.1",
			want: false,
		},
		{
			name: "invalid - attribute key",
			key:  "#attribute#https#443#asset#db.example.com#10.0.0.1",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(tt.key, "#")
			a := NewAsset(parts[2], parts[3])
			a.Key = tt.key

			assert.Equal(t, tt.want, a.Valid(), "Asset.Valid() = %v, want %v", a.Valid(), tt.want)
		})
	}
}

func TestAsset_IsClass(t *testing.T) {
	tests := []struct {
		name  string
		asset Asset
		value string
		want  bool
	}{
		{
			name: "matches class",
			asset: Asset{
				DNS:  "example.com",
				Name: "example.com",
			},
			value: "tld",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.asset.Class = tt.asset.GetClass()
			actual := tt.asset.IsClass(tt.value)
			assert.Equal(t, tt.want, actual, "Asset.Is(%v) = %v, want %v", tt.value, actual, tt.want)
		})
	}
}

func TestAsset_Case(t *testing.T) {
	asset1 := NewAsset("2001:db8::1", "2001:db8::1")
	asset2 := NewAsset("2001:DB8::1", "2001:DB8::1")
	webpageAsset := NewAsset("https://foobar.com", "/WEB-INF")

	assert.Equal(t, asset1.Key, asset2.Key, "Keys should be case insensitive")
	assert.NotEqual(t, asset1.DNS, asset2.DNS, "DNS should be case sensitive")
	assert.Contains(t, webpageAsset.Key, "web-inf", "Keys should be case insensitive")
	assert.Contains(t, webpageAsset.Name, "WEB-INF", "DNS should be case sensitive")
}

func TestAsset_IsPrivate(t *testing.T) {
	tests := []struct {
		name  string
		asset Asset
		want  bool
	}{
		{
			name:  "private ip",
			asset: NewAsset("10.0.0.1", "10.0.0.1"),
			want:  true,
		},
		{
			name:  "public ip",
			asset: NewAsset("1.1.1.1", "1.1.1.1"),
			want:  false,
		},
		{
			name:  "private cidr",
			asset: NewAsset("10.0.0.0/8", "10.0.0.0/8"),
			want:  true,
		},
		{
			name:  "public cidr",
			asset: NewAsset("1.1.1.0/24", "1.1.1.0/24"),
			want:  false,
		},
		{
			name:  "domain",
			asset: NewAsset("subdomain.example.com", "subdomain.example.com"),
			want:  false,
		},
		{
			name:  "tld",
			asset: NewAsset("example.com", "example.com"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.asset.IsPrivate()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAsset_Unmarshall(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		valid bool
	}{
		{
			name:  "valid asset",
			data:  `{"type": "asset", "dns": "example.com", "name": "example.com"}`,
			valid: true,
		},
		{
			name:  "valid asset - group and identifier",
			data:  `{"type": "asset", "group": "example.com", "identifier": "example.com"}`,
			valid: true,
		},
		{
			name:  "invalid asset - missing dns and name",
			data:  `{"type": "asset"}`,
			valid: false,
		},
		{
			name:  "invalid asset - missing dns or name",
			data:  `{"type": "asset", "dns": "example.com", "group": "example.com"}`,
			valid: false,
		},
		{
			name:  "invalid asset - missing group or identifier",
			data:  `{"type": "asset", "name": "example.com", "identifier": "example.com"}`,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a registry.Wrapper[Assetlike]
			err := json.Unmarshal([]byte(tt.data), &a)
			require.NoError(t, err)

			registry.CallHooks(a.Model)
			assert.Equal(t, tt.valid, a.Model.Valid())
		})
	}
}

func TestAsset_SeedModels(t *testing.T) {
	seedAsset := NewAssetSeed("example.com")
	seedModels := seedAsset.SeedModels()

	assert.Equal(t, 1, len(seedModels))
	assert.Equal(t, &seedAsset, seedModels[0])
	assert.Contains(t, seedAsset.GetLabels(), SeedLabel)
}

func TestAsset_DomainVerificationJob(t *testing.T) {
	tests := []struct {
		name   string
		seed   Asset
		config []string
		want   map[string]string
	}{
		{
			name:   "Basic domain verification job",
			seed:   NewAssetSeed("example.com"),
			config: []string{"source", "test-source"},
			want: map[string]string{
				"source": "test-source",
			},
		},
		{
			name: "Domain verification job with multiple config pairs",
			seed: func() Asset {
				s := NewAssetSeed("example.com")
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
			dummy := NewJob("whois", &tt.seed)
			job := tt.seed.DomainVerificationJob(&dummy, tt.config...)

			assert.Equal(t, job.Config, tt.want)

			assert.Equal(t, job.Target.Model.Group(), tt.seed.DNS)
			assert.Equal(t, job.Target.Model.GetStatus(), tt.seed.Status)
			assert.Equal(t, job.Source, "whois")
			assert.True(t, job.Full)
		})
	}
}
