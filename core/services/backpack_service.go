package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
)

// BackpackService manages the inventory (Sac à Dos)
type BackpackService struct {
	queries   *db.Queries
	assetsDir string
}

// NewBackpackService creates a new backpack service
func NewBackpackService(queries *db.Queries, assetsDir string) *BackpackService {
	return &BackpackService{
		queries:   queries,
		assetsDir: assetsDir,
	}
}

// CreateItem creates a new item in the inventory
func (s *BackpackService) CreateItem(ctx context.Context, item *models.Item) error {
	now := time.Now()

	params := db.CreateItemParams{
		Name:         item.Name,
		Category:     item.Category,
		Brand:        item.Brand,
		Model:        item.Model,
		SerialNumber: item.SerialNumber,
		PhotoPath:    item.PhotoPath,
		Notes:        item.Notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if item.PurchaseDate != nil {
		params.PurchaseDate.Time = *item.PurchaseDate
		params.PurchaseDate.Valid = true
	}

	created, err := s.queries.CreateItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	item.ID = created.ID
	item.CreatedAt = created.CreatedAt
	item.UpdatedAt = created.UpdatedAt

	return nil
}

// GetItem retrieves an item by ID
func (s *BackpackService) GetItem(ctx context.Context, id int64) (*models.Item, error) {
	dbItem, err := s.queries.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return s.dbItemToModel(dbItem), nil
}

// GetAllItems retrieves all items
func (s *BackpackService) GetAllItems(ctx context.Context) ([]models.Item, error) {
	dbItems, err := s.queries.GetAllItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}

	items := make([]models.Item, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = *s.dbItemToModel(dbItem)
	}

	return items, nil
}

// UpdateItem updates an existing item
func (s *BackpackService) UpdateItem(ctx context.Context, item *models.Item) error {
	now := time.Now()

	params := db.UpdateItemParams{
		Name:         item.Name,
		Category:     item.Category,
		Brand:        item.Brand,
		Model:        item.Model,
		SerialNumber: item.SerialNumber,
		PhotoPath:    item.PhotoPath,
		Notes:        item.Notes,
		UpdatedAt:    now,
		ID:           item.ID,
	}

	if item.PurchaseDate != nil {
		params.PurchaseDate.Time = *item.PurchaseDate
		params.PurchaseDate.Valid = true
	}

	if err := s.queries.UpdateItem(ctx, params); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	item.UpdatedAt = now

	return nil
}

