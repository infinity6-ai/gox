package cmdbrwjsz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/spf13/cobra"
)

type execOptions struct {
	code     string
	file     string
	args     []string
	argFiles []string
	raw      bool
	headless bool
	bin      string
	timeout  time.Duration
}

func parseArgValue(s string) any {
	var val any
	if err := json.Unmarshal([]byte(s), &val); err == nil {
		return val
	}
	return s
}

func newExecCmd() *cobra.Command {
	opts := &execOptions{}

	cmd := &cobra.Command{
		Use:   "exec [code]",
		Short: "Execute JavaScript in a headless browser session",
		RunE: func(cmd *cobra.Command, args []string) error {
			code := opts.code
			if code == "" && opts.file != "" {
				if opts.file == "-" {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
					code = string(b)
				} else {
					b, err := os.ReadFile(opts.file)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", opts.file, err)
					}
					code = string(b)
				}
			} else if code == "" && len(args) > 0 {
				code = args[0]
			}

			if code == "" {
				return fmt.Errorf("no code specified (use --code, --file, or code argument)")
			}

			var resolvedArgs []any
			for _, a := range opts.args {
				resolvedArgs = append(resolvedArgs, parseArgValue(a))
			}
			for _, f := range opts.argFiles {
				b, err := os.ReadFile(f)
				if err != nil {
					return fmt.Errorf("failed to read arg file %s: %w", f, err)
				}
				resolvedArgs = append(resolvedArgs, parseArgValue(string(b)))
			}

			ctx := cmd.Context()
			if opts.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, opts.timeout)
				defer cancel()
			}

			brwOpts := brwjsz.Options{}
			brwOpts.Chrome(opts.bin, opts.headless, "")

			b := brwjsz.New(ctx, brwOpts)
			defer b.Close()

			if err := b.Start(ctx); err != nil {
				return fmt.Errorf("failed to start browser: %w", err)
			}

			wrapper := `async (code, args) => {
				window.args = args;
				let target;
				try {
					target = (0, eval)('(' + code + ')');
				} catch (_) {
					target = (0, eval)(code);
				}
				if (typeof target === 'function') {
					let res = target.apply(window, args);
					if (res && typeof res.then === 'function') {
						res = await res;
					}
					return res;
				}
				if (target && typeof target.then === 'function') {
					target = await target;
				}
				return target;
			}`

			res, err := brwclient.Call[any](ctx, b.Client(), wrapper, code, resolvedArgs)
			if err != nil {
				return fmt.Errorf("execution error: %w", err)
			}

			if res == nil {
				return nil
			}

			if str, ok := res.(string); ok && opts.raw {
				fmt.Println(str)
				return nil
			}

			data, err := json.Marshal(res)
			if err != nil {
				fmt.Printf("%v\n", res)
				return nil
			}
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.code, "code", "", "JavaScript code to execute")
	cmd.Flags().StringVar(&opts.file, "file", "", "JavaScript file path to execute (or - for stdin)")
	cmd.Flags().StringArrayVar(&opts.args, "arg", nil, "Argument to pass to the function (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&opts.argFiles, "arg-file", nil, "File containing argument value (can be specified multiple times)")
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "Print string output without JSON quotes")
	cmd.Flags().BoolVar(&opts.headless, "headless", true, "Run browser in headless mode")
	cmd.Flags().StringVar(&opts.bin, "bin", "", "Browser binary path (defaults to auto-detect)")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 30*time.Second, "Execution timeout")

	return cmd
}
