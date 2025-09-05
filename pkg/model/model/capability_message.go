package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/segmentio/ksuid"
)

// CapabilityMessage represents a bidirectional communication message for capability jobs
type CapabilityMessage struct {
	registry.BaseModel
	baseTableModel
	Key       string `dynamodbav:"key" json:"key" desc:"Composite key for the message: #message#{jobKey}#{KSUID}" example:"#message#job123#2Q2L5TYy3IKcEU7Jd9sZmdxZBQN"`
	JobKey    string `dynamodbav:"jobKey" json:"jobKey" desc:"Key of the job this message belongs to" example:"#job#example.com#asset#portscan"`
	MessageID string `dynamodbav:"messageID" json:"messageID" desc:"Unique KSUID identifier for this message" example:"2Q2L5TYy3IKcEU7Jd9sZmdxZBQN"`
	Body      string `dynamodbav:"body" json:"body" desc:"Message body content" example:"status: scanning port 443"`
	Sender    string `dynamodbav:"sender" json:"sender" desc:"Origin of the message: 'user' or 'capability'" example:"user"`
	Timestamp string `dynamodbav:"timestamp" json:"timestamp" desc:"RFC3339 timestamp when the message was created" example:"2023-10-27T10:00:00Z"`
	TTL       int64  `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the message record (Unix timestamp), set to 24 hours from creation" example:"1706353200"`
}

func init() {
	registry.Registry.MustRegisterModel(&CapabilityMessage{})
}

func (cm *CapabilityMessage) GetKey() string {
	return cm.Key
}

func (cm *CapabilityMessage) GetDescription() string {
	return "Represents a bidirectional communication message between users and running capability jobs"
}

func (cm *CapabilityMessage) Defaulted() {
	cm.Timestamp = Now()
	cm.TTL = Future(24) // 1-day TTL (24 hours)
}

// TableModel implementation
func (cm *CapabilityMessage) TableModel() {}

func (cm *CapabilityMessage) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				// Generate KSUID if not set
				if cm.MessageID == "" {
					cm.MessageID = ksuid.New().String()
				}
				
				// Generate composite key: #message#{jobKey}#{KSUID}
				if cm.JobKey != "" {
					cm.Key = fmt.Sprintf("#message#%s#%s", cm.JobKey, cm.MessageID)
				}
				
				return nil
			},
		},
	}
}

// NewCapabilityMessage creates a new capability message
func NewCapabilityMessage(jobKey, body, sender string) CapabilityMessage {
	cm := CapabilityMessage{
		JobKey: jobKey,
		Body:   body,
		Sender: sender,
	}
	cm.Defaulted()
	registry.CallHooks(&cm)
	return cm
}

// Validate ensures the message is properly formatted
func (cm *CapabilityMessage) Validate() bool {
	return cm.JobKey != "" && cm.MessageID != "" && cm.Key != ""
}