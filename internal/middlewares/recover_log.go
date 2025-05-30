package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func NewRecovery(lg *zap.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: true,
		LogErrorFunc: func(_ echo.Context, err error, stack []byte) error {
			lg.With(
				zap.Error(err),
				zap.String("stack", string(stack)),
			).Error("panic recovered")
			return err
		},
	})
}
