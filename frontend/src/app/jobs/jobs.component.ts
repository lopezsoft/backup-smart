// Componente de gestión de Jobs de respaldo de backup-smart.
// Permite crear, editar, eliminar y ejecutar manualmente los jobs programados.
import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WailsBridgeService } from '../services/wails-bridge.service';
import { NavComponent } from '../shared/nav.component';
import type { Job, DatabaseConfig, Destination } from '../models';

/**
 * JobsComponent permite al usuario gestionar los jobs de respaldo.
 * Incluye formulario de creación/edición con soporte para expresiones cron.
 */
@Component({
  selector: 'app-jobs',
  standalone: true,
  imports: [CommonModule, DatePipe, FormsModule, NavComponent],
  template: `
    <div class="flex h-screen bg-terminal-bg overflow-hidden">
      <app-nav />

      <main class="flex-1 overflow-y-auto p-6">
        <!-- Header -->
        <div class="mb-6 flex items-center justify-between">
          <div>
            <h1 class="text-terminal-text text-2xl font-bold">Jobs de Respaldo</h1>
            <p class="text-terminal-muted text-sm mt-1">
              Configura y programa tus tareas de respaldo automatizado
            </p>
          </div>
          <button class="btn-primary flex items-center gap-2" (click)="openForm()">
            <span>+</span> Nuevo Job
          </button>
        </div>

        <!-- Error global -->
        @if (bridge.errorMessage()) {
          <div class="mb-4 p-3 rounded border border-terminal-red/50 bg-red-900/20 text-terminal-red text-sm">
            ⚠ {{ bridge.errorMessage() }}
          </div>
        }

        <!-- Modal de formulario de job -->
        @if (showForm()) {
          <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div class="card w-full max-w-2xl max-h-[90vh] overflow-y-auto border-terminal-blue/30">
              <h2 class="text-terminal-text text-lg font-semibold mb-6">
                {{ editingJob()?.id ? 'Editar Job' : 'Crear Nuevo Job' }}
              </h2>

              <form (ngSubmit)="saveJob()" #jobForm="ngForm">
                <div class="grid grid-cols-1 gap-4">
                  <!-- Nombre del job -->
                  <div>
                    <label class="form-label">Nombre del Job *</label>
                    <input type="text" class="form-input" name="name"
                      [(ngModel)]="formData.name" required placeholder="ej: Respaldo Noche BD Producción" />
                  </div>

                  <!-- Expresión Cron -->
                  <div>
                    <label class="form-label">Expresión Cron * (6 campos: s m h d M w)</label>
                    <input type="text" class="form-input font-mono" name="cron"
                      [(ngModel)]="formData.cron_expression" required
                      placeholder="ej: 0 0 2 * * * (todos los días a las 2am)" />
                    <p class="text-terminal-muted text-xs mt-1">
                      Formato: segundos minutos horas día mes día-semana
                    </p>
                    <!-- Presets de cron -->
                    <div class="mt-2 flex flex-wrap gap-1">
                      @for (preset of cronPresets; track preset.expr) {
                        <button type="button"
                          class="text-xs px-2 py-0.5 rounded bg-terminal-border hover:bg-terminal-blue/20 text-terminal-muted hover:text-terminal-blue transition-colors"
                          (click)="applyCronPreset(preset.expr)">
                          {{ preset.label }}
                        </button>
                      }
                    </div>
                  </div>

                  <!-- Base de datos -->
                  <div>
                    <label class="form-label">Base de Datos *</label>
                    <select class="form-input" name="dbConfig"
                      [(ngModel)]="formData.db_config_id" required>
                      <option [value]="0" disabled>Selecciona una base de datos...</option>
                      @for (db of bridge.dbConfigs(); track db.id) {
                        <option [value]="db.id">{{ db.name }} ({{ db.type }} - {{ db.host }})</option>
                      }
                    </select>
                    @if (bridge.dbConfigs().length === 0) {
                      <p class="text-terminal-yellow text-xs mt-1">
                        ⚠ No hay bases de datos configuradas. Crea una desde el panel de jobs.
                      </p>
                    }
                  </div>

                  <!-- Destinos múltiples -->
                  <div>
                    <label class="form-label">Destinos de Almacenamiento</label>
                    <div class="space-y-1 max-h-32 overflow-y-auto p-2 bg-terminal-bg rounded border border-terminal-border">
                      @if (bridge.destinations().length === 0) {
                        <p class="text-terminal-muted text-xs">No hay destinos configurados.</p>
                      }
                      @for (dest of bridge.destinations(); track dest.id) {
                        <label class="flex items-center gap-2 cursor-pointer">
                          <input type="checkbox"
                            class="accent-terminal-blue"
                            [checked]="isDestinationSelected(dest.id)"
                            (change)="toggleDestination(dest.id, $event)" />
                          <span class="text-terminal-text text-sm">
                            {{ dest.name }} <span class="text-terminal-muted">({{ dest.type }})</span>
                          </span>
                        </label>
                      }
                    </div>
                  </div>

                  <!-- Webhook de notificación -->
                  <div>
                    <label class="form-label">Webhook de Notificación (opcional)</label>
                    <input type="url" class="form-input" name="webhook"
                      [(ngModel)]="formData.notify_webhook"
                      placeholder="https://hooks.slack.com/services/..." />
                  </div>

                  <!-- Estado activo/inactivo -->
                  <div class="flex items-center gap-3">
                    <input type="checkbox" id="isActive" class="accent-terminal-green w-4 h-4"
                      name="isActive" [(ngModel)]="formData.is_active" />
                    <label for="isActive" class="text-terminal-text text-sm cursor-pointer">
                      Job activo (incluir en el scheduler)
                    </label>
                  </div>
                </div>

                <!-- Botones del formulario -->
                <div class="flex gap-3 mt-6 justify-end">
                  <button type="button" class="btn-secondary" (click)="closeForm()">
                    Cancelar
                  </button>
                  <button type="submit" class="btn-primary"
                    [disabled]="bridge.isLoading() || !formData.name || !formData.cron_expression || !formData.db_config_id">
                    @if (bridge.isLoading()) {
                      <span class="animate-spin">↻</span>
                    } @else {
                      {{ editingJob()?.id ? 'Actualizar' : 'Crear Job' }}
                    }
                  </button>
                </div>
              </form>
            </div>
          </div>
        }

        <!-- Tabla de jobs -->
        @if (bridge.jobs().length === 0) {
          <div class="card text-center py-12">
            <div class="text-6xl mb-4">🗄️</div>
            <h3 class="text-terminal-text text-lg font-semibold mb-2">No hay jobs configurados</h3>
            <p class="text-terminal-muted text-sm mb-4">
              Crea tu primer job de respaldo automatizado
            </p>
            <button class="btn-primary" (click)="openForm()">+ Crear Job</button>
          </div>
        } @else {
          <div class="card overflow-hidden p-0">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-terminal-border">
                  <th class="text-left text-terminal-muted text-xs py-3 px-4 font-medium uppercase">Nombre</th>
                  <th class="text-left text-terminal-muted text-xs py-3 px-4 font-medium uppercase">Cron</th>
                  <th class="text-left text-terminal-muted text-xs py-3 px-4 font-medium uppercase">BD</th>
                  <th class="text-left text-terminal-muted text-xs py-3 px-4 font-medium uppercase">Último Run</th>
                  <th class="text-left text-terminal-muted text-xs py-3 px-4 font-medium uppercase">Estado</th>
                  <th class="text-right text-terminal-muted text-xs py-3 px-4 font-medium uppercase">Acciones</th>
                </tr>
              </thead>
              <tbody>
                @for (job of bridge.jobs(); track job.id) {
                  <tr class="border-b border-terminal-border/50 last:border-0 hover:bg-terminal-surface/50 transition-colors">
                    <td class="py-3 px-4">
                      <div class="text-terminal-text font-medium">{{ job.name }}</div>
                      @if (!job.is_active) {
                        <span class="text-terminal-muted text-xs">⏸ Pausado</span>
                      }
                    </td>
                    <td class="py-3 px-4">
                      <code class="text-terminal-blue text-xs bg-terminal-bg px-2 py-0.5 rounded">
                        {{ job.cron_expression }}
                      </code>
                    </td>
                    <td class="py-3 px-4 text-terminal-muted text-xs">
                      {{ getDBConfigName(job.db_config_id) }}
                    </td>
                    <td class="py-3 px-4 text-terminal-muted text-xs">
                      {{ job.last_run_at ? (job.last_run_at | date:'dd/MM/yy HH:mm') : 'Nunca' }}
                    </td>
                    <td class="py-3 px-4">
                      @switch (job.last_status) {
                        @case ('success') { <span class="badge-success">✓ OK</span> }
                        @case ('error') { <span class="badge-error">✗ Error</span> }
                        @case ('running') { <span class="badge-info">↻ Ejecutando</span> }
                        @default { <span class="text-terminal-muted text-xs">—</span> }
                      }
                    </td>
                    <td class="py-3 px-4">
                      <div class="flex items-center justify-end gap-2">
                        <!-- Ejecutar manualmente -->
                        <button
                          class="text-terminal-green hover:text-green-400 transition-colors p-1"
                          title="Ejecutar ahora"
                          (click)="runNow(job)"
                          [disabled]="bridge.isLoading()">
                          ▶
                        </button>
                        <!-- Editar -->
                        <button
                          class="text-terminal-blue hover:text-blue-400 transition-colors p-1"
                          title="Editar"
                          (click)="editJob(job)">
                          ✏
                        </button>
                        <!-- Eliminar -->
                        <button
                          class="text-terminal-red hover:text-red-400 transition-colors p-1"
                          title="Eliminar"
                          (click)="deleteJob(job)"
                          [disabled]="bridge.isLoading()">
                          ✕
                        </button>
                      </div>
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </main>
    </div>
  `
})
export class JobsComponent implements OnInit {
  readonly bridge = inject(WailsBridgeService);

