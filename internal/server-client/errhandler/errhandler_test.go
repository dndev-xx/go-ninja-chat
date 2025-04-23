package errhandler_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
	"github.com/dndev-xx/go-ninja-chat/internal/server-client/errhandler"
)

func TestHandler_Handle_InDevMode(t *testing.T) {
	const productionMode = false

	cases := []struct {
		name    string
		err     error
		expResp string
	}{
		{
			name:    "echo error",
			err:     echo.NewHTTPError(http.StatusBadRequest, "bind request"),
			expResp: `{"code": 400, "message": "bind request", "details": "code=400, message=bind request"}`,
		},
		{
			name: "custom error",
			err: internalerrors.NewServerError(
				1000,
				"create chat error",
				fmt.Errorf("create chat: find client: no user found with id 1234: %w", io.EOF),
			),
			expResp: `{"code": 1000, "message": "create chat error",
"details": "create chat error: create chat: find client: no user found with id 1234: EOF"}`,
		},
		{
			name:    "unknown error",
			err:     fmt.Errorf("cannot handle usecase: %w", context.Canceled),
			expResp: `{"code": 500, "message": "something went wrong", "details": "cannot handle usecase: context canceled"}`,
		},
	}

	h, err := errhandler.New(errhandler.NewOptions(zap.L(), productionMode, respBuilder))
	require.NoError(t, err)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp, ctx := newEchoCtx()
			h.Handle(tt.err, ctx)

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.JSONEq(t, tt.expResp, string(b))
		})
	}
}

func TestHandler_Handle_InProductionMode(t *testing.T) {
	const productionMode = true

	cases := []struct {
		name    string
		err     error
		expResp string
	}{
		{
			name:    "echo error",
			err:     echo.NewHTTPError(http.StatusBadRequest, "bind request"),
			expResp: `{"code": 400, "message": "bind request"}`,
		},
		{
			name: "custom error",
			err: internalerrors.NewServerError(
				1000,
				"create chat error",
				fmt.Errorf("create chat: find client: no user found with id 1234: %w", io.EOF),
			),
			expResp: `{"code": 1000, "message": "create chat error"}`,
		},
		{
			name:    "unknown error",
			err:     fmt.Errorf("cannot handle usecase: %w", context.Canceled),
			expResp: `{"code": 500, "message": "something went wrong"}`,
		},
	}

	h, err := errhandler.New(errhandler.NewOptions(zap.L(), productionMode, respBuilder))
	require.NoError(t, err)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			resp, ctx := newEchoCtx()
			h.Handle(tt.err, ctx)

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.JSONEq(t, tt.expResp, string(b))
		})
	}
}

func newEchoCtx() (resp *httptest.ResponseRecorder, ctx echo.Context) {
	req := httptest.NewRequest(http.MethodPost, "/getHistory", bytes.NewBufferString(`{"pageSize": 30, "cursor": ""}`))
	resp = httptest.NewRecorder()
	ctx = echo.New().NewContext(req, resp)
	return resp, ctx
}

type Error struct {
	Code    int     `json:"code"`
	Details *string `json:"details,omitempty"`
	Message string  `json:"message"`
}

var respBuilder = func(code int, msg string, details string) any {
	var d *string
	if details != "" {
		d = &details
	}
	return Error{
		Code:    code,
		Details: d,
		Message: msg,
	}
}
