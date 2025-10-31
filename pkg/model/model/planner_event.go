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

	// Agent completion event (child agent finished, resume parent)
	AgentCompletion *PlannerAgentCompletion `json:"agentCompletion,omitempty"`
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

type PlannerAgentCompletion struct {
	ParentConversationID string   `json:"parentConversationId" desc:"Parent conversation to resume."`
	ChildConversationID  string   `json:"childConversationId" desc:"Child conversation that completed."`
	ParentToolUseID      string   `json:"parentToolUseId" desc:"Tool use ID in parent conversation to respond to."`
	Response             string   `json:"response" desc:"Child agent's final response text."`
	ToolsUsed            []string `json:"toolsUsed" desc:"List of tools the child agent used."`
	Success              bool     `json:"success" desc:"Whether child completed successfully."`
	Error                string   `json:"error,omitempty" desc:"Error message if child failed."`
	CompletedAt          string   `json:"completedAt" desc:"RFC3339 timestamp when child completed."`
}
