package logz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/stretchr/testify/require"
)

func TestManualSample(t *testing.T) {
	ctx := t.Context()

	type tlogger logz.Type

	var logger = logz.Create(tlogger(true))

	logger.Info(ctx, "a1", nil)

	collector := logger.(logz.Collector).Collector()

	logger.Info(ctx, "a2", nil)
	logger.Info(ctx, "a3", nil)

	s := collector()
	require.Equal(t, "a2", s[0].Operation)
	require.Equal(t, "a3", s[1].Operation)

	require.Empty(t, collector())

}
