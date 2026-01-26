package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Event type constants
const (
	EventJobStart   = "job_start"
	EventJobEnd     = "job_end"
	EventAgentStart = "agent_start"
	EventAgentEnd   = "agent_end"
	EventTaskStart  = "task_start"
	EventTaskEnd    = "task_end"
)

// JobEvent represents a single telemetry event in a trace.
type JobEvent struct {
	registry.BaseModel
	baseTableModel

	// Identity & Ownership
	Username string `dynamodbav:"username" json:"username" desc:"Account owner for tenant isolation."`
	Key      string `dynamodbav:"key" json:"key" desc:"Event key: #event#{trace_id}#{event_id}"`

	// Trace/Span Identity
	TraceID      string `dynamodbav:"trace_id" json:"trace_id" desc:"Root trace identifier (UUID)."`
	SpanID       string `dynamodbav:"span_id" json:"span_id" desc:"This span's identifier (UUID)."`
	ParentSpanID string `dynamodbav:"parent_span_id,omitempty" json:"parent_span_id,omitempty" desc:"Parent span ID."`
	EventID      string `dynamodbav:"event_id" json:"event_id" desc:"Unique event identifier (UUID)."`

	// Event Details
	EventType string `dynamodbav:"event_type" json:"event_type" desc:"Type of event (job_start, job_end, etc)."`

	// GSI field - MUST be prefixed with #trace# for filtering
	// Format: #trace#2026-01-26T10:22:13Z
	Created string `dynamodbav:"created" json:"created" desc:"Prefixed timestamp for Created GSI queries."`

	// Context
	JobKey string `dynamodbav:"job_key,omitempty" json:"job_key,omitempty" desc:"Associated job key."`
	Status string `dynamodbav:"status,omitempty" json:"status,omitempty" desc:"Outcome status for end events."`

	// Extensible attributes
	Attributes map[string]string `dynamodbav:"attributes,omitempty" json:"attributes,omitempty" desc:"Additional attributes."`

	// Lifecycle
	TTL int64 `dynamodbav:"ttl" json:"ttl" desc:"Unix timestamp for 90-day expiration."`
}

func init() {
	registry.Registry.MustRegisterModel(&JobEvent{})
}

func (e *JobEvent) GetKey() string {
	return e.Key
}

func (e *JobEvent) GetDescription() string {
	return "Represents a telemetry event in a job execution trace."
}

// NewJobEvent creates a new event with generated IDs and prefixed timestamp.
func NewJobEvent(username, traceID, spanID, parentSpanID, eventType string) JobEvent {
	eventID := uuid.New().String()
	timestamp := Now()

	event := JobEvent{
		Username:     username,
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		EventID:      eventID,
		EventType:    eventType,
		Created:      fmt.Sprintf("#trace#%s", timestamp), // Prefixed for GSI filtering
		Key:          fmt.Sprintf("#event#%s#%s", traceID, eventID),
		TTL:          Future(90 * 24), // 90 days
	}
	return event
}

// NewTraceID generates a new trace ID.
func NewTraceID() string {
	return uuid.New().String()
}

// NewSpanID generates a new span ID.
func NewSpanID() string {
	return uuid.New().String()
}
