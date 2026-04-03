// Package entities contiene las entidades del dominio de backup-smart.
package entities

import "time"

// Job representa una tarea de respaldo automatizado programada con cron.
type Job struct {
	// ID es el identificador único en SQLite.
	ID int64 `json:"id"`
	// Name es el nombre descriptivo del job de respaldo.
	Name string `json:"name"`
	// CronExpression es la expresión cron para la programación (ej: "0 2 * * *").
	CronExpression string `json:"cron_expression"`
	// DBConfigID es la referencia a la configuración de base de datos a respaldar.
	DBConfigID int64 `json:"db_config_id"`
	// DestinationIDs lista los IDs de destinos donde se sube el respaldo.
	DestinationIDs []int64 `json:"destination_ids"`
	// NotifyWebhook es la URL de webhook para notificaciones (puede estar vacía).
	NotifyWebhook string `json:"notify_webhook,omitempty"`
	// IsActive indica si el job está habilitado en el scheduler.
	IsActive bool `json:"is_active"`
	// LastRunAt es la fecha y hora del último ejecutado.
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	// LastStatus indica el resultado del último ejecutado ("success", "error", "running").
	LastStatus string `json:"last_status,omitempty"`
	// CreatedAt es la fecha de creación del registro.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt es la fecha de última modificación.
	UpdatedAt time.Time `json:"updated_at"`
}
