// Package ports define las interfaces (puertos) de la capa de casos de uso.
package ports

import "context"

// Notifier define la interfaz para enviar notificaciones al usuario.
type Notifier interface {
	// Notify envía un mensaje de notificación al webhook especificado.
	// El webhook puede ser una URL de Slack, Discord, Teams, etc.
	Notify(ctx context.Context, msg string, webhook string) error
}
