package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"github.com/mssola/useragent"
)

func LoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			ip := c.Request().RemoteAddr
			method := c.Request().Method
			path := c.Request().URL.Path
			userAgent := c.Request().UserAgent()
			ua := useragent.New(userAgent)
			browserName, _ := ua.Browser()
			host := c.Request().Host
			requestID := c.Request().Header.Get(echo.HeaderXRequestID) 

			err := next(c)

			latency := time.Since(start)
			status := c.Response().Status

			logger.Info("Incoming request client server",
				zap.String("remote_ip", ip),
				zap.String("host", host),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("request_id", requestID),
				zap.String("user_agent", userAgent),
				zap.String("os", ua.OS()),
				zap.String("browser", browserName),
				zap.Duration("latency", latency),
				zap.Int("status", status),
			)

			return err
		}
	}
}
