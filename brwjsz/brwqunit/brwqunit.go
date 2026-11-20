package brwqunit

import (
	"context"
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

// TestError represents assertion failure details in a QUnit test.
type TestError struct {
	Message  string `json:"message"`
	Passed   bool   `json:"passed"`
	Actual   string `json:"actual,omitempty"`
	Expected string `json:"expected,omitempty"`
	Stack    string `json:"stack,omitempty"`
}

// TestResult represents the outcome of an individual QUnit test.
type TestResult struct {
	Name      string      `json:"name"`
	SuiteName string      `json:"suiteName"`
	Status    string      `json:"status"` // "passed", "failed", "skipped", "todo"
	Runtime   int64       `json:"runtime"`
	Errors    []TestError `json:"errors"`
}

// TestReport represents the aggregate summary of a QUnit test run.
type TestReport struct {
	Status  string       `json:"status"` // "passed" or "failed"
	Total   int          `json:"total"`
	Passed  int          `json:"passed"`
	Failed  int          `json:"failed"`
	Skipped int          `json:"skipped"`
	Runtime int64        `json:"runtime"`
	Tests   []TestResult `json:"tests"`
}

// Run executes the specified QUnit test files in the browser and returns the structured test report.
func Run(ctx context.Context, c *brwclient.Client, testUrls ...string) (*TestReport, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	ensureRunner := `async () => {
		if (!window.brwqunit) {
			await import('/api/brwjs/f/qunit/runner.js');
		}
	}`
	if err := brwclient.Proc(ctx, c, ensureRunner); err != nil {
		return nil, fmt.Errorf("failed to load qunit runner: %w", err)
	}

	report, err := brwclient.Call[*TestReport](ctx, c, `(urls) => window.brwqunit.run(urls)`, testUrls)
	if err != nil {
		return nil, fmt.Errorf("failed running qunit tests: %w", err)
	}
	return report, nil
}

// MustRun executes the specified QUnit test files in the browser, panicking on failure.
func MustRun(ctx context.Context, c *brwclient.Client, testUrls ...string) *TestReport {
	report, err := Run(ctx, c, testUrls...)
	errorz.Check(err)
	return report
}

// Test executes the specified QUnit test files in the browser and maps each test to a Go sub-test (t.Run).
func Test(t *testing.T, c *brwclient.Client, testUrls ...string) {
	t.Helper()
	report, err := Run(t.Context(), c, testUrls...)
	if err != nil {
		t.Fatalf("failed running QUnit tests: %v", err)
	}

	for _, test := range report.Tests {
		testName := test.Name
		if test.SuiteName != "" {
			testName = test.SuiteName + " > " + test.Name
		}
		t.Run(testName, func(t *testing.T) {
			if test.Status == "skipped" {
				t.Skip("QUnit test skipped")
				return
			}
			if test.Status != "passed" {
				for _, e := range test.Errors {
					t.Errorf("Assertion failed: %s (expected: %s, actual: %s)\n%s", e.Message, e.Expected, e.Actual, e.Stack)
				}
			}
		})
	}

	if report.Status != "passed" && report.Failed > 0 {
		t.Errorf("QUnit run failed: %d/%d tests failed", report.Failed, report.Total)
	}
}
