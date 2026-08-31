# Generate module by IA

You should generate the same validation function from `validation` module with panic instead of return error.

Generate only `validation/checker/checker.go` and `validation/checker/checker_test.go` with all validation functions.

Functions must have the same signature without returning error. Use `errorz.Check` as a helper to check error and panic if it is not nil.

Sample:

```go
package checker

import (
    "github.com/infinity6-ai/gox/commonz/errorz"
    "github.com/infinity6-ai/gox/commonz/validation"
)

func Equal[T comparable](expected T, actual T, msg string, args ...any) {
    errorz.Check(validation.Equal(expected, actual), msg, args)
}
```

Follow `GEMINI.md` directives to generate tests

