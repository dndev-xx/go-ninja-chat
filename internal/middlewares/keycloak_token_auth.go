package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4/middleware"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/introspector_mock.gen.go -package=middlewaresmocks Introspector

const tokenCtxKey = "user-token"
var ErrNoRequiredResourceRole = errors.New("no required resource role")

type Introspector interface {
	IntrospectToken(ctx context.Context, token string) (*keycloakclient.IntrospectTokenResult, error)
}

// NewKeycloakTokenAuth returns a middleware that implements "active" authentication:
// each request is verified by the Keycloak server.
func NewKeycloakTokenAuth(introspector Introspector, resource, role string) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:Authorization",
		AuthScheme: "Bearer",
		Validator: func(tokenStr string, eCtx echo.Context) (bool, error) {
			// Introspect token first
			_, err := introspector.IntrospectToken(eCtx.Request().Context(), tokenStr)
			if err != nil {
				return false, err
			}

			// Parse token without signature verification
			parser := jwt.Parser{SkipClaimsValidation: true}
			token, _, err := parser.ParseUnverified(tokenStr, &claims{})
			if err != nil {
				return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			// Validate claims
			cl, ok := token.Claims.(*claims)
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
		},
	})
}

func hasResourceRole(cl *claims, resource, role string) bool {
	if resourceAccess, ok := cl.ResourceAccess[resource]; ok {
		for _, r := range resourceAccess.Roles {
			if r == role {
				return true
			}
		}
	}
	return false
}

func MustUserID(eCtx echo.Context) *uuid.UUID {
	uid, ok := userID(eCtx)
	if !ok {
		panic("no user token in request context")
	}
	return uid
}

func userID(eCtx echo.Context) (*uuid.UUID, bool) {
	t := eCtx.Get(tokenCtxKey)
	if t == nil {
		return nil, false
	}

	tt, ok := t.(*jwt.Token)
	if !ok {
		return nil, false
	}

	if claims, ok := tt.Claims.(*claims); ok {
		return claims.UserID(), true
	}
	return nil, false
}