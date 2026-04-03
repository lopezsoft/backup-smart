// Configuración principal de la aplicación Angular de backup-smart.
import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';

import { routes } from './app.routes';

/**
 * Configuración de la aplicación Angular de backup-smart.
 * Registra los providers globales: router con lazy loading y change detection optimizado.
 * No se usa provideHttpClient ya que la comunicación con el backend Go se realiza
 * exclusivamente a través del bridge IPC de Wails.
 */
export const appConfig: ApplicationConfig = {
  providers: [
    // Optimización de change detection con Zone.js
    provideZoneChangeDetection({ eventCoalescing: true }),
    // Router con soporte para input binding en componentes de ruta
    provideRouter(routes, withComponentInputBinding()),
  ]
};
