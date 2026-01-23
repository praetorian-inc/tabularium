package model

import "github.com/praetorian-inc/tabularium/pkg/registry"

// Trace represents a trace event for job lifecycle and collector operations.
// Used for observability and customer-facing trace queries.
type Trace struct {
	registry.BaseModel
	baseTableModel

	// Identity (DynamoDB key structure: username + #trace#{trace_id}#{span_id})
	Username string `dynamodbav:"username" json:"username" desc:"Username who owns this trace event." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the trace event." example:"#trace#abc-123#span-001"`

	// Trace context
	TraceID      string `dynamodbav:"trace_id" json:"trace_id" desc:"Root trace identifier shared by all events in a request chain." example:"550e8400-e29b-41d4-a716-446655440000"`
	SpanID       string `dynamodbav:"span_id" json:"span_id" desc:"Unique identifier for this specific event/span." example:"660e8400-e29b-41d4-a716-446655440001"`
	ParentSpanID string `dynamodbav:"parent_span_id,omitempty" json:"parent_span_id,omitempty" desc:"Parent span identifier for building trace trees." example:"550e8400-e29b-41d4-a716-446655440000"`

	// Event details
	EventType string `dynamodbav:"event_type" json:"event_type" desc:"Type of trace event (e.g., trace.job.created, trace.collector.http.request)." example:"trace.job.created"`
	Timestamp string `dynamodbav:"timestamp" json:"timestamp" desc:"When this event occurred (RFC3339)." example:"2023-10-27T10:00:00Z"`

	// Context (varies by event type)
	JobKey     string            `dynamodbav:"job_key,omitempty" json:"job_key,omitempty" desc:"Associated job key for job lifecycle events." example:"#job#example.com#asset#portscan"`
	Capability string            `dynamodbav:"capability,omitempty" json:"capability,omitempty" desc:"Capability that generated this event." example:"portscan"`
	TargetKey  string            `dynamodbav:"target_key,omitempty" json:"target_key,omitempty" desc:"Target asset key for the operation." example:"#asset#ipv4#192.168.1.1"`
	Status     string            `dynamodbav:"status,omitempty" json:"status,omitempty" desc:"Result status (e.g., success, error, timeout)." example:"success"`
	Error      string            `dynamodbav:"error,omitempty" json:"error,omitempty" desc:"Error message if the operation failed." example:"connection timeout"`
	Duration   int64             `dynamodbav:"duration,omitempty" json:"duration,omitempty" desc:"Duration of the operation in milliseconds." example:"1234"`
	Metadata   map[string]string `dynamodbav:"metadata,omitempty" json:"metadata,omitempty" desc:"Additional event-specific metadata." example:"{\"http_status\": \"200\"}"`

	// Standard fields
	Created string `dynamodbav:"created" json:"created" desc:"Timestamp when this trace record was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	TTL     int64  `dynamodbav:"ttl,omitempty" json:"ttl,omitempty" desc:"Time-to-live for this trace record (Unix timestamp)." example:"1706353200"`
}

func init() {
	registry.Registry.MustRegisterModel(&Trace{})
}

func (t *Trace) GetKey() string {
	return t.Key
}

// GetDescription returns a description for the Trace model.
func (t *Trace) GetDescription() string {
	return "Represents a trace event for job lifecycle and collector operations, enabling observability and customer-facing trace queries."
}

// GetHooks returns hooks for the Trace model.
func (t *Trace) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if t.Metadata == nil {
					t.Metadata = make(map[string]string)
				}
				return nil
			},
		},
	}
}

// NewTrace creates a new Trace event with the given trace context.
func NewTrace(traceID, spanID, parentSpanID, eventType string) Trace {
	now := Now()
	trace := Trace{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		EventType:    eventType,
		Timestamp:    now,
		Created:      now,
		TTL:          Future(30 * 24), // 30 day retention
		Metadata:     make(map[string]string),
	}
	return trace
}

// SetKey generates the DynamoDB key for this trace event.
func (t *Trace) SetKey() {
	t.Key = "#trace#" + t.TraceID + "#" + t.SpanID
}
