// Package ports define las interfaces (puertos) de la capa de casos de uso.
// Estas interfaces son implementadas por los adaptadores de infraestructura.
package ports

import (
	"context"

	"backup-smart/backend/domain/entities"
)

// DatabaseDumper define la interfaz para generar volcados de bases de datos.
type DatabaseDumper interface {
	// Dump genera un volcado de la base de datos y retorna la ruta del archivo generado.
	// Retorna la ruta absoluta del archivo .sql/.dump generado, o un error.
	Dump(ctx context.Context, config entities.DatabaseConfig) (string, error)
}
