package keycloakclient_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
)

func TestIntrospectTokenResult(t *testing.T) {
	cases := []struct {
		name string
		in   string
		exp  keycloakclient.IntrospectTokenResult
	}{
		{
			name: "no aud",
			in: `
{
   "exp": 1662047857,
   "iat": 1662047557,
   "active": true
}`,
			exp: keycloakclient.IntrospectTokenResult{
				Exp:    1662047857,
				Iat:    1662047557,
				Aud:    nil,
				Active: true,
			},
		},

		{
			name: "aud is string",
			in: `
{
   "exp": 1662057857,
   "iat": 1662057557,
   "aud": "account",
   "active": true
}`,
			exp: keycloakclient.IntrospectTokenResult{
				Exp:    1662057857,
				Iat:    1662057557,
				Aud:    []string{"account"},
				Active: true,
			},
		},

		{
			name: "aud is array of string",
			in: `
{
   "exp": 1662087857,
   "iat": 1662087557,
   "aud": [
      "chat-ui-client",
      "account"
   ],
   "active": false
}`,
			exp: keycloakclient.IntrospectTokenResult{
				Exp:    1662087857,
				Iat:    1662087557,
				Aud:    []string{"chat-ui-client", "account"},
				Active: false,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var r keycloakclient.IntrospectTokenResult
			require.NoError(t, json.Unmarshal([]byte(tt.in), &r))
			assert.Equal(t, tt.exp, r)
		})
	}
}
