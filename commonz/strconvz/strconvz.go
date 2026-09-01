package strconvz

import (
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func ParseNumber[T constraintz.Numbers](v string, defaultV ...T) (T, error) {
	var zero T
	if v == "" {
		if len(defaultV) > 0 {
			return defaultV[0], nil
		}
		return zero, fmt.Errorf("empty string with no default value")
	}

	switch any(zero).(type) {
	case float32:
		parsed, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return zero, fmt.Errorf("failed to parse float32 '%s': %w", v, err)
		}
		return T(parsed), nil
	case float64:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return zero, fmt.Errorf("failed to parse float64 '%s': %w", v, err)
		}
		return T(parsed), nil
	case int:
		parsed, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return zero, fmt.Errorf("failed to parse int '%s': %w", v, err)
		}
		return T(parsed), nil
	case int8:
		parsed, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("failed to parse int8 '%s': %w", v, err)
		}
		return T(parsed), nil
	case int16:
		parsed, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("failed to parse int16 '%s': %w", v, err)
		}
		return T(parsed), nil
	case int32:
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("failed to parse int32 '%s': %w", v, err)
		}
		return T(parsed), nil
	case int64:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("failed to parse int64 '%s': %w", v, err)
		}
		return T(parsed), nil
	case uint:
		parsed, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return zero, fmt.Errorf("failed to parse uint '%s': %w", v, err)
		}
		return T(parsed), nil
	case uint8:
		parsed, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("failed to parse uint8 '%s': %w", v, err)
		}
		return T(parsed), nil
	case uint16:
		parsed, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("failed to parse uint16 '%s': %w", v, err)
		}
		return T(parsed), nil
	case uint32:
		parsed, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("failed to parse uint32 '%s': %w", v, err)
		}
		return T(parsed), nil
	case uint64:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("failed to parse uint64 '%s': %w", v, err)
		}
		return T(parsed), nil
	default:
		return zero, fmt.Errorf("unsupported number type: %T", zero)
	}
}

func MustParseNumber[T constraintz.Numbers](v string, defaultV ...T) T {
	value, err := ParseNumber(v, defaultV...)
	errorz.Check(err)
	return value
}

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

func MustParseBool(v string, defaultV ...bool) bool {
	value, err := ParseBool(v, defaultV...)
	errorz.Check(err)
	return value
}
