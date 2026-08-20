package runtimez_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/runtimez"
	"github.com/stretchr/testify/require"
)

func TestUnitStackTraceString(t *testing.T) {
	stack := runtimez.StackTraceString(2)
	require.NotEmpty(t, stack)
	require.True(t, strings.Contains(stack, "runtimez_unit_test.go"))
	require.True(t, strings.Contains(stack, "TestUnitStackTraceString"))

	lines := strings.Split(stack, "\n")
	require.Regexp(t, `^.*runtimez_unit_test.go:[0-9]+\ \(.+\)$`, lines[0])
}
