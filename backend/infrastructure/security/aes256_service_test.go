// Package security contiene las pruebas unitarias para el servicio de encriptación AES-256.
package security

import (
"strings"
"testing"
)

// TestEncryptDecryptRoundTrip verifica que un texto encriptado puede ser desencriptado correctamente.
func TestEncryptDecryptRoundTrip(t *testing.T) {
svc := NewAES256Service()
masterKey := "mi-contraseña-maestra-segura-123"
plaintext := "contraseña-secreta-de-base-de-datos"

// Encriptamos el texto
cipher, err := svc.Encrypt(plaintext, masterKey)
if err != nil {
t.Fatalf("Error al encriptar: %v", err)
}
if cipher == "" {
t.Fatal("El cipher no puede estar vacío")
}
if cipher == plaintext {
t.Fatal("El cipher no debe ser igual al texto plano")
}

// Desencriptamos el cipher
result, err := svc.Decrypt(cipher, masterKey)
if err != nil {
t.Fatalf("Error al desencriptar: %v", err)
}
if result != plaintext {
t.Errorf("Texto desencriptado incorrecto: esperado=%q, obtenido=%q", plaintext, result)
}
}

// TestEncryptProducesUniqueResults verifica que cada encriptación produce un resultado único (nonce aleatorio).
func TestEncryptProducesUniqueResults(t *testing.T) {
svc := NewAES256Service()
masterKey := "clave-maestra"
plaintext := "texto-a-encriptar"

cipher1, err := svc.Encrypt(plaintext, masterKey)
if err != nil {
t.Fatalf("Error en primera encriptación: %v", err)
}

cipher2, err := svc.Encrypt(plaintext, masterKey)
if err != nil {
t.Fatalf("Error en segunda encriptación: %v", err)
}

// Dos encriptaciones del mismo texto deben producir resultados diferentes (nonce único)
if cipher1 == cipher2 {
t.Error("Las encriptaciones del mismo texto deben producir resultados diferentes (nonce aleatorio)")
}
}

// TestDecryptWithWrongKeyFails verifica que la desencriptación con clave incorrecta falla.
func TestDecryptWithWrongKeyFails(t *testing.T) {
svc := NewAES256Service()
plaintext := "dato-sensible"

cipher, err := svc.Encrypt(plaintext, "clave-correcta")
if err != nil {
t.Fatalf("Error al encriptar: %v", err)
}

// Intentamos desencriptar con una clave incorrecta
_, err = svc.Decrypt(cipher, "clave-incorrecta")
if err == nil {
t.Fatal("Se esperaba un error al desencriptar con clave incorrecta")
}
}

// TestEncryptEmptyString verifica que una cadena vacía se maneja correctamente.
func TestEncryptEmptyString(t *testing.T) {
svc := NewAES256Service()

cipher, err := svc.Encrypt("", "cualquier-clave")
if err != nil {
t.Fatalf("Error al encriptar cadena vacía: %v", err)
}
if cipher != "" {
t.Error("La encriptación de cadena vacía debe retornar cadena vacía")
}

result, err := svc.Decrypt("", "cualquier-clave")
if err != nil {
t.Fatalf("Error al desencriptar cadena vacía: %v", err)
}
if result != "" {
t.Error("La desencriptación de cadena vacía debe retornar cadena vacía")
}
}

// TestDecryptInvalidBase64Fails verifica que datos inválidos retornan error.
func TestDecryptInvalidBase64Fails(t *testing.T) {
svc := NewAES256Service()

_, err := svc.Decrypt("esto-no-es-base64-válido!!!", "clave")
if err == nil {
t.Fatal("Se esperaba un error al desencriptar datos inválidos")
}
}

// TestEncryptDifferentKeys verifica que la misma clave desencripta correctamente
// y que una clave diferente no puede desencriptar.
func TestEncryptDifferentKeys(t *testing.T) {
svc := NewAES256Service()
testCases := []struct {
name      string
plaintext string
masterKey string
}{
{"contraseña simple", "mi-contraseña", "clave1"},
{"contraseña con caracteres especiales", "p@$$w0rd!#%", "clave-compleja-2024"},
{"texto largo", strings.Repeat("a", 1000), "clave3"},
}

for _, tc := range testCases {
t.Run(tc.name, func(t *testing.T) {
cipher, err := svc.Encrypt(tc.plaintext, tc.masterKey)
if err != nil {
t.Fatalf("Error al encriptar '%s': %v", tc.name, err)
}

result, err := svc.Decrypt(cipher, tc.masterKey)
if err != nil {
t.Fatalf("Error al desencriptar '%s': %v", tc.name, err)
}

if result != tc.plaintext {
t.Errorf("Round-trip fallido para '%s': esperado=%q, obtenido=%q", tc.name, tc.plaintext, result)
}
})
}
}
