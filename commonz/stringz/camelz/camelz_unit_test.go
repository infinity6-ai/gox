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

	t.Run("mixed case and delimiters a_b-c-pUi", func(t *testing.T) {
		check(t, testScenario{
			input:         "a_b-c-pUi",
			expectedParts: []string{"a", "b", "c", "p", "ui"},
		})
	})

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

	t.Run("mixed delimiters foo_bar-baz", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_bar-baz",
			expectedParts: []string{"foo", "bar", "baz"},
		})
	})

	t.Run("snake with uppercase foo_Bar", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo_Bar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("kebab with uppercase foo-Bar", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo-Bar",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("leading and trailing delimiters", func(t *testing.T) {
		check(t, testScenario{
			input:         "_foo-bar_",
			expectedParts: []string{"foo", "bar"},
		})
	})

	t.Run("consecutive delimiters", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo__bar--baz",
			expectedParts: []string{"foo", "bar", "baz"},
		})
	})

	t.Run("empty string returns nil parts", func(t *testing.T) {
		check(t, testScenario{
			input:         "",
			expectedParts: nil,
		})
	})

	t.Run("contains space error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo bar",
			expectedError: "invalid character",
		})
	})

	t.Run("contains symbol error", func(t *testing.T) {
		check(t, testScenario{
			input:         "foo@bar",
			expectedError: "invalid character",
		})
	})

	t.Run("only delimiters error", func(t *testing.T) {
		check(t, testScenario{
			input:         "---___---",
			expectedError: "no valid parts",
		})
	})
}

func TestUnitTransform(t *testing.T) {
	type testScenario struct {
		input       string
		expectedP   string
		expectedC   string
		expectedSL  string
		expectedSU  string
		expectedQL  string
		expectedQU  string
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

	t.Run("from a_b-c-pUi", func(t *testing.T) {
		check(t, testScenario{
			input:       "a_b-c-pUi",
			expectedP:   "ABCPUi",
			expectedC:   "aBCPUi",
			expectedSL:  "a_b_c_p_ui",
			expectedSU:  "A_B_C_P_UI",
			expectedQL:  "a-b-c-p-ui",
			expectedQU:  "A-B-C-P-UI",
			expectedStr: "a_b_c_p_ui",
		})
	})

	t.Run("from snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:       "foo_bar",
			expectedP:   "FooBar",
			expectedC:   "fooBar",
			expectedSL:  "foo_bar",
			expectedSU:  "FOO_BAR",
			expectedQL:  "foo-bar",
			expectedQU:  "FOO-BAR",
			expectedStr: "foo_bar",
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
			expectedStr: "foo_bar",
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
			expectedStr: "foo_bar",
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
			expectedStr: "foo_bar",
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
			expectedStr: "a_b_c",
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

	t.Run("from empty string", func(t *testing.T) {
		check(t, testScenario{
			input:       "",
			expectedP:   "",
			expectedC:   "",
			expectedSL:  "",
			expectedSU:  "",
			expectedQL:  "",
			expectedQU:  "",
			expectedStr: "",
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
		require.Equal(t, 0, p.Len())
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
		require.Equal(t, 0, p.Len())
		require.Empty(t, p.Parts())
	})
}

func TestUnitGetAndLen(t *testing.T) {
	t.Run("valid identifier parts and len", func(t *testing.T) {
		p := camelz.P("a_b-c-pUi")
		require.Equal(t, 5, p.Len())
		require.Equal(t, "a", p.Get(0))
		require.Equal(t, "b", p.Get(1))
		require.Equal(t, "c", p.Get(2))
		require.Equal(t, "p", p.Get(3))
		require.Equal(t, "ui", p.Get(4))
	})

	t.Run("out of bounds panics", func(t *testing.T) {
		p := camelz.P("foo_bar")
		require.Equal(t, 2, p.Len())
		require.Equal(t, "foo", p.Get(0))
		require.Equal(t, "bar", p.Get(1))
		require.Panics(t, func() { p.Get(-1) })
		require.Panics(t, func() { p.Get(2) })
		require.Panics(t, func() { p.Get(100) })
	})

	t.Run("empty parsed panics on get", func(t *testing.T) {
		p := camelz.P("")
		require.Equal(t, 0, p.Len())
		require.Panics(t, func() { p.Get(0) })
		require.Panics(t, func() { p.Get(-1) })
	})
}

func TestUnitPAndMustParse(t *testing.T) {
	t.Run("P helper valid", func(t *testing.T) {
		p := camelz.P("a_b-c-pUi")
		require.NotNil(t, p)
		require.Equal(t, "ABCPUi", p.P())
		require.Equal(t, []string{"a", "b", "c", "p", "ui"}, p.Parts())
	})

	t.Run("P helper panic on error", func(t *testing.T) {
		require.Panics(t, func() {
			camelz.P("invalid format with space")
		})
	})

	t.Run("MustParse helper valid", func(t *testing.T) {
		p := camelz.MustParse("FooBar")
		require.NotNil(t, p)
		require.Equal(t, "foo_bar", p.SL())
	})

	t.Run("MustParse helper panic on error", func(t *testing.T) {
		require.Panics(t, func() {
			camelz.MustParse("bad string with spaces")
		})
	})
}
