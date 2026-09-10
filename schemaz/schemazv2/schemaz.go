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
	// Documentation Metadata (ignored by JSON, used by your generator)
	Desc *Desc

	// Mutually exclusive data bindings (returning pointers for 2-way binding)
	Raw    func() any
	Object func() map[string]*Schema
	Array  func() (length int, getElement func(idx int) *Schema)
	Str    func(v string)
	Strs   func(v []string)
}

// =====================================
// MARSHAL (Writing to JSON)
// =====================================

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Object != nil {
		// json.Marshal automatically calls MarshalJSON on the *Schema values
		return json.Marshal(s.Object())
	}
	if s.Array != nil {
		length, getElem := s.Array()
		out := make([]*Schema, length)
		for i := 0; i < length; i++ {
			out[i] = getElem(i)
		}
		// Marshaling a slice of *Schema triggers recursive MarshalJSON
		return json.Marshal(out)
	}
	if s.Raw != nil {
		return json.Marshal(s.Raw())
	}
	// if s.Str != nil {
	// 	if ptr := s.Str(); ptr != nil {
	// 		return json.Marshal(*ptr)
	// 	}
	// }
	// if s.Strs != nil {
	// 	if ptr := s.Strs(); ptr != nil {
	// 		return json.Marshal(*ptr)
	// 	}
	// }
	return []byte("null"), nil
}

// =====================================
// UNMARSHAL (Reading from JSON)
// =====================================

func (s *Schema) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if s.Object != nil {
		return s.unmarshalJSONObject(data)
	}
	if s.Array != nil {
		return s.unmarshalJSONArray(data)
	}
	if s.Raw != nil {
		return json.Unmarshal(data, s.Raw())
	}
	if s.Str != nil {
		return s.unmarshalJSONStr(data)
	}
	if s.Strs != nil {
		return s.unmarshalJSONStrs(data)
	}
	return fmt.Errorf("schema must have one field set (Object, Array, Raw, Str, Strs)")
}

func (s *Schema) unmarshalJSONObject(data []byte) error {
	m := s.Object()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, schemaNode := range m {
		if rawVal, exists := raw[k]; exists {
			if err := json.Unmarshal(rawVal, schemaNode); err != nil {
				return fmt.Errorf("error parsing key %q: %w", k, err)
			}
		}
	}
	return nil
}

func (s *Schema) unmarshalJSONArray(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	_, getElem := s.Array() // We only need the element getter for unmarshaling
	for i, rawVal := range raw {
		schemaNode := getElem(i)
		if schemaNode == nil {
			continue
		}
		if err := json.Unmarshal(rawVal, schemaNode); err != nil {
			return fmt.Errorf("error parsing index %d: %w", i, err)
		}
	}
	return nil
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
