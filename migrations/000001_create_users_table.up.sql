CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);


COMMENT ON TABLE  users IS 'Зарегистрированные пользователи системы';
COMMENT ON COLUMN users.id            IS 'Идентификатор пользователя, генерируется БД';
COMMENT ON COLUMN users.login         IS 'Уникальный логин для входа';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля пользователя';
COMMENT ON COLUMN users.created_at    IS 'Дата и время регистрации';