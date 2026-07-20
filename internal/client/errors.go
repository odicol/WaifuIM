package client

import (
	"errors"
	"fmt"
)

type BadResponse struct {
	StatusCode   int
	ContentType  string
	ResponseBody string
}

func (e *BadResponse) Error() string {
	var ct, b string
	if e.ContentType == "" {
		ct = "N/A"
	} else {
		ct = e.ContentType
	}
	if e.ResponseBody == "" {
		b = "N/A"
	} else {
		b = e.ResponseBody
	}
	return fmt.Sprintf("bad response %d - content-type: %s, body: %s", e.StatusCode, ct, b)
}

func (e *BadResponse) Is(target error) bool {
	var t *BadResponse
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode &&
		e.ContentType == t.ContentType &&
		e.ResponseBody == t.ResponseBody
}
