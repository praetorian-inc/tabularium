package model

import (
	"fmt"
	"strings"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Status constants for monitoring sessions
const (
	MonitorStatusActive    = "active"
	MonitorStatusExpired   = "expired"
	MonitorStatusCancelled = "cancelled"
)

// Labels
const (
	MonitoringSessionLabel  = "MonitoringSession"
	MonitoredTechniqueLabel = "MonitoredTechnique"
	MonitorDetectionLabel   = "MonitorDetection"
	HasTechniqueLabel       = "HAS_TECHNIQUE"
	HasDetectionLabel       = "HAS_DETECTION"
)

func init() {
	registry.Registry.MustRegisterModel(&MonitoringSession{})
	registry.Registry.MustRegisterModel(&MonitoredTechnique{})
	registry.Registry.MustRegisterModel(&MonitorDetection{})
	registry.Registry.MustRegisterModel(&HasTechnique{})
	registry.Registry.MustRegisterModel(&HasDetection{})
	MustRegisterLabel(MonitoringSessionLabel)
	MustRegisterLabel(MonitoredTechniqueLabel)
	MustRegisterLabel(MonitorDetectionLabel)
}

// MonitorFilter defines a matching rule for EDR alert correlation.
type MonitorFilter struct {
	Type  string `json:"type"`  // "hostname" | "filehash" | "mitre"
	Value string `json:"value"`
}

// --- MonitoringSession ---

type MonitoringSession struct {
	registry.BaseModel
	Username   string `neo4j:"username" json:"username"`
	Key        string `neo4j:"key" json:"key"`
	SessionID  string `neo4j:"session_id" json:"session_id"`
	Name       string `neo4j:"name" json:"name"`
	Status     string `neo4j:"status" json:"status"`
	Created    string `neo4j:"created" json:"created"`
	ExpiresAt  string `neo4j:"expires_at" json:"expires_at"`
	ExecutedAt string `neo4j:"executed_at" json:"executed_at"`
	LastRunAt  string          `neo4j:"last_run_at" json:"last_run_at"`
	Filters    []MonitorFilter `neo4j:"filters" json:"filters"`
}

func NewMonitoringSession(sessionID, name string, filters []MonitorFilter, executedAt, expiresAt string) MonitoringSession {
	s := MonitoringSession{
		SessionID:  sessionID,
		Name:       name,
		Status:     MonitorStatusActive,
		Created:    Now(),
		ExpiresAt:  expiresAt,
		ExecutedAt: executedAt,
		Filters:    filters,
	}
	registry.CallHooks(&s)
	return s
}

func (s *MonitoringSession) GetKey() string   { return s.Key }
func (s *MonitoringSession) GetLabels() []string {
	return []string{MonitoringSessionLabel}
}
func (s *MonitoringSession) Valid() bool {
	return strings.HasPrefix(s.Key, "#monitoringsession#") && s.SessionID != ""
}
func (s *MonitoringSession) Standalone()            {}
func (s *MonitoringSession) SetUsername(u string) { s.Username = u }
func (s *MonitoringSession) GetAgent() string     { return "" }
func (s *MonitoringSession) GetDescription() string {
	return "A breach and attack validation monitoring session."
}
func (s *MonitoringSession) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Call: func() error {
			s.Key = fmt.Sprintf("#monitoringsession#%s", s.SessionID)
			return nil
		},
	}}
}

// --- MonitoredTechnique ---

type MonitoredTechnique struct {
	registry.BaseModel
	Username    string `neo4j:"username" json:"username"`
	Key         string `neo4j:"key" json:"key"`
	TechniqueID string `neo4j:"technique_id" json:"technique_id"`
	Name        string `neo4j:"name" json:"name"`
}

func NewMonitoredTechnique(techniqueID, name string) MonitoredTechnique {
	t := MonitoredTechnique{
		TechniqueID: techniqueID,
		Name:        name,
	}
	registry.CallHooks(&t)
	return t
}

func (t *MonitoredTechnique) GetKey() string   { return t.Key }
func (t *MonitoredTechnique) GetLabels() []string {
	return []string{MonitoredTechniqueLabel}
}
func (t *MonitoredTechnique) Valid() bool {
	return strings.HasPrefix(t.Key, "#monitoredtechnique#") && t.TechniqueID != ""
}
func (t *MonitoredTechnique) Standalone()            {}
func (t *MonitoredTechnique) SetUsername(u string) { t.Username = u }
func (t *MonitoredTechnique) GetAgent() string     { return "" }
func (t *MonitoredTechnique) GetDescription() string {
	return "A MITRE ATT&CK technique being monitored for detection."
}
func (t *MonitoredTechnique) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Call: func() error {
			t.Key = fmt.Sprintf("#monitoredtechnique#%s", t.TechniqueID)
			return nil
		},
	}}
}

