package app

import (
	grpcapp "exT/internal/app/grpcapp"
	"exT/internal/services/auth"
	"exT/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {

	// Инициализация хранилища
	storage, err := sqlite.New(storagePath)

	if err != nil {
		panic(err)
	}

	// инициализация auth service (сервисный слой)
	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
