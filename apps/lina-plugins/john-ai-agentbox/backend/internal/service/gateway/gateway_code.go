// This file defines structured AgentBox raw gateway error codes. Runtime
// migration gaps use stable bizerr codes so callers never receive raw
// WebSocket, shell, proxy, or container backend strings.

package gateway

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeGatewayInvalidInput reports invalid raw gateway input.
	CodeGatewayInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_GATEWAY_INVALID_INPUT",
		"AgentBox gateway input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeGatewayRuntimeUnavailable reports runtime-backed raw gateway actions
	// that are not yet available in the current migration slice.
	CodeGatewayRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_GATEWAY_RUNTIME_UNAVAILABLE",
		"AgentBox gateway runtime is temporarily unavailable.",
		gcode.CodeNotSupported,
	)
)
