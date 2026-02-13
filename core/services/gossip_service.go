package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
)

// GossipService handles peer discovery and synchronization
type GossipService struct {
	queries      *db.Queries
	instanceID   string
	instanceName string
	listenAddr   string
}

// NewGossipService creates a new GossipService
func NewGossipService(queries *db.Queries, instanceName, listenAddr string) *GossipService {
	// Generate or load instance ID
	instanceID := uuid.New().String()

	return &GossipService{
		queries:      queries,
		instanceID:   instanceID,
		instanceName: instanceName,
		listenAddr:   listenAddr,
	}
}

// GetInstanceInfo returns information about this instance
func (s *GossipService) GetInstanceInfo(ctx context.Context) (*models.SyncInfo, error) {
	count, err := s.queries.CountItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	return &models.SyncInfo{
		InstanceID:   s.instanceID,
		InstanceName: s.instanceName,
		LastSync:     nil, // Will be set per-peer
		ItemCount:    int(count),
	}, nil
}

// AddPeer adds a new peer to the list
func (s *GossipService) AddPeer(ctx context.Context, peer *models.Peer) error {
	// Check if peer already exists
	_, err := s.queries.GetPeer(ctx, peer.ID)
	if err == nil {
		// Peer exists, update last seen
		return s.UpdatePeerLastSeen(ctx, peer.ID)
	}

	// Create new peer
	_, err = s.queries.CreatePeer(ctx, db.CreatePeerParams{
		ID:        peer.ID,
		Name:      peer.Name,
		Address:   peer.Address,
		LastSeen:  sql.NullTime{Time: time.Now(), Valid: true},
		IsTrusted: sql.NullBool{Bool: peer.IsTrusted, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}

	return nil
}

// GetPeers returns all known peers
func (s *GossipService) GetPeers(ctx context.Context) ([]models.Peer, error) {
	dbPeers, err := s.queries.GetAllPeers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	peers := make([]models.Peer, len(dbPeers))
	for i, p := range dbPeers {
		peers[i] = s.dbPeerToModel(p)
	}

	return peers, nil
}

// GetTrustedPeers returns only trusted peers
func (s *GossipService) GetTrustedPeers(ctx context.Context) ([]models.Peer, error) {
	dbPeers, err := s.queries.GetTrustedPeers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted peers: %w", err)
	}

	peers := make([]models.Peer, len(dbPeers))
	for i, p := range dbPeers {
		peers[i] = s.dbPeerToModel(p)
	}

	return peers, nil
}

// UpdatePeerLastSeen updates the last seen timestamp for a peer
func (s *GossipService) UpdatePeerLastSeen(ctx context.Context, peerID string) error {
	return s.queries.UpdatePeerLastSeen(ctx, db.UpdatePeerLastSeenParams{
		LastSeen: sql.NullTime{Time: time.Now(), Valid: true},
		ID:       peerID,
	})
}

// UpdatePeerLastSync updates the last sync timestamp for a peer
func (s *GossipService) UpdatePeerLastSync(ctx context.Context, peerID string) error {
	return s.queries.UpdatePeerLastSync(ctx, db.UpdatePeerLastSyncParams{
		LastSync: sql.NullTime{Time: time.Now(), Valid: true},
		ID:       peerID,
	})
}

// SetPeerTrust sets whether a peer is trusted
func (s *GossipService) SetPeerTrust(ctx context.Context, peerID string, trusted bool) error {
	return s.queries.UpdatePeerTrust(ctx, db.UpdatePeerTrustParams{
		IsTrusted: sql.NullBool{Bool: trusted, Valid: true},
		ID:        peerID,
	})
}

// RemovePeer removes a peer from the list
func (s *GossipService) RemovePeer(ctx context.Context, peerID string) error {
	return s.queries.DeletePeer(ctx, peerID)
}

// GetChanges returns items modified since a given timestamp
func (s *GossipService) GetChanges(ctx context.Context, since time.Time) ([]models.Item, error) {
	dbItems, err := s.queries.GetItemsModifiedSince(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes: %w", err)
	}

	items := make([]models.Item, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = s.dbItemToModel(dbItem)
	}

	return items, nil
}

// LogSync logs a synchronization event
func (s *GossipService) LogSync(ctx context.Context, log *models.SyncLog) error {
	_, err := s.queries.CreateSyncLog(ctx, db.CreateSyncLogParams{
		PeerID:        log.PeerID,
		Timestamp:     sql.NullTime{Time: log.Timestamp, Valid: true},
		ItemsReceived: sql.NullInt64{Int64: int64(log.ItemsReceived), Valid: true},
		ItemsSent:     sql.NullInt64{Int64: int64(log.ItemsSent), Valid: true},
		Conflicts:     sql.NullInt64{Int64: int64(log.Conflicts), Valid: true},
		DurationMs:    sql.NullInt64{Int64: log.DurationMs, Valid: true},
		Error:         sql.NullString{String: log.Error, Valid: log.Error != ""},
	})

	if err != nil {
		return fmt.Errorf("failed to log sync: %w", err)
	}

	return nil
}

// GetSyncHistory returns the sync history for a peer
func (s *GossipService) GetSyncHistory(ctx context.Context, peerID string, limit int) ([]models.SyncLog, error) {
	dbLogs, err := s.queries.GetSyncLogsByPeer(ctx, db.GetSyncLogsByPeerParams{
		PeerID: peerID,
		Limit:  int64(limit),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get sync history: %w", err)
	}

	logs := make([]models.SyncLog, len(dbLogs))
	for i, l := range dbLogs {
		logs[i] = s.dbSyncLogToModel(l)
	}

	return logs, nil
}

// GetRecentSyncHistory returns the most recent sync events
func (s *GossipService) GetRecentSyncHistory(ctx context.Context, limit int) ([]models.SyncLog, error) {
	dbLogs, err := s.queries.GetRecentSyncLogs(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sync history: %w", err)
	}

	logs := make([]models.SyncLog, len(dbLogs))
	for i, l := range dbLogs {
		logs[i] = s.dbSyncLogToModel(l)
	}

	return logs, nil
}

// SyncWithPeer synchronizes with a remote peer
func (s *GossipService) SyncWithPeer(ctx context.Context, peerID string, remoteChanges []models.Item) (*models.SyncResult, error) {
	startTime := time.Now()

	// Get peer info
	peer, err := s.queries.GetPeer(ctx, peerID)
	if err != nil {
		return nil, fmt.Errorf("peer not found: %w", err)
	}

	// Get local changes since last sync
	var since time.Time
	if peer.LastSync.Valid {
		since = peer.LastSync.Time
	}

	localChanges, err := s.GetChanges(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get local changes: %w", err)
	}

	// Apply remote changes with conflict resolution
	conflicts := 0
	for _, remoteItem := range remoteChanges {
		// Check if item exists locally
		localItem, err := s.queries.GetItemByID(ctx, remoteItem.ID)

		if err != nil {
			// Item doesn't exist, create it
			_, err = s.queries.CreateItem(ctx, db.CreateItemParams{
				Name:         remoteItem.Name,
				Category:     remoteItem.Category,
				Brand:        remoteItem.Brand,
				Model:        remoteItem.Model,
				SerialNumber: remoteItem.SerialNumber,
				PurchaseDate: sql.NullTime{Time: *remoteItem.PurchaseDate, Valid: remoteItem.PurchaseDate != nil},
				PhotoPath:    remoteItem.PhotoPath,
				Notes:        remoteItem.Notes,
				CreatedAt:    remoteItem.CreatedAt,
				UpdatedAt:    remoteItem.UpdatedAt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create item: %w", err)
			}
		} else {
			// Item exists, check for conflict
			if localItem.UpdatedAt.After(remoteItem.UpdatedAt) {
				// Local version is newer, skip (Last-Write-Wins)
				conflicts++
				continue
			}

			// Remote version is newer or same, update
			err = s.queries.UpdateItem(ctx, db.UpdateItemParams{
				Name:         remoteItem.Name,
				Category:     remoteItem.Category,
				Brand:        remoteItem.Brand,
				Model:        remoteItem.Model,
				SerialNumber: remoteItem.SerialNumber,
				PurchaseDate: sql.NullTime{Time: *remoteItem.PurchaseDate, Valid: remoteItem.PurchaseDate != nil},
				PhotoPath:    remoteItem.PhotoPath,
				Notes:        remoteItem.Notes,
				UpdatedAt:    remoteItem.UpdatedAt,
				ID:           remoteItem.ID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update item: %w", err)
			}
		}
	}

	// Update peer's last sync timestamp
	if err := s.UpdatePeerLastSync(ctx, peerID); err != nil {
		return nil, fmt.Errorf("failed to update peer last sync: %w", err)
	}

	// Calculate result
	duration := time.Since(startTime)
	result := &models.SyncResult{
		ItemsReceived: len(remoteChanges),
		ItemsSent:     len(localChanges),
		Conflicts:     conflicts,
		DurationMs:    duration.Milliseconds(),
	}

	// Log the sync
	syncLog := &models.SyncLog{
		PeerID:        peerID,
		Timestamp:     time.Now(),
		ItemsReceived: result.ItemsReceived,
		ItemsSent:     result.ItemsSent,
		Conflicts:     result.Conflicts,
		DurationMs:    result.DurationMs,
	}

	if err := s.LogSync(ctx, syncLog); err != nil {
		// Log error but don't fail the sync
		fmt.Printf("failed to log sync: %v\n", err)
	}

	return result, nil
}

// Helper functions to convert DB models to domain models

func (s *GossipService) dbPeerToModel(dbPeer db.Peer) models.Peer {
	peer := models.Peer{
		ID:        dbPeer.ID,
		Name:      dbPeer.Name,
		Address:   dbPeer.Address,
		IsTrusted: dbPeer.IsTrusted.Bool,
		Status:    models.PeerStatusOffline, // Default, will be updated by discovery
	}

	if dbPeer.CreatedAt.Valid {
		peer.CreatedAt = dbPeer.CreatedAt.Time
	}

	if dbPeer.LastSeen.Valid {
		peer.LastSeen = dbPeer.LastSeen.Time
		// Consider online if seen in last 5 minutes
		if time.Since(peer.LastSeen) < 5*time.Minute {
			peer.Status = models.PeerStatusOnline
		}
	}

	if dbPeer.LastSync.Valid {
		peer.LastSync = &dbPeer.LastSync.Time
	}

	return peer
}

func (s *GossipService) dbSyncLogToModel(dbLog db.SyncLog) models.SyncLog {
	log := models.SyncLog{
		ID:     dbLog.ID,
		PeerID: dbLog.PeerID,
	}

	if dbLog.Timestamp.Valid {
		log.Timestamp = dbLog.Timestamp.Time
	}

	if dbLog.ItemsReceived.Valid {
		log.ItemsReceived = int(dbLog.ItemsReceived.Int64)
	}

	if dbLog.ItemsSent.Valid {
		log.ItemsSent = int(dbLog.ItemsSent.Int64)
	}

	if dbLog.Conflicts.Valid {
		log.Conflicts = int(dbLog.Conflicts.Int64)
	}

	if dbLog.DurationMs.Valid {
		log.DurationMs = dbLog.DurationMs.Int64
	}

	if dbLog.Error.Valid {
		log.Error = dbLog.Error.String
	}

	return log
}

func (s *GossipService) dbItemToModel(dbItem db.Item) models.Item {
	item := models.Item{
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