// DeleteItem deletes an item and all its assets
func (s *BackpackService) DeleteItem(ctx context.Context, id int64) error {
	// Get all assets for this item to delete files
	assets, err := s.queries.GetAssetsByItemID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	// Delete asset files from disk
	for _, asset := range assets {
		if err := os.Remove(asset.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but continue
			fmt.Printf("Warning: failed to delete asset file %s: %v\n", asset.FilePath, err)
		}
	}

	// Delete the item (assets will be deleted by CASCADE)
	if err := s.queries.DeleteItem(ctx, id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// SearchItems searches for items by query
func (s *BackpackService) SearchItems(ctx context.Context, query string) ([]models.Item, error) {
	searchTerm := "%" + query + "%"

	dbItems, err := s.queries.SearchItems(ctx, db.SearchItemsParams{
		Name:     searchTerm,
		Brand:    searchTerm,
		Category: searchTerm,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}

	items := make([]models.Item, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = *s.dbItemToModel(dbItem)
	}

	return items, nil
}

// AddAsset adds an asset to an item by copying the file to the assets directory
func (s *BackpackService) AddAsset(ctx context.Context, itemID int64, assetType models.AssetType, name string, sourcePath string) (*models.Asset, error) {
	// Verify item exists
	if _, err := s.queries.GetItemByID(ctx, itemID); err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get file info
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, sourceFile); err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Reset file pointer
	if _, err := sourceFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Create item directory in assets
	itemDir := filepath.Join(s.assetsDir, fmt.Sprintf("item_%d", itemID))
	if err := os.MkdirAll(itemDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create item directory: %w", err)
	}

	// Generate destination path
	ext := filepath.Ext(sourcePath)
	destPath := filepath.Join(itemDir, fmt.Sprintf("%s_%d%s", assetType, time.Now().Unix(), ext))

	// Copy file to assets directory
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Create asset in database
	now := time.Now()
	params := db.CreateAssetParams{
		ItemID:    itemID,
		Type:      string(assetType),
		Name:      name,
		FilePath:  destPath,
		FileSize:  fileInfo.Size(),
		FileHash:  fileHash,
		CreatedAt: now,
	}

	dbAsset, err := s.queries.CreateAsset(ctx, params)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return s.dbAssetToModel(dbAsset), nil
}

// GetItemAssets retrieves all assets for an item
func (s *BackpackService) GetItemAssets(ctx context.Context, itemID int64) ([]models.Asset, error) {
	dbAssets, err := s.queries.GetAssetsByItemID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	assets := make([]models.Asset, len(dbAssets))
	for i, dbAsset := range dbAssets {
		assets[i] = *s.dbAssetToModel(dbAsset)
	}

	return assets, nil
}

// GetItemWithAssets retrieves an item with all its assets and health status
func (s *BackpackService) GetItemWithAssets(ctx context.Context, itemID int64) (*models.ItemWithAssets, error) {
	item, err := s.GetItem(ctx, itemID)
	if err != nil {
		return nil, err
	}

	assets, err := s.GetItemAssets(ctx, itemID)
	if err != nil {
		return nil, err
	}

	health := s.calculateDocumentationHealth(assets)

	return &models.ItemWithAssets{
		Item:   *item,
		Assets: assets,
		Health: health,
	}, nil
}

// DeleteAsset deletes an asset and its file
func (s *BackpackService) DeleteAsset(ctx context.Context, assetID int64) error {
	// Get asset to get file path
	dbAsset, err := s.queries.GetAssetByID(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %w", err)
	}

	// Delete file from disk
	if err := os.Remove(dbAsset.FilePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to delete asset file %s: %v\n", dbAsset.FilePath, err)
	}

	// Delete from database
	if err := s.queries.DeleteAsset(ctx, assetID); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}

// calculateDocumentationHealth determines the health status based on assets
func (s *BackpackService) calculateDocumentationHealth(assets []models.Asset) models.DocumentationHealth {
	if len(assets) == 0 {
		return models.HealthIncomplete
	}

	hasManual := false
	hasServiceManual := false

	for _, asset := range assets {
		switch asset.Type {
		case models.AssetTypeManual:
			hasManual = true
		case models.AssetTypeServiceManual:
			hasServiceManual = true
		}
	}

	if hasManual && hasServiceManual {
		return models.HealthSecured
	}

	if hasManual || len(assets) > 0 {
		return models.HealthPartial
	}

	return models.HealthIncomplete
}

// dbItemToModel converts a DB item to a model item
func (s *BackpackService) dbItemToModel(dbItem db.Item) *models.Item {
	item := &models.Item{
		ID:           dbItem.ID,
		Name:         dbItem.Name,
		Category:     dbItem.Category,
		Brand:        dbItem.Brand,
		Model:        dbItem.Model,
		SerialNumber: dbItem.SerialNumber,
		PhotoPath:    dbItem.PhotoPath,
		Notes:        dbItem.Notes,
		CreatedAt:    dbItem.CreatedAt,
		UpdatedAt:    dbItem.UpdatedAt,
	}

	if dbItem.PurchaseDate.Valid {
		item.PurchaseDate = &dbItem.PurchaseDate.Time
	}

	return item
}

// dbAssetToModel converts a DB asset to a model asset
func (s *BackpackService) dbAssetToModel(dbAsset db.Asset) *models.Asset {
	return &models.Asset{
		ID:        dbAsset.ID,
		ItemID:    dbAsset.ItemID,
		Type:      models.AssetType(dbAsset.Type),
		Name:      dbAsset.Name,
		FilePath:  dbAsset.FilePath,
		FileSize:  dbAsset.FileSize,
		FileHash:  dbAsset.FileHash,
		CreatedAt: dbAsset.CreatedAt,
	}
}
