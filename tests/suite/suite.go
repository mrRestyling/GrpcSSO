package suite

import (
	"context"
	"exT/internal/config"
	"net"
	"strconv"
	"testing"

	ssov1 "github.com/mrRestyling/protos/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T                  // Потребуется для вызова *testing.T внутри Suite
	Cfg        *config.Config   // Конфигурация приложения
	AuthClient ssov1.AuthClient // Клиент для взаимодействия с gRpc-сервером
}

func New(t *testing.T) (context.Context, *Suite) {

	t.Helper()   // при фейле правильно формирует стек вызовов, чтобы эта ф-я не была указана как финальная
	t.Parallel() // тесты выполняются параллельно

	// // Получаем env
	// // (для запуска автоматических тестов на GitHub)
	// key := "CONFIG_PATH"
	// if v := os.Getenv(key); v != "" {
	// 	return v
	// }

	cfg := config.MustLoadByPath("../config/local.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	// Чистка после теста
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	// Создаем grpc-клиент для нашего сервиса

	cc, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		// Используем insecure-коннект (чтобы не заморачиваться с соединениями)
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc), // создаем новый auth клиент (кодогенерация grpc)
	}

}

func grpcAddress(cfg *config.Config) string {
	// возвращае ф-ю, которая объединяет хост и порт
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