  /** Controla la visibilidad del formulario modal. */
  readonly showForm = signal<boolean>(false);

  /** Job que se está editando actualmente (null si es nuevo). */
  readonly editingJob = signal<Job | null>(null);

  /** Datos del formulario de job. */
  formData: Partial<Job> = this.getEmptyForm();

  /** Presets de expresiones cron comunes. */
  readonly cronPresets = [
    { label: 'Cada hora', expr: '0 0 * * * *' },
    { label: 'Diario 2am', expr: '0 0 2 * * *' },
    { label: 'Cada 6h', expr: '0 0 */6 * * *' },
    { label: 'Lunes 3am', expr: '0 0 3 * * 1' },
    { label: 'Mensual', expr: '0 0 2 1 * *' },
  ];

  ngOnInit(): void {
    this.bridge.loadJobs().catch(console.error);
    this.bridge.loadDatabaseConfigs().catch(console.error);
    this.bridge.loadDestinations().catch(console.error);
  }

  /** Abre el formulario para crear un nuevo job. */
  openForm(): void {
    this.editingJob.set(null);
    this.formData = this.getEmptyForm();
    this.showForm.set(true);
  }

  /** Abre el formulario para editar un job existente. */
  editJob(job: Job): void {
    this.editingJob.set(job);
    this.formData = { ...job };
    this.showForm.set(true);
  }

