package model

import (
	"fmt"
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
	r := GCPResource{
		CloudResource: CloudResource{
			Name:         name,
			Provider:     "gcp",
			Properties:   properties,
			ResourceType: rtype,
			AccountRef:   accountRef,
		},
	}

	r.DisplayName = r.GetDisplayName()
	r.Region = r.GetRegion()
	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *GCPResource) Defaulted() {
	a.Origins = []string{"gcp"}
	a.AttackSurface = []string{"cloud"}
	a.CloudResource.Defaulted()
	a.BaseAsset.Defaulted()
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
		useGroupAndIdentifier(a, &a.AccountRef, &a.Name),
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#gcpresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{GCPResourceLabel}
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

func (a *GCPResource) Group() string {
	return a.AccountRef
}

func (a *GCPResource) Identifier() string {
	return a.Name
}

func (a *GCPResource) Visit(other Assetlike) {
	otherResource, ok := other.(*GCPResource)
	if !ok {
		return
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *GCPResource) Merge(other Assetlike) {
	otherResource, ok := other.(*GCPResource)
	if !ok {
		return
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
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
	for _, domain := range a.GetDNS() {
		if _, ok := domainSet[domain]; !ok && domain != "" {
			record(NewAsset(domain, domain))
			domainSet[domain] = true
			for ip := range ipSet {
				record(NewAsset(domain, ip))
			}
		}
	}
	return assets
}
