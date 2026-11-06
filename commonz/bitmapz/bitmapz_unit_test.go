package bitmapz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/bitmapz"
	"github.com/stretchr/testify/require"
)

func TestUnitBitmapExpand(t *testing.T) {

	// type Grant bitmapz.Bit
	// type Grants bitmapz.Bitmap

	var read bitmapz.Bit = 1
	var _ bitmapz.Bit = 2
	var admin bitmapz.Bit = 4

	var grants bitmapz.Bitmap = 6

	require.Equal(t, true, bitmapz.Contains(grants, admin))
	require.Equal(t, false, bitmapz.Contains(grants, read))
}
