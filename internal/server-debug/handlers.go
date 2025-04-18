package serverdebug

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oasdiff/yaml"
)

func (s *Server) getOpenAPISpec(c echo.Context) error {
	yamlFile, err := ioutil.ReadFile("./api/client.v1.swagger.yaml")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error reading OpenAPI specification file")
	}

	var spec interface{}
	if err := yaml.Unmarshal(yamlFile, &spec); err != nil {
		return c.String(http.StatusInternalServerError, "Error parsing YAML")
	}

	jsonData, err := json.Marshal(spec)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error generating JSON")
	}

	return c.Blob(http.StatusOK, "application/json", jsonData)
}