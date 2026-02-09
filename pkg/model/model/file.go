package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type SmartBytes []byte

func (sb SmartBytes) MarshalJSON() ([]byte, error) {
	if len(sb) == 0 {
		return []byte(`""`), nil
	}

	if needsEncoding(sb) {
		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(sb)))
		base64.StdEncoding.Encode(encoded, sb)
		return json.Marshal("base64:" + string(encoded))
	}

	return json.Marshal(string(sb))
}

func (sb *SmartBytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if len(s) > 7 && s[:7] == "base64:" {
		decoded, err := base64.StdEncoding.DecodeString(s[7:])
		if err != nil {
			return err
		}
		*sb = decoded
		return nil
	}

	*sb = []byte(s)
	return nil
}

func needsEncoding(data []byte) bool {
	if !utf8.Valid(data) {
		return true
	}

	for _, c := range data {
		// Control characters (except tab, CR, LF which JSON can handle with escaping)
		if c < 32 && c != '\t' && c != '\n' && c != '\r' {
			return true
		}

		// DEL character
		if c == 127 {
			return true
		}
	}

	return false
}

type File struct {
	registry.BaseModel
	baseTableModel
	Username string `dynamodbav:"username" json:"username" desc:"Chariot username associated with the file." example:"user@example.com"`
	Key      string `dynamodbav:"key" json:"key" desc:"Unique key for the file record." example:"#file#proofs/scan.txt"`
	// Attributes
	Name         string     `dynamodbav:"name" json:"name" desc:"Name or path of the file." example:"proofs/scan.txt" slim:"File"`
	Updated      string     `dynamodbav:"updated" json:"updated" desc:"Timestamp when the file record was last updated (RFC3339)." example:"2023-10-27T11:00:00Z"`
	Bytes        SmartBytes `dynamodbav:"-" json:"bytes" desc:"File content. Must be a JSON-compatible string. Can be encoded with base64 if it begins with the prefix 'base64:'" example:"base64:dGVzdA==" slim:"File"`
	Overwrite    bool       `dynamodbav:"-" json:"-"`
	PlexTracPath string     `dynamodbav:"plextrac_path" json:"-" desc:"Key for the file in PlexTrac." example:"/api/v2/uploads/1234567890"`
}

func init() {
	registry.Registry.MustRegisterModel(&File{})
}

func (f *File) GetKey() string {
	return f.Key
}

func (f *File) Defaulted() {
	f.Updated = Now()
	f.Overwrite = true
}

func (f *File) GetHooks() []registry.Hook {
	return []registry.Hook{
		{
			Call: func() error {
				f.Key = fmt.Sprintf("#file#%s", f.Name)
				return nil
			},
		},
	}
}

func NewFile(name string) File {
	f := File{
		Name: name,
	}
	f.Defaulted()
	registry.CallHooks(&f)
	return f
}

// GetDescription returns a description for the File model.
func (f *File) GetDescription() string {
	return "Represents a file system entity, including its path and potential content hash."
}

func (f *File) HydratableFilepath() string {
	return f.Name
}

func (f *File) Hydrate(data []byte) error {
	f.Bytes = data
	return nil
}

func (f *File) HydratedFile() File {
	return *f
}

func (f *File) Dehydrate() Hydratable {
	return nil
}
