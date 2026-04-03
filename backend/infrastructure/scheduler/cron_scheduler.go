// Package scheduler implementa el programador de tareas cron de backup-smart.
// Utiliza robfig/cron/v3 para la programación de jobs en segundo plano.
package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"backup-smart/backend/domain/entities"
	"backup-smart/backend/infrastructure/dumpers"
	"backup-smart/backend/infrastructure/security"
	"backup-smart/backend/infrastructure/sqlite"
	"backup-smart/backend/infrastructure/storage"

	"github.com/robfig/cron/v3"
)

// CronScheduler gestiona la ejecución programada de jobs de respaldo.
// Utiliza robfig/cron/v3 con soporte para 5 o 6 campos de expresión cron.
type CronScheduler struct {
	ctx       context.Context
	cron      *cron.Cron
	repo      *sqlite.Repository
	secSvc    *security.AES256Service
	masterKey string
	jobMap    map[int64]cron.EntryID // Mapeo de jobID -> entryID de cron
}

// NewCronScheduler crea una nueva instancia del scheduler.
func NewCronScheduler(ctx context.Context, repo *sqlite.Repository, secSvc *security.AES256Service, masterKey string) *CronScheduler {
	return &CronScheduler{
		ctx:       ctx,
		cron:      cron.New(cron.WithSeconds()),
		repo:      repo,
		secSvc:    secSvc,
		masterKey: masterKey,
		jobMap:    make(map[int64]cron.EntryID),
	}
}

// Start inicia el scheduler y carga todos los jobs activos desde la base de datos.
func (s *CronScheduler) Start() {
	if err := s.loadJobs(); err != nil {
		s.logEntry(0, "", "error", fmt.Sprintf("Error al cargar jobs del scheduler: %v", err))
	}
	s.cron.Start()
}

