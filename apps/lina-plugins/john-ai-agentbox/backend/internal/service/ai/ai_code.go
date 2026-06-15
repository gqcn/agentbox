// This file defines structured AgentBox AI error codes. These errors are
// returned through LinaPro bizerr so clients receive stable metadata instead of
// raw provider or storage strings.

package ai

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAIInvalidInput reports invalid tier, provider, model, or filter input.
	CodeAIInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AI_INVALID_INPUT",
		"AgentBox AI input is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeAINotFound reports a missing AI capability, provider, or provider model.
	CodeAINotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AI_NOT_FOUND",
		"AgentBox AI resource was not found",
		gcode.CodeNotFound,
	)
	// CodeAIStoreUnavailable reports plugin-owned AI storage failures.
	CodeAIStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AI_STORE_UNAVAILABLE",
		"AgentBox AI storage is temporarily unavailable",
		gcode.CodeInternalError,
	)
	// CodeAIProviderFailed reports a provider connectivity or response failure.
	CodeAIProviderFailed = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AI_PROVIDER_FAILED",
		"AgentBox AI provider request failed",
		gcode.CodeInvalidOperation,
	)
)
