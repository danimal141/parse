package parse

import "fmt"

type APIError interface {
	error
	Code() int
	Message() string
}

type apiError struct {
	ErrorCode    int    `json:"code" parse:"code"`
	ErrorMessage string `json:"error" parse:"error"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("error %d - %s", e.ErrorCode, e.ErrorMessage)
}

func (e *apiError) Code() int {
	return e.ErrorCode
}

func (e *apiError) Message() string {
	return e.ErrorMessage
}
