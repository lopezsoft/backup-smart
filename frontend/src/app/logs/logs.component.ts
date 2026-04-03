// Componente de consola de logs de backup-smart.
// Terminal UI en tiempo real que muestra eventos del sistema y de los jobs.
import { Component, OnInit, OnDestroy, inject, signal, ViewChild, ElementRef, effect } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { WailsBridgeService } from '../services/wails-bridge.service';
import { NavComponent } from '../shared/nav.component';
import type { LogEntry, LogLevel } from '../models';

/**
 * LogsComponent muestra una consola en tiempo real con los logs de actividad.
 * Escucha eventos del backend Go mediante el bridge de Wails.
 * Simula la estética de una terminal para la experiencia de SysAdmin.
 */
@Component({
  selector: 'app-logs',
  standalone: true,
  imports: [CommonModule, DatePipe, NavComponent],
  template: `
    <div class="flex h-screen bg-terminal-bg overflow-hidden">
      <app-nav />

      <main class="flex-1 flex flex-col overflow-hidden p-6">
        <!-- Header de la consola -->
        <div class="mb-4 flex items-center justify-between flex-shrink-0">
          <div>
            <h1 class="text-terminal-text text-2xl font-bold flex items-center gap-2">
              <span class="text-terminal-green">&#x1F4BB;</span> Consola de Actividad
            </h1>
            <p class="text-terminal-muted text-sm mt-1">
              Logs en tiempo real del sistema de respaldo
            </p>
          </div>
          <div class="flex items-center gap-3">
            <!-- Filtro por nivel -->
            <select
              class="form-input w-auto py-1 text-sm"
              (change)="onFilterChange($event)">
              <option value="">Todos los niveles</option>
              <option value="info">Info</option>
              <option value="success">Success</option>
              <option value="warn">Warning</option>
              <option value="error">Error</option>
            </select>

            <!-- Auto-scroll toggle -->
            <label class="flex items-center gap-2 text-terminal-muted text-sm cursor-pointer">
              <input type="checkbox"
                class="accent-terminal-green"
                [checked]="autoScroll()"
                (change)="toggleAutoScroll()" />
              Auto-scroll
            </label>

            <!-- Limpiar logs de la vista -->
            <button class="btn-secondary py-1 text-sm" (click)="clearView()">
              Limpiar
            </button>

            <!-- Recargar desde BD -->
            <button class="btn-primary py-1 text-sm flex items-center gap-1"
              (click)="reloadLogs()"
              [disabled]="bridge.isLoading()">
              @if (bridge.isLoading()) {
                <span class="animate-spin">↻</span>
              } @else {
                ↻ Recargar
              }
            </button>
          </div>
        </div>

        <!-- Terminal de logs -->
        <div #logContainer
          class="flex-1 bg-[#0a0e13] rounded-lg border border-terminal-border overflow-y-auto font-mono text-xs p-4"
          style="min-height: 0;">

          <!-- Cabecera de la terminal -->
          <div class="text-terminal-muted mb-3 pb-2 border-b border-terminal-border/50">
            <span class="text-terminal-green">backup-smart</span>
            <span class="text-terminal-muted">&#64;</span>
            <span class="text-terminal-blue">system</span>
            <span class="text-terminal-muted">:~$ </span>
            <span class="text-terminal-text">tail -f /var/log/backup-smart.log</span>
          </div>

          <!-- Líneas de log -->
          @if (filteredLogs().length === 0) {
            <div class="text-terminal-muted text-center py-8">
              <div class="text-3xl mb-2">&#x1F4CB;</div>
              <p>No hay logs para mostrar.</p>
              @if (levelFilter()) {
                <p class="text-xs mt-1">Filtro activo: {{ levelFilter() }}</p>
              }
            </div>
          } @else {
            @for (log of filteredLogs(); track log.id) {
              <div class="flex items-start gap-2 py-0.5 hover:bg-white/5 rounded px-1 group">
                <!-- Timestamp -->
                <span class="text-terminal-muted/60 flex-shrink-0 tabular-nums">
                  {{ log.created_at | date:'HH:mm:ss.SSS' }}
                </span>

                <!-- Nivel del log -->
                <span [class]="getLevelBadgeClass(log.level)" class="flex-shrink-0 font-bold uppercase w-8 text-center">
                  {{ getLevelShort(log.level) }}
                </span>

                <!-- Nombre del job (si aplica) -->
                @if (log.job_name) {
                  <span class="text-terminal-purple flex-shrink-0 max-w-[12ch] truncate" [title]="log.job_name">
                    [{{ log.job_name }}]
                  </span>
                } @else {
                  <span class="text-terminal-muted/40 flex-shrink-0">[sys]</span>
                }

                <!-- Mensaje -->
                <span [class]="getMessageClass(log.level)" class="break-all leading-relaxed">
                  {{ log.message }}
                </span>
              </div>
            }
          }

          <!-- Cursor parpadeante -->
          <div class="mt-2 text-terminal-muted">
            <span class="text-terminal-green">&#x25B6;</span>
            <span class="animate-pulse text-terminal-green">█</span>
          </div>
        </div>

        <!-- Barra de estado inferior -->
        <div class="mt-2 flex items-center justify-between text-xs text-terminal-muted flex-shrink-0">
          <span>
            {{ filteredLogs().length }} entradas
            @if (levelFilter()) {
              <span>(filtro: {{ levelFilter() }})</span>
            }
          </span>
          <span class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-terminal-green animate-pulse"></span>
            Escuchando eventos en tiempo real
          </span>
        </div>
      </main>
    </div>
  `
})
export class LogsComponent implements OnInit {
  readonly bridge = inject(WailsBridgeService);

