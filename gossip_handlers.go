package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lhommenul/brique/core/models"
)

// GossipInfoResponse represents instance information for sync
type GossipInfoResponse struct {
	InstanceID   string     `json:"instance_id"`
	InstanceName string     `json:"instance_name"`
	LastSync     *time.Time `json:"last_sync"`
	ItemCount    int        `json:"item_count"`
}

// SyncRequest represents a sync request
type SyncRequest struct {
	PeerID  string         `json:"peer_id"`
	Changes []models.Item  `json:"changes"`
	Since   time.Time      `json:"since"`
}

// SyncResponse represents a sync response
type SyncResponse struct {
	Changes []models.Item       `json:"changes"`
	Result  *models.SyncResult  `json:"result"`
}

// GetGossipInfo returns information about this instance
func (a *App) GetGossipInfo() (*GossipInfoResponse, error) {
	info, err := a.gossipService.GetInstanceInfo(a.ctx)
	if err != nil {
		return nil, err
	}

	return &GossipInfoResponse{
		InstanceID:   info.InstanceID,
		InstanceName: info.InstanceName,
		LastSync:     info.LastSync,
		ItemCount:    info.ItemCount,
	}, nil
}

// GetGossipChanges returns items modified since a given timestamp
func (a *App) GetGossipChanges(since time.Time) ([]ItemDTO, error) {
	items, err := a.gossipService.GetChanges(a.ctx, since)
	if err != nil {
		return nil, err
	}

	dtos := make([]ItemDTO, len(items))
	for i, item := range items {
		dtos[i] = itemToDTO(&item)
	}

	return dtos, nil
}

// SyncWithPeerHTTP performs synchronization with a remote peer via HTTP
func (a *App) SyncWithPeerHTTP(peerID string) (*models.SyncResult, error) {
	// Get peer info
	peers, err := a.gossipService.GetPeers(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	var peer *models.Peer
	for _, p := range peers {
		if p.ID == peerID {
			peer = &p
			break
		}
	}

	if peer == nil {
		a.events.Error("Peer introuvable", fmt.Sprintf("Le pair %s n'existe pas", peerID))
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	// Emit progress start
	progressID := fmt.Sprintf("sync-%s", peerID)
	a.events.EmitProgress(ProgressData{
		ID:        progressID,
		Operation: "Synchronisation",
		Current:   10,
		Total:     100,
	})

	// Get remote info
	infoURL := fmt.Sprintf("http://%s/api/v1/gossip/info", peer.Address)
	resp, err := http.Get(infoURL)
	if err != nil {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de connexion", fmt.Sprintf("Impossible de contacter %s", peer.Name))
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de synchronisation", "Le serveur distant a renvoyé une erreur")
		return nil, fmt.Errorf("remote server returned status %d", resp.StatusCode)
	}

	var remoteInfo GossipInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&remoteInfo); err != nil {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de synchronisation", "Réponse invalide du serveur")
		return nil, fmt.Errorf("failed to decode remote info: %w", err)
	}

	// Update progress
	a.events.EmitProgress(ProgressData{
		ID:        progressID,
		Operation: "Récupération des changements",
		Current:   30,
		Total:     100,
	})

	// Get remote changes since last sync
	var since time.Time
	if peer.LastSync != nil {
		since = *peer.LastSync
	}

	changesURL := fmt.Sprintf("http://%s/api/v1/gossip/changes?since=%s", peer.Address, since.Format(time.RFC3339))
	resp, err = http.Get(changesURL)
	if err != nil {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de synchronisation", "Impossible de récupérer les changements")
		return nil, fmt.Errorf("failed to get remote changes: %w", err)
	}
	defer resp.Body.Close()

	var remoteChanges []models.Item
	if err := json.NewDecoder(resp.Body).Decode(&remoteChanges); err != nil {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de synchronisation", "Données invalides")
		return nil, fmt.Errorf("failed to decode remote changes: %w", err)
	}

	// Update progress
	a.events.EmitProgress(ProgressData{
		ID:        progressID,
		Operation: "Application des changements",
		Current:   60,
		Total:     100,
	})

	// Sync with peer
	result, err := a.gossipService.SyncWithPeer(a.ctx, peerID, remoteChanges)
	if err != nil {
		a.events.EmitProgressComplete(progressID)
		a.events.Error("Erreur de synchronisation", err.Error())
		return nil, err
	}

	// Complete progress
	a.events.EmitProgressComplete(progressID)

	// Emit success notification
	message := fmt.Sprintf("Synchronisé avec %s: %d reçus, %d envoyés", peer.Name, result.ItemsReceived, result.ItemsSent)
	if result.Conflicts > 0 {
		message += fmt.Sprintf(", %d conflits résolus", result.Conflicts)
	}
	a.events.Success("Synchronisation réussie", message)

	return result, nil
}

// ServeGossipAPI starts the HTTP server for gossip protocol (internal use)
// This would be called in main.go startup to expose the API endpoints
func ServeGossipAPI(app *App, port int) error {
	mux := http.NewServeMux()

	// GET /api/v1/gossip/info
	mux.HandleFunc("/api/v1/gossip/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info, err := app.GetGossipInfo()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// GET /api/v1/gossip/changes?since=<timestamp>
	mux.HandleFunc("/api/v1/gossip/changes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sinceStr := r.URL.Query().Get("since")
		var since time.Time
		var err error

		if sinceStr != "" {
			since, err = time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
				return
			}
		}

		changes, err := app.GetGossipChanges(since)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(changes)
	})

	// Start server on specified port
	addr := fmt.Sprintf(":%d", port)
	app.logger.Info("Gossip API server starting", "address", addr)

	return http.ListenAndServe(addr, mux)
}
