package filters

import (
	"encoding/json"
	"fmt"
)

type Filter struct {
	Field           string            `json:"field"`
	Operator        string            `json:"operator"`
	Value           SliceOrValue[any] `json:"value"`
	Not             bool              `json:"not"`
	ReverseOperands bool              `json:"reverseOperands"`
	Alias           string            `json:"alias,omitempty"`
	RelationshipFilter
}

type RelationshipFilter struct {
	RelationshipNodeLabel string `json:"relationshipNodeLabel,omitempty"`
	RelationshipDirection string `json:"relationshipDirection,omitempty"`
	RelationshipLabel     string `json:"relationshipLabel,omitempty"`
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

func (f *Filter) IsRelationshipFilter() bool {
	label, direction, relationship := f.relationshipFieldsPresent()
	return label && direction && relationship
}

func (f *Filter) relationshipFieldsPresent() (label, direction, relationship bool) {
	return f.RelationshipNodeLabel != "", f.RelationshipDirection != "", f.RelationshipLabel != ""
}

func (f *Filter) Validate() error {
	if f.Operator == OperatorOr || f.Operator == OperatorAnd {
		if f.IsRelationshipFilter() {
			return fmt.Errorf("relationship filter is not allowed with logical filters")
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

	label, direction, relationship := f.relationshipFieldsPresent()
	hasAnyRelationshipField := label || direction || relationship
	if hasAnyRelationshipField && !f.IsRelationshipFilter() {
		return fmt.Errorf("relationshipNodeLabel, relationshipDirection, relationshipLabel must be provided together")
	}

	if !hasAnyRelationshipField {
		return nil
	}

	if f.RelationshipDirection != "source" && f.RelationshipDirection != "target" {
		return fmt.Errorf("relationshipDirection must be one of: source, target")
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
	OperatorAnyStartsWith      = "ANY_STARTS_WITH"
)
