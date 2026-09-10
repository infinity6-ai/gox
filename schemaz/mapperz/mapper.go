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
	// 1. Parse the incoming JSON into a map of raw byte slices
	// var raw map[string]json.RawMessage
	// if err := json.Unmarshal(data, &raw); err != nil {
	// 	return err
	// }

	_, ok := b.Target.(Mapper)
	if ok {
		panic("mapper")
	}

	m, ok := b.Target.(map[string]*Mapper)
	if ok {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		for key, ble := range m {
			if rawVal, exists := raw[key]; exists {
				if err := json.Unmarshal(rawVal, ble.Target); err != nil {
					return fmt.Errorf("error parsing key %q: %w", key, err)
				}
			}
		}
		return nil
	}

	err := json.Unmarshal(data, b.Target)
	errorz.Check(err)
	return err

	// 2. Iterate over your defined mappings
	// for key, ble := range b.m {
	// 	// If the JSON payload contains this key, unmarshal it into the Target pointer
	// 	if rawVal, exists := raw[key]; exists {
	// 		if err := json.Unmarshal(rawVal, ble.Target); err != nil {
	// 			return fmt.Errorf("error parsing key %q: %w", key, err)
	// 		}
	// 	}
	// }

	// return nil
}
