package model

import (
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry"
	"maps"
	"slices"
	"strings"
)

type Job struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Username who initiated or owns the job." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the job." example:"#job#example.com#asset#portscan"`
	// Attributes
	DNS                   string            `dynamodbav:"dns" json:"dns" desc:"Primary DNS associated with the job's target." example:"example.com"`
	Source                string            `dynamodbav:"source" json:"source" desc:"The source or capability that generated this job, ordered by updated" example:"portscan#2023-10-27T10:05:00Z"`
	Comment               string            `dynamodbav:"comment" json:"comment,omitempty" desc:"Optional comment about the job." example:"Scanning standard web ports"`
	Created               string            `dynamodbav:"created" json:"created" desc:"Timestamp when the job was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Updated               string            `dynamodbav:"updated" json:"updated" desc:"Timestamp when the job was last updated (RFC3339)." example:"2023-10-27T10:05:00Z"`
	Started               string            `dynamodbav:"started" json:"started" desc:"Timestamp when the job was started (RFC3339)." example:"2023-10-27T10:05:00Z"`
	Finished              string            `dynamodbav:"finished" json:"finished" desc:"Timestamp when the job was finished (RFC3339)." example:"2023-10-27T10:05:00Z"`
	Delayed               string            `dynamodbav:"delayed,omitempty" json:"delayed,omitempty" desc:"Timestamp that this job should be delayed until" example:"2023-10-27T10:00:00Z"`
	Status                string            `dynamodbav:"status" json:"status" desc:"Current status of the job (e.g., JQ#2023-10-27T10:05:00Z)." example:"JQ#2023-10-27T10:05:00Z"`
	TTL                   int64             `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the job record (Unix timestamp)." example:"1706353200"`
	Name                  string            `dynamodbav:"name,omitempty" json:"name,omitempty" desc:"The IP address this job was executed from" example:"1.2.3.4"`
	Config                map[string]string `dynamodbav:"config" json:"config" external:"true" desc:"Configuration parameters for the job capability." example:"{\"test\": \"cve-1111-2222\"}"`
	Secret                map[string]string `dynamodbav:"-" json:"secret" external:"true" desc:"Sensitive configuration parameters (credentials, tokens, keys)."`
	CredentialIDs         []string          `dynamodbav:"credential_ids" json:"credential_ids,omitempty" external:"true" desc:"List of credential IDs to retrieve and inject into job execution." example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	LargeArtifactFileName string            `dynamodbav:"large_artifact_filename" json:"large_artifact_filename,omitempty" desc:"The name of the file that contains the large artifacts." example:"large_artifact.zip"`
	S3DownloadURL         string            `dynamodbav:"s3DownloadURL" json:"s3DownloadURL,omitempty" desc:"The URL of the file that contains the large output." example:"https://s3.amazonaws.com/big_output.zip"`
	Retries               int               `dynamodbav:"retries" json:"retries" desc:"The number of times this job has been retried."`
	AllowRepeat           bool              `dynamodbav:"allowRepeat" json:"allowRepeat" desc:"Indicates if repeating this job should be allowed. Used for manual jobs, or rescan jobs, that should not block other job executions." example:"false"`
	Full                  bool              `dynamodbav:"-" json:"full,omitempty" external:"true" desc:"Indicates if this is a full scan job." example:"false"`
	Capabilities          []string          `dynamodbav:"-" json:"capabilities,omitempty" external:"true" desc:"List of specific capabilities to run for this job." example:"[\"portscan\", \"nuclei\"]"`
	Queue                 string            `dynamodbav:"-" desc:"Target queue for the job." example:"standard"`
	Conversation          string            `dynamodbav:"conversation,omitempty" json:"conversation,omitempty" external:"true" desc:"UUID of the conversation that initiated this job." example:"550e8400-e29b-41d4-a716-446655440000"`
	User                  string            `dynamodbav:"user,omitempty" json:"user,omitempty" desc:"User who initiated this job." example:"user@example.com"`
	Partition             string            `dynamodbav:"partition,omitempty" json:"partition,omitempty" desc:"The partition of the job." example:"user@example.com##asset#test#0.0.0.0"`
	// Trace context for telemetry (propagated to child jobs)
	TraceID       string `dynamodbav:"trace_id,omitempty" json:"trace_id,omitempty" external:"true" desc:"Root trace ID for this job chain."`
	ParentSpanID  string `dynamodbav:"parent_span_id,omitempty" json:"parent_span_id,omitempty" desc:"Parent span ID from spawning job."`
	CurrentSpanID string `dynamodbav:"-" json:"current_span_id,omitempty" desc:"Current execution span ID (not persisted to DynamoDB, but serialized through SQS for trace correlation)."`
	Origin                TargetWrapper     `dynamodbav:"origin" json:"origin" desc:"The job that originally started this chain of jobs."`
	Target                TargetWrapper     `dynamodbav:"target" json:"target" desc:"The primary target of the job."`
	Parent                TargetWrapper     `dynamodbav:"parent" json:"parent,omitempty" desc:"Optional parent target from which this job was spawned."`
	RateLimit             RateLimit         `dynamodbav:"-" json:"-"`
	// internal
	Async  bool                  `dynamodbav:"-" json:"-"`
	stream chan []registry.Model `dynamodbav:"-" json:"-"`
	// records the original status of the job when Chariot received/created it.
	originalStatus string `dynamodbav:"-" json:"-"`
}

func init() {
	registry.Registry.MustRegisterModel(&Job{})
}

func (job *Job) GetKey() string {
	return job.Key
}

func (job *Job) Raw() string {
	rawJSON, _ := json.Marshal(job)
	return string(rawJSON)
}

func (job *Job) Is(status string) bool {
	status = status + "#"
	return strings.HasPrefix(job.Status, status)
}

func (job *Job) Was(status string) bool {
	status = status + "#"
	return strings.HasPrefix(job.originalStatus, status)
}

func (job *Job) GetStatus() string {
	parts := strings.Split(job.Status, "#")
	return parts[0]
}

func (job *Job) GetCapability() string {
	parts := strings.Split(job.Source, "#")
	return parts[0]
}

func (job *Job) Update(status string) {
	job.Updated = Now()
	job.SetStatus(status)
	job.SetCapability(job.GetCapability())
	job.TTL = Future(7 * 24)
}

func (job *Job) SetStatus(status string) {
	if status == Queued {
		job.Started, job.Finished = "", ""
	} else if job.Started == "" && status == Running {
		job.Started = Now()
	} else if job.Finished == "" && (status == Fail || status == Pass) {
		job.Finished = Now()
	}
	job.Status = fmt.Sprintf("%s#%s", status, job.Updated)
}

func (job *Job) SetCapability(capability string) {
	job.Source = fmt.Sprintf("%s#%s", capability, job.Updated)
}

func (job *Job) State() string {
	return strings.Split(job.Status, "#")[0]
}

func (job *Job) GetParent() Target {
	if job.Parent.Model != nil {
		return job.Parent.Model
	}
	return job.Target.Model
}

func (job *Job) Requested(capability string) bool {
	return slices.Contains(job.Capabilities, capability)
}

func (job *Job) Valid() bool {
	validPrefix := strings.HasPrefix(job.Key, "#job#")
	validCredentials := true
	for _, cred := range job.CredentialIDs {
		if cred == "" {
			validCredentials = false
		}
	}

	return validPrefix && validCredentials
}

func (job *Job) Defaulted() {
	job.initializeParameters()
	job.Created = Now()
	job.Updated = Now()
	job.TTL = Future(7 * 24)
	job.Queue = Standard
	job.Async = false
	job.SetStatus(Queued)
}

func (job *Job) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if job.Target.Model != nil {
					group := job.Target.Model.Group()
					identifier := job.Target.Model.Identifier()
					template := fmt.Sprintf("#job#%%s#%s#%s", identifier, job.GetCapability())
					if len(template) <= 1024 {
						shortenedDNS := group[:min(1024-len(template), len(group))]
						job.DNS = shortenedDNS
						job.Key = fmt.Sprintf(template, shortenedDNS)
					}
				}
				job.initializeParameters()
				job.originalStatus = job.Status
				job.SetCapability(job.GetCapability())
				return nil
			},
		},
	}
}

func (job *Job) initializeParameters() {
	if job.Config == nil {
		job.Config = make(map[string]string)
	}
	if job.Secret == nil {
		job.Secret = make(map[string]string)
	}
}

// Parameters returns a combined map of the Config and Secret maps
func (job *Job) Parameters() map[string]string {
	params := make(map[string]string)
	maps.Copy(params, job.Config)
	maps.Copy(params, job.Secret)

	return params
}

func (job *Job) Send(items ...registry.Model) {
	job.stream <- items
}

func (job *Job) Stream() <-chan []registry.Model {
	return job.stream
}

func (job *Job) IsDelayed() bool {
	return job.Delayed != ""
}

// Ensures the job's stream is open. Necessary when Unmarshaling a job from JSON.
func (job *Job) Open() {
	if job.stream == nil {
		job.stream = make(chan []registry.Model)
	}
}

func (job *Job) Close() {
	close(job.stream)
}

func NewJob(source string, target Target) Job {
	job := Job{
		Source: source,
		Target: TargetWrapper{Model: target},
		stream: make(chan []registry.Model),
	}
	job.Defaulted()
	registry.CallHooks(&job)
	return job
}

// Traced enables telemetry tracing for this job and its children.
// Only traced jobs emit TraceEvents. Call this for manually-triggered jobs.
func (job *Job) Traced() *Job {
	if job.TraceID == "" {
		job.TraceID = NewTraceID()
	}
	return job
}

// IsTraced returns true if this job has tracing enabled.
func (job *Job) IsTraced() bool {
	return job.TraceID != ""
}

func NewSystemJob(source, id string) Job {
	t := NewAsset(id, id) // not needed, but required for target to unmarshal. TODO make this a separate type
	j := Job{
		DNS:     id,
		Created: Now(),
		Updated: Now(),
		Queue:   Standard,
		Key:     fmt.Sprintf("#job#%s#system#%s", id, source),
		Target:  TargetWrapper{Model: &t},
		stream:  make(chan []registry.Model),
	}
	j.SetStatus(Queued)
	j.SetCapability(source)
	return j
}

// GetDescription returns a description for the Job model.
func (job *Job) GetDescription() string {
	return "Represents an asynchronous job or task, including its status, specification, and results."
}

func (job *Job) ToContext() ResultContext {
	// Extract agent client ID from config if present (for Aegis agora capabilities)
	agentClientID := ""
	if clientID, ok := job.Config["client_id"]; ok {
		agentClientID = clientID
	}

	return ResultContext{
		Username:      job.Username,
		Source:        job.GetCapability(),
		Config:        job.Config,
		Secret:        job.Secret,
		Target:        job.Target,
		Parent:        job.Parent,
		Queue:         job.Queue,
		Capabilities:  job.Capabilities,
		FullScan:      job.Full,
		AgentClientID: agentClientID,
		TraceID:       job.TraceID,
		CurrentSpanID: job.CurrentSpanID,
	}
}
