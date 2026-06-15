// This file verifies AgentBox container runtime service boundaries. Runtime
// data is intentionally unavailable in this migration slice, but callers must
// still provide authenticated AgentBox user context before runtime state is considered.

package container

import (
	"context"
	"errors"
	"strings"
	"testing"

	dockercontainer "github.com/moby/moby/api/types/container"

	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"

	"lina-core/pkg/bizerr"
)

// TestContainerServiceRequiresUser verifies runtime actions reject missing users.
func TestContainerServiceRequiresUser(t *testing.T) {
	lifecycle := &fakeContainerLifecycleBackend{}
	service, err := New(&fakeDockerHealthBackend{}, lifecycle)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.List(context.Background(), "")
	if !bizerr.Is(err, CodeContainerInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
	if lifecycle.listCalls != 0 {
		t.Fatalf("expected lifecycle backend not to be called, got %d calls", lifecycle.listCalls)
	}
}

// TestDockerHealthRequiresUserBeforeBackend verifies Docker cannot be probed without user context.
func TestDockerHealthRequiresUserBeforeBackend(t *testing.T) {
	backend := &fakeDockerHealthBackend{}
	service, err := New(backend, &fakeContainerLifecycleBackend{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.DockerHealth(context.Background(), "")
	if !bizerr.Is(err, CodeContainerInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
	if backend.calls != 0 {
		t.Fatalf("expected backend not to be called, got %d calls", backend.calls)
	}
}

// TestContainerRuntimeUnavailable verifies visible runtime calls do not fake Docker state.
func TestContainerRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeDockerHealthBackend{err: errors.New("daemon down")}, &fakeContainerLifecycleBackend{
		err: errors.New("daemon down"),
	})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		run  func() error
	}{
		{
			name: "health",
			run: func() error {
				_, err := service.DockerHealth(context.Background(), "usr-owner")
				return err
			},
		},
		{
			name: "list",
			run: func() error {
				_, err := service.List(context.Background(), "usr-owner")
				return err
			},
		},
		{
			name: "detail",
			run: func() error {
				_, err := service.Detail(context.Background(), "usr-owner", "ctr-owned")
				return err
			},
		},
		{
			name: "logs",
			run: func() error {
				_, err := service.Logs(context.Background(), "usr-owner", "ctr-owned")
				return err
			},
		},
	}
	for _, check := range checks {
		if err := check.run(); !bizerr.Is(err, CodeContainerRuntimeUnavailable) {
			t.Fatalf("%s: expected runtime unavailable error, got %v", check.name, err)
		}
	}
}

// TestDockerHealthUsesRuntimeBackend verifies Docker health can return runtime metadata.
func TestDockerHealthUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeDockerHealthBackend{
		response: &DockerHealthResponse{
			OK:         true,
			APIVersion: "1.45",
			OSType:     "linux",
		},
	}, &fakeContainerLifecycleBackend{})
	if err != nil {
		t.Fatal(err)
	}

	item, err := service.DockerHealth(context.Background(), "usr-owner")
	if err != nil {
		t.Fatal(err)
	}
	if item.APIVersion != "1.45" || item.OSType != "linux" {
		t.Fatalf("unexpected health response: %#v", item)
	}
}

