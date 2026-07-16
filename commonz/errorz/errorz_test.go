package errorz_test

import (
	"errors"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/stretchr/testify/assert"
)

func TestUnitNew(t *testing.T) {
	err := errors.New("test error")
	wrappedErr := errorz.New(err)

	assert.NotNil(t, wrappedErr)
	assert.Equal(t, "test error", wrappedErr.Error())
}

func TestUnitNewWithNil(t *testing.T) {
	wrappedErr := errorz.New(nil)
	assert.Nil(t, wrappedErr)
}

func TestUnitUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := errorz.New(originalErr)

	unwrapped := errors.Unwrap(wrappedErr)
	assert.Equal(t, originalErr, unwrapped)
}
