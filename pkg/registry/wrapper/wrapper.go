package wrapper

import (
	"encoding/json"
	"fmt"
	"github.com/praetorian-inc/tabularium/pkg/registry/model"
	"github.com/praetorian-inc/tabularium/pkg/registry/shared"
	"reflect"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Wrapper allows us to unmarshal into a Model interface based on the type registry
// Wrapper is generic, which allows us to use it for interfaces that implement Model
type Wrapper[T model.Model] struct {
	Model          T      `dynamodbav:"model" json:"model"`
	Type           string `dynamodbav:"type" json:"type"`
	SkipDefaulting bool   `dynamodbav:"-" json:"-"`
}

func (t Wrapper[T]) MarshalJSON() ([]byte, error) {
	type Alias Wrapper[T]
	alias := Alias(t)
	if alias.Type == "" && reflect.ValueOf(alias.Model).IsValid() {
		alias.Type = model.Name(t.Model)
	}
	return json.Marshal(alias)
}

func (t *Wrapper[T]) UnmarshalJSON(data []byte) error {
	props := map[string]any{}
	if err := json.Unmarshal(data, &props); err != nil {
		return err
	}

	tipe, err := t.getType(props)
	if tipe == "" && t.isEmpty(props) {
		return nil
	}

	if err != nil {
		return err
	}
	t.Type = tipe

	fromModel := func(model any) error {
		if m, ok := model.(map[string]any); ok {
			return t.fromProps(m)
		} else {
			return nil
		}
	}

	model, ok := props["model"]
	if ok {
		return fromModel(model)
	}
	model, ok = props["Model"]
	if ok {
		return fromModel(model)
	}
	return t.fromProps(props)
}

func (t Wrapper[T]) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	type Alias Wrapper[T]
	alias := Alias(t)
	if alias.Type == "" && reflect.ValueOf(alias.Model).IsValid() {
		alias.Type = model.Name(t.Model)
	}
	return attributevalue.Marshal(alias)
}

func (t *Wrapper[T]) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	if _, ok := av.(*types.AttributeValueMemberNULL); ok {
		return nil
	}

	m, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("model is not a map")
	}

	props := map[string]any{}
	err := attributevalue.Unmarshal(m, &props)
	if err != nil {
		return err
	}

	tipe, err := t.getType(props)
	if tipe == "" && t.isEmpty(props) {
		return nil
	}

	if err != nil {
		return err
	}
	t.Type = tipe

	fromModel := func(model any) error {
		if m, ok := model.(map[string]any); ok {
			return t.fromProps(m)
		} else {
			return nil
		}
	}

	model, ok := props["model"]
	if ok {
		return fromModel(model)
	}
	model, ok = props["Model"]
	if ok {
		return fromModel(model)
	}

	return t.fromProps(props)
}

func (t *Wrapper[T]) isEmpty(props map[string]any) bool {
	model, ok := props["model"]
	if ok && model == nil {
		return true
	}

	return len(props) == 0
}

func (t *Wrapper[T]) getType(props map[string]any) (string, error) {
	if t.Type != "" {
		return t.Type, nil
		// DynamoDB uses title case by default, and most of our models define dynamodb field tags
		// Furthermore, JSON uses lowercase, so we need to check both title-case and lowercase for any field that could be defined on the model
		// Fortunately, of these fields, only "key" is relevant here, so we must check both 'Key' and 'key'
	} else if k, ok := props["Key"].(string); ok {
		v := strings.Split(k, "#")
		if len(v) >= 2 && v[1] != "" {
			return v[1], nil
		}
	} else if k, ok := props["key"].(string); ok {
		v := strings.Split(k, "#")
		if len(v) >= 2 && v[1] != "" {
			return v[1], nil
		}
	} else if t, ok := props["type"].(string); ok && t != "" {
		return t, nil
	}

	if model, ok := props["model"]; ok {
		if m, ok := model.(map[string]any); ok {
			return t.getType(m)
		}
	}

	return "", fmt.Errorf("wrapper contains neither type nor key with type")
}

func (t *Wrapper[T]) fromProps(props map[string]any) error {
	tipe := t.Type
	tipes := model.GetTypes[T](shared.Registry)
	if !slices.Contains(tipes, strings.ToLower(tipe)) {
		return fmt.Errorf("provided type %q not known or does not implement %T", tipe, t.Model)
	}

	model, ok := shared.Registry.MakeType(tipe)
	if !ok {
		return fmt.Errorf("failed to make type %v", tipe)
	}

	if !t.SkipDefaulting {
		model.Defaulted()
	}

	t.Model, ok = model.(T)
	if !ok {
		return fmt.Errorf("failed to convert %v to %T", tipe, t.Model)
	}

	bytes, err := json.Marshal(props)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, t.Model)
	if err != nil {
		return err
	}

	return nil
}
