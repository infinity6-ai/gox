package fszjson_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fszjson"
	"github.com/stretchr/testify/require"
)

type SampleItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnitUploadDownload(t *testing.T) {
	tmpDir := filez.CreateTempDir("fszjson-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "items.json")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	items := []SampleItem{
		{ID: 1, Name: "apple"},
		{ID: 2, Name: "banana"},
	}

	// 1. Upload
	err = fszjson.Upload(ctx, items, fszjson.UploadOptions{
		Url:  u,
		Gzip: false,
	})
	require.NoError(t, err)

	// Verify file was written
	require.FileExists(t, filePath)

	// 2. Download
	var decoded []SampleItem
	found, header, err := fszjson.Download(ctx, &decoded, fszjson.DownloadOptions{
		Url:  u,
		Gzip: false,
	})
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, header)
	require.Len(t, decoded, 2)
	require.Equal(t, items, decoded)
}

func TestUnitUploadDownloadGzip(t *testing.T) {
	tmpDir := filez.CreateTempDir("fszjson-gzip-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "items.json.gz")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	items := []SampleItem{
		{ID: 10, Name: "orange"},
		{ID: 20, Name: "grape"},
	}

	// 1. Upload with Gzip
	err = fszjson.Upload(ctx, items, fszjson.UploadOptions{
		Url:  u,
		Gzip: true,
	})
	require.NoError(t, err)

	// Verify file was written
	require.FileExists(t, filePath)

	// Verify that the file is actually gzip compressed by checking magic bytes (0x1f 0x8b)
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.True(t, len(data) >= 2)
	require.Equal(t, byte(0x1f), data[0])
	require.Equal(t, byte(0x8b), data[1])

	// 2. Download with Gzip
	var decoded []SampleItem
	found, header, err := fszjson.Download(ctx, &decoded, fszjson.DownloadOptions{
		Url:  u,
		Gzip: true,
	})
	require.NoError(t, err)
	require.True(t, found)
	require.NotNil(t, header)
	require.Len(t, decoded, 2)
	require.Equal(t, items, decoded)
}

func TestUnitDownloadNotFound(t *testing.T) {
	tmpDir := filez.CreateTempDir("fszjson-notfound-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "nonexistent.json")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	var decoded []SampleItem
	found, header, err := fszjson.Download(ctx, &decoded, fszjson.DownloadOptions{
		Url:  u,
		Gzip: false,
	})
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, header)
	require.Empty(t, decoded)
}

func TestUnitUploadHeaderPropagation(t *testing.T) {
	tmpDir := filez.CreateTempDir("fszjson-header-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "items_header.json")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	items := []SampleItem{
		{ID: 100, Name: "cherry"},
	}

	customHeader := make(http.Header)
	customHeader.Set("Custom-Key", "Custom-Value")
	customHeader.Set("Content-Type", "application/x-ndjson")

	// 1. Upload with Custom Header
	err = fszjson.Upload(ctx, items, fszjson.UploadOptions{
		Url:    u,
		Gzip:   false,
		Header: customHeader,
	})
	require.NoError(t, err)

	// 2. Download and verify we get correct decoding
	var decoded []SampleItem
	found, _, err := fszjson.Download(ctx, &decoded, fszjson.DownloadOptions{
		Url:  u,
		Gzip: false,
	})
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, decoded, 1)
	require.Equal(t, items, decoded)
}
