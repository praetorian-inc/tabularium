package model

import (
	"log/slog"
	"time"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&AegisAgent{})
}

// CloudflaredStatus represents the cloudflared status information
type CloudflaredStatus struct {
	Status          string `json:"status,omitempty" desc:"High-level status (e.g., connected, not_found)" example:"connected" neo4j:"status"`
	Connected       string `json:"connected,omitempty" desc:"String flag indicating connectivity (legacy)" example:"true" neo4j:"connected"`
	ConnectionCount int    `json:"connection_count,omitempty" desc:"Active connection count" example:"4" neo4j:"connection_count"`
	TunnelName      string `json:"tunnel_name,omitempty" desc:"Configured Cloudflare Tunnel name" example:"corp-agent-tunnel" neo4j:"tunnel_name"`
	Hostname        string `json:"hostname,omitempty" desc:"Cloudflare hostname for the tunnel" example:"agent.example.com" neo4j:"hostname"`
	AuthorizedUsers string `json:"authorized_users,omitempty" desc:"Authorized users (format defined by agent)" example:"alice,bob" neo4j:"authorized_users"`
}

// AegisAgent represents an Aegis agent with all its information
type AegisAgent struct {
	BaseAsset                                 // required for models
	ClientID          string                  `json:"client_id" desc:"Unique agent identifier"`
	FirstSeenAt       int64                   `json:"first_seen_at" desc:"Unix timestamp (seconds) of first check-in"`
	LastSeenAt        int64                   `json:"last_seen_at" desc:"Unix timestamp (seconds) of last check-in"`
	Hostname          string                  `json:"hostname" desc:"Host short name"`
	FQDN              string                  `json:"fqdn" desc:"Fully-qualified domain name"`
	NetworkInterfaces []AegisNetworkInterface `json:"network_interfaces" desc:"Network interfaces and their IPs"`
	OS                string                  `json:"os" desc:"Operating system"`
	OSVersion         string                  `json:"os_version" desc:"Operating system version"`
	Architecture      string                  `json:"architecture" desc:"CPU architecture (e.g., amd64, arm64)"`
	HealthCheck       *AegisHealthCheckData   `json:"health_check,omitempty" desc:"Latest health check payload"`
}

// AegisNetworkInterface represents a network interface on an Aegis agent
type AegisNetworkInterface struct {
	Name        string   `json:"name" desc:"Interface name" example:"eth0"`
	IPAddresses []string `json:"ip_addresses" desc:"Interface IP addresses" example:"[\"10.0.0.5\",\"fe80::1\"]"`
}

// AegisHealthCheckData represents health check information from an Aegis agent
type AegisHealthCheckData struct {
	DiskSpace               int                `json:"disk_space,omitempty" desc:"Free disk space (GB)" example:"256"`
	Memory                  float64            `json:"memory,omitempty" desc:"System memory (GB)" example:"16"`
	VirtualizationSupported bool               `json:"virtualization_supported" desc:"Hardware virtualization support available" example:"true"`
	CloudFlare              bool               `json:"cloudflare" desc:"Cloudflare agent installed/configured" example:"true"`
	CloudflaredStatus       *CloudflaredStatus `json:"cloudflared_status,omitempty" desc:"Cloudflared process/tunnel status"`
}

// HasCloudflareConfigured returns true if the agent has Cloudflare tunnel configuration
func (a *AegisAgent) HasCloudflareConfigured() bool {
	if a.HealthCheck == nil || !a.HealthCheck.CloudFlare {
		slog.Debug("HasCloudflareConfigured: no health check or CloudFlare=false", "clientID", a.ClientID)
		return false
	}

	// Check if there's cloudflared status indicating an active tunnel
	if a.HealthCheck.CloudflaredStatus == nil {
		slog.Debug("HasCloudflareConfigured: no cloudflared status", "clientID", a.ClientID)
		return false
	}

	slog.Debug("HasCloudflareConfigured: checking cloudflared status", "clientID", a.ClientID)

	// Check if status indicates a tunnel is found and configured
	if a.HealthCheck.CloudflaredStatus.Status == "not_found" {
		slog.Debug("HasCloudflareConfigured: tunnel not found", "clientID", a.ClientID)
		return false
	}

	// Check if we have tunnel configuration (name and hostname)
	hasTunnelName := a.HealthCheck.CloudflaredStatus.TunnelName != ""
	hasHostname := a.HealthCheck.CloudflaredStatus.Hostname != ""

	result := hasTunnelName && hasHostname
	slog.Debug("HasCloudflareConfigured: checked tunnel configuration",
		"clientID", a.ClientID,
		"result", result,
		"tunnelName", a.HealthCheck.CloudflaredStatus.TunnelName,
		"hostname", a.HealthCheck.CloudflaredStatus.Hostname)

	return result
}

// IsOnline checks if the agent is currently online (last seen within 60 seconds)
func (a *AegisAgent) IsOnline() bool {
	if a.LastSeenAt <= 0 {
		return false
	}

	// Detect unit by magnitude: seconds < 1e11, ms < 1e14, µs >= 1e14
	var lastSeenTime time.Time
	switch {
	case a.LastSeenAt >= 100000000000000: // ≥1e14 ≈ microseconds since epoch
		lastSeenTime = time.Unix(0, a.LastSeenAt*int64(time.Microsecond))
	case a.LastSeenAt >= 100000000000: // ≥1e11 ≈ milliseconds since epoch
		lastSeenTime = time.Unix(0, a.LastSeenAt*int64(time.Millisecond))
	default: // seconds
		lastSeenTime = time.Unix(a.LastSeenAt, 0)
	}

	// Calculate time difference
	now := time.Now()
	timeDiff := now.Sub(lastSeenTime)

	// Handle future timestamps defensively - treat as online if within threshold
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// Return true if within 60 seconds
	return timeDiff <= 60*time.Second
}

// GetPrimaryIP returns the first non-localhost IP address from network interfaces
func (a *AegisAgent) GetPrimaryIP() string {
	for _, iface := range a.NetworkInterfaces {
		for _, ip := range iface.IPAddresses {
			if !isLocalhost(ip) {
				return ip
			}
		}
	}
	return ""
}

// isLocalhost checks if an IP address is localhost
func isLocalhost(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}
