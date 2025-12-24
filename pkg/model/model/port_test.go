package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPort(t *testing.T) {
	asset := NewAsset("example.com", "192.168.1.1")
	port := NewPort("tcp", 80, &asset)

	assert.Equal(t, "tcp", port.Protocol)
	assert.Equal(t, 80, port.Port)
	assert.Equal(t, Active, port.Status)
	assert.NotEmpty(t, port.Created)
	assert.NotEmpty(t, port.Visited)
	assert.Equal(t, "#port#tcp#80#asset#example.com#192.168.1.1", port.Key)
}

func TestPort_Target(t *testing.T) {
	tests := []struct {
		name     string
		port     Port
		expected string
	}{
		{
			name: "port without service",
			port: func() Port {
				asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#192.168.1.1"}}
				asset.Name = "192.168.1.1"
				return NewPort("tcp", 80, &asset)
			}(),
			expected: "192.168.1.1:80",
		},
		{
			name: "port with http service",
			port: func() Port {
				asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#192.168.1.1"}}
				asset.Name = "192.168.1.1"
				port := NewPort("tcp", 443, &asset)
				port.Service = "https"
				return port
			}(),
			expected: "https://example.com:443",
		},
		{
			name: "port with ssh service",
			port: func() Port {
				asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#192.168.1.1"}}
				asset.Name = "192.168.1.1"
				port := NewPort("tcp", 443, &asset)
				port.Service = "ssh"
				return port
			}(),
			expected: "192.168.1.1:443",
		},
	}

	for _, test := range tests {
		actual := test.port.Target()
		assert.Equal(t, test.expected, actual, "test case %s failed", test.name)
	}
}

func TestPort_Valid(t *testing.T) {
	tests := []struct {
		name string
		port Port
		want bool
	}{
		{
			name: "valid port",
			port: func() Port {
				asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#example.com"}}
				return NewPort("tcp", 80, &asset)
			}(),
			want: true,
		},
		{
			name: "invalid port - zero",
			port: Port{Port: 0, Key: "test"},
			want: false,
		},
		{
			name: "invalid port - too high",
			port: Port{Port: 65536, Key: "test"},
			want: false,
		},
		{
			name: "invalid port - no key",
			port: Port{Port: 80},
			want: false,
		},
	}

	for _, test := range tests {
		actual := test.port.Valid()
		assert.Equal(t, test.want, actual, "test case %s failed", test.name)
	}
}

func TestPort_Asset(t *testing.T) {
	parentAsset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#192.168.1.1"}}
	parentAsset.DNS = "example.com"
	parentAsset.Name = "192.168.1.1"
	port := NewPort("tcp", 80, &parentAsset)

	asset := port.Asset()
	assert.Equal(t, "example.com", asset.DNS)
	assert.Equal(t, "192.168.1.1", asset.Name)
}

func TestPort_IsClass(t *testing.T) {
	asset := Asset{}
	port := NewPort("tcp", 80, &asset)
	port.Service = "http"

	assert.True(t, port.IsClass("http"))
	assert.True(t, port.IsClass("80"))
	assert.False(t, port.IsClass("ssh"))
}

func TestPort_Visit(t *testing.T) {
	tests := []struct {
		name     string
		existing Port
		update   Port
		validate func(*testing.T, Port)
	}{
		{
			name: "basic visit updates status, service, and TTL",
			existing: func() Port {
				asset := Asset{}
				return NewPort("tcp", 80, &asset)
			}(),
			update: Port{
				Status:  "inactive",
				Service: "http",
				TTL:     12345,
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, "inactive", p.Status)
				assert.Equal(t, "http", p.Service)
				assert.Equal(t, int64(12345), p.TTL)
			},
		},
		{
			name: "visit propagates tags without duplicates",
			existing: func() Port {
				asset := Asset{}
				port := NewPort("tcp", 80, &asset)
				port.Tags = Tags{Tags: []string{"production", "web"}}
				return port
			}(),
			update: Port{
				Tags: Tags{Tags: []string{"critical", "monitored"}},
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, []string{"production", "web", "critical", "monitored"}, p.Tags.Tags)
			},
		},
		{
			name: "visit with duplicate tags only adds new ones",
			existing: func() Port {
				asset := Asset{}
				port := NewPort("tcp", 443, &asset)
				port.Tags = Tags{Tags: []string{"production", "web"}}
				return port
			}(),
			update: Port{
				Tags: Tags{Tags: []string{"production", "critical"}},
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, []string{"production", "web", "critical"}, p.Tags.Tags)
			},
		},
		{
			name: "visit with empty tags preserves existing",
			existing: func() Port {
				asset := Asset{}
				port := NewPort("tcp", 22, &asset)
				port.Tags = Tags{Tags: []string{"ssh", "admin"}}
				return port
			}(),
			update: Port{
				Status: Active,
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, []string{"ssh", "admin"}, p.Tags.Tags)
			},
		},
		{
			name: "visit updates service and propagates tags",
			existing: func() Port {
				asset := Asset{}
				port := NewPort("tcp", 3306, &asset)
				port.Tags = Tags{Tags: []string{"database"}}
				return port
			}(),
			update: Port{
				Service: "mysql",
				Tags:    Tags{Tags: []string{"production", "critical"}},
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, "mysql", p.Service)
				assert.Equal(t, []string{"database", "production", "critical"}, p.Tags.Tags)
			},
		},
		{
			name: "visit does not update status when pending",
			existing: func() Port {
				asset := Asset{}
				port := NewPort("tcp", 80, &asset)
				port.Status = Active
				return port
			}(),
			update: Port{
				Status: Pending,
			},
			validate: func(t *testing.T, p Port) {
				assert.Equal(t, Active, p.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.existing.Visit(tt.update)
			tt.validate(t, tt.existing)
		})
	}
}

