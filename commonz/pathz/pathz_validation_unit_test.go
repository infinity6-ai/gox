package pathz_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitValidate(t *testing.T) {
	p := func(s string) *pathz.Path {
		res, err := pathz.Parse(s)
		require.NoError(t, err)
		return res
	}

	type testScenario struct {
		path   *pathz.Path
		opts   pathz.ValidateOptions
		errMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		err := s.path.Validate(s.opts)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)
	}

	t.Run("MinPart 0 allows empty path", func(t *testing.T) {
		check(t, testScenario{
			path: p(""),
			opts: pathz.ValidateOptions{
				MinPart: 0,
			},
		})
	})

	t.Run("MinPart 0 allows non-empty path", func(t *testing.T) {
		check(t, testScenario{
			path: p("a/b"),
			opts: pathz.ValidateOptions{
				MinPart: 0,
			},
		})
	})

	t.Run("MinPart 1 fails on empty path", func(t *testing.T) {
		check(t, testScenario{
			path: p(""),
			opts: pathz.ValidateOptions{
				MinPart: 1,
			},
			errMsg: "min parts allowed",
		})
	})

	t.Run("MinPart 1 succeeds on single part path", func(t *testing.T) {
		check(t, testScenario{
			path: p("a"),
			opts: pathz.ValidateOptions{
				MinPart: 1,
			},
		})
	})

	t.Run("MinPart 2 fails on single part path", func(t *testing.T) {
		check(t, testScenario{
			path: p("a"),
			opts: pathz.ValidateOptions{
				MinPart: 2,
			},
			errMsg: "min parts allowed",
		})
	})

	t.Run("MinPart 2 succeeds on two parts path", func(t *testing.T) {
		check(t, testScenario{
			path: p("a/b"),
			opts: pathz.ValidateOptions{
				MinPart: 2,
			},
		})
	})

	t.Run("MaxPart 0 succeeds on empty path (empty equivalent)", func(t *testing.T) {
		max0 := 0
		check(t, testScenario{
			path: p(""),
			opts: pathz.ValidateOptions{
				MaxPart: &max0,
			},
		})
	})

	t.Run("MaxPart 0 fails on non-empty path", func(t *testing.T) {
		max0 := 0
		check(t, testScenario{
			path: p("a"),
			opts: pathz.ValidateOptions{
				MaxPart: &max0,
			},
			errMsg: "max parts allowed",
		})
	})

	t.Run("MaxPart 1 succeeds on 1 part path", func(t *testing.T) {
		max1 := 1
		check(t, testScenario{
			path: p("a"),
			opts: pathz.ValidateOptions{
				MaxPart: &max1,
			},
		})
	})

	t.Run("MaxPart 1 fails on 2 parts path", func(t *testing.T) {
		max1 := 1
		check(t, testScenario{
			path: p("a/b"),
			opts: pathz.ValidateOptions{
				MaxPart: &max1,
			},
			errMsg: "max parts allowed",
		})
	})

	t.Run("Part function succeeds when returning nil", func(t *testing.T) {
		visited := make(map[int]string)
		check(t, testScenario{
			path: p("foo/bar/baz"),
			opts: pathz.ValidateOptions{
				Part: func(idx int, part string) error {
					visited[idx] = part
					return nil
				},
			},
		})
		require.Equal(t, map[int]string{0: "foo", 1: "bar", 2: "baz"}, visited)
	})

	t.Run("Part function fails and reports error", func(t *testing.T) {
		check(t, testScenario{
			path: p("foo/invalid/baz"),
			opts: pathz.ValidateOptions{
				Part: func(idx int, part string) error {
					if part == "invalid" {
						return errors.New("prohibited word")
					}
					return nil
				},
			},
			errMsg: "part 1 (invalid) validation failed: prohibited word",
		})
	})

	t.Run("Part function stops on first error", func(t *testing.T) {
		callCount := 0
		check(t, testScenario{
			path: p("bad1/bad2/bad3"),
			opts: pathz.ValidateOptions{
				Part: func(idx int, part string) error {
					callCount++
					return errors.New("fail")
				},
			},
			errMsg: "part 0 (bad1) validation failed: fail",
		})
		require.Equal(t, 1, callCount)
	})

	t.Run("Combined MinPart and MaxPart exact length match", func(t *testing.T) {
		max2 := 2
		check(t, testScenario{
			path: p("foo/bar"),
			opts: pathz.ValidateOptions{
				MinPart: 2,
				MaxPart: &max2,
			},
		})
	})

	t.Run("Combined MinPart MaxPart and Part callback", func(t *testing.T) {
		max2 := 2
		check(t, testScenario{
			path: p("FOO/BAR"),
			opts: pathz.ValidateOptions{
				MinPart: 1,
				MaxPart: &max2,
				Part: func(idx int, part string) error {
					if strings.ToUpper(part) != part {
						return errors.New("must be uppercase")
					}
					return nil
				},
			},
		})
	})

	t.Run("EndingSlash check succeeds when matching", func(t *testing.T) {
		endingSlash := true
		check(t, testScenario{
			path: p("a/b/"),
			opts: pathz.ValidateOptions{
				EndingSlash: &endingSlash,
			},
		})
	})

	t.Run("EndingSlash check fails when not matching", func(t *testing.T) {
		endingSlash := false
		check(t, testScenario{
			path: p("a/b/"),
			opts: pathz.ValidateOptions{
				EndingSlash: &endingSlash,
			},
			errMsg: "path ending slash mismatch",
		})
	})

	t.Run("Absolute check succeeds when matching", func(t *testing.T) {
		absolute := true
		check(t, testScenario{
			path: p("/a/b"),
			opts: pathz.ValidateOptions{
				Absolute: &absolute,
			},
		})
	})

	t.Run("Absolute check fails when not matching", func(t *testing.T) {
		absolute := true
		check(t, testScenario{
			path: p("a/b"),
			opts: pathz.ValidateOptions{
				Absolute: &absolute,
			},
			errMsg: "path absolute flag",
		})
	})

	t.Run("Wildchar disallowed fails when path has wildcard", func(t *testing.T) {
		check(t, testScenario{
			path: p("a/*/c"),
			opts: pathz.ValidateOptions{
				Wildchar: false,
			},
			errMsg: "path contains wildcard characters",
		})
	})

	t.Run("Wildchar allowed succeeds when path has wildcard", func(t *testing.T) {
		check(t, testScenario{
			path: p("a/*/c"),
			opts: pathz.ValidateOptions{
				Wildchar: true,
			},
		})
	})
}

