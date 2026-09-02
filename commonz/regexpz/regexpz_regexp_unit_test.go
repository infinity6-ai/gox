package regexpz

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitRegexpJSONMarshalling(t *testing.T) {
	type testScenario struct {
		name          string
		input         *Regexp
		expectedJSON  string
		expectedError string
	}

	checkMarshal := func(t *testing.T, s testScenario) {
		t.Helper()
		jsonData, err := json.Marshal(s.input)
		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.expectedJSON, string(jsonData))
	}

	t.Run("MarshalValidRegexp", func(t *testing.T) {
		checkMarshal(t, testScenario{
			name:         "valid regexp",
			input:        MustCompile(`^a.*b$`),
			expectedJSON: `"^a.*b$"`,
		})
	})

	t.Run("MarshalEmptyRegexp", func(t *testing.T) {
		checkMarshal(t, testScenario{
			name:         "empty regexp",
			input:        MustCompile(``),
			expectedJSON: `""`,
		})
	})

	t.Run("MarshalNilRegexpInStruct", func(t *testing.T) {
		checkMarshal(t, testScenario{
			name:         "nil regexp in struct",
			input:        &Regexp{Regexp: nil},
			expectedJSON: `null`,
		})
	})

	t.Run("MarshalNilRegexpPointer", func(t *testing.T) {
		var r *Regexp
		checkMarshal(t, testScenario{
			name:         "nil regexp pointer",
			input:        r,
			expectedJSON: `null`,
		})
	})
}

func TestUnitRegexpJSONUnmarshalling(t *testing.T) {
	type testScenario struct {
		name           string
		jsonInput      string
		expectedRegexp *Regexp
		expectedError  string
	}

	checkUnmarshal := func(t *testing.T, s testScenario) {
		t.Helper()
		var r Regexp
		err := json.Unmarshal([]byte(s.jsonInput), &r)

		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			return
		}

		require.NoError(t, err)

		if s.expectedRegexp == nil || s.expectedRegexp.Regexp == nil {
			require.Nil(t, r.Regexp)
		} else {
			require.NotNil(t, r.Regexp)
			require.Equal(t, s.expectedRegexp.String(), r.String())
		}
	}

	t.Run("UnmarshalValidRegexp", func(t *testing.T) {
		checkUnmarshal(t, testScenario{
			name:           "valid regexp string",
			jsonInput:      `"^[a-z]+$"`,
			expectedRegexp: MustCompile(`^[a-z]+$`),
		})
	})

	t.Run("UnmarshalEmptyRegexp", func(t *testing.T) {
		checkUnmarshal(t, testScenario{
			name:           "empty regexp string",
			jsonInput:      `""`,
			expectedRegexp: &Regexp{}, // This will have a nil Regexp field
		})
	})

	t.Run("UnmarshalNull", func(t *testing.T) {
		checkUnmarshal(t, testScenario{
			name:           "json null",
			jsonInput:      `null`,
			expectedRegexp: nil, // expecting nil regexp
		})
	})

	t.Run("UnmarshalInvalidRegexpPattern", func(t *testing.T) {
		checkUnmarshal(t, testScenario{
			name:          "invalid regexp pattern",
			jsonInput:     `"["`,
			expectedError: "error parsing regexp", // From regexp.Compile
		})
	})

	t.Run("UnmarshalInvalidJSON", func(t *testing.T) {
		checkUnmarshal(t, testScenario{
			name:          "invalid json",
			jsonInput:     `not-a-json-string`,
			expectedError: "invalid character",
		})
	})
}
