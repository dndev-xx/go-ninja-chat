package serverdebug //nolint:testpackage // special hack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Server) Handler() http.Handler {
	return s.srv.Handler
}

func TestServerVersion(t *testing.T) {
	server, err := New(nil, NewOptions(":8080"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	resp := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, resp)

	err = server.Version(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "go_version")
	assert.Contains(t, response, "path")
	assert.Contains(t, response, "main")
	assert.Contains(t, response, "dependencies")
	assert.Contains(t, response, "settings")
}

func TestServerGetLogLevelHandler(t *testing.T) {
	server, err := New(nil, NewOptions(":8080"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/log/level", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err = server.getLogLevelHandler(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "level")
}

func TestServerLogLevelHandler(t *testing.T) {
	server, err := New(nil, NewOptions(":8080"))
	require.NoError(t, err)

	reqBody := `{"level": "debug"}`
	req := httptest.NewRequest(http.MethodPut, "/log/level", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	err = server.logLevelHandler(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "status")
	assert.Equal(t, "ok", response["status"])
}

func TestServerSetupPprof(t *testing.T) {
	server, err := New(nil, NewOptions(":8080"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	_ = e.NewContext(req, rec)

	handler := server.Handler()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
