package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func RecoveryMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Recovered from panic",
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
					)
					c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
				}
			}()
			return next(c)
		}
	}
}
