package camelz

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// MarshalJSON marshals the Parsed identifier into JSON as a snake_case string.
func (p *Parsed) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte("null"), nil
	}
	return json.Marshal(p.SL())
}

// UnmarshalJSON unmarshals a JSON snake_case (or any supported casing) string into Parsed.
func (p *Parsed) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		p.parts = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal camelz from json: %w", err)
	}
	parsed, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse camelz from json: %w", err)
	}
	p.parts = parsed.parts
	return nil
}

// GobEncode encodes the Parsed identifier as a gob-encoded snake_case string.
func (p *Parsed) GobEncode() ([]byte, error) {
	var s string
	if p != nil {
		s = p.SL()
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s); err != nil {
		return nil, fmt.Errorf("cannot marshal camelz to gob: %w", err)
	}
	return buf.Bytes(), nil
}

// GobDecode decodes a gob-encoded string into Parsed.
func (p *Parsed) GobDecode(data []byte) error {
	var s string
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&s); err != nil {
		return fmt.Errorf("cannot unmarshal camelz from gob: %w", err)
	}

	parsed, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse camelz from gob: %w", err)
	}
	p.parts = parsed.parts
	return nil
}
