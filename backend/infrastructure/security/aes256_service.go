// Package security implementa los servicios criptográficos de backup-smart.
// Utiliza AES-256-GCM para encriptación autenticada de credenciales.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// AES256Service implementa la interfaz SecurityService usando AES-256-GCM.
// El nonce se genera aleatoriamente y se antepone al cipher para almacenamiento.
type AES256Service struct{}

// NewAES256Service crea una nueva instancia del servicio de seguridad AES-256.
func NewAES256Service() *AES256Service {
	return &AES256Service{}
}

// Encrypt encripta el texto plano usando AES-256-GCM con la clave maestra.
// La clave maestra se deriva a 32 bytes usando SHA-256.
// El resultado es base64(nonce + ciphertext + tag).
func (s *AES256Service) Encrypt(text, masterKey string) (string, error) {
	if text == "" {
		return "", nil
	}
	key := deriveKey(masterKey)

	// Creamos el bloque AES con la clave de 32 bytes
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error al crear bloque AES: %w", err)
	}

	// Creamos el GCM (Galois/Counter Mode) para encriptación autenticada
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error al crear GCM: %w", err)
	}

	// Generamos un nonce aleatorio de 12 bytes (tamaño estándar de GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("error al generar nonce: %w", err)
	}

	// Encriptamos y autenticamos el texto: nonce + ciphertext + tag
	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)

	// Codificamos en base64 para almacenamiento como string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt desencripta un cipher (base64) usando AES-256-GCM con la clave maestra.
// Verifica la autenticidad del mensaje mediante el tag GCM.
func (s *AES256Service) Decrypt(cipherB64, masterKey string) (string, error) {
	if cipherB64 == "" {
		return "", nil
	}
	key := deriveKey(masterKey)

	// Decodificamos el base64
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("error al decodificar base64: %w", err)
	}

	// Creamos el bloque AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error al crear bloque AES: %w", err)
	}

	// Creamos el GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error al crear GCM: %w", err)
	}

	// Verificamos que los datos tienen el tamaño mínimo esperado
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("datos encriptados corruptos o incompletos")
	}

	// Separamos el nonce del ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Desencriptamos y verificamos la autenticidad
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error al desencriptar (clave incorrecta o datos corruptos): %w", err)
	}

	return string(plaintext), nil
}

// deriveKey deriva una clave AES de 32 bytes (256 bits) desde el masterKey
// usando SHA-256 para garantizar el tamaño correcto.
func deriveKey(masterKey string) []byte {
	sum := sha256.Sum256([]byte(masterKey))
	return sum[:]
}
