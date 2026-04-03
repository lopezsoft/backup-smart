// Stubs del runtime de Wails para eventos IPC.
// Wails reemplaza este módulo con la implementación real al compilar.

type EventCallback = (...data: unknown[]) => void;

// Registro interno de listeners de eventos
const eventListeners: Map<string, EventCallback[]> = new Map();

/**
 * Suscribe una función callback al evento especificado de Wails.
 * Retorna una función para cancelar la suscripción.
 */
export function EventsOn(eventName: string, callback: EventCallback): () => void {
  const listeners = eventListeners.get(eventName) ?? [];
  listeners.push(callback);
  eventListeners.set(eventName, listeners);

  // Retornamos la función de cancelación
  return () => EventsOff(eventName, callback);
}

/**
 * Emite un evento hacia el backend Go de Wails.
 * En modo de desarrollo (sin runtime), los eventos se registran en consola.
 */
export function EventsEmit(eventName: string, ...data: unknown[]): void {
  // Notificamos a los listeners locales registrados
  const listeners = eventListeners.get(eventName) ?? [];
  listeners.forEach(cb => cb(...data));
}

/**
 * Cancela la suscripción de una función callback a un evento.
 */
export function EventsOff(eventName: string, callback: EventCallback): void {
  const listeners = eventListeners.get(eventName) ?? [];
  const filtered = listeners.filter(cb => cb !== callback);
  if (filtered.length === 0) {
    eventListeners.delete(eventName);
  } else {
    eventListeners.set(eventName, filtered);
  }
}

/**
 * Cancela todas las suscripciones a un evento.
 */
export function EventsOffAll(eventName: string): void {
  eventListeners.delete(eventName);
}

/**
 * Suscribe al evento especificado solo una vez; se cancela automáticamente después del primer disparo.
 */
export function EventsOnce(eventName: string, callback: EventCallback): void {
  const wrapper: EventCallback = (...data) => {
    callback(...data);
    EventsOff(eventName, wrapper);
  };
  EventsOn(eventName, wrapper);
}
