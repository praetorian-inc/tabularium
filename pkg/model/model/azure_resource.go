package model

import (
	"fmt"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	AzureResourceLabel = "AzureResource"
)

type AzureResource struct {
	CloudResource
	ResourceGroup string `neo4j:"resourceGroup" json:"resourceGroup"`
}

func init() {
	MustRegisterLabel(AzureResourceLabel)
	registry.Registry.MustRegisterModel(&AzureResource{})
}

func NewAzureResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (AzureResource, error) {
	r := AzureResource{
		CloudResource: CloudResource{
			Name:         name,
			Provider:     "azure",
			Properties:   properties,
			ResourceType: CloudResourceType(rtype),
			AccountRef:   accountRef,
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
	if displayName, ok := a.Properties["displayName"].(string); ok {
		return displayName
	}
	parts := strings.Split(a.Name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return a.Name
}

func (a *AzureResource) Defaulted() {
	a.Origins = []string{"azure"}
	a.AttackSurface = []string{"cloud"}
	a.CloudResource.Defaulted()
	a.BaseAsset.Defaulted()
}

func (a *AzureResource) GetHooks() []registry.Hook {
	hooks := []registry.Hook{
		useGroupAndIdentifier(a, &a.AccountRef, &a.Name),
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#azureresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{AzureResourceLabel}
				a.CloudResource.IPs = a.GetIPs()
				a.CloudResource.URLs = a.GetURLs()
				return nil
			},
		},
		setGroupAndIdentifier(a, &a.AccountRef, &a.Name),
	}

	hooks = append(hooks, a.CloudResource.GetHooks()...)
	return hooks
}

func (a *AzureResource) NewAssets() []Asset {
	assets := []Asset{}
	ipSet := make(map[string]bool)
	urlSet := make(map[string]bool)

	record := func(asset Asset) {
		asset.CloudId = a.Name
		asset.CloudService = a.ResourceType.String()
		asset.CloudAccount = a.AccountRef
		assets = append(assets, asset)
	}

	for _, ip := range a.GetIPs() {
		if _, ok := ipSet[ip]; !ok && ip != "" {
			ipSet[ip] = true
			record(NewAsset(ip, ip))
		}
	}
	for _, url := range a.GetURLs() {
		if _, ok := urlSet[url]; !ok && url != "" {
			urlSet[url] = true
			record(NewAsset(url, url))
		}
	}

	return assets
}

func (a *AzureResource) GetIPs() []string {
	ips := make([]string, 0)

	// Extract private IPs
	if privateIPs, ok := a.Properties["privateIPs"].([]any); ok {
		for _, ip := range privateIPs {
			if ipStr, ok := ip.(string); ok && ipStr != "" {
				ips = append(ips, ipStr)
			}
		}
	}

	// Extract public IPs
	if publicIPs, ok := a.Properties["publicIPs"].([]any); ok {
		for _, ip := range publicIPs {
			if ipStr, ok := ip.(string); ok && ipStr != "" {
				ips = append(ips, ipStr)
			}
		}
	}

	return ips
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

func (a *AzureResource) GetURLs() []string {
	return []string{}
}

func (a *AzureResource) Visit(other Assetlike) {
	otherResource, ok := other.(*AzureResource)
	if !ok {
		return
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *AzureResource) Merge(other Assetlike) {
	otherResource, ok := other.(*AzureResource)
	if !ok {
		return
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
}

func (a *AzureResource) Group() string {
	return a.AccountRef
}

func (a *AzureResource) Identifier() string {
	return a.Name
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
	if urls := a.GetURLs(); len(urls) > 0 {
		for _, url := range urls {
			if url != "" {
				return false // Has public URL = not private
			}
		}
	}

	// No public IPs or URL = assume private
	return true
}
