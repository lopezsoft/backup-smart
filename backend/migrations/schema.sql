-- Schema SQL de backup-smart para referencia.
-- La creación real de tablas se hace mediante el método Migrate() del repositorio Go.
-- Este archivo es solo documentación del esquema de la base de datos SQLite.

-- ─── Configuraciones de bases de datos ────────────────────────────────────────
-- Almacena las conexiones a bases de datos configuradas por el usuario.
-- La contraseña se almacena encriptada con AES-256-GCM.
CREATE TABLE IF NOT EXISTS database_configs (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    name               TEXT    NOT NULL,
    host               TEXT    NOT NULL,
    port               INTEGER NOT NULL DEFAULT 3306,
    user               TEXT    NOT NULL,
    encrypted_password TEXT    NOT NULL,
    db_name            TEXT    NOT NULL,
    type               TEXT    NOT NULL CHECK(type IN ('mysql', 'postgres')),
    created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── Destinos de almacenamiento ────────────────────────────────────────────────
-- Almacena los destinos donde se suben los respaldos.
-- Las credenciales (para S3/FTP) se almacenan encriptadas con AES-256-GCM.
CREATE TABLE IF NOT EXISTS destinations (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    name                  TEXT    NOT NULL,
    type                  TEXT    NOT NULL CHECK(type IN ('s3', 'local', 'ftp')),
    path_or_bucket        TEXT    NOT NULL,
    encrypted_credentials TEXT    NOT NULL DEFAULT '',
    region                TEXT    NOT NULL DEFAULT '',
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── Jobs de respaldo ──────────────────────────────────────────────────────────
-- Almacena los trabajos de respaldo programados con su configuración.
-- destination_ids es un array JSON de IDs de destinos.
CREATE TABLE IF NOT EXISTS jobs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    cron_expression TEXT    NOT NULL,
    db_config_id    INTEGER NOT NULL REFERENCES database_configs(id) ON DELETE CASCADE,
    destination_ids TEXT    NOT NULL DEFAULT '[]',   -- JSON array: [1, 2, 3]
    notify_webhook  TEXT    NOT NULL DEFAULT '',
    is_active       INTEGER NOT NULL DEFAULT 1,       -- 0=false, 1=true
    last_run_at     DATETIME,
    last_status     TEXT    NOT NULL DEFAULT '',      -- 'success', 'error', 'running'
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── Configuración de la aplicación ───────────────────────────────────────────
-- Almacena parámetros de configuración en formato clave-valor.
-- Incluye el hash SHA-256 de la contraseña maestra.
CREATE TABLE IF NOT EXISTS app_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── Logs de actividad ─────────────────────────────────────────────────────────
-- Registro de actividad del sistema: ejecuciones de jobs, errores, eventos del sistema.
CREATE TABLE IF NOT EXISTS activity_logs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id     INTEGER DEFAULT 0,    -- 0 = evento del sistema (sin job asociado)
    job_name   TEXT    NOT NULL DEFAULT '',
    level      TEXT    NOT NULL DEFAULT 'info', -- 'info', 'warn', 'error', 'success'
    message    TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Índice para búsqueda eficiente de logs recientes (ordenados por fecha DESC)
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
