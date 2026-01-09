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
