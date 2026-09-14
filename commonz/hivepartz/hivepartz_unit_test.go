package hivepartz

import (
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		name           string
		path           *pathz.Path
		expectedHP     *HiveParts
		expectedPath   *pathz.Path
		expectedErrMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		hp, p, err := Parse(s.path)

		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			require.Nil(t, hp)
			require.Nil(t, p)
			return
		}

		require.NoError(t, err)
		if s.expectedHP == nil {
			require.Nil(t, hp)
		} else {
			require.NotNil(t, hp)
			require.Equal(t, s.expectedHP.names, hp.names)
			require.Equal(t, s.expectedHP.values, hp.values)
		}

		if s.expectedPath == nil {
			require.Nil(t, p)
		} else {
			require.NotNil(t, p)
			require.Equal(t, s.expectedPath.Equals(p), true)
		}
	}

	t.Run("Empty path", func(t *testing.T) {
		check(t, testScenario{
			path:         pathz.New(0, nil, false),
			expectedHP:   &HiveParts{values: nil},
			expectedPath: pathz.New(0, nil, false),
		})
	})

	t.Run("Path with no hive parts", func(t *testing.T) {
		check(t, testScenario{
			path:         pathz.New(0, []string{"a", "b", "c"}, false),
			expectedHP:   &HiveParts{values: nil},
			expectedPath: pathz.New(0, []string{"a", "b", "c"}, false),
		})
	})

	t.Run("Path with only hive parts", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"year=2024", "month=09"}, false),
			expectedHP: &HiveParts{
				names:  []string{"year", "month"},
				values: map[string]string{"year": "2024", "month": "09"},
			},
			expectedPath: pathz.New(0, []string{}, false),
		})
	})

	t.Run("Path with mixed parts", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"year=2024", "data", "file.txt"}, false),
			expectedHP: &HiveParts{
				names:  []string{"year"},
				values: map[string]string{"year": "2024"},
			},
			expectedPath: pathz.New(0, []string{"data", "file.txt"}, false),
		})
	})

	t.Run("Path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"year=2024", "month=09"}, true),
			expectedHP: &HiveParts{
				names:  []string{"year", "month"},
				values: map[string]string{"year": "2024", "month": "09"},
			},
			expectedPath: pathz.New(0, []string{}, true),
		})
	})

	t.Run("Path with empty value", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"key="}, false),
			expectedHP: &HiveParts{
				names:  []string{"key"},
				values: map[string]string{"key": ""},
			},
			expectedPath: pathz.New(0, []string{}, false),
		})
	})

	t.Run("Path with empty key fails in Add", func(t *testing.T) {
		// This does not return an error from Parse, but panics from checker.
		// Let's test MustParse for this.
		require.Panics(t, func() {
			MustParse(pathz.New(0, []string{"=value"}, false))
		})
	})

	t.Run("Invalid hive part with too many equals", func(t *testing.T) {
		check(t, testScenario{
			path:           pathz.New(0, []string{"a=b=c"}, false),
			expectedErrMsg: "too many parts in hive partition: a=b=c",
		})
	})

	t.Run("Absolute path is unsupported", func(t *testing.T) {
		check(t, testScenario{
			path:           pathz.MustParse("/a/b"),
			expectedErrMsg: "path unsupported",
		})
	})

	t.Run("Path with parent traversal is unsupported", func(t *testing.T) {
		check(t, testScenario{
			path:           pathz.New(2, nil, false),
			expectedErrMsg: "path unsupported",
		})
	})

	t.Run("Path with wildcard is supported", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"year=*", "data"}, false),
			expectedHP: &HiveParts{
				names:  []string{"year"},
				values: map[string]string{"year": "*"},
			},
			expectedPath: pathz.New(0, []string{"data"}, false),
		})
	})
	
	t.Run("Complex path", func(t *testing.T) {
		check(t, testScenario{
			path: pathz.New(0, []string{"source=s1", "dataset=d1", "date=2026-09-14", "files", "archive.zip"}, true),
			expectedHP: &HiveParts{
				names:  []string{"source", "dataset", "date"},
				values: map[string]string{"source": "s1", "dataset": "d1", "date": "2026-09-14"},
			},
			expectedPath: pathz.New(0, []string{"files", "archive.zip"}, true),
		})
	})
}

func TestUnitHivePartsFormatAndString(t *testing.T) {
	type testScenario struct {
		name       string
		hiveParts  *HiveParts
		expected   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		require.Equal(t, s.expected, s.hiveParts.String())
		require.Equal(t, s.expected, fmt.Sprintf("%s", s.hiveParts))
		require.Equal(t, s.expected, fmt.Sprintf("%v", s.hiveParts))
	}

	t.Run("Empty HiveParts", func(t *testing.T) {
		check(t, testScenario{
			name:       "empty",
			hiveParts:  &HiveParts{},
			expected:   "",
		})
	})

	t.Run("Single hive part", func(t *testing.T) {
		check(t, testScenario{
			name:       "single",
			hiveParts:  (&HiveParts{}).Add("year", "2024"),
			expected:   "year=2024",
		})
	})

	t.Run("Multiple hive parts", func(t *testing.T) {
		check(t, testScenario{
			name:       "multiple",
			hiveParts:  (&HiveParts{}).Add("year", "2024").Add("month", "09"),
			expected:   "year=2024/month=09",
		})
	})

	t.Run("Hive part with empty value", func(t *testing.T) {
		check(t, testScenario{
			name:       "empty_value",
			hiveParts:  (&HiveParts{}).Add("key", ""),
			expected:   "key=",
		})
	})

	t.Run("Hive part with special characters", func(t *testing.T) {
		check(t, testScenario{
			name:       "special_chars",
			hiveParts:  (&HiveParts{}).Add("name", "John Doe/S.A."),
			expected:   "name=John Doe/S.A.",
		})
	})
}

