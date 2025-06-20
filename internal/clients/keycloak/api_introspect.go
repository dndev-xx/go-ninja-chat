package keycloakclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type AudArray []string

func (a *AudArray) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '"' {
		var singleAud string
		if err := json.Unmarshal(data, &singleAud); err != nil {
			return err
		}
		*a = AudArray{singleAud}
		return nil
	}
	var audList []string
	if err := json.Unmarshal(data, &audList); err != nil {
		return err
	}
	*a = AudArray(audList)
	return nil
}

type IntrospectTokenResult struct {
	Exp           int      `json:"exp"`
	Iat           int      `json:"iat"`
	Aud           AudArray `json:"aud"`
	Active        bool     `json:"active"`
	Username      string   `json:"username"`
	ClientID      string   `json:"client_id"`
	TokenType     string   `json:"token_type"`
	Authorization struct {
		Permissions []struct {
			Scopes []string `json:"scopes"`
			Rsid   string   `json:"rsid"`
			Rsname string   `json:"rsname"`
		} `json:"permissions"`
	} `json:"authorization"`
}

// IntrospectToken implements
// https://www.keycloak.org/docs/latest/authorization_services/index.html#obtaining-information-about-an-rpt
func (c *Client) IntrospectToken(ctx context.Context, token string) (*IntrospectTokenResult, error) {
	// Проверяем базовые параметры
	if c.BasePath == "" || c.Realm == "" || c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("invalid client configuration: basepath, realm, clientid and clientsecret must be set")
	}

	baseURL, err := url.Parse(c.BasePath)
	if err != nil {
		return nil, fmt.Errorf("invalid basepath URL: %v", err)
	}

	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect",
		baseURL.String(), url.PathEscape(c.Realm))

	resp, err := c.auth(ctx).
		SetFormData(map[string]string{
			"token":         token,
			"client_id":     c.ClientID,
			"client_secret": c.ClientSecret,
		}).
		Post(introspectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("introspection failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	var result IntrospectTokenResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal introspection response: %v", err)
	}
	if !result.Active {
		return &result, fmt.Errorf("token is not active")
	}

	return &result, nil
}

func (c *Client) auth(ctx context.Context) *resty.Request {
	return c.cli.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetError(&KeycloakError{})
}

type KeycloakError struct {
	Err              string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *KeycloakError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.ErrorDescription)
}
