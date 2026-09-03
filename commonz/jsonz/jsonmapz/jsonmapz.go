package jsonmapz

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/jsonz"
)

// ParseMap parses a map of string slices into a map of structured data.
// It uses the `out` map as a template for types.
func ParseMap(in map[string][]string, out map[string]any) error {
	for key, values := range in {
		target, ok := out[key]
		if !ok || len(values) == 0 {
			continue
		}

		// Handle string and []string as special cases that don't need JSON parsing.
		switch target.(type) {
		case string:
			out[key] = values[0]
			continue
		case []string:
			out[key] = values
			continue
		}

		targetType := reflect.TypeOf(target)

		if targetType.Kind() != reflect.Slice {
			// Not a slice: parse the first value.
			container := reflect.New(targetType).Interface()

			parsed, err := jsonz.Parse(values[0], container)
			if err != nil {
				return fmt.Errorf("failed to parse value for key '%s': %w", key, err)
			}
			out[key] = reflect.ValueOf(parsed).Elem().Interface()
		} else {
			// A slice of elements that need parsing.
			sliceType := targetType
			elemType := sliceType.Elem()
			newSlice := reflect.MakeSlice(sliceType, 0, len(values))

			for _, strValue := range values {
				// Create a new element (as a pointer) to parse into.
				container := reflect.New(elemType).Interface()

				parsed, err := jsonz.Parse(strValue, container)
				if err != nil {
					return fmt.Errorf("failed to parse slice element for key '%s': %w", key, err)
				}

				// Append the parsed element (dereferenced) to the new slice.
				newSlice = reflect.Append(newSlice, reflect.ValueOf(parsed).Elem())
			}
			out[key] = newSlice.Interface()
		}
	}
	return nil
}

// FormatMap converts a map of structured data into a map of string slices.
func FormatMap(in map[string]any, out map[string][]string) error {
	for key, value := range in {
		if value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			out[key] = []string{v}
		case []string:
			out[key] = v
		default:
			val := reflect.ValueOf(value)
			if val.Kind() == reflect.Slice {
				// It's a slice of something other than string.
				var stringSlice []string
				for i := 0; i < val.Len(); i++ {
					b, err := json.Marshal(val.Index(i).Interface())
					if err != nil {
						return fmt.Errorf("failed to marshal slice element for key '%s': %w", key, err)
					}
					stringSlice = append(stringSlice, string(b))
				}
				out[key] = stringSlice
			} else {
				// It's a single item.
				b, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to marshal value for key '%s': %w", key, err)
				}
				out[key] = []string{string(b)}
			}
		}
	}
	return nil
}
