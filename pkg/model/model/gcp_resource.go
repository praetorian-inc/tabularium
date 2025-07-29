package model

import (
	"fmt"
	"maps"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type GCPResource struct {
	CloudResource
}

func NewGCPResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (GCPResource, error) {
	key := fmt.Sprintf("#gcpresource#%s#%s", accountRef, name)

	r := GCPResource{
		CloudResource: CloudResource{
			Key:          key,
			Name:         name,
			Provider:     "gcp",
			Properties:   properties,
			ResourceType: CloudResourceType(rtype),
			AccountRef:   accountRef,
			Labels:       []string{"GCPResource"},
		},
	}

	r.DisplayName = r.GetDisplayName()
	r.Region = r.GetRegion()
	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *GCPResource) GetDisplayName() string {
	parts := strings.Split(a.Name, "/")
	if len(parts) == 6 {
		return parts[len(parts)-1]
	}
	return a.Name
}
func (a *GCPResource) GetHooks() []registry.Hook {
	return a.CloudResource.GetHooks()
}

func (a *GCPResource) GetIPs() []string {
	return make([]string, 0) // Return empty slice instead of nil
}

func (a *GCPResource) GetRegion() string {
	if parts := strings.Split(a.Name, "/"); len(parts) >= 4 {
		for i, part := range parts {
			if part == "zones" || part == "regions" {
				if i+1 < len(parts) {
					if strings.Contains(parts[i+1], "-") {
						region := strings.Join(strings.Split(parts[i+1], "-")[:2], "-")
						return region
					}
					return parts[i+1]
				}
			}
		}
	}
	return ""
}

func (a *GCPResource) GetURL() string {
	return ""
}

func (a *GCPResource) Group() string { return "gcpresource" }

// Insertable interface methods
func (a *GCPResource) Merge(otherModel any) {
	other, ok := otherModel.(*GCPResource)
	if !ok {
		return
	}
	a.Status = other.Status
	a.Visited = other.Visited

	// Safely copy properties with nil checks
	if a.Properties == nil {
		a.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		maps.Copy(a.Properties, other.Properties)
	}
}

func (a *GCPResource) Visit(otherModel any) error {
	other, ok := otherModel.(*GCPResource)
	if !ok {
		return fmt.Errorf("expected *GCPResource, got %T", otherModel)
	}
	a.Visited = other.Visited
	a.Status = other.Status

	// Safely copy properties with nil checks
	if a.Properties == nil {
		a.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		maps.Copy(a.Properties, other.Properties)
	}

	// Fix TTL update logic: update if other has a valid TTL
	if other.TTL != 0 {
		a.TTL = other.TTL
	}
	return nil
}

func (a *GCPResource) WithStatus(status string) Target {
	ret := *a
	ret.Status = status
	return &ret
}

// IsPrivate determines if this GCP resource is private based on its IP/URL
func (a *GCPResource) IsPrivate() bool {
	// Check if resource has any public IP addresses
	if ips := a.GetIPs(); len(ips) > 0 {
		for _, ip := range ips {
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				if parsedIP != nil && !parsedIP.IsPrivate() {
					return false // Has at least one public IP = not private
				}
			}
		}
	}

	// Check if resource has a public URL/endpoint
	if url := a.GetURL(); url != "" {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}
