package jsonz

import (
	"encoding/json"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Json[T any] struct {
	ref T
}

func (p *Json[T]) MarshalJSON() ([]byte, error) {
	r, err := json.Marshal(p.ref)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal Json[%T]: %w", p.ref, err)
	}
	r, err = json.Marshal(string(r))
	errorz.Check(err)
	return r, nil
}

func (p *Json[T]) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal Json[%T]: %w", p.ref, err)
	}
	if err := json.Unmarshal([]byte(s), &p.ref); err != nil {
		return fmt.Errorf("cannot unmarshal Json[%T]: %w", p.ref, err)
	}
	return nil
}
