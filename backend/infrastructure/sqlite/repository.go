// Package sqlite implementa el repositorio de datos usando SQLite.
// Utiliza database/sql con el driver go-sqlite3 para operaciones CRUD.
package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"backup-smart/backend/domain/entities"

	_ "github.com/mattn/go-sqlite3" // Driver SQLite3
)

// Repository implementa las operaciones CRUD para todas las entidades en SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository abre (o crea) la base de datos SQLite en la ruta especificada.
// Configura el modo WAL para mejor concurrencia.
func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al verificar conexión con SQLite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite no soporta múltiples escritores simultáneos
	return &Repository{db: db}, nil
}

// Close cierra la conexión a la base de datos.
func (r *Repository) Close() error {
	return r.db.Close()
}

// Migrate ejecuta las migraciones del esquema de la base de datos.
// Es idempotente: puede ejecutarse múltiples veces sin efectos secundarios.
func (r *Repository) Migrate() error {
	schema := `
	-- Tabla de configuraciones de bases de datos
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

	-- Tabla de destinos de almacenamiento
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

	-- Tabla de jobs de respaldo
	CREATE TABLE IF NOT EXISTS jobs (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		name            TEXT    NOT NULL,
		cron_expression TEXT    NOT NULL,
		db_config_id    INTEGER NOT NULL REFERENCES database_configs(id) ON DELETE CASCADE,
		destination_ids TEXT    NOT NULL DEFAULT '[]',
		notify_webhook  TEXT    NOT NULL DEFAULT '',
		is_active       INTEGER NOT NULL DEFAULT 1,
		last_run_at     DATETIME,
		last_status     TEXT    NOT NULL DEFAULT '',
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Tabla de configuración de la aplicación (master password hash, etc.)
	CREATE TABLE IF NOT EXISTS app_config (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Tabla de logs de actividad
	CREATE TABLE IF NOT EXISTS activity_logs (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id     INTEGER DEFAULT 0,
		job_name   TEXT    NOT NULL DEFAULT '',
		level      TEXT    NOT NULL DEFAULT 'info',
		message    TEXT    NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Índice para búsqueda eficiente de logs recientes
	CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
	`
	if _, err := r.db.Exec(schema); err != nil {
		return fmt.Errorf("error al ejecutar migraciones: %w", err)
	}
	return nil
}

// ─── MasterPassword ───────────────────────────────────────────────────────────

