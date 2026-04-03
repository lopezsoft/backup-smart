// Package entities contiene las entidades del dominio de backup-smart.
// Son los objetos de negocio fundamentales sin dependencias externas.
package entities

import "time"

// DBType representa el tipo de motor de base de datos soportado.
type DBType string

const (
	// DBTypeMySQL indica una base de datos MySQL/MariaDB.
	DBTypeMySQL DBType = "mysql"
	// DBTypePostgres indica una base de datos PostgreSQL.
	DBTypePostgres DBType = "postgres"
)

// DatabaseConfig almacena la configuración de conexión a una base de datos.
// La contraseña se guarda encriptada con AES-256 usando el MasterPassword del usuario.
type DatabaseConfig struct {
	// ID es el identificador único en SQLite.
	ID int64 `json:"id"`
	// Name es el nombre descriptivo de la configuración.
	Name string `json:"name"`
	// Host es la dirección del servidor de base de datos.
	Host string `json:"host"`
	// Port es el puerto del servidor de base de datos.
	Port int `json:"port"`
	// User es el usuario de conexión a la base de datos.
	User string `json:"user"`
	// EncryptedPassword contiene la contraseña encriptada con AES-256.
	EncryptedPassword string `json:"encrypted_password"`
	// DBName es el nombre de la base de datos a respaldar.
	DBName string `json:"db_name"`
	// Type indica el motor de base de datos (mysql/postgres).
	Type DBType `json:"type"`
	// CreatedAt es la fecha de creación del registro.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt es la fecha de última modificación.
	UpdatedAt time.Time `json:"updated_at"`
}
