package apperror

import "net/http"

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func New(status int, message string) *Error {
	return &Error{Status: status, Message: message}
}

func NotFound(message string) *Error     { return New(http.StatusNotFound, message) }
func Conflict(message string) *Error     { return New(http.StatusConflict, message) }
func Unauthorized(message string) *Error { return New(http.StatusUnauthorized, message) }
func Forbidden(message string) *Error    { return New(http.StatusForbidden, message) }
func Unprocessable(message string) *Error {
	return New(http.StatusUnprocessableEntity, message)
}
