package errorz

import (
	"errors"
	"runtime"
)

// Panic panics with a StructuredError.
// If err is already a StructuredError, it's panicked directly.
// Otherwise, err is wrapped in a new StructuredError before panicking.
// If err is nil, Panic does nothing.
func Panic(err error) {
	if err == nil {
		err = Detail(500, "NulPanic", "", false, errors.New("panic with nil not allowed"))
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

type PanicVal struct {
	value any
}

func (p *PanicVal) Error() string {
	return "panic"
}

func (p *PanicVal) Value() any {
	return p.value
}

func (p *PanicVal) Unwrap() error {
	err, ok := p.value.(error)
	if !ok {
		return nil
	}
	return err
}

type PanicError interface {
	error
	Value() any
}

func Unpanic(fn func()) (ret StructuredError) {
	didPanic := true
	defer func() {
		if !didPanic {
			return
		}
		recovered := recover()
		// In Go 1.21+, panic(nil) is recovered as a *runtime.PanicNilError.
		// We normalize this to a nil value to maintain consistent behavior.
		if _, ok := recovered.(*runtime.PanicNilError); ok {
			recovered = nil
		}
		err, ok := recovered.(error)
		if !ok {
			err = &PanicVal{value: recovered}
		}
		ret = Detail(500, "PanicRecoveredError", "", false, err)
	}()

	fn()
	didPanic = false
	return
}
