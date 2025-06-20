package errhandler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
)

//go:generate options-gen -out-filename=errhandler_options.gen.go -from-struct=Options
type Options struct {
	logger          *zap.Logger                                    `option:"mandatory" validate:"required"`
	productionMode  bool                                           `option:"mandatory"`
	responseBuilder func(code int, msg string, details string) any `option:"mandatory" validate:"required"`
}

type Handler struct {
	lg              *zap.Logger
	productionMode  bool
	responseBuilder func(code int, msg string, details string) any
}

func New(opts Options) (Handler, error) {
	if err := opts.Validate(); err != nil {
		return Handler{}, err
	}
	return Handler{
		lg:              opts.logger,
		productionMode:  opts.productionMode,
		responseBuilder: opts.responseBuilder,
	}, nil
}

func (h Handler) Handle(err error, eCtx echo.Context) {
	code, msg, details := internalerrors.ProcessServerError(err)
	httpStatus := http.StatusOK
	responseDetails := ""
	if !h.productionMode {
		responseDetails = details
	}
	response := h.responseBuilder(code, msg, responseDetails)
	if err := eCtx.JSON(httpStatus, response); err != nil {
		h.lg.Error("failed to send error response", zap.Error(err))
	}
}
