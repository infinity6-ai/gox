package jsschemaz

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchema(ctx context.Context, baseId string, schema *schemaz.Schema) (*jstypez.Typedef, *jstypez.Typedefs) {
	var typedef *jstypez.Typedef
	var typedefs *jstypez.Typedefs
	if schema.Object != nil {
		s := schema.Object(false)
		typedef, typedefs = p.ParseSchemaObject(ctx, baseId, s)
	}
	if schema.UrlValue != nil {
		s := schema.UrlValue()
		typedef, typedefs = p.ParseSchemaUrlValue(ctx, baseId, s)
	}
	if schema.Field != nil {
		s := schema.Field()
		typedef, typedefs = p.ParseSchemaField(ctx, baseId, s)
	}
	if schema.Array != nil {
		s := schema.Array()
		typedef, typedefs = p.ParseSchemaArray(ctx, baseId, s)
	}
	if typedef == nil {
		panic(fmt.Sprintf("unsupported: %#v", schema))
	}
	checker.StrEmpty(typedef.Head.Name, "typedef.Head.Name")
	typedef.Head.Name = baseId
	typedef.Head.Summary = DescSummary(schema.Desc)
	return typedef, typedefs
}

func DescName(d *schemaz.Desc) string {
	if d == nil {
		return ""
	}
	return d.Name
}

func DescSummary(d *schemaz.Desc) string {
	if d == nil {
		return ""
	}
	return d.Summary
}
