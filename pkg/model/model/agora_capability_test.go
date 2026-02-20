package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCapabilityParameters(t *testing.T) {
	destinationStr := ""
	destinationInt := 0
	destinationBool := false

	params := []AgoraParameter{
		NewAgoraParameter("string-param", "test string", &destinationStr),
		NewAgoraParameter("int-param", "test integer", &destinationInt),
		NewAgoraParameter("bool-param", "test boolean", &destinationBool),
	}

	config := map[string]string{
		"string-param": "test",
		"int-param":    "1",
		"bool-param":   "true",
	}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
	require.Equal(t, "test", destinationStr)
	require.Equal(t, 1, destinationInt)
	require.Equal(t, true, destinationBool)
}

func TestParseCapabilityParameters_WithDefault(t *testing.T) {
	destinationStr := ""
	destinationInt := 0
	destinationBool := false

	params := []AgoraParameter{
		NewAgoraParameter("string-param", "test string", &destinationStr).WithDefault("default"),
		NewAgoraParameter("int-param", "test integer", &destinationInt).WithDefault("0"),
		NewAgoraParameter("bool-param", "test boolean", &destinationBool).WithDefault("true"),
	}

	config := map[string]string{}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
	require.Equal(t, "default", destinationStr)
	require.Equal(t, 0, destinationInt)
	require.Equal(t, true, destinationBool)
}

func TestParseCapabilityParameters_OverrideDefault(t *testing.T) {
	destinationStr := ""
	destinationInt := 0
	destinationBool := false

	params := []AgoraParameter{
		NewAgoraParameter("string-param", "test string", &destinationStr).WithDefault("default"),
		NewAgoraParameter("int-param", "test integer", &destinationInt).WithDefault("0"),
		NewAgoraParameter("bool-param", "test boolean", &destinationBool).WithDefault("true"),
	}

	config := map[string]string{
		"string-param": "test",
		"int-param":    "1",
		"bool-param":   "false",
	}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
	require.Equal(t, "test", destinationStr)
	require.Equal(t, 1, destinationInt)
	require.Equal(t, false, destinationBool)
}

func TestParseCapabilityParameters_Required(t *testing.T) {
	requiredString := ""

	params := []AgoraParameter{
		NewAgoraParameter("required-string", "test string", &requiredString).WithRequired(),
	}

	config := map[string]string{}

	err := ParseCapabilityParameters(params, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "is required")
}

func TestParseCapabilityParameters_WithParser(t *testing.T) {
	destinationStr := ""

	parser := func(s string) error {
		destinationStr = "parsed:" + s
		return nil
	}

	config := map[string]string{
		"string-param": "test",
	}

	params := []AgoraParameter{
		NewAgoraParameter("string-param", "test string", &destinationStr).WithParser(parser),
	}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
	require.Equal(t, "parsed:test", destinationStr)
}

func TestParseCapabilityParameters_OptionalBoolNoDefault(t *testing.T) {
	destinationBool := false
	destinationInt := 0

	params := []AgoraParameter{
		NewAgoraParameter("bool-param", "optional bool", &destinationBool),
		NewAgoraParameter("int-param", "optional int", &destinationInt),
	}

	// Empty config: no values provided, no defaults set
	config := map[string]string{}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
	require.Equal(t, false, destinationBool)
	require.Equal(t, 0, destinationInt)
}

func TestParseCapabilityParameters_NoParameters(t *testing.T) {
	params := []AgoraParameter{}

	config := map[string]string{
		"some-value": "some-value",
	}

	err := ParseCapabilityParameters(params, config)
	require.NoError(t, err)
}
