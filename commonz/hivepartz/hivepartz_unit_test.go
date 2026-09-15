package hivepartz

import (
	"encoding/json"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitHivePartsAddGetOptionalNames(t *testing.T) {
	type testScenario struct {
		name       string
		value      string
		expectedHP *HiveParts
		expectErr  bool
		errMsg     string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		hp := New()
		err := hp.Add(s.name, s.value)

		if s.expectErr {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expectedHP.names, hp.names)
		require.Equal(t, s.expectedHP.values, hp.values)
		require.Equal(t, s.value, hp.Get(s.name))
		require.True(t, hp.Optional(s.name).IsPresent())
		require.Equal(t, s.value, hp.Optional(s.name).Must())
		require.Equal(t, []string{s.name}, hp.Names())
	}

	t.Run("Valid add operation", func(t *testing.T) {
		check(t, testScenario{
			name:  "key",
			value: "value",
			expectedHP: &HiveParts{
				names:  []string{"key"},
				values: map[string]string{"key": "value"},
			},
			expectErr: false,
		})
	})

	t.Run("Add with invalid name", func(t *testing.T) {
		check(t, testScenario{
			name:      "Key", // Capital letter invalidates
			value:     "value",
			expectErr: true,
			errMsg:    "validation error validation fail StringRegex",
		})
	})

	t.Run("Add with invalid value", func(t *testing.T) {
		check(t, testScenario{
			name:      "key",
			value:     "_value", // Invalid start char
			expectErr: true,
			errMsg:    "validation error validation fail StringRegex",
		})
	})

	t.Run("Add with name starting with number", func(t *testing.T) {
		check(t, testScenario{
			name:      "1key",
			value:     "value",
			expectErr: true,
			errMsg:    "validation error validation fail StringRegex",
		})
	})

	t.Run("Add with value starting with special char", func(t *testing.T) {
		check(t, testScenario{
			name:      "key",
			value:     "-value",
			expectErr: true,
			errMsg:    "validation error validation fail StringRegex",
		})
	})
}

func TestUnitHivePartsFormatString(t *testing.T) {
	type testScenario struct {
		hp        *HiveParts
		want      string
		expectErr bool
		errMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := s.hp.FormatString()
		require.Equal(t, s.want, got)
	}

	t.Run("Single part", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		check(t, testScenario{
			hp:   hp,
			want: "region=us-east-1",
		})
	})

	t.Run("Multiple parts", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		hp.MustAdd("env", "prod")
		check(t, testScenario{
			hp:   hp,
			want: "region=us-east-1/env=prod",
		})
	})

	t.Run("No parts", func(t *testing.T) {
		hp := New()
		check(t, testScenario{
			hp:   hp,
			want: "",
		})
	})

	t.Run("Parts with hyphens", func(t *testing.T) {
		hp := New()
		hp.MustAdd("data-center", "dc-1")
		check(t, testScenario{
			hp:   hp,
			want: "data-center=dc-1",
		})
	})
}

func TestUnitHivePartsFormatStringAndString(t *testing.T) {
	type testScenario struct {
		hp   *HiveParts
		want string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		require.Equal(t, s.want, s.hp.FormatString())
		require.Equal(t, s.want, s.hp.String())
	}

	t.Run("Single part", func(t *testing.T) {
		hp := New()
		hp.MustAdd("key1", "value1")
		check(t, testScenario{hp: hp, want: "key1=value1"})
	})

	t.Run("Multiple parts", func(t *testing.T) {
		hp := New()
		hp.MustAdd("key1", "value1")
		hp.MustAdd("key2", "value2")
		check(t, testScenario{hp: hp, want: "key1=value1/key2=value2"})
	})

	t.Run("Empty hive parts", func(t *testing.T) {
		hp := New()
		check(t, testScenario{hp: hp, want: ""})
	})
}

func TestUnitHivePartsFormat(t *testing.T) {
	type testScenario struct {
		hp   *HiveParts
		want *pathz.Path
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := s.hp.Format()
		require.Equal(t, s.want.String(), got.String())
	}

	t.Run("Single part", func(t *testing.T) {
		hp := New()
		hp.MustAdd("key", "value")
		check(t, testScenario{hp: hp, want: pathz.MustParse("key=value")})
	})

	t.Run("Multiple parts", func(t *testing.T) {
		hp := New()
		hp.MustAdd("key1", "value1")
		hp.MustAdd("key2", "value2")
		check(t, testScenario{hp: hp, want: pathz.MustParse("key1=value1/key2=value2")})
	})

	t.Run("No parts", func(t *testing.T) {
		hp := New()
		check(t, testScenario{hp: hp, want: pathz.MustParse("")})
	})
}

