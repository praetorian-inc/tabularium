package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPort(t *testing.T) {
	asset := NewAsset("example.com", "192.168.1.1")
	port := NewPort("tcp", 80, &asset)

	assert.Equal(t, "tcp", port.Protocol)
	assert.Equal(t, 80, port.PortNumber)
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
			name: "port with service",
			port: func() Port {
				asset := Asset{BaseAsset: BaseAsset{Key: "#asset#example.com#192.168.1.1"}}
				asset.Name = "192.168.1.1"
				port := NewPort("tcp", 443, &asset)
				port.Service = "https"
				return port
			}(),
			expected: "https://192.168.1.1:443",
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
			port: Port{PortNumber: 0, Key: "test"},
			want: false,
		},
		{
			name: "invalid port - too high",
			port: Port{PortNumber: 65536, Key: "test"},
			want: false,
		},
		{
			name: "invalid port - no key",
			port: Port{PortNumber: 80},
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
	asset := Asset{}
	port1 := NewPort("tcp", 80, &asset)
	port2 := Port{
		Status:  "inactive", 
		Service: "http",
		TTL:     12345,
	}

	port1.Visit(port2)

	assert.Equal(t, "inactive", port1.Status)
	assert.Equal(t, "http", port1.Service)
	assert.Equal(t, int64(12345), port1.TTL)
}

func TestPortConditions(t *testing.T) {
	asset := Asset{}
	port := NewPort("tcp", 80, &asset)
	
	conditions := PortConditions(port)
	
	assert.Len(t, conditions, 2)
	assert.Equal(t, "port", conditions[0].Name)
	assert.Equal(t, "", conditions[0].Value)
	assert.Equal(t, "port", conditions[1].Name)
	assert.Equal(t, "80", conditions[1].Value)
}