package schemamold

import (
	"context"
	"maps"
	"slices"

	"github.com/infinity6-ai/gox/commonz/stringz/camelz"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func (p *Parser) ParseSchemaObject(ctx context.Context, baseId string, obj map[string]*schemaz.Schema) (*typeastz.Typedef, *typeastz.Typedefs) {
	ret := typeastz.Typedef{
		Head:       &typeastz.Tuple{Typ: typeastz.TypObject},
		Properties: &typeastz.Tuples{},
	}
	tdfs := &typeastz.Typedefs{}
	for _, field := range slices.Sorted(maps.Keys(obj)) {
		schema := obj[field]
		fieldId := baseId + camelz.P(field).P()
		s, schemaTdfs := p.ParseSchema(ctx, fieldId, schema)
		typ := s.Head.Typ
		if s.Head.Typ == typeastz.TypObject && s.Properties != nil {
			tdfs.Add(s)
			typ = s.Head.Name
		}
		ret.Properties.Add(typeastz.Tuple{
			Name:    field,
			Typ:     typ,
			GoTyp:   s.Head.GoTyp,
			Summary: DescSummary(schema.Desc),
		})
		tdfs.ExtractFrom(schemaTdfs)
	}
	return &ret, tdfs
}
