// Package cmdqunit provides the CLI command to execute QUnit tests in a browser session.
package cmdqunit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/infinity6-ai/gox/brwjsz/brwqunit"
	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdserver"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/spf13/cobra"
)

type qunitOptions struct {
	site            string
	suffix          string
	state           string
	bin             string
	disableHeadless bool
	timeout         time.Duration
	importMap       string
}

const discoveryRunnerJs = `async (suffix, opts) => {
	if (!suffix) {
		suffix = '-test.js';
	}
	if (!window.brwqunit) {
		await import('/api/brwjs/f/qunit/runner.js');
	}
	const res = await fetch('/?t=json&r=true');
	if (!res.ok) {
		throw new Error('Failed to list files from site: ' + res.statusText);
	}
	const data = await res.json();
	const testUrls = (data.entries || [])
		.filter(e => !e.dir && e.path.endsWith(suffix))
		.map(e => e.path);

	if (testUrls.length === 0) {
		return {
			status: 'passed',
			total: 0,
			passed: 0,
			failed: 0,
			skipped: 0,
			tests: []
		};
	}

	return await window.brwqunit.run(testUrls, opts);
}`

func formatReport(report *brwqunit.TestReport) error {
	if report == nil {
		return fmt.Errorf("no report received")
	}

	if report.Total == 0 {
		fmt.Println("No QUnit test files found matching suffix.")
		return nil
	}

	fmt.Printf("\n--- QUnit Test Results (%d tests) ---\n", report.Total)
	for _, t := range report.Tests {
		name := t.Name
		if t.SuiteName != "" {
			name = t.SuiteName + " > " + t.Name
		}
		if t.Status == "passed" {
			fmt.Printf("  ✓ %s (%dms)\n", name, t.Runtime)
		} else if t.Status == "skipped" {
			fmt.Printf("  - %s (skipped)\n", name)
		} else {
			fmt.Printf("  ✗ %s (%dms)\n", name, t.Runtime)
			for _, e := range t.Errors {
				fmt.Printf("      Assertion failed: %s\n", e.Message)
				if e.Expected != "" || e.Actual != "" {
					fmt.Printf("        Expected: %s\n        Actual:   %s\n", e.Expected, e.Actual)
				}
				if e.Stack != "" {
					fmt.Printf("        Stack: %s\n", e.Stack)
				}
			}
		}
	}

	if report.Status == "passed" && report.Failed == 0 {
		fmt.Printf("\nPASS: %d/%d tests passed (%dms)\n", report.Passed, report.Total, report.Runtime)
		return nil
	}

	return fmt.Errorf("FAIL: %d/%d tests failed", report.Failed, report.Total)
}

func parseImportMap(raw string) any {
	if raw == "" {
		return nil
	}
	if content, err := os.ReadFile(raw); err == nil {
		var parsed any
		if err := json.Unmarshal(content, &parsed); err == nil {
			return parsed
		}
		return string(content)
	}
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return parsed
	}
	return raw
}

// Prepare configures and attaches the qunit command to the parent cobra command.
func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	opts := &qunitOptions{}

	cmd := &cobra.Command{
		Use:   "qunit",
		Short: "Discover and execute QUnit tests in a browser session",
		RunE: func(cmd *cobra.Command, args []string) error {
			suffix := opts.suffix
			if suffix == "" {
				suffix = "-test.js"
			}

			runOpts := map[string]any{}
			if im := parseImportMap(opts.importMap); im != nil {
				runOpts["importMap"] = im
			}

			// Mode A: Run against active server daemon
			if opts.state != "" {
				state, err := cmdserver.ReadState(opts.state)
				if err != nil {
					return fmt.Errorf("failed to read state: %w", err)
				}

				reqPayload, err := json.Marshal(handlers.ExecRequest{
					Code: discoveryRunnerJs,
					Args: []any{suffix, runOpts},
				})
				if err != nil {
					return fmt.Errorf("failed to marshal qunit payload: %w", err)
				}

				execUrl := state.URL + handlers.ExecPath
				httpReq, err := http.NewRequestWithContext(cmd.Context(), "POST", execUrl, bytes.NewReader(reqPayload))
				if err != nil {
					return fmt.Errorf("failed to create http request: %w", err)
				}
				httpReq.Header.Set("Content-Type", "application/json")

				httpClient := &http.Client{Timeout: opts.timeout}
				httpResp, err := httpClient.Do(httpReq)
				if err != nil {
					return fmt.Errorf("failed to send qunit request to server: %w", err)
				}
				defer httpResp.Body.Close()

				var execResp handlers.ExecResponse
				if err := json.NewDecoder(httpResp.Body).Decode(&execResp); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}

				if !execResp.Ok {
					return fmt.Errorf("qunit execution error: %s", execResp.Error)
				}

				resBytes, err := json.Marshal(execResp.Result)
				if err != nil {
					return fmt.Errorf("failed to marshal exec result: %w", err)
				}

				var report brwqunit.TestReport
				if err := json.Unmarshal(resBytes, &report); err != nil {
					return fmt.Errorf("failed to parse test report: %w", err)
				}

				return formatReport(&report)
			}

			// Mode B: Standalone one-shot browser run
			siteDir := opts.site
			if siteDir == "" {
				siteDir = "."
			}
			absSite, err := filepath.Abs(siteDir)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute site path %q: %w", siteDir, err)
			}

			cmdCtx := cmd.Context()
			if opts.timeout > 0 {
				var cancel context.CancelFunc
				cmdCtx, cancel = context.WithTimeout(cmdCtx, opts.timeout)
				defer cancel()
			}

			brwOpts := brwjsz.Options{
				Site: &brwjsz.Site{
					Disk: pathz.MustParse(absSite),
				},
			}
			brwOpts.Chrome(opts.bin, !opts.disableHeadless, "")

			b := brwjsz.New(cmdCtx, brwOpts)
			defer b.Close()

			if err := b.Start(cmdCtx); err != nil {
				return fmt.Errorf("failed to start browser: %w", err)
			}

			report, err := brwclient.Call[*brwqunit.TestReport](cmdCtx, b.Client(), discoveryRunnerJs, suffix, runOpts)
			if err != nil {
				return fmt.Errorf("failed to execute qunit discovery runner: %w", err)
			}

			return formatReport(report)
		},
	}

	cmd.Flags().StringVar(&opts.site, "site", "", "Disk directory path to serve as the test site (defaults to .)")
	cmd.Flags().StringVar(&opts.suffix, "suffix", "-test.js", "Filename suffix for test discovery (recursively matched)")
	cmd.Flags().StringVar(&opts.state, "state", "", "Path to state file of running server daemon")
	cmd.Flags().StringVar(&opts.bin, "bin", "", "Browser binary path (defaults to auto-detect)")
	cmd.Flags().BoolVar(&opts.disableHeadless, "disable-headless", false, "Disable headless browser mode")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 30*time.Second, "Execution timeout")
	cmd.Flags().StringVar(&opts.importMap, "importmap", "", "Import map JSON string or file path for bare module specifiers")

	if parent != nil {
		parent.AddCommand(cmd)
	}

	return cmd
}
