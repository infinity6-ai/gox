package camelz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/stringz/camelz"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		input         string
		expectedParts []string
		expectedError string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := camelz.Parse(s.input)
		if s.expectedError != "" {
			require.Error(t, err)
			require.ErrorIs(t, err, camelz.ErrUnsupported)
			require.Contains(t, err.Error(), s.expectedError)
			require.Nil(t, got)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, s.expectedParts, got.Parts())
	}

	t.Run("PascalCase single word", func(t *testing.T) {
		check(t, testScenario{
			input:         "Foo",
			expectedParts: []string{"foo"},
		})
	})

	t.Run("PascalCase multiple words", func(t *testing.T) {
		check(t, testScenario{
			input:         "FooBar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("PascalCase three words", func(t *testing.T) {
		check(t, testScenario{
			input:         "FooBarBaz",
			expectedParts: []string{"foo", "bar", "baz"},
		})
	})

	t.Run("PascalCase corner case ABC", func(t *testing.T) {
		check(t, testScenario{
			input:         "ABC",
			expectedParts: []string{"a", "b", "c"},
		})
	})

	t.Run("PascalCase single letter A", func(t *testing.T) {
		check(t, testScenario{
			input:         "A",
			expectedParts: []string{"a"},
		})
	})

	t.Run("PascalCase with digits", func(t *testing.T) {
		check(t, testScenario{
			input:         "Foo1Bar2",
			expectedParts: []string{"foo1", "bar2"},
		})
	})

	t.Run("camelCase single word", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo",
			expectedParts: []string{"foo"},
		})
	})

	t.Run("camelCase multiple words", func(t *testing.T) {
		check(t, testScenario{
			input:         "fooBar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("camelCase corner case aBC", func(t *testing.T) {
		check(t, testScenario{
			input:         "aBC",
			expectedParts: []string{"a", "b", "c"},
		})
	})

	t.Run("camelCase single letter", func(t *testing.T) {
		check(t, testScenario{
			input:         "a",
			expectedParts: []string{"a"},
		})
	})

	t.Run("snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_bar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("snake lower multiple parts", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_bar_baz",
			expectedParts: []string{"foo", "bar", "baz"},
		})
	})

	t.Run("snake lower with digits", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_1_bar",
			expectedParts: []string{"foo", "1", "bar"},
		})
	})

	t.Run("kebab lower", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-bar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("kebab lower multiple parts", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-bar-baz",
			expectedParts: []string{"foo", "bar", "baz"},
		})
	})

	t.Run("kebab lower with digits", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-1-bar",
			expectedParts: []string{"foo", "1", "bar"},
		})
	})

	t.Run("empty string error", func(t *testing.T) {
		check(t, testScenario{
			input:         "",
			expectedError: "empty string",
		})
	})

	t.Run("contains space error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo bar",
			expectedError: "invalid character",
		})
	})

	t.Run("starts with digit error", func(t *testing.T) {
		check(t, testScenario{
			input:         "123",
			expectedError: "must start with a letter",
		})
	})

	t.Run("snake starts with digit error", func(t *testing.T) {
		check(t, testScenario{
			input:         "1_foo",
			expectedError: "delimited lower string must start with a lowercase letter",
		})
	})

	t.Run("kebab starts with digit error", func(t *testing.T) {
		check(t, testScenario{
			input:         "1-foo",
			expectedError: "delimited lower string must start with a lowercase letter",
		})
	})

	t.Run("snake with uppercase error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_Bar",
			expectedError: "invalid character in lower delimited string",
		})
	})

	t.Run("kebab with uppercase error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-Bar",
			expectedError: "invalid character in lower delimited string",
		})
	})

	t.Run("mixed delimiters error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_bar-baz",
			expectedError: "mixed delimiters",
		})
	})

	t.Run("snake leading delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "_foo",
			expectedError: "delimited lower string must start with a lowercase letter",
		})
	})

	t.Run("snake trailing delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_",
			expectedError: "trailing delimiter",
		})
	})

	t.Run("kebab leading delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "-foo",
			expectedError: "delimited lower string must start with a lowercase letter",
		})
	})

	t.Run("kebab trailing delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-",
			expectedError: "trailing delimiter",
		})
	})

	t.Run("snake consecutive delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo__bar",
			expectedError: "consecutive delimiters",
		})
	})

	t.Run("kebab consecutive delimiter error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo--bar",
			expectedError: "consecutive delimiters",
		})
	})
}

