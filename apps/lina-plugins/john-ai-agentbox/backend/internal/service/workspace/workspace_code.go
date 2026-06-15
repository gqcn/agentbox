// This file defines structured AgentBox workspace error codes. Controller-facing
// failures use LinaPro bizerr codes so clients receive stable machine-readable
// errors instead of raw runtime strings.

package workspace

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeWorkspaceInvalidInput reports invalid AgentBox workspace input.
	CodeWorkspaceInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_WORKSPACE_INVALID_INPUT",
		"AgentBox workspace input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeWorkspaceRuntimeUnavailable reports runtime-backed workspace actions
	// that are not yet available in the current migration slice.
	CodeWorkspaceRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_WORKSPACE_RUNTIME_UNAVAILABLE",
		"AgentBox workspace runtime is temporarily unavailable.",
		gcode.CodeNotSupported,
	)
	// CodeWorkspaceStateConflict reports stale-write or existing-entry conflicts.
	CodeWorkspaceStateConflict = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_WORKSPACE_STATE_CONFLICT",
		"AgentBox workspace resource is in a conflicting state.",
		gcode.CodeInvalidOperation,
	)
)
