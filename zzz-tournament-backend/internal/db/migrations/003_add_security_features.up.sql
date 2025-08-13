-- migrations/003_add_security_features.up.sql

-- Добавляем новые поля в таблицу users
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true NOT NULL,
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT false NOT NULL,
ADD COLUMN IF NOT EXISTS last_login TIMESTAMP,
ADD COLUMN IF NOT EXISTS login_attempts INTEGER DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;

-- Обновляем существующих пользователей
UPDATE users SET is_active = true WHERE is_active IS NULL;
UPDATE users SET is_verified = false WHERE is_verified IS NULL;
UPDATE users SET login_attempts = 0 WHERE login_attempts IS NULL;

-- Изменяем таблицу refresh_tokens для хранения хешированных токенов
ALTER TABLE refresh_tokens 
DROP COLUMN IF EXISTS token,
ADD COLUMN IF NOT EXISTS token_hash VARCHAR(64) NOT NULL;

-- Создаем уникальный индекс на token_hash
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Добавляем индексы для производительности (предупреждения безопасны)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- Создаем таблицу для событий безопасности
CREATE TABLE IF NOT EXISTS security_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    event VARCHAR(100) NOT NULL,
    client_ip INET NOT NULL,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Индексы для security_events
CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events(user_id);
CREATE INDEX IF NOT EXISTS idx_security_events_event ON security_events(event);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);
CREATE INDEX IF NOT EXISTS idx_security_events_client_ip ON security_events(client_ip);

-- Добавляем constraints с проверкой существования
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_users_email_format'
    ) THEN
        ALTER TABLE users 
        ADD CONSTRAINT chk_users_email_format 
        CHECK (email ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_users_username_length'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT chk_users_username_length 
        CHECK (length(username) >= 3 AND length(username) <= 50);
    END IF;
END $$;

-- Удаляем старые функции, если они существуют
DROP FUNCTION IF EXISTS cleanup_expired_tokens();
DROP FUNCTION IF EXISTS unlock_expired_accounts();

-- Функция для очистки токенов
CREATE FUNCTION cleanup_expired_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_refresh_tokens INTEGER;
    deleted_reset_tokens INTEGER;
BEGIN
    DELETE FROM refresh_tokens WHERE expires_at <= CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_refresh_tokens = ROW_COUNT;

    DELETE FROM password_reset_tokens WHERE expires_at <= CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_reset_tokens = ROW_COUNT;

    RETURN deleted_refresh_tokens + deleted_reset_tokens;
END;
$$ LANGUAGE plpgsql;

-- Функция для разблокировки аккаунтов
CREATE FUNCTION unlock_expired_accounts()
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE users
    SET locked_until = NULL, login_attempts = 0
    WHERE locked_until <= CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;


-- Комментарии к таблицам и столбцам
COMMENT ON COLUMN users.is_active IS 'Активен ли аккаунт пользователя';
COMMENT ON COLUMN users.is_verified IS 'Подтвержден ли email пользователя';
COMMENT ON COLUMN users.last_login IS 'Время последнего входа';
COMMENT ON COLUMN users.login_attempts IS 'Количество неудачных попыток входа';
COMMENT ON COLUMN users.locked_until IS 'Время до которого аккаунт заблокирован';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 хеш refresh токена';
COMMENT ON TABLE security_events IS 'Журнал событий безопасности';
