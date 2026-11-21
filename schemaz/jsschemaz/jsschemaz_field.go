package jsschemaz

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/jstypez"
)

func (p *Parser) ParseSchemaField(ctx context.Context, baseId string, s any) (*jstypez.Typedef, *jstypez.Typedefs) {
	return p.ParseAny(ctx, baseId, s)
}