// Stop detiene el scheduler de forma ordenada, esperando a que termine el job en ejecución.
func (s *CronScheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

// Reload recarga todos los jobs desde la base de datos, deteniendo y reiniciando el scheduler.
func (s *CronScheduler) Reload() error {
	// Detenemos el cron actual
	ctx := s.cron.Stop()
	<-ctx.Done()

	// Creamos un nuevo cron y recargamos los jobs
	s.cron = cron.New(cron.WithSeconds())
	s.jobMap = make(map[int64]cron.EntryID)
	if err := s.loadJobs(); err != nil {
		return fmt.Errorf("error al recargar jobs: %w", err)
	}
	s.cron.Start()
	return nil
}

// RunJobNow ejecuta un job de forma inmediata de forma asíncrona.
func (s *CronScheduler) RunJobNow(jobID int64) error {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil {
		return fmt.Errorf("error al obtener job para ejecución manual: %w", err)
	}
	go s.executeJob(job)
	return nil
}

// loadJobs carga todos los jobs activos y los registra en el cron.
func (s *CronScheduler) loadJobs() error {
	jobs, err := s.repo.GetAllJobs()
	if err != nil {
		return fmt.Errorf("error al obtener jobs: %w", err)
	}

	for _, job := range jobs {
		if !job.IsActive {
			continue
		}
		// Capturamos la variable para el closure
		jobCopy := job
		entryID, err := s.cron.AddFunc(jobCopy.CronExpression, func() {
			s.executeJob(jobCopy)
		})
		if err != nil {
			s.logEntry(jobCopy.ID, jobCopy.Name, "error",
				fmt.Sprintf("Expresión cron inválida '%s': %v", jobCopy.CronExpression, err))
			continue
		}
		s.jobMap[jobCopy.ID] = entryID
	}
	return nil
}

// executeJob ejecuta un job completo de respaldo: dump → upload → notificación.
func (s *CronScheduler) executeJob(job entities.Job) {
	startTime := time.Now()
	s.logEntry(job.ID, job.Name, "info", fmt.Sprintf("Iniciando job de respaldo: %s", job.Name))

	// Actualizamos el estado a "running"
	if err := s.repo.UpdateJobStatus(job.ID, "running", startTime); err != nil {
		s.logEntry(job.ID, job.Name, "warn", fmt.Sprintf("No se pudo actualizar estado a running: %v", err))
	}

	// Obtenemos la configuración de la base de datos
	dbConfig, err := s.repo.GetDatabaseConfigByID(job.DBConfigID)
	if err != nil {
		s.finishJob(job, "error", fmt.Sprintf("Error al obtener config de DB: %v", err))
		return
	}

	// Desencriptamos la contraseña
	plainPassword, err := s.secSvc.Decrypt(dbConfig.EncryptedPassword, s.masterKey)
	if err != nil {
		s.finishJob(job, "error", fmt.Sprintf("Error al desencriptar contraseña: %v", err))
		return
	}
	dbConfig.EncryptedPassword = plainPassword

	// Determinamos el volcador según el tipo de base de datos
	outputDir := "backups"
	var dumpFilePath string

	switch dbConfig.Type {
	case entities.DBTypeMySQL:
		dumper := dumpers.NewMySQLDumper(outputDir)
		dumpFilePath, err = dumper.Dump(s.ctx, dbConfig)
	case entities.DBTypePostgres:
		dumper := dumpers.NewPostgresDumper(outputDir)
		dumpFilePath, err = dumper.Dump(s.ctx, dbConfig)
	default:
		s.finishJob(job, "error", fmt.Sprintf("Tipo de base de datos no soportado: %s", dbConfig.Type))
		return
	}

	if err != nil {
		s.finishJob(job, "error", fmt.Sprintf("Error al generar volcado: %v", err))
		return
	}

	s.logEntry(job.ID, job.Name, "info", fmt.Sprintf("Volcado generado: %s", dumpFilePath))

	// Subimos el volcado a todos los destinos configurados
	uploadErrors := 0
	for _, destID := range job.DestinationIDs {
		dest, err := s.repo.GetDestinationByID(destID)
		if err != nil {
			s.logEntry(job.ID, job.Name, "warn", fmt.Sprintf("Destino no encontrado (id=%d): %v", destID, err))
			uploadErrors++
			continue
		}

		// Desencriptamos las credenciales del destino
		if dest.EncryptedCredentials != "" {
			plainCreds, err := s.secSvc.Decrypt(dest.EncryptedCredentials, s.masterKey)
			if err == nil {
				dest.EncryptedCredentials = plainCreds
			}
		}

		// Seleccionamos el uploader según el tipo de destino
		var uploadErr error
		switch dest.Type {
		case entities.DestinationS3:
			uploader := storage.NewS3Uploader()
			uploadErr = uploader.Upload(s.ctx, dumpFilePath, dest)
		case entities.DestinationLocal:
			uploader := storage.NewLocalUploader()
			uploadErr = uploader.Upload(s.ctx, dumpFilePath, dest)
		default:
			uploadErr = fmt.Errorf("tipo de destino no soportado: %s", dest.Type)
		}

		if uploadErr != nil {
			s.logEntry(job.ID, job.Name, "error", fmt.Sprintf("Error al subir a destino '%s': %v", dest.Name, uploadErr))
			uploadErrors++
		} else {
			s.logEntry(job.ID, job.Name, "info", fmt.Sprintf("Volcado subido exitosamente a: %s", dest.Name))
		}
	}

	// Eliminamos el archivo temporal local después de subir
	if dumpFilePath != "" {
		os.Remove(dumpFilePath)
	}

	// Determinamos el estado final
	finalStatus := "success"
	finalMsg := fmt.Sprintf("Job completado en %v", time.Since(startTime).Round(time.Second))
	if uploadErrors > 0 {
		finalStatus = "error"
		finalMsg = fmt.Sprintf("Job completado con %d error(es) de subida", uploadErrors)
	}

	// Enviamos notificación si hay webhook configurado
	if job.NotifyWebhook != "" {
		if err := sendWebhookNotification(s.ctx, job.Name, finalMsg, job.NotifyWebhook); err != nil {
			s.logEntry(job.ID, job.Name, "warn", fmt.Sprintf("Error al enviar notificación: %v", err))
		}
	}

	s.finishJob(job, finalStatus, finalMsg)
}

// finishJob actualiza el estado final del job y registra el resultado en el log.
func (s *CronScheduler) finishJob(job entities.Job, status string, msg string) {
	if err := s.repo.UpdateJobStatus(job.ID, status, time.Now()); err != nil {
		s.logEntry(job.ID, job.Name, "warn", fmt.Sprintf("No se pudo actualizar estado final: %v", err))
	}
	s.logEntry(job.ID, job.Name, status, msg)
}

// logEntry registra un evento en la base de datos de logs.
func (s *CronScheduler) logEntry(jobID int64, jobName string, level string, message string) {
	_ = s.repo.CreateLog(entities.LogEntry{
		JobID:   jobID,
		JobName: jobName,
		Level:   level,
		Message: message,
	})
}

// sendWebhookNotification envía una notificación HTTP POST al webhook configurado.
func sendWebhookNotification(ctx context.Context, jobName string, message string, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}
	// Construimos el payload JSON de la notificación
	payload := fmt.Sprintf(`{"text":"[backup-smart] Job: %s - %s"}`, jobName, message)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("error al crear request de notificación: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error al enviar notificación al webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("el webhook retornó código de error: %d", resp.StatusCode)
	}
	return nil
}
