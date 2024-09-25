package grpcapp

import (
	authgRPC "exT/internal/grpc/auth"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

// New create new gRPC server app
func New(log *slog.Logger, authService authgRPC.Auth, port int) *App {

	// создаем сервер
	gRPCServer := grpc.NewServer()

	// подключаем обработчик
	authgRPC.Register(gRPCServer, authService)

	// возвращаем тип App наружу
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {

	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {

	// метка для дебага
	const op = "grpcapp.App.Run"

	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	// создаем лисенер, который будет слушать tcp сообщения

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gprc server started", slog.String("addr", l.Addr().String()))

	// запускаем сервер
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// метод для graceful shutdown
func (a *App) Stop() {
	const op = "grpcapp.App.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
