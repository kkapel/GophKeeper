# GophKeeper

Клиент-серверный менеджер паролей: хранит логины/пароли, произвольные
текстовые и бинарные данные, а также данные банковских карт. Сервер — gRPC,
хранилище — PostgreSQL.

> ⚠️ Проект в разработке. Готово: регистрация и вход пользователей
> (сервер + CLI-клиент), хранение данных в PostgreSQL, JWT-аутентификация.
> В работе: хранение приватных данных (CRUD), шифрование, TLS.

## Требования

- Go 1.25+
- Docker (для PostgreSQL)

## Запуск сервера

**1. Поднять PostgreSQL** (задайте свой пароль вместо `<your-password>`):

```bash
docker run --name gophkeeper-pg -e POSTGRES_PASSWORD=<your-password> -e POSTGRES_DB=gophkeeper -p 5432:5432 -d postgres:17
```

Миграции накатываются автоматически при старте сервера.

**2. Задать переменные окружения** (читаются с префиксом `GOPHKEEPER_`):

```bash
# Linux / macOS
export GOPHKEEPER_DATABASE_URL="postgres://postgres:<your-password>@localhost:5432/gophkeeper?sslmode=disable"
export GOPHKEEPER_JWT_SECRET="любая-длинная-случайная-строка"
```

```powershell
# Windows (PowerShell)
$env:GOPHKEEPER_DATABASE_URL = "postgres://postgres:<your-password>@localhost:5432/gophkeeper?sslmode=disable"
$env:GOPHKEEPER_JWT_SECRET = "любая-длинная-случайная-строка"
```

> Пароль в DSN и порт должны совпадать с параметрами контейнера PostgreSQL.
> `GOPHKEEPER_JWT_SECRET` — секрет для подписи JWT-токенов; задаётся только через
> окружение, в коде не хранится.

**3. Запустить сервер:**

```bash
go run ./cmd/server
```

Остановка — `Ctrl+C` (корректное завершение). По умолчанию сервер слушает `:8080`.

## Переменные окружения

| Переменная                | Обязательна | По умолчанию | Описание                     |
|---------------------------|-------------|--------------|------------------------------|
| `GOPHKEEPER_DATABASE_URL` | да          | —            | DSN подключения к PostgreSQL |
| `GOPHKEEPER_JWT_SECRET`   | да          | —            | секрет для подписи JWT       |
| `GOPHKEEPER_GRPC_ADDRESS` | нет         | `:8080`      | адрес gRPC-сервера           |
| `GOPHKEEPER_LOGGER_LEVEL` | нет         | `info`       | уровень логирования          |

## Клиент

CLI-клиент подключается к серверу по gRPC.

### Команды

- `register --login <логин>` — регистрация нового пользователя
- `login --login <логин>` — вход
- `version` — версия и дата сборки

Пароль запрашивается интерактивно (ввод скрыт), не передаётся флагом.

### Флаги

- `--address` — адрес gRPC-сервера, по умолчанию `127.0.0.1:8080`
- `--login` — логин пользователя

### Пример

```bash
go run ./cmd/client register --login ivan
```

После успешного входа токен сохраняется в `~/.gophkeeper/token`.
