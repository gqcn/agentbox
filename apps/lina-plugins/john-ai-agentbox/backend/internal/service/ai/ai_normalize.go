// This file keeps AgentBox AI enum normalization and timestamp helpers in one
// place so persistence and provider-test paths share the same validation rules.

package ai

import (
	"strings"
	"time"

	"john-ai-agentbox/backend/internal/service/catalog"
)

func normalizeTierCode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case TierBasic:
		return TierBasic
	case TierStandard:
		return TierStandard
	case TierAdvanced:
		return TierAdvanced
	default:
		return ""
	}
}

func normalizePurpose(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case PurposeCapabilityTest:
		return PurposeCapabilityTest
	case PurposeGitCommitMessage:
		return PurposeGitCommitMessage
	default:
		return ""
	}
}

func normalizeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case InvocationStatusSuccess:
		return InvocationStatusSuccess
	case InvocationStatusError:
		return InvocationStatusError
	default:
		return ""
	}
}

func normalizeProtocol(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case catalog.ProtocolOpenAI:
		return catalog.ProtocolOpenAI
	case catalog.ProtocolAnthropic:
		return catalog.ProtocolAnthropic
	default:
		return ""
	}
}

func sortedTierCodes() []string {
	return []string{TierBasic, TierStandard, TierAdvanced}
}

func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

func unixMilli(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

func sanitizeAIError(value string, secret string) string {
	value = strings.TrimSpace(value)
	if secret = strings.TrimSpace(secret); secret != "" {
		value = strings.ReplaceAll(value, secret, "[redacted]")
	}
	for _, marker := range []string{"Bearer ", "x-api-key", "api-key"} {
		if strings.Contains(value, marker) {
			value = strings.ReplaceAll(value, marker, "[redacted] ")
		}
	}
	if len(value) > 500 {
		value = value[:500]
	}
	if value == "" {
		return "ai provider request failed"
	}
	return value
}
