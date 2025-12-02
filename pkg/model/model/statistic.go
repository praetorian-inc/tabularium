package model

import (
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
)

func init() {
	shared.Registry.MustRegisterModel(&Data{})
	shared.Registry.MustRegisterModel(&Statistic{})
}

const StatisticNow = "now"

// GetDescription returns a description for the Data model (used within Statistic).
func (d *Data) GetDescription() string {
	return "Represents the data points associated with a statistic, often counts or values."
}

type Data struct {
	model.BaseModel
	// all supported types of data. an individual statistic may use one or more.
	Count  int                `dynamodbav:"count,omitempty" json:"count,omitempty" desc:"A singular count value." example:"150"`
	Counts map[string]int     `dynamodbav:"counts,omitempty" json:"counts,omitempty" desc:"A map of named counts." example:"{\"active\": 100, \"pending\": 50}"`
	Values map[string]float64 `dynamodbav:"values,omitempty" json:"values,omitempty" desc:"A map of named float values." example:"{\"average_score\": 8.5}"`
}

// GetDescription returns a description for the Statistic model.
func (s *Statistic) GetDescription() string {
	return "Represents a statistic or metric collected by the system, including its associated data points."
}

type Statistic struct {
	model.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the statistic." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the statistic record." example:"#statistic#asset_count#global#2023-10-27T00:00:00Z#all"`
	Type     string `dynamodbav:"type" json:"type" desc:"The type or category of the statistic." example:"asset_count"`
	Name     string `dynamodbav:"name" json:"name" desc:"The specific name or scope of the statistic." example:"global"`
	Value    string `dynamodbav:"value" json:"value" desc:"A specific value or sub-category for the statistic." example:"all"`
	Data     Data   `dynamodbav:"data" json:"data" desc:"The actual statistical data." example:"{\"count\": 150}"`
	Created  string `dynamodbav:"created" json:"created" desc:"Timestamp when the statistic was generated (RFC3339 or 'now')." example:"2023-10-27T10:00:00Z"`
	Store    bool   `dynamodbav:"store" json:"store" desc:"Flag indicating if this statistic should be stored long-term." example:"true"`
	TTL      int64  `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the statistic record (Unix timestamp)." example:"1706353200"`
}

func (s *Statistic) Defaulted() {
	s.TTL = Future(90 * 24)
}

func (s *Statistic) GetHooks() []model.Hook {
	return []model.Hook{
		{
			Call: func() error {
				tipe := s.Type
				name := s.Name
				value := s.Value
				created := s.Created

				template := fmt.Sprintf("#statistic#%s#%s#%s#%%s", tipe, name, created)
				shortenedValue := value[:min(1024-len(template), len(value))]

				s.Key = fmt.Sprintf(template, shortenedValue)
				return nil
			},
		},
	}
}

func NewStatistic(tipe, name, value, created string) Statistic {
	s := Statistic{
		Type:    tipe,
		Name:    name,
		Value:   value,
		Created: created,
	}
	s.Defaulted()
	model.CallHooks(&s)
	return s
}

func (s *Statistic) Save() Statistic {
	stat := NewStatistic(s.Type, s.Name, s.Value, Now())
	stat.Data = s.Data
	return stat
}
