package fsz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitFszFileProvider(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// 1. Upload a file
	content := "hello world"
	filePath := filepath.Join(tmpDir, "test.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	err = Upload(ctx, u, strings.NewReader(content))
	require.NoError(t, err)

	// 2. Stat the file
	stat, err := Stat(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, stat)
	require.Equal(t, uint64(len(content)), stat.Size)
	require.Equal(t, "text/plain; charset=utf-8", stat.ContentType)
	require.NotEmpty(t, stat.Md5)
	require.Equal(t, stat.Md5, stat.Etag)

	// 3. Download the file
	err = Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) {
		require.True(t, found)
		require.NotNil(t, reader)
		defer (reader.(io.Closer)).Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.Equal(t, content, string(data))
	})
	require.NoError(t, err)

	// 4. Delete the file
	err = Delete(ctx, u)
	require.NoError(t, err)

	// 5. Stat again, should be nil
	stat, err = Stat(ctx, u)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestUnitFszLs(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-ls-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Create some files
	for i := 0; i < 5; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		err := os.WriteFile(filePath, []byte(fmt.Sprintf("content %d", i)), 0644)
		require.NoError(t, err)
	}

	u, err := urlz.Parse("file://" + tmpDir)
	require.NoError(t, err)

	paginator, err := Ls(ctx, u)
	require.NoError(t, err)

	// Paginate with max=2
	stats1 := paginator.Paginate(ctx, 2)
	require.Len(t, stats1, 2)

	// Paginate with max=2
	stats2 := paginator.Paginate(ctx, 2)
	require.Len(t, stats2, 2)

	// Paginate with max=2 (should get the last one)
	stats3 := paginator.Paginate(ctx, 2)
	require.Len(t, stats3, 1)

	// Paginate again, should be empty
	stats4 := paginator.Paginate(ctx, 2)
	require.Len(t, stats4, 0)
}

func TestUnitDownloadNotFound(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-notfound-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	err = Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) {
		require.False(t, found)
		require.Nil(t, headers)
		require.Nil(t, reader)
	})
	require.NoError(t, err)
}

func TestUnitStatNotFound(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-stat-notfound-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")
	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	stat, err := Stat(ctx, u)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestUnitDownloadCallbackReader(t *testing.T) {
	tmpDir := filez.CreateTempDir("fsz-download-callback-test")
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	content := "some test data"
	filePath := filepath.Join(tmpDir, "data.txt")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	u, err := urlz.Parse("file://" + filePath)
	require.NoError(t, err)

	var downloadedData bytes.Buffer
	err = Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) {
		require.True(t, found)
		_, err := io.Copy(&downloadedData, reader)
		require.NoError(t, err)
		(reader.(io.Closer)).Close()
	})
	require.NoError(t, err)
	require.Equal(t, content, downloadedData.String())
}
