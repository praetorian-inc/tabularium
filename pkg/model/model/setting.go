package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Setting defines the structure for storing settings
type Setting struct {
	registry.BaseModel
	baseTableModel
	Username     string    `dynamodbav:"username" json:"username" desc:"Chariot username associated with the setting (if user-specific)." example:"user@example.com"`
	Key          string    `dynamodbav:"key" json:"key" desc:"Unique key for the setting." example:"#setting#notification_email"`
	LastModified time.Time `dynamodbav:"last_modified,omitempty" json:"last_modified,omitempty" desc:"Timestamp when the setting was last modified." example:"2023-01-01T00:00:00Z"`
	// Attributes
	Name  string `dynamodbav:"name" json:"name" desc:"Name of the setting." example:"notification_email"`
	Value any    `dynamodbav:"value" json:"value" desc:"Value of the setting." example:"admin@example.com"`
}

func init() {
	registry.Registry.MustRegisterModel(&Setting{})
	registry.Registry.MustRegisterModel(&Configuration{})
}

func (s *Setting) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				s.Key = fmt.Sprintf("#setting#%s", s.Name)
				return nil
			},
		},
	}
}

func NewSetting(name string, value any) Setting {
	s := Setting{
		Name:         name,
		Value:        value,
		LastModified: time.Now().UTC(),
	}
	s.Defaulted()
	registry.CallHooks(&s)
	return s
}

func (s *Setting) Valid() error {
	validator := settingValidators[s.Name]
	if validator == nil {
		return nil
	}
	return validator(s.Value)
}

func (s *Setting) GetKey() string {
	return s.Key
}

// GetDescription returns a description for the Setting model.
func (s *Setting) GetDescription() string {
	return "Represents a configuration setting within chariot."
}

// Configuration stores configuration settings managed by Praetorian.
type Configuration Setting

func (c *Configuration) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				c.Key = fmt.Sprintf("#configuration#%s", c.Name)
				return nil
			},
		},
	}
}

func NewConfiguration(name string, value any) Configuration {
	c := Configuration{
		Name:         name,
		Value:        value,
		LastModified: time.Now().UTC(),
	}
	registry.CallHooks(&c)
	return c
}

func (c *Configuration) GetKey() string {
	return c.Key
}

func (c *Configuration) Valid() error {
	validator := configurationValidators[c.Name]
	if validator == nil {
		return nil
	}
	return validator(c.Value)
}

func (c *Configuration) GetDescription() string {
	return "Represents a praetorian-only configuration setting within chariot."
}

type RateLimit struct {
	// Maximum number of hosts that can be checked simultaneously
	SimultaneousHosts *IntOrDefault `dynamodbav:"simultaneousHosts,omitempty" json:"simultaneousHosts,omitempty"`

	// Whether the maximum number of hosts is custom set
	SimultaneousHostsIsCustom bool `dynamodbav:"simultaneousHostsIsCustom,omitempty" json:"simultaneousHostsIsCustom,omitempty"`

	// Maximum rate of requests for single capability
	CapabilityRateLimit *IntOrDefault `dynamodbav:"capabilityRateLimit,omitempty" json:"capabilityRateLimit,omitempty"`

	// Whether the capability rate limit is custom set
	CapabilityRateLimitIsCustom bool `dynamodbav:"capabilityRateLimitIsCustom,omitempty" json:"capabilityRateLimitIsCustom,omitempty"`
}

func rateLimitValidator(value any) error {
	bites, err := json.Marshal(value)
	if err != nil || value == nil {
		return fmt.Errorf("failed to marshal JSON")
	}

	var r RateLimit
	if err = json.Unmarshal(bites, &r); err != nil {
		return fmt.Errorf("failed to unmarshal JSON")
	}

	check := func(desc string, val *IntOrDefault, min, max uint) error {
		if val == nil {
			return nil
		}
		if val := uint(*val); val < min || (max > 0 && val > max) {
			return fmt.Errorf("%s must be between %d and %d", desc, min, max)
		}
		return nil
	}
	// 500 as a general safe upper limit, but also to prevent DynamoDB pagination
	if err := check("simultaneous hosts", r.SimultaneousHosts, 10, 500); err != nil {
		return err
	}
	if err := check("capability rate limit", r.CapabilityRateLimit, 25, 0); err != nil {
		return err
	}

	return nil
}

type IntOrDefault uint

func (iod *IntOrDefault) GetOrDefault(defaultValue uint) uint {
	if iod == nil {
		return defaultValue
	}
	return uint(*iod)
}

func (iod *IntOrDefault) IsSet() bool {
	return iod != nil
}

func (iod *IntOrDefault) UnmarshalJSON(bites []byte) error {
	var value uint
	if err := json.Unmarshal(bites, &value); err != nil {
		return fmt.Errorf("failed to unmarshal JSON")
	}
	*iod = IntOrDefault(value)
	return nil
}

func (iod IntOrDefault) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint(iod))
}
