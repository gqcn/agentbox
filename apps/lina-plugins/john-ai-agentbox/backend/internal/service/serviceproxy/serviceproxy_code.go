// This file defines structured AgentBox service-proxy error codes. Runtime
// migration gaps use stable bizerr codes so callers never receive raw proxy or
// container runtime strings.

package serviceproxy

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeServiceProxyInvalidInput reports invalid service-proxy input.
	CodeServiceProxyInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_SERVICE_PROXY_INVALID_INPUT",
		"AgentBox service proxy input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeServiceProxyRuntimeUnavailable reports runtime-backed service proxy
	// actions that are not yet available in the current migration slice.
	CodeServiceProxyRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_SERVICE_PROXY_RUNTIME_UNAVAILABLE",
		"AgentBox service proxy runtime is temporarily unavailable.",
		gcode.CodeNotSupported,
	)
)
