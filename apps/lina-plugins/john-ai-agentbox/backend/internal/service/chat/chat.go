// Package chat owns AgentBox Chat session, message, and interaction
// persistence. It keeps Chat resources in plugin-owned tables and validates
// every entry point against the current AgentBox user boundary before exposing
// session history or interaction state.
package chat

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

const (
	// SessionStatusIdle marks a Chat session with no active runtime work.
	SessionStatusIdle = "idle"
	// SessionStatusRunning marks a Chat session currently processing.
	SessionStatusRunning = "running"
	// SessionStatusWaiting marks a Chat session waiting for user input.
	SessionStatusWaiting = "waiting_input"
	// SessionStatusExited marks a Chat session whose runtime exited.
	SessionStatusExited = "exited"
	// SessionStatusRecovering marks a Chat session being recovered.
	SessionStatusRecovering = "recovering"
	// SessionStatusError marks a Chat session with a runtime failure.
	SessionStatusError = "error"

	// RuntimeStateIdle marks a detached or idle Chat runtime.
	RuntimeStateIdle = "idle"
	// RuntimeStateRunning marks an active Chat runtime.
	RuntimeStateRunning = "running"
	// RuntimeStateWaiting marks a runtime waiting for an interaction response.
	RuntimeStateWaiting = "waiting_input"
	// RuntimeStateExited marks a runtime that has exited.
	RuntimeStateExited = "exited"
	// RuntimeStateRecovering marks a runtime recovery attempt.
	RuntimeStateRecovering = "recovering"
	// RuntimeStateError marks a runtime error state.
	RuntimeStateError = "error"

	// MessageRoleUser identifies messages submitted by the user.
	MessageRoleUser = "user"
	// MessageRoleAssistant identifies assistant responses.
	MessageRoleAssistant = "assistant"
	// MessageRoleSystem identifies system notices.
	MessageRoleSystem = "system"
	// MessageRoleError identifies runtime error messages.
	MessageRoleError = "error"
	// MessageRoleTerminal identifies terminal output messages.
	MessageRoleTerminal = "terminal"

	// MessageStatusStreaming marks an in-progress message.
	MessageStatusStreaming = "streaming"
	// MessageStatusComplete marks a complete message.
	MessageStatusComplete = "complete"
	// MessageStatusError marks a failed message.
	MessageStatusError = "error"

	// InteractionTypePermission identifies a permission request.
	InteractionTypePermission = "permission"
	// InteractionTypeQuestion identifies a direct question.
	InteractionTypeQuestion = "question"
	// InteractionTypeChoice identifies an option choice request.
	InteractionTypeChoice = "choice"
	// InteractionTypeText identifies a free-text request.
	InteractionTypeText = "text"
	// InteractionTypeAuth identifies an authentication request.
	InteractionTypeAuth = "auth"
	// InteractionTypePlan identifies a plan confirmation request.
	InteractionTypePlan = "plan"
	// InteractionTypeCustom identifies a tool-specific interaction.
	InteractionTypeCustom = "custom"

	// InteractionStatusPending marks an unresolved interaction.
	InteractionStatusPending = "pending"
	// InteractionStatusResolved marks a resolved interaction.
	InteractionStatusResolved = "resolved"
	// InteractionStatusRejected marks a rejected interaction.
	InteractionStatusRejected = "rejected"
	// InteractionStatusCancelled marks a cancelled interaction.
	InteractionStatusCancelled = "cancelled"
	// InteractionStatusExpired marks an expired interaction.
	InteractionStatusExpired = "expired"
	// InteractionStatusError marks an interaction error.
	InteractionStatusError = "error"

	// InteractionRiskLow identifies low-risk interactions.
	InteractionRiskLow = "low"
	// InteractionRiskMedium identifies medium-risk interactions.
	InteractionRiskMedium = "medium"
	// InteractionRiskHigh identifies high-risk interactions.
	InteractionRiskHigh = "high"
	// InteractionRiskCritical identifies critical-risk interactions.
	InteractionRiskCritical = "critical"

	// InteractionResponseScopeOnce applies the response once.
	InteractionResponseScopeOnce = "once"
	// InteractionResponseScopeSession applies the response to a session.
	InteractionResponseScopeSession = "session"
	// InteractionResponseScopeAgent applies the response to an agent.
	InteractionResponseScopeAgent = "agent"
	// InteractionResponseScopeProvider applies the response to a provider.
	InteractionResponseScopeProvider = "provider"

	// InteractionResponseModeAllow records an allow response.
	InteractionResponseModeAllow = "allow"
	// InteractionResponseModeAnswer records an answer response.
	InteractionResponseModeAnswer = "answer"
	// InteractionResponseModeReject records a reject response.
	InteractionResponseModeReject = "reject"
	// InteractionResponseModeCancel records a cancel response.
	InteractionResponseModeCancel = "cancel"
	// InteractionResponseModeAllowOnce records a one-time allow response.
	InteractionResponseModeAllowOnce = "allow_once"
	// InteractionResponseModeAllowSession records a session-scoped allow response.
	InteractionResponseModeAllowSession = "allow_session"
)

