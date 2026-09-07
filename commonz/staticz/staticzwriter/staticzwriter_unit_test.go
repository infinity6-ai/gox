package staticzwriter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitLineEncoder(t *testing.T) {
	type testScenario struct {
		input    string
		limit    int
		expected string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var buf bytes.Buffer
		encoder := &lineEncoder{w: &buf, limit: s.limit}
		n, err := encoder.Write([]byte(s.input))
		require.NoError(t, err)
		require.Equal(t, len(s.input), n)
		require.Equal(t, s.expected, buf.String())
	}

	t.Run("no newlines needed for short input", func(t *testing.T) {
		check(t, testScenario{
			input:    "abcde",
			limit:    10,
			expected: "abcde",
		})
	})

	t.Run("one newline inserted at limit", func(t *testing.T) {
		check(t, testScenario{
			input:    "abcdefghij",
			limit:    10,
			expected: "abcdefghij\n",
		})
	})

	t.Run("multiple newlines for long input", func(t *testing.T) {
		check(t, testScenario{
			input:    "123456789012345678901",
			limit:    10,
			expected: "1234567890\n1234567890\n1",
		})
	})

	t.Run("multiple writes build up to limit", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := &lineEncoder{w: &buf, limit: 5}
		_, err := encoder.Write([]byte("123"))
		require.NoError(t, err)
		require.Equal(t, "123", buf.String())
		_, err = encoder.Write([]byte("45"))
		require.NoError(t, err)
		require.Equal(t, "12345\n", buf.String())
		_, err = encoder.Write([]byte("67"))
		require.NoError(t, err)
		require.Equal(t, "12345\n67", buf.String())
	})
}

func TestUnitCreateTarGz(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "staticz_targz_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	err = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("world"), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NotPanics(t, func() {
		CreateTarGz(tmpDir, &buf)
	})

	// Verify the tar.gz content
	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Check file1.txt
	hdr, err := tr.Next()
	require.NoError(t, err)
	require.Equal(t, "file1.txt", hdr.Name)
	require.False(t, hdr.FileInfo().IsDir())
	require.Equal(t, int64(5), hdr.Size)
	file1Contents, err := io.ReadAll(tr)
	require.NoError(t, err)
	require.Equal(t, "hello", string(file1Contents))

	// Check subdir
	hdr, err = tr.Next()
	require.NoError(t, err)
	require.Equal(t, "subdir", hdr.Name)
	require.True(t, hdr.FileInfo().IsDir())

	// Check subdir/file2.txt
	hdr, err = tr.Next()
	require.NoError(t, err)
	require.Equal(t, "subdir/file2.txt", hdr.Name)
	require.False(t, hdr.FileInfo().IsDir())
	file2Contents, err := io.ReadAll(tr)
	require.NoError(t, err)
	require.Equal(t, "world", string(file2Contents))

	// Check for end of archive
	_, err = tr.Next()
	require.Equal(t, io.EOF, err)
}

func TestUnitGenerate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "staticz_generate_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	stzDir := filepath.Join(tmpDir, "files_to_embed")
	err = os.Mkdir(stzDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(stzDir, "test.txt"), []byte("test data"), 0644)
	require.NoError(t, err)

	check := func(t *testing.T, opts GenerateOptions) {
		t.Helper()
		var buf bytes.Buffer
		opts.Out = &buf // Override Out for capturing output

		require.NotPanics(t, func() {
			Generate(context.Background(), opts)
		})

		generatedCode := buf.String()

		// Use a copy to check defaults without modifying the original
		optsCopy := opts
		optsCopy.fix()

		require.True(t, strings.HasPrefix(generatedCode, "package stzfiles\n"), "bad package")
		require.Contains(t, generatedCode, `import "`+optsCopy.Imp+`"`, "bad import")
		require.Contains(t, generatedCode, "func init() {\n", "bad init function start")
		require.Contains(t, generatedCode, "    "+optsCopy.Code+"(Name, func() string {\n", "bad SetCode call")
		require.Contains(t, generatedCode, "        return `\n", "bad return start")
		require.Contains(t, generatedCode, "\n`\n", "bad return end")
		require.Contains(t, generatedCode, "    })\n}\n", "bad init function end")

		// Extract base64 content
		re := regexp.MustCompile("return `\n([\\s\\S]*?)\n`")
		matches := re.FindStringSubmatch(generatedCode)
		require.Len(t, matches, 2, "could not find base64 content in generated code")

		b64content := strings.ReplaceAll(matches[1], "\n", "")

		// Decode base64
		decoded, err := base64.StdEncoding.DecodeString(b64content)
		require.NoError(t, err, "failed to decode base64 content")

		// Decompress and verify tar
		gr, err := gzip.NewReader(bytes.NewReader(decoded))
		require.NoError(t, err)
		defer gr.Close()

		tr := tar.NewReader(gr)
		hdr, err := tr.Next()
		require.NoError(t, err)
		require.Equal(t, "test.txt", hdr.Name)
		fileContents, err := io.ReadAll(tr)
		require.NoError(t, err)
		require.Equal(t, "test data", string(fileContents))

		_, err = tr.Next()
		require.Equal(t, io.EOF, err)
	}

	t.Run("with default options", func(t *testing.T) {
		check(t, GenerateOptions{
			Name: "MyFiles",
			Dir:  stzDir,
		})
	})

	t.Run("with custom options", func(t *testing.T) {
		check(t, GenerateOptions{
			Imp:  "my/custom/loader",
			Code: "myloader.Load",
			Name: "CustomAssets",
			Dir:  stzDir,
		})
	})
}
