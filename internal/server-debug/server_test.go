package serverdebug_test

import (
	"context"
	"encoding/json"
	"io"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
)

func TestServer_LoggerLevel(t *testing.T) {
	// Arrange.
	err := logger.Init(logger.NewOptions("debug"))
	require.NoError(t, err)

	srv, err := serverdebug.New(serverdebug.NewOptions(":80"))
	require.NoError(t, err)

	testSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(testSrv.Close)

	logLevelURL := testSrv.URL + "/log/level"

	cases := []struct {
		name      string
		level     string
		expStatus int
	}{
		{
			name:      "success set debug",
			level:     "debug",
			expStatus: http.StatusOK,
		},
		{
			name:      "set info",
			level:     "info",
			expStatus: http.StatusOK,
		},
		{
			name:      "set warn",
			level:     "warn",
			expStatus: http.StatusOK,
		},
		{
			name:      "set error",
			level:     "error",
			expStatus: http.StatusOK,
		},
		{
			name:      "unsupported level",
			level:     "any_invalid_level",
			expStatus: http.StatusBadRequest,
		},
		{
			name:      "empty level",
			level:     "",
			expStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Action.
			status := setLevel(t, logLevelURL, tt.level)

			// Assert.
			require.Equal(t, tt.expStatus, status)

			if tt.expStatus == http.StatusOK {
				lvl := getLevel(t, logLevelURL)
				assert.Equal(t, tt.level, lvl)
			}
		})
	}
}

func setLevel(t *testing.T, url, level string) int {
    t.Helper()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    reqBody := fmt.Sprintf(`{"level":"%s"}`, level)
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url,
        io.NopCloser(strings.NewReader(reqBody)))
    require.NoError(t, err)

    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer func() { require.NoError(t, resp.Body.Close()) }()
    return resp.StatusCode
}


func getLevel(t *testing.T, url string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var data struct {
		Level string `json:"level"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)

	return data.Level
}
