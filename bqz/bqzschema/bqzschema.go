package bqzschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/commonz/dataz"
)

func fix(s bigquery.Schema) bigquery.Schema {
	if s == nil {
		return nil
	}
	out := make(bigquery.Schema, len(s))
	for i, f := range s {
		c := *f
		c.Schema = fix(c.Schema)
		out[i] = &c
	}
	slices.SortFunc(out, func(a, b *bigquery.FieldSchema) int { return strings.Compare(a.Name, b.Name) })
	return out
}

func Infer(v any) (bigquery.Schema, error) {
	if s, ok := v.(bigquery.Schema); ok {
		return slices.Clone(s), nil
	}
	s, err := inferStructSchema(v)
	if err != nil {
		return nil, fmt.Errorf("failed to infer schema for %T: %w", v, err)
	}
	return fix(s), nil
}

func inferStructSchema(v any) (bigquery.Schema, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil, fmt.Errorf("cannot infer schema for nil value")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %s", t.Kind())
	}
	return inferStruct(t)
}

func inferStruct(t reflect.Type) (bigquery.Schema, error) {
	var schema bigquery.Schema

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// Ignore unexported fields
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		tag := f.Tag.Get("bigquery")
		if tag == "-" {
			continue
		}

		var name string
		var options []string
		if tag != "" {
			parts := strings.Split(tag, ",")
			name = parts[0]
			options = parts[1:]
		}

		if name == "" {
			name = f.Name
		}

		// Handle embedded structs
		if f.Anonymous && f.Type.Kind() == reflect.Struct && tag == "" {
			embeddedSchema, err := inferStruct(f.Type)
			if err != nil {
				return nil, err
			}
			schema = append(schema, embeddedSchema...)
			continue
		}

		fieldSchema := &bigquery.FieldSchema{
			Name: name,
		}

		// Determine if field is repeated or nullable
		ft := f.Type
		isRepeated := false
		isPointer := false

		if ft.Kind() == reflect.Slice && ft != reflect.TypeOf([]byte{}) {
			isRepeated = true
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Ptr {
			isPointer = true
			ft = ft.Elem()
		}

		// Check for custom and native Null types
		isNullType := false
		var bqType bigquery.FieldType

		switch ft {
		case reflect.TypeOf(dataz.NullString{}), reflect.TypeOf(bigquery.NullString{}):
			bqType = bigquery.StringFieldType
			isNullType = true
		case reflect.TypeOf(dataz.NullInt64{}), reflect.TypeOf(bigquery.NullInt64{}):
			bqType = bigquery.IntegerFieldType
			isNullType = true
		case reflect.TypeOf(dataz.NullFloat64{}), reflect.TypeOf(bigquery.NullFloat64{}):
			bqType = bigquery.FloatFieldType
			isNullType = true
		case reflect.TypeOf(dataz.NullBool{}), reflect.TypeOf(bigquery.NullBool{}):
			bqType = bigquery.BooleanFieldType
			isNullType = true
		case reflect.TypeOf(dataz.NullTimestamp{}), reflect.TypeOf(bigquery.NullTimestamp{}):
			bqType = bigquery.TimestampFieldType
			isNullType = true
		case reflect.TypeOf(time.Time{}):
			bqType = bigquery.TimestampFieldType
		default:
			switch ft.Kind() {
			case reflect.String:
				bqType = bigquery.StringFieldType
			case reflect.Slice:
				if ft == reflect.TypeOf([]byte{}) {
					bqType = bigquery.BytesFieldType
				} else {
					return nil, fmt.Errorf("nested slices not supported for field %s", f.Name)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				bqType = bigquery.IntegerFieldType
			case reflect.Float32, reflect.Float64:
				bqType = bigquery.FloatFieldType
			case reflect.Bool:
				bqType = bigquery.BooleanFieldType
			case reflect.Struct:
				bqType = bigquery.RecordFieldType
				subSchema, err := inferStruct(ft)
				if err != nil {
					return nil, fmt.Errorf("failed to infer schema for struct field %s: %w", f.Name, err)
				}
				fieldSchema.Schema = subSchema
			default:
				return nil, fmt.Errorf("unsupported type %s for field %s", ft, f.Name)
			}
		}

		fieldSchema.Type = bqType
		fieldSchema.Repeated = isRepeated

		// Required logic:
		// Slices (repeated) are never required.
		// Null types or pointers are nullable (not required) by default.
		// Basic types are required by default.
		// This can be overridden by tags (nullable/required).
		if isRepeated {
			fieldSchema.Repeated = true
			fieldSchema.Required = false
		} else {
			isRequired := true
			if isPointer || isNullType {
				isRequired = false
			}

			// Override with tag options
			if slices.Contains(options, "nullable") {
				isRequired = false
			}
			if slices.Contains(options, "required") {
				isRequired = true
			}
			fieldSchema.Required = isRequired
		}

		schema = append(schema, fieldSchema)
	}

	return schema, nil
}

type jsonField struct {
	Name                   string       `json:"name"`
	Type                   string       `json:"type"`
	Mode                   string       `json:"mode"`
	DefaultValueExpression string       `json:"defaultValueExpression,omitempty"`
	Fields                 []*jsonField `json:"fields,omitempty"`
}

func toJSON(s bigquery.Schema) []*jsonField {
	out := make([]*jsonField, len(s))
	for i, f := range s {
		typ := string(f.Type)
		if f.Type == bigquery.FloatFieldType {
			typ = "FLOAT64"
		}
		mode := "NULLABLE"
		if f.Repeated {
			mode = "REPEATED"
		} else if f.Required {
			mode = "REQUIRED"
		}
		out[i] = &jsonField{Name: f.Name, Type: typ, Mode: mode, DefaultValueExpression: f.DefaultValueExpression, Fields: toJSON(f.Schema)}
	}
	return out
}

func NativeTableJSON(v any) ([]byte, error) {
	s, err := Infer(v)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema json: %w", err)
	}
	b, err := json.MarshalIndent(toJSON(fix(s)), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema json: %w", err)
	}
	return append(b, '\n'), nil
}
