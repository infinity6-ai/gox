package strconvz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/util/strconvz"
)

func TestUnitParsez(t *testing.T) {
	// Float64
	assert.Equal(t, float64(3.14), strconvz.ParseFloat64("3.14"))
	assert.Equal(t, float64(-42.5), strconvz.ParseFloat64("-42.5"))

	// Float32
	assert.Equal(t, float32(2.71), strconvz.ParseFloat32("2.71"))
	assert.Equal(t, float32(-0.5), strconvz.ParseFloat32("-0.5"))

	// Int64
	assert.Equal(t, int64(42), strconvz.ParseInt64("42"))
	assert.Equal(t, int64(-123456), strconvz.ParseInt64("-123456"))

	// Int32
	assert.Equal(t, int32(123), strconvz.ParseInt32("123"))
	assert.Equal(t, int32(-9999), strconvz.ParseInt32("-9999"))

	// Int
	assert.Equal(t, 7, strconvz.ParseInt("7"))
	assert.Equal(t, -321, strconvz.ParseInt("-321"))

	// Bool
	assert.Equal(t, true, strconvz.ParseBool("true"))
	assert.Equal(t, false, strconvz.ParseBool("false"))
}

func TestUnitParsezDefault(t *testing.T) {
	// Float64Default
	assert.Equal(t, float64(3.14), strconvz.ParseFloat64Default("3.14", 1.0))
	assert.Equal(t, float64(1.23), strconvz.ParseFloat64Default("", 1.23))

	// Float32Default
	assert.Equal(t, float32(2.71), strconvz.ParseFloat32Default("2.71", 1.0))
	assert.Equal(t, float32(4.56), strconvz.ParseFloat32Default("", 4.56))

	// Int64Default
	assert.Equal(t, int64(42), strconvz.ParseInt64Default("42", 100))
	assert.Equal(t, int64(100), strconvz.ParseInt64Default("", 100))

	// Int32Default
	assert.Equal(t, int32(123), strconvz.ParseInt32Default("123", 99))
	assert.Equal(t, int32(99), strconvz.ParseInt32Default("", 99))

	// IntDefault
	assert.Equal(t, 7, strconvz.ParseIntDefault("7", 8))
	assert.Equal(t, 8, strconvz.ParseIntDefault("", 8))

	// BoolDefault
	assert.Equal(t, true, strconvz.ParseBoolDefault("true", false))
	assert.Equal(t, false, strconvz.ParseBoolDefault("false", true))
	assert.Equal(t, true, strconvz.ParseBoolDefault("", true))
	assert.Equal(t, false, strconvz.ParseBoolDefault("", false))
}
