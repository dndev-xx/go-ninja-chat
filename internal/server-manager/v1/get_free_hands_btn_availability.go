package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	"github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1/pkg"
	canreceiveproblems "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability"
)

func (h *Handlers) PostGetFreeHandsBtnAvailability(eCtx echo.Context, params pkg.PostGetFreeHandsBtnAvailabilityParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)
	resp, err := h.getFreeHandsBtnAvailability.Handle(ctx, canreceiveproblems.Request{
		ID:        params.XRequestID,
		ManagerID: clientID,
	})
	if err != nil {
		return err
	}
	return eCtx.JSON(http.StatusOK, pkg.GetFreeHandsBtnAvailabilityResponse{
		Data: resp,
	})
}
