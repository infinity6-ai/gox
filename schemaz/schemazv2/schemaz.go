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
	Array  func(idx int) *Schema
	Strs   func(v []string)
	Str    func(v string)
	Desc   func() *Desc
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Raw != nil {
		return json.Marshal(s.Raw())
	}
	if s.Object != nil {
		return json.Marshal(s.Object())
	}
	// if s.Array != nil {
	// 	return json.Marshal([]*Schema{s.Array()})
	// }
	// panic("IMPLEMENT IT")
	return []byte("null"), nil
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
	if s.Strs != nil {
		return s.unmarshalJSONStrs(data)
	}
	if s.Str != nil {
		return s.unmarshalJSONStr(data)
	}
	panic("schema must have either object or array or raw field set")
}

func (s *Schema) unmarshalJSONStr(data []byte) error {
	if len(data) == 0 {
		panic("data must not be empty")
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		s.Str(str)
		return nil
	}
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	if len(strs) == 0 {
		s.Str("")
	} else {
		s.Str(strs[0])
	}
	return nil
}

func (s *Schema) unmarshalJSONStrs(data []byte) error {
	if len(data) == 0 {
		panic("data must not be empty")
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		s.Strs([]string{str})
		return nil
	}
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	s.Strs(strs)
	return nil
}

func (s *Schema) unmarshalJSONArray(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for i, ble := range raw {
		m := s.Array(i)
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
