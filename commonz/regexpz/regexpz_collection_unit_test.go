package regexpz

import (
	"regexp"
	"sync"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/require"
)

func TestUnitCollectionNew(t *testing.T) {
	t.Run("CreatesCollectionWithBufferSize", func(t *testing.T) {
		bufferSize := 10
		c := New(bufferSize)
		require.NotNil(t, c)
		require.NotNil(t, c.coll)
		// We cannot reliably check the exact capacity of a map in Go,
		// but we can ensure it's initialized and not nil.
		require.Empty(t, c.coll)
		require.IsType(t, &sync.RWMutex{}, &c.mu)
	})

	t.Run("CreatesCollectionWithZeroBufferSize", func(t *testing.T) {
		c := New(0)
		require.NotNil(t, c)
		require.NotNil(t, c.coll)
		require.Empty(t, c.coll)
	})
}

func TestUnitCollectionGet(t *testing.T) {
	type testScenario struct {
		name          string
		pattern       string
		expectedError string
		alreadyCached bool
	}

	check := func(t *testing.T, c *Collection, s testScenario) {
		t.Helper()
		// If the pattern is supposed to be cached, add it manually before the test
		if s.alreadyCached {
			re, err := regexp.Compile(s.pattern)
			require.NoError(t, err)
			c.mu.Lock()
			c.coll[s.pattern] = re
			c.mu.Unlock()
		}

		ret, err := c.Get(s.pattern, nil) // The second argument `re` is unused in the current implementation.

		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			require.Nil(t, ret)
		} else {
			require.NoError(t, err)
			require.NotNil(t, ret)
			require.Equal(t, s.pattern, ret.String())
			// Verify it's now in the cache
			require.Contains(t, c.coll, s.pattern)
			require.Equal(t, ret, c.coll[s.pattern])
		}
	}

	c := New(0) // Create a new collection for these tests

	t.Run("ValidPatternCompilesAndCaches", func(t *testing.T) {
		check(t, c, testScenario{
			name:    "simple valid pattern",
			pattern: "^abc$",
		})
	})

	t.Run("AlreadyCachedPatternReturnsFromCache", func(t *testing.T) {
		check(t, c, testScenario{
			name:          "cached valid pattern",
			pattern:       "def.*ghi",
			alreadyCached: true,
		})
	})

	t.Run("InvalidPatternReturnsError", func(t *testing.T) {
		check(t, c, testScenario{
			name:          "invalid regex pattern",
			pattern:       "[",
			expectedError: "error compiling regexp",
		})
	})

	t.Run("EmptyPatternIsValid", func(t *testing.T) {
		check(t, c, testScenario{
			name:    "empty pattern",
			pattern: "",
		})
	})
}

func TestUnitCollectionMust(t *testing.T) {
	c := New(0) // Create a new collection for these tests

	t.Run("ValidPatternReturnsRegexp", func(t *testing.T) {
		pattern := "test.*123"
		re := c.Must(pattern, nil) // The second argument `re` is unused.
		require.NotNil(t, re)
		require.Equal(t, pattern, re.String())
		// Verify it's in the cache
		require.Contains(t, c.coll, pattern)
		require.Equal(t, re, c.coll[pattern])
	})

	t.Run("InvalidPatternPanics", func(t *testing.T) {
		pattern := "("
		require.Panics(t, func() {
			c.Must(pattern, nil) // The second argument `re` is unused.
		})
		// After panic, get the actual panic value and check its type and content
		var panicValue interface{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicValue = r
				}
			}()
			c.Must(pattern, nil)
		}()
		require.NotNil(t, panicValue)

		structuredErr, ok := panicValue.(errorz.StructuredError)
		require.True(t, ok, "Panic value should be of type errorz.StructuredError")
		require.Contains(t, structuredErr.Error(), "error compiling regexp")
		// Ensure invalid pattern is NOT cached
		require.NotContains(t, c.coll, pattern)
	})
}

func TestUnitPackageLevelGet(t *testing.T) {
	// Reset the root collection for clean testing
	root = New(0)

	type testScenario struct {
		name          string
		pattern       string
		expectedError string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ret, err := Get(s.pattern, nil) // The second argument `re` is unused.

		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			require.Nil(t, ret)
		} else {
			require.NoError(t, err)
			require.NotNil(t, ret)
			require.Equal(t, s.pattern, ret.String())
			// Verify it's now in the cache
			require.Contains(t, root.coll, s.pattern)
			require.Equal(t, ret, root.coll[s.pattern])
		}
	}

	t.Run("ValidPatternCompilesAndCaches", func(t *testing.T) {
		check(t, testScenario{
			name:    "simple valid pattern",
			pattern: "xyz",
		})
	})

	t.Run("InvalidPatternReturnsError", func(t *testing.T) {
		check(t, testScenario{
			name:          "invalid regex pattern",
			pattern:       "++",
			expectedError: "error compiling regexp",
		})
	})
}

func TestUnitPackageLevelMust(t *testing.T) {
	// Reset the root collection for clean testing
	root = New(0)

	t.Run("ValidPatternReturnsRegexp", func(t *testing.T) {
		pattern := "alpha.*beta"
		re := Must(pattern, nil) // The second argument `re` is unused.
		require.NotNil(t, re)
		require.Equal(t, pattern, re.String())
		// Verify it's in the cache
		require.Contains(t, root.coll, pattern)
		require.Equal(t, re, root.coll[pattern])
	})

	t.Run("InvalidPatternPanics", func(t *testing.T) {
		pattern := "????"
		require.Panics(t, func() {
			Must(pattern, nil) // The second argument `re` is unused.
		})
		// After panic, get the actual panic value and check its type and content
		var panicValue interface{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicValue = r
				}
			}()
			Must(pattern, nil)
		}()
		require.NotNil(t, panicValue)

		structuredErr, ok := panicValue.(errorz.StructuredError)
		require.True(t, ok, "Panic value should be of type errorz.StructuredError")
		require.Contains(t, structuredErr.Error(), "error compiling regexp")
		// Ensure invalid pattern is NOT cached
		require.NotContains(t, root.coll, pattern)
	})
}
