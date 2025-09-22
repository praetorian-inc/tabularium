package model

type PlannerEvent struct {
	Username       string   `json:"username"`
	ConversationID string   `json:"conversationId"`
	JobKey         string   `json:"jobKey"`
	Source         string   `json:"source"`
	Target         string   `json:"target"`
	Status         string   `json:"status"`
	ResultKeys     []string `json:"resultKeys"`
	TotalCount     int      `json:"totalCount"`
	Comment        string   `json:"comment,omitempty"`
	CompletedAt    string   `json:"completedAt"`
}