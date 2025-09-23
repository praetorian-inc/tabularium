package model

import (
	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	RoleUser         = "user"
	RoleChariot      = "chariot"
	RoleSystem       = "system"
	RoleToolCall     = "tool call"
	RoleToolResponse = "tool response"
)

type Message struct {
	registry.BaseModel
	baseTableModel
	Username       string `dynamodbav:"username" json:"username" desc:"Username who sent the message." example:"user@example.com"`
	Key            string `dynamodbav:"key" json:"key" desc:"Unique key for the message." example:"#message#550e8400-e29b-41d4-a716-446655440000#1sB5tZfLipTVWQWHVKnDFS6kFRK"`
	ConversationID string `dynamodbav:"conversationId" json:"conversationId" desc:"ID of the conversation this message belongs to." example:"550e8400-e29b-41d4-a716-446655440000"`
	Role           string `dynamodbav:"role" json:"role" desc:"Role of the message sender (user, chariot, system, tool)." example:"user"`
	Content        string `dynamodbav:"content" json:"content" desc:"Content of the message." example:"Hello, how can I help you today?"`
	Timestamp      string `dynamodbav:"timestamp" json:"timestamp" desc:"Timestamp when the message was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	MessageID      string `dynamodbav:"messageId" json:"messageId" desc:"KSUID for message ordering." example:"1sB5tZfLipTVWQWHVKnDFS6kFRK"`
	TTL            int64  `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the message record (Unix timestamp)." example:"1706353200"`
	ToolUseID      string `dynamodbav:"toolUseId,omitempty" json:"toolUseId,omitempty" desc:"Tool use ID for tool result messages." example:"tooluse_kZJMlvQmRJ6eAyJE5GIl7Q"`
	ToolUseContent string `dynamodbav:"toolUseContent,omitempty" json:"toolUseContent,omitempty" desc:"JSON serialized tool use content for assistant tool use messages." example:"{\"name\":\"query\",\"input\":{\"node\":{\"labels\":[\"Asset\"]}}}"`
}

func init() {
	registry.Registry.MustRegisterModel(&Message{})
}

func (m *Message) GetKey() string {
	return m.Key
}

func (m *Message) GetDescription() string {
	return "Represents a message within a conversation, with KSUID ordering for proper sequencing."
}

func (m *Message) Defaulted() {
	m.Timestamp = Now()
	m.TTL = Future(24 * 30)
	if m.MessageID == "" {
		id, _ := uuid.NewV7()
		m.MessageID = id.String()
	}
}

func (m *Message) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if m.Key == "" {
					if m.MessageID == "" {
						id, _ := uuid.NewV7()
						m.MessageID = id.String()
					}
					m.Key = "#message#" + m.ConversationID + "#" + m.MessageID
				}
				return nil
			},
		},
	}
}

func (m *Message) Valid() bool {
	return m.ConversationID != "" && m.Role != "" && m.Content != ""
}

func NewMessage(conversationID, role, content, username string) Message {
	msg := Message{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Username:       username,
	}
	msg.Defaulted()
	registry.CallHooks(&msg)
	return msg
}

