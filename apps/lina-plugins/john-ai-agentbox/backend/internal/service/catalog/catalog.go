// Package catalog owns AgentBox provider, provider-model, and coding-image
// persistence. It uses only plugin-generated DAO/DO/Entity types so AgentBox
// AI configuration data stays isolated from LinaPro host AI tables.
package catalog

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// ProtocolOpenAI identifies OpenAI-compatible provider models.
	ProtocolOpenAI = "openai"
	// ProtocolAnthropic identifies Anthropic-compatible provider models.
	ProtocolAnthropic = "anthropic"
	// ModelSourceManual identifies operator-managed model records.
	ModelSourceManual = "manual"
	// ModelSourceAPI identifies model records synchronized from provider APIs.
	ModelSourceAPI = "api"
	// AgentTypeClaudeCode identifies Claude Code based coding images.
	AgentTypeClaudeCode = "claude_code"
	// AgentTypeCodex identifies Codex based coding images.
	AgentTypeCodex = "codex"
	// AgentTypeCustom identifies custom coding images.
	AgentTypeCustom = "custom"
	// DefaultShell is the fallback shell for coding-image profiles.
	DefaultShell = "/bin/bash"
)

const defaultRemoteModelSyncLimit = 200

// ProviderInput carries provider fields accepted by create and update actions.
type ProviderInput struct {
	Name             string
	HomepageURL      string
	Notes            string
	APIKey           string
	OpenAIBaseURL    string
	AnthropicBaseURL string
}

// ProviderInfo is the service-level provider projection.
type ProviderInfo struct {
	ID               int64
	Name             string
	HomepageURL      string
	Notes            string
	APIKeyMasked     string
	APIKeyConfigured bool
	OpenAIBaseURL    string
	AnthropicBaseURL string
	CreatedAt        int64
	UpdatedAt        int64
	Models           []ProviderModelInfo
}

// ProviderRecord includes secret provider material and must not be serialized.
type ProviderRecord struct {
	ProviderInfo
	APIKey string
}

// ProviderModelInput carries provider-model fields accepted by create actions.
type ProviderModelInput struct {
	Name     string
	Protocol string
}

// ProviderModelInfo is the service-level provider-model projection.
type ProviderModelInfo struct {
	ID           int64
	ProviderID   int64
	Name         string
	Protocol     string
	Source       string
	LastSyncedAt *int64
	CreatedAt    int64
	UpdatedAt    int64
}

// SyncProviderModelsOutput reports one bounded remote model synchronization.
type SyncProviderModelsOutput struct {
	Protocol string
	Count    int
	Models   []ProviderModelInfo
}

// CodingImageInput carries coding-image fields accepted by create and update actions.
type CodingImageInput struct {
	Name         string
	ImageRef     string
	AgentType    string
	DefaultShell string
	Notes        string
	Enabled      bool
}

// CodingImageInfo is the service-level coding-image projection.
type CodingImageInfo struct {
	ID           int64
	Name         string
	ImageRef     string
	AgentType    string
	DefaultShell string
	Notes        string
	Enabled      bool
	IsDefault    bool
	CreatedAt    int64
	UpdatedAt    int64
}

// AgentInput carries coding-agent fields accepted by create and update actions.
type AgentInput struct {
	Name          string
	ProviderID    int64
	ModelName     string
	ModelProtocol string
	ImageID       int64
	AgentType     string
	IconKey       string
	Notes         string
}

// AgentInfo is the service-level coding-agent projection.
type AgentInfo struct {
	ID             string
	UserID         string
	Name           string
	ProviderID     int64
	ProviderName   string
	ModelName      string
	ModelProtocol  string
	ImageID        int64
	ImageName      string
	ImageRef       string
	AgentType      string
	IconKey        string
	Notes          string
	RuntimeStatus  string
	ActivityStatus string
	ContainerID    string
	DockerID       string
	DeletedAt      *int64
	CreatedAt      int64
	UpdatedAt      int64
}

