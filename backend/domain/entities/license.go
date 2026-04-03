// Package entities contiene las entidades del dominio de backup-smart.
package entities

import "time"

// License representa el estado de la licencia del software en la máquina actual.
type License struct {
	// MachineID es el identificador de hardware único del equipo.
	MachineID string `json:"machine_id"`
	// LicenseKey es la clave de licencia ingresada por el usuario.
	LicenseKey string `json:"license_key"`
	// SignedToken es el token JWT/Ed25519 firmado que representa la licencia válida.
	SignedToken string `json:"signed_token,omitempty"`
	// IsValid indica si la licencia es válida en el momento de la verificación.
	IsValid bool `json:"is_valid"`
	// LastChecked es la fecha y hora de la última verificación de licencia.
	LastChecked time.Time `json:"last_checked"`
	// ExpirationDate es la fecha de vencimiento de la licencia.
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	// GracePeriodEnd es la fecha límite del período de gracia offline (14 días).
	GracePeriodEnd *time.Time `json:"grace_period_end,omitempty"`
	// Plan indica el plan de licencia ("free", "pro", "enterprise").
	Plan string `json:"plan"`
}

// LogEntry representa una entrada en el registro de actividad de la aplicación.
type LogEntry struct {
	// ID es el identificador único del log.
	ID int64 `json:"id"`
	// JobID es la referencia al job que generó el log (0 si es del sistema).
	JobID int64 `json:"job_id,omitempty"`
	// JobName es el nombre del job para mostrar en la UI.
	JobName string `json:"job_name,omitempty"`
	// Level indica el nivel del log ("info", "warn", "error", "success").
	Level string `json:"level"`
	// Message es el mensaje descriptivo del evento.
	Message string `json:"message"`
	// CreatedAt es la fecha y hora del evento.
	CreatedAt time.Time `json:"created_at"`
}
