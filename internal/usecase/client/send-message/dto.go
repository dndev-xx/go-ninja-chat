package sendmessage

import (
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

type Request struct {
	ID          types.RequestID 	`validate:"required"`
	ClientID    types.UserID		`validate:"required"`
	MessageBody string				`validate:"required,min=1,max=4000"`
}

func (r Request) Validate() error {
	if err := validator.Validator.Struct(r); err != nil {
		return err
	}
	return nil
}

type Response struct {
	AuthorID  types.UserID  		`json:"authorID"`
	MessageID types.MessageID 		`json:"messageID"`
	CreatedAt time.Time				`json:"createdAt"`
}
