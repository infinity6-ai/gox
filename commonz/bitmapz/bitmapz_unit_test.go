package bitmapz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/bitmapz"
	"github.com/stretchr/testify/require"
)

func TestUnitBitmapExpand(t *testing.T) {
	require.Equal(t, true, bitmapz.Contains(6, 4))
	require.Equal(t, false, bitmapz.Contains(6, 1))
}
