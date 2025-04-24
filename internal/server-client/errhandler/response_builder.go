package errhandler

import (
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
)

type Response struct {
	Error clientv1.Error `json:"error"`
}

var ResponseBuilder = func(code int, msg string, details string) any {
	var currentDetails *string
	if details != "" {
		currentDetails = &details
	}
	return Response{
		Error: clientv1.Error {
			Code: code,
			Message: msg,
			Details: currentDetails,
		},
	}
}
