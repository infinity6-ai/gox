package fsz_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/fsz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

const testBucket = "i6-bucket" // Provided by the user

func TestRemoteGsProvider(t *testing.T) {
	ctx := context.Background()

	// Helper function for common test logic
	check := func(t *testing.T, scenarioName string, fn func(t *testing.T, u *urlz.Url)) {
		t.Helper()
		t.Run(scenarioName, func(t *testing.T) {
			testObjectName := fmt.Sprintf("test-object-%d-%s", time.Now().UnixNano(), scenarioName)
			testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, testObjectName))
			require.NoError(t, err)

			defer func() {
				// Clean up: Delete the test object after each test scenario
				err := fsz.Delete(ctx, testUrl)
				if err != nil {
					t.Logf("Failed to clean up test object %s: %v", testUrl, err)
				}
			}()

			fn(t, testUrl)
		})
	}

	check(t, "UploadAndStat", func(t *testing.T, u *urlz.Url) {
		content := "hello google cloud storage"
		err := fsz.Upload(ctx, u, nil, strings.NewReader(content))
		require.NoError(t, err)

		stat, err := fsz.Stat(ctx, u)
		require.NoError(t, err)
		require.NotNil(t, stat)
		require.Equal(t, uint64(len(content)), stat.Size)
		require.NotNil(t, stat.CreatedAt)
		require.NotNil(t, stat.UpdatedAt)
	})

	check(t, "Download", func(t *testing.T, u *urlz.Url) {
		content := "download test content"
		err := fsz.Upload(ctx, u, nil, strings.NewReader(content))
		require.NoError(t, err)

		var downloadedContent bytes.Buffer
		err = fsz.Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) error {
			require.True(t, found)
			_, err := io.Copy(&downloadedContent, reader)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, content, downloadedContent.String())
	})

	check(t, "DownloadNotFound", func(t *testing.T, u *urlz.Url) {
		err := fsz.Download(ctx, u, func(found bool, headers http.Header, reader io.Reader) error {
			require.False(t, found)
			require.Nil(t, headers)
			require.Nil(t, reader)
			return nil
		})
		require.NoError(t, err)
	})

	check(t, "StatNotFound", func(t *testing.T, u *urlz.Url) {
		stat, err := fsz.Stat(ctx, u)
		require.NoError(t, err)
		require.Nil(t, stat)
	})

	check(t, "Ls", func(t *testing.T, u *urlz.Url) {
		// Create a directory-like structure
		basePath := strings.TrimPrefix(u.Path.String(), "/") // Get the object name part
		file1URL, _ := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1/file1.txt", testBucket, basePath))
		file2URL, _ := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1/file2.txt", testBucket, basePath))
		file3URL, _ := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir2/file3.txt", testBucket, basePath))

		fsz.Upload(ctx, file1URL, nil, strings.NewReader("content1"))
		fsz.Upload(ctx, file2URL, nil, strings.NewReader("content2"))
		fsz.Upload(ctx, file3URL, nil, strings.NewReader("content3"))

		// Ensure cleanup for these files
		defer fsz.Delete(ctx, file1URL)
		defer fsz.Delete(ctx, file2URL)
		defer fsz.Delete(ctx, file3URL)

		prefixURL, _ := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1/", testBucket, basePath))
		paginator, err := fsz.Ls(ctx, prefixURL)
		require.NoError(t, err)

		stats, err := paginator.Paginate(ctx, 10)
		require.NoError(t, err)
		require.Len(t, stats, 2) // Should find file1.txt and file2.txt

		foundFiles := make(map[string]bool)
		for _, stat := range stats {
			foundFiles[stat.Url.String()] = true
		}
		require.True(t, foundFiles[file1URL.String()])
		require.True(t, foundFiles[file2URL.String()])
	})

	check(t, "CopySameBucket", func(t *testing.T, u *urlz.Url) {
		srcContent := "source content for copy"
		srcURL := u
		destURL, _ := urlz.Parse(fmt.Sprintf("gs://%s/%s-copy", testBucket, strings.TrimPrefix(u.Path.String(), "/")))

		err := fsz.Upload(ctx, srcURL, nil, strings.NewReader(srcContent))
		require.NoError(t, err)

		err = fsz.Copy(ctx, srcURL, destURL)
		require.NoError(t, err)

		// Verify destination
		var downloadedContent bytes.Buffer
		err = fsz.Download(ctx, destURL, func(found bool, headers http.Header, reader io.Reader) error {
			require.True(t, found)
			_, err := io.Copy(&downloadedContent, reader)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, srcContent, downloadedContent.String())

		// Clean up destination
		fsz.Delete(ctx, destURL)
	})

	t.Run("SignGet", func(t *testing.T) {
		objectName := fmt.Sprintf("test-signed-get-%d", time.Now().UnixNano())
		testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, objectName))
		require.NoError(t, err)

		content := "signed content"
		err = fsz.Upload(ctx, testUrl, nil, strings.NewReader(content))
		require.NoError(t, err)
		defer fsz.Delete(ctx, testUrl)

		signedURL, err := fsz.SignGet(ctx, testUrl, 5*time.Minute)
		if err != nil {
			if strings.Contains(err.Error(), "missing required GoogleAccessID") {
				t.Skip("Skipping SignGet test: environment not configured for signing URLs")
			}
			require.NoError(t, err)
		}
		require.NotEmpty(t, signedURL)

		// Attempt to download using the signed URL
		resp, err := http.Get(signedURL)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, content, string(data))
	})

	// TODO: Add more tests for SignPut, SignDelete, and cross-bucket copy if applicable
}
