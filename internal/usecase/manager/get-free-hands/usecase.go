package getfreehands

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
type managerPool interface {
	Put(ctx context.Context, managerID types.UserID) error
}

//go:generate options-gen -out-filename=usecase_options.gen.go -from-struct=Options
type Options struct {
	managerPool managerPool
}

type UseCase struct {
	Options
}

func New(opts Options) (*UseCase, error) {
	return &UseCase{Options: opts}, opts.Validate()
}

func (u *UseCase) Handle(ctx context.Context, req Request) error {
	if err := req.Validate(); err != nil {
		return ErrInvalidRequest
	}
	if err := u.managerPool.Put(ctx, req.ManagerID); err != nil {
		return err
	}
	return nil
}
