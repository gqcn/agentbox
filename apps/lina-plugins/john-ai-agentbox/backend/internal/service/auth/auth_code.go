// This file defines structured AgentBox authentication error codes. These
// codes are returned through LinaPro bizerr so HTTP clients receive stable
// machine-readable failure metadata during the migration.

package auth

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAuthInvalidCredentials reports invalid AgentBox login credentials.
	CodeAuthInvalidCredentials = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AUTH_INVALID_CREDENTIALS",
		"Invalid AgentBox username or password",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthRequired reports a missing, expired, revoked, or invalid session.
	CodeAuthRequired = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AUTH_REQUIRED",
		"AgentBox authentication is required",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthUserDisabled reports an AgentBox login attempt by a disabled user.
	CodeAuthUserDisabled = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AUTH_USER_DISABLED",
		"AgentBox user account is disabled",
		gcode.CodeNotAuthorized,
	)
	// CodeAuthStoreUnavailable reports plugin-owned user/session storage errors.
	CodeAuthStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_AUTH_STORE_UNAVAILABLE",
		"AgentBox authentication storage is temporarily unavailable",
		gcode.CodeInternalError,
	)
)
