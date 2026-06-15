// Package serviceproxy owns AgentBox runtime service discovery and bridge JSON
// behavior. The current migration slice supports read-only runtime service
// discovery after AgentBox user ownership checks; actual proxy, tunnel, and
// bridge relay backends are migrated in a later runtime slice.
package serviceproxy

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

const (
	// DefaultRuntimeServiceListLimit caps one service discovery response.
	DefaultRuntimeServiceListLimit = 100
	// AgentServiceProtocolHTTP marks an HTTP service detected by runtime probing.
	AgentServiceProtocolHTTP = "http"
	// AgentServiceProtocolHTTPS marks an HTTPS service detected by runtime probing.
	AgentServiceProtocolHTTPS = "https"
	// AgentServiceProtocolTCP marks a service reachable as raw TCP.
	AgentServiceProtocolTCP = "tcp"
	// AgentServiceProtocolUnknown marks a service whose protocol is unknown.
	AgentServiceProtocolUnknown = "unknown"
)

const (
	// AgentServiceAccessDirect means the service can be proxied without an explicit bridge.
	AgentServiceAccessDirect = "direct"
	// AgentServiceAccessBridgeRequired means loopback access needs an explicit bridge.
	AgentServiceAccessBridgeRequired = "bridge_required"
	// AgentServiceAccessBridged means a loopback service has an active bridge.
	AgentServiceAccessBridged = "bridged"
	// AgentServiceAccessUnavailable means runtime or bridge access is unavailable.
	AgentServiceAccessUnavailable = "unavailable"
)

const (
	// AgentServiceBridgeStatusActive marks a bridge that accepts traffic.
	AgentServiceBridgeStatusActive = "active"
	// AgentServiceBridgeStatusClosed marks a bridge that has been closed.
	AgentServiceBridgeStatusClosed = "closed"
	// AgentServiceBridgeStatusError marks a bridge whose relay backend failed.
	AgentServiceBridgeStatusError = "error"
)

// ListenAddress describes one listen address for an Agent runtime service.
type ListenAddress struct {
	Address           string
	Port              int
	Network           string
	AccessStatus      string
	BridgeID          string
	LocalHost         string
	LocalPort         int
	UnavailableReason string
}

// RuntimeServiceInfo is the service-layer projection of one discovered service.
type RuntimeServiceInfo struct {
	ID                string
	AgentID           string
	Port              int
	Protocol          string
	AccessStatus      string
	ListenAddresses   []ListenAddress
	ProcessName       string
	PID               string
	ProxyURL          string
	TunnelURL         string
	TunnelCommand     string
	BridgeID          string
	LocalHost         string
	LocalPort         int
	UnavailableReason string
	LastCheckedAt     int64
}

// BridgeInfo describes one explicit loopback service bridge.
type BridgeInfo struct {
	ID            string
	AgentID       string
	ServiceID     string
	ListenAddress string
	Port          int
	BridgePort    int
	LocalHost     string
	LocalPort     int
	Status        string
	ErrorMessage  string
	CreatedAt     int64
	ClosedAt      *int64
}

// BridgeInput carries bridge creation inputs after controller projection.
type BridgeInput struct {
	ServiceID     string
	ListenAddress string
	Port          int
}

// Config contains pure value settings for runtime service discovery.
type Config struct {
	RuntimeServiceListLimit int
}

// RuntimeBackend discovers services from one visible Agent runtime. The backend
// must scope every operation to the authenticated AgentBox user and Agent and
// must reject or hide containers that do not carry the plugin ownership labels.
type RuntimeBackend interface {
	// RuntimeServices lists bounded TCP listen sockets for one visible Agent runtime without creating bridges or relays.
	RuntimeServices(ctx context.Context, userID string, agentID string) ([]RuntimeServiceInfo, error)
}

// Service defines AgentBox service discovery and bridge behavior. Methods
// validate the current AgentBox user's ownership before touching runtime-backed
// service, tunnel, or bridge data.
type Service interface {
	// Services lists bounded runtime services for one visible Agent using the injected runtime backend.
	Services(ctx context.Context, userID string, agentID string) ([]RuntimeServiceInfo, error)
	// Service gets one runtime service for one visible Agent. The service ID is checked as part of the Agent-scoped proxy visibility boundary.
	Service(ctx context.Context, userID string, agentID string, serviceID string) (RuntimeServiceInfo, error)
	// ServiceBridges lists explicit bridges for one visible Agent. Runtime-backed
	// bridge storage is not available in this migration slice.
	ServiceBridges(ctx context.Context, userID string, agentID string) ([]BridgeInfo, error)
	// CreateServiceBridge creates an explicit loopback bridge for one visible
	// service. The current slice validates ownership and input, then reports the
	// runtime backend as unavailable.
	CreateServiceBridge(ctx context.Context, userID string, agentID string, input BridgeInput) (BridgeInfo, error)
	// DeleteServiceBridge closes one explicit bridge for a visible Agent. Bridge
	// lookup remains runtime-owned and is unavailable until proxy migration lands.
	DeleteServiceBridge(ctx context.Context, userID string, agentID string, bridgeID string) (bool, error)
}

// serviceImpl is the default service-proxy implementation.
type serviceImpl struct {
	accessSvc      accesssvc.Service
	runtimeBackend RuntimeBackend
	config         Config
}

var _ Service = (*serviceImpl)(nil)

// New creates a service-proxy service with explicit access and runtime dependency injection.
func New(accessSvc accesssvc.Service, runtimeBackend RuntimeBackend, configs ...Config) (Service, error) {
	if accessSvc == nil {
		return nil, gerror.New("agentbox service proxy access service is required")
	}
	config := Config{}
	if len(configs) > 0 {
		config = configs[0]
	}
	return &serviceImpl{accessSvc: accessSvc, runtimeBackend: runtimeBackend, config: normalizeConfig(config)}, nil
}

// DefaultConfig returns conservative runtime service discovery limits.
func DefaultConfig() Config {
	return Config{RuntimeServiceListLimit: DefaultRuntimeServiceListLimit}
}

func normalizeConfig(config Config) Config {
	if config.RuntimeServiceListLimit <= 0 {
		return DefaultConfig()
	}
	return config
}
