package gethistory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	gethistory "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
)

func TestRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		request gethistory.Request
		wantErr bool
	}{
		// Positive.
		{
			name: "cursor specified",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 0,
				Cursor:   "eyJwYWdlX3NpemUiOjUwLCJsYXN0IjoxNjcwNTAyNTAyfQ==", // {"page_size":50,"last":1670502502}
			},
			wantErr: false,
		},
		{
			name: "page size specified",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 50,
				Cursor:   "",
			},
			wantErr: false,
		},

		// Negative.
		{
			name: "neither cursor nor pagesize specified",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 0,
				Cursor:   "",
			},
			wantErr: true,
		},
		{
			name: "cursor and pagesize specified",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 50,
				Cursor:   "eyJwYWdlX3NpemUiOjUwLCJsYXN0IjoxNjcwNTAyNTAyfQ==", // {"page_size":50,"last":1670502502}
			},
			wantErr: true,
		},
		{
			name: "require request id",
			request: gethistory.Request{
				ID:       types.RequestIDNil,
				ClientID: types.NewUserID(),
				PageSize: 10,
				Cursor:   "",
			},
			wantErr: true,
		},
		{
			name: "require client id",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.UserIDNil,
				PageSize: 10,
				Cursor:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid cursor encoding",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 0,
				Cursor:   "eyJwYWdlX3NpemUiOjUwLCJs YXN0IjoxNjcwNTAyNTAyfQ", // With space.
			},
			wantErr: true,
		},
		{
			name: "too small page size",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 9,
				Cursor:   "",
			},
			wantErr: true,
		},
		{
			name: "too big page size",
			request: gethistory.Request{
				ID:       types.NewRequestID(),
				ClientID: types.NewUserID(),
				PageSize: 101,
				Cursor:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
