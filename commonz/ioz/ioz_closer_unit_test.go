package ioz_test

import (
	"errors"
	"testing"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/stretchr/testify/require"
)

func TestUnitCloser(t *testing.T) {
	type testScenario struct {
		f      func() error
		errMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		c := ioz.Closer(s.f)
		require.NotNil(t, c)
		err := c.Close()
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)
	}

	t.Run("returns nil when underlying function succeeds", func(t *testing.T) {
		called := false
		check(t, testScenario{
			f: func() error {
				called = true
				return nil
			},
		})
		require.True(t, called)
	})

	t.Run("returns wrapped error when underlying function fails", func(t *testing.T) {
		check(t, testScenario{
			f: func() error {
				return errors.New("underlying closer failure")
			},
			errMsg: "failed to close: underlying closer failure",
		})
	})

	t.Run("returns nil when function is nil", func(t *testing.T) {
		check(t, testScenario{
			f: nil,
		})
	})
}

func TestUnitCloserV(t *testing.T) {
	type testScenario struct {
		f func()
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		c := ioz.CloserV(s.f)
		require.NotNil(t, c)
		err := c.Close()
		require.NoError(t, err)
	}

	t.Run("invokes void function on close", func(t *testing.T) {
		called := false
		check(t, testScenario{
			f: func() {
				called = true
			},
		})
		require.True(t, called)
	})

	t.Run("handles nil void function safely", func(t *testing.T) {
		check(t, testScenario{
			f: nil,
		})
	})
}
