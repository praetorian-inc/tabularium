package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewThreat(t *testing.T) {
	threat := NewThreat("feed", "cve", "created", map[string]any{})
	assert.Equal(t, "#threat#feed#cve", threat.Key)
	assert.Equal(t, "#threat#feed#created", threat.Created)
}
