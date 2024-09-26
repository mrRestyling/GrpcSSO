package jwt

import (
	"exT/internal/domain/models"
	// "exT/internal/lib/jwt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO Покрыть тестами данный модуль

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	// подписываем свой токен
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil

}
