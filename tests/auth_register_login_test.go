package tests

import (
	"exT/tests/suite"
	"testing"
	"time"

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
	// обращаемся (создаем) suite
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

	loginTime := time.Unix(int64(claims["exp"].(float64)), 0)
	const deltaSeconds = 3600 // точность

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestDoubleReg(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	resp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetUserId)

	resp, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, resp.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

// Табличный тест
func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "empty password",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "empty email",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "empty email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "empty password",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "empty email",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "empty email",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login without AppID",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
