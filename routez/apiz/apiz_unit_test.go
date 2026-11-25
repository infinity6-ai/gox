package apiz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/stretchr/testify/require"
)

type dummyApiWithUnimplemented struct {
	apiz.UnimplementedApi
}

type dummyApiWithShadowedMethods struct {
	apiz.UnimplementedApi
}

func (d *dummyApiWithShadowedMethods) ApiSpec() apiz.ApiSpec {
	return apiz.ApiSpec{
		Id: "dummy-spec-shadowed",
	}
}

func (d *dummyApiWithShadowedMethods) ApiDataRefs() *apiz.DataRefs {
	return &apiz.DataRefs{}
}

func TestUnitUnimplementedApi(t *testing.T) {
	type testScenario struct {
		name              string
		api               apiz.Api
		shouldPanic       bool
		specPanicContains string
		refsPanicContains string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		require.NotNil(t, s.api)

		if s.shouldPanic {
			require.PanicsWithValue(t, s.specPanicContains, func() {
				s.api.ApiSpec()
			})
			require.PanicsWithValue(t, s.refsPanicContains, func() {
				s.api.ApiDataRefs()
			})
			return
		}

		require.NotPanics(t, func() {
			spec := s.api.ApiSpec()
			require.Equal(t, "dummy-spec-shadowed", spec.Id)
			refs := s.api.ApiDataRefs()
			require.NotNil(t, refs)
		})
	}

	t.Run("Panics for both methods when not shadowed", func(t *testing.T) {
		check(t, testScenario{
			name:              "Panics for both methods when not shadowed",
			api:               &dummyApiWithUnimplemented{},
			shouldPanic:       true,
			specPanicContains: "ApiSpec not implemented: run 'go generate ./...'",
			refsPanicContains: "ApiDataRefs not implemented: run 'go generate ./...'",
		})
	})

	t.Run("Dispatches to real methods when shadowed", func(t *testing.T) {
		check(t, testScenario{
			name:        "Dispatches to real methods when shadowed",
			api:         &dummyApiWithShadowedMethods{},
			shouldPanic: false,
		})
	})
}
