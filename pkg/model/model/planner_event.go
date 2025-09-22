package model

import (
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type PlannerEvent struct {
	registry.BaseModel
	baseTableModel
	Username       string   `dynamodbav:"username" json:"username" desc:"Username who owns the job that completed." example:"user@example.com"`
	Key            string   `dynamodbav:"key" json:"key" desc:"Unique key for the planner event." example:"#plannerevent#550e8400-e29b-41d4-a716-446655440000"`
	ConversationID string   `dynamodbav:"conversationId" json:"conversationId" desc:"UUID of the conversation that initiated the job." example:"550e8400-e29b-41d4-a716-446655440000"`
	JobKey         string   `dynamodbav:"jobKey" json:"jobKey" desc:"Key of the completed job." example:"#job#example.com#10.0.1.5#nuclei#1698422400"`
	Source         string   `dynamodbav:"source" json:"source" desc:"Security capability that was executed." example:"nuclei"`
	Target         string   `dynamodbav:"target" json:"target" desc:"Target that was scanned." example:"#asset#example.com#10.0.1.5"`
	Status         string   `dynamodbav:"status" json:"status" desc:"Final job status (JP, JF)." example:"JP"`
	ResultKeys     []string `dynamodbav:"resultKeys" json:"resultKeys" desc:"First 20 result keys from job execution." example:"[\"#asset#example.com#web01\"]"`
	TotalCount     int      `dynamodbav:"totalCount" json:"totalCount" desc:"Total number of results found." example:"15"`
	Comment        string   `dynamodbav:"comment,omitempty" json:"comment,omitempty" desc:"Error message if job failed." example:"Connection timeout"`
	CompletedAt    string   `dynamodbav:"completedAt" json:"completedAt" desc:"Timestamp when job completed (RFC3339)." example:"2023-10-27T10:15:00Z"`
	TTL            int64    `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the event record (Unix timestamp)." example:"1706353200"`
}

func init() {
	registry.Registry.MustRegisterModel(&PlannerEvent{})
}

func (p *PlannerEvent) GetKey() string {
	return p.Key
}

func (p *PlannerEvent) GetDescription() string {
	return "Represents a job completion event for planner processing, containing aggregated results and metadata."
}

func (p *PlannerEvent) Defaulted() {
	p.CompletedAt = Now()
	p.TTL = Future(24 * 7)
}

func (p *PlannerEvent) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if p.Key == "" {
					p.Key = "#plannerevent#" + p.ConversationID + "#" + p.JobKey
				}
				return nil
			},
		},
	}
}

func (p *PlannerEvent) Valid() bool {
	return p.ConversationID != "" && p.JobKey != "" && p.Source != ""
}