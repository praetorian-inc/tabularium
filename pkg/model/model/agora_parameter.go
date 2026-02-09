package model

import (
	"fmt"
	"strconv"

	"github.com/praetorian-inc/tabularium/pkg/registry"
)

func init() {
	registry.Registry.MustRegisterModel(&AgoraParameter{})
}

type AgoraParameter struct {
	registry.BaseModel
	Name        string `json:"name" desc:"The name of the parameter" example:"rate_limit"`
	Description string `json:"description" desc:"A description of the parameter suitable for human or LLM use" example:"The rate limit for the capability"`
	Default     string `json:"default,omitempty" desc:"A string representation of the default value of the parameter" example:"1.2.3.4"`
	Required    bool   `json:"required" desc:"Whether the parameter is required" example:"true"`
	Type        string   `json:"type" desc:"The type of the parameter" example:"string"`
	Options     []string `json:"options,omitempty" desc:"Suggested choices for enum-style parameters, used as UI hints" example:"[\"option_a\",\"option_b\"]"`
	// internal
	parser func(string) error
}

func (a AgoraParameter) WithRequired() AgoraParameter {
	a.Required = true
	return a
}

func (a AgoraParameter) WithDefault(defaultVal string) AgoraParameter {
	a.Default = defaultVal
	return a
}

func (a AgoraParameter) WithOptions(opts []string) AgoraParameter {
	a.Options = opts
	return a
}

func (a AgoraParameter) WithParser(parser func(string) error) AgoraParameter {
	a.parser = parser
	return a
}

func (a AgoraParameter) Parse(s *string) error {
	if a.Required && s == nil {
		return fmt.Errorf("AgoraParameter %s is required", a.Name)
	}

	val := a.Default
	if s != nil {
		val = *s
	}

	err := a.parser(val)
	if err != nil {
		return fmt.Errorf("AgoraParameter %s encountered error parsing %s: %w", a.Name, val, err)
	}

	return nil
}

func NewAgoraParameter[T any](name, description string, destination *T) AgoraParameter {
	tipe, parser := getTypeAndParser(destination)

	ap := AgoraParameter{
		Name:        name,
		Description: description,
		Type:        tipe,
		parser:      parser,
	}

	return ap
}

func getTypeAndParser(destination any) (string, func(string) error) {
	switch d := any(destination).(type) {
	case *string:
		return wrap(d, parseString, "string")
	case *int:
		return wrap(d, parseInt, "int")
	case *bool:
		return wrap(d, parseBool, "bool")
	}

	errorFunc := func(string) error {
		return fmt.Errorf("unknown/unsupported type: %T", destination)
	}

	return "unknown", errorFunc
}

func wrap[T any](destination *T, parser func(string, *T) error, tipe string) (string, func(string) error) {
	f := func(s string) error {
		return parser(s, destination)
	}

	return tipe, f
}

func parseString(s string, destination *string) error {
	*destination = s
	return nil
}

func parseInt(s string, destination *int) error {
	i, err := strconv.Atoi(s)
	*destination = i
	return err
}

func parseBool(s string, destination *bool) error {
	b, err := strconv.ParseBool(s)
	*destination = b
	return err
}

// GetDescription returns a description for the AgoraParameter model.
func (a *AgoraParameter) GetDescription() string {
	return "Describes a single parameter for an Agora capability."
}
