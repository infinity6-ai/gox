package cryptzsalt_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/cryptz/cryptzsalt"
)

func TestUnitSalt(t *testing.T) {
	salted, err := cryptzsalt.Generate("abc")
	assert.NoError(t, err)
	assert.NotEmpty(t, salted.Salt)
	assert.NotEmpty(t, salted.Result)
	assert.NotEqual(t, []byte("abc"), salted.Result)

	other, err := cryptzsalt.Generate("abc")
	assert.NoError(t, err)
	assert.NotEmpty(t, other.Salt)
	assert.NotEmpty(t, other.Result)
	assert.NotEqual(t, []byte("abc"), other.Result)

	assert.NotEqual(t, salted.Salt, other.Salt)
	assert.NotEqual(t, salted.Result, other.Result)

	bundle := salted.Format()
	assert.NotEmpty(t, bundle)

	expected, err := cryptzsalt.Verify(bundle, "abc")
	assert.NoError(t, err)
	assert.Equal(t, salted.Salt, expected.Salt)
	assert.Equal(t, salted.Result, expected.Result)
	assert.Equal(t, 50, expected.BundleSize())

	greaterBundle := append(slices.Clone(bundle), []byte("defgh")...)
	expected, err = cryptzsalt.Verify(greaterBundle, "abc")
	assert.NoError(t, err)
	assert.Equal(t, salted.Salt, expected.Salt)
	assert.Equal(t, salted.Result, expected.Result)
	assert.Equal(t, 50, expected.BundleSize())
	assert.Equal(t, "defgh", string(greaterBundle[expected.BundleSize():]))

	_, err = cryptzsalt.Verify(bundle, "abc2")
	assert.Error(t, err)
	assert.EqualError(t, err, "hash mismatch")

	_, err = cryptzsalt.Generate(make([]byte, cryptzsalt.MaxSize+1))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original data size")
	assert.Contains(t, err.Error(), "exceeds max size")
}
