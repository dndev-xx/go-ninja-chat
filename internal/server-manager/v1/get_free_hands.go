package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	"github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1/pkg"
	freeHands "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands"
)

func (h *Handlers) PostFreeHands(eCtx echo.Context, params pkg.PostFreeHandsParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)
	err := h.getFreeHands.Handle(ctx, freeHands.Request{
		ManagerID: clientID,
	})
	if err != nil {
		return err
	}
	return eCtx.JSON(http.StatusOK, pkg.FreeHandsResponse{
		Data: nil,
	})
}
