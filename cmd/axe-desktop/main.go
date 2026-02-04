package main

import (
	"fmt"
	"os"

	"axe-desktop/internal/agent"
	"axe-desktop/internal/config"
	"axe-desktop/internal/storage"
	"axe-desktop/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	store, err := storage.New(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize agent service
	agentService, err := agent.NewService(cfg, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize agent service: %v\n", err)
		os.Exit(1)
	}

	// Create Fyne app with Vercel theme
	a := app.New()
	a.Settings().SetTheme(ui.NewVercelTheme())

	w := a.NewWindow("Axe Desktop")
	w.Resize(fyne.NewSize(1400, 900))
	w.CenterOnScreen()

	// Create and run UI
	mainUI := ui.New(w, store, cfg, agentService)
	mainUI.Initialize()

	w.ShowAndRun()
}
