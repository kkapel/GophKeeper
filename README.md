# GophKeeper

Клиент-серверный менеджер паролей: хранит логины/пароли, произвольные
текстовые и бинарные данные, а также данные банковских карт. Сервер — gRPC,
хранилище — PostgreSQL.

> ⚠️ Проект в разработке: собран скелет сервера (конфигурация, логирование,
> подключение к БД с миграциями, каркас gRPC с graceful shutdown).

## Требования

- Go 1.25+
- Docker (для PostgreSQL)

## Запуск

**1. Поднять PostgreSQL** (задайте свой пароль вместо `<your-password>`):

```bash
docker run --name gophkeeper-pg -e POSTGRES_PASSWORD=<your-password> -e POSTGRES_DB=gophkeeper -p 5432:5432 -d postgres:17
```

Миграции накатываются автоматически при старте сервера.

**2. Задать DSN** (переменная читается с префиксом `GOPHKEEPER_`; подставьте тот
же пароль, что задали контейнеру):

```bash
# Linux / macOS
export GOPHKEEPER_DATABASE_URL="postgres://postgres:<your-password>@localhost:5432/gophkeeper?sslmode=disable"
```

```powershell
# Windows (PowerShell)
$env:GOPHKEEPER_DATABASE_URL = "postgres://postgres:<your-password>@localhost:5432/gophkeeper?sslmode=disable"
```

**3. Запустить сервер:**

```bash
go run ./cmd/server
```

Остановка — `Ctrl+C` (корректное завершение). По умолчанию сервер слушает `:8080`.

## Переменные окружения

| Переменная                | Обязательность | По умолчанию | Описание                     |
|---------------------------|-------------   |--------------|------------------------------|
| `GOPHKEEPER_DATABASE_URL` | да             | —            | DSN подключения к PostgreSQL |
| `GOPHKEEPER_GRPC_ADDRESS` | нет            | `:8080`      | адрес gRPC-сервера           |
| `GOPHKEEPER_LOGGER_LEVEL` | нет            | `info`       | уровень логирования          |
| `GOPHKEEPER_JWT_SECRET`   | да             | —            | JWT-секрет                   |