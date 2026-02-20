package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMonitoringSession(t *testing.T) {
	filters := []MonitorFilter{{Type: "hostname", Value: "dc01.corp.local"}}
	s := NewMonitoringSession("sess-123", "Test Session", filters, "2026-01-01T00:00:00Z", "2026-01-08T00:00:00Z")

	assert.Equal(t, "#monitoringsession#sess-123", s.Key)
	assert.Equal(t, "sess-123", s.SessionID)
	assert.Equal(t, "Test Session", s.Name)
	assert.Equal(t, MonitorStatusActive, s.Status)
	assert.Equal(t, "2026-01-01T00:00:00Z", s.ExecutedAt)
	assert.Equal(t, "2026-01-08T00:00:00Z", s.ExpiresAt)
	assert.NotEmpty(t, s.Created)
	assert.True(t, s.Valid())
	assert.Equal(t, []string{MonitoringSessionLabel}, s.GetLabels())
	assert.Len(t, s.Filters, 1)
}

func TestMonitoringSession_Invalid(t *testing.T) {
	s := MonitoringSession{Key: "bad-key", SessionID: "x"}
	assert.False(t, s.Valid())

	s2 := MonitoringSession{Key: "#monitoringsession#x"}
	assert.False(t, s2.Valid())
}

func TestNewMonitoredTechnique(t *testing.T) {
	tech := NewMonitoredTechnique("T1003.001", "OS Credential Dumping: LSASS Memory")

	assert.Equal(t, "#monitoredtechnique#T1003.001", tech.Key)
	assert.Equal(t, "T1003.001", tech.TechniqueID)
	assert.Equal(t, "OS Credential Dumping: LSASS Memory", tech.Name)
	assert.True(t, tech.Valid())
	assert.Equal(t, []string{MonitoredTechniqueLabel}, tech.GetLabels())
}

func TestMonitoredTechnique_GlobalKey(t *testing.T) {
	t1 := NewMonitoredTechnique("T1003.001", "Name A")
	t2 := NewMonitoredTechnique("T1003.001", "Name B")
	assert.Equal(t, t1.Key, t2.Key, "Same technique ID should produce same key (global)")
}

func TestNewMonitorDetection(t *testing.T) {
	alert := &MonitorAlert{
		ID:       "alert-456",
		Title:    "Suspicious Process",
		Hostname: "dc01",
	}
	d := NewMonitorDetection(alert, "sess-123", "T1003.001", "defender", "mitre")

	assert.Equal(t, "#monitordetection#sess-123#T1003.001#defender#alert-456", d.Key)
	assert.Equal(t, "alert-456", d.AlertID)
	assert.Equal(t, "Suspicious Process", d.Title)
	assert.Equal(t, "dc01", d.Hostname)
	assert.Equal(t, "mitre", d.MatchMethod)
	assert.True(t, d.Valid())
	assert.Equal(t, []string{MonitorDetectionLabel}, d.GetLabels())
}

func TestMonitorDetection_InvalidWithoutAlertID(t *testing.T) {
	d := &MonitorDetection{
		Key:       "#monitordetection#sess#tech#src#",
		SessionID: "sess",
	}
	assert.False(t, d.Valid())
}

func TestHasTechnique(t *testing.T) {
	session := &MonitoringSession{Key: "#monitoringsession#s1"}
	technique := &MonitoredTechnique{Key: "#monitoredtechnique#T1003"}

	rel := NewHasTechnique(session, technique)
	assert.Equal(t, HasTechniqueLabel, rel.Label())
	assert.True(t, rel.Valid())

	src, tgt := rel.Nodes()
	assert.Equal(t, session, src)
	assert.Equal(t, technique, tgt)
}

func TestHasDetection(t *testing.T) {
	technique := &MonitoredTechnique{Key: "#monitoredtechnique#T1003"}
	detection := &MonitorDetection{Key: "#monitordetection#s1#T1003#defender#a1"}

	rel := NewHasDetection(technique, detection)
	assert.Equal(t, HasDetectionLabel, rel.Label())
	assert.True(t, rel.Valid())

	src, tgt := rel.Nodes()
	assert.Equal(t, technique, src)
	assert.Equal(t, detection, tgt)
}
