-- name: CreateUser :one
-- CreateUser создаёт нового пользователя и возвращает его данные без хеша пароля.
INSERT INTO users (login, password_hash)
VALUES ($1, $2)
RETURNING id, login, created_at;

-- name: GetUserByLogin :one
-- GetUserByLogin возвращает пользователя по логину вместе с password_hash для сверки при входе.
SELECT id, login, password_hash, created_at
FROM users
WHERE login = $1;