// SetMasterPassword guarda el hash SHA-256 de la contraseña maestra en la configuración.
func (r *Repository) SetMasterPassword(password string) error {
	hash := hashPassword(password)
	_, err := r.db.Exec(
		`INSERT INTO app_config (key, value, updated_at) VALUES ('master_password_hash', ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		hash, time.Now(),
	)
	return err
}

// ValidateMasterPassword verifica si la contraseña proporcionada coincide con el hash almacenado.
func (r *Repository) ValidateMasterPassword(password string) (bool, error) {
	var storedHash string
	err := r.db.QueryRow(`SELECT value FROM app_config WHERE key='master_password_hash'`).Scan(&storedHash)
	if err == sql.ErrNoRows {
		// Si no hay contraseña configurada, la guardamos y retornamos true
		return true, r.SetMasterPassword(password)
	}
	if err != nil {
		return false, fmt.Errorf("error al consultar contraseña maestra: %w", err)
	}
	return hashPassword(password) == storedHash, nil
}

// hashPassword calcula el hash SHA-256 de una contraseña.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// ─── DatabaseConfig CRUD ──────────────────────────────────────────────────────

// GetAllDatabaseConfigs retorna todas las configuraciones de bases de datos.
func (r *Repository) GetAllDatabaseConfigs() ([]entities.DatabaseConfig, error) {
	rows, err := r.db.Query(`SELECT id, name, host, port, user, encrypted_password, db_name, type, created_at, updated_at FROM database_configs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("error al consultar database_configs: %w", err)
	}
	defer rows.Close()

	var configs []entities.DatabaseConfig
	for rows.Next() {
		var c entities.DatabaseConfig
		if err := rows.Scan(&c.ID, &c.Name, &c.Host, &c.Port, &c.User, &c.EncryptedPassword, &c.DBName, &c.Type, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al leer database_config: %w", err)
		}
		configs = append(configs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando database_configs: %w", err)
	}
	return configs, nil
}

// GetDatabaseConfigByID retorna una configuración de base de datos por su ID.
func (r *Repository) GetDatabaseConfigByID(id int64) (entities.DatabaseConfig, error) {
	var c entities.DatabaseConfig
	err := r.db.QueryRow(`SELECT id, name, host, port, user, encrypted_password, db_name, type, created_at, updated_at FROM database_configs WHERE id=?`, id).
		Scan(&c.ID, &c.Name, &c.Host, &c.Port, &c.User, &c.EncryptedPassword, &c.DBName, &c.Type, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return c, fmt.Errorf("configuración de base de datos no encontrada: id=%d", id)
	}
	if err != nil {
		return c, fmt.Errorf("error al consultar database_config: %w", err)
	}
	return c, nil
}

// CreateDatabaseConfig inserta una nueva configuración de base de datos.
func (r *Repository) CreateDatabaseConfig(c entities.DatabaseConfig) error {
	now := time.Now()
	_, err := r.db.Exec(
		`INSERT INTO database_configs (name, host, port, user, encrypted_password, db_name, type, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		c.Name, c.Host, c.Port, c.User, c.EncryptedPassword, c.DBName, c.Type, now, now,
	)
	if err != nil {
		return fmt.Errorf("error al crear database_config: %w", err)
	}
	return nil
}

// UpdateDatabaseConfig actualiza una configuración de base de datos existente.
func (r *Repository) UpdateDatabaseConfig(c entities.DatabaseConfig) error {
	result, err := r.db.Exec(
		`UPDATE database_configs SET name=?, host=?, port=?, user=?, encrypted_password=?, db_name=?, type=?, updated_at=? WHERE id=?`,
		c.Name, c.Host, c.Port, c.User, c.EncryptedPassword, c.DBName, c.Type, time.Now(), c.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar database_config: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("database_config no encontrado: id=%d", c.ID)
	}
	return nil
}

// DeleteDatabaseConfig elimina una configuración de base de datos por su ID.
func (r *Repository) DeleteDatabaseConfig(id int64) error {
	result, err := r.db.Exec(`DELETE FROM database_configs WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("error al eliminar database_config: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("database_config no encontrado: id=%d", id)
	}
	return nil
}

// ─── Destination CRUD ─────────────────────────────────────────────────────────

// GetAllDestinations retorna todos los destinos de almacenamiento.
func (r *Repository) GetAllDestinations() ([]entities.Destination, error) {
	rows, err := r.db.Query(`SELECT id, name, type, path_or_bucket, encrypted_credentials, region, created_at, updated_at FROM destinations ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("error al consultar destinations: %w", err)
	}
	defer rows.Close()

	var dests []entities.Destination
	for rows.Next() {
		var d entities.Destination
		if err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.PathOrBucket, &d.EncryptedCredentials, &d.Region, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al leer destination: %w", err)
		}
		dests = append(dests, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando destinations: %w", err)
	}
	return dests, nil
}

// GetDestinationByID retorna un destino de almacenamiento por su ID.
func (r *Repository) GetDestinationByID(id int64) (entities.Destination, error) {
	var d entities.Destination
	err := r.db.QueryRow(`SELECT id, name, type, path_or_bucket, encrypted_credentials, region, created_at, updated_at FROM destinations WHERE id=?`, id).
		Scan(&d.ID, &d.Name, &d.Type, &d.PathOrBucket, &d.EncryptedCredentials, &d.Region, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return d, fmt.Errorf("destino no encontrado: id=%d", id)
	}
	if err != nil {
		return d, fmt.Errorf("error al consultar destination: %w", err)
	}
	return d, nil
}

// CreateDestination inserta un nuevo destino de almacenamiento.
func (r *Repository) CreateDestination(d entities.Destination) error {
	now := time.Now()
	_, err := r.db.Exec(
		`INSERT INTO destinations (name, type, path_or_bucket, encrypted_credentials, region, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		d.Name, d.Type, d.PathOrBucket, d.EncryptedCredentials, d.Region, now, now,
	)
	if err != nil {
		return fmt.Errorf("error al crear destination: %w", err)
	}
	return nil
}

// UpdateDestination actualiza un destino de almacenamiento existente.
func (r *Repository) UpdateDestination(d entities.Destination) error {
	result, err := r.db.Exec(
		`UPDATE destinations SET name=?, type=?, path_or_bucket=?, encrypted_credentials=?, region=?, updated_at=? WHERE id=?`,
		d.Name, d.Type, d.PathOrBucket, d.EncryptedCredentials, d.Region, time.Now(), d.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar destination: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("destino no encontrado: id=%d", d.ID)
	}
	return nil
}

// DeleteDestination elimina un destino por su ID.
func (r *Repository) DeleteDestination(id int64) error {
	result, err := r.db.Exec(`DELETE FROM destinations WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("error al eliminar destination: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("destino no encontrado: id=%d", id)
	}
	return nil
}

// ─── Job CRUD ─────────────────────────────────────────────────────────────────

// GetAllJobs retorna todos los jobs de respaldo con sus IDs de destinos desserializados.
func (r *Repository) GetAllJobs() ([]entities.Job, error) {
	rows, err := r.db.Query(`SELECT id, name, cron_expression, db_config_id, destination_ids, notify_webhook, is_active, last_run_at, last_status, created_at, updated_at FROM jobs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("error al consultar jobs: %w", err)
	}
	defer rows.Close()

	var jobs []entities.Job
	for rows.Next() {
		var j entities.Job
		var destIDsJSON string
		var lastRunAt sql.NullTime
		if err := rows.Scan(&j.ID, &j.Name, &j.CronExpression, &j.DBConfigID, &destIDsJSON, &j.NotifyWebhook, &j.IsActive, &lastRunAt, &j.LastStatus, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al leer job: %w", err)
		}
		if lastRunAt.Valid {
			j.LastRunAt = &lastRunAt.Time
		}
		// Deserializamos el JSON de IDs de destinos
		if err := json.Unmarshal([]byte(destIDsJSON), &j.DestinationIDs); err != nil {
			j.DestinationIDs = []int64{}
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando jobs: %w", err)
	}
	return jobs, nil
}

// GetJobByID retorna un job por su ID.
func (r *Repository) GetJobByID(id int64) (entities.Job, error) {
	var j entities.Job
	var destIDsJSON string
	var lastRunAt sql.NullTime

	err := r.db.QueryRow(`SELECT id, name, cron_expression, db_config_id, destination_ids, notify_webhook, is_active, last_run_at, last_status, created_at, updated_at FROM jobs WHERE id=?`, id).
		Scan(&j.ID, &j.Name, &j.CronExpression, &j.DBConfigID, &destIDsJSON, &j.NotifyWebhook, &j.IsActive, &lastRunAt, &j.LastStatus, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return j, fmt.Errorf("job no encontrado: id=%d", id)
	}
	if err != nil {
		return j, fmt.Errorf("error al consultar job: %w", err)
	}
	if lastRunAt.Valid {
		j.LastRunAt = &lastRunAt.Time
	}
	if err := json.Unmarshal([]byte(destIDsJSON), &j.DestinationIDs); err != nil {
		j.DestinationIDs = []int64{}
	}
	return j, nil
}

// CreateJob inserta un nuevo job de respaldo.
func (r *Repository) CreateJob(j entities.Job) error {
	destIDsJSON, err := json.Marshal(j.DestinationIDs)
	if err != nil {
		return fmt.Errorf("error al serializar destination_ids: %w", err)
	}
	now := time.Now()
	_, err = r.db.Exec(
		`INSERT INTO jobs (name, cron_expression, db_config_id, destination_ids, notify_webhook, is_active, last_status, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		j.Name, j.CronExpression, j.DBConfigID, string(destIDsJSON), j.NotifyWebhook, j.IsActive, "", now, now,
	)
	if err != nil {
		return fmt.Errorf("error al crear job: %w", err)
	}
	return nil
}

// UpdateJob actualiza un job de respaldo existente.
func (r *Repository) UpdateJob(j entities.Job) error {
	destIDsJSON, err := json.Marshal(j.DestinationIDs)
	if err != nil {
		return fmt.Errorf("error al serializar destination_ids: %w", err)
	}
	result, err := r.db.Exec(
		`UPDATE jobs SET name=?, cron_expression=?, db_config_id=?, destination_ids=?, notify_webhook=?, is_active=?, updated_at=? WHERE id=?`,
		j.Name, j.CronExpression, j.DBConfigID, string(destIDsJSON), j.NotifyWebhook, j.IsActive, time.Now(), j.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar job: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("job no encontrado: id=%d", j.ID)
	}
	return nil
}

// UpdateJobStatus actualiza el estado y la fecha del último ejecutado de un job.
func (r *Repository) UpdateJobStatus(jobID int64, status string, runAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE jobs SET last_run_at=?, last_status=?, updated_at=? WHERE id=?`,
		runAt, status, time.Now(), jobID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar estado del job: %w", err)
	}
	return nil
}

// DeleteJob elimina un job por su ID.
func (r *Repository) DeleteJob(id int64) error {
	result, err := r.db.Exec(`DELETE FROM jobs WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("error al eliminar job: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("job no encontrado: id=%d", id)
	}
	return nil
}

// ─── ActivityLog ──────────────────────────────────────────────────────────────

// CreateLog inserta una nueva entrada en el registro de actividad.
func (r *Repository) CreateLog(entry entities.LogEntry) error {
	_, err := r.db.Exec(
		`INSERT INTO activity_logs (job_id, job_name, level, message, created_at) VALUES (?,?,?,?,?)`,
		entry.JobID, entry.JobName, entry.Level, entry.Message, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error al crear log: %w", err)
	}
	return nil
}

// GetRecentLogs retorna los últimos N registros de actividad ordenados por fecha descendente.
func (r *Repository) GetRecentLogs(limit int) ([]entities.LogEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, job_id, job_name, level, message, created_at FROM activity_logs ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("error al consultar logs: %w", err)
	}
	defer rows.Close()

	var logs []entities.LogEntry
	for rows.Next() {
		var l entities.LogEntry
		if err := rows.Scan(&l.ID, &l.JobID, &l.JobName, &l.Level, &l.Message, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("error al leer log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando logs: %w", err)
	}
	return logs, nil
}
