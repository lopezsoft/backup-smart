// Configuración de rutas de backup-smart.
// Las rutas usan carga diferida (lazy loading) para optimizar el tiempo de inicio.
import { Routes } from '@angular/router';
import { authGuard } from './auth/auth.guard';

export const routes: Routes = [
  {
    // Ruta raíz: redirige al dashboard o al login según estado de autenticación
    path: '',
    redirectTo: '/auth',
    pathMatch: 'full'
  },
  {
    // Ruta de autenticación: ingreso de Master Password y License Key
    path: 'auth',
    loadComponent: () => import('./auth/auth.component').then(m => m.AuthComponent)
  },
  {
    // Dashboard principal: métricas, estado de licencia, próximos crons
    path: 'dashboard',
    loadComponent: () => import('./dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [authGuard]
  },
  {
    // Gestión de jobs de respaldo: CRUD con expresiones cron
    path: 'jobs',
    loadComponent: () => import('./jobs/jobs.component').then(m => m.JobsComponent),
    canActivate: [authGuard]
  },
  {
    // Consola de logs en tiempo real: terminal UI
    path: 'logs',
    loadComponent: () => import('./logs/logs.component').then(m => m.LogsComponent),
    canActivate: [authGuard]
  },
  {
    // Ruta de captura de URLs no encontradas
    path: '**',
    redirectTo: '/auth'
  }
];
