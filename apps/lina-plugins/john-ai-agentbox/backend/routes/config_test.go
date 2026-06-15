// This file verifies AgentBox plugin-owned business configuration loading
// without requiring a full Lina host registrar.

package routes

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestLoadAgentBoxConfigUsesDefaultsWithoutReader verifies focused tests can
// construct plugin routes without a host-published config service.
func TestLoadAgentBoxConfigUsesDefaultsWithoutReader(t *testing.T) {
	config, err := loadAgentBoxConfigFromReader(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if config.RuntimeMode != runtimeModeSingleNode {
		t.Fatalf("unexpected runtime mode %q", config.RuntimeMode)
	}
	if config.Auth.SessionTTL != 24*time.Hour {
		t.Fatalf("unexpected session ttl %s", config.Auth.SessionTTL)
	}
	if config.Docker.ContainerLogTail != 400 || config.Docker.StopTimeout != 10*time.Second {
		t.Fatalf("unexpected docker defaults: %#v", config.Docker)
	}
	if config.Docker.Workspace.WorkspaceRootPath != config.Workspace.WorkspaceRootPath {
		t.Fatalf("docker workspace config was not propagated: %#v", config.Docker.Workspace)
	}
}

// TestLoadAgentBoxConfigReadsBusinessSettings verifies manifest/config values
// are projected into the plugin service configuration graph.
func TestLoadAgentBoxConfigReadsBusinessSettings(t *testing.T) {
	reader := fakeAgentBoxConfigReader{values: map[string]string{
		"runtime.mode":                              "single-node",
		"auth.sessionTtl":                           "2h",
		"providers.requestTimeout":                  "7s",
		"providers.remoteModelSyncLimit":            "17",
		"ai.requestTimeout":                         "9s",
		"runtime.docker.host":                       "tcp://127.0.0.1:2375",
		"runtime.docker.containerLogTail":           "33",
		"runtime.docker.stopTimeout":                "4s",
		"runtime.workspace.rootPath":                "/workspace",
		"runtime.workspace.sharedPath":              "/shared",
		"runtime.workspace.previewLimitBytes":       "1234",
		"runtime.workspace.uploadFileLimitBytes":    "5678",
		"runtime.workspace.uploadCountLimit":        "3",
		"runtime.workspace.skillListLimit":          "4",
		"runtime.workspace.skillManifestLimitBytes": "99",
		"runtime.services.discoveryLimit":           "11",
	}}

	config, err := loadAgentBoxConfigFromReader(context.Background(), reader)
	if err != nil {
		t.Fatal(err)
	}
	if config.Auth.SessionTTL != 2*time.Hour || config.Catalog.RemoteRequestTimeout != 7*time.Second || config.AI.RequestTimeout != 9*time.Second {
		t.Fatalf("unexpected timeout config: %#v", config)
	}
	if config.Catalog.RemoteModelSyncLimit != 17 {
		t.Fatalf("unexpected provider sync limit %d", config.Catalog.RemoteModelSyncLimit)
	}
	if config.Docker.Host != "tcp://127.0.0.1:2375" || config.Docker.ContainerLogTail != 33 || config.Docker.StopTimeout != 4*time.Second {
		t.Fatalf("unexpected docker config: %#v", config.Docker)
	}
	if config.Workspace.WorkspaceRootPath != "/workspace" || config.Workspace.SharedRootPath != "/shared" {
		t.Fatalf("unexpected workspace roots: %#v", config.Workspace)
	}
	if config.Workspace.PreviewLimitBytes != 1234 || config.Workspace.UploadFileLimitBytes != 5678 || config.Workspace.UploadCountLimit != 3 {
		t.Fatalf("unexpected workspace limits: %#v", config.Workspace)
	}
	if config.Workspace.SkillListLimit != 4 || config.Workspace.SkillManifestLimitBytes != 99 {
		t.Fatalf("unexpected skill limits: %#v", config.Workspace)
	}
	if config.Service.RuntimeServiceListLimit != 11 || config.Docker.Service.RuntimeServiceListLimit != 11 {
		t.Fatalf("unexpected service discovery config: %#v", config.Service)
	}
	if config.Docker.Workspace.WorkspaceRootPath != "/workspace" || config.Docker.Workspace.SharedRootPath != "/shared" {
		t.Fatalf("docker runtime did not receive workspace config: %#v", config.Docker.Workspace)
	}
}

// TestLoadAgentBoxConfigRejectsUnsupportedRuntimeMode verifies future runtime
// modes cannot silently run under single-node assumptions.
func TestLoadAgentBoxConfigRejectsUnsupportedRuntimeMode(t *testing.T) {
	_, err := loadAgentBoxConfigFromReader(context.Background(), fakeAgentBoxConfigReader{
		values: map[string]string{"runtime.mode": "cluster"},
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported runtime mode error, got %v", err)
	}
}

type fakeAgentBoxConfigReader struct {
	values map[string]string
}

func (r fakeAgentBoxConfigReader) String(_ context.Context, key string, defaultValue string) (string, error) {
	value := strings.TrimSpace(r.values[key])
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (r fakeAgentBoxConfigReader) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(r.values[key])
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func (r fakeAgentBoxConfigReader) Duration(_ context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(r.values[key])
	if value == "" {
		return defaultValue, nil
	}
	return time.ParseDuration(value)
}
