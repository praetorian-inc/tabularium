package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"strings"
)

type Condition struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the condition." example:"system"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the condition." example:"#condition#exposure-open_port-80"`
	// Attributes
	Name    string `dynamodbav:"name" json:"name" desc:"Name of the condition or attribute." example:"https"`
	Value   string `dynamodbav:"value" json:"value" desc:"Value associated with the condition." example:"443"`
	Count   int    `dynamodbav:"-" json:"-"`
	Source  string `dynamodbav:"source" json:"source" desc:"Source of the condition information." example:"system"`
	Updated string `dynamodbav:"updated" json:"updated" desc:"Timestamp when the condition was last updated (RFC3339)." example:"2023-10-27T10:00:00Z"`
}

func init() {
	shared.Registry.MustRegisterModel(&Condition{})
}

func AttributeConditions(attribute Attribute) []Condition {
	return []Condition{
		NewCondition(attribute.Name, ""),
		NewCondition(attribute.Name, attribute.Value),
	}
}

func (c *Condition) GetExposure() string {
	return alertName(c.Name, c.Value)
}

func NewCondition(name, value string) Condition {
	c := Condition{
		Name:  name,
		Value: value,
	}
	c.Defaulted()
	model.CallHooks(&c)
	return c
}

func alertName(name, value string) string {
	sanitizedName := strings.ReplaceAll(name, "-", "_")
	sanitizedValue := strings.ReplaceAll(value, "-", "_")

	var exposure string
	if value == "" {
		exposure = fmt.Sprintf("exposure-%s", sanitizedName)
	} else {
		exposure = fmt.Sprintf("exposure-%s-%s", sanitizedName, sanitizedValue)
	}

	return exposure
}

// GetDescription returns a description for the Condition model.
func (c *Condition) GetDescription() string {
	return "Represents a specific condition or finding observed on an asset, often related to vulnerabilities or misconfigurations."
}

func (c *Condition) Defaulted() {
	c.Source = "system"
	c.Updated = Now()
}

func (c *Condition) GetHooks() []model.Hook {
	return []model.Hook{
		{
			Call: func() error {
				c.Key = fmt.Sprintf("#condition#%s", alertName(c.Name, c.Value))
				return nil
			},
		},
	}
}
