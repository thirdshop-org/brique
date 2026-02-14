package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lhommenul/brique/core/models"
	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

// QRCodeData represents the data structure embedded in the QR code
type QRCodeData struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	Type         string `json:"type"` // "brique-item"
}

// GenerateQRCode generates a QR code for an item and returns it as base64 PNG
func (a *App) GenerateQRCode(itemID int64) (string, error) {
	// Get item details
	item, err := a.backpackService.GetItem(a.ctx, itemID)
	if err != nil {
		a.events.Error("Erreur QR Code", "Impossible de charger l'item")
		return "", err
	}

	// Create QR code data
	qrData := QRCodeData{
		ID:           item.ID,
		Name:         item.Name,
		Brand:        item.Brand,
		Model:        item.Model,
		SerialNumber: item.SerialNumber,
		Type:         "brique-item",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal QR data: %w", err)
	}

	// Generate QR code (256x256 pixels, medium error correction)
	png, err := qrcode.Encode(string(jsonData), qrcode.Medium, 256)
	if err != nil {
		a.events.Error("Erreur QR Code", "Impossible de générer le QR code")
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Encode to base64
	base64Str := base64.StdEncoding.EncodeToString(png)

	a.events.Success("QR Code généré", fmt.Sprintf("QR Code pour '%s' créé", item.Name))
	return base64Str, nil
}

// ExportData represents the full export structure
type ExportData struct {
	ExportDate string                  `json:"exportDate"`
	Version    string                  `json:"version"`
	Items      []ItemWithAssetsDTO     `json:"items"`
	Stats      map[string]interface{}  `json:"stats"`
}

// ExportToJSON exports all inventory data to a JSON file
func (a *App) ExportToJSON() error {
	// Get all items with their assets
	items, err := a.backpackService.GetAllItems(a.ctx)
	if err != nil {
		a.events.Error("Erreur d'export", "Impossible de charger les items")
		return err
	}

	exportItems := make([]ItemWithAssetsDTO, 0, len(items))
	for _, item := range items {
		itemWithAssets, err := a.backpackService.GetItemWithAssets(a.ctx, item.ID)
		if err != nil {
			continue
		}

		itemDTO := itemToDTO(&itemWithAssets.Item)
		assetDTOs := make([]AssetDTO, len(itemWithAssets.Assets))
		for i, asset := range itemWithAssets.Assets {
			assetDTOs[i] = assetToDTO(&asset)
		}

		exportItems = append(exportItems, ItemWithAssetsDTO{
			Item:   itemDTO,
			Assets: assetDTOs,
			Health: string(itemWithAssets.Health),
		})
	}

	// Create export data
	exportData := ExportData{
		ExportDate: time.Now().Format(time.RFC3339),
		Version:    "1.0",
		Items:      exportItems,
		Stats: map[string]interface{}{
			"totalItems": len(items),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		a.events.Error("Erreur d'export", "Impossible de générer le JSON")
		return err
	}

	// Show save dialog
	defaultFilename := fmt.Sprintf("brique-export-%s.json", time.Now().Format("2006-01-02"))
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Exporter l'inventaire",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil || savePath == "" {
		return fmt.Errorf("export cancelled")
	}

	// Write to file
	if err := os.WriteFile(savePath, jsonData, 0644); err != nil {
		a.events.Error("Erreur d'export", "Impossible d'écrire le fichier")
		return err
	}

	a.events.Success("Export réussi", fmt.Sprintf("Inventaire exporté: %d items", len(items)))
	return nil
}

// ImportFromJSON imports inventory data from a JSON file
func (a *App) ImportFromJSON() error {
	// Show open file dialog
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Importer un inventaire",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil || filepath == "" {
		return fmt.Errorf("import cancelled")
	}

	// Read file
	jsonData, err := os.ReadFile(filepath)
	if err != nil {
		a.events.Error("Erreur d'import", "Impossible de lire le fichier")
		return err
	}

	// Parse JSON
	var exportData ExportData
	if err := json.Unmarshal(jsonData, &exportData); err != nil {
		a.events.Error("Erreur d'import", "Format JSON invalide")
		return err
	}

	// Import items
	imported := 0
	skipped := 0
	for _, itemDTO := range exportData.Items {
		// Check if item already exists (by brand+model+serial)
		existing, _ := a.backpackService.SearchItems(a.ctx, itemDTO.Item.SerialNumber)
		if len(existing) > 0 {
			// Skip if serial number already exists
			skipped++
			continue
		}

		// Create item
		item := &models.Item{
			Name:         itemDTO.Item.Name,
			Category:     itemDTO.Item.Category,
			Brand:        itemDTO.Item.Brand,
			Model:        itemDTO.Item.Model,
			SerialNumber: itemDTO.Item.SerialNumber,
			PhotoPath:    itemDTO.Item.PhotoPath,
			Notes:        itemDTO.Item.Notes,
		}

		if err := a.backpackService.CreateItem(a.ctx, item); err != nil {
			skipped++
			continue
		}

		imported++

		// Note: Assets are not imported because they reference file paths
		// that may not exist on this machine
	}

	message := fmt.Sprintf("Import réussi: %d items importés, %d ignorés", imported, skipped)
	if imported > 0 {
		a.events.Success("Import réussi", message)
	} else {
		a.events.Warning("Import partiel", message)
	}

	return nil
}

