package errorz

import "errors"

// Panic panics with a StructuredError.
// If err is already a StructuredError, it's panicked directly.
// Otherwise, err is wrapped in a new StructuredError before panicking.
// If err is nil, Panic does nothing.
func Panic(err error) {
	if err == nil {
		return
	}
	var se StructuredError
	if errors.As(err, &se) {
		panic(se)
	} else {
		panic(Detail(0, "", "", false, err))
	}
}

// Check panics if the given error is not nil.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}
