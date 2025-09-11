package model

import (
	"encoding/json"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&Cache{})
}

type Cache struct {
	registry.BaseModel
	baseTableModel
	Username string          `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string          `dynamodbav:"key" json:"key" desc:"Unique key for the cache record." example:"#cache#ipv4scan#127.0.0.1"`
	TTL      int64           `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the cache record (Unix timestamp)." example:"1706353200"`
	Keys     []string        `dynamodbav:"-" json:"keys"`
	Cached   json.RawMessage `dynamodbav:"cached" json:"cached"`
	Large    bool            `dynamodbav:"large" json:"large"`
}

func (a *Cache) GetDescription() string {
	return "Represents a cache line in the chariot cache"
}

func (a *Cache) Defaulted() {
	a.TTL = Future(12)
}

func (a *Cache) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				a.Key = "#cache"
				for _, key := range a.Keys {
					a.Key += "#" + key
				}
				return nil
			},
		},
	}
}

func NewCache(keys ...string) Cache {
	a := Cache{
		Keys: keys,
	}
	a.Defaulted()
	registry.CallHooks(&a)
	return a
}
