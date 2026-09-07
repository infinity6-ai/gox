package staticzloader_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	code := staticzloader.GetCode(stzfiles.Name)
	require.NotEmpty(t, code)
}
