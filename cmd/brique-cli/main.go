package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
	"github.com/lhommenul/brique/core/services"
	"github.com/lhommenul/brique/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfg             *config.Config
	database        *db.Database
	backpackService *services.BackpackService
	logger          *slog.Logger
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

	itemGetCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get detailed information about an item",
		Args:  cobra.ExactArgs(1),
		RunE:  runItemGet,
	}

	itemUpdateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing item",
		Args:  cobra.ExactArgs(1),
		RunE:  runItemUpdate,
	}

	itemDeleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an item from the inventory",
		Args:  cobra.ExactArgs(1),
		RunE:  runItemDelete,
	}

	itemSearchCmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search items by name, brand, or category",
		Args:  cobra.ExactArgs(1),
		RunE:  runItemSearch,
	}

	itemCmd.AddCommand(itemAddCmd, itemListCmd, itemGetCmd, itemUpdateCmd, itemDeleteCmd, itemSearchCmd)

	// Asset commands
	assetCmd := &cobra.Command{
		Use:   "asset",
		Short: "Manage item assets (documents, files)",
	}

	assetAddCmd := &cobra.Command{
		Use:   "add <item-id> <file>",
		Short: "Add an asset file to an item",
		Args:  cobra.ExactArgs(2),
		RunE:  runAssetAdd,
	}
	assetAddCmd.Flags().StringP("type", "t", "manual", "Asset type (manual, service_manual, exploded_view, stl, firmware, driver, schematic, other)")
	assetAddCmd.Flags().StringP("name", "n", "", "Asset name (defaults to filename)")

	assetListCmd := &cobra.Command{
		Use:   "list <item-id>",
		Short: "List all assets for an item",
		Args:  cobra.ExactArgs(1),
		RunE:  runAssetList,
	}

	assetDeleteCmd := &cobra.Command{
		Use:   "delete <asset-id>",
		Short: "Delete an asset",
		Args:  cobra.ExactArgs(1),
		RunE:  runAssetDelete,
	}

	assetCmd.AddCommand(assetAddCmd, assetListCmd, assetDeleteCmd)

	rootCmd.AddCommand(itemCmd, assetCmd)

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

// Item commands implementation

func runItemAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Add New Item ===\n")

	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Category: ")
	category, _ := reader.ReadString('\n')
	category = strings.TrimSpace(category)

	fmt.Print("Brand: ")
	brand, _ := reader.ReadString('\n')
	brand = strings.TrimSpace(brand)

	fmt.Print("Model: ")
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)

	fmt.Print("Serial Number (optional): ")
	serialNumber, _ := reader.ReadString('\n')
	serialNumber = strings.TrimSpace(serialNumber)

	fmt.Print("Notes (optional): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	item := &models.Item{
		Name:         name,
		Category:     category,
		Brand:        brand,
		Model:        model,
		SerialNumber: serialNumber,
		Notes:        notes,
	}

	if err := backpackService.CreateItem(ctx, item); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	fmt.Printf("\n✓ Item created successfully with ID: %d\n", item.ID)

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

	fmt.Printf("\n=== Inventory (%d items) ===\n\n", len(items))

	for _, item := range items {
		fmt.Printf("ID: %d | %s\n", item.ID, item.Name)
		fmt.Printf("  Category: %s | Brand: %s | Model: %s\n", item.Category, item.Brand, item.Model)
		if item.SerialNumber != "" {
			fmt.Printf("  Serial: %s\n", item.SerialNumber)
		}
		fmt.Println()
	}

	return nil
}

func runItemGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	itemWithAssets, err := backpackService.GetItemWithAssets(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	item := itemWithAssets.Item

	fmt.Printf("\n=== Item #%d ===\n\n", item.ID)
	fmt.Printf("Name:         %s\n", item.Name)
	fmt.Printf("Category:     %s\n", item.Category)
	fmt.Printf("Brand:        %s\n", item.Brand)
	fmt.Printf("Model:        %s\n", item.Model)
	if item.SerialNumber != "" {
		fmt.Printf("Serial:       %s\n", item.SerialNumber)
	}
	if item.PurchaseDate != nil {
		fmt.Printf("Purchase:     %s\n", item.PurchaseDate.Format("2006-01-02"))
	}
	if item.Notes != "" {
		fmt.Printf("Notes:        %s\n", item.Notes)
	}
	fmt.Printf("Created:      %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:      %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Display health and assets
	fmt.Printf("\nDocumentation Health: %s\n", getHealthEmoji(itemWithAssets.Health))

	if len(itemWithAssets.Assets) > 0 {
		fmt.Printf("\nAssets (%d files):\n", len(itemWithAssets.Assets))
		for _, asset := range itemWithAssets.Assets {
			fmt.Printf("  [%d] %s (%s) - %s\n", asset.ID, asset.Name, asset.Type, formatFileSize(asset.FileSize))
		}
	} else {
		fmt.Println("\nNo assets attached.")
	}

	fmt.Println()

	return nil
}

func runItemUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Get existing item
	item, err := backpackService.GetItem(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	fmt.Printf("\n=== Update Item #%d ===\n\n", id)
	fmt.Println("Press Enter to keep current value, or type new value:")
	fmt.Println()

	fmt.Printf("Name [%s]: ", item.Name)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.Name = strings.TrimSpace(input)
	}

	fmt.Printf("Category [%s]: ", item.Category)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.Category = strings.TrimSpace(input)
	}

	fmt.Printf("Brand [%s]: ", item.Brand)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.Brand = strings.TrimSpace(input)
	}

	fmt.Printf("Model [%s]: ", item.Model)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.Model = strings.TrimSpace(input)
	}

	fmt.Printf("Serial Number [%s]: ", item.SerialNumber)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.SerialNumber = strings.TrimSpace(input)
	}

	fmt.Printf("Notes [%s]: ", item.Notes)
	if input, _ := reader.ReadString('\n'); strings.TrimSpace(input) != "" {
		item.Notes = strings.TrimSpace(input)
	}

	if err := backpackService.UpdateItem(ctx, item); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	fmt.Printf("\n✓ Item #%d updated successfully\n", id)

	return nil
}

func runItemDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Get item to show what will be deleted
	item, err := backpackService.GetItem(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	fmt.Printf("\n=== Delete Item #%d ===\n\n", id)
	fmt.Printf("Name: %s\n", item.Name)
	fmt.Printf("Brand: %s %s\n", item.Brand, item.Model)
	fmt.Println()
	fmt.Print("Are you sure you want to delete this item? (yes/no): ")

	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.ToLower(strings.TrimSpace(confirmation))

	if confirmation != "yes" && confirmation != "y" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := backpackService.DeleteItem(ctx, id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	fmt.Printf("\n✓ Item #%d deleted successfully\n", id)

	return nil
}

func runItemSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	query := args[0]

	items, err := backpackService.SearchItems(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search items: %w", err)
	}

	if len(items) == 0 {
		fmt.Printf("\nNo items found matching '%s'.\n", query)
		return nil
	}

	fmt.Printf("\n=== Search Results for '%s' (%d items) ===\n\n", query, len(items))

	for _, item := range items {
		fmt.Printf("ID: %d | %s\n", item.ID, item.Name)
		fmt.Printf("  Category: %s | Brand: %s | Model: %s\n", item.Category, item.Brand, item.Model)
		if item.SerialNumber != "" {
			fmt.Printf("  Serial: %s\n", item.SerialNumber)
		}
		fmt.Println()
	}

	return nil
}

// Asset commands implementation

func runAssetAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	itemID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	filePath := args[1]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get flags
	assetType, _ := cmd.Flags().GetString("type")
	assetName, _ := cmd.Flags().GetString("name")

	// Default name to filename if not provided
	if assetName == "" {
		assetName = filepath.Base(filePath)
	}

	// Validate asset type
	validTypes := map[string]bool{
		"manual":         true,
		"service_manual": true,
		"exploded_view":  true,
		"stl":            true,
		"firmware":       true,
		"driver":         true,
		"schematic":      true,
		"other":          true,
	}

	if !validTypes[assetType] {
		return fmt.Errorf("invalid asset type: %s", assetType)
	}

	fmt.Printf("\nAdding asset to item #%d...\n", itemID)

	asset, err := backpackService.AddAsset(ctx, itemID, models.AssetType(assetType), assetName, filePath)
	if err != nil {
		return fmt.Errorf("failed to add asset: %w", err)
	}

	fmt.Printf("\n✓ Asset added successfully\n")
	fmt.Printf("  ID: %d\n", asset.ID)
	fmt.Printf("  Name: %s\n", asset.Name)
	fmt.Printf("  Type: %s\n", asset.Type)
	fmt.Printf("  Size: %s\n", formatFileSize(asset.FileSize))
	fmt.Printf("  Hash: %s\n", asset.FileHash[:16]+"...")

	return nil
}

func runAssetList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	itemID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Get item info
	item, err := backpackService.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Get assets
	assets, err := backpackService.GetItemAssets(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	fmt.Printf("\n=== Assets for Item #%d: %s ===\n\n", itemID, item.Name)

	if len(assets) == 0 {
		fmt.Println("No assets attached to this item.")
		return nil
	}

	fmt.Printf("Total: %d file(s)\n\n", len(assets))

	totalSize := int64(0)
	for _, asset := range assets {
		fmt.Printf("ID: %d\n", asset.ID)
		fmt.Printf("  Name: %s\n", asset.Name)
		fmt.Printf("  Type: %s\n", asset.Type)
		fmt.Printf("  Size: %s\n", formatFileSize(asset.FileSize))
		fmt.Printf("  Path: %s\n", asset.FilePath)
		fmt.Printf("  Hash: %s\n", asset.FileHash[:16]+"...")
		fmt.Printf("  Added: %s\n", asset.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()

		totalSize += asset.FileSize
	}

	fmt.Printf("Total size: %s\n", formatFileSize(totalSize))

	return nil
}

func runAssetDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	assetID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid asset ID: %w", err)
	}

	fmt.Printf("\n=== Delete Asset #%d ===\n\n", assetID)
	fmt.Print("Are you sure you want to delete this asset? (yes/no): ")

	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.ToLower(strings.TrimSpace(confirmation))

	if confirmation != "yes" && confirmation != "y" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := backpackService.DeleteAsset(ctx, assetID); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	fmt.Printf("\n✓ Asset #%d deleted successfully\n", assetID)

	return nil
}

// Helper functions

func getHealthEmoji(health models.DocumentationHealth) string {
	switch health {
	case models.HealthSecured:
		return "🟢 Secured (Complete documentation)"
	case models.HealthPartial:
		return "🟡 Partial (Some documentation)"
	case models.HealthIncomplete:
		return "🔴 Incomplete (No documentation)"
	default:
		return "❓ Unknown"
	}
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
