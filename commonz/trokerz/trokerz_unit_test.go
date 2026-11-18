package trokerz_test

import (
	"errors"
	"testing"

	"github.com/infinity6-ai/gox/commonz/trokerz"
	"github.com/stretchr/testify/require"
)

func TestUnitTroke(t *testing.T) {
	type testScenario struct {
		original   string
		opts       trokerz.Options
		want       string
		wantErrMsg string
		wantErr    error
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		got, err := trokerz.Troke(s.original, s.opts)
		if s.wantErr != nil {
			require.Error(t, err)
			require.ErrorIs(t, err, s.wantErr)
			if s.wantErrMsg != "" {
				require.Contains(t, err.Error(), s.wantErrMsg)
			}
			return
		}
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

	t.Run("Unclosed placeholder returns ErrUnclosedPlaceholder", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{unclosed placeholder",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "replaced", true, nil
				},
			},
			wantErr:    trokerz.ErrUnclosedPlaceholder,
			wantErrMsg: "unclosed placeholder",
		})
	})

	t.Run("Interleaved opening delimiter returns ErrUnclosedPlaceholder", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{ unclosed {{first}} and {{last}}!",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			wantErr:    trokerz.ErrUnclosedPlaceholder,
			wantErrMsg: "unclosed placeholder",
		})
	})

	t.Run("Nested delimiters without escaping return ErrUnclosedPlaceholder", func(t *testing.T) {
		check(t, testScenario{
			original: "{{ {{first}} }}",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			wantErr:    trokerz.ErrUnclosedPlaceholder,
			wantErrMsg: "unclosed placeholder",
		})
	})

	t.Run("Unmatched closing delimiter returns ErrUnmatchedClosingDelimiter", func(t *testing.T) {
		check(t, testScenario{
			original: "hello }} world",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			wantErr:    trokerz.ErrUnmatchedClosingDelimiter,
			wantErrMsg: "unmatched closing delimiter",
		})
	})

	t.Run("Unmatched closing delimiter before valid placeholder returns ErrUnmatchedClosingDelimiter", func(t *testing.T) {
		check(t, testScenario{
			original: "hello }} {{first}}",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			wantErr:    trokerz.ErrUnmatchedClosingDelimiter,
			wantErrMsg: "unmatched closing delimiter",
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

	t.Run("Same begin and end delimiters with no delimiters in input", func(t *testing.T) {
		check(t, testScenario{
			original: "no delimiters",
			opts: trokerz.Options{
				Begin: "%",
				End:   "%",
				Get: func(v string) (string, bool, error) {
					return "should not call", true, nil
				},
			},
			want: "no delimiters",
		})
	})

	t.Run("Same begin and end delimiters unclosed returns ErrUnclosedPlaceholder", func(t *testing.T) {
		check(t, testScenario{
			original: "hello %unclosed",
			opts: trokerz.Options{
				Begin: "%",
				End:   "%",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			wantErr:    trokerz.ErrUnclosedPlaceholder,
			wantErrMsg: "unclosed placeholder",
		})
	})

	t.Run("Escaped symmetric delimiters", func(t *testing.T) {
		check(t, testScenario{
			original: "100%% complete: %task% (100%% done)",
			opts: trokerz.Options{
				Begin: "%",
				End:   "%",
				Get: func(v string) (string, bool, error) {
					if v == "task" {
						return "deploy", true, nil
					}
					return "", false, nil
				},
			},
			want: "100% complete: deploy (100% done)",
		})
	})

	t.Run("Escaped symmetric delimiters followed by placeholder", func(t *testing.T) {
		check(t, testScenario{
			original: "%%%task%",
			opts: trokerz.Options{
				Begin: "%",
				End:   "%",
				Get: func(v string) (string, bool, error) {
					if v == "task" {
						return "deploy", true, nil
					}
					return "", false, nil
				},
			},
			want: "%deploy",
		})
	})

	t.Run("Escaped asymmetric delimiters with single characters", func(t *testing.T) {
		check(t, testScenario{
			original: "use {{name}} or {val}",
			opts: trokerz.Options{
				Begin: "{",
				End:   "}",
				Get: func(v string) (string, bool, error) {
					if v == "val" {
						return "42", true, nil
					}
					return "", false, nil
				},
			},
			want: "use {name} or 42",
		})
	})

	t.Run("Escaped asymmetric delimiters combined with placeholder", func(t *testing.T) {
		check(t, testScenario{
			original: "nested: {{{val}}}",
			opts: trokerz.Options{
				Begin: "{",
				End:   "}",
				Get: func(v string) (string, bool, error) {
					if v == "val" {
						return "42", true, nil
					}
					return "", false, nil
				},
			},
			want: "nested: {42}",
		})
	})

	t.Run("Escaped multi-character delimiters", func(t *testing.T) {
		check(t, testScenario{
			original: "raw {{{{key}}}} and {{key}}",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(v string) (string, bool, error) {
					if v == "key" {
						return "replaced", true, nil
					}
					return "", false, nil
				},
			},
			want: "raw {{key}} and replaced",
		})
	})

	t.Run("Escaped closing delimiters alone", func(t *testing.T) {
		check(t, testScenario{
			original: "closing }} and more",
			opts: trokerz.Options{
				Begin: "{",
				End:   "}",
				Get: func(v string) (string, bool, error) {
					return "val", true, nil
				},
			},
			want: "closing } and more",
		})
	})

	t.Run("Returns error on missing Begin option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				End: "}",
				Get: func(v string) (string, bool, error) {
					return "", true, nil
				},
			},
			wantErr:    trokerz.ErrInvalidOptions,
			wantErrMsg: "begin, end and get are required",
		})
	})

	t.Run("Returns error on missing End option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				Begin: "{",
				Get: func(v string) (string, bool, error) {
					return "", true, nil
				},
			},
			wantErr:    trokerz.ErrInvalidOptions,
			wantErrMsg: "begin, end and get are required",
		})
	})

	t.Run("Returns error on missing Get callback option", func(t *testing.T) {
		check(t, testScenario{
			original: "hello",
			opts: trokerz.Options{
				Begin: "{",
				End:   "}",
			},
			wantErr:    trokerz.ErrInvalidOptions,
			wantErrMsg: "begin, end and get are required",
		})
	})
}