func TestUnitTransform(t *testing.T) {
	type testScenario struct {
		input      string
		expectedP  string
		expectedC  string
		expectedSL string
		expectedSU string
		expectedQL string
		expectedQU string
		expectedStr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p, err := camelz.Parse(s.input)
		require.NoError(t, err)
		require.NotNil(t, p)

		require.Equal(t, s.expectedP, p.P())
		require.Equal(t, s.expectedC, p.C())
		require.Equal(t, s.expectedSL, p.SL())
		require.Equal(t, s.expectedSU, p.SU())
		require.Equal(t, s.expectedQL, p.QL())
		require.Equal(t, s.expectedQU, p.QU())
		require.Equal(t, s.expectedStr, p.String())
	}

	t.Run("from snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:       "foo_bar",
			expectedP:   "FooBar",
			expectedC:   "fooBar",
			expectedSL:  "foo_bar",
			expectedSU:  "FOO_BAR",
			expectedQL:  "foo-bar",
			expectedQU:  "FOO-BAR",
			expectedStr: "foo-bar",
		})
	})

	t.Run("from kebab lower", func(t *testing.T) {
		check(t, testScenario{
			input:       "foo-bar",
			expectedP:   "FooBar",
			expectedC:   "fooBar",
			expectedSL:  "foo_bar",
			expectedSU:  "FOO_BAR",
			expectedQL:  "foo-bar",
			expectedQU:  "FOO-BAR",
			expectedStr: "foo-bar",
		})
	})

	t.Run("from camelCase", func(t *testing.T) {
		check(t, testScenario{
			input:       "fooBar",
			expectedP:   "FooBar",
			expectedC:   "fooBar",
			expectedSL:  "foo_bar",
			expectedSU:  "FOO_BAR",
			expectedQL:  "foo-bar",
			expectedQU:  "FOO-BAR",
			expectedStr: "foo-bar",
		})
	})

	t.Run("from PascalCase", func(t *testing.T) {
		check(t, testScenario{
			input:       "FooBar",
			expectedP:   "FooBar",
			expectedC:   "fooBar",
			expectedSL:  "foo_bar",
			expectedSU:  "FOO_BAR",
			expectedQL:  "foo-bar",
			expectedQU:  "FOO-BAR",
			expectedStr: "foo-bar",
		})
	})

	t.Run("from ABC", func(t *testing.T) {
		check(t, testScenario{
			input:       "ABC",
			expectedP:   "ABC",
			expectedC:   "aBC",
			expectedSL:  "a_b_c",
			expectedSU:  "A_B_C",
			expectedQL:  "a-b-c",
			expectedQU:  "A-B-C",
			expectedStr: "a-b-c",
		})
	})

	t.Run("from single letter a", func(t *testing.T) {
		check(t, testScenario{
			input:       "a",
			expectedP:   "A",
			expectedC:   "a",
			expectedSL:  "a",
			expectedSU:  "A",
			expectedQL:  "a",
			expectedQU:  "A",
			expectedStr: "a",
		})
	})

	t.Run("from single letter A", func(t *testing.T) {
		check(t, testScenario{
			input:       "A",
			expectedP:   "A",
			expectedC:   "a",
			expectedSL:  "a",
			expectedSU:  "A",
			expectedQL:  "a",
			expectedQU:  "A",
			expectedStr: "a",
		})
	})
}

func TestUnitNilAndEmpty(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var p *camelz.Parsed
		require.Equal(t, "", p.P())
		require.Equal(t, "", p.C())
		require.Equal(t, "", p.SL())
		require.Equal(t, "", p.SU())
		require.Equal(t, "", p.QL())
		require.Equal(t, "", p.QU())
		require.Equal(t, "", p.String())
		require.Nil(t, p.Parts())
	})

	t.Run("empty parsed", func(t *testing.T) {
		p := &camelz.Parsed{}
		require.Equal(t, "", p.P())
		require.Equal(t, "", p.C())
		require.Equal(t, "", p.SL())
		require.Equal(t, "", p.SU())
		require.Equal(t, "", p.QL())
		require.Equal(t, "", p.QU())
		require.Equal(t, "", p.String())
		require.Empty(t, p.Parts())
	})
}

func TestUnitPAndMustParse(t *testing.T) {
	t.Run("P helper valid", func(t *testing.T) {
		p := camelz.P("foo_bar")
		require.NotNil(t, p)
		require.Equal(t, "FooBar", p.P())
	})

	t.Run("P helper panic on error", func(t *testing.T) {
		require.Panics(t, func() {
			camelz.P("invalid format")
		})
	})

	t.Run("MustParse helper valid", func(t *testing.T) {
		p := camelz.MustParse("FooBar")
		require.NotNil(t, p)
		require.Equal(t, "foo_bar", p.SL())
	})

	t.Run("MustParse helper panic on error", func(t *testing.T) {
		require.Panics(t, func() {
			camelz.MustParse("")
		})
	})
}
