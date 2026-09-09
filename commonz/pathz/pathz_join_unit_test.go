package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathJoin(t *testing.T) {
	type testScenario struct {
		name   string
		base   *Path
		others []*Path
		want   string
		errMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := s.base.Join(s.others...)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, s.want, got.String())
	}

	t.Run("join one", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("c")},
			want:   "a/b/c",
		})
	})

	t.Run("join multiple", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("c"), MustParse("d")},
			want:   "a/b/c/d",
		})
	})

	t.Run("join with parents", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("../c")},
			want:   "a/c",
			errMsg: "path escaped error",
		})
	})

	t.Run("escape fails", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("../../c")},
			want:   "c",
			errMsg: "path escaped error",
		})
	})

	t.Run("escape fails too much", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("../../../c")},
			want:   "../c",
			errMsg: "path escaped error",
		})
	})

	t.Run("absolute child", func(t *testing.T) {
		check(t, testScenario{
			base:   MustParse("a/b"),
			others: []*Path{MustParse("/c")},
			want:   "/c",
			errMsg: "path escaped error",
		})
	})
}

func TestUnitPathJoinNames(t *testing.T) {
	type testScenario struct {
		name   string
		base   string
		others []string
		want   string
		errMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		p, err := Parse(s.base)
		require.NoError(t, err)
		got, err := p.JoinNamesString(s.others...)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)

			// Test Must* version
			require.Panics(t, func() {
				p, err := Parse(s.base)
				require.NoError(t, err)
				p.MustJoinNamesString(s.others...)
			})
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.want, got.String())

		// Test Must* version
		p, err = Parse(s.base)
		require.NoError(t, err)
		mustGot := p.MustJoinNamesString(s.others...)
		require.Equal(t, s.want, mustGot.String())
	}

	t.Run("valid names", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"c", "d"},
			want:   "a/b/c/d",
		})
	})

	t.Run("empty name", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{""},
			errMsg: "it is not a name:",
		})
	})

	t.Run("multi-part name", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"c/d"},
			errMsg: "it is not a name: c/d",
		})
	})

	t.Run("name with ..", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"../c"},
			errMsg: "max parents allowed",
		})
	})

	t.Run("absolute name", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"/c"},
			errMsg: "path absolute flag",
		})
	})

	t.Run("wildcard in name", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"c*"},
			errMsg: "wildcard characters when not allowed",
		})
	})

	t.Run("mixed valid and invalid names", func(t *testing.T) {
		check(t, testScenario{
			base:   "a/b",
			others: []string{"c", "d/e"},
			errMsg: "it is not a name: d/e",
		})
	})
}
