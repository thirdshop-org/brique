package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lhommenul/brique/core/db"
	"github.com/lhommenul/brique/core/models"
	"github.com/lhommenul/brique/core/services"
	"github.com/lhommenul/brique/pkg/config"
)

type Server struct {
	cfg              *config.Config
	database         *db.Database
	backpackService  *services.BackpackService
	gossipService    *services.GossipService
	discoveryService *services.DiscoveryService
	logger           *slog.Logger
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded", "data_dir", cfg.DataDir)

	// Initialize database
	database, err := db.NewDatabase(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create queries
	queries := db.New(database.DB)

	// Create backpack service
	backpackService := services.NewBackpackService(queries, cfg.AssetsDir)

	// Get port from environment or use default
	port := 8080
	if portStr := os.Getenv("BRIQUE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Create gossip service
	instanceName := os.Getenv("BRIQUE_INSTANCE_NAME")
	if instanceName == "" {
		instanceName = fmt.Sprintf("Brique-%s", os.Getenv("USER"))
		if instanceName == "Brique-" {
			instanceName = "Brique-Server"
		}
	}

	gossipAddr := fmt.Sprintf(":%d", port)
	gossipService := services.NewGossipService(queries, instanceName, gossipAddr)

	// Get instance info
	instanceInfo, err := gossipService.GetInstanceInfo(ctx)
	if err != nil {
		logger.Error("Failed to get instance info", "error", err)
		os.Exit(1)
	}

	// Create discovery service
	discoveryService := services.NewDiscoveryService(
		instanceInfo.InstanceID,
		instanceName,
		port,
		logger,
		gossipService,
	)

	// Start discovery (announce and browse)
	if err := discoveryService.Start(ctx); err != nil {
		logger.Warn("Failed to start discovery service", "error", err)
		// Not a fatal error, continue without discovery
	}

	// Create server
	srv := &Server{
		cfg:              cfg,
		database:         database,
		backpackService:  backpackService,
		gossipService:    gossipService,
		discoveryService: discoveryService,
		logger:           logger,
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	srv.setupRoutes(mux)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      srv.loggingMiddleware(srv.corsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Stop discovery service
	if srv.discoveryService != nil {
		if err := srv.discoveryService.Stop(); err != nil {
			logger.Error("Failed to stop discovery service", "error", err)
		}
	}

	logger.Info("Server exited")
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Items endpoints
	mux.HandleFunc("/api/v1/items", s.handleItems)
	mux.HandleFunc("/api/v1/items/", s.handleItemByID)

	// Assets endpoints
	mux.HandleFunc("/api/v1/items/{id}/assets", s.handleAssets)
	mux.HandleFunc("/api/v1/assets/", s.handleAssetByID)

	// Gossip endpoints
	mux.HandleFunc("/api/v1/gossip/info", s.handleGossipInfo)
	mux.HandleFunc("/api/v1/gossip/changes", s.handleGossipChanges)
	mux.HandleFunc("/api/v1/gossip/peers", s.handlePeers)
	mux.HandleFunc("/api/v1/gossip/peers/", s.handlePeerByID)
	mux.HandleFunc("/api/v1/gossip/sync/", s.handleSync)
}

// Middleware
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handlers
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		// List items
		items, err := s.backpackService.GetAllItems(ctx)
		if err != nil {
			s.jsonError(w, "Failed to list items", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, items)

	case http.MethodPost:
		// Create item
		var item models.Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			s.jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.backpackService.CreateItem(ctx, &item); err != nil {
			s.jsonError(w, "Failed to create item", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, item)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleItemByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.URL.Path[len("/api/v1/items/"):]
	if idStr == "" {
		s.jsonError(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.jsonError(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get item by ID
		item, err := s.backpackService.GetItem(ctx, id)
		if err != nil {
			s.jsonError(w, "Item not found", http.StatusNotFound)
			return
		}
		s.jsonResponse(w, item)

	case http.MethodPut:
		// Update item
		var item models.Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			s.jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		item.ID = id

		if err := s.backpackService.UpdateItem(ctx, &item); err != nil {
			s.jsonError(w, "Failed to update item", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, item)

	case http.MethodDelete:
		// Delete item
		if err := s.backpackService.DeleteItem(ctx, id); err != nil {
			s.jsonError(w, "Failed to delete item", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract item ID from path
	idStr := r.URL.Path[len("/api/v1/items/"):]
	idStr = idStr[:len(idStr)-len("/assets")]

	itemID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.jsonError(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	assets, err := s.backpackService.GetItemAssets(ctx, itemID)
	if err != nil {
		s.jsonError(w, "Failed to list assets", http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, assets)
}

func (s *Server) handleAssetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.URL.Path[len("/api/v1/assets/"):]
	if idStr == "" {
		s.jsonError(w, "Asset ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.jsonError(w, "Invalid asset ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Delete asset
		if err := s.backpackService.DeleteAsset(ctx, id); err != nil {
			s.jsonError(w, "Failed to delete asset", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGossipInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info, err := s.gossipService.GetInstanceInfo(ctx)
	if err != nil {
		s.jsonError(w, "Failed to get instance info", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"instance_id":   info.InstanceID,
		"instance_name": info.InstanceName,
		"last_sync":     info.LastSync,
		"item_count":    info.ItemCount,
	})
}

func (s *Server) handleGossipChanges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
			s.jsonError(w, "Invalid timestamp format", http.StatusBadRequest)
			return
		}
	}

	changes, err := s.gossipService.GetChanges(ctx, since)
	if err != nil {
		s.jsonError(w, "Failed to get changes", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, changes)
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		// List all peers
		peers, err := s.gossipService.GetPeers(ctx)
		if err != nil {
			s.jsonError(w, "Failed to get peers", http.StatusInternalServerError)
			return
		}
		s.jsonResponse(w, peers)

	case http.MethodPost:
		// Add a new peer manually
		var req struct {
			Name      string `json:"name"`
			Address   string `json:"address"`
			IsTrusted bool   `json:"is_trusted"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Address == "" {
			s.jsonError(w, "Name and address are required", http.StatusBadRequest)
			return
		}

		// Generate a unique ID for the peer
		peerID := fmt.Sprintf("%s.manual", req.Address)

		peer := &models.Peer{
			ID:        peerID,
			Name:      req.Name,
			Address:   req.Address,
			IsTrusted: req.IsTrusted,
		}

		if err := s.gossipService.AddPeer(ctx, peer); err != nil {
			s.jsonError(w, "Failed to add peer", http.StatusInternalServerError)
			return
		}

		s.jsonResponse(w, peer)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePeerByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract peer ID from path
	peerID := r.URL.Path[len("/api/v1/gossip/peers/"):]
	if peerID == "" {
		s.jsonError(w, "Peer ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Remove a peer
		if err := s.gossipService.RemovePeer(ctx, peerID); err != nil {
			s.jsonError(w, "Failed to remove peer", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPut:
		// Update peer trust status
		var req struct {
			IsTrusted bool `json:"is_trusted"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.gossipService.SetPeerTrust(ctx, peerID, req.IsTrusted); err != nil {
			s.jsonError(w, "Failed to update peer", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract peer ID from path
	peerID := r.URL.Path[len("/api/v1/gossip/sync/"):]
	if peerID == "" {
		s.jsonError(w, "Peer ID is required", http.StatusBadRequest)
		return
	}

	// Get peer info
	peers, err := s.gossipService.GetPeers(ctx)
	if err != nil {
		s.jsonError(w, "Failed to get peers", http.StatusInternalServerError)
		return
	}

	var peer *models.Peer
	for _, p := range peers {
		if p.ID == peerID {
			peer = &p
			break
		}
	}

	if peer == nil {
		s.jsonError(w, "Peer not found", http.StatusNotFound)
		return
	}

	// Get remote changes
	var since time.Time
	if peer.LastSync != nil {
		since = *peer.LastSync
	}

	changesURL := fmt.Sprintf("http://%s/api/v1/gossip/changes?since=%s", peer.Address, since.Format(time.RFC3339))
	resp, err := http.Get(changesURL)
	if err != nil {
		s.jsonError(w, "Failed to connect to peer", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var remoteChanges []models.Item
	if err := json.NewDecoder(resp.Body).Decode(&remoteChanges); err != nil {
		s.jsonError(w, "Failed to decode remote changes", http.StatusBadGateway)
		return
	}

	// Sync with peer
	result, err := s.gossipService.SyncWithPeer(ctx, peerID, remoteChanges)
	if err != nil {
		s.jsonError(w, "Failed to sync", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, result)
}

// Helper functions
func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
