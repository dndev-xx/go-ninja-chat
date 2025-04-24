package v1

import (
	"net/http"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	view "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/views"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	usecase "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
	"github.com/dndev-xx/go-ninja-chat/pkg/pointer"
	"github.com/labstack/echo/v4"
)

func (h Handlers) PostGetHistory(eCtx echo.Context, params clientv1.PostGetHistoryParams) error {
	ctx := eCtx.Request().Context()
	clientID := middlewares.MustUserID(eCtx)
	var req clientv1.GetHistoryRequest
	if err := eCtx.Bind(&req); err != nil {
		return err
	}
	reqUsecase := usecase.Request {
		ID: types.RequestID(params.XRequestID),
		ClientID: types.UserID(pointer.Indirect(clientID)),
		PageSize: pointer.Indirect(req.PageSize),
		Cursor: pointer.Indirect(req.Cursor),
	}
	history, err := h.getHistory.Handle(ctx, reqUsecase)
	if err != nil {
		return err
	}
	msg := make([]clientv1.Message, 0)
	for _,cur := range history.Messages {
		msg = append(msg, view.ConvertOriginalMessageToMessage(cur))
	}
	response := clientv1.GetHistoryResponse{
		Data: clientv1.MessagesPage{
			Messages: msg,
			TotalCount: len(msg),
		},
	}

	//eCtx.Set("responseData", response)
	return eCtx.JSON(http.StatusOK, response)
}