// TestContainerLifecycleUsesScopedBackend verifies label-scoped lifecycle calls
// are delegated only after user and container identifiers are validated.
func TestContainerLifecycleUsesScopedBackend(t *testing.T) {
	lifecycle := &fakeContainerLifecycleBackend{
		items: []ContainerInfo{{
			ID:       "ctr-owned",
			Name:     "owned",
			DockerID: "docker-owned",
		}},
		item: &ContainerInfo{
			ID:       "ctr-owned",
			Name:     "owned",
			DockerID: "docker-owned",
		},
		logs: &LogsResponse{Logs: "hello"},
	}
	service, err := New(&fakeDockerHealthBackend{}, lifecycle)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	items, err := service.List(ctx, "usr-owner")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "ctr-owned" {
		t.Fatalf("unexpected list response: %#v", items)
	}
	if lifecycle.lastUserID != "usr-owner" {
		t.Fatalf("unexpected list user %q", lifecycle.lastUserID)
	}

	item, err := service.Detail(ctx, "usr-owner", "ctr-owned")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "ctr-owned" || lifecycle.lastContainerID != "ctr-owned" {
		t.Fatalf("unexpected detail response item=%#v backend id=%q", item, lifecycle.lastContainerID)
	}

	if _, err := service.Start(ctx, "usr-owner", "ctr-owned"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Stop(ctx, "usr-owner", "ctr-owned"); err != nil {
		t.Fatal(err)
	}
	deleted, err := service.Delete(ctx, "usr-owner", "ctr-owned")
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatalf("expected delete result true")
	}
	logs, err := service.Logs(ctx, "usr-owner", "ctr-owned")
	if err != nil {
		t.Fatal(err)
	}
	if logs.Logs != "hello" {
		t.Fatalf("unexpected logs response: %#v", logs)
	}
}

