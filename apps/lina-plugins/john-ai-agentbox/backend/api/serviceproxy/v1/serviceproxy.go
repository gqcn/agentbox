// This file defines version-one runtime service-proxy DTOs for the AgentBox
// plugin. Paths are plugin-relative and are published under
// /x/john-ai-agentbox/api/v1 by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ServicesReq lists runtime services discovered in one Agent runtime.
type ServicesReq struct {
	g.Meta `path:"/agents/{id}/services" method:"get" tags:"AgentBox Services" summary:"List AgentBox runtime services" dc:"List bounded TCP listen ports discovered inside one authenticated-user-owned Agent runtime. This read-only discovery response does not create bridges, proxy relays, or tunnel URLs."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// ServicesRes returns runtime services.
type ServicesRes = []AgentRuntimeServiceInfo

// ServiceReq gets one discovered runtime service by ID.
type ServiceReq struct {
	g.Meta    `path:"/agents/{id}/services/{serviceId}" method:"get" tags:"AgentBox Services" summary:"Get AgentBox runtime service" dc:"Get one bounded runtime service for an authenticated-user-owned Agent from read-only runtime service discovery. Proxy, tunnel, and bridge relay metadata remain omitted until those runtime backends are available."`
	ID        string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	ServiceID string `json:"serviceId" v:"required" dc:"Runtime service ID for the Agent-scoped port group" eg:"svc-5666"`
}

// ServiceRes returns one runtime service.
type ServiceRes = AgentRuntimeServiceInfo

// ServiceBridgesReq lists explicit loopback bridges for one Agent.
type ServiceBridgesReq struct {
	g.Meta `path:"/agents/{id}/service-bridges" method:"get" tags:"AgentBox Services" summary:"List AgentBox service bridges" dc:"List currently tracked loopback service bridges for one authenticated-user-owned Agent runtime."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
}

// ServiceBridgesRes returns bridge records.
type ServiceBridgesRes = []AgentServiceBridgeInfo

// CreateServiceBridgeReq creates an explicit loopback bridge.
type CreateServiceBridgeReq struct {
	g.Meta        `path:"/agents/{id}/service-bridges" method:"post" tags:"AgentBox Services" summary:"Create AgentBox service bridge" dc:"Create an explicit bridge for a visible service that only listens on container loopback, enabling host-local, HTTP proxy, or TCP tunnel access until the bridge is closed or the Agent runtime stops."`
	ID            string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	ServiceID     string `json:"serviceId" v:"required" dc:"Runtime service ID to bridge" eg:"svc-3000"`
	ListenAddress string `json:"listenAddress" v:"required" dc:"Loopback listen address to bridge: 127.0.0.1 or ::1" eg:"127.0.0.1"`
	Port          int    `json:"port" v:"required|min:1|max:65535" dc:"Loopback listen port to bridge" eg:"3000"`
}

// CreateServiceBridgeRes returns the active bridge.
type CreateServiceBridgeRes = AgentServiceBridgeInfo

// DeleteServiceBridgeReq closes one explicit loopback bridge.
type DeleteServiceBridgeReq struct {
	g.Meta   `path:"/agents/{id}/service-bridges/{bridgeId}" method:"delete" tags:"AgentBox Services" summary:"Delete AgentBox service bridge" dc:"Close one active loopback service bridge and stop its relay backend so the service returns to bridge-required access when runtime support is available."`
	ID       string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	BridgeID string `json:"bridgeId" v:"required" dc:"Bridge ID to close" eg:"brg-1234567890abcdef"`
}

// DeleteServiceBridgeRes reports bridge deletion state.
type DeleteServiceBridgeRes struct {
	Deleted bool `json:"deleted" dc:"Whether the bridge was closed" eg:"true"`
}

// AgentServiceListenAddress describes one listen address for an Agent runtime service.
type AgentServiceListenAddress struct {
	Address           string `json:"address" dc:"Container listen address for this service endpoint" eg:"0.0.0.0"`
	Port              int    `json:"port" dc:"Container listen port for this service endpoint" eg:"5666"`
	Network           string `json:"network" dc:"Network family: tcp4 or tcp6" eg:"tcp4"`
	AccessStatus      string `json:"accessStatus" dc:"Access status: direct, bridge_required, bridged, unavailable" eg:"direct"`
	BridgeID          string `json:"bridgeId,omitempty" dc:"Active bridge ID for loopback access; omitted when no bridge is active" eg:"brg-1234567890abcdef"`
	LocalHost         string `json:"localHost,omitempty" dc:"Host-local address for bridged service access; omitted when no host-local bridge is active" eg:"127.0.0.1"`
	LocalPort         int    `json:"localPort,omitempty" dc:"Host-local port for bridged service access; omitted when no host-local bridge is active" eg:"49152"`
	UnavailableReason string `json:"unavailableReason,omitempty" dc:"Human-readable reason when this listen address cannot currently be accessed" eg:"container network address is unavailable"`
}

