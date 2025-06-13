package getfreehandsbtnavailability_test

import (
	"testing"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
	getFreeHandsBtnAvailability "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability"
)

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		r       getFreeHandsBtnAvailability.Request
		wantErr bool
	}{
		{
			name: "valid",
			r: getFreeHandsBtnAvailability.Request{
				ID:        types.NewRequestID(),
				ManagerID: types.NewUserID(),
			},
			wantErr: false,
		},
		{
			name: "invalid",
			r: getFreeHandsBtnAvailability.Request{
				ID:        types.RequestIDNil,
				ManagerID: types.UserIDNil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
