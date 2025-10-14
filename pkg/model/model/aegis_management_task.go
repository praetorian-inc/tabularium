package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&AegisManagementTask{}, "aegis-mgmt")
}

// AegisManagementTask represents a task for Aegis infrastructure management operations
type AegisManagementTask struct {
	registry.BaseModel
	baseTableModel
	Username                  string            `dynamodbav:"username" json:"username" desc:"Username who initiated the Aegis management task." example:"user@example.com"`
	Key                       string            `dynamodbav:"key" json:"key" desc:"Unique key for the Aegis management task." example:"#aegis-mgmt#user@example.com#tunnel_management#2024-01-15T10:30:00Z"`
	AegisManagementCapability string            `dynamodbav:"aegisManagementCapability" json:"aegisManagementCapability" desc:"Name of the Aegis management capability." example:"tunnel_management"`
	AegisClientID             string            `dynamodbav:"aegisClientId" json:"aegisClientId" desc:"Aegis client ID target for the management operation." example:"C.12345abcdef"`
	AegisAgentID              string            `dynamodbav:"aegisAgentId,omitempty" json:"aegisAgentId,omitempty" desc:"Specific Aegis agent ID if targeting single agent." example:"agent-001"`
	FlowID                    string            `dynamodbav:"flowId,omitempty" json:"flowId,omitempty" desc:"Velociraptor flow ID for tracking execution." example:"F.D36EN790K9V6G"`
	Parameters                map[string]string `dynamodbav:"parameters" json:"parameters" desc:"Parameters for the Aegis management capability." example:"{\"tunnel_name\": \"my-tunnel\", \"action\": \"create\"}"`
	Status                    string            `dynamodbav:"status" json:"status" desc:"Current status of the Aegis management task." example:"AMT_PENDING"`
	Created                   string            `dynamodbav:"created" json:"created" desc:"Timestamp when the task was created (RFC3339)." example:"2024-01-15T10:30:00Z"`
	Updated                   string            `dynamodbav:"updated" json:"updated" desc:"Timestamp when the task was last updated (RFC3339)." example:"2024-01-15T10:35:00Z"`
	Started                   string            `dynamodbav:"started,omitempty" json:"started,omitempty" desc:"Timestamp when the task execution started (RFC3339)." example:"2024-01-15T10:31:00Z"`
	Completed                 string            `dynamodbav:"completed,omitempty" json:"completed,omitempty" desc:"Timestamp when the task completed (RFC3339)." example:"2024-01-15T10:35:00Z"`
	Result                    string            `dynamodbav:"result,omitempty" json:"result,omitempty" desc:"Result data from the Aegis management operation." example:"{\"tunnel_id\": \"tunnel-123\", \"status\": \"active\"}"`
	ErrorMessage              string            `dynamodbav:"errorMessage,omitempty" json:"errorMessage,omitempty" desc:"Error message if the task failed." example:"Failed to create tunnel: connection timeout"`
	CommandResult             *CommandResult    `dynamodbav:"commandResult,omitempty" json:"commandResult,omitempty" desc:"Detailed command execution result including exit code and output."`
	Async                     bool              `dynamodbav:"async" json:"async" desc:"Whether the Aegis management task is asynchronous." example:"true"`
	HealthCheck               bool              `dynamodbav:"healthCheck" json:"healthCheck" desc:"Whether to trigger a healthcheck after task completion." example:"true"`
	TTL                       int64             `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the task record (Unix timestamp)." example:"1706353200"`
}

// Aegis Management Task status constants
const (
	AegisManagementPending   = "AMT_PENDING"
	AegisManagementRunning   = "AMT_RUNNING"
	AegisManagementCompleted = "AMT_COMPLETED"
	AegisManagementFailed    = "AMT_FAILED"
)

// GetKey returns the unique key for the Aegis management task
func (amt *AegisManagementTask) GetKey() string {
	return amt.Key
}

// Raw returns the JSON representation of the Aegis management task
func (amt *AegisManagementTask) Raw() string {
	rawJSON, _ := json.Marshal(amt)
	return string(rawJSON)
}

// Is checks if the task has a specific status
func (amt *AegisManagementTask) Is(status string) bool {
	return amt.Status == status
}

// GetStatus returns the current status of the task
func (amt *AegisManagementTask) GetStatus() string {
	return amt.Status
}

// Update sets the status and updates the timestamp
func (amt *AegisManagementTask) Update(status string) {
	amt.Status = status
	amt.Updated = Now()
	amt.TTL = Future(7 * 24) // 7 days TTL
}

// SetRunning marks the task as running and sets the started timestamp
func (amt *AegisManagementTask) SetRunning() {
	amt.Status = AegisManagementRunning
	amt.Started = Now()
	amt.Updated = Now()
}

// SetCompleted marks the task as completed with result data
func (amt *AegisManagementTask) SetCompleted(result string) {
	amt.Status = AegisManagementCompleted
	amt.Result = result
	amt.Completed = Now()
	amt.Updated = Now()
}

// SetFailed marks the task as failed with error message
func (amt *AegisManagementTask) SetFailed(errorMessage string) {
	amt.Status = AegisManagementFailed
	amt.ErrorMessage = errorMessage
	amt.Completed = Now()
	amt.Updated = Now()
}

// Valid checks if the task has a valid key format
func (amt *AegisManagementTask) Valid() bool {
	return strings.HasPrefix(amt.Key, "#aegis-mgmt#")
}

// Defaulted sets default values for a new task
func (amt *AegisManagementTask) Defaulted() {
	amt.Status = AegisManagementPending
	if amt.Parameters == nil {
		amt.Parameters = make(map[string]string)
	}
	amt.Created = Now()
	amt.Updated = Now()
	amt.TTL = Future(7 * 24) // 7 days TTL
}

// GetHooks returns hooks for the Aegis management task
func (amt *AegisManagementTask) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if amt.Key == "" && amt.Username != "" && amt.AegisManagementCapability != "" {
					taskID := uuid.New().String()
					amt.Key = fmt.Sprintf("#aegis-mgmt#%s#%s#%s", amt.Username, amt.AegisManagementCapability, taskID)
				}
				return nil
			},
		},
	}
}

// GetDescription returns a description for the AegisManagementTask model
func (amt *AegisManagementTask) GetDescription() string {
	return "Represents an Aegis infrastructure management task, including its status, parameters, and results."
}

// NewAegisManagementTask creates a new Aegis management task
func NewAegisManagementTask(username, capability, clientID string, parameters map[string]string, async bool, healthCheck bool) AegisManagementTask {
	task := AegisManagementTask{
		Username:                  username,
		AegisManagementCapability: capability,
		AegisClientID:             clientID,
		Parameters:                parameters,
		Async:                     async,
		HealthCheck:               healthCheck,
	}
	task.Defaulted()
	registry.CallHooks(&task)
	return task
}
