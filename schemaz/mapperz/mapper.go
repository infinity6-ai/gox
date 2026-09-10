package mapperz

import (
	"encoding/json"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Mapper struct {
	Target any
}

func (b *Mapper) UnmarshalJSON(data []byte) error {
	m, ok := b.Target.(map[string]*Mapper)
	if ok {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		for key, ble := range m {
			if rawVal, exists := raw[key]; exists {
				if err := json.Unmarshal(rawVal, &ble.Target); err != nil {
					return fmt.Errorf("error parsing key %q: %w", key, err)
				}
			}
		}
		return nil
	}
	err := json.Unmarshal(data, b.Target)
	errorz.Check(err)
	return err
}

type Array struct {
	Element func() *Mapper
}

func (b *Array) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for i, ble := range raw {
		m := b.Element()
		if err := json.Unmarshal(ble, m); err != nil {
			return fmt.Errorf("error parsing key %d: %w", i, err)
		}
		print(111)
	}
	return nil
}
