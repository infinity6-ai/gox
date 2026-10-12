package errorz

import (
	"errors"
	"fmt"
)

// Data is the structured, machine-readable data of a StructuredError.
type Data struct {
	Code     int    `json:"code,omitempty"`
	Name     string `json:"name,omitempty"`
	Payload  string `json:"payload,omitempty"`
	Business bool   `json:"business,omitempty"`
	Cause    string `json:"cause"`
	Stack    string `json:"stack,omitempty"`
}

// Internal returns a sanitized copy of Data suitable for non-business internal errors,
// ensuring no sensitive internal details leak to the caller.
func (d *Data) Internal() *Data {
	return &Data{
		Code:     500,
		Name:     "InternalError",
		Cause:    "InternalError",
		Business: false,
	}
}

// From converts any error or panic value into structured *Data.
// If err is nil, it returns a generic internal error Data.
// If err is a StructuredError, it returns its Data with stack trace cleared.
// Otherwise, it wraps err as an internal error with stack trace cleared.
func From(err any) *Data {
	if err == nil {
		return (&Data{}).Internal()
	}
	e, ok := err.(error)
	if !ok {
		e = fmt.Errorf("%v", err)
	}
	se := As(e)
	d := se.Data()
	d.Stack = ""
	return d
}

// BusinessData sanitizes an error for business-level display. It strips sensitive
// information like stack traces and payloads, returning a machine-readable *Data struct.
//
//   - If the error is not a StructuredError, it's wrapped as a generic internal error.
//   - If it's a business-facing StructuredError, its data is returned, including the payload.
//   - If it's a non-business StructuredError, internal details are replaced with a
//     generic "InternalError" message, while retaining the original code and name.
func BusinessData(err error) *Data {
	if err == nil {
		return nil
	}

	var se StructuredError
	var data *Data

	if !errors.As(err, &se) {
		// Not a structured error. Wrap it as a generic internal error.
		data = Detail(500, "InternalError", "", false, err).Data()
	} else if se.Business() {
		// It's a business-facing error. Use its data directly, including the payload.
		data = se.Data()
	} else {
		// It's a non-business structured error. Sanitize it completely by replacing
		// the cause with a generic message, ensuring no internal details leak.
		data = Detail(se.Code(), se.Name(), "", false, errors.New("InternalError")).Data()
	}

	// Always strip the stack trace before returning to the caller.
	data.Stack = ""

	return data
}
