// Package sqlite contiene las pruebas unitarias para el repositorio SQLite.
package sqlite

import (
"testing"
"time"

"backup-smart/backend/domain/entities"
)

// setupTestRepo crea un repositorio en memoria para las pruebas.
func setupTestRepo(t *testing.T) *Repository {
t.Helper()
repo, err := NewRepository(":memory:")
if err != nil {
t.Fatalf("Error al crear repositorio de prueba: %v", err)
}
if err := repo.Migrate(); err != nil {
t.Fatalf("Error al migrar esquema de prueba: %v", err)
}
t.Cleanup(func() { repo.Close() })
return repo
}

// TestMigrateIsIdempotent verifica que las migraciones pueden ejecutarse múltiples veces sin error.
func TestMigrateIsIdempotent(t *testing.T) {
repo := setupTestRepo(t)
// Ejecutamos la migración nuevamente para verificar idempotencia
if err := repo.Migrate(); err != nil {
t.Errorf("La segunda migración no debe fallar: %v", err)
}
}

// TestDatabaseConfigCRUD verifica las operaciones CRUD en configuraciones de base de datos.
func TestDatabaseConfigCRUD(t *testing.T) {
repo := setupTestRepo(t)

// Creamos una nueva configuración
config := entities.DatabaseConfig{
Name:              "BD de Producción",
Host:              "localhost",
Port:              3306,
User:              "admin",
EncryptedPassword: "encrypted_pass_123",
DBName:            "production_db",
Type:              entities.DBTypeMySQL,
}

// CREATE
if err := repo.CreateDatabaseConfig(config); err != nil {
t.Fatalf("Error al crear database_config: %v", err)
}

// READ ALL
configs, err := repo.GetAllDatabaseConfigs()
if err != nil {
t.Fatalf("Error al obtener database_configs: %v", err)
}
if len(configs) != 1 {
t.Fatalf("Se esperaba 1 config, se obtuvo %d", len(configs))
}

saved := configs[0]
if saved.Name != config.Name {
t.Errorf("Nombre incorrecto: esperado=%q, obtenido=%q", config.Name, saved.Name)
}
if saved.Type != entities.DBTypeMySQL {
t.Errorf("Tipo incorrecto: esperado=%q, obtenido=%q", entities.DBTypeMySQL, saved.Type)
}

// READ BY ID
byID, err := repo.GetDatabaseConfigByID(saved.ID)
if err != nil {
t.Fatalf("Error al obtener database_config por ID: %v", err)
}
if byID.DBName != config.DBName {
t.Errorf("DBName incorrecto: esperado=%q, obtenido=%q", config.DBName, byID.DBName)
}

// UPDATE
saved.Host = "db.production.com"
saved.Port = 3307
if err := repo.UpdateDatabaseConfig(saved); err != nil {
t.Fatalf("Error al actualizar database_config: %v", err)
}
updated, _ := repo.GetDatabaseConfigByID(saved.ID)
if updated.Host != "db.production.com" {
t.Errorf("Host no actualizado correctamente: esperado=%q, obtenido=%q", "db.production.com", updated.Host)
}

// DELETE
if err := repo.DeleteDatabaseConfig(saved.ID); err != nil {
t.Fatalf("Error al eliminar database_config: %v", err)
}
configs, _ = repo.GetAllDatabaseConfigs()
if len(configs) != 0 {
t.Errorf("Se esperaban 0 configs después de eliminar, se obtuvo %d", len(configs))
}
}

