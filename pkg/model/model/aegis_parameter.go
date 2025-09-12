package model

import (
	"fmt"
	"strconv"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&AegisParameter{})
}

type AegisParameter struct {
	registry.BaseModel
	Name        string `json:"name" desc:"The name of the Aegis parameter" example:"timeout"`
	Description string `json:"description" desc:"A description of the parameter suitable for human or LLM use" example:"The timeout in seconds for the Aegis operation"`
	Default     string `json:"default,omitempty" desc:"A string representation of the default value of the parameter" example:"30"`
	Required    bool   `json:"required" desc:"Whether the parameter is required" example:"true"`
	Type        string `json:"type" desc:"The type of the parameter" example:"string"`
	// internal
	parser func(string) error
}

func (a AegisParameter) WithRequired() AegisParameter {
	a.Required = true
	return a
}

func (a AegisParameter) WithDefault(defaultVal string) AegisParameter {
	a.Default = defaultVal
	return a
}

func (a AegisParameter) WithParser(parser func(string) error) AegisParameter {
	a.parser = parser
	return a
}

func (a AegisParameter) Parse(s *string) error {
	if a.Required && s == nil {
		return fmt.Errorf("AegisParameter %s is required", a.Name)
	}

	val := a.Default
	if s != nil {
		val = *s
	}

	err := a.parser(val)
	if err != nil {
		return fmt.Errorf("AegisParameter %s encountered error parsing %s: %w", a.Name, val, err)
	}

	return nil
}

func NewAegisParameter[T any](name, description string, destination *T) AegisParameter {
	tipe, parser := getAegisTypeAndParser(destination)

	ap := AegisParameter{
		Name:        name,
		Description: description,
		Type:        tipe,
		parser:      parser,
	}

	return ap
}

func getAegisTypeAndParser(destination any) (string, func(string) error) {
	switch d := any(destination).(type) {
	case *string:
		return wrapAegis(d, parseAegisString, "string")
	case *int:
		return wrapAegis(d, parseAegisInt, "int")
	case *bool:
		return wrapAegis(d, parseAegisBool, "bool")
	}

	errorFunc := func(string) error {
		return fmt.Errorf("unknown/unsupported type: %T", destination)
	}

	return "unknown", errorFunc
}

func wrapAegis[T any](destination *T, parser func(string, *T) error, tipe string) (string, func(string) error) {
	f := func(s string) error {
		return parser(s, destination)
	}

	return tipe, f
}

func parseAegisString(s string, destination *string) error {
	*destination = s
	return nil
}

func parseAegisInt(s string, destination *int) error {
	i, err := strconv.Atoi(s)
	*destination = i
	return err
}

func parseAegisBool(s string, destination *bool) error {
	b, err := strconv.ParseBool(s)
	*destination = b
	return err
}

// GetDescription returns a description for the AegisParameter model.
func (a *AegisParameter) GetDescription() string {
	return "Describes a single parameter for an Aegis management capability."
}