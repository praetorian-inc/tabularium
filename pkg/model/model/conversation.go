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
	UUID     string `dynamodbav:"uuid" json:"uuid" desc:"UUID of the conversation for reference." example:"550e8400-e29b-41d4-a716-446655440000"`
	User               string `dynamodbav:"user" json:"user" desc:"Who started the conversation." example:"user@example.com"`
	Created            string `dynamodbav:"created" json:"created" desc:"Timestamp when the conversation was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Topic              string `dynamodbav:"topic" json:"topic" desc:"Topic extracted from first message (first 256 chars)." example:"Find all active assets in my infrastructure"`
	ParentConversation string `dynamodbav:"parent_conversation,omitempty" json:"parent_conversation,omitempty" desc:"UUID of parent conversation for subagent hierarchies." example:"550e8400-e29b-41d4-a716-446655440000"`
	// Blocking tool support
	Status          string           `dynamodbav:"status,omitempty" json:"status,omitempty" desc:"active or suspended"`
	PendingToolCall *PendingToolCall `dynamodbav:"pending_tool_call,omitempty" json:"pendingToolCall,omitempty" desc:"Pending tool call when suspended"`
}

// PendingToolCall tracks a suspended tool call awaiting async completion
type PendingToolCall struct {
	ToolCallID  string `dynamodbav:"tool_call_id" json:"toolCallId"`
	ToolName    string `dynamodbav:"tool_name" json:"toolName"`
	PendingID   string `dynamodbav:"pending_id" json:"pendingId"`
	EventType   string `dynamodbav:"event_type" json:"eventType"`
	SuspendedAt string `dynamodbav:"suspended_at" json:"suspendedAt"`
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
	return c.Key != ""
}

func NewConversation(topic string) Conversation {
	conv := Conversation{
		Topic: topic,
	}
	conv.Defaulted()
	registry.CallHooks(&conv)
	return conv
}
