package errorz_test

import (
	"errors"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitPanicWithNilError(t *testing.T) {
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		errorz.Panic(nil)
	}()

	assert.NotNil(t, recovered)
	se, ok := recovered.(errorz.StructuredError)
	assert.True(t, ok)
	assert.Equal(t, "NulPanic", se.Name())
}

func TestUnitPanicWithStandardError(t *testing.T) {
	stdErr := errors.New("standard error")

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		errorz.Panic(stdErr)
	}()

	assert.NotNil(t, recovered)
	se, ok := recovered.(errorz.StructuredError)
	assert.True(t, ok)
	assert.Equal(t, stdErr, errors.Unwrap(se))
}

func TestUnitPanicWithStructuredError(t *testing.T) {
	structuredErr := errorz.Detail(500, "TestError", "", false, errors.New("test"))

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		errorz.Panic(structuredErr)
	}()

	assert.NotNil(t, recovered)
	se, ok := recovered.(errorz.StructuredError)
	assert.True(t, ok)
	assert.Equal(t, structuredErr, se)
}

func TestUnitCheckWithNilError(t *testing.T) {
	assert.NotPanics(t, func() {
		errorz.Check(nil)
	})
}

func TestUnitCheckWithStandardError(t *testing.T) {
	stdErr := errors.New("standard error")

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		errorz.Check(stdErr)
	}()

	assert.NotNil(t, recovered)
	se, ok := recovered.(errorz.StructuredError)
	assert.True(t, ok)
	assert.Equal(t, stdErr, errors.Unwrap(se))
}

func TestUnitCheckWithStructuredError(t *testing.T) {
	structuredErr := errorz.Detail(500, "TestError", "", false, errors.New("test"))
	assert.PanicsWithValue(t, structuredErr, func() {
		errorz.Check(structuredErr)
	})
}
