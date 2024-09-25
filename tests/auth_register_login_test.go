package tests

import (
	"exT/tests/suite"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/mrRestyling/protos/proto/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	// обрващаемся (создаем) suite
	ctx, st := suite.New(t)

	// генерация случайного логина и пароля
	// go get github.com/brianvoe/gofakeit/v6

	email := gofakeit.Email()

	pass := randomFakePassword()

	// используем клиент, который есть в нашем suite

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err) // аналог, который не останавливает тест -> assert.Equal(t,err)
	assert.NotEmpty(t, respReg.GetUserId())

	//Вызываем функцию логин
	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	// Первое - получаем токен
	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	// Второе - пытаемся парсить этот токен
	// appSecret  = "test-secret"
	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	// Проверяем, что он проходит валидацию
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	// Проверяем, что в токене содержится корректная информация
	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64))) // особенности хранения данных в MapClaims
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	// Проверяем, что время истечения токена совпадает с ожидаемым

}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
