package managerload

import (
	"context"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/service_mocks.gen.go -package=clientv1mocks
type problemsRepository interface {
	GetManagerOpenProblemsCount(ctx context.Context, managerID types.UserID) (int, error)
}

//go:generate options-gen -out-filename=service.gen.go -from-struct=Options
type Options struct {
	maxProblemsAtTime int                `option:"mandatory" validate:"min=1,max=30"`
	problemsRepo      problemsRepository `option:"mandatory" validate:"required"`
}

type Service struct {
	Options
}

func New(opts Options) (*Service, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		Options: opts,
	}, nil
}
