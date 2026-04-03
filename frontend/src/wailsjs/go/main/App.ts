// Stubs de los bindings generados por Wails para el módulo App de Go.
// En desarrollo local (sin Wails runtime), estas funciones simulan las llamadas al backend.
// Wails reemplaza este archivo con bindings reales al compilar la aplicación de escritorio.

import type { Job, DatabaseConfig, Destination, License, LogEntry } from '../../../app/models';

// Referencia al runtime de Wails inyectado por el proceso Go
declare const window: Window & {
  go?: {
    main?: {
      App?: Record<string, (...args: unknown[]) => Promise<unknown>>;
    };
  };
};

/**
 * Invoca un método del backend Go a través del bridge Wails.
 * Si el runtime no está disponible (modo web de desarrollo), lanza un error descriptivo.
 */
function invoke<T>(method: string, ...args: unknown[]): Promise<T> {
  const goApp = window?.go?.main?.App;
  if (!goApp || typeof goApp[method] !== 'function') {
    return Promise.reject(
      new Error(`Wails runtime no disponible. Método: ${method}. ¿Está corriendo dentro de la aplicación de escritorio?`)
    );
  }
  return goApp[method](...args) as Promise<T>;
}

// ─── Configuración e inicio ─────────────────────────────────────────────────

/** Inicializa la aplicación con la contraseña maestra y la clave de licencia. */
export function InitSetup(masterPassword: string, licenseKey: string): Promise<void> {
  return invoke<void>('InitSetup', masterPassword, licenseKey);
}

/** Valida la contraseña maestra y abre la sesión. */
export function Login(masterPassword: string): Promise<boolean> {
  return invoke<boolean>('Login', masterPassword);
}

// ─── Jobs ───────────────────────────────────────────────────────────────────

/** Retorna todos los jobs de respaldo configurados. */
export function GetJobs(): Promise<Job[]> {
  return invoke<Job[]>('GetJobs');
}

/** Crea o actualiza un job de respaldo. Si id=0 crea uno nuevo. */
export function SaveJob(job: Job): Promise<void> {
  return invoke<void>('SaveJob', job);
}

/** Elimina un job por su ID. */
export function DeleteJob(id: number): Promise<void> {
  return invoke<void>('DeleteJob', id);
}

/** Ejecuta un job de forma manual sin esperar el cron. */
export function RunJobManual(jobID: number): Promise<void> {
  return invoke<void>('RunJobManual', jobID);
}

// ─── Configuraciones de base de datos ──────────────────────────────────────

/** Retorna todas las configuraciones de bases de datos guardadas. */
export function GetDatabaseConfigs(): Promise<DatabaseConfig[]> {
  return invoke<DatabaseConfig[]>('GetDatabaseConfigs');
}

/** Guarda una configuración de base de datos. */
export function SaveDatabaseConfig(config: DatabaseConfig): Promise<void> {
  return invoke<void>('SaveDatabaseConfig', config);
}

/** Elimina una configuración de base de datos por su ID. */
export function DeleteDatabaseConfig(id: number): Promise<void> {
  return invoke<void>('DeleteDatabaseConfig', id);
}

// ─── Destinos ──────────────────────────────────────────────────────────────

/** Retorna todos los destinos de almacenamiento configurados. */
export function GetDestinations(): Promise<Destination[]> {
  return invoke<Destination[]>('GetDestinations');
}

/** Guarda un destino de almacenamiento. */
export function SaveDestination(dest: Destination): Promise<void> {
  return invoke<void>('SaveDestination', dest);
}

/** Elimina un destino por su ID. */
export function DeleteDestination(id: number): Promise<void> {
  return invoke<void>('DeleteDestination', id);
}

// ─── Licenciamiento ─────────────────────────────────────────────────────────

/** Retorna el ID de hardware único del equipo actual. */
export function GetHardwareID(): Promise<string> {
  return invoke<string>('GetHardwareID');
}

/** Valida una clave de licencia y retorna el estado detallado. */
export function VerifyLicense(licenseKey: string): Promise<License> {
  return invoke<License>('VerifyLicense', licenseKey);
}

// ─── Logs ────────────────────────────────────────────────────────────────────

/** Retorna los últimos N registros de log del sistema. */
export function GetLogs(limit: number): Promise<LogEntry[]> {
  return invoke<LogEntry[]>('GetLogs', limit);
}
