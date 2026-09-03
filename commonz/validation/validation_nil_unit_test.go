package validation_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/stretchr/testify/require"
)

type myError struct{}

func (e *myError) Error() string { return "my error" }

func TestUnitNil(t *testing.T) {
	type testScenario struct {
		value   any
		wantErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		err := validation.Nil(s.value, "test message")
		if s.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}

	t.Run("nil value should pass", func(t *testing.T) {
		check(t, testScenario{
			value:   nil,
			wantErr: false,
		})
	})

	t.Run("typed nil pointer should pass", func(t *testing.T) {
		check(t, testScenario{
			value:   (*myError)(nil),
			wantErr: false,
		})
	})

	t.Run("typed nil interface should pass", func(t *testing.T) {
		var err error = nil
		check(t, testScenario{
			value:   err,
			wantErr: false,
		})
	})

	t.Run("non-nil value should fail", func(t *testing.T) {
		check(t, testScenario{
			value:   "hello",
			wantErr: true,
		})
	})

	t.Run("non-nil pointer should fail", func(t *testing.T) {
		check(t, testScenario{
			value:   &myError{},
			wantErr: true,
		})
	})
}

func TestUnitNotNil(t *testing.T) {
	type testScenario struct {
		value   any
		wantErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		err := validation.NotNil(s.value, "test message")
		if s.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}

	t.Run("nil value should fail", func(t *testing.T) {
		check(t, testScenario{
			value:   nil,
			wantErr: true,
		})
	})

	t.Run("typed nil pointer should fail", func(t *testing.T) {
		check(t, testScenario{
			value:   (*myError)(nil),
			wantErr: true,
		})
	})

	t.Run("typed nil interface should fail", func(t *testing.T) {
		var err error = nil
		check(t, testScenario{
			value:   err,
			wantErr: true,
		})
	})

	t.Run("non-nil value should pass", func(t *testing.T) {
		check(t, testScenario{
			value:   "hello",
			wantErr: false,
		})
	})

	t.Run("non-nil pointer should pass", func(t *testing.T) {
		check(t, testScenario{
			value:   &myError{},
			wantErr: false,
		})
	})
}
