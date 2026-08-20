# Commonz Go Project

This is an open-source Go project providing several helper utilities and common functionalities.

## Testing Conventions

We use `github.com/stretchr/testify` for our testing assertions. When writing tests, prefer `require` over `assert` to ensure that tests fail immediately on the first assertion failure, preventing subsequent dependent assertions from running.

### Test Naming Pattern

All test functions should follow the pattern `Test[Type][Scenario]`. The `[Type]` segment categorizes the test based on its dependencies and execution environment:

-   **Unit**: Pure unit tests with no external dependencies.
-   **Remote**: Tests that require external resources or cloud services (e.g., APIs, databases).
-   **Manual**: Tests that are intended to be run manually only, often for complex scenarios or integration with systems not suitable for automated CI.

We do not mix different test types within the same test file. Test file names should reflect the type of tests they contain, using the pattern `[testfilename]_[type]_test.go`.

### Explicit Scenario Sub-Tests

To achieve clear test isolation, direct IDE navigation to failing test data, and maintain code reusability, we employ an "Explicit Scenario Sub-Test" pattern instead of traditional table-driven tests. This approach ensures that when a test fails, the IDE can directly link to the specific data for that scenario.

This pattern involves:
1.  A main `Test[Type][Scenario]` function.
2.  A local `check` helper function (often a closure) containing the common test logic and assertions.
3.  A `testScenario` struct (defined within the test function) to encapsulate all input and expected output data for a single test case.
4.  Explicit `t.Run` calls for each scenario, passing an instantiated `testScenario` struct directly to the `check` helper.

This allows the IDE to link directly to the line where the `testScenario` is instantiated upon a test failure, streamlining debugging.

**Example:**
```go
func TestUnitMyFunction(t *testing.T) {
    type testScenario struct {
        input  string
        want   string
        errMsg string // Example for expected error messages
    }

    check := func(t *testing.T, s testScenario) {
        t.Helper() // Mark as helper to ensure correct line reporting
        got := MyFunction(s.input)
        if s.errMsg != "" {
            require.Error(t, got)
            require.Contains(t, got.Error(), s.errMsg)
            return
        }
        require.NoError(t, got)
        require.Equal(t, s.want, got)
    }

    t.Run("Valid input returns expected output", func(t *testing.T) {
        check(t, testScenario{
            input: "hello",
            want:  "HELLO",
        })
    })

    t.Run("Empty input returns error", func(t *testing.T) {
        check(t, testScenario{
            input:  "",
            errMsg: "input cannot be empty",
        })
    })

    t.Run("Input with numbers is processed correctly", func(t *testing.T) {
        check(t, testScenario{
            input: "123world",
            want:  "123WORLD",
        })
    })
}
```

## Error Handling Conventions

When handling errors, every `err := ... if (err != nil) return err` must be encapsulated using `fmt.Errorf`. This ensures that errors provide more context as they propagate up the call stack, making debugging easier.

For example, instead of:
```go
if err != nil {
    return err
}
```

Use:
```go
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}
```

For situations where a non-nil error should immediately halt execution, `errorz.Check` provides a concise and consistent way to panic. It ensures that if `err` is not `nil`, the program panics, and critically, it wraps the error in a `StructuredError` (if it isn't already) to provide consistent error reporting, and prevents panicking with a `nil` error.

Instead of:
```go
if err != nil {
    panic(err)
}
```

Use:
```go
errorz.Check(err)
```

### Running Tests

To run Unit and Remote tests, use the following command:

```bash
go test -run '^(TestUnit.*|TestRemote.*)$' ./...
```
