package model

import (
	"log/slog"
	"time"
)

// CloudflaredStatus represents the cloudflared status information
type CloudflaredStatus struct {
	Status          string `json:"status,omitempty"`
	Connected       string `json:"connected,omitempty"`
	ConnectionCount int    `json:"connection_count,omitempty"`
	TunnelName      string `json:"tunnel_name,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	AuthorizedUsers string `json:"authorized_users,omitempty"`
}

// AegisAgent represents an Aegis agent with all its information
type AegisAgent struct {
	ClientID          string                  `json:"client_id"`
	LastSeenAt        int64                   `json:"last_seen_at"`
	Hostname          string                  `json:"hostname"`
	FQDN              string                  `json:"fqdn"`
	NetworkInterfaces []AegisNetworkInterface `json:"network_interfaces"`
	OS                string                  `json:"os"`
	OSVersion         string                  `json:"os_version"`
	Architecture      string                  `json:"architecture"`
	HealthCheck       *AegisHealthCheckData   `json:"health_check,omitempty"`
}

// AegisNetworkInterface represents a network interface on an Aegis agent
type AegisNetworkInterface struct {
	Name        string   `json:"name"`
	IPAddresses []string `json:"ip_addresses"`
}

// AegisHealthCheckData represents health check information from an Aegis agent
type AegisHealthCheckData struct {
	DiskSpace               int                `json:"disk_space,omitempty"`
	Memory                  float64            `json:"memory,omitempty"`
	VirtualizationSupported bool               `json:"virtualization_supported"`
	CloudFlare              bool               `json:"cloudflare"`
	CloudflaredStatus       *CloudflaredStatus `json:"cloudflared_status,omitempty"`
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

	// Determine if LastSeenAt is in microseconds or seconds
	// If > 1 trillion, it's in microseconds; otherwise it's in seconds
	var lastSeenTime time.Time
	if a.LastSeenAt > 1000000000000 {
		// Convert microseconds to time.Time
		lastSeenTime = time.Unix(0, a.LastSeenAt*int64(time.Microsecond))
	} else {
		// Convert seconds to time.Time
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
