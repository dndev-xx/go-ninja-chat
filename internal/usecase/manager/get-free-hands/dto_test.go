package getfreehands_test

import (
	"testing"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	freehands "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands"
)

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request freehands.Request
		wantErr bool
	}{
		{
			name: "valid",
			request: freehands.Request{
				ManagerID: types.NewUserID(),
			},
			wantErr: false,
		},
		{
			name: "invalid: empty manager id",
			request: freehands.Request{
				ManagerID: types.UserIDNil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
