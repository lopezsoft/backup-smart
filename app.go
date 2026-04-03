// Package main contiene la capa de aplicación de backup-smart.
// El struct App expone los métodos de negocio a Angular mediante Wails bindings.
package main

import (
	"context"
	"fmt"

	"backup-smart/backend/domain/entities"
	"backup-smart/backend/infrastructure/licensing"
	"backup-smart/backend/infrastructure/scheduler"
	"backup-smart/backend/infrastructure/security"
	"backup-smart/backend/infrastructure/sqlite"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App es el struct principal que expone los métodos a Angular mediante Wails.
type App struct {
	ctx         context.Context
	repo        *sqlite.Repository
	secSvc      *security.AES256Service
	scheduler   *scheduler.CronScheduler
	licenseSvc  *licensing.LicensingService
	masterKey   string
	initialized bool
}

// NewApp crea e inicializa una nueva instancia de la aplicación.
func NewApp() *App {
	return &App{}
}

// startup se ejecuta al arrancar la aplicación Wails; recibe el contexto del runtime.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.secSvc = security.NewAES256Service()
	a.licenseSvc = licensing.NewLicensingService()
	runtime.LogInfo(ctx, "backup-smart iniciado correctamente")
}

// ─── Configuración e inicio ────────────────────────────────────────────────────

// InitSetup inicializa la aplicación con la contraseña maestra y la clave de licencia.
// Crea las tablas en SQLite y valida la licencia proporcionada.
func (a *App) InitSetup(masterPassword string, licenseKey string) error {
	if masterPassword == "" {
		return fmt.Errorf("la contraseña maestra no puede estar vacía")
	}

	// Inicializamos el repositorio SQLite
	repo, err := sqlite.NewRepository("backup-smart.db")
	if err != nil {
		return fmt.Errorf("error al inicializar la base de datos: %w", err)
	}

	// Ejecutamos las migraciones del esquema
	if err := repo.Migrate(); err != nil {
		return fmt.Errorf("error al ejecutar migraciones: %w", err)
	}

	a.repo = repo
	a.masterKey = masterPassword

	// Validamos la licencia si se proporcionó
	if licenseKey != "" {
		if _, err := a.licenseSvc.VerifyLicense(licenseKey); err != nil {
			return fmt.Errorf("licencia inválida: %w", err)
		}
	}

	// Inicializamos el scheduler de cron
	a.scheduler = scheduler.NewCronScheduler(a.ctx, repo, a.secSvc, a.masterKey)
	a.scheduler.Start()

	a.initialized = true
	return nil
}

// Login valida la contraseña maestra comparando con el hash almacenado.
func (a *App) Login(masterPassword string) bool {
	if a.repo == nil {
		return false
	}
	ok, err := a.repo.ValidateMasterPassword(masterPassword)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Error al validar contraseña: %v", err))
		return false
	}
	if ok {
		a.masterKey = masterPassword
	}
	return ok
}

// ─── Gestión de Jobs ───────────────────────────────────────────────────────────

// GetJobs retorna la lista de todos los jobs de respaldo configurados.
func (a *App) GetJobs() ([]entities.Job, error) {
	if err := a.checkInit(); err != nil {
		return nil, err
	}
	jobs, err := a.repo.GetAllJobs()
	if err != nil {
		return nil, fmt.Errorf("error al obtener jobs: %w", err)
	}
	return jobs, nil
}

// SaveJob crea o actualiza un job de respaldo. Si el ID es 0, crea uno nuevo.
func (a *App) SaveJob(job entities.Job) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	if job.Name == "" {
		return fmt.Errorf("el nombre del job no puede estar vacío")
	}
	if job.CronExpression == "" {
		return fmt.Errorf("la expresión cron no puede estar vacía")
	}

	var err error
	if job.ID == 0 {
		err = a.repo.CreateJob(job)
	} else {
		err = a.repo.UpdateJob(job)
	}
	if err != nil {
		return fmt.Errorf("error al guardar el job: %w", err)
	}

	// Recargamos el scheduler con los jobs actualizados
	if err := a.scheduler.Reload(); err != nil {
		return fmt.Errorf("error al recargar el scheduler: %w", err)
	}
	return nil
}

// DeleteJob elimina un job de respaldo por su ID.
func (a *App) DeleteJob(id int64) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	if err := a.repo.DeleteJob(id); err != nil {
		return fmt.Errorf("error al eliminar el job: %w", err)
	}
	return a.scheduler.Reload()
}

