// Package licensing implementa la gestión del licenciamiento de backup-smart.
// Soporta validación de licencias con hardware fingerprinting y período de gracia offline.
package licensing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"backup-smart/backend/domain/entities"

	"github.com/denisbrodbeck/machineid"
)

// licensePayload representa el contenido decodificado de un token de licencia.
type licensePayload struct {
	// MachineID es el ID de hardware para el que se emitió la licencia.
	MachineID string `json:"machine_id"`
	// LicenseKey es la clave de licencia única.
	LicenseKey string `json:"license_key"`
	// Plan es el nivel de la licencia ("free", "pro", "enterprise").
	Plan string `json:"plan"`
	// ExpirationDate es la fecha de vencimiento en formato RFC3339.
	ExpirationDate string `json:"expiration_date,omitempty"`
	// IssuedAt es la fecha de emisión de la licencia.
	IssuedAt string `json:"issued_at"`
}

// publicKeyB64 es la clave pública Ed25519 embebida para verificar firmas de licencias.
// En producción, esta clave debe corresponder a la clave privada usada para firmar licencias.
// NOTA: Esta es una clave de demostración. Reemplazar con la clave real en producción.
const publicKeyB64 = "MCowBQYDK2VwAyEA3pJ7epVECiQaS2HHmKO8ICEbOCRkpV7kpKHLYhCdOFU="

// gracePeriodDays es el número de días de período de gracia para validación offline.
const gracePeriodDays = 14

// LicensingService implementa la interfaz LicenseManager para gestión de licencias.
type LicensingService struct {
	cachedLicense *entities.License
	publicKey     ed25519.PublicKey
}

// NewLicensingService crea una nueva instancia del servicio de licenciamiento.
func NewLicensingService() *LicensingService {
	svc := &LicensingService{}

	// Intentamos cargar la clave pública embebida
	pubKey, err := loadPublicKey(publicKeyB64)
	if err == nil {
		svc.publicKey = pubKey
	}

	return svc
}

// GetHardwareID retorna el identificador de hardware único del equipo actual.
// Utiliza la librería machineid que genera un ID estable basado en características del hardware.
func (s *LicensingService) GetHardwareID() string {
	id, err := machineid.ProtectedID("backup-smart")
	if err != nil {
		// Si falla la lectura del hardware ID, retornamos un identificador de error
		return fmt.Sprintf("error-hardware-id: %v", err)
	}
	return id
}

// VerifyLicense valida una clave de licencia y retorna el estado detallado.
// El formato de la licenseKey es: base64(payload_json) + "." + base64(signature_ed25519)
func (s *LicensingService) VerifyLicense(licenseKey string) (entities.License, error) {
	now := time.Now()
	gracePeriodEnd := now.Add(gracePeriodDays * 24 * time.Hour)

	// Si hay una licencia en caché válida, la retornamos directamente
	if s.cachedLicense != nil && s.cachedLicense.IsValid {
		timeSinceCheck := now.Sub(s.cachedLicense.LastChecked)
		if timeSinceCheck < 24*time.Hour {
			return *s.cachedLicense, nil
		}
	}

	// Licencia vacía = modo libre sin licencia
	if licenseKey == "" {
		return entities.License{
			MachineID:   s.GetHardwareID(),
			IsValid:     false,
			LastChecked: now,
			Plan:        "free",
		}, nil
	}

	// Parseamos y verificamos la licencia
	payload, err := s.parseLicense(licenseKey)
	if err != nil {
		// Intentamos usar la caché con período de gracia si existe
		if s.cachedLicense != nil && s.cachedLicense.GracePeriodEnd != nil {
			if now.Before(*s.cachedLicense.GracePeriodEnd) {
				return *s.cachedLicense, nil
			}
		}
		return entities.License{
			MachineID:   s.GetHardwareID(),
			LicenseKey:  licenseKey,
			IsValid:     false,
			LastChecked: now,
			Plan:        "free",
		}, fmt.Errorf("licencia inválida: %w", err)
	}

	// Verificamos que el hardware ID coincida
	currentHardwareID := s.GetHardwareID()
	if payload.MachineID != "" && payload.MachineID != currentHardwareID {
		return entities.License{
			MachineID:   currentHardwareID,
			LicenseKey:  licenseKey,
			IsValid:     false,
			LastChecked: now,
			Plan:        "free",
		}, fmt.Errorf("la licencia no corresponde a este equipo (hardware ID no coincide)")
	}

	// Verificamos la fecha de vencimiento
	var expDate *time.Time
	if payload.ExpirationDate != "" {
		exp, err := time.Parse(time.RFC3339, payload.ExpirationDate)
		if err != nil {
			return entities.License{}, fmt.Errorf("fecha de vencimiento inválida en la licencia: %w", err)
		}
		expDate = &exp
		if now.After(exp) {
			return entities.License{
				MachineID:      currentHardwareID,
				LicenseKey:     licenseKey,
				IsValid:        false,
				LastChecked:    now,
				ExpirationDate: expDate,
				Plan:           payload.Plan,
			}, fmt.Errorf("la licencia ha vencido el %s", exp.Format("02/01/2006"))
		}
	}

	// Licencia válida
	license := entities.License{
		MachineID:      currentHardwareID,
		LicenseKey:     licenseKey,
		IsValid:        true,
		LastChecked:    now,
		ExpirationDate: expDate,
		GracePeriodEnd: &gracePeriodEnd,
		Plan:           payload.Plan,
	}

	// Guardamos en caché para validación offline posterior
	s.cachedLicense = &license
	return license, nil
}

