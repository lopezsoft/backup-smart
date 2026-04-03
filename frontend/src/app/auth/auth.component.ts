// Componente de autenticación de backup-smart.
// Permite ingresar la contraseña maestra y la clave de licencia.
import { Component, signal, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { WailsBridgeService } from '../services/wails-bridge.service';

/**
 * AuthComponent maneja el flujo de autenticación inicial de la aplicación.
 * En el primer uso, solicita crear una contraseña maestra.
 * En usos posteriores, valida la contraseña contra el hash almacenado.
 */
@Component({
  selector: 'app-auth',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="min-h-screen bg-terminal-bg flex items-center justify-center p-4">
      <div class="w-full max-w-md">
        <!-- Logo y título -->
        <div class="text-center mb-8">
          <div class="text-terminal-green text-5xl mb-3 font-bold font-mono">
            <span class="text-terminal-blue">&#x1F5C4;</span> backup-smart
          </div>
          <p class="text-terminal-muted text-sm">
            DevTool de respaldo inteligente para SysAdmins y DevOps
          </p>
          <div class="mt-2 flex items-center justify-center gap-2">
            <span class="inline-block w-2 h-2 rounded-full bg-terminal-green animate-pulse"></span>
            <span class="text-terminal-green text-xs">Sistema listo</span>
          </div>
        </div>

        <!-- Tarjeta de login -->
        <div class="card border-terminal-border">
          <h2 class="text-terminal-text text-lg font-semibold mb-6 flex items-center gap-2">
            <span class="text-terminal-blue">&#x1F511;</span>
            {{ isFirstSetup() ? 'Configuración inicial' : 'Acceso al sistema' }}
          </h2>

          <!-- Indicador de error global -->
          @if (errorMsg()) {
            <div class="mb-4 p-3 rounded border border-terminal-red/50 bg-red-900/20 text-terminal-red text-sm flex items-start gap-2">
              <span>&#x26A0;</span>
              <span>{{ errorMsg() }}</span>
            </div>
          }

          <form (ngSubmit)="onSubmit()" #authForm="ngForm">
            <!-- Contraseña maestra -->
            <div class="mb-4">
              <label class="form-label">
                {{ isFirstSetup() ? 'Crear contraseña maestra' : 'Contraseña maestra' }}
              </label>
              <input
                type="password"
                class="form-input"
                name="masterPassword"
                [(ngModel)]="masterPassword"
                required
                minlength="8"
                placeholder="Mínimo 8 caracteres"
                [disabled]="bridge.isLoading()"
                autocomplete="current-password"
              />
              @if (isFirstSetup()) {
                <p class="text-terminal-muted text-xs mt-1">
                  &#x26A0; Esta contraseña encripta todas tus credenciales. No puede recuperarse.
                </p>
              }
            </div>

            <!-- Confirmación de contraseña (solo en primer setup) -->
            @if (isFirstSetup()) {
              <div class="mb-4">
                <label class="form-label">Confirmar contraseña maestra</label>
                <input
                  type="password"
                  class="form-input"
                  name="confirmPassword"
                  [(ngModel)]="confirmPassword"
                  required
                  placeholder="Repetir contraseña"
                  [disabled]="bridge.isLoading()"
                  autocomplete="new-password"
                />
              </div>

              <!-- Clave de licencia (opcional en setup) -->
              <div class="mb-6">
                <label class="form-label">
                  Clave de licencia <span class="text-terminal-muted">(opcional - Pro)</span>
                </label>
                <input
                  type="text"
                  class="form-input"
                  name="licenseKey"
                  [(ngModel)]="licenseKey"
                  placeholder="XXXX-XXXX-XXXX-XXXX"
                  [disabled]="bridge.isLoading()"
                />
                <p class="text-terminal-muted text-xs mt-1">
                  ID de hardware: <code class="text-terminal-blue">{{ hardwareID() || 'Cargando...' }}</code>
                </p>
              </div>
            }

            <!-- Botón de submit -->
            <button
              type="submit"
              class="btn-primary w-full flex items-center justify-center gap-2"
              [disabled]="bridge.isLoading() || !masterPassword"
            >
              @if (bridge.isLoading()) {
                <span class="animate-spin">&#x21BB;</span>
                <span>Procesando...</span>
              } @else {
                <span>{{ isFirstSetup() ? '&#x2713; Configurar sistema' : '&#x1F513; Ingresar' }}</span>
              }
            </button>
          </form>

          <!-- Enlace para cambiar entre modos -->
          <div class="mt-4 text-center">
            <button
              type="button"
              class="text-terminal-muted text-xs hover:text-terminal-blue transition-colors"
              (click)="toggleMode()"
            >
              {{ isFirstSetup() ? '¿Ya tienes una contraseña? Ingresar' : '¿Primera vez? Configurar' }}
            </button>
          </div>
        </div>

        <!-- Footer -->
        <p class="text-center text-terminal-muted text-xs mt-6">
          backup-smart v1.0.0 · LopezSoft · 2024
        </p>
      </div>
    </div>
  `
})
export class AuthComponent {
  /** Bridge de Wails para comunicación con el backend. */
  readonly bridge = inject(WailsBridgeService);
  private readonly router = inject(Router);

  // ─── Estado del formulario ───────────────────────────────────────────

  /** Contraseña maestra ingresada. */
  masterPassword = '';
  /** Confirmación de contraseña (solo en primer setup). */
  confirmPassword = '';
  /** Clave de licencia Pro (opcional). */
  licenseKey = '';

  /** Indica si está en modo de primer setup (true) o login (false). */
  readonly isFirstSetup = signal<boolean>(true);

  /** Mensaje de error a mostrar al usuario. */
  readonly errorMsg = signal<string>('');

  /** ID de hardware del equipo (para activación de licencia). */
  readonly hardwareID = signal<string>('');

  constructor() {
    // Cargamos el hardware ID al iniciar para mostrarlo al usuario
    this.bridge.loadHardwareID().then(() => {
      this.hardwareID.set(this.bridge.hardwareID());
    }).catch(() => {
      this.hardwareID.set('No disponible');
    });
  }

  /** Alterna entre modo de primer setup y modo de login. */
  toggleMode(): void {
    this.isFirstSetup.update(v => !v);
    this.errorMsg.set('');
    this.masterPassword = '';
    this.confirmPassword = '';
  }

  /** Procesa el formulario de autenticación. */
  async onSubmit(): Promise<void> {
    this.errorMsg.set('');

    if (!this.masterPassword || this.masterPassword.length < 8) {
      this.errorMsg.set('La contraseña maestra debe tener al menos 8 caracteres.');
      return;
    }

    try {
      if (this.isFirstSetup()) {
        // Validamos que las contraseñas coincidan
        if (this.masterPassword !== this.confirmPassword) {
          this.errorMsg.set('Las contraseñas no coinciden.');
          return;
        }
        // Inicializamos el sistema por primera vez
        await this.bridge.initSetup(this.masterPassword, this.licenseKey);
      } else {
        // Intentamos el login con la contraseña existente
        const ok = await this.bridge.login(this.masterPassword);
        if (!ok) {
          this.errorMsg.set('Contraseña incorrecta. Intenta nuevamente.');
          return;
        }
      }
      // Redirigimos al dashboard tras autenticación exitosa
      await this.router.navigate(['/dashboard']);
    } catch (err) {
      this.errorMsg.set(err instanceof Error ? err.message : 'Error inesperado.');
    }
  }
}
