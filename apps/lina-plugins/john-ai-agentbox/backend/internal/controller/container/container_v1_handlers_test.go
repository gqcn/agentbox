// This file verifies container controller authentication plumbing. Handlers
// must use the authenticated AgentBox user from context and reject calls
// without plugin authentication context.

package container

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/container/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
	containersvc "john-ai-agentbox/backend/internal/service/container"
)

// TestListUsesAuthenticatedUser verifies container list calls are user scoped.
func TestListUsesAuthenticatedUser(t *testing.T) {
	service := &fakeContainerService{
		containers: []containersvc.ContainerInfo{{
			ID:        "ctr-owned",
			Name:      "owned",
			CreatedAt: 1704067200000,
		}},
	}
	controller := newTestController(t, service)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.List(ctx, &v1.ListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" {
		t.Fatalf("unexpected user id %q", service.lastUserID)
	}
	if len(*res) != 1 || (*res)[0].ID != "ctr-owned" {
		t.Fatalf("unexpected list response: %#v", *res)
	}
}

// TestDockerHealthRequiresAuthenticatedUser verifies missing auth is rejected.
func TestDockerHealthRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeContainerService{})

	_, err := controller.DockerHealth(context.Background(), &v1.DockerHealthReq{})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

func newTestController(t *testing.T, containerSvc containersvc.Service) *ControllerV1 {
	t.Helper()
	controller, err := NewV1(containerSvc)
	if err != nil {
		t.Fatal(err)
	}
	typed, ok := controller.(*ControllerV1)
	if !ok {
		t.Fatalf("unexpected controller type %T", controller)
	}
	return typed
}

type fakeContainerService struct {
	containersvc.Service
	lastUserID string
	containers []containersvc.ContainerInfo
	err        error
}

func (s *fakeContainerService) DockerHealth(_ context.Context, userID string) (*containersvc.DockerHealthResponse, error) {
	s.lastUserID = userID
	if s.err != nil {
		return nil, s.err
	}
	return &containersvc.DockerHealthResponse{OK: true}, nil
}

func (s *fakeContainerService) List(_ context.Context, userID string) ([]containersvc.ContainerInfo, error) {
	s.lastUserID = userID
	if s.err != nil {
		return nil, s.err
	}
	return s.containers, nil
}
