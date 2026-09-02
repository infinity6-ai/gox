package regexpz

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitRuleJSONSerialization(t *testing.T) {
	type testScenario struct {
		name          string
		inputRule     Rule
		expectedJSON  string
		expectedError string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Test MarshalJSON
		marshaled, err := json.Marshal(&s.inputRule)
		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.expectedJSON, string(marshaled))

		// Test UnmarshalJSON
		var unmarshaled Rule
		err = json.Unmarshal(marshaled, &unmarshaled)
		require.NoError(t, err)

		// if we marshal to null, we can't get the original data back
		if string(marshaled) == "null" {
			require.Nil(t, unmarshaled.Regexp)
			return
		}

		require.Equal(t, s.inputRule.Name, unmarshaled.Name)
		require.Equal(t, s.inputRule.Operation, unmarshaled.Operation)
		require.Equal(t, s.inputRule.Out, unmarshaled.Out)
		require.Equal(t, s.inputRule.Regexp.String(), unmarshaled.Regexp.String())
	}

	t.Run("Valid Rule with MatchOperation", func(t *testing.T) {
		check(t, testScenario{
			name: "Valid Rule with MatchOperation",
			inputRule: Rule{
				Name:      "test_match",
				Regexp:    regexp.MustCompile("foo"),
				Operation: MatchOperation,
				Out:       "",
			},
			expectedJSON: `{"name":"test_match","regexp_pattern":"foo","operation":"match","out":""}`,
		})
	})

	t.Run("Valid Rule with ReplaceOperation", func(t *testing.T) {
		check(t, testScenario{
			name: "Valid Rule with ReplaceOperation",
			inputRule: Rule{
				Name:      "test_replace",
				Regexp:    regexp.MustCompile("bar"),
				Operation: ReplaceOperation,
				Out:       "baz",
			},
			expectedJSON: `{"name":"test_replace","regexp_pattern":"bar","operation":"replace","out":"baz"}`,
		})
	})

	t.Run("Rule with empty Regexp", func(t *testing.T) {
		check(t, testScenario{
			name: "Rule with empty Regexp",
			inputRule: Rule{
				Name:      "test_empty",
				Regexp:    regexp.MustCompile(""),
				Operation: MismatchOperation,
				Out:       "",
			},
			expectedJSON: `{"name":"test_empty","regexp_pattern":"","operation":"mismatch","out":""}`,
		})
	})

	t.Run("Rule with nil Regexp", func(t *testing.T) {
		check(t, testScenario{
			name: "Rule with nil Regexp",
			inputRule: Rule{
				Name:      "test_nil_regexp",
				Regexp:    nil, // This will be marshaled as null for the object
				Operation: MatchOperation,
				Out:       "",
			},
			expectedJSON: `null`, // As per MarshalJSON for nil Regexp
		})
	})
}

func TestUnitEngineJSONSerialization(t *testing.T) {
	type testScenario struct {
		name          string
		inputEngine   Engine
		expectedJSON  string
		expectedError string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Test MarshalJSON
		marshaled, err := json.Marshal(&s.inputEngine)
		require.NoError(t, err)
		require.JSONEq(t, s.expectedJSON, string(marshaled))

		// Test UnmarshalJSON
		var unmarshaled Engine
		err = json.Unmarshal(marshaled, &unmarshaled)
		if s.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedError)
			return
		}
		require.NoError(t, err)

		require.Len(t, unmarshaled.Rules, len(s.inputEngine.Rules))
		for i, rule := range s.inputEngine.Rules {
			require.Equal(t, rule.Name, unmarshaled.Rules[i].Name)
			require.Equal(t, rule.Operation, unmarshaled.Rules[i].Operation)
			require.Equal(t, rule.Out, unmarshaled.Rules[i].Out)
			if rule.Regexp != nil {
				require.Equal(t, rule.Regexp.String(), unmarshaled.Rules[i].Regexp.String())
			} else {
				require.Nil(t, unmarshaled.Rules[i].Regexp)
			}
		}
	}

	t.Run("Engine with multiple Rules", func(t *testing.T) {
		check(t, testScenario{
			name: "Engine with multiple Rules",
			inputEngine: Engine{
				Rules: []Rule{
					{
						Name:      "rule1",
						Regexp:    regexp.MustCompile("pattern1"),
						Operation: MatchOperation,
						Out:       "out1",
					},
					{
						Name:      "rule2",
						Regexp:    regexp.MustCompile("pattern2"),
						Operation: ReplaceOperation,
						Out:       "out2",
					},
				},
			},
			expectedJSON: `{"rules":[{"name":"rule1","regexp_pattern":"pattern1","operation":"match","out":"out1"},{"name":"rule2","regexp_pattern":"pattern2","operation":"replace","out":"out2"}]}`,
		})
	})

	t.Run("Engine with an invalid regex pattern in JSON", func(t *testing.T) {
		// This test specifically checks UnmarshalJSON error handling
		invalidJSON := `{"rules":[{"name":"bad_rule","regexp_pattern":"[invalid","operation":"match","out":""}]}`
		var unmarshaled Engine
		err := json.Unmarshal([]byte(invalidJSON), &unmarshaled)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error parsing regexp: missing closing ]")
	})

	t.Run("Engine with a nil Rule in the slice", func(t *testing.T) {
		check(t, testScenario{
			name: "Engine with a nil Rule in the slice",
			inputEngine: Engine{
				Rules: []Rule{
					{
						Name:      "rule1",
						Regexp:    regexp.MustCompile("pattern1"),
						Operation: MatchOperation,
						Out:       "out1",
					},
					{}, // This would have a nil regexp, resulting in a null object for this rule
				},
			},
			expectedJSON: `{"rules":[{"name":"rule1","regexp_pattern":"pattern1","operation":"match","out":"out1"},null]}`,
		})
	})
}
