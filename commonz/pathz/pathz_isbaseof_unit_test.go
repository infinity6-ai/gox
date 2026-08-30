package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitIsBaseOf(t *testing.T) {
	type testScenario struct {
		p1str  string
		p2str  string
		isBase bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p1, err := Parse(s.p1str)
		require.NoError(t, err)
		p2, err := Parse(s.p2str)
		require.NoError(t, err)

		got := p1.IsBaseOf(p2)
		require.Equal(t, s.isBase, got)
	}

	t.Run("simple base", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b",
			p2str:  "a/b/c",
			isBase: true,
		})
	})

	t.Run("identical paths", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b",
			p2str:  "a/b",
			isBase: true,
		})
	})

	t.Run("identical paths with trailing slash", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b/",
			p2str:  "a/b/",
			isBase: true,
		})
	})

	t.Run("base without slash, other with", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b",
			p2str:  "a/b/",
			isBase: true,
		})
	})

	t.Run("base with slash, other without", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b/",
			p2str:  "a/b",
			isBase: false,
		})
	})

	t.Run("not a base - different parts", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/x",
			p2str:  "a/b/c",
			isBase: false,
		})
	})

	t.Run("not a base - p1 is longer", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "a/b/c",
			p2str:  "a/b",
			isBase: false,
		})
	})

	t.Run("absolute vs relative", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "/a/b",
			p2str:  "a/b",
			isBase: false,
		})
	})

	t.Run("root is base of absolute", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "/",
			p2str:  "/a/b",
			isBase: true,
		})
	})

	t.Run("root is not base of relative", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "/",
			p2str:  "a/b",
			isBase: false,
		})
	})

	t.Run("different relative parents", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "../a",
			p2str:  "../../a",
			isBase: false,
		})
	})

	t.Run("same relative parents", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "../a",
			p2str:  "../a/b",
			isBase: true,
		})
	})

	t.Run("empty path is base of relative path", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "",
			p2str:  "a",
			isBase: true,
		})
	})

	t.Run("empty path is not base of absolute path", func(t *testing.T) {
		check(t, testScenario{
			p1str:  "",
			p2str:  "/a",
			isBase: false,
		})
	})

	t.Run("dot path is base of relative path", func(t *testing.T) {
		check(t, testScenario{
			p1str:  ".",
			p2str:  "a",
			isBase: true,
		})
	})

	t.Run("dot path is not base of absolute path", func(t *testing.T) {
		check(t, testScenario{
			p1str:  ".",
			p2str:  "/a",
			isBase: false,
		})
	})
}
