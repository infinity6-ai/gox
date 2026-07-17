package errorz

import "errors"

// Data is the structured, machine-readable data of a StructuredError.
type Data struct {
	Code     int    `json:"code,omitempty"`
	Name     string `json:"name,omitempty"`
	Payload  string `json:"payload,omitempty"`
	Business bool   `json:"business,omitempty"`
	Cause    string `json:"cause"`
	Stack    string `json:"stack,omitempty"`
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
