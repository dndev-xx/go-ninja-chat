package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	defaultMsg = "something went wrong"
)

// ServerError is used to return custom error codes to client.
type ServerError struct {
	Code    int
	Message string
	cause   error
}

func NewServerError(code int, msg string, err error) *ServerError {
	return &ServerError{
		Code:    code,
		Message: msg,
		cause:   err,
	}
}

func (s *ServerError) Error() string {
	if s.cause != nil {
		return fmt.Sprintf("%s: %v", s.Message, s.cause)
	}
	return s.Message
}

func (s *ServerError) Unwrap() error {
	return s.cause
}

func GetServerErrorCode(err error) int {
	code, _, _ := ProcessServerError(err)
	return code
}

// ProcessServerError tries to retrieve from given error its code, message and some details.
// For example, that fields can be used to build error response for client.
func ProcessServerError(err error) (code int, msg string, details string) {
	if err == nil {
		return http.StatusOK, "", ""
	}

	var serverErr *ServerError
	if errors.As(err, &serverErr) {
		return serverErr.Code, serverErr.Message, serverErr.Error()
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		msg := fmt.Sprint(httpErr.Message)
		return httpErr.Code, msg, fmt.Sprintf("code=%d, message=%s", httpErr.Code, msg)
	}

	return http.StatusInternalServerError, "something went wrong", err.Error()
}
