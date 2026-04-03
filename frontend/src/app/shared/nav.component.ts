// Componente de navegación lateral compartido.
// Contiene el menú principal de la aplicación con Dark Mode.
import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { WailsBridgeService } from '../services/wails-bridge.service';

/**
 * NavComponent es el sidebar de navegación principal de backup-smart.
 * Muestra las rutas principales y el estado del sistema.
 */
@Component({
  selector: 'app-nav',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  template: `
    <nav class="h-screen w-56 bg-terminal-surface border-r border-terminal-border flex flex-col">
      <!-- Logo -->
      <div class="p-4 border-b border-terminal-border">
        <div class="text-terminal-text font-bold flex items-center gap-2">
          <span class="text-terminal-blue text-xl">&#x1F5C4;</span>
          <span>backup<span class="text-terminal-green">-smart</span></span>
        </div>
        <div class="text-terminal-muted text-xs mt-1">v1.0.0</div>
      </div>

      <!-- Menú de navegación -->
      <div class="flex-1 p-2 overflow-y-auto">
        <ul class="space-y-1">
          @for (item of navItems; track item.path) {
            <li>
              <a
                [routerLink]="item.path"
                routerLinkActive="bg-terminal-bg text-terminal-blue border-l-2 border-terminal-blue"
                class="flex items-center gap-3 px-3 py-2 rounded text-terminal-muted hover:text-terminal-text hover:bg-terminal-bg transition-colors text-sm"
              >
                <span>{{ item.icon }}</span>
                <span>{{ item.label }}</span>
              </a>
            </li>
          }
        </ul>
      </div>

      <!-- Estado de la licencia -->
      <div class="p-3 border-t border-terminal-border">
        @if (bridge.isProLicensed()) {
          <div class="badge-success w-full justify-center text-xs">
            &#x2605; Licencia Pro activa
          </div>
        } @else {
          <div class="badge-warning w-full justify-center text-xs">
            Plan gratuito
          </div>
        }
      </div>
    </nav>
  `
})
export class NavComponent {
  readonly bridge = inject(WailsBridgeService);

  /** Elementos del menú de navegación. */
  readonly navItems = [
    { path: '/dashboard', icon: '&#x1F4CA;', label: 'Dashboard' },
    { path: '/jobs', icon: '&#x1F504;', label: 'Jobs de Respaldo' },
    { path: '/logs', icon: '&#x1F4BB;', label: 'Consola de Logs' },
  ];
}
