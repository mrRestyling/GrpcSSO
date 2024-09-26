package auth

import (
	"context"
	"errors"
	"exT/internal/services/auth"

	ssov1 "github.com/mrRestyling/protos/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userId int64, err error)
	IsAdmin(ctx context.Context, userId int64) (bool, error)
}

// обработчик всех входящих запросов
type ServerAPI struct {
	ssov1.UnimplementedAuthServer //заглушка, чтобы не реализовывать все методы
	auth                          Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})

}

func (s *ServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {

	// проверяем входные данные отдельной функцией
	if err := validationLogin(req); err != nil {
		return nil, err
	}

	// TODO Сервисный слой (auth) к интерфейсу auth

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		// обработка ошибки
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *ServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {

	if err := validationRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// обрабатываем случаи ошибки

		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "user already exists")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *ServerAPI) IsAdminS(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {

	if err := validationAdmin(req); err != nil {
		return nil, err
	}

	admin, err := s.auth.IsAdmin(ctx, req.UserId)
	if err != nil {

		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: admin,
	}, nil

}

// validation - дополнительная функция к методу Login. Проверка полей валидации
func validationLogin(req *ssov1.LoginRequest) error {
	// проверяем, что Email не пустой
	// возвращаем ошибку из GRPC(пакет статус)
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}

	// проверяем, что password не пустой
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}
	return nil
}

func validationRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	return nil
}

func validationAdmin(req *ssov1.IsAdminRequest) error {

	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty user_id")
	}

	return nil
}
