package model

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

// Label constants
const (
	MonitoringSessionLabel  = "MonitoringSession"
	MonitoredTechniqueLabel = "MonitoredTechnique"
	MonitorDetectionLabel   = "MonitorDetection"
)

// Status constants for MonitoringSession
const (
	MonitorStatusActive    = "active"
	MonitorStatusExpired   = "expired"
	MonitorStatusCancelled = "cancelled"
)

// Relationship label constants
const (
	HasTechniqueLabel = "HAS_TECHNIQUE"
	HasDetectionLabel = "HAS_DETECTION"
)

func init() {
	registry.Registry.MustRegisterModel(&MonitoringSession{})
	registry.Registry.MustRegisterModel(&MonitoredTechnique{})
	registry.Registry.MustRegisterModel(&MonitorDetection{})
	registry.Registry.MustRegisterModel(&HasTechnique{})
	registry.Registry.MustRegisterModel(&HasDetection{})
}

// Compile-time interface checks
var _ GraphModel = (*MonitoringSession)(nil)
var _ GraphModel = (*MonitoredTechnique)(nil)
var _ GraphModel = (*MonitorDetection)(nil)

// ============================================================================
// MonitoringSession
// ============================================================================

var monitoringSessionKeyRegex = regexp.MustCompile(`^#monitoring_session#.+$`)

// MonitoringSession represents a breach and attack validation monitoring session.
type MonitoringSession struct {
	registry.BaseModel
	Username   string `neo4j:"username" json:"username"`
	Key        string `neo4j:"key" json:"key"`
	SessionID  string `neo4j:"session_id" json:"session_id"`
	Name       string `neo4j:"name" json:"name"`
	Status     string `neo4j:"status" json:"status"`
	CreatedAt  string `neo4j:"created_at" json:"created_at"`
	ExpiresAt  string `neo4j:"expires_at" json:"expires_at"`
	ExecutedAt string `neo4j:"executed_at" json:"executed_at"`
	LastRunAt  string `neo4j:"last_run_at" json:"last_run_at"`
	Filters    string `neo4j:"filters" json:"filters"`
}

// NewMonitoringSession creates a new MonitoringSession with hooks applied.
func NewMonitoringSession(sessionID, name, filters, executedAt, expiresAt string) *MonitoringSession {
	s := &MonitoringSession{
		SessionID:  sessionID,
		Name:       name,
		Status:     MonitorStatusActive,
		CreatedAt:  Now(),
		ExpiresAt:  expiresAt,
		ExecutedAt: executedAt,
		Filters:    filters,
	}
	registry.CallHooks(s)
	return s
}

// GraphModel interface
func (m *MonitoringSession) GetLabels() []string { return []string{MonitoringSessionLabel} }
func (m *MonitoringSession) GetKey() string       { return m.Key }
func (m *MonitoringSession) Valid() bool           { return monitoringSessionKeyRegex.MatchString(m.Key) }
func (m *MonitoringSession) SetUsername(u string)  { m.Username = u }
func (m *MonitoringSession) GetAgent() string      { return "" }

func (m *MonitoringSession) GetDescription() string {
	return "Represents a breach and attack validation monitoring session."
}

func (m *MonitoringSession) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Description: "set Key from SessionID",
		Call: func() error {
			if m.SessionID != "" && m.Key == "" {
				m.Key = fmt.Sprintf("#monitoring_session#%s", m.SessionID)
			}
			return nil
		},
	}}
}

// ============================================================================
// MonitoredTechnique — global reference node, keyed by technique ID only
// ============================================================================

var monitoredTechniqueKeyRegex = regexp.MustCompile(`^#monitored_technique#.+$`)

// MonitoredTechnique represents a MITRE ATT&CK technique. Keyed globally by
// technique ID — shared across all sessions. Session-specific data (executed_at,
// description, status) lives on the HAS_TECHNIQUE relationship.
type MonitoredTechnique struct {
	registry.BaseModel
	Username    string `neo4j:"username" json:"username"`
	Key         string `neo4j:"key" json:"key"`
	TechniqueID string `neo4j:"technique_id" json:"technique_id"`
	Name        string `neo4j:"name" json:"name"`
}

// NewMonitoredTechnique creates a new MonitoredTechnique with hooks applied.
func NewMonitoredTechnique(techniqueID, name string) *MonitoredTechnique {
	t := &MonitoredTechnique{
		TechniqueID: techniqueID,
		Name:        name,
	}
	registry.CallHooks(t)
	return t
}

func (m *MonitoredTechnique) GetLabels() []string { return []string{MonitoredTechniqueLabel} }
func (m *MonitoredTechnique) GetKey() string       { return m.Key }
func (m *MonitoredTechnique) Valid() bool           { return monitoredTechniqueKeyRegex.MatchString(m.Key) }
func (m *MonitoredTechnique) SetUsername(u string)  { m.Username = u }
func (m *MonitoredTechnique) GetAgent() string      { return "" }

func (m *MonitoredTechnique) GetDescription() string {
	return "Represents a MITRE ATT&CK technique (global reference node)."
}

func (m *MonitoredTechnique) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Description: "set Key from TechniqueID",
		Call: func() error {
			if m.TechniqueID != "" && m.Key == "" {
				m.Key = fmt.Sprintf("#monitored_technique#%s", m.TechniqueID)
			}
			return nil
		},
	}}
}

