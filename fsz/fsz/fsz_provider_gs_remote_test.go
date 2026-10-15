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

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/infinity6-ai/gox/fsz/fsz/fszsignmethod"
	"github.com/stretchr/testify/require"
)

const (
	testBucket  = "i6-bucket" // Provided by the user
	testBucket2 = "i6-tmp"
)

func TestRemoteGsProvider(t *testing.T) {
	ctx := context.Background()

	testObjectName := fmt.Sprintf("test-provider-%d", time.Now().UnixNano())
	testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, testObjectName))
	require.NoError(t, err)

	defer func() {
		_ = fsz.Delete(ctx, testUrl)
	}()

	// 1. Upload
	content := "hello google cloud storage"
	err = fsz.Upload(ctx, testUrl, nil, strings.NewReader(content))
	require.NoError(t, err)

	// 2. Stat
	stat, err := fsz.Stat(ctx, testUrl)
	require.NoError(t, err)
	require.NotNil(t, stat)
	require.Equal(t, uint64(len(content)), stat.Size)
	require.NotNil(t, stat.CreatedAt)
	require.NotNil(t, stat.UpdatedAt)

	// 3. Download
	var downloadedContent bytes.Buffer
	err = fsz.Download(ctx, testUrl, func(found bool, headers http.Header, reader io.Reader) error {
		require.True(t, found)
		_, err := io.Copy(&downloadedContent, reader)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, content, downloadedContent.String())

	// 4. Delete
	err = fsz.Delete(ctx, testUrl)
	require.NoError(t, err)

	// 5. Stat again, should be nil
	stat, err = fsz.Stat(ctx, testUrl)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestRemoteGsDownloadNotFound(t *testing.T) {
	ctx := context.Background()

	testObjectName := fmt.Sprintf("test-download-notfound-%d", time.Now().UnixNano())
	testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, testObjectName))
	require.NoError(t, err)

	err = fsz.Download(ctx, testUrl, func(found bool, headers http.Header, reader io.Reader) error {
		require.False(t, found)
		require.Nil(t, headers)
		require.Nil(t, reader)
		return nil
	})
	require.NoError(t, err)
}

func TestRemoteGsStatNotFound(t *testing.T) {
	ctx := context.Background()

	testObjectName := fmt.Sprintf("test-stat-notfound-%d", time.Now().UnixNano())
	testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, testObjectName))
	require.NoError(t, err)

	stat, err := fsz.Stat(ctx, testUrl)
	require.NoError(t, err)
	require.Nil(t, stat)
}

func TestRemoteGsLs(t *testing.T) {
	ctx := context.Background()

	basePath := fmt.Sprintf("test-ls-%d", time.Now().UnixNano())
	file1URL, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1/file1.txt", testBucket, basePath))
	require.NoError(t, err)
	file2URL, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1/file2.txt", testBucket, basePath))
	require.NoError(t, err)
	file3URL, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir2/file3.txt", testBucket, basePath))
	require.NoError(t, err)

	defer func() {
		_ = fsz.Delete(ctx, file1URL)
		_ = fsz.Delete(ctx, file2URL)
		_ = fsz.Delete(ctx, file3URL)
	}()

	err = fsz.Upload(ctx, file1URL, nil, strings.NewReader("content1"))
	require.NoError(t, err)
	err = fsz.Upload(ctx, file2URL, nil, strings.NewReader("content2"))
	require.NoError(t, err)
	err = fsz.Upload(ctx, file3URL, nil, strings.NewReader("content3"))
	require.NoError(t, err)

	prefixURL, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/dir1", testBucket, basePath))
	require.NoError(t, err)

	paginator, err := fsz.Ls(ctx, prefixURL)
	require.NoError(t, err)
	require.Equal(t, "", paginator.GetCursor())

	stats, err := paginator.Paginate(ctx, 10)
	require.NoError(t, err)
	require.Len(t, stats, 2)
	require.Equal(t, "", paginator.GetCursor())

	foundFiles := make(map[string]bool)
	for _, stat := range stats {
		foundFiles[stat.Url.String()] = true
	}
	require.True(t, foundFiles[file1URL.String()])
	require.True(t, foundFiles[file2URL.String()])
}

