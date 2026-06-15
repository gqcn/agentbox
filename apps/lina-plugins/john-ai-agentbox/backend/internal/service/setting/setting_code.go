// This file defines structured AgentBox setting error codes. These errors are
// returned through LinaPro bizerr so clients receive stable metadata instead of
// raw storage strings.

package setting

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSettingInvalidInput reports invalid setting keys or user IDs.
	CodeSettingInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_SETTING_INVALID_INPUT",
		"AgentBox setting input is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingNotFound reports a missing current-user setting.
	CodeSettingNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_SETTING_NOT_FOUND",
		"AgentBox setting was not found",
		gcode.CodeNotFound,
	)
	// CodeSettingStoreUnavailable reports setting storage failures.
	CodeSettingStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_SETTING_STORE_UNAVAILABLE",
		"AgentBox setting storage is temporarily unavailable",
		gcode.CodeInternalError,
	)
)
