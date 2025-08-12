-- internal/db/migrations/002_additional_tables.up.sql - исправленная версия

-- Add additional columns to users table
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'is_admin') THEN
        ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT false;
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'users' AND column_name = 'last_seen') THEN
        ALTER TABLE users ADD COLUMN last_seen TIMESTAMP;
    END IF;
END $$;

-- Add updated_at column to messages table
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'messages' AND column_name = 'updated_at') THEN
        ALTER TABLE messages ADD COLUMN updated_at TIMESTAMP;
    END IF;
END $$;

-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create password_reset_tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create room_mutes table
CREATE TABLE IF NOT EXISTS room_mutes (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    muted_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    UNIQUE(room_id, user_id)
);

-- Create audit_logs table for security logging
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id INTEGER,
    ip_address INET,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for new tables (only if they don't exist)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_refresh_tokens_user_id') THEN
        CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_refresh_tokens_token') THEN
        CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_refresh_tokens_expires_at') THEN
        CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_password_reset_tokens_user_id') THEN
        CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_password_reset_tokens_token') THEN
        CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_password_reset_tokens_expires_at') THEN
        CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_room_mutes_room_id') THEN
        CREATE INDEX idx_room_mutes_room_id ON room_mutes(room_id);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_room_mutes_user_id') THEN
        CREATE INDEX idx_room_mutes_user_id ON room_mutes(user_id);
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_room_mutes_expires_at') THEN
        CREATE INDEX idx_room_mutes_expires_at ON room_mutes(expires_at);
    END IF;
END $$;

DO $ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_audit_logs_user_id') THEN
        CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
    END IF;
END $;

DO $ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_audit_logs_action') THEN
        CREATE INDEX idx_audit_logs_action ON audit_logs(action);
    END IF;
END $;

DO $ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_audit_logs_created_at') THEN
        CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
    END IF;
END $;

-- Add trigger for updating updated_at column
DO $ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_refresh_tokens_updated_at') THEN
        CREATE TRIGGER trigger_refresh_tokens_updated_at
            BEFORE UPDATE ON refresh_tokens
            FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
    END IF;
END $;

-- Clean up expired tokens function
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $
BEGIN
    DELETE FROM refresh_tokens WHERE expires_at < NOW();
    DELETE FROM password_reset_tokens WHERE expires_at < NOW();
    DELETE FROM room_mutes WHERE expires_at IS NOT NULL AND expires_at < NOW();
END;
$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE refresh_tokens IS 'Stores JWT refresh tokens for users';
COMMENT ON TABLE password_reset_tokens IS 'Stores password reset tokens';
COMMENT ON TABLE room_mutes IS 'Stores information about muted users in rooms';
COMMENT ON TABLE audit_logs IS 'Stores audit trail for security and compliance';

COMMENT ON COLUMN room_mutes.expires_at IS 'NULL means permanent mute until manually removed';
COMMENT ON COLUMN audit_logs.details IS 'Additional metadata about the action in JSON format';