package models

import "time"

// Item represents a physical object in the user's inventory
type Item struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Brand        string    `json:"brand"`
	Model        string    `json:"model"`
	SerialNumber string    `json:"serial_number"`
	PurchaseDate *time.Time `json:"purchase_date,omitempty"`
	PhotoPath    string    `json:"photo_path,omitempty"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Asset represents a file associated with an item (PDF, STL, firmware, etc.)
type Asset struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	Type      AssetType `json:"type"`
	Name      string    `json:"name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	FileHash  string    `json:"file_hash"` // SHA256 for integrity
	CreatedAt time.Time `json:"created_at"`
}

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeManual         AssetType = "manual"
	AssetTypeServiceManual  AssetType = "service_manual"
	AssetTypeExplodedView   AssetType = "exploded_view"
	AssetTypeSTL            AssetType = "stl"
	AssetTypeFirmware       AssetType = "firmware"
	AssetTypeDriver         AssetType = "driver"
	AssetTypeSchematic      AssetType = "schematic"
	AssetTypeOther          AssetType = "other"
)

// DocumentationHealth represents the completeness of an item's documentation
type DocumentationHealth string

const (
	HealthIncomplete DocumentationHealth = "incomplete"
	HealthPartial    DocumentationHealth = "partial"
	HealthSecured    DocumentationHealth = "secured"
)

// ItemWithAssets is a DTO that includes an item with its associated assets
type ItemWithAssets struct {
	Item   Item    `json:"item"`
	Assets []Asset `json:"assets"`
	Health DocumentationHealth `json:"health"`
}
