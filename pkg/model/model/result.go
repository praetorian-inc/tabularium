package model

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type ResultContext struct {
	Username     string            `json:"username" desc:"Username who initiated or owns the job."`
	Source       string            `json:"source" desc:"The source or capability that generated this job."`
	Config       map[string]string `json:"config" desc:"Configuration parameters for the job capability."`
	Secret       map[string]string `json:"secret" desc:"Sensitive configuration parameters (credentials, tokens, keys)."`
	Target       TargetWrapper     `json:"target" desc:"The primary target of the job."`
	Parent       TargetWrapper     `json:"parent,omitempty" desc:"Optional parent target from which this job was spawned."`
	Origin       TargetWrapper     `json:"origin" desc:"The origin for this chain of jobs. Defaults to target unless set here."`
	Queue        string            `json:"queue,omitempty" desc:"Target queue for the job."`
	Capabilities []string          `json:"capabilities,omitempty" desc:"List of specific capabilities to run for this job."`
}

func _importEntity(entity string, config map[string]string) bool {
	if value, ok := config[entity]; ok {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			slog.Error(fmt.Sprintf("Error parsing %s config value", entity), "error", err)
			return false
		}
		return parsed
	}
	return true
}

func (rc *ResultContext) ImportAssets() bool {
	return _importEntity("importAssets", rc.Config)
}

func (rc *ResultContext) ImportVulnerabilities() bool {
	return _importEntity("importVulnerabilities", rc.Config)
}

func (rc *ResultContext) GetParent() Target {
	if rc.Parent.Model != nil {
		return ReturnParent(rc.Parent.Model)
	}
	return ReturnParent(rc.Target.Model)
}

func ReturnParent(parent Target) Target {
	if IsNoInput(parent) {
		return nil
	}
	return parent
}

func (rc *ResultContext) GetOrigin() Target {
	if rc.Origin.Model != nil {
		return rc.Origin.Model
	}
	return rc.Target.Model
}

type SpawnJobOption func(job *Job)

func (rc *ResultContext) SpawnJob(source string, target Target, config map[string]string) Job {
	job := NewJob(source, target)
	if config != nil {
		job.Config = config
	}
	job.Capabilities = rc.Capabilities
	job.Origin = rc.Origin
	if job.Origin.Model == nil {
		job.Origin = TargetWrapper{Model: rc.GetParent()}
	}
	return job
}

type Result struct {
	registry.BaseModel
	Context ResultContext    `json:"context" desc:"The context associated with this result."`
	Items   []registry.Model `json:"items" desc:"The actual result items."`
}

func init() {
	registry.Registry.MustRegisterModel(&Result{})
}

// GetDescription returns a description for the Result model.
func (r *Result) GetDescription() string {
	return "Represents the result of a job, encapsulating the job details and the resulting item(s)."
}
