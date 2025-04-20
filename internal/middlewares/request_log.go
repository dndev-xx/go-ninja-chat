package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewRequestLogger(lg *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().Method == http.MethodOptions
		},
		LogValuesFunc: func(eCtx echo.Context, v middleware.RequestLoggerValues) error {
			lg := lg.With(
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("host", v.Host),
				zap.String("method", v.Method),
				zap.String("path", v.URIPath),
				zap.String("request_id", v.RequestID),
				zap.String("user_agent", v.UserAgent),
				zap.Int("status", v.Status),
			)

			uid, _ := userID(eCtx)
			lg = lg.With(zap.Stringer("user_id", uid))

			if err := v.Error; err != nil {
				lg = lg.With(zap.Error(err))
			}

			switch s := v.Status; {
			case s >= 500:
				lg.Error("server error")
			case s >= 400:
				lg.Error("client error")
			default:
				lg.Info("success")
			}

			return nil
		},
		LogLatency:   true,
		LogRemoteIP:  true,
		LogHost:      true,
		LogMethod:    true,
		LogURIPath:   true,
		LogRequestID: true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
	})
}