func TestUnitParseString(t *testing.T) {
	type testScenario struct {
		input     string
		wantHP    *HiveParts
		wantPath  *pathz.Path
		expectErr bool
		errMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		gotHP, gotPath, err := ParseString(s.input)

		if s.expectErr {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.wantHP.names, gotHP.names)
		require.Equal(t, s.wantHP.values, gotHP.values)
		require.Equal(t, s.wantPath.String(), gotPath.String())
	}

	t.Run("Single hive part", func(t *testing.T) {
		check(t, testScenario{
			input:    "region=us-east-1",
			wantHP:   From("region", "us-east-1"),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Multiple hive parts", func(t *testing.T) {
		check(t, testScenario{
			input:    "region=us-east-1/env=prod",
			wantHP:   From("region", "us-east-1", "env", "prod"),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Hive parts with remaining path", func(t *testing.T) {
		check(t, testScenario{
			input:     "region=us-east-1/data/file.txt",
			wantHP:    From("region", "us-east-1"), // Still want to check what it would parse if it didn't error
			wantPath:  pathz.MustParse("data/file.txt"),
			expectErr: true,
			errMsg:    "path unsupported: region=us-east-1/data/file.txt, remaning: data/file.txt",
		})
	})

	t.Run("Empty string", func(t *testing.T) {
		check(t, testScenario{
			input:    "",
			wantHP:   New(),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Only remaining path", func(t *testing.T) {
		check(t, testScenario{
			input:     "data/file.txt",
			wantHP:    &HiveParts{names: []string{}, values: map[string]string{}},
			wantPath:  pathz.MustParse("data/file.txt"),
			expectErr: true,
			errMsg:    "path unsupported: data/file.txt, remaning: data/file.txt",
		})
	})

	t.Run("Invalid path format", func(t *testing.T) {
		check(t, testScenario{
			input:     "region==us-east-1",
			expectErr: true,
			errMsg:    "too many parts in hive partition: region==us-east-1",
		})
	})

	t.Run("Invalid hive part name", func(t *testing.T) {
		check(t, testScenario{
			input:     "Region=us-east-1/data",
			expectErr: true,
			errMsg:    "invalid hive part: validation error validation fail StringRegex (regex=^[a-z][a-z0-9\\-]*, actual=Region): name",
		})
	})

	t.Run("Path with ending slash", func(t *testing.T) {
		check(t, testScenario{
			input:     "region=us-east-1/",
			expectErr: true,
			errMsg:    "path unsupported: cannot end with slash",
		})
	})

	t.Run("Path with parent directory", func(t *testing.T) {
		check(t, testScenario{
			input:     "../region=us-east-1",
			expectErr: true,
			errMsg:    "path unsupported: validation error must be less or equal than (threshold=0, actual=1): max parents allowed: ../region=us-east-1",
		})
	})
}

func TestUnitParse(t *testing.T) {
	type testScenario struct {
		input     *pathz.Path
		wantHP    *HiveParts
		wantPath  *pathz.Path
		expectErr bool
		errMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		gotHP, gotPath, err := Parse(s.input)

		if s.expectErr {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.wantHP.names, gotHP.names)
		require.Equal(t, s.wantHP.values, gotHP.values)
		require.Equal(t, s.wantPath.String(), gotPath.String())
	}

	t.Run("Single hive part", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse("region=us-east-1"),
			wantHP:   From("region", "us-east-1"),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Multiple hive parts", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse("region=us-east-1/env=prod"),
			wantHP:   From("region", "us-east-1", "env", "prod"),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Hive parts with remaining path", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse("region=us-east-1/data/file.txt"),
			wantHP:   From("region", "us-east-1"),
			wantPath: pathz.MustParse("data/file.txt"),
		})
	})

	t.Run("Empty path", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse(""),
			wantHP:   New(),
			wantPath: pathz.MustParse(""),
		})
	})

	t.Run("Only remaining path", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse("data/file.txt"),
			wantHP:   &HiveParts{names: []string{}, values: map[string]string{}},
			wantPath: pathz.MustParse("data/file.txt"),
		})
	})

	t.Run("Path with ending slash is unsupported", func(t *testing.T) {
		check(t, testScenario{
			input:     pathz.MustParse("region=us-east-1/"),
			expectErr: true,
			errMsg:    "path unsupported: cannot end with slash",
		})
	})
	t.Run("Path with parent directory is unsupported", func(t *testing.T) {
		check(t, testScenario{
			input:     pathz.MustParse("../region=us-east-1"),
			expectErr: true,
			errMsg:    "path unsupported: validation error must be less or equal than (threshold=0, actual=1): max parents allowed: ../region=us-east-1",
		})
	})
}

