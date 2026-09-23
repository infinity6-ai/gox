package trokerz_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/trokerz"
)

func TestUnitTroke(t *testing.T) {
	type testScenario struct {
		name        string
		original    string
		opts        trokerz.Options
		want        string
		wantErrMsg  string
		expectPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.expectPanic {
			require.Panics(t, func() {
				_, _ = trokerz.Troke(s.original, s.opts)
			})
			return
		}

		got, err := trokerz.Troke(s.original, s.opts)
		if s.wantErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErrMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.want, got)
	}

	t.Run("Replace multiple placeholders", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{first}} {{last}}!",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					switch v {
					case "first":
						return "John", true, nil
					case "last":
						return "Doe", true, nil
					default:
						return "", false, nil
					}
				},
			},
			want: "hello John Doe!",
		})
	})

	t.Run("Retain placeholder when shouldReplace is false", func(t *testing.T) {
		check(t, testScenario{
			original: "keep {{this}} but change {{that}}",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					if v == "that" {
						return "something else", true, nil
					}
					return "", false, nil
				},
			},
			want: "keep {{this}} but change something else",
		})
	})

	t.Run("Unclosed placeholder keeps delimiter and remaining text", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{unclosed placeholder",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "replaced", true, nil
				},
			},
			want: "hello {{unclosed placeholder",
		})
	})

	t.Run("Propagates and wraps error from Get", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{error_key}}!",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "", false, errors.New("db timeout")
				},
			},
			wantErrMsg: `failed to get replacement for key "error_key": db timeout`,
		})
	})

	t.Run("No placeholders in input returns original", func(t *testing.T) {
		check(t, testScenario{
			original: "no placeholders here",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "should not call", true, nil
				},
			},
			want: "no placeholders here",
		})
	})

	t.Run("Empty original string returns empty", func(t *testing.T) {
		check(t, testScenario{
			original: "",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "should not call", true, nil
				},
			},
			want: "",
		})
	})

	t.Run("Same begin and end delimiters", func(t *testing.T) {
		check(t, testScenario{
			original: "hello %first% %last%!",
			opts: trokerz.Options{
				Begin: "%",
				End:   "%",
				Get: func(v string) (string, bool, error) {
					switch v {
					case "first":
						return "Jane", true, nil
					case "last":
						return "Smith", true, nil
					default:
						return "", false, nil
					}
				},
			},
			want: "hello Jane Smith!",
		})
	})

	t.Run("Panic on missing Begin option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				End: "}",
				Get: func(v string) (string, bool, error) {
					return "", true, nil
				},
			},
			expectPanic: true,
		})
	})

	t.Run("Panic on missing End option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				Begin: "{",
				Get: func(v string) (string, bool, error) {
					return "", true, nil
				},
			},
			expectPanic: true,
		})
	})

	t.Run("Panic on missing Get callback option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				Begin: "{",
				End:   "}",
			},
			expectPanic: true,
		})
	})
}
