package model

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Job struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Username who initiated or owns the job." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the job." example:"#job#example.com#asset#portscan"`
	// Attributes
	DNS                   string            `dynamodbav:"dns" json:"dns" desc:"Primary DNS associated with the job's target." example:"example.com"`
	Source                string            `dynamodbav:"source" json:"source" desc:"The source or capability that generated this job." example:"portscan"`
	Comment               string            `dynamodbav:"comment" json:"comment,omitempty" desc:"Optional comment about the job." example:"Scanning standard web ports"`
	Created               string            `dynamodbav:"created" json:"created" desc:"Timestamp when the job was created (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Updated               string            `dynamodbav:"updated" json:"updated" desc:"Timestamp when the job was last updated (RFC3339)." example:"2023-10-27T10:05:00Z"`
	Delayed               string            `dynamodbav:"delayed,omitempty" json:"delayed,omitempty" desc:"Timestamp that this job should be delayed until" example:"2023-10-27T10:00:00Z"`
	Status                string            `dynamodbav:"status" json:"status" desc:"Current status of the job (e.g., JQ#portscan)." example:"JQ#portscan"`
	TTL                   int64             `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the job record (Unix timestamp)." example:"1706353200"`
	Name                  string            `dynamodbav:"name,omitempty" json:"name,omitempty" desc:"The IP address this job was executed from" example:"1.2.3.4"`
	Config                map[string]string `dynamodbav:"-" json:"config" desc:"Configuration parameters for the job capability." example:"{\"test\": \"cve-1111-2222\"}"`
	LargeArtifactFileName string            `dynamodbav:"largeArtifactFileName" json:"largeArtifactFileName,omitempty" desc:"The name of the file that contains the large artifacts." example:"large_artifact.zip"`
	S3DownloadURL         string            `dynamodbav:"s3DownloadURL" json:"s3DownloadURL,omitempty" desc:"The URL of the file that contains the large output." example:"https://s3.amazonaws.com/big_output.zip"`
	Full                  bool              `dynamodbav:"-" json:"full,omitempty" desc:"Indicates if this is a full scan job." example:"false"`
	Capabilities          []string          `dynamodbav:"-" json:"capabilities,omitempty" desc:"List of specific capabilities to run for this job." example:"[\"portscan\", \"nuclei\"]"`
	Queue                 string            `dynamodbav:"-" desc:"Target queue for the job." example:"standard"`
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

func (job *Job) Update(status string) {
	job.SetStatus(status)
	job.Updated = Now()

	job.TTL = Future(7 * 24)
}

func (job *Job) SetStatus(status string) {
	job.Status = fmt.Sprintf("%s#%s", status, job.Source)
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
	return strings.HasPrefix(job.Key, "#job#")
}

func (job *Job) Defaulted() {
	job.Status = fmt.Sprintf("%s#%s", Queued, job.Source)
	job.Config = make(map[string]string)
	job.Created = Now()
	job.Updated = Now()
	job.TTL = Future(7 * 24)
	job.Queue = Standard
	job.Async = false
}

func (job *Job) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				if job.Target.Model != nil {
					template := fmt.Sprintf("#job#%%s#%s#%s", job.Target.Model.Identifier(), job.Source)
					if len(template) <= 1024 {
						shortenedDNS := job.Target.Model.Group()[:min(1024-len(template), len(job.Target.Model.Group()))]
						job.DNS = shortenedDNS
						job.Key = fmt.Sprintf(template, shortenedDNS)
					}
				}
				job.originalStatus = job.Status
				return nil
			},
		},
	}
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

func NewSystemJob(source, id string) Job {
	t := NewAsset(id, id) // not needed, but required for target to unmarshal. TODO make this a separate type
	return Job{
		DNS:     id,
		Source:  source,
		Status:  fmt.Sprintf("%s#%s", Queued, source),
		Created: Now(),
		Updated: Now(),
		Queue:   Standard,
		Key:     fmt.Sprintf("#job#%s#system#%s", id, source),
		Target:  TargetWrapper{Model: &t},
		stream:  make(chan []registry.Model),
	}
}

// GetDescription returns a description for the Job model.
func (job *Job) GetDescription() string {
	return "Represents an asynchronous job or task, including its status, specification, and results."
}

func (job *Job) ToContext() ResultContext {
	return ResultContext{
		Username:     job.Username,
		Source:       job.Source,
		Config:       job.Config,
		Target:       job.Target,
		Parent:       job.Parent,
		Queue:        job.Queue,
		Capabilities: job.Capabilities,
	}
}
