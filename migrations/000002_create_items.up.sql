CREATE TABLE items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type              SMALLINT NOT NULL,
    encrypted_payload BYTEA NOT NULL,
    metadata          TEXT NOT NULL DEFAULT '',
    version           BIGINT NOT NULL DEFAULT 1,
    deleted           BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_items_user_id ON items (user_id);

COMMENT ON TABLE  items IS 'Приватные данные пользователей: зашифрованный payload + открытая метаинформация';
COMMENT ON COLUMN items.user_id           IS 'Владелец записи, FK на users.id';
COMMENT ON COLUMN items.type              IS 'Тип данных: 1=credentials, 2=text, 3=binary, 4=card';
COMMENT ON COLUMN items.encrypted_payload IS 'Зашифрованные данные пользователя';
COMMENT ON COLUMN items.metadata          IS 'Открытая текстовая метаинформация: сайт, банк, личность и т.п.';
COMMENT ON COLUMN items.version           IS 'Версия записи';
COMMENT ON COLUMN items.deleted           IS 'Удалена ли запись';