package getfreehandsbtnavailability

import (
	"context"
	"errors"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrIdempotencyKey   = errors.New("invalid idempotency key")
	ErrManagerNotExists = errors.New("manager not exists")
	ErrManagerBusy      = errors.New("manager busy")
)

//go:generate mockgen -source=$GOFILE -destination=mocks/usecase_mocks.gen.go -package=managerv1mocks
type managerLoadService interface {
	CanManagerTakeProblem(ctx context.Context, managerID types.UserID) (bool, error)
}

type managerPool interface {
	Contains(ctx context.Context, managerID types.UserID) (bool, error)
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	managerLoadService managerLoadService
	managerPool        managerPool
}

type UseCase struct {
	Options
}

func New(opts Options) (*UseCase, error) {
	return &UseCase{Options: opts}, opts.Validate()
}

func (u *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	// FIXME: 1) Если менеджер есть в пуле менеджеров, то возвращаем false (он уже жмакал кнопку ранее).
	// FIXME: 2) Иначе строим ответ на базе факта, может ли менеджер брать новую проблему в принципе.
	if err := req.Validate(); err != nil {
		return Response{}, ErrInvalidRequest
	}
	resolution, err := u.managerPool.Contains(ctx, req.ManagerID)
	if err != nil {
		return Response{
			Result: resolution,
		}, ErrManagerNotExists
	}
	if resolution {
		return Response{
			Result: !resolution,
		}, nil
	}
	resolution, err = u.managerLoadService.CanManagerTakeProblem(ctx, req.ManagerID)
	if err != nil {
		return Response{
			Result: resolution,
		}, ErrManagerBusy
	}
	return Response{
		Result: resolution,
	}, nil
}
