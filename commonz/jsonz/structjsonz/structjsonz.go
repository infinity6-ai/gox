package structjsonz

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/infinity6-ai/gox/commonz/jsonz/jsonmapz"
)

func ParseSingle(in map[string]string, out any) error {
	n := make(map[string][]string, len(in))
	for k, v := range in {
		n[k] = []string{v}
	}
	return Parse(n, out)
}

// Parse populates the fields of a struct `out` from a map of string slices `in`.
// `out` must be a pointer to a struct.
// The mapping between `in` keys and struct fields is determined by the struct's `json` tags.
func Parse(in map[string][]string, out any) error {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("out must be a pointer to a struct")
	}

	structMap, err := structToMap(out)
	if err != nil {
		return fmt.Errorf("failed to convert struct to map: %w", err)
	}

	if err := jsonmapz.ParseMap(in, structMap); err != nil {
		return fmt.Errorf("error parsing map: %w", err)
	}

	if err := mapToStruct(structMap, out); err != nil {
		return fmt.Errorf("failed to convert map back to struct: %w", err)
	}

	return nil
}

// Format converts a struct `in` to a map of string slices.
// `in` must be a pointer to a struct.
// The mapping is determined by the struct's `json` tags.
func Format(in any) (map[string][]string, error) {
	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("in must be a pointer to a struct")
	}

	structMap, err := structToMap(in)
	if err != nil {
		return nil, fmt.Errorf("failed to convert struct to map: %w", err)
	}

	out := make(map[string][]string)
	if err := jsonmapz.FormatMap(structMap, out); err != nil {
		return nil, fmt.Errorf("error formatting map: %w", err)
	}

	return out, nil
}

func structToMap(data any) (map[string]any, error) {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()
	outMap := make(map[string]any)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		jsonTag := typeField.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		parts := strings.Split(jsonTag, ",")
		keyName := parts[0]

		containsOmitempty := false
		for _, part := range parts[1:] {
			if strings.TrimSpace(part) == "omitempty" {
				containsOmitempty = true
				break
			}
		}

		if containsOmitempty && field.IsZero() {
			continue
		}

		if keyName == "" {
			keyName = typeField.Name
		}

		outMap[keyName] = field.Interface()
	}
	return outMap, nil
}

func mapToStruct(m map[string]any, structPtr any) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, structPtr); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into struct: %w", err)
	}

	return nil
}
