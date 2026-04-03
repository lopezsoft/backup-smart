// Package entities contiene las entidades del dominio de backup-smart.
package entities

import "time"

// DestinationType representa el tipo de destino de almacenamiento.
type DestinationType string

const (
	// DestinationS3 indica un bucket de Amazon S3.
	DestinationS3 DestinationType = "s3"
	// DestinationLocal indica un directorio local del sistema de archivos.
	DestinationLocal DestinationType = "local"
	// DestinationFTP indica un servidor FTP.
	DestinationFTP DestinationType = "ftp"
)

// Destination almacena la configuración de un destino de almacenamiento para los respaldos.
// Las credenciales se guardan encriptadas con AES-256 usando el MasterPassword.
type Destination struct {
	// ID es el identificador único en SQLite.
	ID int64 `json:"id"`
	// Name es el nombre descriptivo del destino.
	Name string `json:"name"`
	// Type indica el tipo de destino (s3, local, ftp).
	Type DestinationType `json:"type"`
	// PathOrBucket es la ruta local, bucket de S3 o directorio FTP.
	PathOrBucket string `json:"path_or_bucket"`
	// EncryptedCredentials contiene las credenciales encriptadas (JSON con keys/secrets).
	EncryptedCredentials string `json:"encrypted_credentials"`
	// Region es la región de AWS (solo para S3).
	Region string `json:"region,omitempty"`
	// CreatedAt es la fecha de creación del registro.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt es la fecha de última modificación.
	UpdatedAt time.Time `json:"updated_at"`
}
