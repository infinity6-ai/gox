package gobz

import "reflect"

// Copy encodes the input to gob and decodes it into the output.
// This is useful for deep copying.
// It returns the populated output object.
func Copy[I any, O any](input I, output O) (O, error) {
	data, err := Format(input)
	if err != nil {
		return output, err
	}
	return Parse(data, output)
}

// Clone creates a deep copy of the input object using gob encoding and decoding.
// It returns a new instance of the same type as the input.
func Clone[T any](input T, output T) (T, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		var zero T
		return zero, nil
	}
	return Copy(input, output)
}
