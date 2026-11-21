package jsschemaz

import (
	"context"
	"reflect"
	"time"

	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Parser struct {
	known map[reflect.Type]*typeastz.Tuple
}

// New creates a new base Parser initialized with standard library types (e.g. time.Time).
func New() *Parser {
	p := &Parser{
		known: make(map[reflect.Type]*typeastz.Tuple),
	}
	p.AddKnownType(time.Time{}, typeastz.TypTime)
	return p
}

// AddKnownType registers how an external Go type is represented in JavaScript.
// Supports fluent chaining: p.AddKnownType(...).AddKnownType(...)
func (p *Parser) AddKnownType(sample any, jsType string) *Parser {
	t := reflect.TypeOf(sample)
	p.known[t] = &typeastz.Tuple{
		Typ:   jsType,
		GoTyp: t.String(),
	}
	return p
}

func (p *Parser) knownTypes(ctx context.Context, baseId string, val reflect.Value) (*typeastz.Typedef, *typeastz.Typedefs) {
	if !val.IsValid() {
		return nil, nil
	}
	v := val.Interface()
	switch s := v.(type) {
	case schemaz.Schema:
		a, b := p.ParseSchema(ctx, baseId, &s)
		a.Head.Name = ""
		return a, b
	case *schemaz.Schema:
		a, b := p.ParseSchema(ctx, baseId, s)
		a.Head.Name = ""
		return a, b
	}

	t := reflect.TypeOf(v)
	if tuple, ok := p.known[t]; ok {
		return &typeastz.Typedef{
			Head: &typeastz.Tuple{
				Typ:   tuple.Typ,
				GoTyp: tuple.GoTyp,
			},
		}, nil
	}

	return nil, nil
}
