package v1

import (
	"net/http"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	sendmessage "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
	"github.com/labstack/echo/v4"
)

func (h Handlers) PostSendMessage(eCtx echo.Context, params clientv1.PostSendMessageParams) error {
    ctx := eCtx.Request().Context()
    userID := middlewares.MustUserID(eCtx)

    var req clientv1.SendMessageRequest
    if err := eCtx.Bind(&req); err != nil {
        return err
    }

    if err := validator.Validator.Struct(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    reqUsecase := sendmessage.Request{
        ID:         params.XRequestID,
        ClientID:   userID,
        MessageBody: req.MessageBody,
    }

    resp, err := h.sendMsg.Handle(ctx, reqUsecase)
    if err != nil {
        return err
    }

    return eCtx.JSON(http.StatusOK, resp)
}
