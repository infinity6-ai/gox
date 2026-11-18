package cmdtrokerz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/trokerz"
	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
)

type trokerzCmdOptions struct {
	Begin        string
	End          string
	InFile       string
	OutFile      string
	Inplace      bool
	Recursive    bool
	DryRun       bool
	ParamFiles   []string
	ParamEnvs    bool
	ParamEnvList []string
	Params       []string
	FailMissing  bool
}

func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	var (
		begin        string
		end          string
		inFile       string
		outFile      string
		inplace      bool
		recursive    bool
		dryRun       bool
		paramFiles   []string
		paramEnvs    bool
		paramEnvList []string
		cliParams    []string
		failMissing  bool
	)

	cmd := &cobra.Command{
		Version:       version.Version(),
		Use:           "trokerz",
		Short:         "trokerz template replacer",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrokerz(cmd, trokerzCmdOptions{
				Begin:        begin,
				End:          end,
				InFile:       inFile,
				OutFile:      outFile,
				Inplace:      inplace,
				Recursive:    recursive,
				DryRun:       dryRun,
				ParamFiles:   paramFiles,
				ParamEnvs:    paramEnvs,
				ParamEnvList: paramEnvList,
				Params:       cliParams,
				FailMissing:  failMissing,
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&begin, "begin", "{{", "opening delimiter")
	flags.StringVar(&end, "end", "}}", "closing delimiter")
	flags.StringVar(&inFile, "in", "", "input file (defaults to stdin if omitted or '-')")
	flags.StringVarP(&outFile, "out", "o", "", "output file (defaults to stdout if omitted or '-')")
	flags.BoolVarP(&inplace, "inplace", "i", false, "modify input file in-place")
	flags.BoolVarP(&recursive, "recursive", "r", false, "recursively process directory (only with --inplace)")
	flags.BoolVar(&dryRun, "dry-run", false, "simulate run without writing to files (requires --inplace or --out)")
	flags.StringSliceVar(&paramFiles, "param-file", nil, "JSON file(s) containing parameters")
	flags.BoolVar(&paramEnvs, "param-envs", false, "load all environment variables as parameters")
	flags.StringSliceVar(&paramEnvList, "param-env", nil, "specific environment variable(s) to load")
	flags.StringSliceVar(&cliParams, "params", nil, "explicit parameter key=value (can be repeated)")
	flags.BoolVar(&failMissing, "fail-missing", false, "fail if a placeholder has no replacement")

	if parent != nil {
		parent.AddCommand(cmd)
	}

	return cmd
}

func Execute(ctx context.Context, cmd *cobra.Command) {
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runTrokerz(cmd *cobra.Command, opts trokerzCmdOptions) error {
	if opts.Recursive && !opts.Inplace {
		return fmt.Errorf("--recursive can only be used with --inplace")
	}

	if opts.DryRun {
		if !opts.Inplace && (opts.OutFile == "" || opts.OutFile == "-") {
			return fmt.Errorf("--dry-run requires --inplace or --out")
		}
	}

	if opts.Inplace {
		if opts.InFile == "" || opts.InFile == "-" {
			return fmt.Errorf("--inplace requires a valid input file path (--in)")
		}
		if opts.OutFile != "" && opts.OutFile != "-" {
			return fmt.Errorf("--inplace cannot be used with --out")
		}
	}

	params := make(map[string]string)

	for _, path := range opts.ParamFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read param file %q: %w", path, err)
		}
		var rawMap map[string]any
		if err := json.Unmarshal(data, &rawMap); err != nil {
			return fmt.Errorf("failed to parse JSON from %q: %w", path, err)
		}
		for k, v := range rawMap {
			params[k] = valueToString(v)
		}
	}

	if opts.ParamEnvs {
		for _, env := range os.Environ() {
			if k, v, ok := strings.Cut(env, "="); ok {
				params[k] = v
			}
		}
	}
	for _, envKey := range opts.ParamEnvList {
		if val, ok := os.LookupEnv(envKey); ok {
			params[envKey] = val
		}
	}

	for _, p := range opts.Params {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			return fmt.Errorf("invalid param %q, expected key=value", p)
		}
		params[k] = v
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.Inplace {
		if opts.Recursive {
			if _, err := os.Stat(opts.InFile); err != nil {
				return fmt.Errorf("failed to read input path %q: %w", opts.InFile, err)
			}
			err := filepath.WalkDir(opts.InFile, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("failed to access %q: %w", path, err)
				}
				if d.IsDir() {
					return nil
				}
				if !d.Type().IsRegular() {
					return nil
				}
				if strings.HasPrefix(d.Name(), "trokerz-") && strings.HasSuffix(d.Name(), ".tmp") {
					return nil
				}
				if err := replaceFileInplace(ctx, path, opts, params); err != nil {
					return fmt.Errorf("failed to process %q: %w", path, err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to recursively process files: %w", err)
			}
			return nil
		}

		return replaceFileInplace(ctx, opts.InFile, opts, params)
	}

	var inputBytes []byte
	var inPath string
	if opts.InFile != "" && opts.InFile != "-" {
		inPath = opts.InFile
		data, err := os.ReadFile(inPath)
		if err != nil {
			return fmt.Errorf("failed to read input file %q: %w", inPath, err)
		}
		inputBytes = data
	} else {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		inputBytes = data
	}

	fileCtx := ctx
	if inPath != "" {
		fileCtx = logz.SetAttr(ctx, "file", inPath)
	}

	result, err := trokerz.Troke(string(inputBytes), trokerz.Options{
		Ctx:   fileCtx,
		Begin: opts.Begin,
		End:   opts.End,
		Get: func(k string) (string, bool, error) {
			val, exists := params[k]
			if exists {
				return val, true, nil
			}
			if opts.FailMissing {
				return "", false, fmt.Errorf("missing parameter %q", k)
			}
			return "", false, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	if opts.DryRun {
		return nil
	}

	if opts.OutFile != "" && opts.OutFile != "-" {
		if err := os.MkdirAll(filepath.Dir(opts.OutFile), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %q: %w", opts.OutFile, err)
		}
		if err := os.WriteFile(opts.OutFile, []byte(result), 0o644); err != nil {
			return fmt.Errorf("failed to write output file %q: %w", opts.OutFile, err)
		}
		return nil
	}

	if _, err := cmd.OutOrStdout().Write([]byte(result)); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	return nil
}

func replaceFileInplace(ctx context.Context, filePath string, opts trokerzCmdOptions, params map[string]string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read input file %q: %w", filePath, err)
	}

	fileCtx := logz.SetAttr(ctx, "file", filePath)

	result, err := trokerz.Troke(string(data), trokerz.Options{
		Ctx:   fileCtx,
		Begin: opts.Begin,
		End:   opts.End,
		Get: func(k string) (string, bool, error) {
			val, exists := params[k]
			if exists {
				return val, true, nil
			}
			if opts.FailMissing {
				return "", false, fmt.Errorf("missing parameter %q", k)
			}
			return "", false, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	if opts.DryRun {
		return nil
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat input file %q: %w", filePath, err)
	}

	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "trokerz-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.WriteString(result); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Chmod(info.Mode()); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("failed to replace file in-place: %w", err)
	}
	return nil
}

func valueToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprint(val)
	}
}
