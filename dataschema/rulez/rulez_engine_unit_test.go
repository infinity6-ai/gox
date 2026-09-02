package rulez

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitEngineApply(t *testing.T) {
	type testScenario struct {
		name        string
		engine      *Engine
		input       string
		expected    string
		ok          bool
		expectedIdx int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		output, idx, ok := s.engine.Apply(s.input)
		require.Equal(t, s.ok, ok)
		require.Equal(t, s.expected, output)
		require.Equal(t, s.expectedIdx, idx)
	}

	t.Run("ComplexTransformWithGrouping", func(t *testing.T) {
		engine := &Engine{
			Rules: []Rule{
				{
					Name:      "Swap words",
					Regexp:    regexp.MustCompile(`(quick)\s(brown)`),
					Operation: ReplaceOperation,
					Out:       `$2 $1`,
				},
				{
					Name:      "Capitalize fox",
					Regexp:    regexp.MustCompile(`(fox)`),
					Operation: ReplaceOperation,
					Out:       `FOX`,
				},
				{
					Name:      "Join with hyphens",
					Regexp:    regexp.MustCompile(`\s`),
					Operation: SplitOperation,
					Out:       "-",
				},
			},
		}

		check(t, testScenario{
			name:        "should transform the string",
			engine:      engine,
			input:       "The quick brown fox jumps over the lazy dog",
			expected:    "The-brown-quick-FOX-jumps-over-the-lazy-dog",
			ok:          true,
			expectedIdx: -1,
		})
	})

	t.Run("MismatchOperationFailsChain", func(t *testing.T) {
		engine := &Engine{
			Rules: []Rule{
				{
					Name:      "Should match",
					Regexp:    regexp.MustCompile(`^start`),
					Operation: MatchOperation,
				},
				{
					Name:      "Should mismatch and fail",
					Regexp:    regexp.MustCompile(`processing`),
					Operation: MismatchOperation,
				},
			},
		}

		check(t, testScenario{
			name:        "mismatch should fail the chain",
			engine:      engine,
			input:       "start processing",
			expected:    "start processing",
			ok:          false,
			expectedIdx: 1,
		})
	})

	t.Run("DeleteOperation", func(t *testing.T) {
		engine := &Engine{
			Rules: []Rule{
				{
					Name:      "Delete vowels",
					Regexp:    regexp.MustCompile(`[aeiou]`),
					Operation: DeleteOperation,
				},
			},
		}

		check(t, testScenario{
			name:        "should delete all vowels",
			engine:      engine,
			input:       "hello world",
			expected:    "hll wrld",
			ok:          true,
			expectedIdx: -1,
		})
	})
}
