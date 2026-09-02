package regexpz

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitEngine(t *testing.T) {
	type testScenario struct {
		name   string
		engine Engine
		input  string
		want   string
		ok     bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, ok := s.engine.Apply(s.input)
		require.Equal(t, s.ok, ok, s.name)
		require.Equal(t, s.want, got, s.name)
	}

	t.Run("Replace then match success", func(t *testing.T) {
		check(t, testScenario{
			name: "Replace then match success",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Replace 'a' with 'b'",
						Regexp:    regexp.MustCompile("a"),
						Operation: ReplaceOperation,
						Out:       "b",
					},
					{
						Name:      "Match 'b'",
						Regexp:    regexp.MustCompile("b"),
						Operation: MatchOperation,
					},
				},
			},
			input: "aca",
			want:  "bcb",
			ok:    true,
		})
	})

	t.Run("Replace then match fail", func(t *testing.T) {
		check(t, testScenario{
			name: "Replace then match fail",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Replace 'a' with 'b'",
						Regexp:    regexp.MustCompile("a"),
						Operation: ReplaceOperation,
						Out:       "b",
					},
					{
						Name:      "Match 'c'",
						Regexp:    regexp.MustCompile("c"),
						Operation: MatchOperation,
					},
				},
			},
			input: "ada",
			// Engine should return original input on failure
			want: "ada",
			ok:   false,
		})
	})

	t.Run("Mismatch success (no match)", func(t *testing.T) {
		check(t, testScenario{
			name: "Mismatch success (no match)",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Mismatch 'c'",
						Regexp:    regexp.MustCompile("c"),
						Operation: MismatchOperation,
					},
					{
						Name:      "Replace 'a' with 'b'",
						Regexp:    regexp.MustCompile("a"),
						Operation: ReplaceOperation,
						Out:       "b",
					},
				},
			},
			input: "ada",
			want:  "bdb",
			ok:    true,
		})
	})

	t.Run("Mismatch fail (match found)", func(t *testing.T) {
		check(t, testScenario{
			name: "Mismatch fail (match found)",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Mismatch 'c'",
						Regexp:    regexp.MustCompile("c"),
						Operation: MismatchOperation,
					},
					{
						Name:      "Replace 'a' with 'b'",
						Regexp:    regexp.MustCompile("a"),
						Operation: ReplaceOperation,
						Out:       "b",
					},
				},
			},
			input: "aca",
			// Engine should return original input on failure
			want: "aca",
			ok:   false,
		})
	})
    
    t.Run("Split operation", func(t *testing.T) {
		check(t, testScenario{
			name: "Split operation",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Split by comma",
						Regexp:    regexp.MustCompile(","),
						Operation: SplitOperation,
						Out:       "-",
					},
				},
			},
			input: "a,b,c",
			want:  "a-b-c",
			ok:    true,
		})
	})

    t.Run("Delete operation", func(t *testing.T) {
		check(t, testScenario{
			name: "Delete operation",
			engine: Engine{
				Rules: []Rule{
					{
						Name:      "Delete vowels",
						Regexp:    regexp.MustCompile("[aeiou]"),
						Operation: DeleteOperation,
					},
				},
			},
			input: "hello world",
			want:  "hll wrld",
			ok:    true,
		})
	})

    t.Run("Empty rules", func(t *testing.T) {
		check(t, testScenario{
			name:   "Empty rules",
			engine: Engine{Rules: []Rule{}},
			input:  "anything",
			want:   "anything",
			ok:     true,
		})
	})

    t.Run("Chain of replacements", func(t *testing.T) {
		check(t, testScenario{
			name: "Chain of replacements",
			engine: Engine{
				Rules: []Rule{
					{
						Operation: ReplaceOperation,
						Regexp:    regexp.MustCompile("a"),
						Out:       "b",
					},
					{
						Operation: ReplaceOperation,
						Regexp:    regexp.MustCompile("b"),
						Out:       "c",
					},
					{
						Operation: ReplaceOperation,
						Regexp:    regexp.MustCompile("c"),
						Out:       "d",
					},
				},
			},
			input: "a",
			want:  "d",
			ok:    true,
		})
	})
}
