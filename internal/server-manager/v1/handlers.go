package v1

import (
	"context"
	"fmt"

	freeHands "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands"
	getFreeHandsBtnAvailability "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/handlers_mocks.gen.go -package=clientv1mocks
type getFreeHandsBtnAvailabilityUseCase interface {
	Handle(ctx context.Context, req getFreeHandsBtnAvailability.Request) (getFreeHandsBtnAvailability.Response, error)
}

type getFreeHands interface {
	Handle(ctx context.Context, req freeHands.Request) error
}

//go:generate options-gen -out-filename=handlers.gen.go -from-struct=Options
type Options struct {
	getFreeHandsBtnAvailability getFreeHandsBtnAvailabilityUseCase `option:"mandatory" validate:"required"`
	getFreeHands                getFreeHands                       `option:"mandatory" validate:"required"`
}

type Handlers struct {
	Options
}

func NewHandlers(opts Options) (*Handlers, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %v", err)
	}
	return &Handlers{Options: opts}, nil
}
