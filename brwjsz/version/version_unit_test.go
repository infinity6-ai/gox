package version_test

import (
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/version"
	"github.com/stretchr/testify/require"
)

func TestUnitVersion(t *testing.T) {
	require.NotEmpty(t, version.Version())
}
