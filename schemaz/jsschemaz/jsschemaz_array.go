package jsschemaz

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchemaArray(ctx context.Context, baseId string, s *schemaz.Array) (*typeastz.Typedef, *typeastz.Typedefs) {
	element := s.Get(0, false)
	eTypedef, eTypedefs := p.ParseSchema(ctx, baseId+"Element", element)
	typedef := &typeastz.Typedef{
		Head: &typeastz.Tuple{Typ: fmt.Sprintf("Array<%s>", eTypedef.Head.Name)},
	}
	typedefs := eTypedefs
	if typedefs == nil {
		typedefs = &typeastz.Typedefs{}
	}
	typedefs.Add(eTypedef)
	return typedef, typedefs
}