// TestContainerCreateRemainsUnavailable verifies creation is still blocked
// until trusted runtime image and configuration migration is complete.
func TestContainerCreateRemainsUnavailable(t *testing.T) {
	service, err := New(&fakeDockerHealthBackend{}, &fakeContainerLifecycleBackend{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Create(context.Background(), "usr-owner", "demo")
	if !bizerr.Is(err, CodeContainerRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
}

// TestManagedContainerLabelScopeRejectsOtherUsers verifies helper predicates do
// not treat unlabelled or other-user Docker containers as AgentBox resources.
func TestManagedContainerLabelScopeRejectsOtherUsers(t *testing.T) {
	owned := map[string]string{
		containerLabelManaged:     containerManagedValue,
		containerLabelUser:        "usr-owner",
		containerLabelContainerID: "ctr-owned",
		containerLabelName:        "owned",
	}
	if !isIndependentManagedContainer(owned, "usr-owner") {
		t.Fatalf("expected owned labels to be visible")
	}
	for name, labels := range map[string]map[string]string{
		"other user": {
			containerLabelManaged:     containerManagedValue,
			containerLabelUser:        "usr-other",
			containerLabelContainerID: "ctr-secret",
		},
		"missing logical id": {
			containerLabelManaged: containerManagedValue,
			containerLabelUser:    "usr-owner",
		},
		"unmanaged": {
			containerLabelUser:        "usr-owner",
			containerLabelContainerID: "ctr-owned",
		},
	} {
		if isIndependentManagedContainer(labels, "usr-owner") {
			t.Fatalf("%s labels unexpectedly visible: %#v", name, labels)
		}
	}
}

// TestContainerNotFoundDoesNotLeakIdentifier verifies invisible-container
// errors stay stable and do not include the requested Docker or logical ID.
func TestContainerNotFoundDoesNotLeakIdentifier(t *testing.T) {
	lifecycle := &fakeContainerLifecycleBackend{err: newContainerNotFoundError()}
	service, err := New(&fakeDockerHealthBackend{}, lifecycle)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Detail(context.Background(), "usr-owner", "ctr-secret")
	if !bizerr.Is(err, CodeContainerNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "ctr-secret") {
		t.Fatalf("container error leaked requested id: %v", err)
	}
}

// TestContainerProjectionFromDockerSummary verifies Docker summaries are mapped
// to bounded public projections without requiring per-item inspect calls.
func TestContainerProjectionFromDockerSummary(t *testing.T) {
	backend := &dockerRuntimeBackend{config: DefaultRuntimeConfig()}
	item := backend.containerInfoFromSummary(dockercontainer.Summary{
		ID:      "docker-owned",
		Names:   []string{"/agentbox-owned"},
		Image:   "ghcr.io/example/agentbox:latest",
		Created: 1704067200,
		State:   dockercontainer.StateRunning,
		Status:  "Up 5 minutes",
		Labels: map[string]string{
			containerLabelManaged:     containerManagedValue,
			containerLabelUser:        "usr-owner",
			containerLabelContainerID: "ctr-owned",
			containerLabelName:        "owned",
		},
		Mounts: []dockercontainer.MountPoint{{
			Type:        "volume",
			Source:      "agentbox-workspace",
			Destination: workspacesvc.DefaultWorkspaceRootPath,
			Name:        "agentbox-workspace",
		}},
	})
	if item.ID != "ctr-owned" || item.Name != "owned" || item.DockerID != "docker-owned" {
		t.Fatalf("unexpected container projection: %#v", item)
	}
	if item.CreatedAt != 1704067200000 || item.Workspace != workspacesvc.DefaultWorkspaceRootPath {
		t.Fatalf("unexpected timestamp/workspace projection: %#v", item)
	}
	if len(item.Mounts) != 1 || item.Mounts[0].Name != "agentbox-workspace" {
		t.Fatalf("unexpected mounts: %#v", item.Mounts)
	}
}

// TestParseAgentGitStatus verifies repository status headers preserve NUL-delimited porcelain output.
func TestParseAgentGitStatus(t *testing.T) {
	status, err := parseAgentGitStatus("repo\n/home/agent/workspace/project\n M tracked.txt\x00?? new.txt\x00", "/home/agent/workspace/project")
	if err != nil {
		t.Fatal(err)
	}
	if status.NotRepository || status.RepositoryRoot != "/home/agent/workspace/project" || status.Porcelain != " M tracked.txt\x00?? new.txt\x00" {
		t.Fatalf("unexpected git status: %#v", status)
	}

	notRepo, err := parseAgentGitStatus("not_repo\n/home/agent/workspace\n", "/home/agent/workspace")
	if err != nil {
		t.Fatal(err)
	}
	if !notRepo.NotRepository || notRepo.Path != "/home/agent/workspace" {
		t.Fatalf("unexpected not-repo git status: %#v", notRepo)
	}
}

type fakeDockerHealthBackend struct {
	response *DockerHealthResponse
	err      error
	calls    int
}

func (b *fakeDockerHealthBackend) Ping(context.Context) (*DockerHealthResponse, error) {
	b.calls++
	if b.err != nil {
		return nil, b.err
	}
	if b.response != nil {
		return b.response, nil
	}
	return &DockerHealthResponse{OK: true}, nil
}

type fakeContainerLifecycleBackend struct {
	items           []ContainerInfo
	item            *ContainerInfo
	logs            *LogsResponse
	err             error
	lastUserID      string
	lastContainerID string
	listCalls       int
}

func (b *fakeContainerLifecycleBackend) List(_ context.Context, userID string) ([]ContainerInfo, error) {
	b.listCalls++
	b.lastUserID = userID
	if b.err != nil {
		return nil, b.err
	}
	return b.items, nil
}

func (b *fakeContainerLifecycleBackend) Inspect(_ context.Context, userID string, containerID string) (*ContainerInfo, error) {
	b.lastUserID = userID
	b.lastContainerID = containerID
	return b.single()
}

func (b *fakeContainerLifecycleBackend) Start(_ context.Context, userID string, containerID string) (*ContainerInfo, error) {
	b.lastUserID = userID
	b.lastContainerID = containerID
	return b.single()
}

func (b *fakeContainerLifecycleBackend) Stop(_ context.Context, userID string, containerID string) (*ContainerInfo, error) {
	b.lastUserID = userID
	b.lastContainerID = containerID
	return b.single()
}

func (b *fakeContainerLifecycleBackend) Delete(_ context.Context, userID string, containerID string) (bool, error) {
	b.lastUserID = userID
	b.lastContainerID = containerID
	if b.err != nil {
		return false, b.err
	}
	return true, nil
}

func (b *fakeContainerLifecycleBackend) Logs(_ context.Context, userID string, containerID string) (*LogsResponse, error) {
	b.lastUserID = userID
	b.lastContainerID = containerID
	if b.err != nil {
		return nil, b.err
	}
	if b.logs != nil {
		return b.logs, nil
	}
	return &LogsResponse{}, nil
}

func (b *fakeContainerLifecycleBackend) single() (*ContainerInfo, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.item != nil {
		return b.item, nil
	}
	return &ContainerInfo{ID: "ctr-owned"}, nil
}
