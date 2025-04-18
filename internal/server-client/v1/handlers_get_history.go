package v1

import (
	"net/http"
	"time"

	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var stub = clientv1.MessagesPage{Messages: []clientv1.Message{
	{
		AuthorId:  uuid.New(),
		Body:      "Здравствуйте! Разберёмся.",
		CreatedAt: time.Now(),
		Id:        uuid.New(),
	},
	{
		AuthorId:  uuid.MustParse("7d67b14d-221e-4499-9be2-6707d7df1adc"),
		Body:      "Привет! Не могу снять денег с карты,\nпишет 'карта заблокирована'",
		CreatedAt: time.Now().Add(-time.Minute),
		Id:        uuid.New(),
	},

}, TotalCount: 2,}

func (h Handlers) PostGetHistory(eCtx echo.Context, params clientv1.PostGetHistoryParams)error {
	return eCtx.JSON(http.StatusOK, stub)
}
