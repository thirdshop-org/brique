package models

import "time"

// PeerStatus represents the status of a peer
type PeerStatus string

const (
	PeerStatusOnline  PeerStatus = "online"
	PeerStatusOffline PeerStatus = "offline"
	PeerStatusSyncing PeerStatus = "syncing"
)

// Peer represents a remote Brique instance
type Peer struct {
	ID        string
	Name      string
	Address   string    // IP:Port
	LastSeen  time.Time
	LastSync  *time.Time // Pointer because it can be null
	IsTrusted bool
	CreatedAt time.Time
	Status    PeerStatus // Computed field, not stored in DB
}

// SyncResult represents the result of a synchronization
type SyncResult struct {
	ItemsReceived int
	ItemsSent     int
	Conflicts     int
	DurationMs    int64
}

// SyncLog represents a synchronization log entry
type SyncLog struct {
	ID            int64
	PeerID        string
	Timestamp     time.Time
	ItemsReceived int
	ItemsSent     int
	Conflicts     int
	DurationMs    int64
	Error         string
}

// ChangeSet represents a set of changes to synchronize
type ChangeSet struct {
	Items   []Item
	Since   time.Time
	PeerID  string
}

// SyncInfo represents instance information for synchronization
type SyncInfo struct {
	InstanceID   string
	InstanceName string
	LastSync     *time.Time
	ItemCount    int
}
