package pathz

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathGobEncodingRoundtrip(t *testing.T) {
	type testScenario struct {
		name string
		path *Path
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// GobEncode the Path
		encoded, err := s.path.GobEncode()
		require.NoError(t, err, "GobEncode should not return an error for %s", s.name)
		require.NotEmpty(t, encoded, "Encoded bytes should not be empty for %s", s.name)

		// GobDecode into a new Path object
		var decodedP Path
		err = decodedP.GobDecode(encoded)
		require.NoError(t, err, "GobDecode should not return an error for %s", s.name)

		// Verify the decoded Path is equal to the original
		require.Equal(t, s.path, &decodedP, "Decoded path mismatch for %s", s.name)
	}

	t.Run("Valid absolute path", func(t *testing.T) {
		check(t, testScenario{
			name: "Valid absolute path",
			path: New(-1, []string{"a", "b", "c"}, false),
		})
	})

	t.Run("Valid relative path", func(t *testing.T) {
		check(t, testScenario{
			name: "Valid relative path",
			path: New(0, []string{"a", "b", "c"}, false),
		})
	})

	t.Run("Root path", func(t *testing.T) {
		check(t, testScenario{
			name: "Root path",
			path: New(-1, nil, false),
		})
	})

	t.Run("Path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			name: "Path with ending slash",
			path: New(-1, []string{"a", "b"}, true),
		})
	})

	t.Run("Path with special characters", func(t *testing.T) {
		check(t, testScenario{
			name: "Path with special characters",
			path: New(-1, []string{"a-b", "c_d", "e.f"}, false),
		})
	})

	t.Run("Empty path string", func(t *testing.T) {
		check(t, testScenario{
			name: "Empty path string",
			path: New(0, nil, false),
		})
	})
	t.Run("Path with multiple dots", func(t *testing.T) {
		check(t, testScenario{
			name: "Path with multiple dots cleaned",
			path: New(0, []string{"a", "c"}, false),
		})
	})
}

func TestUnitPathGobEncodingInvalid(t *testing.T) {
	type testScenario struct {
		name         string
		pathString   string
		expectGobErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(s.pathString)
		require.NoError(t, err)

		var decodedP Path
		err = decodedP.GobDecode(buf.Bytes())
		require.Error(t, err, "GobDecode should return an error for %s", s.name)
		require.Contains(t, err.Error(), s.expectGobErr, "GobDecode error message mismatch for %s", s.name)
	}

	t.Run("Path with tilde", func(t *testing.T) {
		check(t, testScenario{
			name:         "Path with tilde",
			pathString:   "~/a/b",
			expectGobErr: "path contains illegal character 0 '~' in '~/a/b'",
		})
	})

	t.Run("Path with spaces", func(t *testing.T) {
		check(t, testScenario{
			name:         "Path with spaces",
			pathString:   "/my path/with spaces",
			expectGobErr: "path contains illegal character 3 ' ' in '/my path/with spaces'",
		})
	})

	t.Run("Path with complex segment (gob-safe string)", func(t *testing.T) {
		check(t, testScenario{
			name:         "Path with complex segment",
			pathString:   "/path/with/segment-containing-gob-special!@#$",
			expectGobErr: "path contains illegal character 41 '!' in '/path/with/segment-containing-gob-special!@#$'", // The first illegal character will be reported
		})
	})

	t.Run("Path with Unicode characters", func(t *testing.T) {
		check(t, testScenario{
			name:         "Path with Unicode characters",
			pathString:   "/\u4f60\u597d/\u4e16\u754c/\u30d1\u30b9",
			expectGobErr: "path contains illegal character 1 '你' in '/你好/世界/パス'", // The first illegal character will be reported
		})
	})
}
