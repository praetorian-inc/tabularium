package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	K8sProvider      = "kubernetes"
	K8sResourceLabel = "K8sResource"
)

// K8sResource represents a Kubernetes workload or configuration object
// discovered during cloud security scanning. The AccountRef is the cloud
// account that owns the cluster, and Name is the Orca-assigned unique ID
// (typically a UUID).
type K8sResource struct {
	CloudResource

	// Cluster is the name of the K8s cluster this resource belongs to.
	Cluster   string `neo4j:"cluster" json:"cluster"`
	Namespace string `neo4j:"namespace" json:"namespace"`
}

func init() {
	MustRegisterLabel(K8sResourceLabel)
	registry.Registry.MustRegisterModel(&K8sResource{})
}

// NewK8sResource creates a K8sResource. Name is the unique identifier
// (typically a UUID from Orca), accountRef is the cloud account owning
// the cluster.
func NewK8sResource(name, accountRef string, rtype CloudResourceType, properties map[string]any) (K8sResource, error) {
	r := K8sResource{
		CloudResource: CloudResource{
			Name:         name,
			Provider:     K8sProvider,
			Properties:   properties,
			ResourceType: rtype,
			AccountRef:   accountRef,
		},
	}

	r.Defaulted()
	registry.CallHooks(&r)
	return r, nil
}

func (a *K8sResource) Defaulted() {
	a.Origins = []string{"kubernetes"}
	a.AttackSurface = []string{"cloud"}
	a.CloudResource.Defaulted()
	a.BaseAsset.Defaulted()
}

func (a *K8sResource) GetHooks() []registry.Hook {
	hooks := []registry.Hook{
		useGroupAndIdentifier(a, &a.AccountRef, &a.Name),
		{
			Call: func() error {
				a.CloudResource.Key = fmt.Sprintf("#k8sresource#%s#%s", a.AccountRef, a.Name)
				a.CloudResource.Labels = []string{K8sResourceLabel}
				return nil
			},
		},
		setGroupAndIdentifier(a, &a.AccountRef, &a.Name),
	}

	hooks = append(hooks, a.CloudResource.GetHooks()...)
	return hooks
}

func (a *K8sResource) WithStatus(status string) Target {
	ret := *a
	ret.Status = status
	return &ret
}

func (a *K8sResource) GetIPs() []string {
	return []string{}
}

func (a *K8sResource) GetURLs() []string {
	return []string{}
}

func (a *K8sResource) NewAssets() []Asset {
	asset := NewAsset(a.Name, a.Name)
	asset.CloudId = a.Name
	asset.CloudService = a.ResourceType.String()
	asset.CloudAccount = a.AccountRef
	return []Asset{asset}
}

func (a *K8sResource) Group() string {
	return a.AccountRef
}

func (a *K8sResource) Identifier() string {
	return a.Name
}

func (a *K8sResource) Visit(other Assetlike) {
	otherResource, ok := other.(*K8sResource)
	if !ok {
		return
	}

	if otherResource.Cluster != "" {
		a.Cluster = otherResource.Cluster
	}
	if otherResource.Namespace != "" {
		a.Namespace = otherResource.Namespace
	}

	a.CloudResource.Visit(&otherResource.CloudResource)
	a.BaseAsset.Visit(otherResource)
}

func (a *K8sResource) Merge(other Assetlike) {
	otherResource, ok := other.(*K8sResource)
	if !ok {
		return
	}

	if otherResource.Cluster != "" {
		a.Cluster = otherResource.Cluster
	}
	if otherResource.Namespace != "" {
		a.Namespace = otherResource.Namespace
	}

	a.CloudResource.Merge(&otherResource.CloudResource)
	a.BaseAsset.Merge(otherResource)
}

func (a *K8sResource) IsPrivate() bool {
	return true
}