// --- MonitorAlert ---

// MonitorAlert is a normalized EDR alert. Transient input to matching, not persisted as a graph node.
type MonitorAlert struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Severity        string   `json:"severity"`
	DetectedAt      string   `json:"detected_at"`
	Hostname        string   `json:"hostname"`
	Hashes          []string `json:"hashes,omitempty"`
	MitreTechniques []string `json:"mitre_techniques,omitempty"`
	Evidence        string   `json:"evidence,omitempty"`
	SourceURL       string   `json:"source_url,omitempty"`
}

// --- MonitorDetection ---

// MonitorDetection is a graph node created when an alert matches a session technique.
type MonitorDetection struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username"`
	Key      string `neo4j:"key" json:"key"`

	// Match result
	SessionID   string `neo4j:"session_id" json:"session_id"`
	TechniqueID string `neo4j:"technique_id" json:"technique_id"`
	Source      string `neo4j:"source" json:"source"`             // "defender", "extrahop"
	MatchMethod string `neo4j:"match_method" json:"match_method"` // "mitre", "filehash", "llm"
	Latency     string `neo4j:"latency" json:"latency"`
	LLMScore    int    `neo4j:"llm_score" json:"llm_score,omitempty"`
	LLMReason   string `neo4j:"llm_reason" json:"llm_reason,omitempty"`

	// Alert snapshot (for display without re-fetching)
	AlertID     string `neo4j:"alert_id" json:"alert_id"`
	Title       string `neo4j:"title" json:"title"`
	Description string `neo4j:"description" json:"description"`
	Severity    string `neo4j:"severity" json:"severity"`
	DetectedAt  string `neo4j:"detected_at" json:"detected_at"`
	Hostname    string `neo4j:"hostname" json:"hostname"`
	SourceURL   string `neo4j:"source_url" json:"source_url,omitempty"`
}

func NewMonitorDetection(alert *MonitorAlert, sessionID, techniqueID, source, method string) MonitorDetection {
	d := MonitorDetection{
		SessionID:   sessionID,
		TechniqueID: techniqueID,
		Source:      source,
		MatchMethod: method,
		AlertID:     alert.ID,
		Title:       alert.Title,
		Description: alert.Description,
		Severity:    alert.Severity,
		DetectedAt:  alert.DetectedAt,
		Hostname:    alert.Hostname,
		SourceURL:   alert.SourceURL,
	}
	registry.CallHooks(&d)
	return d
}

func (d *MonitorDetection) GetKey() string { return d.Key }
func (d *MonitorDetection) GetLabels() []string {
	return []string{MonitorDetectionLabel}
}
func (d *MonitorDetection) Valid() bool {
	return strings.HasPrefix(d.Key, "#monitordetection#") && d.AlertID != ""
}
func (d *MonitorDetection) Standalone()            {}
func (d *MonitorDetection) SetUsername(u string) { d.Username = u }
func (d *MonitorDetection) GetAgent() string     { return "" }
func (d *MonitorDetection) GetDescription() string {
	return "A detection event from an EDR matched to a monitored technique."
}
func (d *MonitorDetection) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Call: func() error {
			d.Key = fmt.Sprintf("#monitordetection#%s#%s#%s#%s",
				d.SessionID, d.TechniqueID, d.Source, d.AlertID)
			return nil
		},
	}}
}

// --- Relationships ---

type HasTechnique struct {
	*BaseRelationship
}

func NewHasTechnique(session *MonitoringSession, technique *MonitoredTechnique) *HasTechnique {
	return &HasTechnique{
		BaseRelationship: NewBaseRelationship(session, technique, HasTechniqueLabel),
	}
}

func (ht HasTechnique) Label() string { return HasTechniqueLabel }

func (ht *HasTechnique) GetDescription() string {
	return "Links a monitoring session to a technique being tested."
}

type HasDetection struct {
	*BaseRelationship
}

func NewHasDetection(technique *MonitoredTechnique, detection *MonitorDetection) *HasDetection {
	return &HasDetection{
		BaseRelationship: NewBaseRelationship(technique, detection, HasDetectionLabel),
	}
}

func (hd HasDetection) Label() string { return HasDetectionLabel }

func (hd *HasDetection) GetDescription() string {
	return "Links a monitored technique to a detection from an EDR."
}
