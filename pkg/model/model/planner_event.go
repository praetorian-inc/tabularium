package model

type PlannerEvent struct {
	Type           string                 `json:"type"`
	Username       string                 `json:"username"`
	ConversationID string                 `json:"conversationId"`
	
	// Job completion event
	JobCompletion  *PlannerJobCompletion  `json:"jobCompletion,omitempty"`
	
	// User message event
	UserMessage    *PlannerUserMessage    `json:"userMessage,omitempty"`
}

type PlannerJobCompletion struct {
	JobKey      string   `json:"jobKey"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Status      string   `json:"status"`
	ResultKeys  []string `json:"resultKeys"`
	TotalCount  int      `json:"totalCount"`
	Comment     string   `json:"comment,omitempty"`
	CompletedAt string   `json:"completedAt"`
}

type PlannerUserMessage struct {
	Message string `json:"message"`
}