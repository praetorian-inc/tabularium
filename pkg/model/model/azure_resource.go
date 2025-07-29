package model

import (
	"fmt"
	"maps"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type AzureResource struct {
	CloudResource
	ResourceGroup string `neo4j:"resourceGroup" json:"resourceGroup"`
}

func NewAzureResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (AzureResource, error) {
	key := fmt.Sprintf("#azureresource#%s#%s", accountRef, name)

	r := AzureResource{
		CloudResource: CloudResource{
			Key:          key,
			Name:         name,
			Provider:     "azure",
			Properties:   properties,
			ResourceType: CloudResourceType(rtype),
			AccountRef:   accountRef,
			Labels:       []string{"AzureResource"},
		},
	}

	r.DisplayName = r.GetDisplayName()
	r.Region = r.GetRegion()
	r.ResourceGroup = r.GetResourceGroup()
	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *AzureResource) GetDisplayName() string {
	if displayName, ok := a.Properties["name"].(string); ok {
		return displayName
	}
	parts := strings.Split(a.Name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return a.Name
}

func (a *AzureResource) GetHooks() []registry.Hook {
	return a.CloudResource.GetHooks()
}

func (a *AzureResource) GetIPs() []string {
	return make([]string, 0) // Return empty slice instead of nil
}

// Azure-specific methods remain unchanged
func (a *AzureResource) GetRegion() string {
	if location, ok := a.Properties["location"].(string); ok {
		return location
	}
	return ""
}
func (a *AzureResource) GetResourceGroup() string {
	if resourceGroup, ok := a.Properties["resourceGroup"].(string); ok {
		return resourceGroup
	}
	return ""
}

func (a *AzureResource) GetURL() string {
	return ""
}

func (a *AzureResource) Group() string { return "azureresource" }

func (a *AzureResource) Merge(otherModel any) {
	other, ok := otherModel.(*AzureResource)
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

func (a *AzureResource) Visit(otherModel any) error {
	other, ok := otherModel.(*AzureResource)
	if !ok {
		return fmt.Errorf("expected *AzureResource, got %T", otherModel)
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
func (a *AzureResource) WithStatus(status string) Target {
	ret := *a
	ret.Status = status
	return &ret
}

// IsPrivate determines if this Azure resource is private based on its IP/URL
func (a *AzureResource) IsPrivate() bool {
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
