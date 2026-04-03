// Package storage implementa los adaptadores de almacenamiento para backup-smart.
// Incluye soporte para Amazon S3 y almacenamiento local.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"backup-smart/backend/domain/entities"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// s3Credentials representa las credenciales de AWS S3 almacenadas en JSON.
type s3Credentials struct {
	// AccessKeyID es la clave de acceso de AWS IAM.
	AccessKeyID string `json:"access_key_id"`
	// SecretAccessKey es la clave secreta de AWS IAM.
	SecretAccessKey string `json:"secret_access_key"`
}

// S3Uploader implementa la interfaz Uploader para subir archivos a Amazon S3.
type S3Uploader struct{}

// NewS3Uploader crea una nueva instancia del uploader de S3.
func NewS3Uploader() *S3Uploader {
	return &S3Uploader{}
}

// Upload sube el archivo al bucket de S3 especificado en el destino.
// Las credenciales se deserializan desde el campo EncryptedCredentials (ya desencriptado).
func (u *S3Uploader) Upload(ctx context.Context, filePath string, dest entities.Destination) error {
	// Deserializamos las credenciales de AWS desde el JSON
	var creds s3Credentials
	if err := json.Unmarshal([]byte(dest.EncryptedCredentials), &creds); err != nil {
		return fmt.Errorf("error al deserializar credenciales S3: %w", err)
	}

	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return fmt.Errorf("las credenciales de S3 son inválidas o están vacías")
	}

	// Configuramos la región (default us-east-1 si no se especifica)
	region := dest.Region
	if region == "" {
		region = "us-east-1"
	}

	// Creamos la sesión de AWS con las credenciales proporcionadas
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			"", // Token de sesión (solo para credenciales temporales)
		),
	})
	if err != nil {
		return fmt.Errorf("error al crear sesión de AWS: %w", err)
	}

	// Abrimos el archivo local a subir
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir archivo para subir: %w", err)
	}
	defer file.Close()

	// Obtenemos la información del archivo para el Content-Length
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error al obtener información del archivo: %w", err)
	}

	// Generamos la clave en S3 con la fecha para organizar los backups
	datePrefix := time.Now().Format("2006/01/02")
	s3Key := fmt.Sprintf("%s/%s", datePrefix, filepath.Base(filePath))

	// Subimos el archivo a S3
	svc := s3.New(sess)
	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(dest.PathOrBucket),
		Key:           aws.String(s3Key),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("error al subir archivo a S3 (bucket: %s, key: %s): %w", dest.PathOrBucket, s3Key, err)
	}

	return nil
}

// LocalUploader implementa la interfaz Uploader para copiar archivos a un directorio local.
type LocalUploader struct{}

// NewLocalUploader crea una nueva instancia del uploader local.
func NewLocalUploader() *LocalUploader {
	return &LocalUploader{}
}

// Upload copia el archivo al directorio local especificado en el destino.
func (u *LocalUploader) Upload(_ context.Context, filePath string, dest entities.Destination) error {
	// Creamos el directorio de destino con subdirectorio de fecha
	dateDir := filepath.Join(dest.PathOrBucket, time.Now().Format("2006/01/02"))
	if err := os.MkdirAll(dateDir, 0750); err != nil {
		return fmt.Errorf("error al crear directorio de destino: %w", err)
	}

	// Construimos la ruta de destino
	destPath := filepath.Join(dateDir, filepath.Base(filePath))

	// Abrimos el archivo origen
	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir archivo origen: %w", err)
	}
	defer src.Close()

	// Creamos el archivo destino
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error al crear archivo destino: %w", err)
	}
	defer dst.Close()

	// Copiamos el contenido
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath) // Eliminamos el archivo incompleto en caso de error
		return fmt.Errorf("error al copiar archivo: %w", err)
	}

	return nil
}