func TestPortConditions(t *testing.T) {
	asset := Asset{}
	port := NewPort("tcp", 80, &asset)
	port.Service = "http"

	conditions := PortConditions(port)

	assert.Len(t, conditions, 3)
	assert.Equal(t, "port", conditions[0].Name)
	assert.Equal(t, "80", conditions[0].Value)
	assert.Equal(t, "protocol", conditions[1].Name)
	assert.Equal(t, "http", conditions[1].Value)
	assert.Equal(t, "http", conditions[2].Name)
	assert.Equal(t, "", conditions[2].Value)
}

func TestPort_GetPartitionKey(t *testing.T) {
	// Test that different ports on the same IP have the same partition key
	asset := NewAsset("example.com", "192.168.1.1")
	port80 := NewPort("tcp", 80, &asset)
	port443 := NewPort("tcp", 443, &asset)
	port8080 := NewPort("tcp", 8080, &asset)

	partition80 := port80.GetPartitionKey()
	partition443 := port443.GetPartitionKey()
	partition8080 := port8080.GetPartitionKey()

	// All ports on the same IP should have the same partition key
	assert.Equal(t, "192.168.1.1", partition80, "port 80 should use parent asset IP")
	assert.Equal(t, "192.168.1.1", partition443, "port 443 should use parent asset IP")
	assert.Equal(t, "192.168.1.1", partition8080, "port 8080 should use parent asset IP")
	assert.Equal(t, partition80, partition443, "port 80 and 443 should have same partition")
	assert.Equal(t, partition80, partition8080, "port 80 and 8080 should have same partition")
}

func TestPort_GetPartitionKey_DifferentIPs(t *testing.T) {
	// Test that ports on different IPs have different partition keys
	asset1 := NewAsset("example.com", "192.168.1.1")
	asset2 := NewAsset("test.com", "10.0.0.1")

	port1 := NewPort("tcp", 80, &asset1)
	port2 := NewPort("tcp", 80, &asset2)

	partition1 := port1.GetPartitionKey()
	partition2 := port2.GetPartitionKey()

	assert.Equal(t, "192.168.1.1", partition1, "port1 should use first IP")
	assert.Equal(t, "10.0.0.1", partition2, "port2 should use second IP")
	assert.NotEqual(t, partition1, partition2, "different IPs should have different partitions")
}

func TestPort_GetPartitionKey_SameIPDifferentDNS(t *testing.T) {
	// Critical test: different DNS names pointing to the same IP
	// should result in the same partition key for all their ports
	asset1 := NewAsset("example.com", "192.168.1.1")
	asset2 := NewAsset("test.com", "192.168.1.1")
	asset3 := NewAsset("another.com", "192.168.1.1")

	port1 := NewPort("tcp", 80, &asset1)
	port2 := NewPort("tcp", 443, &asset2)
	port3 := NewPort("tcp", 8080, &asset3)

	partition1 := port1.GetPartitionKey()
	partition2 := port2.GetPartitionKey()
	partition3 := port3.GetPartitionKey()

	// All ports should partition by the shared IP address, not by DNS
	assert.Equal(t, "192.168.1.1", partition1, "port on example.com should use IP")
	assert.Equal(t, "192.168.1.1", partition2, "port on test.com should use IP")
	assert.Equal(t, "192.168.1.1", partition3, "port on another.com should use IP")
	assert.Equal(t, partition1, partition2, "ports on different DNS but same IP should partition together")
	assert.Equal(t, partition1, partition3, "all ports on same IP should partition together")
}

func TestPort_GetPartitionKey_MatchesParentAsset(t *testing.T) {
	// Verify that port partition key matches the parent asset's partition key
	asset := NewAsset("example.com", "192.168.1.1")
	port := NewPort("tcp", 443, &asset)

	assetPartition := asset.GetPartitionKey()
	portPartition := port.GetPartitionKey()

	assert.Equal(t, assetPartition, portPartition,
		"port partition key should match parent asset partition key")
}
