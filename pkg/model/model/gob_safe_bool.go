package model

import (
	"bytes"
	"encoding/gob"
)

func init() {
	b := GobSafeBool(true)
	gob.Register(b)
}

// GobSafeBool is a wrapper around bool that allows structs to embed a
// *bool field without gob losing zero values.
type GobSafeBool bool

func (gsb GobSafeBool) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	var alias = bool(gsb)
	if err := encoder.Encode(alias); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (gsb *GobSafeBool) GobDecode(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var alias = bool(*gsb)
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&alias); err != nil {
		return err
	}

	*gsb = GobSafeBool(alias)

	return nil
}
