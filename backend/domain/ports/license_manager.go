// Package ports define las interfaces (puertos) de la capa de casos de uso.
package ports

import "backup-smart/backend/domain/entities"

// LicenseManager define la interfaz para la gestión del licenciamiento de la aplicación.
// Soporta validación online/offline con un período de gracia de 14 días.
type LicenseManager interface {
	// VerifyLicense valida una clave de licencia y retorna el estado detallado.
	// Intenta validación online; si falla, usa la última validación con período de gracia.
	VerifyLicense(licenseKey string) (entities.License, error)

	// GetHardwareID retorna el identificador de hardware único del equipo actual.
	// Utilizado para anclar la licencia a un dispositivo específico.
	GetHardwareID() string
}