func TestUnitValidatePanics(t *testing.T) {
	p, err := pathz.Parse("a/b")
	require.NoError(t, err)

	t.Run("panic on negative MinPart", func(t *testing.T) {
		require.PanicsWithValue(t, "min part cannot be negative", func() {
			_ = p.Validate(pathz.ValidateOptions{MinPart: -1})
		})
	})

	t.Run("panic on negative MaxPart", func(t *testing.T) {
		neg := -1
		require.PanicsWithValue(t, "max part cannot be negative", func() {
			_ = p.Validate(pathz.ValidateOptions{MaxPart: &neg})
		})
	})

	t.Run("panic when MinPart is greater than MaxPart", func(t *testing.T) {
		max := 1
		require.PanicsWithValue(t, "min part cannot be greater than max part", func() {
			_ = p.Validate(pathz.ValidateOptions{MinPart: 2, MaxPart: &max})
		})
	})

	t.Run("panic on negative MaxParents", func(t *testing.T) {
		neg := -1
		require.PanicsWithValue(t, "max parents cannot be negative, use opts.Absolute instead", func() {
			_ = p.Validate(pathz.ValidateOptions{MaxParents: &neg})
		})
	})
}

func TestUnitValidateCheck(t *testing.T) {
	pValid, err := pathz.Parse("a/b")
	require.NoError(t, err)

	pInvalid, err := pathz.Parse("")
	require.NoError(t, err)

	t.Run("Check does not panic on valid path", func(t *testing.T) {
		require.NotPanics(t, func() {
			pValid.Check(pathz.ValidateOptions{MinPart: 1})
		})
	})

	t.Run("Check panics on invalid path", func(t *testing.T) {
		require.Panics(t, func() {
			pInvalid.Check(pathz.ValidateOptions{MinPart: 1})
		})
	})
}

func TestUnitValidateAbsoluteFile(t *testing.T) {
	t.Run("valid absolute path succeeds", func(t *testing.T) {
		p, err := pathz.Parse("/a/b/c.txt")
		require.NoError(t, err)
		require.NoError(t, p.ValidateAbsoluteFile())
	})

	t.Run("relative path fails", func(t *testing.T) {
		p, err := pathz.Parse("a/b/c.txt")
		require.NoError(t, err)
		require.Error(t, p.ValidateAbsoluteFile())
	})

	t.Run("path with wildcard fails", func(t *testing.T) {
		p, err := pathz.Parse("/a/*/c.txt")
		require.NoError(t, err)
		require.Error(t, p.ValidateAbsoluteFile())
	})
}
