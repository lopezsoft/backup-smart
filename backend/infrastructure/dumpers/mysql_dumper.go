// Package dumpers implementa los volcadores de bases de datos de backup-smart.
// Utiliza os/exec para invocar herramientas nativas del sistema (mysqldump, pg_dump).
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

// MySQLDumper implementa la interfaz DatabaseDumper para bases de datos MySQL/MariaDB.
// Requiere que mysqldump esté instalado en el sistema.
type MySQLDumper struct {
	// OutputDir es el directorio donde se guardan los volcados generados.
	OutputDir string
}

// NewMySQLDumper crea una nueva instancia del volcador de MySQL.
// outputDir es el directorio de salida para los archivos .sql generados.
func NewMySQLDumper(outputDir string) *MySQLDumper {
	return &MySQLDumper{OutputDir: outputDir}
}

// Dump genera un volcado completo de la base de datos MySQL especificada.
// Retorna la ruta absoluta del archivo .sql generado.
func (d *MySQLDumper) Dump(ctx context.Context, config entities.DatabaseConfig) (string, error) {
	// Verificamos que mysqldump esté disponible
	mysqldumpPath, err := exec.LookPath("mysqldump")
	if err != nil {
		return "", fmt.Errorf("mysqldump no encontrado en el sistema: %w", err)
	}

	// Creamos el directorio de salida si no existe
	if err := os.MkdirAll(d.OutputDir, 0750); err != nil {
		return "", fmt.Errorf("error al crear directorio de salida: %w", err)
	}

	// Generamos un nombre de archivo único con timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.sql", config.DBName, config.Type, timestamp)
	outputPath := filepath.Join(d.OutputDir, filename)

	// Construimos los argumentos de mysqldump
	// Nota: la contraseña se pasa con --password= para evitar exposición en la línea de comandos
	args := []string{
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		fmt.Sprintf("--user=%s", config.User),
		fmt.Sprintf("--password=%s", config.EncryptedPassword), // Ya desencriptada
		"--single-transaction",
		"--routines",
		"--triggers",
		"--add-drop-table",
		config.DBName,
	}

	// Creamos el archivo de salida
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("error al crear archivo de volcado: %w", err)
	}
	defer outFile.Close()

	// Ejecutamos mysqldump redirigiendo stdout al archivo
	cmd := exec.CommandContext(ctx, mysqldumpPath, args...)
	cmd.Stdout = outFile
	// Capturamos stderr para mensajes de error
	stderrOutput, err := cmd.CombinedOutput()
	if err != nil {
		// Eliminamos el archivo incompleto en caso de error
		os.Remove(outputPath)
		return "", fmt.Errorf("error al ejecutar mysqldump: %w - %s", err, string(stderrOutput))
	}

	return outputPath, nil
}
