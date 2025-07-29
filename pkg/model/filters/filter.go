package filters

import (
	"encoding/json"
)

type Filter struct {
	Field           string            `json:"field"`
	Operator        string            `json:"operator"`
	Value           SliceOrValue[any] `json:"value"`
	Not             bool              `json:"not"`
	ReverseOperands bool              `json:"reverseOperands"`
	Alias           string            `json:"alias,omitempty"`
}

func NewFilter(field, operator string, value any, opts ...Option) Filter {
	f := Filter{
		Field:    field,
		Operator: operator,
		Value:    SliceOrValue[any]{},
	}
	switch i := value.(type) {
	case []Filter:
		for _, filter := range i {
			f.Value = append(f.Value, filter)
		}
	default:
		f.Value = SliceOrValue[any]{value}
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

type Option func(*Filter)

func WithNot() Option {
	return func(f *Filter) {
		f.Not = true
	}
}

func WithReverseOperands() Option {
	return func(f *Filter) {
		f.ReverseOperands = true
	}
}

func (f *Filter) UnmarshalJSON(data []byte) error {

	type filterT Filter // trick to use regular unmarshalJSON logic and then post process with

	var filter filterT
	if err := json.Unmarshal(data, &filter); err != nil {
		return err
	}

	if filter.Operator == OperatorOr || filter.Operator == OperatorAnd {
		// if we have operator OR or AND, do some extra work to get a concrete type
		b, err := json.Marshal(filter.Value)
		if err != nil {
			return err
		}

		concrete := SliceOrValue[Filter]{}
		if err := json.Unmarshal(b, &concrete); err != nil {
			return err
		}

		filter.Value = SliceOrValue[any]{}
		for _, c := range concrete {
			filter.Value = append(filter.Value, c)
		}
	}

	*f = Filter(filter)
	return nil
}

const (
	OperatorEqual              = "="
	OperatorContains           = "CONTAINS"
	OperatorLessThan           = "<"
	OperatorLessThanEqualTo    = "<="
	OperatorGreaterThan        = ">"
	OperatorGreaterThanEqualTo = ">="
	OperatorStartsWith         = "STARTS WITH"
	OperatorEndsWith           = "ENDS WITH"
	OperatorOr                 = "OR"
	OperatorAnd                = "AND"
	OperatorIn                 = "IN"
	OperatorIsNotNull          = "IS NOT NULL"
)
