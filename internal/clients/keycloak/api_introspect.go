package keycloakclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AudArray []string

func (a *AudArray) UnmarshalJSON(data []byte) error {
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
	Exp    int      `json:"exp"`
	Iat    int      `json:"iat"`
	Aud    AudArray `json:"aud"`
	Active bool     `json:"active"`
}

// IntrospectToken implements
// https://www.keycloak.org/docs/latest/authorization_services/index.html#obtaining-information-about-an-rpt
func (c *Client) IntrospectToken(ctx context.Context, token string) (*IntrospectTokenResult, error) {
	url := fmt.Sprintf("%s/protocol/openid-connect/token/introspect", c.BasePath)

	resp, err := c.auth(ctx).
		SetFormData(map[string]string{
			"token": token,
		}).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("introspection failed: %s", resp.String())
	}

	var result struct {
		Exp    int             `json:"exp"`
		Iat    int             `json:"iat"`
		Aud    json.RawMessage `json:"aud"`
		Active bool            `json:"active"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal introspection response: %v", err)
	}

	var audArray []string
	if result.Aud[0] == '"' {
		var singleAud string
		if err := json.Unmarshal(result.Aud, &singleAud); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aud field: %v", err)
		}
		audArray = append(audArray, singleAud)
	} else {
		if err := json.Unmarshal(result.Aud, &audArray); err != nil {
			return nil, fmt.Errorf("failed to unmarshal aud field: %v", err)
		}
	}

	finalResult := IntrospectTokenResult{
		Exp:    result.Exp,
		Iat:    result.Iat,
		Aud:    audArray,
		Active: result.Active,
	}

	return &finalResult, nil
}

func (c *Client) auth(ctx context.Context) *resty.Request {
	return c.cli.R().SetContext(ctx).SetHeader("Content-Type", "application/x-www-form-urlencoded")
}