// RunJobManual ejecuta un job de forma manual sin esperar el cron.
func (a *App) RunJobManual(jobID int64) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	return a.scheduler.RunJobNow(jobID)
}

// ─── Gestión de configuraciones de base de datos ──────────────────────────────

// GetDatabaseConfigs retorna todas las configuraciones de bases de datos guardadas.
func (a *App) GetDatabaseConfigs() ([]entities.DatabaseConfig, error) {
	if err := a.checkInit(); err != nil {
		return nil, err
	}
	configs, err := a.repo.GetAllDatabaseConfigs()
	if err != nil {
		return nil, fmt.Errorf("error al obtener configuraciones: %w", err)
	}
	// Desencriptamos las contraseñas para el frontend
	for i := range configs {
		plain, err := a.secSvc.Decrypt(configs[i].EncryptedPassword, a.masterKey)
		if err == nil {
			configs[i].EncryptedPassword = plain
		}
	}
	return configs, nil
}

// SaveDatabaseConfig guarda una configuración de base de datos (encriptando la contraseña).
func (a *App) SaveDatabaseConfig(config entities.DatabaseConfig) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	if config.Host == "" || config.DBName == "" || config.User == "" {
		return fmt.Errorf("host, usuario y nombre de base de datos son requeridos")
	}
	// Encriptamos la contraseña antes de guardar
	encrypted, err := a.secSvc.Encrypt(config.EncryptedPassword, a.masterKey)
	if err != nil {
		return fmt.Errorf("error al encriptar contraseña: %w", err)
	}
	config.EncryptedPassword = encrypted

	if config.ID == 0 {
		return a.repo.CreateDatabaseConfig(config)
	}
	return a.repo.UpdateDatabaseConfig(config)
}

// DeleteDatabaseConfig elimina una configuración de base de datos por su ID.
func (a *App) DeleteDatabaseConfig(id int64) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	return a.repo.DeleteDatabaseConfig(id)
}

// ─── Gestión de destinos ──────────────────────────────────────────────────────

// GetDestinations retorna todos los destinos de almacenamiento configurados.
func (a *App) GetDestinations() ([]entities.Destination, error) {
	if err := a.checkInit(); err != nil {
		return nil, err
	}
	dests, err := a.repo.GetAllDestinations()
	if err != nil {
		return nil, fmt.Errorf("error al obtener destinos: %w", err)
	}
	return dests, nil
}

// SaveDestination guarda un destino de almacenamiento (encriptando las credenciales).
func (a *App) SaveDestination(dest entities.Destination) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	if dest.PathOrBucket == "" {
		return fmt.Errorf("la ruta o bucket no puede estar vacía")
	}
	// Encriptamos las credenciales si las hay
	if dest.EncryptedCredentials != "" {
		encrypted, err := a.secSvc.Encrypt(dest.EncryptedCredentials, a.masterKey)
		if err != nil {
			return fmt.Errorf("error al encriptar credenciales: %w", err)
		}
		dest.EncryptedCredentials = encrypted
	}

	if dest.ID == 0 {
		return a.repo.CreateDestination(dest)
	}
	return a.repo.UpdateDestination(dest)
}

// DeleteDestination elimina un destino de almacenamiento por su ID.
func (a *App) DeleteDestination(id int64) error {
	if err := a.checkInit(); err != nil {
		return err
	}
	return a.repo.DeleteDestination(id)
}

// ─── Licenciamiento ───────────────────────────────────────────────────────────

// GetHardwareID retorna el identificador de hardware del equipo actual.
func (a *App) GetHardwareID() string {
	return a.licenseSvc.GetHardwareID()
}

// VerifyLicense valida una clave de licencia y retorna el estado de la misma.
func (a *App) VerifyLicense(licenseKey string) (entities.License, error) {
	return a.licenseSvc.VerifyLicense(licenseKey)
}

// ─── Logs ─────────────────────────────────────────────────────────────────────

// GetLogs retorna los últimos N registros de log del sistema.
func (a *App) GetLogs(limit int) ([]entities.LogEntry, error) {
	if err := a.checkInit(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	return a.repo.GetRecentLogs(limit)
}

// ─── Helpers privados ─────────────────────────────────────────────────────────

// checkInit verifica que la aplicación haya sido inicializada antes de operar.
func (a *App) checkInit() error {
	if !a.initialized || a.repo == nil {
		return fmt.Errorf("la aplicación no ha sido inicializada, por favor ejecute InitSetup primero")
	}
	return nil
}
