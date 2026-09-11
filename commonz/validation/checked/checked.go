package checked

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// 1. The interface that enforces validation
type Checker[V comparable] interface {
	Get() V
	String() string
	Validate(v V) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type Value[T comparable] struct {
	value   T
	present bool
}

func (v Value[T]) Get() T {
	if !v.present {
		panic("x has not been set")
	}
	return v.value
}

// 2. Change 'x any' to 'x checker[V]'.
// Now, the compiler enforces the interface instead of a runtime type assertion!
func Set[V comparable](x Checker[V], v V) error {
	// Guaranteed to be safe, no type assertion needed
	err := x.Validate(v)
	if err != nil {
		return fmt.Errorf("supervalue validation error: %s", err)
	}

	val := reflect.ValueOf(x)

	if val.Kind() != reflect.Ptr {
		panic("Set requires a pointer to a struct")
	}

	targetType := reflect.TypeOf((*Value[V])(nil))

	if val.Type().ConvertibleTo(targetType) {
		convertedPtr := val.Convert(targetType).Interface().(*Value[V])
		if convertedPtr.present {
			panic("x has already been set")
		}
		convertedPtr.value = v
		convertedPtr.present = true
	} else {
		panic("x is not derived from SuperValue[V]")
	}
	return nil
}

// Marshal extracts the hidden 'v' and marshals it.
func Marshal[V comparable](x any) ([]byte, error) {
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	targetType := reflect.TypeOf(Value[V]{})
	if val.Type().ConvertibleTo(targetType) {
		// Convert it back to SuperValue so we can read the unexported 'v'
		base := val.Convert(targetType).Interface().(Value[V])
		return json.Marshal(base.value)
	}
	return nil, fmt.Errorf("type is not convertible to SuperValue")
}

// Unmarshal decodes the JSON, runs your Check(), and safely assigns the value.
func Unmarshal[V comparable](x Checker[V], data []byte) (err error) {
	var temp V
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	if err := Set(x, temp); err != nil {
		return err
	}
	return nil
}
