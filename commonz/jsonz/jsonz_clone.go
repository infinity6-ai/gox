package jsonz

import (
	"reflect"
)

// Copy marshals the input and unmarshals it into the output.
func Copy[I any, O any](input I, output O) (O, error) {
	data, err := Format(input)
	if err != nil {
		return output, err
	}
	return Parse(data.Bytes(), output)
}

// Clone deeply copies the input to the output of the same type.
// If the input is a nil pointer, it returns the zero value of the type.
func Clone[T any](input T, output T) (T, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		var zero T
		return zero, nil
	}
	return Copy(input, output)
}
