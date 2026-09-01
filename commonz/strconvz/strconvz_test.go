package strconvz_test

import (
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/stretchr/testify/require"
)

func TestUnitParseNumber(t *testing.T) {
	// checkFunc is a helper for float64 test scenarios
	checkFloat64 := func(t *testing.T, name, input string, defaultV []float64, want float64, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, float64(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for float32 test scenarios
	checkFloat32 := func(t *testing.T, name, input string, defaultV []float32, want float32, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, float32(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for int test scenarios
	checkInt := func(t *testing.T, name, input string, defaultV []int, want int, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, int(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for int64 test scenarios
	checkInt64 := func(t *testing.T, name, input string, defaultV []int64, want int64, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, int64(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for uint test scenarios
	checkUint := func(t *testing.T, name, input string, defaultV []uint, want uint, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, uint(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for uint64 test scenarios
	checkUint64 := func(t *testing.T, name, input string, defaultV []uint64, want uint64, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, uint64(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for int8 test scenarios
	checkInt8 := func(t *testing.T, name, input string, defaultV []int8, want int8, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, int8(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	// checkFunc is a helper for uint8 test scenarios
	checkUint8 := func(t *testing.T, name, input string, defaultV []uint8, want uint8, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, uint8(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	t.Run("float64 parsing", func(t *testing.T) {
		checkFloat64(t, "Valid float64", "3.14", nil, float64(3.14), "")
		checkFloat64(t, "Negative float64", "-42.5", nil, float64(-42.5), "")
		checkFloat64(t, "Empty string with default float64", "", []float64{99.9}, float64(99.9), "")
		checkFloat64(t, "Invalid float64 string", "abc", nil, float64(0), "failed to parse float64 'abc'")
		checkFloat64(t, "Empty string no default float64", "", nil, float64(0), "empty string with no default value")
	})

	t.Run("float32 parsing", func(t *testing.T) {
		checkFloat32(t, "Valid float32", "2.71", nil, float32(2.71), "")
		checkFloat32(t, "Negative float32", "-0.5", nil, float32(-0.5), "")
		checkFloat32(t, "Empty string with default float32", "", []float32{1.23}, float32(1.23), "")
		checkFloat32(t, "Invalid float32 string", "xyz", nil, float32(0), "failed to parse float32 'xyz'")
		checkFloat32(t, "Empty string no default float32", "", nil, float32(0), "empty string with no default value")
	})

	t.Run("int parsing", func(t *testing.T) {
		checkInt(t, "Valid int", "42", nil, int(42), "")
		checkInt(t, "Negative int", "-123", nil, int(-123), "")
		checkInt(t, "Empty string with default int", "", []int{100}, int(100), "")
		checkInt(t, "Invalid int string", "abc", nil, int(0), "failed to parse int 'abc'")
		checkInt(t, "Empty string no default int", "", nil, int(0), "empty string with no default value")
		checkInt(t, "Overflow int", "92233720368547758071", nil, int(0), "value out of range") // Assuming 64-bit int
	})

	t.Run("int64 parsing", func(t *testing.T) {
		checkInt64(t, "Valid int64", "1234567890", nil, int64(1234567890), "")
		checkInt64(t, "Negative int64", "-9876543210", nil, int64(-9876543210), "")
		checkInt64(t, "Empty string with default int64", "", []int64{500}, int64(500), "")
		checkInt64(t, "Invalid int64 string", "abc", nil, int64(0), "failed to parse int64 'abc'")
		checkInt64(t, "Empty string no default int64", "", nil, int64(0), "empty string with no default value")
		checkInt64(t, "Overflow int64", "92233720368547758071", nil, int64(0), "value out of range")
	})

	t.Run("uint parsing", func(t *testing.T) {
		checkUint(t, "Valid uint", "42", nil, uint(42), "")
		checkUint(t, "Empty string with default uint", "", []uint{100}, uint(100), "")
		checkUint(t, "Invalid uint string", "abc", nil, uint(0), "failed to parse uint 'abc'")
		checkUint(t, "Negative uint string", "-10", nil, uint(0), "invalid syntax")
		checkUint(t, "Empty string no default uint", "", nil, uint(0), "empty string with no default value")
		checkUint(t, "Overflow uint", "184467440737095516151", nil, uint(0), "value out of range") // Assuming 64-bit uint
	})

	t.Run("uint64 parsing", func(t *testing.T) {
		checkUint64(t, "Valid uint64", "1234567890", nil, uint64(1234567890), "")
		checkUint64(t, "Empty string with default uint64", "", []uint64{500}, uint64(500), "")
		checkUint64(t, "Invalid uint64 string", "abc", nil, uint64(0), "failed to parse uint64 'abc'")
		checkUint64(t, "Negative uint64 string", "-10", nil, uint64(0), "invalid syntax")
		checkUint64(t, "Empty string no default uint64", "", nil, uint64(0), "empty string with no default value")
		checkUint64(t, "Overflow uint64", "184467440737095516151", nil, uint64(0), "value out of range")
	})

	t.Run("int8 parsing", func(t *testing.T) {
		checkInt8(t, "Valid int8", "127", nil, int8(127), "")
		checkInt8(t, "Negative int8", "-128", nil, int8(-128), "")
		checkInt8(t, "Overflow int8", "128", nil, int8(0), "value out of range")
		checkInt8(t, "Underflow int8", "-129", nil, int8(0), "value out of range")
	})

	t.Run("uint8 parsing", func(t *testing.T) {
		checkUint8(t, "Valid uint8", "255", nil, uint8(255), "")
		checkUint8(t, "Overflow uint8", "256", nil, uint8(0), "value out of range")
		checkUint8(t, "Negative uint8", "-1", nil, uint8(0), "invalid syntax")
	})

	// Add more integer types like int16, int32, uint16, uint32 as needed
	checkInt16 := func(t *testing.T, name, input string, defaultV []int16, want int16, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, int16(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	t.Run("int16 parsing", func(t *testing.T) {
		checkInt16(t, "Valid int16", "32767", nil, int16(32767), "")
		checkInt16(t, "Negative int16", "-32768", nil, int16(-32768), "")
		checkInt16(t, "Overflow int16", "32768", nil, int16(0), "value out of range")
		checkInt16(t, "Underflow int16", "-32769", nil, int16(0), "value out of range")
	})

	checkInt32 := func(t *testing.T, name, input string, defaultV []int32, want int32, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, int32(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	t.Run("int32 parsing", func(t *testing.T) {
		checkInt32(t, "Valid int32", "2147483647", nil, int32(2147483647), "")
		checkInt32(t, "Negative int32", "-2147483648", nil, int32(-2147483648), "")
		checkInt32(t, "Overflow int32", "2147483648", nil, int32(0), "value out of range")
		checkInt32(t, "Underflow int32", "-2147483649", nil, int32(0), "value out of range")
	})

	checkUint16 := func(t *testing.T, name, input string, defaultV []uint16, want uint16, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, uint16(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	t.Run("uint16 parsing", func(t *testing.T) {
		checkUint16(t, "Valid uint16", "65535", nil, uint16(65535), "")
		checkUint16(t, "Overflow uint16", "65536", nil, uint16(0), "value out of range")
		checkUint16(t, "Negative uint16", "-1", nil, uint16(0), "invalid syntax")
	})

	checkUint32 := func(t *testing.T, name, input string, defaultV []uint32, want uint32, wantErr string) {
		t.Helper()
		result, err := strconvz.ParseNumber(input, defaultV...)
		if wantErr != "" {
			require.Error(t, err, fmt.Sprintf("Expected error for %s", name))
			require.Contains(t, err.Error(), wantErr, fmt.Sprintf("Error message mismatch for %s", name))
			require.Equal(t, uint32(0), result, fmt.Sprintf("Expected zero value on error for %s", name))
		} else {
			require.NoError(t, err, fmt.Sprintf("Did not expect error for %s", name))
			require.Equal(t, want, result, fmt.Sprintf("Result mismatch for %s", name))
		}
	}

	t.Run("uint32 parsing", func(t *testing.T) {
		checkUint32(t, "Valid uint32", "4294967295", nil, uint32(4294967295), "")
		checkUint32(t, "Overflow uint32", "4294967296", nil, uint32(0), "value out of range")
		checkUint32(t, "Negative uint32", "-1", nil, uint32(0), "invalid syntax")
	})
}
