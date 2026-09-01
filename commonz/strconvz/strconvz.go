package strconvz

import (
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/commonz/constraintz"
)

func ParseNumber[T constraintz.Numbers](v string, defaultV ...T) (T, error) {
	if v == "" && len(defaultV) > 0 {
		return defaultV[0], nil
	}

	//float64, float32, int64, int32, int
	value := new(T)

	// value, err := strconv.ParseFloat(v, 64)
	// if err != nil {
	// 	return 0, fmt.Errorf("could not parse float64: %s, %w", v, err)
	// }
	return *value, nil
}

func ParseBool(v string, defaultV ...bool) (bool, error) {
	if v == "" && len(defaultV) > 0 {
		return defaultV[0], nil
	}

	value, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("could not parse bool: %s, %w", v, err)
	}
	return value, nil
}
