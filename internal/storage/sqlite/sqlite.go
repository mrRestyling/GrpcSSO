package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"exT/internal/domain/models"
	"exT/internal/storage"
	"fmt"
	"log"

	"github.com/mattn/go-sqlite3"
	// _ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Конструктор Storage
func New(storagePath string) (*Storage, error) {

	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		log.Println("no connect to DB")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil

}

// Запрос на добавление пользователя
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	// Подготовка запроса
	// (1. защищаем от SQL-инъекций)
	// (2. подготовленный запрос можно выполнить несколько раз с разными параметрами)
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES (?,?)")
	if err != nil {
		log.Println("err create req")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Выполнение подготовленного запроса с заданными параметрами
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {

		// Обработка ошибки, которая в sqlite
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем ID созданной записи

	id, err := res.LastInsertId()
	if err != nil {
		log.Println("err get id")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// Запрос получения пользователя
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {

	const op = "storage.sqlite.User"

	// Подготовка запроса
	smtp, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		log.Println("err create req")
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// Получение ответа

	row := smtp.QueryRowContext(ctx, email)

	var user models.User

	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdminS(ctx context.Context, userID int64) (bool, error) {

	const op = "storage.sqlite.UserAdmin"

	smtp, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")

	if err != nil {
		log.Println("err create req")
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := smtp.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.sqlite.App"

	smtp, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := smtp.QueryRowContext(ctx, appID)

	var app models.App

	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)

	}
	return app, nil
}
