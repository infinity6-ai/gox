package schemazv2

import (
	"encoding/json"
	"fmt"
	"slices"
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
	Object func(read bool) map[string]*Schema
	Array  func() (length int, getElement func(idx int, read bool) *Schema)
	Str    func() (func(v string), func() string)
	Strs   func() (func(v []string), func() []string)
}

// =====================================
// MARSHAL (Writing to JSON)
// =====================================

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Object != nil {
		// json.Marshal automatically calls MarshalJSON on the *Schema values
		return json.Marshal(s.Object(true))
	}
	if s.Array != nil {
		length, getElem := s.Array()
		out := make([]*Schema, length)
		for i := 0; i < length; i++ {
			out[i] = getElem(i, true)
		}
		// Marshaling a slice of *Schema triggers recursive MarshalJSON
		return json.Marshal(out)
	}
	if s.Raw != nil {
		return json.Marshal(s.Raw())
	}
	if s.Str != nil {
		_, f := s.Str()
		return json.Marshal([]string{f()})
	}
	if s.Strs != nil {
		_, f := s.Strs()
		return json.Marshal(f())
	}
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
	m := s.Object(false)
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
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		return err
	}

	_, getElem := s.Array()

	// We iterate backwards to give the schema implementation an opportunity to
	// grow its underlying slice to the correct size from the beginning. The first
	// index seen will be the largest (len-1), allowing for a single allocation
	// to the final size.
	for i, raw := range slices.Backward(raws) {
		schemaNode := getElem(i, false)
		if schemaNode == nil {
			continue
		}
		if err := json.Unmarshal(raw, schemaNode); err != nil {
			return fmt.Errorf("error parsing index %d: %w", i, err)
		}
	}
	return nil
}

func (s *Schema) unmarshalJSONStr(data []byte) error {
	if len(data) == 0 {
		panic("data must not be empty")
	}
	parser, _ := s.Str()
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	if len(strs) == 0 {
		parser("")
	} else {
		parser(strs[0])
	}
	return nil
}

func (s *Schema) unmarshalJSONStrs(data []byte) error {
	if len(data) == 0 {
		panic("data must not be empty")
	}
	parser, _ := s.Strs()
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	parser(strs)
	return nil
}
