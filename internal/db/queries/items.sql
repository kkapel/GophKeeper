-- name: CreateItem :one
-- CreateItem создаёт новую запись приватных данных и возвращает её.
INSERT INTO items (user_id, type, encrypted_payload, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetItem :one
-- GetItem возвращает запись пользователя по id, исключая мягко удалённые.
SELECT * FROM items
WHERE id = $1 AND user_id = $2 AND NOT deleted;

-- name: ListItems :many
-- ListItems возвращает все неудалённые записи пользователя.
SELECT * FROM items
WHERE user_id = $1 AND NOT deleted
ORDER BY updated_at DESC;

-- name: UpdateItem :one
-- UpdateItem обновляет запись с проверкой версии.
-- Если версия не совпала, ни одна строка не будет затронута.
UPDATE items
SET encrypted_payload = $3,
    metadata = $4,
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $5 AND NOT deleted
RETURNING *;

-- name: DeleteItem :execrows
-- DeleteItem выполняет мягкое удаление записи.
UPDATE items
SET deleted = TRUE, version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND NOT deleted;