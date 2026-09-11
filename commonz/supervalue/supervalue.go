package supervalue

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// 1. The interface that enforces validation
type checker[V any] interface {
	Check(v V)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type SuperValue[T comparable] struct {
	value   T
	present bool
}

// 2. Change 'x any' to 'x checker[V]'.
// Now, the compiler enforces the interface instead of a runtime type assertion!
func Set[V comparable](x checker[V], v V) {
	// Guaranteed to be safe, no type assertion needed
	x.Check(v)

	val := reflect.ValueOf(x)

	if val.Kind() != reflect.Ptr {
		panic("Set requires a pointer to a struct")
	}

	targetType := reflect.TypeOf((*SuperValue[V])(nil))

	if val.Type().ConvertibleTo(targetType) {
		convertedPtr := val.Convert(targetType).Interface().(*SuperValue[V])
		if convertedPtr.present {
			panic("x has already been set")
		}
		convertedPtr.value = v
		convertedPtr.present = true
	} else {
		panic("x is not derived from SuperValue[V]")
	}
}

// Marshal extracts the hidden 'v' and marshals it.
func Marshal[V comparable](x any) ([]byte, error) {
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	targetType := reflect.TypeOf(SuperValue[V]{})
	if val.Type().ConvertibleTo(targetType) {
		// Convert it back to SuperValue so we can read the unexported 'v'
		base := val.Convert(targetType).Interface().(SuperValue[V])
		return json.Marshal(base.value)
	}
	return nil, fmt.Errorf("type is not convertible to SuperValue")
}

// Unmarshal decodes the JSON, runs your Check(), and safely assigns the value.
func Unmarshal[V comparable](x checker[V], data []byte) (err error) {
	var temp V
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Catch any panics thrown by the user's Check() method
	// and convert them into standard JSON unmarshaling errors.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("validation failed: %v", r)
		}
	}()

	// Set() will automatically call x.Check(temp)
	Set(x, temp)
	return nil
}
