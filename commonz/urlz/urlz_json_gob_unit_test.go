package urlz_test

import (
	"encoding/json"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/stretchr/testify/require"
)

func TestUnitUrlGobSerialization(t *testing.T) {
	type testScenario struct {
		name     string
		inputURL *urlz.Url
		wantErr  string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Encode
		encoded, err := s.inputURL.GobEncode()
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, encoded)

		// Decode
		decodedURL := &urlz.Url{}
		err = decodedURL.GobDecode(encoded)
		require.NoError(t, err)
		require.Equal(t, s.inputURL, decodedURL)
		require.NotSame(t, s.inputURL, decodedURL)
		require.Equal(t, s.inputURL.Path, decodedURL.Path)
		require.NotSame(t, s.inputURL.Path, decodedURL.Path)
	}

	p1, err := pathz.Parse("/a/b/c")
	require.NoError(t, err)
	p2, err := pathz.Parse("/")
	require.NoError(t, err)

	t.Run("http url", func(t *testing.T) {
		check(t, testScenario{
			name: "gob serialization of http url",
			inputURL: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "example.com",
				Port:     "8080",
				Path:     p1,
				Query:    "param=value",
				Fragment: "frag",
			},
		})
	})

	t.Run("file url", func(t *testing.T) {
		check(t, testScenario{
			name: "gob serialization of file url",
			inputURL: &urlz.Url{
				Scheme: "file",
				Path:   p1,
			},
		})
	})

	t.Run("gs url with root path", func(t *testing.T) {
		check(t, testScenario{
			name: "gob serialization of gs url with root path",
			inputURL: &urlz.Url{
				Scheme: "gs",
				Host:   "my-bucket",
				Path:   p2,
			},
		})
	})

	t.Run("empty url", func(t *testing.T) {
		check(t, testScenario{
			name:     "gob serialization of empty url",
			inputURL: &urlz.Url{},
			wantErr:  "unknown scheme", // Because String() will be called, and an empty scheme is invalid
		})
	})

	t.Run("unix url", func(t *testing.T) {
		p, err := pathz.Parse("/tmp/socket.sock")
		require.NoError(t, err)
		check(t, testScenario{
			name: "gob serialization of unix url",
			inputURL: &urlz.Url{
				Scheme: "unix",
				Path:   p,
			},
		})
	})

	t.Run("boxlocal url", func(t *testing.T) {
		p, err := pathz.Parse("/path/to/box")
		require.NoError(t, err)
		check(t, testScenario{
			name: "gob serialization of boxlocal url",
			inputURL: &urlz.Url{
				Scheme: "boxlocal",
				Path:   p,
			},
		})
	})
}

func TestUnitUrlJsonSerialization(t *testing.T) {
	type testScenario struct {
		name     string
		inputURL *urlz.Url
		wantErr  string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		// Marshal
		marshaled, err := json.Marshal(s.inputURL)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, marshaled)

		// Unmarshal
		decodedURL := &urlz.Url{}
		err = json.Unmarshal(marshaled, decodedURL)
		require.NoError(t, err)
		require.Equal(t, s.inputURL, decodedURL)
		require.NotSame(t, s.inputURL, decodedURL)
		require.Equal(t, s.inputURL.Path, decodedURL.Path)
		require.NotSame(t, s.inputURL.Path, decodedURL.Path)
	}

	p1, err := pathz.Parse("/a/b/c")
	require.NoError(t, err)
	p2, err := pathz.Parse("/")
	require.NoError(t, err)

	t.Run("http url", func(t *testing.T) {
		check(t, testScenario{
			name: "json serialization of http url",
			inputURL: &urlz.Url{
				Scheme:   "http",
				User:     "user",
				Password: "password",
				Host:     "example.com",
				Port:     "8080",
				Path:     p1,
				Query:    "param=value",
				Fragment: "frag",
			},
		})
	})

	t.Run("file url", func(t *testing.T) {
		check(t, testScenario{
			name: "json serialization of file url",
			inputURL: &urlz.Url{
				Scheme: "file",
				Path:   p1,
			},
		})
	})

	t.Run("gs url with root path", func(t *testing.T) {
		check(t, testScenario{
			name: "json serialization of gs url with root path",
			inputURL: &urlz.Url{
				Scheme: "gs",
				Host:   "my-bucket",
				Path:   p2,
			},
		})
	})

	t.Run("empty url", func(t *testing.T) {
		check(t, testScenario{
			name:     "json serialization of empty url",
			inputURL: &urlz.Url{},
			wantErr:  "unknown scheme", // Because String() will be called, and an empty scheme is invalid
		})
	})

	t.Run("unix url", func(t *testing.T) {
		p, err := pathz.Parse("/tmp/socket.sock")
		require.NoError(t, err)
		check(t, testScenario{
			name: "json serialization of unix url",
			inputURL: &urlz.Url{
				Scheme: "unix",
				Path:   p,
			},
		})
	})

	t.Run("boxlocal url", func(t *testing.T) {
		p, err := pathz.Parse("/path/to/box")
		require.NoError(t, err)
		check(t, testScenario{
			name: "json serialization of boxlocal url",
			inputURL: &urlz.Url{
				Scheme: "boxlocal",
				Path:   p,
			},
		})
	})

	t.Run("unmarshal empty string", func(t *testing.T) {
		emptyURL := &urlz.Url{}
		err := json.Unmarshal([]byte(`""`), emptyURL)
		require.NoError(t, err)
		require.Equal(t, &urlz.Url{}, emptyURL)
	})
}