func TestUnitMustParse(t *testing.T) {
	type testScenario struct {
		input     *pathz.Path
		wantHP    *HiveParts
		wantPath  *pathz.Path
		expectErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.expectErr {
			require.Panics(t, func() {
				MustParse(s.input)
			})
			return
		}

		gotHP, gotPath := MustParse(s.input)
		require.Equal(t, s.wantHP.names, gotHP.names)
		require.Equal(t, s.wantHP.values, gotHP.values)
		require.Equal(t, s.wantPath.String(), gotPath.String())
	}

	t.Run("Valid parse", func(t *testing.T) {
		check(t, testScenario{
			input:    pathz.MustParse("key=val/data.txt"),
			wantHP:   From("key", "val"),
			wantPath: pathz.MustParse("data.txt"),
		})
	})

	t.Run("Invalid parse panics", func(t *testing.T) {
		check(t, testScenario{
			input:     pathz.MustParse("../key=val"),
			expectErr: true,
		})
	})
}

func TestUnitHivePartsGobEncoding(t *testing.T) {
	type testScenario struct {
		hp        *HiveParts
		expectErr bool
		errMsg    string
	}

	checkEncodeDecode := func(t *testing.T, s testScenario) {
		t.Helper()
		encoded, err := s.hp.GobEncode()
		if s.expectErr {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)

		var decodedHP HiveParts
		err = decodedHP.GobDecode(encoded)
		if s.expectErr { // Check decode error if encode was successful
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.hp.names, decodedHP.names)
		require.Equal(t, s.hp.values, decodedHP.values)
	}

	t.Run("Encode/Decode single part", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		checkEncodeDecode(t, testScenario{hp: hp})
	})

	t.Run("Encode/Decode multiple parts", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		hp.MustAdd("env", "prod")
		checkEncodeDecode(t, testScenario{hp: hp})
	})

	t.Run("Encode/Decode empty", func(t *testing.T) {
		hp := New()
		checkEncodeDecode(t, testScenario{hp: hp})
	})
}

func TestUnitHivePartsJsonEncoding(t *testing.T) {
	type testScenario struct {
		hp        *HiveParts
		expectErr bool
		errMsg    string
	}

	checkMarshalUnmarshal := func(t *testing.T, s testScenario) {
		t.Helper()
		marshalled, err := json.Marshal(s.hp)
		if s.expectErr {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)

		var unmarshalledHP HiveParts
		err = json.Unmarshal(marshalled, &unmarshalledHP)
		if s.expectErr { // Check unmarshal error if marshal was successful
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.hp.names, unmarshalledHP.names)
		require.Equal(t, s.hp.values, unmarshalledHP.values)
	}

	t.Run("Marshal/Unmarshal single part", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		checkMarshalUnmarshal(t, testScenario{hp: hp})
	})

	t.Run("Marshal/Unmarshal multiple parts", func(t *testing.T) {
		hp := New()
		hp.MustAdd("region", "us-east-1")
		hp.MustAdd("env", "prod")
		checkMarshalUnmarshal(t, testScenario{hp: hp})
	})

	t.Run("Marshal/Unmarshal empty", func(t *testing.T) {
		hp := New()
		checkMarshalUnmarshal(t, testScenario{hp: hp})
	})
}

func TestUnitNew(t *testing.T) {
	hp := New()
	require.NotNil(t, hp)
	require.Empty(t, hp.names)
	require.Nil(t, hp.values) // values map is initialized on first Add
}

func TestUnitFrom(t *testing.T) {
	type testScenario struct {
		args      []string
		wantHP    *HiveParts
		expectPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.expectPanic {
			require.Panics(t, func() {
				From(s.args...)
			})
			return
		}

		gotHP := From(s.args...)
		require.Equal(t, s.wantHP.names, gotHP.names)
		require.Equal(t, s.wantHP.values, gotHP.values)
	}

	t.Run("From with valid arguments", func(t *testing.T) {
		check(t, testScenario{
			args:   []string{"key1", "value1", "key2", "value2"},
			wantHP: &HiveParts{names: []string{"key1", "key2"}, values: map[string]string{"key1": "value1", "key2": "value2"}},
		})
	})

	t.Run("From with single argument panics", func(t *testing.T) {
		check(t, testScenario{
			args:        []string{"key1"},
			expectPanic: true,
		})
	})

	t.Run("From with odd number of arguments panics", func(t *testing.T) {
		check(t, testScenario{
			args:        []string{"key1", "value1", "key2"},
			expectPanic: true,
		})
	})

	t.Run("From with no arguments", func(t *testing.T) {
		check(t, testScenario{
			args:   []string{},
			wantHP: New(),
		})
	})
}
