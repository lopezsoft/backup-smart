// Package ports define las interfaces (puertos) de la capa de casos de uso.
package ports

import (
	"context"

	"backup-smart/backend/domain/entities"
)

// Uploader define la interfaz para subir archivos de respaldo a un destino remoto.
type Uploader interface {
	// Upload sube el archivo en filePath al destino especificado.
	// Retorna un error si la operación falla.
	Upload(ctx context.Context, filePath string, dest entities.Destination) error
}
