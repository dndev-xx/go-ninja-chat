package wshandshake

import (
	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/websocket-stream"
)

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	httpUpgrade *websocketstream.HTTPHandler
}

type UseCase struct {
	Options
}

func New(opt Options) (*UseCase, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &UseCase{
		Options: opt,
	}, nil
}

func (u *UseCase) Handle(ctx echo.Context) error {
	u.httpUpgrade.Serve(ctx)
	return nil
}
