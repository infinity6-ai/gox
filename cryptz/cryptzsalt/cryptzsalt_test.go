package cryptzsalt_test

import (
	"slices"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzsalt"
	"github.com/stretchr/testify/require"
)

func TestUnitSalt(t *testing.T) {
	salted, err := cryptzsalt.Generate("abc")
	require.NoError(t, err)
	require.NotEmpty(t, salted.Salt)
	require.NotEmpty(t, salted.Result)
	require.NotEqual(t, []byte("abc"), salted.Result)

	other, err := cryptzsalt.Generate("abc")
	require.NoError(t, err)
	require.NotEmpty(t, other.Salt)
	require.NotEmpty(t, other.Result)
	require.NotEqual(t, []byte("abc"), other.Result)

	require.NotEqual(t, salted.Salt, other.Salt)
	require.NotEqual(t, salted.Result, other.Result)

	bundle := salted.Format()
	require.NotEmpty(t, bundle)

	expected, err := cryptzsalt.Verify(bundle, "abc")
	require.NoError(t, err)
	require.Equal(t, salted.Salt, expected.Salt)
	require.Equal(t, salted.Result, expected.Result)
	require.Equal(t, 50, expected.BundleSize())

	greaterBundle := append(slices.Clone(bundle), []byte("defgh")...)
	expected, err = cryptzsalt.Verify(greaterBundle, "abc")
	require.NoError(t, err)
	require.Equal(t, salted.Salt, expected.Salt)
	require.Equal(t, salted.Result, expected.Result)
	require.Equal(t, 50, expected.BundleSize())
	require.Equal(t, "defgh", string(greaterBundle[expected.BundleSize():]))

	_, err = cryptzsalt.Verify(bundle, "abc2")
	require.Error(t, err)
	require.EqualError(t, err, "hash mismatch")

	_, err = cryptzsalt.Generate(make([]byte, cryptzsalt.MaxSize+1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "original data size")
	require.Contains(t, err.Error(), "exceeds max size")
}
