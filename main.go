package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/services"
	"github.com/lhommenul/brique/pkg/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx              context.Context
	cfg              *config.Config
	database         *db.Database
	backpackService  *services.BackpackService
	gossipService    *services.GossipService
	discoveryService *services.DiscoveryService
	logger           *slog.Logger
	events           *EventEmitter
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Setup event emitter
	a.events = NewEventEmitter(ctx)

	// Setup logger
	a.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	var err error
	a.cfg, err = config.Load()
	if err != nil {
		a.logger.Error("Failed to load config", "error", err)
		a.events.Error("Configuration Error", "Failed to load application configuration")
		os.Exit(1)
	}

	a.logger.Info("Configuration loaded", "data_dir", a.cfg.DataDir)

	// Initialize database
	a.database, err = db.NewDatabase(a.cfg.DatabasePath, a.logger)
	if err != nil {
		a.logger.Error("Failed to initialize database", "error", err)
		a.events.Error("Database Error", "Failed to initialize database")
		os.Exit(1)
	}

	// Create queries
	queries := db.New(a.database.DB)

	// Create backpack service
	a.backpackService = services.NewBackpackService(queries, a.cfg.AssetsDir)

	// Create gossip service
	instanceName := fmt.Sprintf("Brique-%s", os.Getenv("USER"))
	a.gossipService = services.NewGossipService(queries, instanceName, "localhost:9090")

	// Get instance info
	instanceInfo, err := a.gossipService.GetInstanceInfo(ctx)
	if err != nil {
		a.logger.Error("Failed to get instance info", "error", err)
		os.Exit(1)
	}

	// Create discovery service
	a.discoveryService = services.NewDiscoveryService(
		instanceInfo.InstanceID,
		instanceName,
		9090, // Port for gossip API
		a.logger,
		a.gossipService,
	)

	// Start discovery (announce and browse)
	if err := a.discoveryService.Start(ctx); err != nil {
		a.logger.Warn("Failed to start discovery service", "error", err)
		// Not a fatal error, continue without discovery
	}

	a.logger.Info("Application initialized successfully")
	a.events.Success("Brique démarré", "L'application est prête")
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Stop discovery service
	if a.discoveryService != nil {
		if err := a.discoveryService.Stop(); err != nil {
			a.logger.Error("Failed to stop discovery service", "error", err)
		}
	}

	// Close database
	if a.database != nil {
		a.database.Close()
	}

	a.logger.Info("Application shutdown")
}

// domReady is called after the front-end dom has been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue,
// false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Brique",
		Width:     1280,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "Brique",
				Message: "L'infrastructure de résilience pour la réparation",
			},
		},
		Linux: &linux.Options{
			Icon: []byte{}, // TODO: Add icon
		},
		LogLevel: logger.DEBUG,
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