// SessionInfo is the service-level Chat session projection.
type SessionInfo struct {
	ID                 string
	AgentID            string
	Title              string
	Status             string
	ToolType           string
	ToolSessionID      string
	RuntimeState       string
	LastError          string
	MessageCount       int64
	LastMessagePreview string
	CreatedAt          int64
	UpdatedAt          int64
	LastActiveAt       int64
}

// MessageInfo is the service-level persisted Chat message projection.
type MessageInfo struct {
	ID        int64
	SessionID string
	Sequence  int64
	Role      string
	Content   string
	Status    string
	Metadata  string
	CreatedAt int64
	UpdatedAt int64
}

// MessagesOutput returns a Chat session and its bounded message history.
type MessagesOutput struct {
	Session  SessionInfo
	Messages []MessageInfo
}

// InteractionInfo is the service-level Chat interaction projection.
type InteractionInfo struct {
	ID                 string
	AgentID            string
	SessionID          string
	AssistantMessageID int64
	ToolType           string
	ToolInteractionID  string
	Type               string
	Status             string
	Title              string
	Body               string
	RiskLevel          string
	Payload            string
	Response           string
	ResponseMode       string
	ResponseScope      string
	ExpiresAt          *int64
	ResolvedAt         *int64
	CreatedAt          int64
	UpdatedAt          int64
}

// InteractionFilter carries optional interaction list filters. Empty fields
// mean no filter for that dimension.
type InteractionFilter struct {
	Status string
	Type   string
}

// InteractionResponseInput carries a user response for one pending interaction.
type InteractionResponseInput struct {
	Response      string
	ResponseMode  string
	ResponseScope string
}

// InteractionStatusInput carries a requested terminal status update.
type InteractionStatusInput struct {
	Status string
}

// RecoverOutput returns recovery state once runtime migration is available.
type RecoverOutput struct {
	Session *SessionInfo
	Message *MessageInfo
}

// Service defines AgentBox Chat persistence behavior. Every method accepts the
// current AgentBox user ID and validates resource ownership before loading or
// mutating Chat records; invisible resources return not-found semantics.
type Service interface {
	// ListSessions returns all sessions for one current-user-owned Agent,
	// ordered by recent activity. It returns an empty slice for Agents with no
	// sessions and propagates structured access errors for invisible Agents.
	ListSessions(ctx context.Context, userID string, agentID string) ([]SessionInfo, error)
	// CreateSession creates one empty Chat session for a current-user-owned
	// Agent. The session tool type defaults to the Agent type.
	CreateSession(ctx context.Context, userID string, agentID string) (*SessionInfo, error)
	// GetSession returns one current-user-visible Chat session.
	GetSession(ctx context.Context, userID string, agentID string, sessionID string) (*SessionInfo, error)
	// UpdateSessionTitle updates one visible Chat session title.
	UpdateSessionTitle(ctx context.Context, userID string, agentID string, sessionID string, title string) (*SessionInfo, error)
	// DeleteSession deletes one visible idle Chat session and cascades persisted
	// messages and interactions through the plugin table foreign keys.
	DeleteSession(ctx context.Context, userID string, agentID string, sessionID string) error
	// Messages returns one visible Chat session and its persisted messages in
	// sequence order, without exposing other users' session metadata.
	Messages(ctx context.Context, userID string, agentID string, sessionID string) (*MessagesOutput, error)
	// ListInteractions returns visible interactions for one Chat session with
	// optional status and type filters applied in the database query.
	ListInteractions(ctx context.Context, userID string, agentID string, sessionID string, filter InteractionFilter) ([]InteractionInfo, error)
	// GetInteraction returns one visible interaction for a Chat session.
	GetInteraction(ctx context.Context, userID string, agentID string, sessionID string, interactionID string) (*InteractionInfo, error)
	// UpdateInteractionResponse stores a user response for one pending visible
	// interaction and returns the updated persisted state.
	UpdateInteractionResponse(ctx context.Context, userID string, agentID string, sessionID string, interactionID string, input InteractionResponseInput) (*InteractionInfo, error)
	// UpdateInteractionStatus updates a pending visible interaction to an
	// allowed terminal status such as cancelled, expired, or error.
	UpdateInteractionStatus(ctx context.Context, userID string, agentID string, sessionID string, interactionID string, input InteractionStatusInput) (*InteractionInfo, error)
	// Recover validates session visibility and starts runtime recovery when the
	// runtime service is available. Until WebSocket/runtime migration lands, it
	// returns a structured dependency error.
	Recover(ctx context.Context, userID string, agentID string, sessionID string) (*RecoverOutput, error)
}

// serviceImpl is the default DAO-backed Chat service implementation.
type serviceImpl struct {
	accessSvc accesssvc.Service
}

var _ Service = (*serviceImpl)(nil)

// New creates a Chat service using explicit access-service dependency injection.
func New(accessSvc accesssvc.Service) (Service, error) {
	if accessSvc == nil {
		return nil, gerror.New("agentbox chat access service is required")
	}
	return &serviceImpl{accessSvc: accessSvc}, nil
}
