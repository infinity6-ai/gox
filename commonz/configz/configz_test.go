package configz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/configz"
	"github.com/stretchr/testify/require"
)

func TestUnitReverter(t *testing.T) {
	configSample := configz.Create("I6_CONFIG_SAMPLE", "S1")

	ctx := t.Context()
	v, err := configSample.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "S1", v)

	require.Panics(t, func() { configz.Create("I6_CONFIG_SAMPLE", "S1") })

	reverter, err := configSample.Set(ctx, "anothervalue")
	require.NoError(t, err)
	defer reverter.Close()

	v, err = configSample.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "anothervalue", v)

	reverter.Close()
	v, err = configSample.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "S1", v)
}
