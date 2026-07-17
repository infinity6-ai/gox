package errorz



// Panic panics with a StructuredError.
// If err is already a StructuredError, it's panicked directly.
// Otherwise, err is wrapped in a new StructuredError before panicking.
// If err is nil, Panic does nothing.
func Panic(err error) {
	if err == nil {
		return
	}
	err = As(err)
	panic(err)
}

// Check panics if the given error is not nil.
func Check(err error) {
	if err != nil {
		Panic(err)
	}
}
