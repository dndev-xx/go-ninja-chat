package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	sendmessage "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

func (h Handlers) PostV1SendMessage(eCtx echo.Context, params clientv1.PostV1SendMessageParams) error {
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
		ID:          params.XRequestID,
		ClientID:    userID,
		MessageBody: req.MessageBody,
	}

	us, err := h.sendMsg.Handle(ctx, reqUsecase)
	if err != nil {
		return err
	}

	resp := clientv1.SendMessageResponse{
		Data: &clientv1.MessageHeader{
			AuthorID:  &us.AuthorID,
			MessageID: &us.MessageID,
			CreatedAt: &us.CreatedAt,
		},
	}

	return eCtx.JSON(http.StatusOK, resp)
}
