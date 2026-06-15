// This file verifies Agent runtime Docker helper behavior without requiring a
// live Docker daemon.

package container

import (
	"testing"

	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// TestAgentRuntimeLabelsAreSeparateFromIndependentContainers verifies Agent runtime containers are not exposed by generic container list semantics.
func TestAgentRuntimeLabelsAreSeparateFromIndependentContainers(t *testing.T) {
	input := catalogsvc.AgentRuntimeContainerInput{
		AgentID:   "agt-owned",
		UserID:    "usr-owned",
		Name:      "Agent",
		ImageID:   42,
		ImageRef:  "ghcr.io/example/codex:latest",
		AgentType: "codex",
	}
	labels := agentRuntimeLabels(input)
	if !isAgentManagedContainer(labels, "usr-owned", "agt-owned") {
		t.Fatalf("expected agent labels to be visible to owner: %#v", labels)
	}
	if isAgentManagedContainer(labels, "usr-other", "agt-owned") {
		t.Fatal("agent labels should not be visible to another user")
	}
	if isIndependentManagedContainer(labels, "usr-owned") {
		t.Fatal("agent runtime labels should not satisfy generic independent container visibility")
	}
}

// TestNormalizeAgentRuntimeInputRejectsUntrustedCreateInputs verifies create uses trusted Agent fields.
func TestNormalizeAgentRuntimeInputRejectsUntrustedCreateInputs(t *testing.T) {
	if _, err := normalizeAgentRuntimeInput(catalogsvc.AgentRuntimeContainerInput{
		AgentID:  "agt-owned",
		UserID:   "usr-owned",
		ImageRef: "ghcr.io/example/codex:latest",
	}); err == nil {
		t.Fatal("expected missing image ID to be rejected")
	}
	out, err := normalizeAgentRuntimeInput(catalogsvc.AgentRuntimeContainerInput{
		AgentID:  " agt-owned ",
		UserID:   " usr-owned ",
		ImageID:  42,
		ImageRef: " ghcr.io/example/codex:latest ",
	})
	if err != nil {
		t.Fatalf("normalizeAgentRuntimeInput: %v", err)
	}
	if out.AgentID != "agt-owned" || out.UserID != "usr-owned" || out.ImageRef != "ghcr.io/example/codex:latest" || out.Name != "agt-owned" {
		t.Fatalf("unexpected normalized input: %#v", out)
	}
}

// TestAgentRuntimeContainerNameIsDockerSafe verifies container names are deterministic and bounded.
func TestAgentRuntimeContainerNameIsDockerSafe(t *testing.T) {
	name := agentRuntimeContainerName(" AGT_One With Spaces ")
	if name != "agentbox-agt-one-with-spaces" {
		t.Fatalf("container name = %q", name)
	}
}