// ExportToCSV exports inventory data to a CSV file
func (a *App) ExportToCSV() error {
	// Get all items
	items, err := a.backpackService.GetAllItems(a.ctx)
	if err != nil {
		a.events.Error("Erreur d'export", "Impossible de charger les items")
		return err
	}

	// Show save dialog
	defaultFilename := fmt.Sprintf("brique-export-%s.csv", time.Now().Format("2006-01-02"))
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Exporter l'inventaire",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil || savePath == "" {
		return fmt.Errorf("export cancelled")
	}

	// Create file
	file, err := os.Create(savePath)
	if err != nil {
		a.events.Error("Erreur d'export", "Impossible de créer le fichier")
		return err
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Nom", "Catégorie", "Marque", "Modèle", "Numéro de série", "Date d'achat", "Notes", "Date de création"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, item := range items {
		purchaseDate := ""
		if item.PurchaseDate != nil {
			purchaseDate = item.PurchaseDate.Format("2006-01-02")
		}

		row := []string{
			fmt.Sprintf("%d", item.ID),
			item.Name,
			item.Category,
			item.Brand,
			item.Model,
			item.SerialNumber,
			purchaseDate,
			item.Notes,
			item.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	a.events.Success("Export réussi", fmt.Sprintf("Inventaire exporté: %d items", len(items)))
	return nil
}

// PeerDTO is the Data Transfer Object for peers
type PeerDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	LastSeen  string `json:"lastSeen"`
	LastSync  string `json:"lastSync"`
	IsTrusted bool   `json:"isTrusted"`
	Status    string `json:"status"`
}

// SyncResultDTO is the Data Transfer Object for sync results
type SyncResultDTO struct {
	ItemsReceived int   `json:"itemsReceived"`
	ItemsSent     int   `json:"itemsSent"`
	Conflicts     int   `json:"conflicts"`
	DurationMs    int64 `json:"durationMs"`
}

// SyncLogDTO is the Data Transfer Object for sync logs
type SyncLogDTO struct {
	ID            int64  `json:"id"`
	PeerName      string `json:"peerName"`
	Timestamp     string `json:"timestamp"`
	ItemsReceived int    `json:"itemsReceived"`
	ItemsSent     int    `json:"itemsSent"`
	Conflicts     int    `json:"conflicts"`
	DurationMs    int64  `json:"durationMs"`
	Error         string `json:"error"`
}

// GetPeers returns all discovered peers
func (a *App) GetPeers() ([]PeerDTO, error) {
	peers, err := a.gossipService.GetPeers(a.ctx)
	if err != nil {
		a.events.Error("Erreur", "Impossible de charger les pairs")
		return nil, err
	}

	dtos := make([]PeerDTO, len(peers))
	for i, peer := range peers {
		dtos[i] = peerToDTO(&peer)
	}

	return dtos, nil
}

// SyncWithPeer synchronizes with a specific peer
func (a *App) SyncWithPeer(peerID string) (*SyncResultDTO, error) {
	result, err := a.SyncWithPeerHTTP(peerID)
	if err != nil {
		return nil, err
	}

	return &SyncResultDTO{
		ItemsReceived: result.ItemsReceived,
		ItemsSent:     result.ItemsSent,
		Conflicts:     result.Conflicts,
		DurationMs:    result.DurationMs,
	}, nil
}

// SetPeerTrusted sets whether a peer is trusted
func (a *App) SetPeerTrusted(peerID string, trusted bool) error {
	if err := a.gossipService.SetPeerTrust(a.ctx, peerID, trusted); err != nil {
		a.events.Error("Erreur", "Impossible de modifier la confiance du pair")
		return err
	}

	action := "révoqué"
	if trusted {
		action = "approuvé"
	}
	a.events.Success("Pair " + action, fmt.Sprintf("Le pair a été %s", action))

	return nil
}