  @ViewChild('logContainer') private logContainer!: ElementRef<HTMLDivElement>;

  /** Controla si se hace scroll automático al recibir nuevos logs. */
  readonly autoScroll = signal<boolean>(true);

  /** Filtro de nivel de log activo. */
  readonly levelFilter = signal<string>('');

  /** Logs filtrados según el nivel seleccionado. */
  readonly filteredLogs = signal<LogEntry[]>([]);

  constructor() {
    // Efecto que actualiza los logs filtrados cuando cambian los logs o el filtro
    effect(() => {
      const logs = this.bridge.logs();
      const filter = this.levelFilter();
      const filtered = filter ? logs.filter(l => l.level === filter) : logs;
      this.filteredLogs.set(filtered);

      // Auto-scroll al último log
      if (this.autoScroll()) {
        setTimeout(() => this.scrollToBottom(), 50);
      }
    });
  }

  ngOnInit(): void {
    this.bridge.loadLogs(200).catch(console.error);
  }

  /** Maneja el cambio de filtro de nivel. */
  onFilterChange(event: Event): void {
    const select = event.target as HTMLSelectElement;
    this.levelFilter.set(select.value);
  }

  /** Alterna el auto-scroll. */
  toggleAutoScroll(): void {
    this.autoScroll.update(v => !v);
  }

  /** Limpia la vista de logs (solo visualmente, no elimina de la BD). */
  clearView(): void {
    this.bridge.logs.set([]);
    this.filteredLogs.set([]);
  }

  /** Recarga los logs desde la base de datos. */
  async reloadLogs(): Promise<void> {
    await this.bridge.loadLogs(500);
  }

  /** Hace scroll al final del contenedor de logs. */
  private scrollToBottom(): void {
    if (this.logContainer?.nativeElement) {
      const el = this.logContainer.nativeElement;
      el.scrollTop = el.scrollHeight;
    }
  }

  /** Retorna la clase CSS para el badge de nivel de log. */
  getLevelBadgeClass(level: LogLevel | string): string {
    switch (level) {
      case 'success': return 'text-terminal-green';
      case 'error': return 'text-terminal-red';
      case 'warn': return 'text-terminal-yellow';
      default: return 'text-terminal-blue';
    }
  }

  /** Retorna la abreviatura del nivel de log (3 caracteres). */
  getLevelShort(level: LogLevel | string): string {
    switch (level) {
      case 'success': return 'OK';
      case 'error': return 'ERR';
      case 'warn': return 'WRN';
      default: return 'INF';
    }
  }

  /** Retorna la clase CSS del texto del mensaje según el nivel. */
  getMessageClass(level: LogLevel | string): string {
    switch (level) {
      case 'error': return 'text-terminal-red/90';
      case 'warn': return 'text-terminal-yellow/90';
      case 'success': return 'text-terminal-green/90';
      default: return 'text-terminal-text/80';
    }
  }
}