// ChangeAgentImageOutput reports one coding-agent image switch.
type ChangeAgentImageOutput struct {
	Agent          AgentInfo
	LostPaths      []string
	PreservedPaths []string
}

// AgentLogsOutput carries runtime logs for one Agent when the runtime backend is available.
type AgentLogsOutput struct {
	Logs string
}

// AgentRuntimeContainerInput carries the trusted fields needed to create or
// recreate one long-lived Agent runtime container.
type AgentRuntimeContainerInput struct {
	AgentID   string
	UserID    string
	Name      string
	ImageID   int64
	ImageRef  string
	AgentType string
}

// AgentRuntimeContainerInfo is the runtime backend projection persisted into
// john_ai_agentbox_agent_runtimes after successful lifecycle actions.
type AgentRuntimeContainerInfo struct {
	ContainerID string
	DockerID    string
	Status      string
}

// AgentRuntimeLogsOutput carries backend runtime logs.
type AgentRuntimeLogsOutput struct {
	Logs string
}

// AgentRuntimeBackend is the narrow runtime dependency used by Agent lifecycle
// actions. Implementations must scope all Docker operations to plugin-owned
// labels and the authenticated AgentBox user.
type AgentRuntimeBackend interface {
	// Create starts from trusted Agent and image fields and returns the managed
	// container identity. It must not accept arbitrary Docker create options.
	CreateAgentRuntime(ctx context.Context, input AgentRuntimeContainerInput) (*AgentRuntimeContainerInfo, error)
	// Start starts an existing plugin-managed Agent runtime container.
	StartAgentRuntime(ctx context.Context, userID string, agentID string, containerID string) (*AgentRuntimeContainerInfo, error)
	// Stop stops an existing plugin-managed Agent runtime container.
	StopAgentRuntime(ctx context.Context, userID string, agentID string, containerID string) (*AgentRuntimeContainerInfo, error)
	// Logs returns recent logs from an existing plugin-managed Agent runtime container.
	AgentRuntimeLogs(ctx context.Context, userID string, agentID string, containerID string) (*AgentRuntimeLogsOutput, error)
}

// RemoteHTTPClient is the minimal HTTP dependency required for provider sync.
type RemoteHTTPClient interface {
	// Do sends one HTTP request to a remote provider.
	Do(req *http.Request) (*http.Response, error)
}

// Config contains pure value settings for the catalog service.
type Config struct {
	// RemoteModelSyncLimit caps the number of remote models persisted per sync.
	// A zero value uses the default limit.
	RemoteModelSyncLimit int
	// RemoteRequestTimeout caps outbound provider model-list requests when the
	// service owns the default HTTP client.
	RemoteRequestTimeout time.Duration
	// AgentRuntimeBackend creates and manages Agent runtime containers. A nil
	// backend keeps runtime lifecycle APIs in structured unavailable mode.
	AgentRuntimeBackend AgentRuntimeBackend
}

