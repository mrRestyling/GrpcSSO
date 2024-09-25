package auth

import (
	"context"
	"errors"
	"exT/internal/domain/models"
	"exT/internal/lib/jwt"
	"exT/internal/storage"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdminS(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New returns a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    24 * time.Hour,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appId int) (string, error) {

	// const op - метка для дебага
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {

		if errors.Is(err, storage.ErrUserNotFound) { // !!! нужно изучить !!!
			a.log.Warn("user not found")
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials")
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	// создаем токен + отдельный пакет (package jwt)

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {

		a.log.Error("failed to create token")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {

	// const op - метка для дебага
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering new user")

	// хэшируем и солим пароль

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed generate password hash")

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// сохраняем в базу данных
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("failed to save user")
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("user registered", slog.Int64("user_id", id))

	return id, nil
}

// IsAdmin проверяет, является ли пользователь админом
func (a *Auth) IsAdmin(ctx context.Context, userId int64) (bool, error) {

	const op = "auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.Int64("user_id", userId))

	log.Info("checking if user is admin")

	IsAdmin, err := a.usrProvider.IsAdminS(ctx, userId)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {

			log.Warn("user not found")

			return false, fmt.Errorf("%s: %w", op, err)

		}
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", IsAdmin))

	return IsAdmin, nil
}
