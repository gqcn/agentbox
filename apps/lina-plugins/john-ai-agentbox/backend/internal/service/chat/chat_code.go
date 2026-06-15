// This file defines structured AgentBox Chat error codes. Controller-facing
// failures use LinaPro bizerr codes so clients receive stable machine-readable
// errors instead of raw storage or runtime strings.

package chat

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeChatInvalidInput reports malformed Chat session or interaction input.
	CodeChatInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CHAT_INVALID_INPUT",
		"AgentBox chat input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeChatNotFound reports a missing or invisible Chat resource.
	CodeChatNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CHAT_NOT_FOUND",
		"AgentBox chat resource was not found.",
		gcode.CodeNotFound,
	)
	// CodeChatStateConflict reports an invalid Chat runtime or interaction state.
	CodeChatStateConflict = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CHAT_STATE_CONFLICT",
		"AgentBox chat resource is in a conflicting state.",
		gcode.CodeInvalidOperation,
	)
	// CodeChatStoreUnavailable reports plugin-owned Chat storage failures.
	CodeChatStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CHAT_STORE_UNAVAILABLE",
		"AgentBox chat storage is temporarily unavailable.",
		gcode.CodeInternalError,
	)
	// CodeChatRuntimeUnavailable reports runtime-dependent Chat actions that are
	// not yet available in the current migration slice.
	CodeChatRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CHAT_RUNTIME_UNAVAILABLE",
		"AgentBox chat runtime is temporarily unavailable.",
		gcode.CodeNotSupported,
	)
)
