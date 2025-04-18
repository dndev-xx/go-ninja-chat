package middlewares

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var (
	ErrNoAllowedResources = errors.New("no allowed resources")
	ErrSubjectNotDefined  = errors.New(`"sub" is not defined`)
)

type claims struct {
	*jwt.StandardClaims
	RealmAccess      struct {
        Roles []string `json:"roles"`
    } `json:"realm_access"`
    ResourceAccess map[string]struct {
        Roles []string `json:"roles"`
    } `json:"resource_access"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	Sid               string   `json:"sid"`
	Audience          audience `json:"aud"`
}

// audience can be string or []string
type audience []string

func (a *audience) UnmarshalJSON(data []byte) error {
	var singleAud string
	if err := json.Unmarshal(data, &singleAud); err == nil {
		*a = []string{singleAud}
		return nil
	}
	var multiAud []string
	if err := json.Unmarshal(data, &multiAud); err != nil {
		return err
	}
	*a = multiAud
	return nil
}

// Valid validates the claims
func (c *claims) Valid() error {
	tempClaims := &jwt.StandardClaims{
		ExpiresAt: c.ExpiresAt,
		Id:        c.Id,
		IssuedAt:  c.IssuedAt,
		Issuer:    c.Issuer,
		NotBefore: c.NotBefore,
		Subject:   c.Subject,
	}

	if err := tempClaims.Valid(); err != nil {
		return err
	}

	if c.Subject == "" {
		return ErrSubjectNotDefined
	}

	if len(c.ResourceAccess) == 0 {
		return ErrNoAllowedResources
	}

	return nil
}

func (c *claims) UserID() *uuid.UUID {
	if c.Subject == "" {
		return nil
	}
	uid, err := uuid.Parse(c.Subject)
	if err != nil {
		return nil
	}
	return &uid
}