// AddPeer adds a peer manually
func (a *App) AddPeer(name, address string, trusted bool) error {
	// Generate a unique ID for the peer
	peerID := fmt.Sprintf("%s.manual", address)

	peer := &models.Peer{
		ID:        peerID,
		Name:      name,
		Address:   address,
		IsTrusted: trusted,
		LastSeen:  time.Now(),
	}

	if err := a.gossipService.AddPeer(a.ctx, peer); err != nil {
		a.events.Error("Erreur", "Impossible d'ajouter le pair")
		return err
	}

	a.events.Success("Pair ajouté", fmt.Sprintf("%s a été ajouté à la liste", name))
	return nil
}

// RemovePeer removes a peer from the list
func (a *App) RemovePeer(peerID string) error {
	if err := a.gossipService.RemovePeer(a.ctx, peerID); err != nil {
		a.events.Error("Erreur", "Impossible de supprimer le pair")
		return err
	}

	a.events.Success("Pair supprimé", "Le pair a été retiré de la liste")
	return nil
}

// GetSyncHistory returns recent synchronization history
func (a *App) GetSyncHistory(limit int) ([]SyncLogDTO, error) {
	logs, err := a.gossipService.GetRecentSyncHistory(a.ctx, limit)
	if err != nil {
		a.events.Error("Erreur", "Impossible de charger l'historique")
		return nil, err
	}

	// Get peer names
	peers, _ := a.gossipService.GetPeers(a.ctx)
	peerNames := make(map[string]string)
	for _, peer := range peers {
		peerNames[peer.ID] = peer.Name
	}

	dtos := make([]SyncLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = SyncLogDTO{
			ID:            log.ID,
			PeerName:      peerNames[log.PeerID],
			Timestamp:     log.Timestamp.Format("2006-01-02 15:04:05"),
			ItemsReceived: log.ItemsReceived,
			ItemsSent:     log.ItemsSent,
			Conflicts:     log.Conflicts,
			DurationMs:    log.DurationMs,
			Error:         log.Error,
		}
	}

	return dtos, nil
}

// Helper function to convert peer to DTO
func peerToDTO(peer *models.Peer) PeerDTO {
	dto := PeerDTO{
		ID:        peer.ID,
		Name:      peer.Name,
		Address:   peer.Address,
		IsTrusted: peer.IsTrusted,
		Status:    string(peer.Status),
	}

	if !peer.LastSeen.IsZero() {
		dto.LastSeen = peer.LastSeen.Format("2006-01-02 15:04:05")
	}

	if peer.LastSync != nil {
		dto.LastSync = peer.LastSync.Format("2006-01-02 15:04:05")
	}

	return dto
}

// CreateBackup creates a backup of the database and assets directory
func (a *App) CreateBackup() error {
	a.events.EmitProgress(ProgressData{
		ID:        "backup",
		Operation: "Création du backup",
		Current:   0,
		Total:     100,
	})

	// Get config paths
	dataDir := a.cfg.DataDir
	dbPath := filepath.Join(dataDir, "brique.db")
	assetsDir := filepath.Join(dataDir, "assets")

	// Create backups directory
	backupsDir := filepath.Join(dataDir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		a.events.EmitProgressComplete("backup")
		a.events.Error("Erreur de backup", "Impossible de créer le dossier de backup")
		return err
	}

	// Create backup directory with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDir := filepath.Join(backupsDir, fmt.Sprintf("backup_%s", timestamp))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		a.events.EmitProgressComplete("backup")
		a.events.Error("Erreur de backup", "Impossible de créer le dossier de backup")
		return err
	}

	// Update progress
	a.events.EmitProgress(ProgressData{
		ID:        "backup",
		Operation: "Copie de la base de données",
		Current:   30,
		Total:     100,
	})

	// Backup database
	dbBackupPath := filepath.Join(backupDir, "brique.db")
	if err := copyFile(dbPath, dbBackupPath); err != nil {
		a.events.EmitProgressComplete("backup")
		a.events.Error("Erreur de backup", "Impossible de copier la base de données")
		return err
	}

	// Update progress
	a.events.EmitProgress(ProgressData{
		ID:        "backup",
		Operation: "Copie des assets",
		Current:   60,
		Total:     100,
	})

	// Backup assets directory
	assetsBackupDir := filepath.Join(backupDir, "assets")
	if err := copyDir(assetsDir, assetsBackupDir); err != nil {
		// Not a critical error if assets dir doesn't exist or is empty
		a.events.Warning("Backup partiel", "Assets non copiés")
	}

	// Create a metadata file
	metadata := map[string]string{
		"timestamp": timestamp,
		"version":   "1.0",
	}
	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(filepath.Join(backupDir, "metadata.json"), metadataJSON, 0644)

	// Complete progress
	a.events.EmitProgressComplete("backup")
	a.events.Success("Backup créé", fmt.Sprintf("Backup enregistré: %s", backupDir))

	return nil
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Helper function to copy a directory recursively
func copyDir(src, dst string) error {
	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

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
