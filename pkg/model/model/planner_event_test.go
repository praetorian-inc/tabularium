package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeOutput_Marshal(t *testing.T) {
	output := FinalizeOutput{
		Summary: "Analysis complete",
		Data: map[string]interface{}{
			"count":  5,
			"status": "success",
		},
		Recommendations: []string{
			"Review findings",
			"Update configuration",
		},
	}

	bytes, err := json.Marshal(output)
	require.NoError(t, err)

	var decoded FinalizeOutput
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)

	assert.Equal(t, output.Summary, decoded.Summary)
	assert.Equal(t, 2, len(decoded.Data))
	assert.Equal(t, 2, len(decoded.Recommendations))
}
