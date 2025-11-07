package model

import (
	"fmt"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type Record struct {
	registry.BaseModel
	baseTableModel
	Job
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the account." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the job." example:"#job#example.com#asset#portscan"`
}

func (r *Record) GetKey() string {
	return r.Key
}

func (r *Record) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				r.Key = fmt.Sprintf("#record#%s#%s", r.Job.Key, r.Job.Updated)
				return nil
			},
		},
	}
}

func (r *Record) Defaulted() {}

func NewRecord(job Job) Record {
	record := Record{
		Job: job,
	}
	registry.CallHooks(&record)
	return record
}
