package apispecz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/apispecz"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseApis(t *testing.T) {
	ctx := t.Context()
	type testScenario struct {
		in   []apiz.Api
		want *apispecz.Apis
	}

	parser := apispecz.New(nil)

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		actual := parser.ParseApis(ctx, s.in)
		var expected *apispecz.Apis
		if s.want != nil {
			jsonz.MustParse(jsonz.MustFormat(s.want).String(), &expected)
		}
		jsonz.MustParse(jsonz.MustFormat(actual).String(), &actual)
		require.Equal(t, expected, actual)
	}

	t.Run("empty apis slice", func(t *testing.T) {
		check(t, testScenario{
			in: []apiz.Api{},
			want: &apispecz.Apis{
				Apis:     []*apispecz.Api{},
				Typedefs: typeastz.NewTypedefs(nil),
			},
		})
	})

	t.Run("nil apis slice", func(t *testing.T) {
		check(t, testScenario{
			in: nil,
			want: &apispecz.Apis{
				Apis:     []*apispecz.Api{},
				Typedefs: typeastz.NewTypedefs(nil),
			},
		})
	})

	t.Run("single api", func(t *testing.T) {
		api := routezsamplefraction.NewService().Api()
		expectedApi := parser.ParseApi(ctx, api)
		check(t, testScenario{
			in: []apiz.Api{api},
			want: &apispecz.Apis{
				Apis:     []*apispecz.Api{expectedApi},
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
			want: &apispecz.Apis{
				Apis:     []*apispecz.Api{expectedApi, expectedApi},
				Typedefs: expectedApi.Typedefs,
			},
		})
	})
}
