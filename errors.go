package parse

import "fmt"

type APIError interface {
	error
	Code() int
	Message() string
}

type apiErrorT struct {
	ErrorCode    int    `json:"code" parse:"code"`
	ErrorMessage string `json:"error" parse:"error"`
}

func (e *apiErrorT) Error() string {
	return fmt.Sprintf("error %d - %s", e.ErrorCode, e.ErrorMessage)
}

func (e *apiErrorT) Code() int {
	return e.ErrorCode
}

func (e *apiErrorT) Message() string {
	return e.ErrorMessage
}
