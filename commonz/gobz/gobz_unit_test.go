package gobz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name string
	Age  int
}

func TestUnitFormatAndParse(t *testing.T) {
	type testScenario struct {
		name  string
		input any
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Format the input
		encoded, err := Format(s.input)
		require.NoError(t, err)
		require.NotEmpty(t, encoded)

		// Create a new instance of the same type to parse into
		var decoded any
		var expected any
		switch v := s.input.(type) {
		case testStruct:
			decoded = &testStruct{}
			expected = v
		case *testStruct:
			decoded = &testStruct{}
			expected = *v
		case int:
			var i int
			decoded = &i
			expected = v
		case string:
			var s string
			decoded = &s
			expected = v
		default:
			t.Fatalf("unhandled type for decoding: %T", s.input)
		}

		// Parse the encoded data
		err = Parse(encoded, decoded)
		require.NoError(t, err)

		// Dereference pointer for comparison
		var finalDecoded any
		switch v := decoded.(type) {
		case *testStruct:
			finalDecoded = *v
		case *int:
			finalDecoded = *v
		case *string:
			finalDecoded = *v
		default:
			finalDecoded = v // Should not happen with the switch above
		}

		require.Equal(t, expected, finalDecoded)
	}

	t.Run("Struct round-trip", func(t *testing.T) {
		check(t, testScenario{
			name:  "struct value",
			input: testStruct{Name: "John", Age: 30},
		})
	})

	t.Run("Pointer to struct round-trip", func(t *testing.T) {
		check(t, testScenario{
			name:  "pointer to struct",
			input: &testStruct{Name: "Jane", Age: 25},
		})
	})

	t.Run("Integer round-trip", func(t *testing.T) {
		check(t, testScenario{
			name:  "integer",
			input: 42,
		})
	})

	t.Run("String round-trip", func(t *testing.T) {
		check(t, testScenario{
			name:  "string",
			input: "hello, gob",
		})
	})
}

func TestUnitParseErrors(t *testing.T) {
	t.Run("Invalid gob data", func(t *testing.T) {
		invalidData := []byte("this is not gob data")
		var target testStruct
		err := Parse(invalidData, &target)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to gob decode")
	})

	t.Run("Parse into non-pointer", func(t *testing.T) {
		encoded, err := Format(testStruct{Name: "test", Age: 1})
		require.NoError(t, err)

		var target testStruct
		err = Parse(encoded, target) // Passing value, not pointer
		require.Error(t, err)
		require.Contains(t, err.Error(), "gob: attempt to decode into a non-pointer")
	})
}

func TestUnitFormatErrors(t *testing.T) {
	t.Run("Unsupported type", func(t *testing.T) {
		// Functions and channels are not supported by gob
		unsupported := make(chan int)
		_, err := Format(unsupported)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to gob encode")
		require.Contains(t, err.Error(), "can't handle type")
	})
}
