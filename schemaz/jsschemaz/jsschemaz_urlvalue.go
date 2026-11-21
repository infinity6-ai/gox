package jsschemaz

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchemaUrlValue(ctx context.Context, baseId string, parser *schemaz.Parser[string]) (*jstypez.Typedef, *jstypez.Typedefs) {
	_, v := parser.Format()
	return p.ParseAny(ctx, baseId, v)
}
