// Package ports define las interfaces (puertos) de la capa de casos de uso.
package ports

// SecurityService define la interfaz para operaciones criptográficas.
// Implementado por AES256Service usando AES-256-GCM.
type SecurityService interface {
	// Encrypt encripta el texto plano usando la clave maestra y retorna el cipher en base64.
	// Retorna error si la operación falla.
	Encrypt(text, masterKey string) (string, error)

	// Decrypt desencripta el cipher (base64) usando la clave maestra y retorna el texto plano.
	// Retorna error si la operación falla o los datos están corruptos.
	Decrypt(cipher, masterKey string) (string, error)
}