// AgentRuntimeServiceInfo is the public projection of one discovered runtime service.
type AgentRuntimeServiceInfo struct {
	ID                string                      `json:"id" dc:"Stable service ID derived from the Agent-scoped runtime port group" eg:"svc-5666"`
	AgentID           string                      `json:"agentId" dc:"Agent ID that owns this runtime service" eg:"agt-1234567890abcdef"`
	Port              int                         `json:"port" dc:"Container listen port" eg:"5666"`
	Protocol          string                      `json:"protocol" dc:"Detected protocol: http, https, tcp, unknown" eg:"http"`
	AccessStatus      string                      `json:"accessStatus" dc:"Aggregated access status: direct, bridge_required, bridged, unavailable" eg:"direct"`
	ListenAddresses   []AgentServiceListenAddress `json:"listenAddresses" dc:"Listen addresses discovered for this service port" eg:"[]"`
	ProcessName       string                      `json:"processName,omitempty" dc:"Best-effort process name that owns the listen socket; omitted when unavailable" eg:"node"`
	PID               string                      `json:"pid,omitempty" dc:"Best-effort process ID that owns the listen socket; omitted when unavailable" eg:"1234"`
	ProxyURL          string                      `json:"proxyUrl,omitempty" dc:"Stable short HTTP proxy URL for HTTP or HTTPS services; omitted until the proxy relay runtime is available" eg:"/x/john-ai-agentbox/api/v1/proxy/5G3r1q8FQbT0xYz9/"`
	TunnelURL         string                      `json:"tunnelUrl,omitempty" dc:"Stable WebSocket tunnel URL for raw TCP services; omitted until the tunnel runtime is available" eg:"/x/john-ai-agentbox/api/v1/ws/agents/agt-123/services/svc-5432/tcp?key=5G3r1q8FQbT0xYz9"`
	TunnelCommand     string                      `json:"tunnelCommand,omitempty" dc:"Suggested local helper command for connecting to this TCP service; omitted until the tunnel runtime is available" eg:"john-ai-agentbox tunnel --agent agt-123 --service svc-5432 --local 15432"`
	BridgeID          string                      `json:"bridgeId,omitempty" dc:"Active bridge ID when this service is currently bridged" eg:"brg-1234567890abcdef"`
	LocalHost         string                      `json:"localHost,omitempty" dc:"Host-local address for bridged service access; omitted when no host-local bridge is active" eg:"127.0.0.1"`
	LocalPort         int                         `json:"localPort,omitempty" dc:"Host-local port for bridged service access; omitted when no host-local bridge is active" eg:"49152"`
	UnavailableReason string                      `json:"unavailableReason,omitempty" dc:"Human-readable reason when this service cannot currently be accessed" eg:"agent container is not running"`
	LastCheckedAt     int64                       `json:"lastCheckedAt" dc:"Last service discovery time as Unix timestamp in milliseconds" eg:"1704067200000"`
}

// AgentServiceBridgeInfo describes one explicit loopback bridge resource.
type AgentServiceBridgeInfo struct {
	ID            string `json:"id" dc:"Bridge ID" eg:"brg-1234567890abcdef"`
	AgentID       string `json:"agentId" dc:"Agent ID that owns this bridge" eg:"agt-1234567890abcdef"`
	ServiceID     string `json:"serviceId" dc:"Service ID bridged by this resource" eg:"svc-3000"`
	ListenAddress string `json:"listenAddress" dc:"Loopback listen address bridged by this resource" eg:"127.0.0.1"`
	Port          int    `json:"port" dc:"Target loopback listen port" eg:"3000"`
	BridgePort    int    `json:"bridgePort" dc:"Bridge relay port inside the target container network namespace" eg:"43001"`
	LocalHost     string `json:"localHost,omitempty" dc:"Host-local address that can access this bridge from the AgentBox machine" eg:"127.0.0.1"`
	LocalPort     int    `json:"localPort,omitempty" dc:"Host-local port that can access this bridge from the AgentBox machine" eg:"49152"`
	Status        string `json:"status" dc:"Bridge status: active, closed, error" eg:"active"`
	ErrorMessage  string `json:"errorMessage,omitempty" dc:"Bridge failure details when status is error" eg:"create bridge sidecar failed"`
	CreatedAt     int64  `json:"createdAt" dc:"Bridge creation time as Unix timestamp in milliseconds" eg:"1704067200000"`
	ClosedAt      *int64 `json:"closedAt,omitempty" dc:"Bridge close time as Unix timestamp in milliseconds; omitted while active" eg:"1704067500000"`
}
