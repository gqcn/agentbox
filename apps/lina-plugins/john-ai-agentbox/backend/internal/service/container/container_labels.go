// This file centralizes Docker label names used by the AgentBox plugin. Runtime
// lifecycle APIs only manage containers carrying these plugin-owned labels and
// the authenticated AgentBox user label.

package container

import "strings"

const (
	containerLabelManaged     = "john-ai-agentbox.managed"
	containerLabelUser        = "john-ai-agentbox.user"
	containerLabelContainerID = "john-ai-agentbox.container_id"
	containerLabelName        = "john-ai-agentbox.name"
	containerLabelAgentID     = "john-ai-agentbox.agent_id"
	containerLabelAgentType   = "john-ai-agentbox.agent_type"
	containerLabelImageID     = "john-ai-agentbox.image_id"
	containerManagedValue     = "true"
)

func managedContainerLabelFilter() string {
	return containerLabelManaged + "=" + containerManagedValue
}

func userContainerLabelFilter(userID string) string {
	return containerLabelUser + "=" + strings.TrimSpace(userID)
}

func labelValue(labels map[string]string, key string, fallback string) string {
	if labels == nil {
		return fallback
	}
	if value := strings.TrimSpace(labels[key]); value != "" {
		return value
	}
	return fallback
}

func isIndependentManagedContainer(labels map[string]string, userID string) bool {
	if labels == nil {
		return false
	}
	return strings.TrimSpace(labels[containerLabelManaged]) == containerManagedValue &&
		strings.TrimSpace(labels[containerLabelUser]) == strings.TrimSpace(userID) &&
		strings.TrimSpace(labels[containerLabelContainerID]) != "" &&
		strings.TrimSpace(labels[containerLabelAgentID]) == ""
}

func isAgentManagedContainer(labels map[string]string, userID string, agentID string) bool {
	if labels == nil {
		return false
	}
	return strings.TrimSpace(labels[containerLabelManaged]) == containerManagedValue &&
		strings.TrimSpace(labels[containerLabelUser]) == strings.TrimSpace(userID) &&
		strings.TrimSpace(labels[containerLabelAgentID]) == strings.TrimSpace(agentID)
}

func copyLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	out := make(map[string]string, len(labels))
	for key, value := range labels {
		out[key] = value
	}
	return out
}
