// This file verifies agent controller ownership plumbing. The controller must
// use the authenticated AgentBox user from context for every user-scoped
// service call.

package agent

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/agent/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
	chatsvc "john-ai-agentbox/backend/internal/service/chat"
	terminalsvc "john-ai-agentbox/backend/internal/service/terminal"
)

// TestAgentListUsesAuthenticatedUser verifies list calls are scoped to auth context.
func TestAgentListUsesAuthenticatedUser(t *testing.T) {
	service := &fakeAgentCatalogService{
		agents: []catalogsvc.AgentInfo{{
			ID:     "agt-test",
			UserID: "usr-owner",
			Name:   "Owned",
		}},
	}
	controller := newTestController(t, service, &fakeChatService{}, &fakeTerminalService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.List(ctx, &v1.ListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" {
		t.Fatalf("expected user scoped list for usr-owner, got %q", service.lastUserID)
	}
	if len(*res) != 1 || (*res)[0].ID != "agt-test" {
		t.Fatalf("unexpected agent response: %#v", *res)
	}
}

// TestAgentDetailRequiresAuthenticatedUser verifies missing auth context is rejected.
func TestAgentDetailRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeAgentCatalogService{}, &fakeChatService{}, &fakeTerminalService{})

	_, err := controller.Detail(context.Background(), &v1.DetailReq{ID: "agt-test"})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

