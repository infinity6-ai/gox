package cryptzrand_test

import (
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrand"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	k1 := cryptzrand.Rand(10)
	require.Equal(t, 10, len(k1))

	k2 := cryptzrand.Rand(10)
	require.Equal(t, 10, len(k2))

	require.NotEqual(t, k1, k2)
}
