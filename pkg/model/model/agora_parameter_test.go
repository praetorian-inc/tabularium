package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgoraParameter(t *testing.T) {
	destinationStr := ""
	ap := NewAgoraParameter("test", "test", &destinationStr)
	inputStr := "test"
	err := ap.Parse(&inputStr)
	require.NoError(t, err)
	require.Equal(t, "test", destinationStr)

	ap = ap.WithDefault("default")
	err = ap.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "default", destinationStr)

	destinationInt := 0
	ap = NewAgoraParameter("test", "test", &destinationInt)
	inputInt := "1"
	err = ap.Parse(&inputInt)
	require.NoError(t, err)
	require.Equal(t, 1, destinationInt)

	destinationBool := false
	ap = NewAgoraParameter("test", "test", &destinationBool)
	inputBool := "true"
	err = ap.Parse(&inputBool)
	require.NoError(t, err)
	require.True(t, destinationBool)

	unsupportedType := struct{}{}
	ap = NewAgoraParameter("test", "test", &unsupportedType)
	inputUnsupported := "test"
	err = ap.Parse(&inputUnsupported)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown/unsupported type")
}
