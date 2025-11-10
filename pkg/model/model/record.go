package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

const RecordTTLInHours = 24 * 90 // 90 days

func init() {
	registry.Registry.MustRegisterModel(&JobRecord{})
}

type JobRecord struct {
	registry.BaseModel
	baseTableModel
	Username   string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key        string `dynamodbav:"key" json:"key" desc:"Unique key for the job." example:"#job#example.com#asset#portscan"`
	JobKey     string `dynamodbav:"jobKey" json:"jobKey" desc:"Key of the job that was recorded" example:"#job#example.com#asset#portscan"`
	RecordTime string `dynamodbav:"recordTime" json:"recordTime" desc:"Timestamp when the recorded job was last updated (RFC3339)." example:"2023-10-27T10:00:00Z"`
	Full       bool   `dynamodbav:"full" json:"full" desc:"Whether the record is full." example:"true"`
	TTL        int64  `dynamodbav:"ttl" json:"ttl" desc:"Time-to-live for the record record (Unix timestamp)." example:"1706353200"`
}

func (r *JobRecord) GetDescription() string {
	return "Represents a record of a job. Used to track when a job was last successfully completed."
}

func (r *JobRecord) GetKey() string {
	return r.Key
}

func (r *JobRecord) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				r.Key = fmt.Sprintf("%s%s", RecordSearchKey(Job{Key: r.JobKey}), r.RecordTime)
				return nil
			},
		},
	}
}

func (r *JobRecord) Defaulted() {}

func NewRecord(job Job) JobRecord {
	record := JobRecord{
		JobKey:     job.Key,
		RecordTime: job.Updated,
		Full:       job.Full,
		TTL:        Future(RecordTTLInHours),
	}
	registry.CallHooks(&record)
	return record
}

func RecordSearchKey(job Job) string {
	return fmt.Sprintf("#jobrecord%s#", job.Key)
}
