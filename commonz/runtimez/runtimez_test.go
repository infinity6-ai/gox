package runtimez_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/runtimez"
	"github.com/stretchr/testify/assert"
)

func TestUnitStackTraceString(t *testing.T) {
	stack := runtimez.StackTraceString(2)
	assert.NotEmpty(t, stack)
	assert.True(t, strings.Contains(stack, "runtimez_test.go"))
	assert.True(t, strings.Contains(stack, "TestUnitStackTraceString"))
}
