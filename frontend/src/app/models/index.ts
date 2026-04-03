// Modelos de dominio de backup-smart para el frontend Angular.
// Corresponden exactamente a las entidades Go del backend.

/** Tipos de motores de base de datos soportados. */
export type DBType = 'mysql' | 'postgres';

/** Tipos de destinos de almacenamiento soportados. */
export type DestinationType = 's3' | 'local' | 'ftp';

/** Planes de licencia disponibles. */
export type LicensePlan = 'free' | 'pro' | 'enterprise';

/** Estados posibles de un job en su última ejecución. */
export type JobStatus = 'success' | 'error' | 'running' | '';

/** Niveles de log de actividad. */
export type LogLevel = 'info' | 'warn' | 'error' | 'success';

/**
 * Configuración de conexión a una base de datos.
 * La contraseña se almacena encriptada con AES-256 en el backend.
 */
export interface DatabaseConfig {
  id: number;
  name: string;
  host: string;
  port: number;
  user: string;
  /** En el frontend, este campo contiene la contraseña en texto plano. El backend encripta. */
  encrypted_password: string;
  db_name: string;
  type: DBType;
  created_at?: string;
  updated_at?: string;
}

/**
 * Destino de almacenamiento para los respaldos.
 * Las credenciales se almacenan encriptadas con AES-256 en el backend.
 */
export interface Destination {
  id: number;
  name: string;
  type: DestinationType;
  path_or_bucket: string;
  /** JSON con las credenciales (ej: {access_key_id, secret_access_key} para S3). */
  encrypted_credentials: string;
  region?: string;
  created_at?: string;
  updated_at?: string;
}

/**
 * Job de respaldo automatizado programado con cron.
 */
export interface Job {
  id: number;
  name: string;
  /** Expresión cron en formato de 6 campos (con segundos): "s m h d M w" */
  cron_expression: string;
  db_config_id: number;
  destination_ids: number[];
  notify_webhook?: string;
  is_active: boolean;
  last_run_at?: string;
  last_status?: JobStatus;
  created_at?: string;
  updated_at?: string;
}

/**
 * Estado de la licencia del software.
 */
export interface License {
  machine_id: string;
  license_key: string;
  signed_token?: string;
  is_valid: boolean;
  last_checked: string;
  expiration_date?: string;
  grace_period_end?: string;
  plan: LicensePlan;
}

/**
 * Entrada de log de actividad del sistema.
 */
export interface LogEntry {
  id: number;
  job_id?: number;
  job_name?: string;
  level: LogLevel;
  message: string;
  created_at: string;
}

/** Opciones de tipo de base de datos para selectores del formulario. */
export const DB_TYPE_OPTIONS: { value: DBType; label: string }[] = [
  { value: 'mysql', label: 'MySQL / MariaDB' },
  { value: 'postgres', label: 'PostgreSQL' },
];

/** Opciones de tipo de destino para selectores del formulario. */
export const DESTINATION_TYPE_OPTIONS: { value: DestinationType; label: string }[] = [
  { value: 's3', label: 'Amazon S3' },
  { value: 'local', label: 'Directorio Local' },
  { value: 'ftp', label: 'Servidor FTP' },
];

/** Puerto por defecto según el tipo de base de datos. */
export const DEFAULT_PORTS: Record<DBType, number> = {
  mysql: 3306,
  postgres: 5432,
};
