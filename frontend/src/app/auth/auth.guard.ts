// Guard de autenticación para proteger rutas que requieren login.
// Redirige a /auth si el usuario no ha iniciado sesión.
import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { WailsBridgeService } from '../services/wails-bridge.service';

/**
 * Guard que verifica si el usuario está autenticado.
 * Si no lo está, redirige a la pantalla de login (/auth).
 */
export const authGuard: CanActivateFn = () => {
  const bridge = inject(WailsBridgeService);
  const router = inject(Router);

  if (bridge.isAuthenticated()) {
    return true;
  }

  // Redirigimos al usuario a la pantalla de autenticación
  return router.createUrlTree(['/auth']);
};
