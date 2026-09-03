package jsonz

import (
	"reflect"
)

func Copy[I any, O any](input I, output O) (O, error) {
	data, err := Format(input)
	if err != nil {
		return output, err
	}
	return Parse(data.Bytes(), output)
}

func Clone[T any](input T, output T) (T, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		var zero T
		return zero, nil
	}
	return Copy(input, output)
}
