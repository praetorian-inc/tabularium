package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_MarshalJSON(t *testing.T) {
	file := NewFile("test.txt")
	file.Bytes = []byte("test")

	jsonBytes, err := json.Marshal(file)
	require.Nil(t, err)

	assert.Contains(t, string(jsonBytes), `"name":"test.txt"`)
	assert.Contains(t, string(jsonBytes), `"bytes":"test"`)
	assert.NotContains(t, string(jsonBytes), `"encoded"`)
}

func TestFile_MarshalJSON_Encoded(t *testing.T) {
	file := NewFile("test.txt")
	file.Bytes = []byte("this has a bad character: \x7f")

	jsonBytes, err := json.Marshal(file)
	require.Nil(t, err)
	assert.Contains(t, string(jsonBytes), `"bytes":"base64:dGhpcyB`)
}

func TestFile_UnmarshalJSON_Encoded(t *testing.T) {
	fileData := `{"username":"test","encoded":true,"bytes":"base64:dGVzdA==","name":"test.txt"}`

	var file File
	err := json.Unmarshal([]byte(fileData), &file)

	assert.Nil(t, err)
	assert.Equal(t, "test", file.Username)
	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, []byte("test"), []byte(file.Bytes), "expected 'test' but got %q", string(file.Bytes))
}

func TestFile_UnmarshalJSON_Raw(t *testing.T) {
	fileData := `{"username":"test","bytes":"test","name":"test.txt"}`

	var file File
	err := json.Unmarshal([]byte(fileData), &file)

	assert.Nil(t, err)
	assert.Equal(t, "test", file.Username)
	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, []byte("test"), []byte(file.Bytes), "expected 'test' but got %q", string(file.Bytes))
}
