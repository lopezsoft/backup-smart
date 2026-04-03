// Servicio puente de Wails para Angular.
// Envuelve todas las llamadas al backend Go en un servicio Angular con manejo de errores.
import { Injectable, signal, computed } from '@angular/core';
import {
  InitSetup, Login, GetJobs, SaveJob, DeleteJob, RunJobManual,
  GetDatabaseConfigs, SaveDatabaseConfig, DeleteDatabaseConfig,
  GetDestinations, SaveDestination, DeleteDestination,
  GetHardwareID, VerifyLicense, GetLogs
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import type { Job, DatabaseConfig, Destination, License, LogEntry } from '../models';

/**
 * WailsBridgeService es el servicio central que actúa como puente entre
 * Angular y el backend Go a través de Wails IPC.
 * Gestiona el estado de la aplicación usando Angular Signals.
 */
@Injectable({
  providedIn: 'root'
})
export class WailsBridgeService {
  // ─── Estado de la aplicación con Signals ────────────────────────────────

  /** Indica si el usuario ha iniciado sesión con la contraseña maestra. */
  readonly isAuthenticated = signal<boolean>(false);

  /** Estado de carga global para mostrar indicadores de progreso. */
  readonly isLoading = signal<boolean>(false);

  /** Último mensaje de error para mostrar al usuario. */
  readonly errorMessage = signal<string>('');

  /** Lista de jobs de respaldo activos. */
  readonly jobs = signal<Job[]>([]);

  /** Lista de configuraciones de bases de datos. */
  readonly dbConfigs = signal<DatabaseConfig[]>([]);

  /** Lista de destinos de almacenamiento. */
  readonly destinations = signal<Destination[]>([]);

  /** Estado actual de la licencia. */
  readonly license = signal<License | null>(null);

  /** Logs de actividad en tiempo real. */
  readonly logs = signal<LogEntry[]>([]);

  /** ID de hardware del equipo actual. */
  readonly hardwareID = signal<string>('');

  /** Computa si la licencia Pro está activa y válida. */
  readonly isProLicensed = computed(() => {
    const lic = this.license();
    return lic !== null && lic.is_valid && lic.plan !== 'free';
  });

  /** Número total de jobs activos. */
  readonly activeJobsCount = computed(() =>
    this.jobs().filter(j => j.is_active).length
  );

  constructor() {
    // Suscribimos a eventos en tiempo real emitidos desde Go
    this.subscribeToRuntimeEvents();
  }

  // ─── Autenticación ─────────────────────────────────────────────────────

  /**
   * Inicializa la aplicación por primera vez con la contraseña maestra y licencia.
   * Debe llamarse al primer uso antes de Login.
   */
  async initSetup(masterPassword: string, licenseKey: string): Promise<boolean> {
    return this.withLoading(async () => {
      await InitSetup(masterPassword, licenseKey);
      this.isAuthenticated.set(true);
      await this.loadAllData();
      return true;
    });
  }

  /**
   * Inicia sesión con la contraseña maestra.
   * Retorna true si la autenticación fue exitosa.
   */
  async login(masterPassword: string): Promise<boolean> {
    return this.withLoading(async () => {
      const ok = await Login(masterPassword);
      if (ok) {
        this.isAuthenticated.set(true);
        await this.loadAllData();
      }
      return ok;
    });
  }

  /** Cierra la sesión del usuario limpiando el estado. */
  logout(): void {
    this.isAuthenticated.set(false);
    this.jobs.set([]);
    this.dbConfigs.set([]);
    this.destinations.set([]);
    this.license.set(null);
    this.logs.set([]);
  }

  // ─── Jobs ───────────────────────────────────────────────────────────────

  /** Carga todos los jobs desde el backend. */
  async loadJobs(): Promise<void> {
    return this.withLoading(async () => {
      const result = await GetJobs();
      this.jobs.set(result ?? []);
    });
  }

  /** Guarda un job (crea si id=0, actualiza si id>0). */
  async saveJob(job: Job): Promise<void> {
    return this.withLoading(async () => {
      await SaveJob(job);
      await this.loadJobs();
    });
  }

  /** Elimina un job por su ID. */
  async deleteJob(id: number): Promise<void> {
    return this.withLoading(async () => {
      await DeleteJob(id);
      await this.loadJobs();
    });
  }

  /** Ejecuta un job de forma manual inmediata. */
  async runJobNow(jobID: number): Promise<void> {
    return this.withLoading(async () => {
      await RunJobManual(jobID);
    });
  }

  // ─── Configuraciones de base de datos ──────────────────────────────────

  /** Carga todas las configuraciones de bases de datos. */
  async loadDatabaseConfigs(): Promise<void> {
    return this.withLoading(async () => {
      const result = await GetDatabaseConfigs();
      this.dbConfigs.set(result ?? []);
    });
  }

  /** Guarda una configuración de base de datos. */
  async saveDatabaseConfig(config: DatabaseConfig): Promise<void> {
    return this.withLoading(async () => {
      await SaveDatabaseConfig(config);
      await this.loadDatabaseConfigs();
    });
  }

  /** Elimina una configuración de base de datos. */
  async deleteDatabaseConfig(id: number): Promise<void> {
    return this.withLoading(async () => {
      await DeleteDatabaseConfig(id);
      await this.loadDatabaseConfigs();
    });
  }

  // ─── Destinos ──────────────────────────────────────────────────────────

  /** Carga todos los destinos de almacenamiento. */
  async loadDestinations(): Promise<void> {
    return this.withLoading(async () => {
      const result = await GetDestinations();
      this.destinations.set(result ?? []);
    });
  }

  /** Guarda un destino de almacenamiento. */
  async saveDestination(dest: Destination): Promise<void> {
    return this.withLoading(async () => {
      await SaveDestination(dest);
      await this.loadDestinations();
    });
  }

  /** Elimina un destino. */
  async deleteDestination(id: number): Promise<void> {
    return this.withLoading(async () => {
      await DeleteDestination(id);
      await this.loadDestinations();
    });
  }

  // ─── Licenciamiento ─────────────────────────────────────────────────────

  /** Carga el ID de hardware del equipo. */
  async loadHardwareID(): Promise<void> {
    const id = await GetHardwareID();
    this.hardwareID.set(id);
  }

  /** Verifica y carga el estado de la licencia. */
  async verifyLicense(licenseKey: string): Promise<License> {
    return this.withLoading(async () => {
      const lic = await VerifyLicense(licenseKey);
      this.license.set(lic);
      return lic;
    });
  }

  // ─── Logs ────────────────────────────────────────────────────────────────

  /** Carga los últimos N logs de actividad. */
  async loadLogs(limit = 100): Promise<void> {
    const result = await GetLogs(limit);
    this.logs.set(result ?? []);
  }

  // ─── Helpers privados ───────────────────────────────────────────────────

  /**
   * Carga todos los datos necesarios para el dashboard después del login.
   */
  private async loadAllData(): Promise<void> {
    await Promise.all([
      this.loadJobs(),
      this.loadDatabaseConfigs(),
      this.loadDestinations(),
      this.loadLogs(),
      this.loadHardwareID(),
    ]);
  }

  /**
   * Envuelve una operación asíncrona con manejo de estado de carga y errores.
   */
  private async withLoading<T>(fn: () => Promise<T>): Promise<T> {
    this.isLoading.set(true);
    this.errorMessage.set('');
    try {
      return await fn();
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      this.errorMessage.set(message);
      throw err;
    } finally {
      this.isLoading.set(false);
    }
  }

  /**
   * Suscribe a los eventos de tiempo real emitidos desde el backend Go.
   * Estos eventos se emiten cuando un job termina de ejecutarse.
   */
  private subscribeToRuntimeEvents(): void {
    // Evento emitido cuando un job actualiza su estado
    EventsOn('job:status', (data: unknown) => {
      const entry = data as LogEntry;
      if (entry) {
        this.logs.update(current => [entry, ...current].slice(0, 500));
      }
    });

    // Evento emitido cuando hay una nueva entrada de log
    EventsOn('log:entry', (data: unknown) => {
      const entry = data as LogEntry;
      if (entry) {
        this.logs.update(current => [entry, ...current].slice(0, 500));
      }
    });
  }
}
