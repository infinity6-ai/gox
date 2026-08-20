package constraintz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/constraintz"
)

// This function is a compile-time check to ensure the constraints are valid.
// It is not intended to be a runtime test.
func acceptsNumber[T constraintz.Numbers](_ T) bool {
	return true
}

func TestUnitCompileTimeConstraints(t *testing.T) {
	// This test simply calls the generic function with different types
	// to ensure the constraints work as expected at compile time.
	// If the code compiles, the test passes.
	if !acceptsNumber(1) {
		t.Error("expected true for int")
	}

	if !acceptsNumber(uint(1)) {
		t.Error("expected true for uint")
	}

	if !acceptsNumber(1.0) {
		t.Error("expected true for float64")
	}
}
