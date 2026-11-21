package jsschemaz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/schemaz/jsschemaz"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
	"github.com/stretchr/testify/require"
)

type externalCustomType struct {
	Value string
}

func TestUnitParserInstanceAndAddKnownType(t *testing.T) {
	ctx := t.Context()

	parser := jsschemaz.New().AddKnownType(externalCustomType{}, "MyCustomJsString")

	var val externalCustomType
	schema := &schemaz.Schema{
		Object: func(_ bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"custom_field": {
					Field: func() any { return &val },
				},
			}
		},
	}

	td, _ := parser.ParseSchema(ctx, "testCustom", schema)
	require.NotNil(t, td)
	require.Len(t, td.Properties.All(), 1)
	require.Equal(t, "custom_field", td.Properties.All()[0].Name)
	require.Equal(t, "MyCustomJsString", td.Properties.All()[0].Typ)
}

func TestUnitDescHelpers(t *testing.T) {
	d := &schemaz.Desc{
		Name:    "TestName",
		Summary: "TestSummary",
	}
	require.Equal(t, "TestName", jsschemaz.DescName(d))
	require.Equal(t, "TestSummary", jsschemaz.DescSummary(d))
	require.Equal(t, "", jsschemaz.DescName(nil))
	require.Equal(t, "", jsschemaz.DescSummary(nil))
}

func TestUnitParseSchemaNestedObject(t *testing.T) {
	ctx := t.Context()
	var s string
	schema := &schemaz.Schema{
		Object: func(_ bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"params": {
					Object: func(_ bool) map[string]*schemaz.Schema {
						return map[string]*schemaz.Schema{
							"min": {Field: func() any { return &s }},
						}
					},
				},
			}
		},
	}

	parser := jsschemaz.New()
	td, tds := parser.ParseSchema(ctx, "in", schema)
	require.Equal(t, "inParams", td.Properties.All()[0].Typ, "nested object must be typed, not a bare object")

	var found bool
	for _, d := range tds.All() {
		if d.Head.Name == "inParams" {
			found = true
			require.Equal(t, "string", d.Properties.All()[0].Typ)
		}
	}
	require.True(t, found)
}

type sampleCustomEnum int

func (sampleCustomEnum) JSTypedef() (*jstypez.Typedef, *jstypez.Typedefs) {
	ref := &jstypez.Typedef{
		Head: &jstypez.Tuple{
			Typ:   "SampleCustomEnum",
			GoTyp: "sampleCustomEnum",
		},
	}
	def := &jstypez.Typedef{
		Head: &jstypez.Tuple{
			Name:  "SampleCustomEnum",
			Typ:   "1 | 2 | 4 | number",
			GoTyp: "sampleCustomEnum",
		},
	}
	tds := jstypez.NewTypedefs([]*jstypez.Typedef{def})
	return ref, &tds
}

func TestUnitParseSchemaCustomType(t *testing.T) {
	ctx := t.Context()
	var g sampleCustomEnum
	schema := &schemaz.Schema{
		Object: func(_ bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"custom_enum": {
					Field: func() any { return &g },
				},
			}
		},
	}

	parser := jsschemaz.New()
	td, tds := parser.ParseSchema(ctx, "testReq", schema)
	require.NotNil(t, td)
	require.Equal(t, "SampleCustomEnum", td.Properties.All()[0].Typ)

	var found bool
	for _, d := range tds.All() {
		if d.Head.Name == "SampleCustomEnum" {
			found = true
			require.Equal(t, "1 | 2 | 4 | number", d.Head.Typ)
			require.Equal(t, "sampleCustomEnum", d.Head.GoTyp)
		}
	}
	require.True(t, found)
}
