// This file defines structured AgentBox terminal error codes. Controller-facing
// failures use LinaPro bizerr codes so clients receive stable responses instead
// of raw database or runtime strings.

package terminal

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTerminalInvalidInput reports malformed terminal session input.
	CodeTerminalInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_TERMINAL_INVALID_INPUT",
		"AgentBox terminal input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeTerminalNotFound reports a missing or invisible terminal session.
	CodeTerminalNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_TERMINAL_NOT_FOUND",
		"AgentBox terminal session was not found.",
		gcode.CodeNotFound,
	)
	// CodeTerminalStateConflict reports an invalid terminal lifecycle transition.
	CodeTerminalStateConflict = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_TERMINAL_STATE_CONFLICT",
		"AgentBox terminal session is in a conflicting state.",
		gcode.CodeInvalidOperation,
	)
	// CodeTerminalStoreUnavailable reports plugin-owned terminal storage failures.
	CodeTerminalStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_TERMINAL_STORE_UNAVAILABLE",
		"AgentBox terminal storage is temporarily unavailable.",
		gcode.CodeInternalError,
	)
)
