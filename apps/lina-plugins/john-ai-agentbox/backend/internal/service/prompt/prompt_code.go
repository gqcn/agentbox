// This file defines structured AgentBox prompt error codes. These errors are
// returned through LinaPro bizerr so clients receive stable metadata instead of
// raw template parsing or storage strings.

package prompt

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePromptInvalidInput reports invalid prompt codes or template content.
	CodePromptInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_PROMPT_INVALID_INPUT",
		"AgentBox prompt input is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodePromptNotFound reports an unknown prompt template or missing override.
	CodePromptNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_PROMPT_NOT_FOUND",
		"AgentBox prompt template was not found",
		gcode.CodeNotFound,
	)
	// CodePromptStoreUnavailable reports prompt storage failures.
	CodePromptStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_PROMPT_STORE_UNAVAILABLE",
		"AgentBox prompt storage is temporarily unavailable",
		gcode.CodeInternalError,
	)
)
