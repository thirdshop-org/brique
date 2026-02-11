package services_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
	"github.com/lhommenul/brique/core/services"
)

func setupTestService(t *testing.T) (*services.BackpackService, func()) {
	// Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	assetsDir := filepath.Join(tempDir, "assets")

	// Create assets directory
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create assets dir: %v", err)
	}

	// Initialize database
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	database, err := db.NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	queries := db.New(database.DB)
	service := services.NewBackpackService(queries, assetsDir)

	cleanup := func() {
		database.Close()
	}

	return service, cleanup
}

func TestCreateItem(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	item := &models.Item{
		Name:     "Lave-Linge",
		Category: "Gros Électroménager",
		Brand:    "Brandt",
		Model:    "WTC1234",
	}

	err := service.CreateItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if item.ID == 0 {
		t.Error("item ID should be set after creation")
	}

	if item.CreatedAt.IsZero() {
		t.Error("item CreatedAt should be set")
	}

	if item.UpdatedAt.IsZero() {
		t.Error("item UpdatedAt should be set")
	}
}

func TestGetAllItems(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a few items
	items := []models.Item{
		{Name: "Item 1", Category: "Cat1", Brand: "Brand1", Model: "M1"},
		{Name: "Item 2", Category: "Cat2", Brand: "Brand2", Model: "M2"},
		{Name: "Item 3", Category: "Cat3", Brand: "Brand3", Model: "M3"},
	}

	for i := range items {
		if err := service.CreateItem(ctx, &items[i]); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	// Get all items
	retrieved, err := service.GetAllItems(ctx)
	if err != nil {
		t.Fatalf("failed to get all items: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 items, got %d", len(retrieved))
	}
}

func TestSearchItems(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create items
	items := []models.Item{
		{Name: "Lave-Linge Brandt", Category: "Électroménager", Brand: "Brandt", Model: "WTC1234"},
		{Name: "Perceuse Bosch", Category: "Outils", Brand: "Bosch", Model: "PSB500"},
		{Name: "Lave-Vaisselle Whirlpool", Category: "Électroménager", Brand: "Whirlpool", Model: "WDL123"},
	}

	for i := range items {
		if err := service.CreateItem(ctx, &items[i]); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	// Search by brand
	results, err := service.SearchItems(ctx, "Bosch")
	if err != nil {
		t.Fatalf("failed to search items: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Bosch', got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "Perceuse Bosch" {
		t.Errorf("expected 'Perceuse Bosch', got '%s'", results[0].Name)
	}

	// Search by category
	results, err = service.SearchItems(ctx, "Électroménager")
	if err != nil {
		t.Fatalf("failed to search items: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Électroménager', got %d", len(results))
	}
}

func TestUpdateItem(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create item
	item := &models.Item{
		Name:     "Original Name",
		Category: "Original Category",
		Brand:    "Original Brand",
		Model:    "Original Model",
	}

	if err := service.CreateItem(ctx, item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	originalUpdatedAt := item.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update item
	item.Name = "Updated Name"
	item.Notes = "Some notes"

	if err := service.UpdateItem(ctx, item); err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	if item.UpdatedAt == originalUpdatedAt {
		t.Error("UpdatedAt should be different after update")
	}

	// Verify update
	retrieved, err := service.GetItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", retrieved.Name)
	}

	if retrieved.Notes != "Some notes" {
		t.Errorf("expected notes 'Some notes', got '%s'", retrieved.Notes)
	}
}

func TestAddAsset(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create item
	item := &models.Item{
		Name:     "Test Item",
		Category: "Test",
		Brand:    "Test",
		Model:    "Test",
	}

	if err := service.CreateItem(ctx, item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create a temporary test file
	tempFile := filepath.Join(t.TempDir(), "test_manual.pdf")
	testContent := []byte("This is a test PDF content")
	if err := os.WriteFile(tempFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add asset
	asset, err := service.AddAsset(ctx, item.ID, models.AssetTypeManual, "User Manual", tempFile)
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	if asset.ID == 0 {
		t.Error("asset ID should be set")
	}

	if asset.FileHash == "" {
		t.Error("file hash should be calculated")
	}

	if asset.FileSize != int64(len(testContent)) {
		t.Errorf("expected file size %d, got %d", len(testContent), asset.FileSize)
	}

	// Verify file was copied
	if _, err := os.Stat(asset.FilePath); os.IsNotExist(err) {
		t.Error("asset file should exist")
	}
}

func TestDocumentationHealth(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create item
	item := &models.Item{
		Name:     "Test Item",
		Category: "Test",
		Brand:    "Test",
		Model:    "Test",
	}

	if err := service.CreateItem(ctx, item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Initially should be incomplete
	itemWithAssets, err := service.GetItemWithAssets(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get item with assets: %v", err)
	}

	if itemWithAssets.Health != models.HealthIncomplete {
		t.Errorf("expected health 'incomplete', got '%s'", itemWithAssets.Health)
	}

	// Add a manual - should be partial
	tempFile1 := filepath.Join(t.TempDir(), "manual.pdf")
	if err := os.WriteFile(tempFile1, []byte("manual"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if _, err := service.AddAsset(ctx, item.ID, models.AssetTypeManual, "Manual", tempFile1); err != nil {
		t.Fatalf("failed to add manual: %v", err)
	}

	itemWithAssets, err = service.GetItemWithAssets(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get item with assets: %v", err)
	}

	if itemWithAssets.Health != models.HealthPartial {
		t.Errorf("expected health 'partial', got '%s'", itemWithAssets.Health)
	}

	// Add service manual - should be secured
	tempFile2 := filepath.Join(t.TempDir(), "service_manual.pdf")
	if err := os.WriteFile(tempFile2, []byte("service manual"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if _, err := service.AddAsset(ctx, item.ID, models.AssetTypeServiceManual, "Service Manual", tempFile2); err != nil {
		t.Fatalf("failed to add service manual: %v", err)
	}

	itemWithAssets, err = service.GetItemWithAssets(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to get item with assets: %v", err)
	}

	if itemWithAssets.Health != models.HealthSecured {
		t.Errorf("expected health 'secured', got '%s'", itemWithAssets.Health)
	}
}
