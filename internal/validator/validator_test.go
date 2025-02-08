package validator_test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
)

type options struct {
	DB      *sql.DB      `validate:"required"`
	Handler http.Handler `validate:"required"`
}

func TestValidate_TrickyNils(t *testing.T) {
	cases := []struct {
		in      options
		wantErr bool
	}{
		// Negative.
		{
			in:      options{DB: nil, Handler: new(handlerMock)},
			wantErr: true,
		},
		{
			in:      options{DB: new(sql.DB), Handler: http.HandlerFunc(nil)},
			wantErr: true,
		},
		{
			in:      options{DB: new(sql.DB), Handler: (*handlerMock)(nil)},
			wantErr: true,
		},

		// Positive.
		{
			in:      options{DB: new(sql.DB), Handler: new(handlerMock)},
			wantErr: false,
		},
		{
			in: options{
				DB:      new(sql.DB),
				Handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			err := validator.Validator.Struct(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

var _ http.Handler = (*handlerMock)(nil)

type handlerMock struct{}

func (h *handlerMock) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
}