func TestRemoteGsCopy(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		destBucket string
		content    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		srcObjectName := fmt.Sprintf("test-copy-src-%d", time.Now().UnixNano())
		srcUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, srcObjectName))
		require.NoError(t, err)

		destObjectName := fmt.Sprintf("test-copy-dest-%d", time.Now().UnixNano())
		destUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", s.destBucket, destObjectName))
		require.NoError(t, err)

		defer func() {
			_ = fsz.Delete(ctx, srcUrl)
			_ = fsz.Delete(ctx, destUrl)
		}()

		err = fsz.Upload(ctx, srcUrl, nil, strings.NewReader(s.content))
		require.NoError(t, err)

		err = fsz.Copy(ctx, srcUrl, destUrl)
		require.NoError(t, err)

		// Verify destination content
		var downloadedContent bytes.Buffer
		err = fsz.Download(ctx, destUrl, func(found bool, headers http.Header, reader io.Reader) error {
			require.True(t, found)
			_, err := io.Copy(&downloadedContent, reader)
			return err
		})
		require.NoError(t, err)
		require.Equal(t, s.content, downloadedContent.String())

		// Verify destination stat
		stat, err := fsz.Stat(ctx, destUrl)
		require.NoError(t, err)
		require.NotNil(t, stat)
		require.Equal(t, uint64(len(s.content)), stat.Size)
	}

	t.Run("Copy within same bucket", func(t *testing.T) {
		check(t, testScenario{
			destBucket: testBucket,
			content:    "source content for copy",
		})
	})

	t.Run("Copy across buckets", func(t *testing.T) {
		check(t, testScenario{
			destBucket: testBucket2,
			content:    "cross-bucket content",
		})
	})
}

func TestRemoteGsSign(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		method  fszsignmethod.SignMethod
		content string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		objectName := fmt.Sprintf("test-signed-%s-%d", strings.ToLower(s.method.String()), time.Now().UnixNano())
		testUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s", testBucket, objectName))
		require.NoError(t, err)

		defer func() {
			_ = fsz.Delete(ctx, testUrl)
		}()

		switch s.method {
		case fszsignmethod.SignMethodGet:
			err = fsz.Upload(ctx, testUrl, nil, strings.NewReader(s.content))
			require.NoError(t, err)

			signedURL, err := fsz.Sign(ctx, s.method, testUrl, 5*time.Minute)
			if err != nil {
				if strings.Contains(err.Error(), "GoogleAccessID") {
					t.Skip("Skipping SignMethodGet test: environment not configured for signing URLs")
				}
				require.NoError(t, err)
			}
			require.NotEmpty(t, signedURL)

			resp, err := http.Get(signedURL)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, s.content, string(data))

		case fszsignmethod.SignMethodPut:
			signedURL, err := fsz.Sign(ctx, s.method, testUrl, 5*time.Minute)
			if err != nil {
				if strings.Contains(err.Error(), "GoogleAccessID") {
					t.Skip("Skipping SignMethodPut test: environment not configured for signing URLs")
				}
				require.NoError(t, err)
			}
			require.NotEmpty(t, signedURL)

			req, err := http.NewRequest(http.MethodPut, signedURL, strings.NewReader(s.content))
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var downloadedContent bytes.Buffer
			err = fsz.Download(ctx, testUrl, func(found bool, headers http.Header, reader io.Reader) error {
				require.True(t, found)
				_, err := io.Copy(&downloadedContent, reader)
				return err
			})
			require.NoError(t, err)
			require.Equal(t, s.content, downloadedContent.String())

		case fszsignmethod.SignMethodDelete:
			err = fsz.Upload(ctx, testUrl, nil, strings.NewReader(s.content))
			require.NoError(t, err)

			signedURL, err := fsz.Sign(ctx, s.method, testUrl, 5*time.Minute)
			if err != nil {
				if strings.Contains(err.Error(), "GoogleAccessID") {
					t.Skip("Skipping SignMethodDelete test: environment not configured for signing URLs")
				}
				require.NoError(t, err)
			}
			require.NotEmpty(t, signedURL)

			req, err := http.NewRequest(http.MethodDelete, signedURL, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusNoContent, resp.StatusCode)

			stat, err := fsz.Stat(ctx, testUrl)
			require.NoError(t, err)
			require.Nil(t, stat)
		}
	}

	t.Run("Sign GET method", func(t *testing.T) {
		check(t, testScenario{
			method:  fszsignmethod.SignMethodGet,
			content: "signed content",
		})
	})

	t.Run("Sign PUT method", func(t *testing.T) {
		check(t, testScenario{
			method:  fszsignmethod.SignMethodPut,
			content: "signed put content",
		})
	})

	t.Run("Sign DELETE method", func(t *testing.T) {
		check(t, testScenario{
			method:  fszsignmethod.SignMethodDelete,
			content: "to be deleted",
		})
	})
}

