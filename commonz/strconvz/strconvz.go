package strconvz

import (
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

// ParseNumber parses a string into a numeric type T.
// T can be any of the supported number types defined in constraintz.Numbers.
//
// If the input string v is empty, it returns the provided default value.
// If no default value is given for an empty string, it returns an error.
//
// The function handles various integer and floating-point types,
// using the appropriate bit size for parsing.
//
// It returns the parsed value of type T and an error if parsing fails
// or if the type is not supported.
func ParseNumber[T constraintz.Numbers](v string, defaultV ...T) (T, error) {
	var zero T
	if v == "" {
		if len(defaultV) > 0 {
			return defaultV[0], nil
		}
		return zero, fmt.Errorf("empty string with no default value")
	}

	var err error

	switch any(zero).(type) {
	case float32:
		f, e := strconv.ParseFloat(v, 32)
		if e == nil {
			return T(f), nil
		}
		err = e
	case float64:
		f, e := strconv.ParseFloat(v, 64)
		if e == nil {
			return T(f), nil
		}
		err = e
	case int, int8, int16, int32, int64:
		bitSize := getIntegerBitSize(any(zero))
		parsed, e := strconv.ParseInt(v, 10, bitSize)
		if e == nil {
			return T(parsed), nil
		}
		err = e
	case uint, uint8, uint16, uint32, uint64:
		bitSize := getIntegerBitSize(any(zero))
		parsed, e := strconv.ParseUint(v, 10, bitSize)
		if e == nil {
			return T(parsed), nil
		}
		err = e
	default:
		return zero, fmt.Errorf("unsupported number type: %T", zero)
	}

	if err != nil {
		return zero, fmt.Errorf("failed to parse %T '%s': %w", zero, v, err)
	}

	return zero, fmt.Errorf("unexpected error in ParseNumber for type %T", zero)
}

// MustParseNumber is like ParseNumber but panics if an error occurs.
// This is useful for cases where a parsing failure should be a fatal error.
func MustParseNumber[T constraintz.Numbers](v string, defaultV ...T) T {
	value, err := ParseNumber(v, defaultV...)
	errorz.Check(err)
	return value
}

// ParseBool parses a string into a boolean value.
// It accepts standard boolean strings like "true", "false", "1", "0", etc.
// If the input string v is empty, it returns the provided default value.
// If no default value is given for an empty string, it returns an error.
func ParseBool(v string, defaultV ...bool) (bool, error) {
	if v == "" {
		if len(defaultV) > 0 {
			return defaultV[0], nil
		}
		return false, fmt.Errorf("empty string with no default value")
	}

	value, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("could not parse bool: %s, %w", v, err)
	}
	return value, nil
}

// MustParseBool is like ParseBool but panics if an error occurs.
// This is useful for cases where a parsing failure should be a fatal error.
func MustParseBool(v string, defaultV ...bool) bool {
	value, err := ParseBool(v, defaultV...)
	errorz.Check(err)
	return value
}

// getIntegerBitSize returns the bit size for any integer type.
// It returns 0 for int and uint (system-dependent bit size).
func getIntegerBitSize(val interface{}) int {
	switch val.(type) {
	case int, uint:
		return 0
	case int8, uint8:
		return 8
	case int16, uint16:
		return 16
	case int32, uint32:
		return 32
	case int64, uint64:
		return 64
	default:
		return 0 // Should not happen if called with a non-integer type
	}
}
