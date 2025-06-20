package middlewares

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func SetToken(c echo.Context, uid types.UserID) error {
	now := time.Now()
	expiresAt := now.Add(time.Hour * 72).Unix()
	issuedAt := now.Unix()

	claims := jwt.MapClaims{
		"exp":             expiresAt,
		"iat":             issuedAt,
		"auth_time":       issuedAt,
		"jti":             uuid.New().String(),
		"iss":             "http://localhost:3010/realms/Bank",
		"aud":             "account",
		"sub":             uid.String(),
		"typ":             "Bearer",
		"azp":             "chat-ui-client",
		"nonce":           uuid.New().String(),
		"session_state":   uuid.New().String(),
		"acr":             "1",
		"allowed-origins": []string{"*"},
		"realm_access": map[string][]string{
			"roles": {"offline_access", "default-roles-bank", "uma_authorization"},
		},
		"resource_access": map[string]any{
			"chat-ui-client": map[string][]string{
				"roles": {"support-chat-client"},
			},
			"account": map[string][]string{
				"roles": {"manage-account", "manage-account-links", "view-profile"},
			},
		},
		"scope":              "openid profile email",
		"sid":                uuid.New().String(),
		"email_verified":     false,
		"preferred_username": "test",
		"given_name":         "",
		"family_name":        "",
		"email":              "test@test.ru",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte("test_secret"))
	if err != nil {
		c.Logger().Error("Ошибка при подписывании токена:", err)
		return err
	}

	c.Set(string(tokenCtxKey), signedToken)
	return nil
}