// parseLicense decodifica y verifica la firma de una clave de licencia.
// Formato: base64url(json_payload).base64url(ed25519_signature)
func (s *LicensingService) parseLicense(licenseKey string) (*licensePayload, error) {
	// Buscamos el separador entre payload y firma
	dotIdx := -1
	for i := len(licenseKey) - 1; i >= 0; i-- {
		if licenseKey[i] == '.' {
			dotIdx = i
			break
		}
	}

	if dotIdx < 0 {
		return nil, fmt.Errorf("formato de licencia inválido (falta separador)")
	}

	payloadB64 := licenseKey[:dotIdx]
	sigB64 := licenseKey[dotIdx+1:]

	// Decodificamos el payload JSON
	payloadJSON, err := base64.URLEncoding.DecodeString(payloadB64)
	if err != nil {
		// Intentamos con base64 estándar
		payloadJSON, err = base64.StdEncoding.DecodeString(payloadB64)
		if err != nil {
			return nil, fmt.Errorf("error al decodificar payload de licencia: %w", err)
		}
	}

	// Decodificamos la firma Ed25519
	sig, err := base64.URLEncoding.DecodeString(sigB64)
	if err != nil {
		sig, err = base64.StdEncoding.DecodeString(sigB64)
		if err != nil {
			return nil, fmt.Errorf("error al decodificar firma de licencia: %w", err)
		}
	}

	// Verificamos la firma si tenemos la clave pública
	if s.publicKey != nil {
		if !ed25519.Verify(s.publicKey, payloadJSON, sig) {
			return nil, fmt.Errorf("firma de licencia inválida")
		}
	}

	// Parseamos el payload JSON
	var payload licensePayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de licencia: %w", err)
	}

	return &payload, nil
}

// loadPublicKey carga la clave pública Ed25519 desde base64.
func loadPublicKey(b64Key string) (ed25519.PublicKey, error) {
	// La clave puede estar en formato DER (SubjectPublicKeyInfo) o raw de 32 bytes
	keyBytes, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar clave pública: %w", err)
	}

	// Si tiene la longitud exacta de Ed25519 (32 bytes), la usamos directamente
	if len(keyBytes) == ed25519.PublicKeySize {
		return ed25519.PublicKey(keyBytes), nil
	}

	// Si está en formato DER (SubjectPublicKeyInfo), extraemos los últimos 32 bytes
	if len(keyBytes) > ed25519.PublicKeySize {
		return ed25519.PublicKey(keyBytes[len(keyBytes)-ed25519.PublicKeySize:]), nil
	}

	return nil, fmt.Errorf("tamaño de clave pública inválido: %d bytes", len(keyBytes))
}