// Service defines AgentBox catalog persistence and remote sync behavior.
type Service interface {
	// ListProviders returns all providers with model projections. Model records
	// are loaded by provider ID set in one batch, avoiding provider-row N+1 queries.
	ListProviders(ctx context.Context) ([]ProviderInfo, error)
	// CreateProvider creates one provider configuration in plugin-owned storage.
	CreateProvider(ctx context.Context, input ProviderInput) (*ProviderInfo, error)
	// GetProvider returns one provider and its model projections.
	GetProvider(ctx context.Context, id int64) (*ProviderInfo, error)
	// UpdateProvider updates one provider configuration; an empty API key keeps
	// the previous secret value.
	UpdateProvider(ctx context.Context, id int64, input ProviderInput) (*ProviderInfo, error)
	// DeleteProvider deletes one unused provider configuration.
	DeleteProvider(ctx context.Context, id int64) error
	// CreateProviderModel creates or updates one manually managed model record.
	CreateProviderModel(ctx context.Context, providerID int64, input ProviderModelInput) (*ProviderModelInfo, error)
	// DeleteProviderModel deletes one unused provider model record.
	DeleteProviderModel(ctx context.Context, providerID int64, modelID int64) error
	// SyncProviderModels fetches remote model IDs for one provider protocol and
	// upserts them into plugin-owned provider-model storage with a bounded limit.
	SyncProviderModels(ctx context.Context, providerID int64, protocol string) (*SyncProviderModelsOutput, error)
	// ListImages returns all coding images ordered for selector presentation.
	ListImages(ctx context.Context) ([]CodingImageInfo, error)
	// CreateImage creates one coding-image profile.
	CreateImage(ctx context.Context, input CodingImageInput) (*CodingImageInfo, error)
	// UpdateImage updates one coding-image profile.
	UpdateImage(ctx context.Context, id int64, input CodingImageInput) (*CodingImageInfo, error)
	// DeleteImage deletes one unused non-default coding-image profile.
	DeleteImage(ctx context.Context, id int64) error
	// ListUserAgents returns non-deleted coding agents owned by one AgentBox user.
	ListUserAgents(ctx context.Context, userID string) ([]AgentInfo, error)
	// CreateUserAgent creates one coding agent owned by one AgentBox user.
	CreateUserAgent(ctx context.Context, userID string, input AgentInput) (*AgentInfo, error)
	// GetUserAgent returns one coding agent owned by one AgentBox user.
	GetUserAgent(ctx context.Context, userID string, agentID string) (*AgentInfo, error)
	// UpdateUserAgent updates one coding agent owned by one AgentBox user.
	UpdateUserAgent(ctx context.Context, userID string, agentID string, input AgentInput) (*AgentInfo, error)
	// SetUserAgentImage switches one coding agent's image after ownership validation.
	SetUserAgentImage(ctx context.Context, userID string, agentID string, imageID int64) (*ChangeAgentImageOutput, error)
	// StartUserAgentRuntime validates ownership, creates or starts the plugin-
	// managed Agent runtime container, and persists the Agent runtime mapping.
	StartUserAgentRuntime(ctx context.Context, userID string, agentID string) (*AgentInfo, error)
	// StopUserAgentRuntime validates ownership, stops the plugin-managed Agent
	// runtime container, and persists stopped status for the Agent.
	StopUserAgentRuntime(ctx context.Context, userID string, agentID string) (*AgentInfo, error)
	// UserAgentRuntimeLogs validates ownership and reads logs from the plugin-
	// managed Agent runtime container without leaking invisible resources.
	UserAgentRuntimeLogs(ctx context.Context, userID string, agentID string) (*AgentLogsOutput, error)
	// DeleteUserAgent soft-deletes one coding agent after ownership validation.
	DeleteUserAgent(ctx context.Context, userID string, agentID string) error
}

// serviceImpl is the default AgentBox catalog service implementation.
type serviceImpl struct {
	httpClient           RemoteHTTPClient
	remoteModelSyncLimit int
	runtimeBackend       AgentRuntimeBackend
}

var _ Service = (*serviceImpl)(nil)

// New creates the AgentBox catalog service.
func New(httpClient RemoteHTTPClient, config Config) (Service, error) {
	if httpClient == nil {
		timeout := config.RemoteRequestTimeout
		if timeout == 0 {
			timeout = 20 * time.Second
		}
		if timeout < time.Second {
			return nil, gerror.New("agentbox catalog remote request timeout must be at least one second")
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	limit := config.RemoteModelSyncLimit
	if limit == 0 {
		limit = defaultRemoteModelSyncLimit
	}
	if limit < 1 {
		return nil, gerror.New("agentbox catalog remote model sync limit must be positive")
	}
	return &serviceImpl{
		httpClient:           httpClient,
		remoteModelSyncLimit: limit,
		runtimeBackend:       config.AgentRuntimeBackend,
	}, nil
}
