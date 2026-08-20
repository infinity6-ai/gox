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
