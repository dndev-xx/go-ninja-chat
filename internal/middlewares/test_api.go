package middlewares

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	"github.com/labstack/echo/v4"
)

func SetToken(c echo.Context, uid types.UserID) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	secret := []byte("your_secret_key")
	signedToken, err := token.SignedString(secret)
	if err != nil {
		c.Logger().Error("Ошибка при подписывании токена:", err)
		return
	}
	c.Set(string(tokenCtxKey), signedToken)
}
