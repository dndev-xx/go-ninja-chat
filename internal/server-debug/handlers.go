package serverdebug

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	mn "github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1/pkg"
)

func (s *Server) getOpenAPISpec(c echo.Context) error {
	swagger, err := pkg.GetSwagger()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error loading OpenAPI specification")
	}

	jsonData, err := json.Marshal(swagger)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error generating JSON")
	}

	return c.Blob(http.StatusOK, "application/json", jsonData)
}

func (s *Server) getOpenAPISpecManger(c echo.Context) error {
	swagger, err := mn.GetSwagger()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error loading OpenAPI specification")
	}
	jsonData, err := json.Marshal(swagger)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error generating JSON")
	}

	return c.Blob(http.StatusOK, "application/json", jsonData)
}

func (s *Server) AllLogs(c echo.Context) error {
	return c.String(http.StatusOK, s.logBuffer.String())
}
