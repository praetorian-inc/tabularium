package model

import (
	"encoding/gob"
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/model/label"
	"regexp"
	"slices"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

var CloudLabel = label.New("Cloud")

var neo4jNegateLabelRegex = regexp.MustCompile(`[^a-zA-Z0-9\-_]`)

type CloudResource struct {
	registry.BaseModel
	History
	Key          string            `neo4j:"key" json:"key"`
	Name         string            `neo4j:"name" json:"name"`
	DisplayName  string            `neo4j:"displayName" json:"displayName"`
	Provider     string            `neo4j:"provider" json:"provider"`
	ResourceType CloudResourceType `neo4j:"resourceType" json:"resourceType"`
	Region       string            `neo4j:"region" json:"region"`
	AccountRef   string            `neo4j:"accountRef" json:"accountRef"`
	Status       string            `neo4j:"status" json:"status"`
	Created      string            `neo4j:"created" json:"created"`
	Visited      string            `neo4j:"visited" json:"visited"`
	TTL          int64             `neo4j:"ttl" json:"ttl"`
	Properties   map[string]any    `neo4j:"properties" json:"properties"`
	Labels       []string          `neo4j:"labels" json:"labels"`
	Secret       *string           `neo4j:"secret" json:"secret"`
	Username     string            `neo4j:"username" json:"username"`
}

func (c *CloudResource) Defaulted() {
	c.Status = Active
	c.Created = Now()
	c.Visited = Now()
	c.TTL = Future(7 * 24)
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
				labels = append(labels, CloudLabel, TTLLabel)
				slices.Sort(labels)
				a.Labels = slices.Compact(labels)

				for i, label := range a.Labels {
					label = strings.ReplaceAll(label, "::", "_")
					a.Labels[i] = strings.ReplaceAll(label, "/", "_")
				}

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
		labels[i] = neo4jNegateLabelRegex.ReplaceAllString(label, "_")
	}
	return labels
}

func (c *CloudResource) GetStatus() string {
	return c.Status
}

func (c *CloudResource) Identifier() string {
	return c.Name
}

func (c *CloudResource) IsClass(value string) bool {
	return strings.HasPrefix(string(c.ResourceType), value)
}

func (c *CloudResource) IsStatus(value string) bool {
	return strings.HasPrefix(c.Status, value)
}

func (c *CloudResource) Valid() bool {
	return c.Key != ""
}

func (c *CloudResource) GetSecret() string {
	if c.Secret != nil {
		return *c.Secret
	}
	return ""
}

func init() {
	registry.Registry.MustRegisterModel(&CloudResource{})
	registry.Registry.MustRegisterModel(&AWSResource{})
	registry.Registry.MustRegisterModel(&AzureResource{})
	registry.Registry.MustRegisterModel(&GCPResource{})

	gob.Register(map[string]any{})
	gob.Register(map[string]string{})
	gob.Register(map[string][]string{})
}
