package model

import (
	"encoding/gob"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const CloudResourceLabel = "CloudResource"

var specialCharRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_]`) // to conform with label validator

func init() {
	MustRegisterLabel(CloudResourceLabel)
	registry.Registry.MustRegisterModel(&CloudResource{})

	// register the type for properties
	gob.Register(map[string]any{})
	gob.Register(map[string]string{})
	gob.Register(map[string][]string{})
}

type AssetBuilder interface {
	NewAssets() []Asset
	GraphModel
}

type CloudResource struct {
	IPs          []string          `neo4j:"ips" json:"ips"`
	URLs         []string          `neo4j:"urls" json:"urls"`
	Name         string            `neo4j:"name" json:"name"`
	DisplayName  string            `neo4j:"displayName" json:"displayName"`
	Provider     string            `neo4j:"provider" json:"provider"`
	ResourceType CloudResourceType `neo4j:"resourceType" json:"resourceType"`
	Region       string            `neo4j:"region" json:"region"`
	AccountRef   string            `neo4j:"accountRef" json:"accountRef"`
	Properties   map[string]any    `neo4j:"properties" json:"properties"`
	Labels       []string          `neo4j:"labels" json:"labels"`
	BaseAsset
	OriginationData
}

// Defaulted sets sensible default values for CloudResource
func (c *CloudResource) Defaulted() {
	c.Status = Active
	c.Created = Now()
	c.Visited = Now()
	c.TTL = Future(30 * 24) // 30 days
	if c.Properties == nil {
		c.Properties = make(map[string]any)
	}
}
func (c *CloudResource) GetDescription() string {
	return fmt.Sprintf("%s (%s)", c.Name, c.Provider)
}

func (a *CloudResource) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				labels := append(a.Labels, resourceLabels[a.ResourceType]...)
				labels = append(labels, a.ResourceType.String())
				labels = append(labels, AssetLabel, CloudResourceLabel, TTLLabel)
				slices.Sort(labels)
				a.Labels = slices.Compact(labels)

				for i, label := range a.Labels {
					label = strings.ReplaceAll(label, "::", "_")
					a.Labels[i] = strings.ReplaceAll(label, "/", "_")
				}

				a.Class = string(a.ResourceType)

				return nil
			},
			Description: "Set labels for the resource",
		},
	}
}

func (c *CloudResource) GetKey() string {
	return c.Key
}

func (c *CloudResource) GetLabels() []string {
	labels := make([]string, len(c.Labels))
	for i, label := range c.Labels {
		labels[i] = specialCharRegex.ReplaceAllString(label, "_")
	}
	return labels
}

func (c *CloudResource) GetStatus() string {
	return c.Status
}

func (c *CloudResource) IsStatus(value string) bool {
	return strings.HasPrefix(c.Status, value)
}

func (c *CloudResource) Valid() bool {
	return c.Key != ""
}

// GetSecret returns the secret reference for this cloud resource
func (c *CloudResource) GetSecret() string {
	if c.Secret != nil {
		return *c.Secret
	}
	return ""
}

func (c *CloudResource) Merge(other *CloudResource) {
	c.Status = other.Status
	c.Visited = other.Visited
	c.TTL = other.TTL

	if c.Properties == nil {
		c.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		for k, v := range other.Properties {
			c.Properties[k] = v
		}
	}

	c.OriginationData.Merge(other.OriginationData)
}

func (c *CloudResource) Visit(other *CloudResource) {
	c.Visited = other.Visited
	c.Status = other.Status

	if other.TTL != 0 {
		c.TTL = other.TTL
	}

	if c.Properties == nil {
		c.Properties = make(map[string]any)
	}
	if other.Properties != nil {
		maps.Copy(c.Properties, other.Properties)
	}
	c.OriginationData.Visit(other.OriginationData)
}
