// Package dumpers implementa los volcadores de bases de datos de backup-smart.
package dumpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"backup-smart/backend/domain/entities"
)

// PostgresDumper implementa la interfaz DatabaseDumper para bases de datos PostgreSQL.
// Requiere que pg_dump esté instalado en el sistema.
type PostgresDumper struct {
	// OutputDir es el directorio donde se guardan los volcados generados.
	OutputDir string
}

// NewPostgresDumper crea una nueva instancia del volcador de PostgreSQL.
// outputDir es el directorio de salida para los archivos .dump generados.
func NewPostgresDumper(outputDir string) *PostgresDumper {
	return &PostgresDumper{OutputDir: outputDir}
}

// Dump genera un volcado completo de la base de datos PostgreSQL especificada.
// Utiliza el formato custom de pg_dump para permitir restauración selectiva.
// Retorna la ruta absoluta del archivo .dump generado.
func (d *PostgresDumper) Dump(ctx context.Context, config entities.DatabaseConfig) (string, error) {
	// Verificamos que pg_dump esté disponible
	pgDumpPath, err := exec.LookPath("pg_dump")
	if err != nil {
		return "", fmt.Errorf("pg_dump no encontrado en el sistema: %w", err)
	}

	// Creamos el directorio de salida si no existe
	if err := os.MkdirAll(d.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("error al crear directorio de salida: %w", err)
	}

	// Generamos un nombre de archivo único con timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.dump", config.DBName, config.Type, timestamp)
	outputPath := filepath.Join(d.OutputDir, filename)

	// Construimos la cadena de conexión PostgreSQL (DSN)
	// Usamos PGPASSWORD como variable de entorno para evitar exposición de contraseña
	dsn := fmt.Sprintf("postgresql://%s@%s:%d/%s",
		config.User, config.Host, config.Port, config.DBName)

	// Argumentos de pg_dump: formato custom para máxima compresión y flexibilidad
	args := []string{
		"--format=custom",
		"--compress=9",
		"--no-password",
		fmt.Sprintf("--file=%s", outputPath),
		dsn,
	}

	// Ejecutamos pg_dump con la contraseña en variable de entorno
	cmd := exec.CommandContext(ctx, pgDumpPath, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", config.EncryptedPassword), // Ya desencriptada
	)

	// Capturamos stderr para mensajes de error
	if output, err := cmd.CombinedOutput(); err != nil {
		// Eliminamos el archivo incompleto en caso de error
		os.Remove(outputPath)
		return "", fmt.Errorf("error al ejecutar pg_dump: %w - %s", err, string(output))
	}

	return outputPath, nil
}
