package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4/middleware"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
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
		},
	})
}

func hasResourceRole(cl *Claims, resource, role string) bool {
	if resourceAccess, ok := cl.ResourceAccess[resource]; ok {
		for _, r := range resourceAccess.Roles {
			if r == role {
				return true
			}
		}
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
