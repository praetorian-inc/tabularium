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
}

type PlannerJobCompletion struct {
	JobKey      string                   `json:"jobKey"`
	Source      string                   `json:"source"`
	Target      string                   `json:"target"`
	Status      string                   `json:"status"`
	Results     []map[string]interface{} `json:"results"`
	TotalCount  int                      `json:"totalCount"`
	Comment     string                   `json:"comment,omitempty"`
	CompletedAt string                   `json:"completedAt"`
}

type PlannerUserMessage struct {
	Message string `json:"message"`
}
