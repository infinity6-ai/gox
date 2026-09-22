package bqclienterror

import (
	"errors"

	"google.golang.org/api/googleapi"
)

func ParseErrorCode(err error) int {
	if err == nil {
		return -1
	}
	var e *googleapi.Error
	ok := errors.As(err, &e)
	if !ok {
		return -1
	}
	return e.Code
}
