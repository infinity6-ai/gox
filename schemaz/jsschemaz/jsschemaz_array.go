package jsschemaz

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchemaArray(ctx context.Context, baseId string, s *schemaz.Array) (*jstypez.Typedef, *jstypez.Typedefs) {
	element := s.Get(0, false)
	eTypedef, eTypedefs := p.ParseSchema(ctx, baseId+"Element", element)
	typedef := &jstypez.Typedef{
		Head: &jstypez.Tuple{Typ: fmt.Sprintf("Array<%s>", eTypedef.Head.Name)},
	}
	typedefs := eTypedefs
	if typedefs == nil {
		typedefs = &jstypez.Typedefs{}
	}
	typedefs.Add(eTypedef)
	return typedef, typedefs
}
