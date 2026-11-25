package bqzimpl

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/commonz/dataz"
)

// LoadRow maps a slice of bigquery.Value to a destination struct pointer v
// based on the provided bigquery.Schema.
func LoadRow(v any, schema bigquery.Schema, values []bigquery.Value) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer, got %T", v)
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to a struct, got %T", v)
	}

	return loadStruct(elem, schema, values)
}

func loadStruct(val reflect.Value, schema bigquery.Schema, values []bigquery.Value) error {
	if len(schema) != len(values) {
		return fmt.Errorf("schema length (%d) and values length (%d) mismatch", len(schema), len(values))
	}

	m := buildFieldIndexMap(val.Type())

	for i, fs := range schema {
		rawVal := values[i]
		fieldName := strings.ToLower(fs.Name)
		path, found := m[fieldName]
		if !found {
			// Field not found in struct, ignore it (matches standard Go behavior)
			continue
		}

		fv := val.FieldByIndex(path)
		if err := setField(fv, fs, rawVal); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fs.Name, err)
		}
	}

	return nil
}

func buildFieldIndexMap(t reflect.Type) map[string][]int {
	m := make(map[string][]int)
	buildFieldIndexMapRecursive(t, nil, m)
	return m
}

func buildFieldIndexMapRecursive(t reflect.Type, path []int, m map[string][]int) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		tag := f.Tag.Get("bigquery")
		if tag == "-" {
			continue
		}

		currentPath := append(append([]int(nil), path...), i)

		// Handle embedded structs
		if f.Anonymous && f.Type.Kind() == reflect.Struct && tag == "" {
			buildFieldIndexMapRecursive(f.Type, currentPath, m)
			continue
		}

		var name string
		if tag != "" {
			parts := strings.Split(tag, ",")
			name = parts[0]
		}
		if name == "" {
			name = f.Name
		}
		m[strings.ToLower(name)] = currentPath
	}
}

