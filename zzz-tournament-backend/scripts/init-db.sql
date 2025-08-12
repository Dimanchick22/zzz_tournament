-- Скрипт инициализации базы данных для ZZZ Tournament

-- Создание базы данных (если не существует)
SELECT 'CREATE DATABASE zzz_tournament'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'zzz_tournament')\gexec

-- Подключение к базе данных
\c zzz_tournament;

-- Включение расширений
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- для полнотекстового поиска
CREATE EXTENSION IF NOT EXISTS "unaccent"; -- для поиска без учета диакритики

-- Создание пользователя для приложения (если не существует)
DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'tournament_app') THEN
      CREATE USER tournament_app WITH PASSWORD 'app_password';
   END IF;
END
$$;

-- Предоставление прав пользователю
GRANT CONNECT ON DATABASE zzz_tournament TO tournament_app;
GRANT USAGE ON SCHEMA public TO tournament_app;
GRANT CREATE ON SCHEMA public TO tournament_app;

-- Создание схемы для логов (опционально)
CREATE SCHEMA IF NOT EXISTS logs;
GRANT USAGE ON SCHEMA logs TO tournament_app;
GRANT CREATE ON SCHEMA logs TO tournament_app;

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Комментарий
COMMENT ON FUNCTION trigger_set_timestamp() IS 'Автоматически обновляет поле updated_at при изменении записи';

-- Создание функции для генерации username slug
CREATE OR REPLACE FUNCTION generate_username_slug(input_text TEXT)
RETURNS TEXT AS $$
BEGIN
  RETURN LOWER(
    REGEXP_REPLACE(
      REGEXP_REPLACE(
        UNACCENT(input_text),
        '[^a-zA-Z0-9\s\-_]', '', 'g'
      ),
      '\s+', '-', 'g'
    )
  );
END;
$$ LANGUAGE plpgsql;

-- Создание индекса для полнотекстового поиска пользователей
-- (будет создан после создания таблицы users в миграциях)

-- Настройки для лучшей производительности
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_duration = on;
ALTER SYSTEM SET log_min_duration_statement = 1000; -- логировать запросы дольше 1 сек

-- Создание роли только для чтения (для аналитики)
DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'readonly_user') THEN
      CREATE ROLE readonly_user;
   END IF;
END
$$;

-- Предоставление прав только для чтения
GRANT CONNECT ON DATABASE zzz_tournament TO readonly_user;
GRANT USAGE ON SCHEMA public TO readonly_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly_user;

-- Вывод информации
SELECT 
    'Database zzz_tournament initialized successfully' as status,
    version() as postgres_version,
    current_timestamp as initialized_at;