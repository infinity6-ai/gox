package errorz

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitNew(t *testing.T) {
	err := errors.New("test error")
	wrappedErr := New(err)

	assert.NotNil(t, wrappedErr)
	assert.Equal(t, "test error", wrappedErr.Error())

	customErr, ok := wrappedErr.(*customError)
	assert.True(t, ok)
	assert.NotNil(t, customErr.cause)
	assert.NotEmpty(t, customErr.stack)
}

func TestUnitNewWithNil(t *testing.T) {
	wrappedErr := New(nil)
	assert.Nil(t, wrappedErr)
}

func TestUnitUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := New(originalErr)

	unwrapped := errors.Unwrap(wrappedErr)
	assert.Equal(t, originalErr, unwrapped)
}

func TestUnitStackTrace(t *testing.T) {
	err := errors.New("another test error")
	wrappedErr := New(err)

	customErr, ok := wrappedErr.(*customError)
	assert.True(t, ok)

	stack := customErr.StackTrace()
	assert.NotEmpty(t, stack)
	assert.True(t, strings.Contains(stack, "errorz_test.go"))
	assert.True(t, strings.Contains(stack, "TestUnitStackTrace"))
}