// TestJobCRUD verifica las operaciones CRUD en jobs.
func TestJobCRUD(t *testing.T) {
repo := setupTestRepo(t)

// Primero creamos una configuración de BD (necesaria por FK)
dbConfig := entities.DatabaseConfig{
Name: "BD Test", Host: "localhost", Port: 5432,
User: "user", EncryptedPassword: "pass", DBName: "testdb", Type: entities.DBTypePostgres,
}
if err := repo.CreateDatabaseConfig(dbConfig); err != nil {
t.Fatalf("Error al crear dbConfig para test de job: %v", err)
}
configs, _ := repo.GetAllDatabaseConfigs()
dbConfigID := configs[0].ID

// Creamos un job
job := entities.Job{
Name:           "Respaldo Nocturno",
CronExpression: "0 0 2 * * *",
DBConfigID:     dbConfigID,
DestinationIDs: []int64{},
IsActive:       true,
}

// CREATE
if err := repo.CreateJob(job); err != nil {
t.Fatalf("Error al crear job: %v", err)
}

// READ ALL
jobs, err := repo.GetAllJobs()
if err != nil {
t.Fatalf("Error al obtener jobs: %v", err)
}
if len(jobs) != 1 {
t.Fatalf("Se esperaba 1 job, se obtuvo %d", len(jobs))
}

savedJob := jobs[0]
if savedJob.Name != job.Name {
t.Errorf("Nombre del job incorrecto: esperado=%q, obtenido=%q", job.Name, savedJob.Name)
}
if !savedJob.IsActive {
t.Error("El job debería estar activo")
}

// UPDATE STATUS
now := time.Now()
if err := repo.UpdateJobStatus(savedJob.ID, "success", now); err != nil {
t.Fatalf("Error al actualizar estado del job: %v", err)
}
updatedJob, _ := repo.GetJobByID(savedJob.ID)
if updatedJob.LastStatus != "success" {
t.Errorf("Estado del job incorrecto: esperado=%q, obtenido=%q", "success", updatedJob.LastStatus)
}

// DELETE
if err := repo.DeleteJob(savedJob.ID); err != nil {
t.Fatalf("Error al eliminar job: %v", err)
}
}

// TestMasterPasswordValidation verifica la gestión de la contraseña maestra.
func TestMasterPasswordValidation(t *testing.T) {
repo := setupTestRepo(t)

// Primera validación: no hay contraseña configurada, debe guardarla
ok, err := repo.ValidateMasterPassword("mi-contraseña-inicial")
if err != nil {
t.Fatalf("Error en primera validación: %v", err)
}
if !ok {
t.Error("La primera validación debe ser exitosa (configura la contraseña)")
}

// Segunda validación: contraseña correcta
ok, err = repo.ValidateMasterPassword("mi-contraseña-inicial")
if err != nil {
t.Fatalf("Error en segunda validación: %v", err)
}
if !ok {
t.Error("La contraseña correcta debe validarse exitosamente")
}

// Tercera validación: contraseña incorrecta
ok, err = repo.ValidateMasterPassword("contraseña-incorrecta")
if err != nil {
t.Fatalf("Error en tercera validación: %v", err)
}
if ok {
t.Error("Una contraseña incorrecta no debe validarse")
}
}

// TestActivityLogs verifica la creación y consulta de logs de actividad.
func TestActivityLogs(t *testing.T) {
repo := setupTestRepo(t)

// Insertamos varios logs
entries := []entities.LogEntry{
{JobID: 1, JobName: "Job A", Level: "info", Message: "Iniciando respaldo"},
{JobID: 1, JobName: "Job A", Level: "success", Message: "Respaldo completado"},
{JobID: 0, JobName: "", Level: "warn", Message: "Sistema iniciado"},
}

for _, entry := range entries {
if err := repo.CreateLog(entry); err != nil {
t.Fatalf("Error al crear log: %v", err)
}
}

// Consultamos los últimos logs
logs, err := repo.GetRecentLogs(10)
if err != nil {
t.Fatalf("Error al obtener logs: %v", err)
}
if len(logs) != 3 {
t.Errorf("Se esperaban 3 logs, se obtuvieron %d", len(logs))
}

// Verificamos el límite
logsLimited, err := repo.GetRecentLogs(2)
if err != nil {
t.Fatalf("Error al obtener logs limitados: %v", err)
}
if len(logsLimited) != 2 {
t.Errorf("Se esperaban 2 logs con límite=2, se obtuvieron %d", len(logsLimited))
}
}
