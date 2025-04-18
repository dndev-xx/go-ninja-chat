package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"github.com/mssola/useragent"
)

func LoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.Request().RemoteAddr
			method := c.Request().Method
			path := c.Request().URL.Path
			userAgent := c.Request().UserAgent()
			ua := useragent.New(userAgent)
			browserName, _ := ua.Browser()
			logger.Info("Incoming request client server",
				zap.String("ip", ip),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("os", ua.OS()),
				zap.String("browser", browserName),
			)
			return next(c)
		}
	}
}