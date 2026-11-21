package jsapiz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/jsapiz"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/stretchr/testify/require"
)

func TestUnitParseApis(t *testing.T) {
	ctx := t.Context()
	type testScenario struct {
		in   []apiz.Api
		want *jsapiz.Apis
	}

	parser := jsapiz.New(nil)

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		actual := parser.ParseApis(ctx, s.in)
		var expected *jsapiz.Apis
		if s.want != nil {
			jsonz.MustParse(jsonz.MustFormat(s.want).String(), &expected)
		}
		jsonz.MustParse(jsonz.MustFormat(actual).String(), &actual)
		require.Equal(t, expected, actual)
	}

	t.Run("empty apis slice", func(t *testing.T) {
		check(t, testScenario{
			in: []apiz.Api{},
			want: &jsapiz.Apis{
				Apis:     []*jsapiz.Api{},
				Typedefs: jstypez.NewTypedefs(nil),
			},
		})
	})

	t.Run("nil apis slice", func(t *testing.T) {
		check(t, testScenario{
			in: nil,
			want: &jsapiz.Apis{
				Apis:     []*jsapiz.Api{},
				Typedefs: jstypez.NewTypedefs(nil),
			},
		})
	})

	t.Run("single api", func(t *testing.T) {
		api := routezsamplefraction.NewService().Api()
		expectedApi := parser.ParseApi(ctx, api)
		check(t, testScenario{
			in: []apiz.Api{api},
			want: &jsapiz.Apis{
				Apis:     []*jsapiz.Api{expectedApi},
				Typedefs: expectedApi.Typedefs,
			},
		})
	})

	t.Run("duplicate apis merge typedefs idempotently", func(t *testing.T) {
		api1 := routezsamplefraction.NewService().Api()
		api2 := routezsamplefraction.NewService().Api()
		expectedApi := parser.ParseApi(ctx, api1)

		check(t, testScenario{
			in: []apiz.Api{api1, api2},
			want: &jsapiz.Apis{
				Apis:     []*jsapiz.Api{expectedApi, expectedApi},
				Typedefs: expectedApi.Typedefs,
			},
		})
	})
}
