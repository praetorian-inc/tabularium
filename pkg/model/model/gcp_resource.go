package model

import (
	"fmt"
	"maps"
	"net"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	GCPResourceLabel = "GCPResource"
)

type GCPResource struct {
	CloudResource
}

func init() {
	MustRegisterLabel(GCPResourceLabel)
	registry.Registry.MustRegisterModel(&GCPResource{})
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
	hooks := []registry.Hook{
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#gcpresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{GCPResourceLabel}
				return nil
			},
		},
	}

	hooks = append(hooks, a.CloudResource.GetHooks()...)
	return hooks
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

func (a *GCPResource) GetIPs() []string {
	ipList := []string{}
	if ip, ok := a.Properties["publicIP"].(string); ok && ip != "" {
		ipList = append(ipList, ip)
	}
	if ip, ok := a.Properties["publicIPv6"].(string); ok && ip != "" {
		ipList = append(ipList, ip)
	}
	if ips, ok := a.Properties["publicIPs"].([]string); ok {
		ipList = append(ipList, ips...)
	}
	return ipList
}

func (a *GCPResource) GetURLs() []string {
	urlList := []string{}
	if url, ok := a.Properties["publicURL"].(string); ok && url != "" {
		urlList = append(urlList, url)
	}
	if urls, ok := a.Properties["publicURLs"].([]string); ok {
		urlList = append(urlList, urls...)
	}
	return urlList
}

func (a *GCPResource) GetDNS() []string {
	domainList := []string{}
	if domain, ok := a.Properties["publicDomain"].(string); ok && domain != "" {
		domainList = append(domainList, domain)
	}
	if domains, ok := a.Properties["publicDomains"].([]string); ok {
		domainList = append(domainList, domains...)
	}
	return domainList
}

func (a *GCPResource) NewAssets() []Asset {
	assets := []Asset{}
	ipSet := make(map[string]bool)
	urlSet := make(map[string]bool)
	domainSet := make(map[string]bool)
	for _, ip := range a.GetIPs() {
		if _, ok := ipSet[ip]; !ok && ip != "" {
			ipSet[ip] = true
			toAdd := NewAsset(ip, ip)
			toAdd.CloudId = a.Name
			toAdd.CloudService = a.ResourceType.String()
			toAdd.CloudAccount = a.AccountRef
			assets = append(assets, toAdd)
		}
	}
	for _, url := range a.GetURLs() {
		if _, ok := urlSet[url]; !ok && url != "" {
			urlSet[url] = true
			toAdd := NewAsset(url, url)
			toAdd.CloudId = a.Name
			toAdd.CloudService = a.ResourceType.String()
			toAdd.CloudAccount = a.AccountRef
			assets = append(assets, toAdd)
		}
	}
	for _, domain := range a.GetDNS() {
		if _, ok := domainSet[domain]; !ok && domain != "" {
			domainSet[domain] = true
			toAdd := NewAsset(domain, domain)
			toAdd.CloudId = a.Name
			toAdd.CloudService = a.ResourceType.String()
			toAdd.CloudAccount = a.AccountRef
			assets = append(assets, toAdd)
		}
	}
	return assets
}

// TODO: worth revisiting logic; currently properties only include public
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
	if url := a.GetURLs(); len(url) > 0 {
		return false // Has public URL = not private
	}

	// No public IPs or URL = assume private
	return true
}
