package gethistory

import (
	"errors"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

// Tags https://github.com/go-playground/validator?tab=readme-ov-file#other
type Request struct {
	ID       types.RequestID `validate:"required"`
	ClientID types.UserID    `validate:"required"`
	PageSize int             `validate:"omitempty,gte=10,lte=100,excluded_with=Cursor"`
	Cursor   string          `validate:"omitempty,base64url,excluded_with=PageSize"`
}

func (r Request) Validate() error {
	if err := validator.Validator.Struct(r); err != nil {
		return err
	}
	if r.Cursor == "" && r.PageSize == 0 {
        return errors.New("either cursor or page size must be specified")
    }

    return nil
}

type Response struct {
	Messages 		[]Message 	`json:"messages"`
	NextCursor 		string 		`json:"next_cursor"`
}

type Message struct {
	ID                   types.MessageID 	`json:"id"`
	AuthorID             types.UserID 		`json:"author_id"`
	Body                 string 			`json:"body"`
	IsBlocked            bool 				`json:"is_blocked"`
	IsService            bool 				`json:"is_service"`
	IsReceived			 bool 				`json:"is_received"`
	CreatedAt            time.Time 			`json:"created_at"`
}
