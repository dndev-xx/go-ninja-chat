package v1

import (
	"context"
	"fmt"

	usecaseHist "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
	sendMessage "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=clientv1mocks

type getHistoryUseCase interface {
	Handle(ctx context.Context, req usecaseHist.Request) (usecaseHist.Response, error)
}

type sendMsgUseCase interface {
	Handle(ctx context.Context, req sendMessage.Request) (sendMessage.Response, error)
}

//go:generate options-gen -out-filename=handlers.gen.go -from-struct=Options
type Options struct {
	getHistory getHistoryUseCase 		`option:"mandatory" validate:"required"`
	sendMsg sendMsgUseCase 				`option:"mandatory"`
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (Handlers, error) {
	if err := opts.Validate(); err != nil {
		return Handlers{}, fmt.Errorf("validate options: %v", err)
	}
	return Handlers{Options: opts}, nil
}
