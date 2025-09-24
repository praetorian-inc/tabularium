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
	Name     string `dynamodbav:"name" json:"name" desc:"Name of the conversation." example:"My AI Assistant Chat"`
	UUID     string `dynamodbav:"uuid" json:"uuid" desc:"UUID of the conversation for reference." example:"550e8400-e29b-41d4-a716-446655440000"`
	Source   string `dynamodbav:"source" json:"source" desc:"Who started the conversation." example:"user"`
	Created  string `dynamodbav:"created" json:"created" desc:"Timestamp when the conversation was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Topic    string `dynamodbav:"topic" json:"topic" desc:"Topic extracted from first message (first 256 chars)." example:"Find all active assets in my infrastructure"`
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
	c.Created = Now()
	c.Source = "user"
}

func (c *Conversation) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if c.Key == "" {
					conversationID := uuid.New().String()
					c.UUID = conversationID
					c.Key = "#conversation#" + conversationID
				}
				return nil
			},
		},
	}
}

func (c *Conversation) Valid() bool {
	return c.Name != ""
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