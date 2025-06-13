package getfreehands

import (
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

type Request struct {
	ManagerID types.UserID `validate:"required"`
}

func (r Request) Validate() error {
	if err := validator.Validator.Struct(r); err != nil {
		return err
	}
	return nil
}
