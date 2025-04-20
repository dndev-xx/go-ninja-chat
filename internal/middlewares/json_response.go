package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type jsonResponse struct {
	Data  any `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func JSONResponseMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					return c.JSON(he.Code, jsonResponse{
						Error: he.Message.(string),
						Data:  nil,
					})
				}
				return c.JSON(http.StatusInternalServerError, jsonResponse{
					Error: err.Error(),
					Data:  nil,
				})
			}

			responseData := c.Get("responseData")

			return c.JSON(http.StatusOK, jsonResponse{
				Data: responseData,
				Error: "",
			})
		}
	}
}