  /** Cierra el formulario modal. */
  closeForm(): void {
    this.showForm.set(false);
    this.editingJob.set(null);
    this.formData = this.getEmptyForm();
  }

  /** Aplica un preset de expresión cron al formulario. */
  applyCronPreset(expr: string): void {
    this.formData.cron_expression = expr;
  }

  /** Verifica si un destino está seleccionado en el formulario. */
  isDestinationSelected(destId: number): boolean {
    return (this.formData.destination_ids ?? []).includes(destId);
  }

  /** Activa o desactiva un destino en la selección. */
  toggleDestination(destId: number, event: Event): void {
    const checkbox = event.target as HTMLInputElement;
    const current = this.formData.destination_ids ?? [];
    if (checkbox.checked) {
      this.formData.destination_ids = [...current, destId];
    } else {
      this.formData.destination_ids = current.filter(id => id !== destId);
    }
  }

  /** Guarda el job (crea o actualiza según si tiene ID). */
  async saveJob(): Promise<void> {
    if (!this.formData.name || !this.formData.cron_expression || !this.formData.db_config_id) {
      return;
    }
    try {
      await this.bridge.saveJob(this.formData as Job);
      this.closeForm();
    } catch {
      // El error se muestra via bridge.errorMessage signal
    }
  }

  /** Ejecuta un job de forma manual. */
  async runNow(job: Job): Promise<void> {
    if (confirm(`¿Ejecutar "${job.name}" ahora?`)) {
      try {
        await this.bridge.runJobNow(job.id);
      } catch {
        // El error se muestra via bridge.errorMessage signal
      }
    }
  }

  /** Elimina un job con confirmación. */
  async deleteJob(job: Job): Promise<void> {
    if (confirm(`¿Eliminar el job "${job.name}"? Esta acción no se puede deshacer.`)) {
      try {
        await this.bridge.deleteJob(job.id);
      } catch {
        // El error se muestra via bridge.errorMessage signal
      }
    }
  }

  /** Retorna el nombre de la configuración de BD por su ID. */
  getDBConfigName(dbConfigId: number): string {
    const config = this.bridge.dbConfigs().find(c => c.id === dbConfigId);
    return config ? config.name : `BD #${dbConfigId}`;
  }

  /** Retorna un formulario vacío con valores por defecto. */
  private getEmptyForm(): Partial<Job> {
    return {
      id: 0,
      name: '',
      cron_expression: '0 0 2 * * *',
      db_config_id: 0,
      destination_ids: [],
      notify_webhook: '',
      is_active: true,
    };
  }
}
