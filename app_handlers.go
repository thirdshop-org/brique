package main

import (
	"fmt"

	"github.com/lhommenul/brique/core/models"
)

// ItemDTO is the Data Transfer Object for items
type ItemDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Brand        string  `json:"brand"`
	Model        string  `json:"model"`
	SerialNumber string  `json:"serialNumber"`
	PurchaseDate *string `json:"purchaseDate"`
	PhotoPath    string  `json:"photoPath"`
	Notes        string  `json:"notes"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// AssetDTO is the Data Transfer Object for assets
type AssetDTO struct {
	ID        int64  `json:"id"`
	ItemID    int64  `json:"itemId"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	FilePath  string `json:"filePath"`
	FileSize  int64  `json:"fileSize"`
	FileHash  string `json:"fileHash"`
	CreatedAt string `json:"createdAt"`
}

// ItemWithAssetsDTO is the Data Transfer Object for items with assets
type ItemWithAssetsDTO struct {
	Item   ItemDTO    `json:"item"`
	Assets []AssetDTO `json:"assets"`
	Health string     `json:"health"`
}

// GetAllItems returns all items in the inventory
func (a *App) GetAllItems() ([]ItemDTO, error) {
	items, err := a.backpackService.GetAllItems(a.ctx)
	if err != nil {
		a.events.Error("Erreur de chargement", "Impossible de charger les items")
		return nil, err
	}

	dtos := make([]ItemDTO, len(items))
	for i, item := range items {
		dtos[i] = itemToDTO(&item)
	}

	return dtos, nil
}

// GetItem returns a single item by ID
func (a *App) GetItem(id int64) (*ItemDTO, error) {
	item, err := a.backpackService.GetItem(a.ctx, id)
	if err != nil {
		return nil, err
	}

	dto := itemToDTO(item)
	return &dto, nil
}

// GetItemWithAssets returns an item with all its assets and health status
func (a *App) GetItemWithAssets(id int64) (*ItemWithAssetsDTO, error) {
	itemWithAssets, err := a.backpackService.GetItemWithAssets(a.ctx, id)
	if err != nil {
		return nil, err
	}

	itemDTO := itemToDTO(&itemWithAssets.Item)

	assetDTOs := make([]AssetDTO, len(itemWithAssets.Assets))
	for i, asset := range itemWithAssets.Assets {
		assetDTOs[i] = assetToDTO(&asset)
	}

	return &ItemWithAssetsDTO{
		Item:   itemDTO,
		Assets: assetDTOs,
		Health: string(itemWithAssets.Health),
	}, nil
}

// CreateItem creates a new item
func (a *App) CreateItem(name, category, brand, model, serialNumber, notes string) (*ItemDTO, error) {
	item := &models.Item{
		Name:         name,
		Category:     category,
		Brand:        brand,
		Model:        model,
		SerialNumber: serialNumber,
		Notes:        notes,
	}

	if err := a.backpackService.CreateItem(a.ctx, item); err != nil {
		a.events.Error("Erreur de création", fmt.Sprintf("Impossible de créer l'item '%s'", name))
		return nil, err
	}

	a.events.Success("Item créé", fmt.Sprintf("'%s' a été ajouté à l'inventaire", name))
	dto := itemToDTO(item)
	return &dto, nil
}

// UpdateItem updates an existing item
func (a *App) UpdateItem(id int64, name, category, brand, model, serialNumber, notes string) error {
	item := &models.Item{
		ID:           id,
		Name:         name,
		Category:     category,
		Brand:        brand,
		Model:        model,
		SerialNumber: serialNumber,
		Notes:        notes,
	}

	if err := a.backpackService.UpdateItem(a.ctx, item); err != nil {
		a.events.Error("Erreur de mise à jour", fmt.Sprintf("Impossible de mettre à jour l'item #%d", id))
		return err
	}

	a.events.Success("Item mis à jour", fmt.Sprintf("'%s' a été modifié", name))
	return nil
}

// DeleteItem deletes an item
func (a *App) DeleteItem(id int64) error {
	// Get item name before deletion for notification
	item, err := a.backpackService.GetItem(a.ctx, id)
	if err != nil {
		a.events.Error("Erreur de suppression", "Item introuvable")
		return err
	}

	if err := a.backpackService.DeleteItem(a.ctx, id); err != nil {
		a.events.Error("Erreur de suppression", fmt.Sprintf("Impossible de supprimer l'item #%d", id))
		return err
	}

	a.events.Success("Item supprimé", fmt.Sprintf("'%s' a été supprimé de l'inventaire", item.Name))
	return nil
}

// SearchItems searches for items
func (a *App) SearchItems(query string) ([]ItemDTO, error) {
	items, err := a.backpackService.SearchItems(a.ctx, query)
	if err != nil {
		return nil, err
	}

	dtos := make([]ItemDTO, len(items))
	for i, item := range items {
		dtos[i] = itemToDTO(&item)
	}

	return dtos, nil
}

// GetAssets returns all assets for an item
func (a *App) GetAssets(itemID int64) ([]AssetDTO, error) {
	assets, err := a.backpackService.GetItemAssets(a.ctx, itemID)
	if err != nil {
		return nil, err
	}

	dtos := make([]AssetDTO, len(assets))
	for i, asset := range assets {
		dtos[i] = assetToDTO(&asset)
	}

	return dtos, nil
}

// AddAsset adds an asset to an item
func (a *App) AddAsset(itemID int64, assetType, name, sourcePath string) (*AssetDTO, error) {
	// Emit progress start
	progressID := fmt.Sprintf("asset-upload-%d", itemID)
	a.events.EmitProgress(ProgressData{
		ID:        progressID,
		Operation: "Ajout de fichier",
		Current:   0,
		Total:     100,
		Filename:  name,
	})

	asset, err := a.backpackService.AddAsset(a.ctx, itemID, models.AssetType(assetType), name, sourcePath)

	// Complete progress
	a.events.EmitProgressComplete(progressID)

	if err != nil {
		a.events.Error("Erreur d'ajout", fmt.Sprintf("Impossible d'ajouter le fichier '%s'", name))
		return nil, err
	}

	a.events.Success("Fichier ajouté", fmt.Sprintf("'%s' a été ajouté à l'item", name))
	dto := assetToDTO(asset)
	return &dto, nil
}

// DeleteAsset deletes an asset
func (a *App) DeleteAsset(assetID int64) error {
	if err := a.backpackService.DeleteAsset(a.ctx, assetID); err != nil {
		a.events.Error("Erreur de suppression", "Impossible de supprimer le fichier")
		return err
	}

	a.events.Success("Fichier supprimé", "Le fichier a été supprimé")
	return nil
}

// Helper functions to convert models to DTOs

func itemToDTO(item *models.Item) ItemDTO {
	dto := ItemDTO{
		ID:           item.ID,
		Name:         item.Name,
		Category:     item.Category,
		Brand:        item.Brand,
		Model:        item.Model,
		SerialNumber: item.SerialNumber,
		PhotoPath:    item.PhotoPath,
		Notes:        item.Notes,
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if item.PurchaseDate != nil {
		dateStr := item.PurchaseDate.Format("2006-01-02")
		dto.PurchaseDate = &dateStr
	}

	return dto
}

func assetToDTO(asset *models.Asset) AssetDTO {
	return AssetDTO{
		ID:        asset.ID,
		ItemID:    asset.ItemID,
		Type:      string(asset.Type),
		Name:      asset.Name,
		FilePath:  asset.FilePath,
		FileSize:  asset.FileSize,
		FileHash:  asset.FileHash,
		CreatedAt: asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
