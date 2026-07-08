package main

import (
	"context"
	"embed"
	"log"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// wailsContext is an alias so the anonymous function signature resolves correctly.
type wailsContext = context.Context

func main() {
	// Load .env so CHIMERA_API_KEY, CHIMERA_BASE_URL, etc. are available
	// via os.Getenv. Missing .env is not fatal (env vars may be set directly).
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables directly")
	}

	app := NewApp()

	bindings, err := buildAgentBindings()
	if err != nil {
		log.Fatalf("agent wiring: %v", err)
	}
	agentApp, err := NewAgentApp(bindings)
	if err != nil {
		log.Fatalf("agent app: %v", err)
	}

	err = wails.Run(&options.App{
		Title:            "Relic Ring Protocol",
		Width:            1440,
		Height:           900,
		MinWidth:         1100,
		MinHeight:        700,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 7, G: 11, B: 30, A: 1},
		OnStartup: func(ctx wailsContext) {
			app.startup(ctx)
			agentApp.Startup(ctx)
		},
		OnShutdown: app.shutdown,
		Bind:       []any{app, agentApp},
	})
	if err != nil {
		log.Fatal(err)
	}
}
