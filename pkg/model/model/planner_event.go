package model

type PlannerEvent struct {
	Type string `json:"type"`
	// Username of the tenant we're in
	Username string `json:"username"`
	// User that requested the job, possibly assumed in to the tenant
	User           string `json:"user"`
	ConversationID string `json:"conversationId"`

	// Job completion event
	JobCompletion *PlannerJobCompletion `json:"jobCompletion,omitempty"`

	// User message event
	UserMessage *PlannerUserMessage `json:"userMessage,omitempty"`

	// Parent conversation for subagent events
	ParentConversation string `json:"parentConversation,omitempty"`

	// Subagent execution event
	SubagentExecution *SubagentExecution `json:"subagentExecution,omitempty"`

	// Subagent completion event
	SubagentCompletion *SubagentCompletion `json:"subagentCompletion,omitempty"`

	// Unified agent execution (new)
	AgentExecution *AgentExecution `json:"agentExecution,omitempty"`
	Parent         *ParentContext  `json:"parent,omitempty"`

	// Unified completion (new)
	Completion *CompletionEvent `json:"completion,omitempty"`
}

type PlannerJobCompletion struct {
	JobKey      string                   `json:"jobKey"`
	Source      string                   `json:"source"`
	Target      string                   `json:"target"`
	Status      string                   `json:"status"`
	Results     []map[string]interface{} `json:"results"`
	ResultKeys  []string                 `json:"resultKeys"`
	TotalCount  int                      `json:"totalCount"`
	Comment     string                   `json:"comment,omitempty"`
	CompletedAt string                   `json:"completedAt"`
}

type PlannerUserMessage struct {
	Message string `json:"message"`
	Mode    string `json:"mode,omitempty"`
}

type SubagentExecution struct {
	Agent   string         `json:"agent"`   // Agent type to spawn
	Task    string         `json:"task"`    // Task description
	Context map[string]any `json:"context"` // Optional context data
}

type SubagentCompletion struct {
	Agent   string `json:"agent"`           // Agent type that executed
	Task    string `json:"task"`            // Original task
	Status  string `json:"status"`          // "success" or "error"
	Results string `json:"results"`         // Subagent output/summary
	Error   string `json:"error,omitempty"` // Error message if failed
}

// AgentExecution is the unified event for all agent invocations.
// Replaces separate planner_execution and subagent_execution events.
type AgentExecution struct {
	Agent        string         `json:"agent"`                   // Agent type
	Message      string         `json:"message"`                 // Task/prompt
	Context      map[string]any `json:"context,omitempty"`       // Optional context
	TargetEntity map[string]any `json:"targetEntity,omitempty"`  // For capability agents
	Blocking     bool           `json:"blocking"`                // Suspend conversation?
	ToolCallID   string         `json:"toolCallId,omitempty"`    // For blocking: tool to complete
}

// ParentContext tracks the parent conversation for subagent calls
type ParentContext struct {
	ConversationID string `json:"conversationId"`
	ToolCallID     string `json:"toolCallId"`
	Blocking       bool   `json:"blocking"`
}

// CompletionEvent is the unified completion for all async operations
type CompletionEvent struct {
	PendingID      string `json:"pendingId"`
	ConversationID string `json:"conversationId"`
	ToolCallID     string `json:"toolCallId,omitempty"`
	Blocking       bool   `json:"blocking"`
	Status         string `json:"status"` // "success" or "error"
	Result         string `json:"result"`
	Error          string `json:"error,omitempty"`
}
