// Package main es el punto de entrada de la aplicación backup-smart.
// Utiliza Wails v2 para crear una aplicación de escritorio con frontend Angular.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// assets embebe el contenido del directorio frontend/dist para ser servido por Wails.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Creamos la instancia principal de la aplicación
	app := NewApp()

	// Inicializamos la aplicación Wails con las opciones de configuración
	err := wails.Run(&options.App{
		Title:  "backup-smart",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 18, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error al iniciar la aplicación: %v", err)
	}
}
