package schemamold

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/typeastz"
)

func (p *Parser) ParseSchemaField(ctx context.Context, baseId string, s any) (*typeastz.Typedef, *typeastz.Typedefs) {
	return p.ParseAny(ctx, baseId, s)
}
