// This file centralizes Chat string normalization, JSON validation defaults,
// time projection, and opaque ID generation used by the DAO-backed Chat
// service. Runtime code must use the same helpers so persisted values stay
// bounded to known enum strings.

package chat

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	defaultSessionTitle = "新对话"
	emptyChatPreview    = "尚无消息"
	titleMaxRunes       = 32
	previewMaxRunes     = 80
)

func newChatSessionID() string {
	return "chat-" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UnixMilli()
}

func unixMilliPtrFromTimePtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	out := value.UnixMilli()
	return &out
}

func normalizeSessionStatus(value string) string {
	switch strings.TrimSpace(value) {
	case SessionStatusIdle, SessionStatusRunning, SessionStatusWaiting, SessionStatusExited, SessionStatusRecovering, SessionStatusError:
		return strings.TrimSpace(value)
	default:
		return SessionStatusIdle
	}
}

func normalizeRuntimeState(value string) string {
	switch strings.TrimSpace(value) {
	case RuntimeStateIdle, RuntimeStateRunning, RuntimeStateWaiting, RuntimeStateExited, RuntimeStateRecovering, RuntimeStateError:
		return strings.TrimSpace(value)
	default:
		return RuntimeStateIdle
	}
}

func normalizeMessageRole(value string) string {
	switch strings.TrimSpace(value) {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem, MessageRoleError, MessageRoleTerminal:
		return strings.TrimSpace(value)
	default:
		return MessageRoleSystem
	}
}

func normalizeMessageStatus(value string) string {
	switch strings.TrimSpace(value) {
	case MessageStatusStreaming, MessageStatusComplete, MessageStatusError:
		return strings.TrimSpace(value)
	default:
		return MessageStatusComplete
	}
}

func normalizeInteractionType(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionTypePermission, InteractionTypeQuestion, InteractionTypeChoice, InteractionTypeText, InteractionTypeAuth, InteractionTypePlan, InteractionTypeCustom:
		return strings.TrimSpace(value)
	default:
		return InteractionTypeCustom
	}
}

func normalizeInteractionStatus(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionStatusPending, InteractionStatusResolved, InteractionStatusRejected, InteractionStatusCancelled, InteractionStatusExpired, InteractionStatusError:
		return strings.TrimSpace(value)
	default:
		return InteractionStatusPending
	}
}

func normalizeInteractionControlStatus(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionStatusCancelled, InteractionStatusExpired, InteractionStatusError:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeInteractionResponseScope(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionResponseScopeOnce, InteractionResponseScopeSession, InteractionResponseScopeAgent, InteractionResponseScopeProvider:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeInteractionResponseMode(value string) string {
	switch strings.TrimSpace(value) {
	case InteractionResponseModeAllow, InteractionResponseModeAnswer, InteractionResponseModeReject, InteractionResponseModeCancel, InteractionResponseModeAllowOnce, InteractionResponseModeAllowSession:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(value)
	}
}

func terminalStatusFromResponseMode(responseMode string) string {
	switch normalizeInteractionResponseMode(responseMode) {
	case InteractionResponseModeReject, InteractionResponseModeCancel:
		return InteractionStatusRejected
	default:
		return InteractionStatusResolved
	}
}

func normalizeTitle(value string) string {
	return truncateRunes(collapseWhitespace(value), titleMaxRunes)
}

func normalizePreview(value string) string {
	preview := truncateRunes(collapseWhitespace(value), previewMaxRunes)
	if preview == "" {
		return emptyChatPreview
	}
	return preview
}

func defaultJSON(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "{}"
	}
	return value, json.Valid([]byte(value))
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateRunes(value string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes])
}
