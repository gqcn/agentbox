// This file defines structured AgentBox catalog error codes. These errors are
// returned through LinaPro bizerr so clients receive stable machine-readable
// metadata instead of raw storage or validation strings.

package catalog

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCatalogInvalidInput reports invalid provider, model, or image input.
	CodeCatalogInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_INVALID_INPUT",
		"AgentBox catalog input is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCatalogNotFound reports a missing catalog resource.
	CodeCatalogNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_NOT_FOUND",
		"AgentBox catalog resource was not found",
		gcode.CodeNotFound,
	)
	// CodeCatalogResourceInUse reports an attempted delete of a referenced resource.
	CodeCatalogResourceInUse = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_RESOURCE_IN_USE",
		"AgentBox catalog resource is still in use",
		gcode.CodeInvalidOperation,
	)
	// CodeCatalogStoreUnavailable reports plugin-owned catalog storage failures.
	CodeCatalogStoreUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_STORE_UNAVAILABLE",
		"AgentBox catalog storage is temporarily unavailable",
		gcode.CodeInternalError,
	)
	// CodeCatalogRemoteSyncFailed reports remote provider model synchronization failures.
	CodeCatalogRemoteSyncFailed = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_REMOTE_SYNC_FAILED",
		"AgentBox provider model synchronization failed",
		gcode.CodeInvalidOperation,
	)
	// CodeCatalogRuntimeUnavailable reports Agent runtime lifecycle gaps that
	// have passed ownership checks but do not yet have a trusted runtime backend.
	CodeCatalogRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CATALOG_RUNTIME_UNAVAILABLE",
		"AgentBox runtime is temporarily unavailable",
		gcode.CodeNotSupported,
	)
)