func TestRemoteGsFind(t *testing.T) {
	ctx := context.Background()

	baseObjectName := fmt.Sprintf("test-find-base-%d", time.Now().UnixNano())
	prefixUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/", testBucket, baseObjectName))
	require.NoError(t, err)

	objects := []string{
		"a/b/c.txt",
		"a/b/d.txt",
		"a/e.txt",
		"f.txt",
		"g/h/i.log",
		"g/j.txt",
	}
	for _, obj := range objects {
		objUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/%s", testBucket, baseObjectName, obj))
		require.NoError(t, err)
		err = fsz.Upload(ctx, objUrl, nil, strings.NewReader("content"))
		require.NoError(t, err)
		defer func(u *urlz.Url) {
			_ = fsz.Delete(ctx, u)
		}(objUrl)
	}

	paginator, err := fsz.Find(ctx, prefixUrl)
	require.NoError(t, err)
	require.Equal(t, "", paginator.GetCursor())

	var foundFiles int
	for {
		stats, err := paginator.Paginate(ctx, 2)
		require.NoError(t, err)
		if len(stats) == 0 {
			require.Equal(t, "", paginator.GetCursor())
			break
		}
		foundFiles += len(stats)
		if foundFiles < len(objects) {
			require.NotEmpty(t, paginator.GetCursor())
		}
	}

	require.Equal(t, len(objects), foundFiles)
	require.Equal(t, "", paginator.GetCursor())
}

func TestRemoteGsFindSetCursor(t *testing.T) {
	ctx := context.Background()

	baseObjectName := fmt.Sprintf("test-cursor-base-%d", time.Now().UnixNano())
	prefixUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/", testBucket, baseObjectName))
	require.NoError(t, err)

	objects := []string{
		"file1.txt",
		"file2.txt",
		"file3.txt",
		"file4.txt",
	}
	for _, obj := range objects {
		objUrl, err := urlz.Parse(fmt.Sprintf("gs://%s/%s/%s", testBucket, baseObjectName, obj))
		require.NoError(t, err)
		err = fsz.Upload(ctx, objUrl, nil, strings.NewReader("content"))
		require.NoError(t, err)
		defer func(u *urlz.Url) {
			_ = fsz.Delete(ctx, u)
		}(objUrl)
	}

	p1, err := fsz.Find(ctx, prefixUrl)
	require.NoError(t, err)

	stats1, err := p1.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats1, 2)
	cursor := p1.GetCursor()
	require.NotEmpty(t, cursor)

	p2, err := fsz.Find(ctx, prefixUrl)
	require.NoError(t, err)
	p2.SetCursor(cursor)

	stats2, err := p2.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, stats2, 2)

	p1Files := make(map[string]bool)
	for _, s := range stats1 {
		p1Files[s.Url.String()] = true
	}
	for _, s := range stats2 {
		require.False(t, p1Files[s.Url.String()], "resumed page should not contain items from first page")
	}

	// Empty cursor resets back to start
	p2.SetCursor("")
	statsReset, err := p2.Paginate(ctx, 2)
	require.NoError(t, err)
	require.Len(t, statsReset, 2)
	for _, s := range statsReset {
		require.True(t, p1Files[s.Url.String()], "reset page should contain initial items")
	}
}
