// Configuración principal de la aplicación Angular de backup-smart.
import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';

import { routes } from './app.routes';

/**
 * Configuración de la aplicación Angular de backup-smart.
 * Registra los providers globales: router con lazy loading, HTTP client.
 */
export const appConfig: ApplicationConfig = {
  providers: [
    // Optimización de change detection con Zone.js
    provideZoneChangeDetection({ eventCoalescing: true }),
    // Router con soporte para input binding en componentes
    provideRouter(routes, withComponentInputBinding()),
    // HTTP client para llamadas externas (webhooks, etc.)
    provideHttpClient(),
  ]
};
