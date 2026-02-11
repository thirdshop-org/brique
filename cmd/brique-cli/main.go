package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
	"github.com/lhommenul/brique/core/services"
	"github.com/lhommenul/brique/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
	database *db.Database
	backpackService *services.BackpackService
	logger *slog.Logger
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "brique",
		Short: "Brique - L'infrastructure de résilience pour la réparation",
		Long: `Brique est une application offline-first pour gérer votre inventaire
d'objets et leurs documentations de réparation.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initApp()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeApp()
		},
	}

	// Item commands
	itemCmd := &cobra.Command{
		Use:   "item",
		Short: "Manage inventory items",
	}

	itemAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new item to the inventory",
		RunE:  runItemAdd,
	}

	itemListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all items in the inventory",
		RunE:  runItemList,
	}

	itemCmd.AddCommand(itemAddCmd, itemListCmd)
	rootCmd.AddCommand(itemCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initApp() error {
	// Setup logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Info("Configuration loaded", "data_dir", cfg.DataDir)

	// Initialize database
	database, err = db.NewDatabase(cfg.DatabasePath, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create queries
	queries := db.New(database.DB)

	// Create backpack service
	backpackService = services.NewBackpackService(queries, cfg.AssetsDir)

	logger.Info("Application initialized successfully")

	return nil
}

func closeApp() error {
	if database != nil {
		return database.Close()
	}
	return nil
}

func runItemAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var name, category, brand, model string

	fmt.Print("Name: ")
	fmt.Scanln(&name)

	fmt.Print("Category: ")
	fmt.Scanln(&category)

	fmt.Print("Brand: ")
	fmt.Scanln(&brand)

	fmt.Print("Model: ")
	fmt.Scanln(&model)

	item := &models.Item{
		Name:     name,
		Category: category,
		Brand:    brand,
		Model:    model,
	}

	if err := backpackService.CreateItem(ctx, item); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	fmt.Printf("\nItem created successfully with ID: %d\n", item.ID)

	return nil
}

func runItemList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	items, err := backpackService.GetAllItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to get items: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No items in inventory.")
		return nil
	}

	fmt.Printf("\nInventory (%d items):\n\n", len(items))

	for _, item := range items {
		fmt.Printf("ID: %d\n", item.ID)
		fmt.Printf("  Name: %s\n", item.Name)
		fmt.Printf("  Category: %s\n", item.Category)
		fmt.Printf("  Brand: %s\n", item.Brand)
		fmt.Printf("  Model: %s\n", item.Model)
		if item.SerialNumber != "" {
			fmt.Printf("  Serial: %s\n", item.SerialNumber)
		}
		fmt.Println()
	}

	return nil
}
