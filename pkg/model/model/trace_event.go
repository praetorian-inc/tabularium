package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const (
	EventJobStart   = "job_start"
	EventJobEnd     = "job_end"
	EventJobQueued  = "job_queued"
	EventAgentStart = "agent_start"
	EventAgentEnd   = "agent_end"
	EventTaskStart  = "task_start"
	EventTaskEnd    = "task_end"
)

type TraceEvent struct {
	registry.BaseModel
	baseTableModel
	Username     string            `dynamodbav:"username" json:"username" desc:"Account owner for tenant isolation."`
	Key          string            `dynamodbav:"key" json:"key" desc:"Event key: #event#{trace_id}#{event_id}"`
	TraceID      string            `dynamodbav:"trace_id" json:"trace_id" desc:"Root trace identifier (UUID)."`
	SpanID       string            `dynamodbav:"span_id" json:"span_id" desc:"This span's identifier (UUID)."`
	ParentSpanID string            `dynamodbav:"parent_span_id,omitempty" json:"parent_span_id,omitempty" desc:"Parent span ID."`
	EventID      string            `dynamodbav:"event_id" json:"event_id" desc:"Unique event identifier (UUID)."`
	EventType    string            `dynamodbav:"event_type" json:"event_type" desc:"Type of event (job_start, job_end, etc)."`
	Created      string            `dynamodbav:"created" json:"created" desc:"Prefixed timestamp for Created GSI queries."`
	Attributes   map[string]string `dynamodbav:"attributes,omitempty" json:"attributes,omitempty" desc:"Additional attributes."`
}

func init() {
	registry.Registry.MustRegisterModel(&TraceEvent{}, "event")
}

func (e *TraceEvent) GetKey() string {
	return e.Key
}

func (e *TraceEvent) GetDescription() string {
	return "Represents a telemetry event in a trace."
}

// NewTraceEvent creates a new event with generated IDs and prefixed timestamp.
func NewTraceEvent(username, traceID, spanID, parentSpanID, eventType string) TraceEvent {
	eventID := uuid.New().String()
	timestamp := Now()

	event := TraceEvent{
		Username:     username,
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		EventID:      eventID,
		EventType:    eventType,
		Created:      fmt.Sprintf("#trace#%s", timestamp), // Prefixed for GSI filtering
		Key:          fmt.Sprintf("#event#%s#%s", traceID, eventID),
	}
	return event
}

func NewTraceID() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func NewSpanID() string {
	id, _ := uuid.NewV7()
	return id.String()
}
