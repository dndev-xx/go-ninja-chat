package middlewares

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/introspector_mock.gen.go -package=middlewaresmocks Introspector

const tokenCtxKey = "user-token"

var (
	ErrNoRequiredResourceRole = errors.New("no required resource role")
	ErrWSProtocolError        = errors.New("ws protocol error")
)

type Introspector interface {
	IntrospectToken(ctx context.Context, token string) (*keycloakclient.IntrospectTokenResult, error)
}

// NewKeycloakTokenAuth returns a middleware that implements "active" authentication:
// each request is verified by the Keycloak server.
func NewKeycloakTokenAuth(introspector Introspector, resource, role string) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			// Пропускаем аутентификацию для WebSocket соединений
			return strings.Contains(c.Request().URL.Path, "/ws") &&
				c.Request().Header.Get("Upgrade") == "websocket"
		},
		KeyLookup:  "header:Authorization",
		AuthScheme: "Bearer",
		Validator: func(tokenStr string, eCtx echo.Context) (bool, error) {
			return validateTokenn(introspector, resource, role, tokenStr, eCtx)
		},
	})
}

// NewWebSocketAuth возвращает middleware для аутентификации WebSocket соединений
func NewWebSocketAuth(introspector Introspector, resource, role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Проверяем, что это WebSocket запрос
			if !strings.Contains(c.Request().URL.Path, "/ws") ||
				c.Request().Header.Get("Upgrade") != "websocket" {
				return next(c)
			}

			// Извлекаем токен из Sec-WebSocket-Protocol
			protocol := c.Request().Header.Get("Sec-WebSocket-Protocol")
			if protocol == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrWSProtocolError.Error())
			}

			// Парсим протокол: ожидаем формат "chat-service-protocol, <token>"
			parts := strings.Split(protocol, ",")
			if len(parts) < 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrWSProtocolError.Error())
			}

			// Проверяем наличие chat-service-protocol
			if !strings.Contains(strings.TrimSpace(parts[0]), "chat-service-protocol") {
				return echo.NewHTTPError(http.StatusUnauthorized, ErrWSProtocolError.Error())
			}

			// Извлекаем токен (второй элемент после запятой)
			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "empty token in protocol")
			}

			valid, err := validateTokenn(introspector, resource, role, tokenStr, c)
			if err != nil || !valid {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			return next(c)
		}
	}
}

func validateTokenn(introspector Introspector, resource, role, tokenStr string, eCtx echo.Context) (bool, error) {
	// Introspect token first
	_, err := introspector.IntrospectToken(eCtx.Request().Context(), tokenStr)
	if err != nil {
		return false, err
	}

	// Parse token without signature verification
	parser := jwt.Parser{SkipClaimsValidation: true}
	token, _, err := parser.ParseUnverified(tokenStr, &Claims{})
	if err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Validate claims
	cl, ok := token.Claims.(*Claims)
	if !ok {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	// Manually validate standard claims to preserve error types
	stdClaims := &jwt.StandardClaims{
		ExpiresAt: cl.ExpiresAt,
		IssuedAt:  cl.IssuedAt,
		NotBefore: cl.NotBefore,
	}
	if err := stdClaims.Valid(); err != nil {
		return false, err
	}
	// Validate custom claims
	if err := cl.Valid(); err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Check required resource role
	if !hasResourceRole(cl, resource, role) {
		return false, ErrNoRequiredResourceRole
	}

	// Store token in context
	eCtx.Set(tokenCtxKey, token)
	return true, nil
}

func hasResourceRole(cl *Claims, resource, role string) bool {
	if resourceAccess, ok := cl.ResourceAccess[resource]; ok {
		return slices.Contains(resourceAccess.Roles, role)
	}
	return false
}

func MustUserID(eCtx echo.Context) types.UserID {
	uid, ok := userID(eCtx)
	if !ok {
		panic("no user token in request context")
	}
	return uid
}

func userID(eCtx echo.Context) (types.UserID, bool) {
	t := eCtx.Get(tokenCtxKey)
	if t == nil {
		return types.UserIDNil, false
	}

	tt, ok := t.(*jwt.Token)
	if !ok {
		return types.UserIDNil, false
	}

	if claims, ok := tt.Claims.(*Claims); ok {
		return types.UserID(*claims.UserID()), true
	}
	return types.UserIDNil, false
}
