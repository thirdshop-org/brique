package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/lhommenul/brique/core/models"
)

const (
	// Service type for mDNS discovery
	ServiceType = "_brique._tcp"

	// Domain for mDNS
	Domain = "local."
)

// DiscoveryService handles peer discovery via mDNS
type DiscoveryService struct {
	instanceID   string
	instanceName string
	port         int
	logger       *slog.Logger
	gossipSvc    *GossipService

	server  *mdns.Server
	browser chan *mdns.ServiceEntry
	stopCh  chan struct{}
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(instanceID, instanceName string, port int, logger *slog.Logger, gossipSvc *GossipService) *DiscoveryService {
	return &DiscoveryService{
		instanceID:   instanceID,
		instanceName: instanceName,
		port:         port,
		logger:       logger,
		gossipSvc:    gossipSvc,
		browser:      make(chan *mdns.ServiceEntry, 32),
		stopCh:       make(chan struct{}),
	}
}

// Start begins the discovery service (announce and browse)
func (d *DiscoveryService) Start(ctx context.Context) error {
	// Start announcing this instance
	if err := d.startAnnouncing(); err != nil {
		return fmt.Errorf("failed to start announcing: %w", err)
	}

	// Start browsing for other instances
	go d.startBrowsing(ctx)

	d.logger.Info("Discovery service started",
		"instance_id", d.instanceID,
		"instance_name", d.instanceName,
		"port", d.port)

	return nil
}

// Stop stops the discovery service
func (d *DiscoveryService) Stop() error {
	close(d.stopCh)

	if d.server != nil {
		if err := d.server.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown mDNS server: %w", err)
		}
	}

	d.logger.Info("Discovery service stopped")
	return nil
}

// startAnnouncing announces this instance via mDNS
func (d *DiscoveryService) startAnnouncing() error {
	// Get local IP addresses
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("failed to get interface addresses: %w", err)
	}

	var ips []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}

	if len(ips) == 0 {
		return fmt.Errorf("no valid IP addresses found")
	}

	// Setup mDNS service
	service, err := mdns.NewMDNSService(
		d.instanceID,                    // Instance (unique ID)
		ServiceType,                     // Service type
		Domain,                          // Domain
		"",                              // Host (empty = use hostname)
		d.port,                          // Port
		ips,                             // IPs
		[]string{fmt.Sprintf("name=%s", d.instanceName)}, // TXT records
	)

	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	// Create and start the mDNS server
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("failed to create mDNS server: %w", err)
	}

	d.server = server
	d.logger.Info("mDNS announcing started", "service", ServiceType, "ips", ips)

	return nil
}

// startBrowsing continuously browses for other Brique instances
func (d *DiscoveryService) startBrowsing(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Browse every 10 seconds
	defer ticker.Stop()

	// Initial browse
	d.browse(ctx)

	for {
		select {
		case <-ticker.C:
			d.browse(ctx)
		case <-d.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// browse performs a single mDNS browse operation
func (d *DiscoveryService) browse(ctx context.Context) {
	entriesCh := make(chan *mdns.ServiceEntry, 32)

	// Start browsing
	go func() {
		params := &mdns.QueryParam{
			Service: ServiceType,
			Domain:  Domain,
			Entries: entriesCh,
			Timeout: 5 * time.Second,
		}

		if err := mdns.Query(params); err != nil {
			d.logger.Error("mDNS query failed", "error", err)
		}
		close(entriesCh)
	}()

	// Process discovered entries
	for entry := range entriesCh {
		// Skip self
		if entry.Name == d.instanceID+"."+ServiceType+"."+Domain {
			continue
		}

		d.handleDiscoveredPeer(ctx, entry)
	}
}

// handleDiscoveredPeer processes a discovered peer
func (d *DiscoveryService) handleDiscoveredPeer(ctx context.Context, entry *mdns.ServiceEntry) {
	// Extract instance name from TXT records
	instanceName := d.instanceName // Default
	for _, txt := range entry.InfoFields {
		if len(txt) > 5 && txt[:5] == "name=" {
			instanceName = txt[5:]
			break
		}
	}

	// Create peer
	peer := &models.Peer{
		ID:        entry.Name, // Use mDNS name as ID
		Name:      instanceName,
		Address:   fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port),
		IsTrusted: false, // Not trusted by default
	}

	// Add or update peer
	if err := d.gossipSvc.AddPeer(ctx, peer); err != nil {
		d.logger.Error("Failed to add discovered peer",
			"peer_id", peer.ID,
			"peer_name", peer.Name,
			"address", peer.Address,
			"error", err)
		return
	}

	d.logger.Info("Peer discovered",
		"peer_id", peer.ID,
		"peer_name", peer.Name,
		"address", peer.Address)
}

// GetDiscoveredPeers returns the list of discovered peers (delegates to GossipService)
func (d *DiscoveryService) GetDiscoveredPeers(ctx context.Context) ([]models.Peer, error) {
	return d.gossipSvc.GetPeers(ctx)
}
