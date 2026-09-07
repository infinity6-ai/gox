package staticzlocal_test

import (
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzlocal"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	dir, err := staticzlocal.LookupCurrentDir(stzfiles.Name)
	require.NoError(t, err)
	require.NotEmpty(t, dir)
	require.Equal(t, "commonz sample\n", filez.MustReadFile(filepath.Join(dir, "commonz", "commonz-sample.txt"), 256).String())
}
