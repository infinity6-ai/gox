package bqzschema

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"cloud.google.com/go/bigquery"
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
	s, err := bigquery.InferSchema(v)
	if err != nil {
		return nil, fmt.Errorf("failed to infer schema for %T: %w", v, err)
	}
	return fix(s), nil
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
	if !slices.ContainsFunc(s, func(f *bigquery.FieldSchema) bool { return f.Name == "inserted_at" }) {
		s = append(s, &bigquery.FieldSchema{Name: "inserted_at", Type: bigquery.TimestampFieldType, DefaultValueExpression: "CURRENT_TIMESTAMP()"})
	}
	b, err := json.MarshalIndent(toJSON(fix(s)), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema json: %w", err)
	}
	return append(b, '\n'), nil
}