// ============================================================================
// MonitorDetection — detection event, keyed per session+technique+source+id
// ============================================================================

var monitorDetectionKeyRegex = regexp.MustCompile(`^#monitor_detection#[^#]+#[^#]+#[^#]+#.+$`)

// MonitorDetection represents a normalized EDR alert.
// Monitor integrations produce these directly; the matcher decides which to keep.
// SessionID, TechniqueID, Source, MatchMethod, and Latency are populated by the matcher on match.
type MonitorDetection struct {
	registry.BaseModel
	Username string `neo4j:"username" json:"username"`
	Key      string `neo4j:"key" json:"key"`

	// Populated by monitor integration (raw alert data)
	DetectionID     string   `neo4j:"detection_id" json:"detection_id"`
	Title           string   `neo4j:"title" json:"title"`
	Description     string   `neo4j:"description" json:"description"`
	Severity        string   `neo4j:"severity" json:"severity"`
	DetectedAt      string   `neo4j:"detected_at" json:"detected_at"`
	Hostname        string   `neo4j:"hostname" json:"hostname"`
	MitreTechniques []string `neo4j:"mitre_techniques" json:"mitre_techniques"`
	SHA1            string   `neo4j:"sha1" json:"sha1,omitempty"`
	SHA256          string   `neo4j:"sha256" json:"sha256,omitempty"`
	Evidence        string   `neo4j:"evidence" json:"evidence"`
	SourceURL       string   `neo4j:"source_url" json:"source_url"`

	// Populated by matcher on match
	SessionID   string `neo4j:"session_id" json:"session_id"`
	TechniqueID string `neo4j:"technique_id" json:"technique_id"`
	Source      string `neo4j:"source" json:"source"`
	MatchMethod string `neo4j:"match_method" json:"match_method"`
	Latency     string `neo4j:"latency" json:"latency"`
}

// RerunHooks clears the key and re-runs hooks to regenerate it.
func RerunHooks(d *MonitorDetection) {
	registry.CallHooks(d)
}

func (m *MonitorDetection) GetLabels() []string { return []string{MonitorDetectionLabel} }
func (m *MonitorDetection) GetKey() string       { return m.Key }
func (m *MonitorDetection) Valid() bool           { return monitorDetectionKeyRegex.MatchString(m.Key) }
func (m *MonitorDetection) SetUsername(u string)  { m.Username = u }
func (m *MonitorDetection) GetAgent() string      { return "" }

func (m *MonitorDetection) GetDescription() string {
	return "Represents a single EDR detection matched to a monitored technique."
}

func (m *MonitorDetection) GetHooks() []registry.Hook {
	return []registry.Hook{{
		Description: "set Key from SessionID, TechniqueID, Source, DetectionID",
		Call: func() error {
			if m.SessionID != "" && m.TechniqueID != "" && m.Source != "" && m.DetectionID != "" && m.Key == "" {
				m.Key = fmt.Sprintf("#monitor_detection#%s#%s#%s#%s", m.SessionID, m.TechniqueID, m.Source, m.DetectionID)
			}
			return nil
		},
	}}
}

// ============================================================================
// Relationships
// ============================================================================

// HasTechnique links MonitoringSession → MonitoredTechnique.
// Carries session-specific properties: executed_at, description, and cached status.
type HasTechnique struct {
	*BaseRelationship
	ExecutedAt  string `neo4j:"executed_at" json:"executed_at"`
	Description string `neo4j:"description" json:"description"`
	Status      string `neo4j:"status" json:"status"`
}

// NewHasTechnique creates a HAS_TECHNIQUE relationship with session-specific properties.
func NewHasTechnique(session *MonitoringSession, technique *MonitoredTechnique, executedAt, description string) GraphRelationship {
	return &HasTechnique{
		BaseRelationship: NewBaseRelationship(session, technique, HasTechniqueLabel),
		ExecutedAt:       executedAt,
		Description:      description,
		Status:           "Undetected",
	}
}

func (h *HasTechnique) Label() string { return HasTechniqueLabel }

func (h *HasTechnique) GetDescription() string {
	return "Links a monitoring session to a declared technique with session-specific execution context."
}

// HasDetection links MonitoredTechnique → MonitorDetection.
type HasDetection struct {
	*BaseRelationship
}

func NewHasDetection(technique *MonitoredTechnique, detection *MonitorDetection) GraphRelationship {
	return &HasDetection{
		BaseRelationship: NewBaseRelationship(technique, detection, HasDetectionLabel),
	}
}

func (h *HasDetection) Label() string { return HasDetectionLabel }

func (h *HasDetection) GetDescription() string {
	return "Links a monitored technique to a detection event."
}

// ============================================================================
// MonitorFilter — JSON helper for session filters
// ============================================================================

// MonitorFilter defines a matching criterion stored as JSON in MonitoringSession.Filters.
type MonitorFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func ParseFilters(jsonStr string) ([]MonitorFilter, error) {
	if jsonStr == "" {
		return nil, nil
	}
	var f []MonitorFilter
	if err := json.Unmarshal([]byte(jsonStr), &f); err != nil {
		return nil, fmt.Errorf("failed to parse filters: %w", err)
	}
	return f, nil
}

func SerializeFilters(filters []MonitorFilter) (string, error) {
	if len(filters) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(filters)
	if err != nil {
		return "", fmt.Errorf("failed to serialize filters: %w", err)
	}
	return string(data), nil
}


