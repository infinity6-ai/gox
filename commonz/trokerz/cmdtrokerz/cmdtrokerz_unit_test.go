package cmdtrokerz_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/trokerz/cmdtrokerz"
	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdTrokerzPrepare(t *testing.T) {
	ctx := context.Background()

	t.Run("Prepare mounts subcommand when parent provided", func(t *testing.T) {
		parent := &cobra.Command{Use: "parent"}
		cmd := cmdtrokerz.Prepare(ctx, parent)
		require.NotNil(t, cmd)
		require.Equal(t, "trokerz", cmd.Use)
		require.Equal(t, version.Version(), cmd.Version)

		found, _, err := parent.Find([]string{"trokerz"})
		require.NoError(t, err)
		require.Equal(t, cmd, found)

		require.Equal(t, "{{", cmd.Flags().Lookup("begin").DefValue)
		require.Equal(t, "}}", cmd.Flags().Lookup("end").DefValue)
		require.Equal(t, "", cmd.Flags().Lookup("in").DefValue)
		require.Equal(t, "", cmd.Flags().Lookup("out").DefValue)
		require.Equal(t, "false", cmd.Flags().Lookup("inplace").DefValue)
		require.Equal(t, "false", cmd.Flags().Lookup("recursive").DefValue)
		require.Equal(t, "r", cmd.Flags().Lookup("recursive").Shorthand)
		require.Equal(t, "false", cmd.Flags().Lookup("dry-run").DefValue)
		require.Equal(t, "false", cmd.Flags().Lookup("fail-missing").DefValue)
	})

	t.Run("Prepare returns standalone command when parent is nil", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		require.NotNil(t, cmd)
		require.Equal(t, "trokerz", cmd.Use)
	})
}

func TestUnitCmdTrokerzPipe(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		input  string
		args   []string
		want   string
		errMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cmd := cmdtrokerz.Prepare(ctx, nil)

		var outBuf bytes.Buffer
		cmd.SetIn(bytes.NewBufferString(s.input))
		cmd.SetOut(&outBuf)
		cmd.SetArgs(s.args)

		err := cmd.ExecuteContext(ctx)
		if s.errMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.errMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.want, outBuf.String())
	}

	t.Run("Pipes stdin to stdout with explicit params", func(t *testing.T) {
		check(t, testScenario{
			input: "Hello {{first}} {{last}}!",
			args:  []string{"--params", "first=John", "--params", "last=Doe"},
			want:  "Hello John Doe!",
		})
	})

	t.Run("Custom delimiters work via flags", func(t *testing.T) {
		check(t, testScenario{
			input: "val: <<name>>",
			args:  []string{"--begin", "<<", "--end", ">>", "--params", "name=custom"},
			want:  "val: custom",
		})
	})

	t.Run("Escaped delimiters in input are unescaped", func(t *testing.T) {
		check(t, testScenario{
			input: "raw {{{{name}}}} and {{name}}",
			args:  []string{"--params", "name=replaced"},
			want:  "raw {{name}} and replaced",
		})
	})

	t.Run("Retains unreplaced placeholder by default", func(t *testing.T) {
		check(t, testScenario{
			input: "keep {{unreplaced}} and change {{replaced}}",
			args:  []string{"--params", "replaced=OK"},
			want:  "keep {{unreplaced}} and change OK",
		})
	})

	t.Run("Fails on missing parameter when --fail-missing is set", func(t *testing.T) {
		check(t, testScenario{
			input:  "Hello {{missing}}!",
			args:   []string{"--fail-missing"},
			errMsg: `missing parameter "missing"`,
		})
	})

	t.Run("Fails on invalid template syntax", func(t *testing.T) {
		check(t, testScenario{
			input:  "Hello {{unclosed",
			args:   []string{},
			errMsg: "unclosed placeholder",
		})
	})

	t.Run("Fails on invalid param format without equal sign", func(t *testing.T) {
		check(t, testScenario{
			input:  "Hello {{name}}",
			args:   []string{"--params", "invalid_param"},
			errMsg: `invalid param "invalid_param", expected key=value`,
		})
	})
}

