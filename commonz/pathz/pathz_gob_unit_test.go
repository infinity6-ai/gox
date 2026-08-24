package pathz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitPathGobEncoding(t *testing.T) {
	type testScenario struct {
		name           string
		pathString     string
		expectParseErr string // Expected error message from Parse
		expectGobErr   string // Expected error message from GobEncode/GobDecode
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Test Parse first
		p, err := Parse(s.pathString)
		if err != nil && s.name == "Path with Unicode characters" {
			t.Logf("Parse error for %s: %v", s.name, err)
		}
		if s.expectParseErr != "" {
			require.Error(t, err, "Parse should return an error for %s", s.name)
			require.Contains(t, err.Error(), s.expectParseErr, "Parse error message mismatch for %s", s.name)
			return // Skip Gob encoding/decoding if Parse failed as expected
		}
		require.NoError(t, err, "Parse should not return an error for %s", s.name)
		require.NotNil(t, p, "Parsed path should not be nil for %s", s.name)

		// GobEncode the Path
		encoded, err := p.GobEncode()
		if s.expectGobErr != "" && encoded == nil { // Expect GobEncode to fail
			require.Error(t, err, "GobEncode should return an error for %s", s.name)
			require.Contains(t, err.Error(), s.expectGobErr, "GobEncode error message mismatch for %s", s.name)
			return
		}
		require.NoError(t, err, "GobEncode should not return an error for %s", s.name)
		require.NotEmpty(t, encoded, "Encoded bytes should not be empty for %s", s.name)

		// GobDecode into a new Path object
		var decodedP Path
		err = decodedP.GobDecode(encoded)
		if s.expectGobErr != "" && decodedP.Parts() == "" { // Expect GobDecode to fail
			require.Error(t, err, "GobDecode should return an error for %s", s.name)
			require.Contains(t, err.Error(), s.expectGobErr, "GobDecode error message mismatch for %s", s.name)
			return
		}
		require.NoError(t, err, "GobDecode should not return an error for %s", s.name)

		// Verify the decoded Path is equal to the original
		require.Equal(t, p.String(), decodedP.String(), "String representation mismatch for %s", s.name)
		require.Equal(t, p.Parts(), decodedP.Parts(), "Path parts mismatch for %s", s.name)
		require.Equal(t, p.Parents(), decodedP.Parents(), "Path parents mismatch for %s", s.name)
		require.Equal(t, p.HasEndingSlash(), decodedP.HasEndingSlash(), "HasEndingSlash mismatch for %s", s.name)
	}

	t.Run("Valid absolute path", func(t *testing.T) {
		check(t, testScenario{
			name:       "Valid absolute path",
			pathString: "/a/b/c",
		})
	})

	t.Run("Valid relative path", func(t *testing.T) {
		check(t, testScenario{
			name:       "Valid relative path",
			pathString: "a/b/c",
		})
	})

	t.Run("Root path", func(t *testing.T) {
		check(t, testScenario{
			name:       "Root path",
			pathString: "/",
		})
	})

	t.Run("Path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			name:       "Path with ending slash",
			pathString: "/a/b/",
		})
	})

	t.Run("Path with special characters", func(t *testing.T) {
		check(t, testScenario{
			name:       "Path with special characters",
			pathString: "/a-b/c_d/e.f",
		})
	})

	t.Run("Empty path string", func(t *testing.T) {
		check(t, testScenario{
			name:       "Empty path string",
			pathString: "",
		})
	})
	t.Run("Path with multiple dots", func(t *testing.T) {
		check(t, testScenario{
			name:       "Path with multiple dots",
			pathString: "a/../b/./c",
		})
	})

	t.Run("Path with tilde", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path with tilde",
			pathString:     "~/a/b",
			expectParseErr: "path contains illegal character 0 '~' in '~/a/b'",
		})
	})

	t.Run("Path with spaces", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path with spaces",
			pathString:     "/my path/with spaces",
			expectParseErr: "path contains illegal character 3 ' ' in '/my path/with spaces'",
		})
	})
	// Add a test case for a path that, when stringified, contains characters that might
	// cause issues with gob encoding/decoding if not handled correctly.
	// For example, a path segment that looks like a control character or a specific gob type.
	t.Run("Path with complex segment (gob-safe string)", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path with complex segment",
			pathString:     "/path/with/segment-containing-gob-special!@#$",
			expectParseErr: "path contains illegal character 41 '!' in '/path/with/segment-containing-gob-special!@#$'", // The first illegal character will be reported
		})
	})

	// Test case for a path containing Unicode characters
	t.Run("Path with Unicode characters", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path with Unicode characters",
			pathString:     "/\u4f60\u597d/\u4e16\u754c/\u30d1\u30b9",
			expectParseErr: "path contains illegal character 1 '你' in '/你好/世界/パス'", // The first illegal character will be reported
		})
	})
}
