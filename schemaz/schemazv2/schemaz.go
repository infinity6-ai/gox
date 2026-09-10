package schemazv2

import (
	"encoding/json"
	"fmt"
)

type Desc struct {
	Name     string
	Summary  string
	Markdown string
}

type Schema struct {
	Raw    func() any
	Object func() map[string]*Schema
	Array  func() *Schema
	Values func(unformatted []string)
	Desc   func() *Desc
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	if s.Object != nil {
		return s.unmarshalJSONObject(data)
	}
	if s.Array != nil {
		return s.unmarshalJSONArray(data)
	}
	if s.Raw != nil {
		return json.Unmarshal(data, s.Raw())
	}
	if s.Values != nil {
		return s.unmarshalJSONValues(data)
	}
	panic("schema must have either object or array or raw field set")
}

func (s *Schema) unmarshalJSONValues(data []byte) error {
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Values([]string{string(raw)})
	return nil
}

func (s *Schema) unmarshalJSONArray(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for i, ble := range raw {
		m := s.Array()
		if err := json.Unmarshal(ble, m); err != nil {
			return fmt.Errorf("error parsing key %d: %w", i, err)
		}
	}
	return nil
}

func (s *Schema) unmarshalJSONObject(data []byte) error {
	m := s.Object()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range m {
		if rawVal, exists := raw[k]; exists {
			if err := json.Unmarshal(rawVal, &v); err != nil {
				return fmt.Errorf("error parsing key %q: %w", k, err)
			}
		}
	}
	return nil
}