func TestUnitMustTroke(t *testing.T) {
	type testScenario struct {
		original    string
		opts        trokerz.Options
		want        string
		expectPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.expectPanic {
			require.Panics(t, func() {
				_ = trokerz.MustTroke(s.original, s.opts)
			})
			return
		}

		got := trokerz.MustTroke(s.original, s.opts)
		require.Equal(t, s.want, got)
	}

	t.Run("Valid replacement returns result", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{name}}!",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(k string) (string, bool, error) {
					return "world", true, nil
				},
			},
			want: "hello world!",
		})
	})

	t.Run("Panics on invalid options", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{name}}",
			opts: trokerz.Options{
				Begin: "{{",
			},
			expectPanic: true,
		})
	})

	t.Run("Panics on Get callback error", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{name}}",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(k string) (string, bool, error) {
					return "", false, errors.New("lookup failed")
				},
			},
			expectPanic: true,
		})
	})

	t.Run("Panics on unclosed placeholder", func(t *testing.T) {
		check(t, testScenario{
			original: "hello {{name",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(k string) (string, bool, error) {
					return "world", true, nil
				},
			},
			expectPanic: true,
		})
	})

	t.Run("Panics on unmatched closing delimiter", func(t *testing.T) {
		check(t, testScenario{
			original: "hello }} world",
			opts: trokerz.Options{
				Begin: "{{",
				End:   "}}",
				Get: func(k string) (string, bool, error) {
					return "world", true, nil
				},
			},
			expectPanic: true,
		})
	})
}
