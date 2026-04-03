// Componente Dashboard de backup-smart.
// Muestra métricas de uso, estado de licencia y próximas ejecuciones de cron.
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { WailsBridgeService } from '../services/wails-bridge.service';
import { NavComponent } from '../shared/nav.component';

/**
 * DashboardComponent es el panel principal de backup-smart.
 * Muestra un resumen del estado del sistema: jobs activos, licencia,
 * últimos logs y métricas de respaldo.
 */
@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, DatePipe, RouterLink, NavComponent],
  template: `
    <div class="flex h-screen bg-terminal-bg overflow-hidden">
      <!-- Sidebar de navegación -->
      <app-nav />

      <!-- Contenido principal -->
      <main class="flex-1 overflow-y-auto p-6">
        <!-- Header -->
        <div class="mb-6 flex items-center justify-between">
          <div>
            <h1 class="text-terminal-text text-2xl font-bold">Dashboard</h1>
            <p class="text-terminal-muted text-sm mt-1">
              {{ now | date:'EEEE, d MMMM y HH:mm' }}
            </p>
          </div>
          <!-- Indicador de carga -->
          @if (bridge.isLoading()) {
            <div class="text-terminal-blue text-sm flex items-center gap-2">
              <span class="animate-spin inline-block">&#x21BB;</span>
              <span>Actualizando...</span>
            </div>
          }
        </div>

        <!-- Grid de métricas principales -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <!-- Total de jobs -->
          <div class="card flex items-start gap-4">
            <div class="text-3xl">&#x1F504;</div>
            <div>
              <div class="text-terminal-muted text-xs uppercase tracking-wide">Total Jobs</div>
              <div class="text-terminal-text text-2xl font-bold">{{ bridge.jobs().length }}</div>
              <div class="text-terminal-green text-xs mt-1">
                {{ bridge.activeJobsCount() }} activos
              </div>
            </div>
          </div>

          <!-- Configuraciones de BD -->
          <div class="card flex items-start gap-4">
            <div class="text-3xl">&#x1F5C4;</div>
            <div>
              <div class="text-terminal-muted text-xs uppercase tracking-wide">Bases de Datos</div>
              <div class="text-terminal-text text-2xl font-bold">{{ bridge.dbConfigs().length }}</div>
              <div class="text-terminal-blue text-xs mt-1">configuradas</div>
            </div>
          </div>

          <!-- Destinos -->
          <div class="card flex items-start gap-4">
            <div class="text-3xl">&#x2601;</div>
            <div>
              <div class="text-terminal-muted text-xs uppercase tracking-wide">Destinos</div>
              <div class="text-terminal-text text-2xl font-bold">{{ bridge.destinations().length }}</div>
              <div class="text-terminal-blue text-xs mt-1">de almacenamiento</div>
            </div>
          </div>

          <!-- Estado de licencia -->
          <div class="card flex items-start gap-4">
            <div class="text-3xl">&#x1F511;</div>
            <div>
              <div class="text-terminal-muted text-xs uppercase tracking-wide">Licencia</div>
              <div class="text-terminal-text text-lg font-bold capitalize">
                {{ bridge.license()?.plan ?? 'Free' }}
              </div>
              @if (bridge.isProLicensed()) {
                <div class="badge-success text-xs mt-1">&#x2713; Activa</div>
              } @else {
                <div class="badge-warning text-xs mt-1">Plan gratuito</div>
              }
            </div>
          </div>
        </div>

        <!-- Estado de Jobs Recientes -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Lista de jobs con su último estado -->
          <div class="card">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-terminal-text font-semibold">Jobs de Respaldo</h2>
              <a routerLink="/jobs" class="text-terminal-blue text-xs hover:underline">
                Administrar →
              </a>
            </div>

            @if (bridge.jobs().length === 0) {
              <div class="text-terminal-muted text-sm text-center py-6">
                <div class="text-4xl mb-2">&#x1F4C2;</div>
                <p>No hay jobs configurados.</p>
                <a routerLink="/jobs" class="text-terminal-blue hover:underline">Crear primer job</a>
              </div>
            } @else {
              <ul class="space-y-2">
                @for (job of bridge.jobs().slice(0, 5); track job.id) {
                  <li class="flex items-center justify-between py-2 border-b border-terminal-border/50 last:border-0">
                    <div class="flex items-center gap-3">
                      <!-- Indicador de estado -->
                      <span [class]="getStatusDot(job.last_status)"></span>
                      <div>
                        <div class="text-terminal-text text-sm font-medium">{{ job.name }}</div>
                        <div class="text-terminal-muted text-xs font-mono">{{ job.cron_expression }}</div>
                      </div>
                    </div>
                    <div class="text-right">
                      @if (job.last_run_at) {
                        <div class="text-terminal-muted text-xs">
                          {{ job.last_run_at | date:'dd/MM HH:mm' }}
                        </div>
                      }
                      @if (!job.is_active) {
                        <span class="text-terminal-muted text-xs">Pausado</span>
                      }
                    </div>
                  </li>
                }
              </ul>
            }
          </div>

          <!-- Últimos logs -->
          <div class="card">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-terminal-text font-semibold">Actividad Reciente</h2>
              <a routerLink="/logs" class="text-terminal-blue text-xs hover:underline">
                Ver consola →
              </a>
            </div>

            @if (bridge.logs().length === 0) {
              <div class="text-terminal-muted text-sm text-center py-6">
                <div class="text-4xl mb-2">&#x1F4CB;</div>
                <p>No hay actividad registrada.</p>
              </div>
            } @else {
              <ul class="space-y-1 font-mono text-xs">
                @for (log of bridge.logs().slice(0, 8); track log.id) {
                  <li class="flex items-start gap-2 py-1 border-b border-terminal-border/30 last:border-0">
                    <span [class]="getLogLevelClass(log.level)" class="flex-shrink-0">
                      {{ getLogLevelIcon(log.level) }}
                    </span>
                    <span class="text-terminal-muted flex-shrink-0">
                      {{ log.created_at | date:'HH:mm:ss' }}
                    </span>
                    <span class="text-terminal-text truncate">{{ log.message }}</span>
                  </li>
                }
              </ul>
            }
          </div>
        </div>

        <!-- Hardware ID y licencia -->
        @if (bridge.hardwareID()) {
          <div class="mt-6 card">
            <div class="flex items-center gap-2 text-terminal-muted text-xs">
              <span>&#x1F4BB;</span>
              <span>Hardware ID:</span>
              <code class="text-terminal-blue font-mono bg-terminal-bg px-2 py-0.5 rounded text-xs">
                {{ bridge.hardwareID() }}
              </code>
            </div>
          </div>
        }
      </main>
    </div>
  `
})
export class DashboardComponent implements OnInit {
  readonly bridge = inject(WailsBridgeService);
  readonly now = new Date();

  ngOnInit(): void {
    // Recargamos los datos al entrar al dashboard
    this.bridge.loadJobs().catch(console.error);
    this.bridge.loadLogs().catch(console.error);
  }

  /** Retorna la clase CSS para el punto de estado de un job. */
  getStatusDot(status: string | undefined): string {
    const base = 'inline-block w-2 h-2 rounded-full flex-shrink-0 mt-1.5';
    switch (status) {
      case 'success': return `${base} bg-terminal-green`;
      case 'error': return `${base} bg-terminal-red`;
      case 'running': return `${base} bg-terminal-blue animate-pulse`;
      default: return `${base} bg-terminal-border`;
    }
  }

  /** Retorna la clase CSS para el color del nivel de log. */
  getLogLevelClass(level: string): string {
    switch (level) {
      case 'success': return 'text-terminal-green';
      case 'error': return 'text-terminal-red';
      case 'warn': return 'text-terminal-yellow';
      default: return 'text-terminal-blue';
    }
  }

  /** Retorna el icono para el nivel de log. */
  getLogLevelIcon(level: string): string {
    switch (level) {
      case 'success': return '✓';
      case 'error': return '✗';
      case 'warn': return '⚠';
      default: return '●';
    }
  }
}
