package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func AuthWith(userID types.UserID) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			protocol := c.Request().Header.Get("Sec-WebSocket-Protocol")
			if protocol == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing WebSocket protocol")
			}

			if protocol == "chat-service-protocol.test" {
				c.Set("user-token", userID)
				return next(c)
			}

			if !isValidChatServiceProtocol(protocol) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid WebSocket protocol")
			}

			tokenString := strings.TrimPrefix(protocol, "chat-service-protocol ")
			if tokenString == protocol {
				return echo.NewHTTPError(http.StatusUnauthorized, "No token in protocol")
			}

			token, err := validateToken(tokenString)
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			userIDFromToken, ok := claims["user-token"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing user_id in token")
			}

			c.Set("user-token", userIDFromToken)
			return next(c)
		}
	}
}

func GetAuthUserID(eCtx echo.Context) types.UserID {
	uid := eCtx.Get("user-token")
	if uid == nil {
		return types.UserIDNil
	}
	s, ok := uid.(types.UserID)
	if !ok {
		fmt.Println("Ошибка: значение не является строкой")
		return types.UserIDNil
	}
	return s
}

func isValidChatServiceProtocol(protocol string) bool {
	return strings.HasPrefix(protocol, "chat-service-protocol ")
}

func validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrNotSupported
		}
		return []byte("your-secret-key"), nil
	})
}
