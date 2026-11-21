package schemamold

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchemaUrlValue(ctx context.Context, baseId string, parser *schemaz.Parser[string]) (*typeastz.Typedef, *typeastz.Typedefs) {
	_, v := parser.Format()
	return p.ParseAny(ctx, baseId, v)
}
