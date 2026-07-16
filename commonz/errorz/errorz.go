package errorz

type WrapperError struct {
	Cause   error
	Payload any
}

func (e *WrapperError) Error() string {
	return e.Cause.Error()
}

func New(code int, payload any, err error) *WrapperError {
	return &WrapperError{
		Cause:   err,
		Payload: payload,
	}
}

// func Newf(format string, a ...any) *Errorz {
// 	return &Errorz{
// 		Code:   code,
// 		Params: params,
// 		Error:  fmt.Errorf(format, a...),
// 	}
// }

// type Errorz struct {
// 	Message string `json:"message"`
// 	Code    string `json:"code"`
// 	Details any    `json:"details,omitempty"`
// 	Cause   error  `json:"-"`
// }

// func New(code, message string) *Errorz {
// 	return &Errorz{
// 		Code:    code,
// 		Message: message,
// 	}
// }

// func Newf(code, format string, a ...any) *Errorz {
// 	return &Errorz{
// 		Code:    code,
// 		Message: fmt.Sprintf(format, a...),
// 	}
// }

// func Wrap(err error, code, message string) *Errorz {
// 	return &Errorz{
// 		Code:    code,
// 		Message: message,
// 		Cause:   err,
// 	}
// }

// func Wrapf(err error, code, format string, a ...any) *Errorz {
// 	return &Errorz{
// 		Code:    code,
// 		Message: fmt.Sprintf(format, a...),
// 		Cause:   err,
// 	}
// }

// func (e *Errorz) Error() string {
// 	if e.Cause != nil {
// 		return fmt.Sprintf("code: %s, message: %s, cause: %v", e.Code, e.Message, e.Cause)
// 	}
// 	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
// }

// func (e *Errorz) Unwrap() error {
// 	return e.Cause
// }

// func GetCode(err error) string {
// 	if err == nil {
// 		return ""
// 	}
// 	if appErr, ok := err.(*Errorz); ok {
// 		return appErr.Code
// 	}
// 	return ""
// }

// func GetMessage(err error) string {
// 	if err == nil {
// 		return ""
// 	}
// 	if appErr, ok := err.(*Errorz); ok {
// 		return appErr.Message
// 	}
// 	return err.Error()
// }

// func IsCode(err error, code string) bool {
// 	if err == nil {
// 		return false
// 	}
// 	if appErr, ok := err.(*Errorz); ok {
// 		return appErr.Code == code
// 	}
// 	return false

// }
