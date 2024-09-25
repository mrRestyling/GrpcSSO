package main

import (
	"exT/internal/app"
	"exT/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	// Запуск go run cmd/sso/main.go --config=./config/local.yaml

	// Инициализация объекта конфига...
	cfg := config.MustLoad()

	// Инициализация логгера
	log := setupLogger(cfg.Env)

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Any("cfg", cfg),
		slog.Int("port", cfg.GRPC.Port),
	)

	log.Debug("debug message")

	log.Error("error message")

	log.Warn("warn message")

	// Инициализация приложения
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// Запуск gRPC-сервера в отдельной горутине
	go application.GRPCSrv.MustRun()

	// GF<-
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	signNew := <-stop

	log.Info("graceful shutdown", slog.String("signal", signNew.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")

	// <-GF

}

// инициализация логгера
func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {

	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}
