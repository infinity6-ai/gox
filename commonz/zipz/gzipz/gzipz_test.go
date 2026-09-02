package gzipz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/zipz/gzipz"
	"github.com/stretchr/testify/assert"
)

func TestUnitGzipGunzip(t *testing.T) {
	originalData := "This is some data to compress"

	compressedData, err := gzipz.Gzip(blobz.New(originalData))
	assert.NoError(t, err)

	decompressedData, err := gzipz.Gunzip(compressedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decompressedData.String())
}
