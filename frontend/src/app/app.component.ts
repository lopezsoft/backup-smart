// Componente raíz de backup-smart.
// Aplica el tema dark y el router outlet principal.
import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

/**
 * AppComponent es el componente raíz de la aplicación.
 * Solo contiene el router-outlet; la navegación y el layout
 * están manejados por cada componente de ruta individualmente.
 */
@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  template: `
    <!-- Clase 'dark' aplicada al contenedor raíz para el modo oscuro de TailwindCSS -->
    <div class="dark min-h-screen bg-terminal-bg">
      <router-outlet />
    </div>
  `
})
export class AppComponent {
  title = 'backup-smart';
}