func setField(fv reflect.Value, fs *bigquery.FieldSchema, val any) error {
	// If the value is nil, we set the field to its zero value or null representation
	if val == nil {
		return setNil(fv)
	}

	// Dereference pointers if needed
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return setField(fv.Elem(), fs, val)
	}

	// Handle our custom and bigquery Null types first
	switch fv.Type() {
	case reflect.TypeOf(dataz.NullString{}):
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string for NullString, got %T", val)
		}
		fv.Set(reflect.ValueOf(dataz.NullString{StringVal: str, Valid: true}))
		return nil
	case reflect.TypeOf(bigquery.NullString{}):
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string for bigquery.NullString, got %T", val)
		}
		fv.Set(reflect.ValueOf(bigquery.NullString{StringVal: str, Valid: true}))
		return nil

	case reflect.TypeOf(dataz.NullInt64{}):
		i, err := toInt64(val)
		if err != nil {
			return fmt.Errorf("failed to parse int64 for NullInt64: %w", err)
		}
		fv.Set(reflect.ValueOf(dataz.NullInt64{Int64: i, Valid: true}))
		return nil
	case reflect.TypeOf(bigquery.NullInt64{}):
		i, err := toInt64(val)
		if err != nil {
			return fmt.Errorf("failed to parse int64 for bigquery.NullInt64: %w", err)
		}
		fv.Set(reflect.ValueOf(bigquery.NullInt64{Int64: i, Valid: true}))
		return nil

	case reflect.TypeOf(dataz.NullFloat64{}):
		f, ok := val.(float64)
		if !ok {
			return fmt.Errorf("expected float64 for NullFloat64, got %T", val)
		}
		fv.Set(reflect.ValueOf(dataz.NullFloat64{Float64: f, Valid: true}))
		return nil
	case reflect.TypeOf(bigquery.NullFloat64{}):
		f, ok := val.(float64)
		if !ok {
			return fmt.Errorf("expected float64 for bigquery.NullFloat64, got %T", val)
		}
		fv.Set(reflect.ValueOf(bigquery.NullFloat64{Float64: f, Valid: true}))
		return nil

	case reflect.TypeOf(dataz.NullBool{}):
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("expected bool for NullBool, got %T", val)
		}
		fv.Set(reflect.ValueOf(dataz.NullBool{Bool: b, Valid: true}))
		return nil
	case reflect.TypeOf(bigquery.NullBool{}):
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("expected bool for bigquery.NullBool, got %T", val)
		}
		fv.Set(reflect.ValueOf(bigquery.NullBool{Bool: b, Valid: true}))
		return nil

	case reflect.TypeOf(dataz.NullTimestamp{}):
		t, ok := val.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time for NullTimestamp, got %T", val)
		}
		fv.Set(reflect.ValueOf(dataz.NullTimestamp{Timestamp: t, Valid: true}))
		return nil
	case reflect.TypeOf(bigquery.NullTimestamp{}):
		t, ok := val.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time for bigquery.NullTimestamp, got %T", val)
		}
		fv.Set(reflect.ValueOf(bigquery.NullTimestamp{Timestamp: t, Valid: true}))
		return nil
	}

	// Handle slices (repeated fields)
	if fv.Kind() == reflect.Slice && fv.Type() != reflect.TypeOf([]byte{}) {
		valList, ok := val.([]bigquery.Value)
		if !ok {
			// Fallback: in some testing situations it might be a slice of any
			anyList, okAny := val.([]any)
			if !okAny {
				return fmt.Errorf("expected slice for repeated field, got %T", val)
			}
			valList = make([]bigquery.Value, len(anyList))
			for idx, anyVal := range anyList {
				valList[idx] = anyVal
			}
		}
		sliceVal := reflect.MakeSlice(fv.Type(), len(valList), len(valList))
		elemSchema := &bigquery.FieldSchema{
			Type:   fs.Type,
			Schema: fs.Schema,
		}
		for j := 0; j < len(valList); j++ {
			if err := setField(sliceVal.Index(j), elemSchema, valList[j]); err != nil {
				return err
			}
		}
		fv.Set(sliceVal)
		return nil
	}

	// Handle nested structs (records)
	if fv.Kind() == reflect.Struct && fv.Type() != reflect.TypeOf(time.Time{}) {
		valList, ok := val.([]bigquery.Value)
		if !ok {
			// Fallback: in some testing situations it might be a slice of any
			anyList, okAny := val.([]any)
			if !okAny {
				return fmt.Errorf("expected record values slice for struct, got %T", val)
			}
			valList = make([]bigquery.Value, len(anyList))
			for idx, anyVal := range anyList {
				valList[idx] = anyVal
			}
		}
		return loadStruct(fv, fs.Schema, valList)
	}

	// Handle primitive types
	switch fv.Kind() {
	case reflect.String:
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
		fv.SetString(s)
	case reflect.Bool:
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", val)
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := toInt64(val)
		if err != nil {
			return err
		}
		fv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := toInt64(val)
		if err != nil {
			return err
		}
		fv.SetUint(uint64(i))
	case reflect.Float32, reflect.Float64:
		f, ok := val.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", val)
		}
		fv.SetFloat(f)
	default:
		// e.g. time.Time
		fv.Set(reflect.ValueOf(val))
	}

	return nil
}

func toInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", val)
	}
}

func setNil(fv reflect.Value) error {
	if fv.Kind() == reflect.Ptr {
		fv.Set(reflect.Zero(fv.Type()))
		return nil
	}

	// Set null representations
	switch fv.Type() {
	case reflect.TypeOf(dataz.NullString{}):
		fv.Set(reflect.ValueOf(dataz.NullString{Valid: false}))
	case reflect.TypeOf(bigquery.NullString{}):
		fv.Set(reflect.ValueOf(bigquery.NullString{Valid: false}))
	case reflect.TypeOf(dataz.NullInt64{}):
		fv.Set(reflect.ValueOf(dataz.NullInt64{Valid: false}))
	case reflect.TypeOf(bigquery.NullInt64{}):
		fv.Set(reflect.ValueOf(bigquery.NullInt64{Valid: false}))
	case reflect.TypeOf(dataz.NullFloat64{}):
		fv.Set(reflect.ValueOf(dataz.NullFloat64{Valid: false}))
	case reflect.TypeOf(bigquery.NullFloat64{}):
		fv.Set(reflect.ValueOf(bigquery.NullFloat64{Valid: false}))
	case reflect.TypeOf(dataz.NullBool{}):
		fv.Set(reflect.ValueOf(dataz.NullBool{Valid: false}))
	case reflect.TypeOf(bigquery.NullBool{}):
		fv.Set(reflect.ValueOf(bigquery.NullBool{Valid: false}))
	case reflect.TypeOf(dataz.NullTimestamp{}):
		fv.Set(reflect.ValueOf(dataz.NullTimestamp{Valid: false}))
	case reflect.TypeOf(bigquery.NullTimestamp{}):
		fv.Set(reflect.ValueOf(bigquery.NullTimestamp{Valid: false}))
	default:
		fv.Set(reflect.Zero(fv.Type()))
	}
	return nil
}
