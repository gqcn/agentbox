// This file defines structured AgentBox access errors used by resource
// ownership checks before chat, workspace, terminal, and proxy entry points
// expose runtime data.

package access

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAccessInvalidInput reports a malformed access check request.
	CodeAccessInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_ACCESS_INVALID_INPUT",
		"AgentBox access request is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeAccessResourceUnavailable reports resources that are missing or not
	// visible to the current AgentBox user without leaking existence details.
	CodeAccessResourceUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_ACCESS_RESOURCE_UNAVAILABLE",
		"AgentBox resource is not available.",
		gcode.CodeNotFound,
	)
	// CodeAccessStoreUnavailable reports storage failures while checking access.
	CodeAccessStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_ACCESS_STORE_UNAVAILABLE",
		"AgentBox access store is temporarily unavailable.",
		gcode.CodeInternalError,
	)
)
