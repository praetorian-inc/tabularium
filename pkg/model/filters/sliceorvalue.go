package filters

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type SliceOrValue[T any] []T

func (sov *SliceOrValue[T]) UnmarshalJSON(data []byte) error {
	slice := []T{}
	var sErr, vErr error
	if sErr = json.Unmarshal(data, &slice); sErr == nil {
		*sov = slice
		return nil
	}

	var v T
	if vErr = json.Unmarshal(data, &v); vErr == nil {
		*sov = []T{v}
		return nil
	}

	return fmt.Errorf("failed to unmarshal: failed to unmarshal into slice: %v, failed to unmarshal into value: %v", sErr, vErr)
}

func (sov SliceOrValue[T]) MarshalJSON() ([]byte, error) {
	if len(sov) == 0 {
		return json.Marshal(nil)
	}
	if len(sov) > 1 || reflect.TypeOf(sov[0]).Kind() == reflect.Slice {
		return json.Marshal([]T(sov))
	}
	return json.Marshal(sov[0])
}
