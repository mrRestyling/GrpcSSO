package tests

import (
	"exT/tests/suite"
	"testing"
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

}
