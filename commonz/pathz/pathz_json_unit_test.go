package pathz

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathJson(t *testing.T) {
	type testScenario struct {
		name      string
		path      *Path
		json      string
		expectErr bool
	}

	check := func(t *testing.T, s testScenario) {
		// Test Marshal
		b, err := json.Marshal(s.path)
		if s.expectErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.json, string(b))

		// Test Unmarshal
		var p *Path
		err = json.Unmarshal([]byte(s.json), &p)
		if s.expectErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.path, p)
	}

	t.Run("absolute path", func(t *testing.T) {
		check(t, testScenario{
			name: "absolute path",
			path: &Path{
				parts:   "a/b",
				parents: -1,
			},
			json: `"/a/b"`,
		})
	})

	t.Run("relative path", func(t *testing.T) {
		check(t, testScenario{
			name: "relative path",
			path: &Path{
				parts:   "a/b",
				parents: 2,
			},
			json: `"../../a/b"`,
		})
	})

	t.Run("path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			name: "path with ending slash",
			path: &Path{
				parts:          "a/b",
				parents:        -1,
				hasEndingSlash: true,
			},
			json: `"/a/b/"`,
		})
	})

	t.Run("empty path", func(t *testing.T) {
		check(t, testScenario{
			name: "empty path",
			path: &Path{},
			json: `""`,
		})
	})

	t.Run("nil path", func(t *testing.T) {
		check(t, testScenario{
			name: "nil path",
			path: nil,
			json: `null`,
		})
	})

	t.Run("invalid json", func(t *testing.T) {
		var p Path
		err := json.Unmarshal([]byte(`"/a/b/invalid`), &p)
		require.Error(t, err)
	})
}