func TestUnitCmdTrokerzParamPrecedence(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "params.json")
	err := os.WriteFile(jsonFile, []byte(`{
		"key1": "from_file",
		"key2": "from_file",
		"key3": "from_file",
		"num": 42
	}`), 0o644)
	require.NoError(t, err)

	t.Setenv("TROKERZ_TEST_KEY2", "from_env")
	t.Setenv("TROKERZ_TEST_KEY3", "from_env")

	cmd := cmdtrokerz.Prepare(ctx, nil)
	var outBuf bytes.Buffer
	cmd.SetIn(bytes.NewBufferString("k1: {{key1}}, k2: {{TROKERZ_TEST_KEY2}}, k3: {{TROKERZ_TEST_KEY3}}, num: {{num}}"))
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{
		"--param-file", jsonFile,
		"--param-envs",
		"--params", "TROKERZ_TEST_KEY3=from_cli",
	})

	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "k1: from_file, k2: from_env, k3: from_cli, num: 42", outBuf.String())
}

func TestUnitCmdTrokerzFileInOut(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	inFile := filepath.Join(tmpDir, "input.txt")
	outFile := filepath.Join(tmpDir, "output.txt")

	err := os.WriteFile(inFile, []byte("Service: {{service}} running on port {{port}}"), 0o644)
	require.NoError(t, err)

	cmd := cmdtrokerz.Prepare(ctx, nil)
	cmd.SetArgs([]string{
		"--in", inFile,
		"--out", outFile,
		"--params", "service=api",
		"--params", "port=8080",
	})

	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(outFile)
	require.NoError(t, err)
	require.Equal(t, "Service: api running on port 8080", string(data))
}

func TestUnitCmdTrokerzInplace(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "inplace.txt")

	err := os.WriteFile(targetFile, []byte("Current version: {{version}}"), 0o600)
	require.NoError(t, err)

	cmd := cmdtrokerz.Prepare(ctx, nil)
	cmd.SetArgs([]string{
		"-i",
		"--in", targetFile,
		"--params", "version=v2.0.0",
	})

	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	require.Equal(t, "Current version: v2.0.0", string(data))

	info, err := os.Stat(targetFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestUnitCmdTrokerzRecursiveInplace(t *testing.T) {
	ctx := context.Background()

	t.Run("Recursively replaces templates across multiple nested directories and files", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "sub")
		subSubDir := filepath.Join(subDir, "sub2")

		require.NoError(t, os.MkdirAll(subSubDir, 0o755))

		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(subDir, "file2.json")
		file3 := filepath.Join(subSubDir, "file3.env")
		file4 := filepath.Join(tmpDir, "plain.txt")

		require.NoError(t, os.WriteFile(file1, []byte("Hello {{name}}!"), 0o644))
		require.NoError(t, os.WriteFile(file2, []byte(`{"app": "{{app}}"}`), 0o600))
		require.NoError(t, os.WriteFile(file3, []byte("APP={{app}}\nNAME={{name}}"), 0o640))
		require.NoError(t, os.WriteFile(file4, []byte("plain text no placeholders"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"-r",
			"-i",
			"--in", tmpDir,
			"--params", "name=Alice",
			"--params", "app=Dashboard",
		})

		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		data1, err := os.ReadFile(file1)
		require.NoError(t, err)
		require.Equal(t, "Hello Alice!", string(data1))

		data2, err := os.ReadFile(file2)
		require.NoError(t, err)
		require.Equal(t, `{"app": "Dashboard"}`, string(data2))

		data3, err := os.ReadFile(file3)
		require.NoError(t, err)
		require.Equal(t, "APP=Dashboard\nNAME=Alice", string(data3))

		data4, err := os.ReadFile(file4)
		require.NoError(t, err)
		require.Equal(t, "plain text no placeholders", string(data4))

		// Check permissions preserved
		info1, err := os.Stat(file1)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o644), info1.Mode().Perm())

		info2, err := os.Stat(file2)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info2.Mode().Perm())

		info3, err := os.Stat(file3)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o640), info3.Mode().Perm())
	})

	t.Run("Recursively replaces with custom delimiters", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "nested")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		targetFile := filepath.Join(subDir, "config.ini")
		require.NoError(t, os.WriteFile(targetFile, []byte("value = <<custom>>"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"--recursive",
			"--inplace",
			"--in", tmpDir,
			"--begin", "<<",
			"--end", ">>",
			"--params", "custom=success",
		})

		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		data, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		require.Equal(t, "value = success", string(data))
	})

	t.Run("Fails and reports error when nested file template fails missing check", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "sub")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		targetFile := filepath.Join(subDir, "missing.txt")
		require.NoError(t, os.WriteFile(targetFile, []byte("unresolved: {{missing_val}}"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"-r",
			"-i",
			"--in", tmpDir,
			"--fail-missing",
		})

		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `missing parameter "missing_val"`)
		require.Contains(t, err.Error(), "failed to recursively process files")
	})
}

