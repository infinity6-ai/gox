package apiz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/stretchr/testify/require"
)

type dummyApiWithUnimplemented struct {
	apiz.UnimplementedDataRefs
}

func (d *dummyApiWithUnimplemented) ApiSpec() apiz.ApiSpec {
	return apiz.ApiSpec{
		Id: "dummy-spec",
	}
}

type dummyApiWithShadowedMethod struct {
	apiz.UnimplementedDataRefs
}

func (d *dummyApiWithShadowedMethod) ApiSpec() apiz.ApiSpec {
	return apiz.ApiSpec{
		Id: "dummy-spec-shadowed",
	}
}

func (d *dummyApiWithShadowedMethod) ApiDataRefs() *apiz.DataRefs {
	return &apiz.DataRefs{}
}

func TestUnitUnimplementedDataRefs(t *testing.T) {
	type testScenario struct {
		name          string
		api           apiz.Api
		shouldPanic   bool
		panicContains string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		require.NotNil(t, s.api)
		require.NotEmpty(t, s.api.ApiSpec().Id)

		if s.shouldPanic {
			require.PanicsWithValue(t, s.panicContains, func() {
				s.api.ApiDataRefs()
			})
			return
		}

		require.NotPanics(t, func() {
			refs := s.api.ApiDataRefs()
			require.NotNil(t, refs)
		})
	}

	t.Run("Panics when ApiDataRefs is not shadowed", func(t *testing.T) {
		check(t, testScenario{
			name:          "Panics when ApiDataRefs is not shadowed",
			api:           &dummyApiWithUnimplemented{},
			shouldPanic:   true,
			panicContains: "ApiDataRefs not implemented: run 'go generate ./...'",
		})
	})

	t.Run("Dispatches to real method when shadowed", func(t *testing.T) {
		check(t, testScenario{
			name:        "Dispatches to real method when shadowed",
			api:         &dummyApiWithShadowedMethod{},
			shouldPanic: false,
		})
	})
}
