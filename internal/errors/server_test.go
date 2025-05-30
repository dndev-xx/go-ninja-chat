package errors_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
)

func TestServerError(t *testing.T) {
	err := internalerrors.NewServerError(
		4242,
		"cannot handle something",
		fmt.Errorf("closed: %w", context.Canceled),
	)
	assert.Equal(t, 4242, err.Code)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestProcessServerError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		expCode    int
		expMsg     string
		expDetails string
	}{
		{
			name:       "echo error",
			err:        echo.NewHTTPError(http.StatusBadRequest, "bind request"),
			expCode:    http.StatusBadRequest,
			expMsg:     "bind request",
			expDetails: "code=400, message=bind request",
		},
		{
			name: "custom error",
			err: internalerrors.NewServerError(
				1000,
				"create chat error",
				fmt.Errorf("create chat: find client: no user found with id 1234: %w", io.EOF),
			),
			expCode:    1000,
			expMsg:     "create chat error",
			expDetails: "create chat error: create chat: find client: no user found with id 1234: EOF",
		},
		{
			name:       "unknown error",
			err:        fmt.Errorf("cannot handle usecase: %w", context.Canceled),
			expCode:    http.StatusInternalServerError,
			expMsg:     "something went wrong",
			expDetails: "cannot handle usecase: context canceled",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			code, msg, details := internalerrors.ProcessServerError(tt.err)
			assert.Equal(t, tt.expCode, code)
			assert.Equal(t, tt.expMsg, msg)
			assert.Equal(t, tt.expDetails, details)
		})
	}
}
