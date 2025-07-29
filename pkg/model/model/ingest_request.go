package model

import "encoding/json"

type IngestRequest struct {
	Type  string          `json:"type"`
	Model json.RawMessage `json:"model"`
	ResultContext
}
