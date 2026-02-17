package model

import (
	"testing"

	"github.com/praetorian-inc/tabularium/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- MonitoringSession ----

func TestNewMonitoringSession(t *testing.T) {
	s := NewMonitoringSession("abc123", "LSASS test", `[{"type":"hostname","value":"DC01"}]`, "2026-02-17T10:00:00Z", "2026-02-24T10:00:00Z")

	assert.Equal(t, "#monitoring_session#abc123", s.Key)
	assert.Equal(t, "abc123", s.SessionID)
	assert.Equal(t, "LSASS test", s.Name)
	assert.Equal(t, MonitorStatusActive, s.Status)
	assert.NotEmpty(t, s.CreatedAt)
	assert.Equal(t, "2026-02-24T10:00:00Z", s.ExpiresAt)
	assert.Equal(t, "2026-02-17T10:00:00Z", s.ExecutedAt)
	assert.True(t, s.Valid())
}

func TestMonitoringSession_Valid(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{"valid key", "#monitoring_session#abc123", true},
		{"empty key", "", false},
		{"wrong prefix", "#session#abc123", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MonitoringSession{Key: tt.key}
			assert.Equal(t, tt.valid, m.Valid())
		})
	}
}

func TestMonitoringSession_Hooks(t *testing.T) {
	m := &MonitoringSession{SessionID: "test123"}
	registry.CallHooks(m)
	assert.Equal(t, "#monitoring_session#test123", m.Key)
}

// ---- MonitoredTechnique ----

func TestNewMonitoredTechnique(t *testing.T) {
	tech := NewMonitoredTechnique("T1003.001", "OS Credential Dumping: LSASS Memory")

	assert.Equal(t, "#monitored_technique#T1003.001", tech.Key)
	assert.Equal(t, "T1003.001", tech.TechniqueID)
	assert.Equal(t, "OS Credential Dumping: LSASS Memory", tech.Name)
	assert.True(t, tech.Valid())
}

func TestMonitoredTechnique_GlobalKey(t *testing.T) {
	tech := NewMonitoredTechnique("T1059.001", "PowerShell")
	assert.Equal(t, "#monitored_technique#T1059.001", tech.Key)
	assert.NotContains(t, tech.Key, "session")
}

func TestMonitoredTechnique_Valid(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{"valid sub-technique", "#monitored_technique#T1003.001", true},
		{"valid parent", "#monitored_technique#T1003", true},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MonitoredTechnique{Key: tt.key}
			assert.Equal(t, tt.valid, m.Valid())
		})
	}
}

// ---- MonitorDetection ----

func TestMonitorDetection_KeyGeneration(t *testing.T) {
	d := &MonitorDetection{
		SessionID:   "sess1",
		TechniqueID: "T1003.001",
		Source:      "defender",
		DetectionID: "det456",
	}
	RerunHooks(d)

	assert.Equal(t, "#monitor_detection#sess1#T1003.001#defender#det456", d.Key)
	assert.True(t, d.Valid())
}

func TestMonitorDetection_Valid(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{"valid key", "#monitor_detection#s1#T1003.001#defender#det1", true},
		{"missing parts", "#monitor_detection#s1#T1003.001", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MonitorDetection{Key: tt.key}
			assert.Equal(t, tt.valid, m.Valid())
		})
	}
}

// ---- Relationships ----

func TestNewHasTechnique(t *testing.T) {
	session := NewMonitoringSession("s1", "test", "[]", "2026-02-17T10:00:00Z", "")
	technique := NewMonitoredTechnique("T1003.001", "LSASS")

	rel := NewHasTechnique(session, technique, "2026-02-17T10:00:00Z", "procdump on DC01")
	ht := rel.(*HasTechnique)

	assert.Equal(t, HasTechniqueLabel, rel.Label())
	assert.Equal(t, "2026-02-17T10:00:00Z", ht.ExecutedAt)
	assert.Equal(t, "procdump on DC01", ht.Description)
	assert.Equal(t, "Undetected", ht.Status)

	src, tgt := rel.Nodes()
	assert.Equal(t, session.GetKey(), src.GetKey())
	assert.Equal(t, technique.GetKey(), tgt.GetKey())
}

func TestHasTechnique_Properties(t *testing.T) {
	session := NewMonitoringSession("s1", "test", "[]", "2026-02-17T10:00:00Z", "")
	technique := NewMonitoredTechnique("T1003.001", "LSASS")

	rel := NewHasTechnique(session, technique, "2026-02-17T10:00:00Z", "ran procdump")
	ht := rel.(*HasTechnique)

	assert.Equal(t, "2026-02-17T10:00:00Z", ht.ExecutedAt)
	assert.Equal(t, "ran procdump", ht.Description)
	assert.Equal(t, "Undetected", ht.Status)

	base := rel.Base()
	assert.NotEmpty(t, base.Created)
	assert.NotEmpty(t, base.Key)
}

func TestNewHasDetection(t *testing.T) {
	technique := NewMonitoredTechnique("T1003.001", "LSASS")
	detection := &MonitorDetection{
		SessionID:   "s1",
		TechniqueID: "T1003.001",
		Source:      "defender",
		DetectionID: "det1",
	}
	RerunHooks(detection)

	rel := NewHasDetection(technique, detection)
	assert.Equal(t, HasDetectionLabel, rel.Label())

	src, tgt := rel.Nodes()
	assert.Equal(t, technique.GetKey(), src.GetKey())
	assert.Equal(t, detection.GetKey(), tgt.GetKey())
}

// ---- Helpers ----

func TestParseFilters(t *testing.T) {
	j := `[{"type":"hostname","value":"DC01"},{"type":"filehash","value":"abc123"}]`
	f, err := ParseFilters(j)
	require.NoError(t, err)
	require.Len(t, f, 2)
	assert.Equal(t, "hostname", f[0].Type)
	assert.Equal(t, "DC01", f[0].Value)
}

func TestParseFilters_Empty(t *testing.T) {
	f, err := ParseFilters("")
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestSerializeFilters_Roundtrip(t *testing.T) {
	original := []MonitorFilter{{Type: "hostname", Value: "DC01"}}
	s, err := SerializeFilters(original)
	require.NoError(t, err)

	parsed, err := ParseFilters(s)
	require.NoError(t, err)
	require.Len(t, parsed, 1)
	assert.Equal(t, "DC01", parsed[0].Value)
}

// ---- Registration ----

func TestMonitoringModels_Registered(t *testing.T) {
	names := []string{"monitoringsession", "monitoredtechnique", "monitordetection"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			_, ok := registry.Registry.MakeType(name)
			assert.True(t, ok, "type %q not found in registry", name)
		})
	}
}
