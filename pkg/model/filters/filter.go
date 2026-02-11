package filters

import (
	"encoding/json"
	"fmt"
)

type Filter struct {
	Field                string            `json:"field"`
	Operator             string            `json:"operator"`
	Value                SliceOrValue[any] `json:"value"`
	Not                  bool              `json:"not"`
	ReverseOperands      bool              `json:"reverseOperands"`
	Alias                string            `json:"alias,omitempty"`
	MetadataLabel        string            `json:"metadataLabel,omitempty"`
	MetadataDirection    string            `json:"metadataDirection,omitempty"`
	MetadataRelationship string            `json:"metadataRelationship,omitempty"`
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

	for i := range f.Value {
		num, ok := f.Value[i].(float64)
		if !ok {
			continue
		}

		if num == float64(int64(num)) {
			f.Value[i] = int64(num)
		}
	}

	return nil
}

func (f *Filter) HasMetadata() bool {
	label, direction, relationship := f.metadataFieldsPresent()
	return label && direction && relationship
}

func (f *Filter) metadataFieldsPresent() (label, direction, relationship bool) {
	return f.MetadataLabel != "", f.MetadataDirection != "", f.MetadataRelationship != ""
}

func (f *Filter) Validate() error {
	if f.Operator == OperatorOr || f.Operator == OperatorAnd {
		if f.HasMetadata() {
			return fmt.Errorf("metadata is not allowed on logical filters")
		}
		for _, value := range f.Value {
			nested, ok := value.(Filter)
			if !ok {
				continue
			}
			if err := nested.Validate(); err != nil {
				return err
			}
		}
		return nil
	}

	label, direction, relationship := f.metadataFieldsPresent()
	hasAnyMetadata := label || direction || relationship
	if hasAnyMetadata && !f.HasMetadata() {
		return fmt.Errorf("metadataLabel, metadataDirection, metadataRelationship must be provided together")
	}
	if !hasAnyMetadata {
		return nil
	}

	if f.MetadataDirection != "source" && f.MetadataDirection != "target" {
		return fmt.Errorf("metadataDirection must be one of: source, target")
	}

	return nil
}

const (
	OperatorEqual              = "="
	OperatorContains           = "CONTAINS"
	OperatorLike               = "LIKE"
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
	OperatorIsNull             = "IS NULL"
)