// TestAgentStartUsesAuthenticatedUser verifies start calls are scoped before runtime handling.
func TestAgentStartUsesAuthenticatedUser(t *testing.T) {
	service := &fakeAgentCatalogService{
		err: bizerr.NewCode(catalogsvc.CodeCatalogRuntimeUnavailable),
	}
	controller := newTestController(t, service, &fakeChatService{}, &fakeTerminalService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	_, err := controller.Start(ctx, &v1.StartReq{ID: "agt-test"})
	if !bizerr.Is(err, catalogsvc.CodeCatalogRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" {
		t.Fatalf("expected start scoped to usr-owner/agt-test, got %q/%q", service.lastUserID, service.lastAgentID)
	}
}

// TestAgentStopUsesAuthenticatedUser verifies stop calls are scoped before runtime handling.
func TestAgentStopUsesAuthenticatedUser(t *testing.T) {
	service := &fakeAgentCatalogService{
		err: bizerr.NewCode(catalogsvc.CodeCatalogRuntimeUnavailable),
	}
	controller := newTestController(t, service, &fakeChatService{}, &fakeTerminalService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	_, err := controller.Stop(ctx, &v1.StopReq{ID: "agt-test"})
	if !bizerr.Is(err, catalogsvc.CodeCatalogRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" {
		t.Fatalf("expected stop scoped to usr-owner/agt-test, got %q/%q", service.lastUserID, service.lastAgentID)
	}
}

// TestAgentStopRequiresAuthenticatedUser verifies missing auth is rejected before runtime handling.
func TestAgentStopRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeAgentCatalogService{}, &fakeChatService{}, &fakeTerminalService{})

	_, err := controller.Stop(context.Background(), &v1.StopReq{ID: "agt-test"})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

// TestAgentLogsUsesAuthenticatedUser verifies runtime logs are scoped before runtime handling.
func TestAgentLogsUsesAuthenticatedUser(t *testing.T) {
	service := &fakeAgentCatalogService{
		err: bizerr.NewCode(catalogsvc.CodeCatalogRuntimeUnavailable),
	}
	controller := newTestController(t, service, &fakeChatService{}, &fakeTerminalService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	_, err := controller.Logs(ctx, &v1.LogsReq{ID: "agt-test"})
	if !bizerr.Is(err, catalogsvc.CodeCatalogRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" {
		t.Fatalf("expected logs scoped to usr-owner/agt-test, got %q/%q", service.lastUserID, service.lastAgentID)
	}
}

// TestChatSessionsUseAuthenticatedUser verifies Chat handlers are scoped to auth context.
func TestChatSessionsUseAuthenticatedUser(t *testing.T) {
	chatService := &fakeChatService{
		sessions: []chatsvc.SessionInfo{{
			ID:                 "chat-test",
			AgentID:            "agt-test",
			Title:              "新对话",
			Status:             chatsvc.SessionStatusIdle,
			RuntimeState:       chatsvc.RuntimeStateIdle,
			LastMessagePreview: "",
			CreatedAt:          1718000000000,
			UpdatedAt:          1718000001000,
			LastActiveAt:       1718000002000,
		}},
	}
	controller := newTestController(t, &fakeAgentCatalogService{}, chatService, &fakeTerminalService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.ChatSessions(ctx, &v1.ChatSessionsReq{ID: "agt-test"})
	if err != nil {
		t.Fatal(err)
	}
	if chatService.lastUserID != "usr-owner" || chatService.lastAgentID != "agt-test" {
		t.Fatalf("expected chat scoped to usr-owner/agt-test, got %q/%q", chatService.lastUserID, chatService.lastAgentID)
	}
	if len(*res) != 1 || (*res)[0].ID != "chat-test" {
		t.Fatalf("unexpected chat sessions response: %#v", *res)
	}
}

// TestChatMessagesRequiresAuthenticatedUser verifies missing auth is rejected.
func TestChatMessagesRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeAgentCatalogService{}, &fakeChatService{}, &fakeTerminalService{})

	_, err := controller.ChatMessages(context.Background(), &v1.ChatMessagesReq{ID: "agt-test", SessionID: "chat-test"})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

// TestTerminalSessionsUseAuthenticatedUser verifies terminal handlers are scoped to auth context.
func TestTerminalSessionsUseAuthenticatedUser(t *testing.T) {
	terminalService := &fakeTerminalService{
		sessions: []terminalsvc.SessionInfo{{
			ID:          "term-row",
			AgentID:     "agt-test",
			TerminalID:  "term-test",
			BackendType: terminalsvc.BackendTypeTmux,
			WorkingDir:  terminalsvc.DefaultWorkingDir,
			Shell:       terminalsvc.DefaultShell,
			Status:      terminalsvc.SessionStatusActive,
			CreatedAt:   1718000000000,
			UpdatedAt:   1718000001000,
		}},
	}
	controller := newTestController(t, &fakeAgentCatalogService{}, &fakeChatService{}, terminalService)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.TerminalSessions(ctx, &v1.TerminalSessionsReq{ID: "agt-test"})
	if err != nil {
		t.Fatal(err)
	}
	if terminalService.lastUserID != "usr-owner" || terminalService.lastAgentID != "agt-test" {
		t.Fatalf("expected terminal scoped to usr-owner/agt-test, got %q/%q", terminalService.lastUserID, terminalService.lastAgentID)
	}
	if len(*res) != 1 || (*res)[0].TerminalID != "term-test" {
		t.Fatalf("unexpected terminal sessions response: %#v", *res)
	}
}

func newTestController(t *testing.T, catalogSvc catalogsvc.Service, chatSvc chatsvc.Service, terminalSvc terminalsvc.Service) *ControllerV1 {
	t.Helper()
	controller, err := NewV1(catalogSvc, chatSvc, terminalSvc)
	if err != nil {
		t.Fatal(err)
	}
	typed, ok := controller.(*ControllerV1)
	if !ok {
		t.Fatalf("unexpected controller type %T", controller)
	}
	return typed
}

type fakeAgentCatalogService struct {
	catalogsvc.Service
	lastUserID  string
	lastAgentID string
	agents      []catalogsvc.AgentInfo
	changed     *catalogsvc.ChangeAgentImageOutput
	err         error
}

func (s *fakeAgentCatalogService) ListUserAgents(_ context.Context, userID string) ([]catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	if s.err != nil {
		return nil, s.err
	}
	return s.agents, nil
}

func (s *fakeAgentCatalogService) CreateUserAgent(_ context.Context, userID string, _ catalogsvc.AgentInput) (*catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	if s.err != nil {
		return nil, s.err
	}
	return &s.agents[0], nil
}

func (s *fakeAgentCatalogService) GetUserAgent(_ context.Context, userID string, agentID string) (*catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return &s.agents[0], nil
}

func (s *fakeAgentCatalogService) UpdateUserAgent(_ context.Context, userID string, agentID string, _ catalogsvc.AgentInput) (*catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return &s.agents[0], nil
}

func (s *fakeAgentCatalogService) SetUserAgentImage(_ context.Context, userID string, agentID string, _ int64) (*catalogsvc.ChangeAgentImageOutput, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return s.changed, nil
}

func (s *fakeAgentCatalogService) StartUserAgentRuntime(_ context.Context, userID string, agentID string) (*catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return &s.agents[0], nil
}

func (s *fakeAgentCatalogService) StopUserAgentRuntime(_ context.Context, userID string, agentID string) (*catalogsvc.AgentInfo, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return &s.agents[0], nil
}

func (s *fakeAgentCatalogService) UserAgentRuntimeLogs(_ context.Context, userID string, agentID string) (*catalogsvc.AgentLogsOutput, error) {
	s.lastUserID = userID
	s.lastAgentID = agentID
	if s.err != nil {
		return nil, s.err
	}
	return &catalogsvc.AgentLogsOutput{}, nil
}

func (s *fakeAgentCatalogService) DeleteUserAgent(_ context.Context, userID string, agentID string) error {
	s.lastUserID = userID
	s.lastAgentID = agentID
	return s.err
}

type fakeChatService struct {
	chatsvc.Service
	lastUserID        string
	lastAgentID       string
	lastSessionID     string
	lastInteractionID string
	sessions          []chatsvc.SessionInfo
	messages          *chatsvc.MessagesOutput
	interactions      []chatsvc.InteractionInfo
	interaction       *chatsvc.InteractionInfo
	recover           *chatsvc.RecoverOutput
	err               error
}

func (s *fakeChatService) ListSessions(_ context.Context, userID string, agentID string) ([]chatsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID = userID, agentID
	if s.err != nil {
		return nil, s.err
	}
	return s.sessions, nil
}

func (s *fakeChatService) CreateSession(_ context.Context, userID string, agentID string) (*chatsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID = userID, agentID
	if s.err != nil {
		return nil, s.err
	}
	return &s.sessions[0], nil
}

func (s *fakeChatService) GetSession(_ context.Context, userID string, agentID string, sessionID string) (*chatsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	if s.err != nil {
		return nil, s.err
	}
	return &s.sessions[0], nil
}

func (s *fakeChatService) UpdateSessionTitle(_ context.Context, userID string, agentID string, sessionID string, _ string) (*chatsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	if s.err != nil {
		return nil, s.err
	}
	return &s.sessions[0], nil
}

func (s *fakeChatService) DeleteSession(_ context.Context, userID string, agentID string, sessionID string) error {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	return s.err
}

func (s *fakeChatService) Messages(_ context.Context, userID string, agentID string, sessionID string) (*chatsvc.MessagesOutput, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	if s.err != nil {
		return nil, s.err
	}
	return s.messages, nil
}

func (s *fakeChatService) ListInteractions(_ context.Context, userID string, agentID string, sessionID string, _ chatsvc.InteractionFilter) ([]chatsvc.InteractionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	if s.err != nil {
		return nil, s.err
	}
	return s.interactions, nil
}

func (s *fakeChatService) GetInteraction(_ context.Context, userID string, agentID string, sessionID string, interactionID string) (*chatsvc.InteractionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID, s.lastInteractionID = userID, agentID, sessionID, interactionID
	if s.err != nil {
		return nil, s.err
	}
	return s.interaction, nil
}

func (s *fakeChatService) UpdateInteractionResponse(_ context.Context, userID string, agentID string, sessionID string, interactionID string, _ chatsvc.InteractionResponseInput) (*chatsvc.InteractionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID, s.lastInteractionID = userID, agentID, sessionID, interactionID
	if s.err != nil {
		return nil, s.err
	}
	return s.interaction, nil
}

func (s *fakeChatService) UpdateInteractionStatus(_ context.Context, userID string, agentID string, sessionID string, interactionID string, _ chatsvc.InteractionStatusInput) (*chatsvc.InteractionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID, s.lastInteractionID = userID, agentID, sessionID, interactionID
	if s.err != nil {
		return nil, s.err
	}
	return s.interaction, nil
}

func (s *fakeChatService) Recover(_ context.Context, userID string, agentID string, sessionID string) (*chatsvc.RecoverOutput, error) {
	s.lastUserID, s.lastAgentID, s.lastSessionID = userID, agentID, sessionID
	if s.err != nil {
		return nil, s.err
	}
	return s.recover, nil
}

type fakeTerminalService struct {
	terminalsvc.Service
	lastUserID     string
	lastAgentID    string
	lastTerminalID string
	lastInput      terminalsvc.SessionInput
	sessions       []terminalsvc.SessionInfo
	session        *terminalsvc.SessionInfo
	err            error
}

func (s *fakeTerminalService) ListSessions(_ context.Context, userID string, agentID string, _ terminalsvc.SessionFilter) ([]terminalsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID = userID, agentID
	if s.err != nil {
		return nil, s.err
	}
	return s.sessions, nil
}

func (s *fakeTerminalService) EnsureSession(_ context.Context, userID string, agentID string, input terminalsvc.SessionInput) (*terminalsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastTerminalID, s.lastInput = userID, agentID, input.TerminalID, input
	if s.err != nil {
		return nil, s.err
	}
	return s.session, nil
}

func (s *fakeTerminalService) GetSession(_ context.Context, userID string, agentID string, terminalID string) (*terminalsvc.SessionInfo, error) {
	s.lastUserID, s.lastAgentID, s.lastTerminalID = userID, agentID, terminalID
	if s.err != nil {
		return nil, s.err
	}
	return s.session, nil
}

func (s *fakeTerminalService) CloseSession(_ context.Context, userID string, agentID string, terminalID string) error {
	s.lastUserID, s.lastAgentID, s.lastTerminalID = userID, agentID, terminalID
	return s.err
}
