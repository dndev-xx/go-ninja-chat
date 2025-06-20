package views

import (
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	usecase "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
)

func ConvertOriginalMessageToMessage(origMsg usecase.Message) clientv1.Message {
	return clientv1.Message{
		AuthorId:  origMsg.AuthorID,
		Body:      origMsg.Body,
		CreatedAt: origMsg.CreatedAt,
		Id:        origMsg.ID,
	}
}