func TestUnitCmdTrokerzDryRun(t *testing.T) {
	ctx := context.Background()

	t.Run("Single file inplace dry run does not modify file", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetFile := filepath.Join(tmpDir, "target.txt")
		originalContent := "Hello {{name}}!"
		require.NoError(t, os.WriteFile(targetFile, []byte(originalContent), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"--inplace",
			"--dry-run",
			"--in", targetFile,
			"--params", "name=DryRunBob",
		})

		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		data, err := os.ReadFile(targetFile)
		require.NoError(t, err)
		require.Equal(t, originalContent, string(data))
	})

	t.Run("Recursive inplace dry run does not modify nested files", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "sub")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		file1 := filepath.Join(tmpDir, "a.txt")
		file2 := filepath.Join(subDir, "b.txt")
		content1 := "A: {{valA}}"
		content2 := "B: {{valB}}"
		require.NoError(t, os.WriteFile(file1, []byte(content1), 0o644))
		require.NoError(t, os.WriteFile(file2, []byte(content2), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"-r",
			"-i",
			"--dry-run",
			"--in", tmpDir,
			"--params", "valA=Alpha",
			"--params", "valB=Beta",
		})

		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		data1, err := os.ReadFile(file1)
		require.NoError(t, err)
		require.Equal(t, content1, string(data1))

		data2, err := os.ReadFile(file2)
		require.NoError(t, err)
		require.Equal(t, content2, string(data2))
	})

	t.Run("Out file dry run does not create output file", func(t *testing.T) {
		tmpDir := t.TempDir()
		inFile := filepath.Join(tmpDir, "in.txt")
		outFile := filepath.Join(tmpDir, "nested", "out.txt")
		require.NoError(t, os.WriteFile(inFile, []byte("Hello {{name}}!"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"--in", inFile,
			"--out", outFile,
			"--dry-run",
			"--params", "name=World",
		})

		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		_, err = os.Stat(outFile)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("Dry run with fail-missing still validates template errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetFile := filepath.Join(tmpDir, "target.txt")
		require.NoError(t, os.WriteFile(targetFile, []byte("Missing: {{not_found}}"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{
			"-i",
			"--dry-run",
			"--in", targetFile,
			"--fail-missing",
		})

		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `missing parameter "not_found"`)
	})
}

func TestUnitCmdTrokerzValidationErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("Dry run without inplace or out returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--dry-run", "--in", "somepath"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--dry-run requires --inplace or --out")
	})

	t.Run("Dry run with stdout out returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--dry-run", "--out", "-"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--dry-run requires --inplace or --out")
	})

	t.Run("Recursive without inplace returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--recursive", "--in", "somepath"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--recursive can only be used with --inplace")
	})

	t.Run("Recursive shorthand -r without inplace returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"-r", "--in", "somepath"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--recursive can only be used with --inplace")
	})

	t.Run("Recursive with non-existent input path returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"-r", "-i", "--in", "/non/existent/path/for/trokerz"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read input path")
	})

	t.Run("Inplace without in file returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--inplace"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--inplace requires a valid input file path")
	})

	t.Run("Inplace with dash in file returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--inplace", "--in", "-"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--inplace requires a valid input file path")
	})

	t.Run("Inplace with out file returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--inplace", "--in", "somefile", "--out", "outfile"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--inplace cannot be used with --out")
	})

	t.Run("Non-existent input file returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetArgs([]string{"--in", "/non/existent/file.txt"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read input file")
	})

	t.Run("Non-existent param file returns error", func(t *testing.T) {
		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetIn(bytes.NewBufferString("hello"))
		cmd.SetArgs([]string{"--param-file", "/non/existent/params.json"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read param file")
	})

	t.Run("Malformed JSON in param file returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		badJson := filepath.Join(tmpDir, "bad.json")
		require.NoError(t, os.WriteFile(badJson, []byte("not valid json"), 0o644))

		cmd := cmdtrokerz.Prepare(ctx, nil)
		cmd.SetIn(bytes.NewBufferString("hello"))
		cmd.SetArgs([]string{"--param-file", badJson})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JSON")
	})
}
