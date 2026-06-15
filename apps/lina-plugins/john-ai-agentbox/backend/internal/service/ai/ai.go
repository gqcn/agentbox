// Package ai owns AgentBox AI capability tiers, provider-model bindings,
// provider connectivity tests, and sanitized invocation logs. All persisted
// state lives in plugin-owned john_ai_agentbox_* tables.
package ai

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// TierBasic is the low-cost tier for simple text generation.
	TierBasic = "basic"
	// TierStandard is the default tier for moderate coding tasks.
	TierStandard = "standard"
	// TierAdvanced is the high-capability tier for complex coding work.
	TierAdvanced = "advanced"

	// PurposeCapabilityTest marks lightweight tier connectivity tests.
	PurposeCapabilityTest = "capability_test"
	// PurposeGitCommitMessage marks Git commit message suggestion requests.
	PurposeGitCommitMessage = "git_commit_message"

	// InvocationStatusSuccess marks a completed AI invocation.
	InvocationStatusSuccess = "success"
	// InvocationStatusError marks a failed AI invocation.
	InvocationStatusError = "error"
)

const (
	defaultRequestTimeout  = 30 * time.Second
	defaultLogLimit        = 50
	maxLogLimit            = 200
	primaryBindingPriority = 0
)

// Config contains pure value settings for the AI service.
type Config struct {
	// RequestTimeout caps outbound provider test requests. A zero value uses
	// the service default.
	RequestTimeout time.Duration
}

// HTTPClient is the minimal dependency required for provider connectivity tests.
type HTTPClient interface {
	// Do sends one HTTP request to a remote provider.
	Do(req *http.Request) (*http.Response, error)
}

// CapabilityTierInfo describes one fixed AgentBox AI capability tier.
type CapabilityTierInfo struct {
	Code        string
	DisplayName string
	Description string
	Enabled     bool
	Configured  bool
	Available   bool
	Binding     *CapabilityBindingInfo
	LastTest    *InvocationLogInfo
	CreatedAt   int64
	UpdatedAt   int64
}

// CapabilityBindingInfo describes a tier-to-provider-model binding.
type CapabilityBindingInfo struct {
	ID              int64
	TierCode        string
	ProviderID      int64
	ProviderName    string
	ProviderModelID int64
	ModelName       string
	Protocol        string
	Priority        int
	Enabled         bool
	CreatedAt       int64
	UpdatedAt       int64
}

// UpdateTierInput carries editable fields for one tier.
type UpdateTierInput struct {
	Enabled         bool
	ProviderID      int64
	ProviderModelID int64
	Protocol        string
}

// TestTierInput carries optional draft provider/model fields for connectivity tests.
type TestTierInput struct {
	ProviderID      int64
	ProviderModelID int64
	Protocol        string
}

// InvocationLogFilter carries optional query filters for invocation logs.
type InvocationLogFilter struct {
	Purpose  string
	TierCode string
	Status   string
	Limit    int
}

// InvocationLogInfo exposes the minimal audit record for one AgentBox AI call.
type InvocationLogInfo struct {
	ID              int64
	Purpose         string
	TierCode        string
	ProviderID      int64
	ProviderName    string
	ProviderModelID int64
	ModelName       string
	Protocol        string
	Status          string
	LatencyMS       int64
	ErrorMessage    string
	CreatedAt       int64
}

// CapabilityTestResult returns the result of a lightweight tier test.
type CapabilityTestResult struct {
	Status          string
	TierCode        string
	ProviderID      int64
	ProviderName    string
	ProviderModelID int64
	ModelName       string
	Protocol        string
	LatencyMS       int64
	ErrorMessage    string
	TestedAt        int64
}

// ResolvedBinding contains binding data and secret provider material needed by tests.
type ResolvedBinding struct {
	Tier     CapabilityTierInfo
	Binding  CapabilityBindingInfo
	Provider providerRecord
	Model    providerModelRecord
}

// Service manages AgentBox AI capability configuration and provider tests.
type Service interface {
	// ListTiers returns all fixed tiers with their current binding and last test.
	ListTiers(ctx context.Context) ([]CapabilityTierInfo, error)
	// UpdateTier updates one tier's enabled flag and primary provider/model binding.
	UpdateTier(ctx context.Context, code string, input UpdateTierInput) (*CapabilityTierInfo, error)
	// TestTier runs a lightweight generation call against a saved tier binding
	// or an explicitly supplied draft provider/model pair. Draft tests validate
	// provider/model ownership but do not persist tier bindings.
	TestTier(ctx context.Context, code string, input TestTierInput) (*CapabilityTestResult, error)
	// ListInvocations returns sanitized AI invocation logs with optional filters.
	ListInvocations(ctx context.Context, filter InvocationLogFilter) ([]InvocationLogInfo, error)
}

// serviceImpl is the default AgentBox AI service implementation.
type serviceImpl struct {
	httpClient HTTPClient
	timeout    time.Duration
}

var _ Service = (*serviceImpl)(nil)

// New creates the AgentBox AI service.
func New(httpClient HTTPClient, config Config) (Service, error) {
	timeout := config.RequestTimeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}
	if timeout < time.Second {
		return nil, gerror.New("agentbox ai request timeout must be at least one second")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &serviceImpl{
		httpClient: httpClient,
		timeout:    timeout,
	}, nil
}
