package errorz_test

import (
	"errors"
	"runtime"
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

func TestUnitUnpanicNoPanic(t *testing.T) {
	err := errorz.Unpanic(func() {})
	assert.Nil(t, err)
}

func TestUnitUnpanicWithStandardError(t *testing.T) {
	stdErr := errors.New("standard error")
	err := errorz.Unpanic(func() {
		panic(stdErr)
	})
	assert.NotNil(t, err)
	assert.Equal(t, "PanicRecoveredError", err.Name())
	unwrappedErr := errors.Unwrap(err)
	assert.Equal(t, stdErr, unwrappedErr)
}

func TestUnitUnpanicWithStructuredError(t *testing.T) {
	structuredErr := errorz.Detail(500, "TestError", "", false, errors.New("test"))
	err := errorz.Unpanic(func() {
		panic(structuredErr)
	})
	assert.NotNil(t, err)
	assert.Equal(t, "PanicRecoveredError", err.Name())
	unwrappedErr := errors.Unwrap(err)
	assert.Equal(t, structuredErr, unwrappedErr)
}

func TestUnitUnpanicWithString(t *testing.T) {
	err := errorz.Unpanic(func() {
		panic("panic string")
	})
	assert.NotNil(t, err)
	assert.Equal(t, "PanicRecoveredError", err.Name())
	unwrappedErr := errors.Unwrap(err)

	pe, ok := unwrappedErr.(*errorz.PanicVal)
	assert.True(t, ok)
	assert.Equal(t, "panic string", pe.Value())
}

func TestUnitUnpanicWithNil(t *testing.T) {
	err := errorz.Unpanic(func() {
		panic(nil)
	})
	assert.NotNil(t, err)
	assert.Equal(t, "PanicRecoveredError", err.Name())
	unwrappedErr := errors.Unwrap(err)

	_, ok := unwrappedErr.(*runtime.PanicNilError)
	assert.True(t, ok, "Expected unwrapped error to be *runtime.PanicNilError")
}
