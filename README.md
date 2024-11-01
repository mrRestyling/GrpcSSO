SSO
1. Аутентификация и авторизация (Auth)
2. Работа с правами (Permissions)
3. Сервис, который предоставляет информацию о пользователях (User Info)

User делает запрос на какой-то сервер URL Shortener
User -> Auth сервер по логину, возвращаем токен


Слои:
Транспортный - взаимодействие (gRPC Server)
Сервисный - Бизнес-логика (Auth, TODO Permissions, TODO EventSender(отправлять например в кафку))
Работа с данными - SQLite
APP - Запуск, остановка, healthchecks
Tests





Дополнительные пакеты:
- github.com/mrRestyling/protos (прото модуль с гитхаба)
- go get golang.org/x/crypto (токены)
- go get github.com/golang-migrate/migrate/v4 (для миграции бд)
- go get github.com/mattn/go-sqlite3 (драйвер sqlite3)
- go get github.com/brianvoe/gofakeit/v6 (генерация данных для тестов)
- go get github.com/stretchr/testify (пакет testify)

Генерация прото-файла:
protoc -I protos proto/sso/sso.proto --go_out=./protos --go_opt=paths=source_relative --go-grpc_out=./protos --go-grpc_opt=paths=source_relative

Команды:
- go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations
- go run ./cmd/migrator/main.go --storage-path=./storage/sso.db --migrations-path=./tests/migrations --migrations-table=migrations_test (выпоняем тестовые миграции)


Запуск сервера:
 go run cmd/sso/main.go --config=./config/local.yaml
