package model

import (
	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Conversation struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Username who owns the conversation." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the conversation." example:"#conversation#example-conversation#550e8400-e29b-41d4-a716-446655440000"`
	// Attributes
	Name        string `dynamodbav:"name" json:"name" desc:"Name of the conversation." example:"My AI Assistant Chat"`
	Description string `dynamodbav:"description" json:"description,omitempty" desc:"Optional description of the conversation." example:"Discussion about implementing new features"`
	Status      string `dynamodbav:"status" json:"status" desc:"Current status of the conversation." example:"active"`
	Created     string `dynamodbav:"created" json:"created" desc:"Timestamp when the conversation was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	TTL         int64  `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the conversation record (Unix timestamp)." example:"1706353200"`
}

func init() {
	registry.Registry.MustRegisterModel(&Conversation{})
}

func (c *Conversation) GetKey() string {
	return c.Key
}

func (c *Conversation) GetDescription() string {
	return "Represents a conversation between a user and AI assistant with running capabilities."
}

func (c *Conversation) Defaulted() {
	c.Status = "active"
	c.Created = Now()
	c.TTL = Future(24 * 30) // 30 days
}

func (c *Conversation) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if c.Key == "" {
					conversationID := uuid.New().String()
					c.Key = "#conversation#" + c.Name + "#" + conversationID
				}
				return nil
			},
		},
	}
}

func (c *Conversation) Valid() bool {
	return c.Name != "" && c.Username != ""
}

func NewConversation(name, username string) Conversation {
	conv := Conversation{
		Name:     name,
		Username: username,
	}
	conv.Defaulted()
	registry.CallHooks(&conv)
	return conv
}