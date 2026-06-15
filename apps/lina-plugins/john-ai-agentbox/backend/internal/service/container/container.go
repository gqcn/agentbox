// Package container owns AgentBox container runtime JSON behavior. Docker
// health and the safe lifecycle subset use explicitly injected runtime
// backends. The lifecycle subset only manages plugin-labelled containers scoped
// to the authenticated AgentBox user; creation remains unavailable until image
// and runtime configuration are migrated.
package container

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"

	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
	"john-ai-agentbox/backend/internal/service/container/internal/dockerhealth"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// DockerHealthResponse describes Docker runtime health.
type DockerHealthResponse struct {
	OK         bool
	APIVersion string
	OSType     string
	Error      string
}

// MountInfo describes one runtime container mount.
type MountInfo struct {
	Type        string
	Source      string
	Destination string
	Name        string
}

// ContainerInfo describes one AgentBox managed runtime container.
type ContainerInfo struct {
	ID        string
	Name      string
	DockerID  string
	Image     string
	State     string
	Status    string
	CreatedAt int64
	Mounts    []MountInfo
	Labels    map[string]string
	Workspace string
}

// LogsResponse returns runtime logs.
type LogsResponse struct {
	Logs string
}

// RuntimeConfig contains pure value settings for the Docker-backed AgentBox
// runtime adapter.
type RuntimeConfig struct {
	Host             string
	ContainerLogTail int
	StopTimeout      time.Duration
	Workspace        workspacesvc.Config
	Service          serviceproxysvc.Config
}

// DockerHealthBackend is the narrow runtime dependency used by health checks.
// It must not expose Docker client internals.
type DockerHealthBackend interface {
	// Ping verifies the Docker daemon and returns a public health projection.
	// Implementations may return low-level causes; the service wraps them in
	// CodeContainerRuntimeUnavailable before they reach HTTP callers.
	Ping(ctx context.Context) (*DockerHealthResponse, error)
}

// ContainerLifecycleBackend is the narrow runtime dependency for the safe
// lifecycle subset. Implementations must scope all reads and mutations to
// plugin-labelled containers owned by the provided AgentBox user ID.
type ContainerLifecycleBackend interface {
	// List returns plugin-managed containers owned by userID.
	List(ctx context.Context, userID string) ([]ContainerInfo, error)
	// Inspect returns one plugin-managed container owned by userID.
	Inspect(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Start starts one plugin-managed container owned by userID.
	Start(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Stop stops one plugin-managed container owned by userID.
	Stop(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Delete removes one plugin-managed container owned by userID.
	Delete(ctx context.Context, userID string, containerID string) (bool, error)
	// Logs returns recent logs for one plugin-managed container owned by userID.
	Logs(ctx context.Context, userID string, containerID string) (*LogsResponse, error)
}

// DockerRuntimeBackend combines the Docker health and safe lifecycle backends.
type DockerRuntimeBackend interface {
	DockerHealthBackend
	ContainerLifecycleBackend
	catalogsvc.AgentRuntimeBackend
	workspacesvc.RuntimeBackend
	serviceproxysvc.RuntimeBackend
}

// Service defines AgentBox container runtime behavior.
type Service interface {
	// DockerHealth reports runtime health for the authenticated AgentBox user.
	DockerHealth(ctx context.Context, userID string) (*DockerHealthResponse, error)
	// List lists runtime containers for the authenticated AgentBox user.
	List(ctx context.Context, userID string) ([]ContainerInfo, error)
	// Create creates a runtime container for the authenticated AgentBox user.
	Create(ctx context.Context, userID string, name string) (*ContainerInfo, error)
	// Detail gets one runtime container visible to the authenticated AgentBox user.
	Detail(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Start starts one runtime container visible to the authenticated AgentBox user.
	Start(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Stop stops one runtime container visible to the authenticated AgentBox user.
	Stop(ctx context.Context, userID string, containerID string) (*ContainerInfo, error)
	// Delete deletes one runtime container visible to the authenticated AgentBox user.
	Delete(ctx context.Context, userID string, containerID string) (bool, error)
	// Logs reads logs for one runtime container visible to the authenticated AgentBox user.
	Logs(ctx context.Context, userID string, containerID string) (*LogsResponse, error)
}

// serviceImpl is the default container service implementation.
type serviceImpl struct {
	healthBackend    DockerHealthBackend
	lifecycleBackend ContainerLifecycleBackend
}

var _ Service = (*serviceImpl)(nil)

// New creates a container service with explicit runtime dependency injection.
func New(healthBackend DockerHealthBackend, lifecycleBackend ContainerLifecycleBackend) (Service, error) {
	if healthBackend == nil {
		return nil, gerror.New("agentbox container health backend is required")
	}
	if lifecycleBackend == nil {
		return nil, gerror.New("agentbox container lifecycle backend is required")
	}
	return &serviceImpl{healthBackend: healthBackend, lifecycleBackend: lifecycleBackend}, nil
}

// NewDockerHealthBackend creates the default Docker-backed health dependency.
func NewDockerHealthBackend() DockerHealthBackend {
	return &dockerHealthBackend{runtime: dockerhealth.New()}
}

// NewDockerRuntimeBackend creates the default Docker-backed runtime dependency.
func NewDockerRuntimeBackend(configs ...RuntimeConfig) DockerRuntimeBackend {
	config := RuntimeConfig{}
	if len(configs) > 0 {
		config = configs[0]
	}
	return newDockerRuntimeBackend(normalizeRuntimeConfig(config))
}

// DefaultRuntimeConfig returns conservative Docker runtime defaults.
func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		ContainerLogTail: 400,
		StopTimeout:      10 * time.Second,
		Workspace:        workspacesvc.DefaultConfig(),
		Service:          serviceproxysvc.DefaultConfig(),
	}
}

func normalizeRuntimeConfig(config RuntimeConfig) RuntimeConfig {
	defaults := DefaultRuntimeConfig()
	config.Host = strings.TrimSpace(config.Host)
	if config.ContainerLogTail <= 0 {
		config.ContainerLogTail = defaults.ContainerLogTail
	}
	if config.StopTimeout <= 0 {
		config.StopTimeout = defaults.StopTimeout
	}
	config.Workspace = workspacesvc.NormalizeConfig(config.Workspace)
	config.Service = normalizeServiceRuntimeConfig(config.Service)
	return config
}

func normalizeServiceRuntimeConfig(config serviceproxysvc.Config) serviceproxysvc.Config {
	if config.RuntimeServiceListLimit <= 0 {
		return serviceproxysvc.DefaultConfig()
	}
	return config
}

func normalizeRuntimeUserAndID(userID string, values ...string) (string, []string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", nil, newInvalidInputError()
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return "", nil, newInvalidInputError()
		}
		normalized = append(normalized, value)
	}
	return userID, normalized, nil
}

func wrapRuntimeUnavailable(cause error) error {
	if cause == nil {
		return newRuntimeUnavailableError()
	}
	return bizerr.WrapCode(cause, CodeContainerRuntimeUnavailable)
}
