//go:build integration

package keycloakclient

import (
	"context"
	"fmt"
	"net/http"
)

type RPT struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func (c *Client) Auth(ctx context.Context, username, password string) (*RPT, error) {
	url := fmt.Sprintf("realms/%s/protocol/openid-connect/token", c.realm)

	var token RPT

	resp, err := c.auth(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"username":   username,
			"password":   password,
			"grant_type": "password",
		}).
		SetResult(&token).
		Post(url)
	if err != nil {
		return nil, fmt.Errorf("send request to keycloak: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("errored keycloak response: %v", resp.Status())
	}

	return &token, nil
}
