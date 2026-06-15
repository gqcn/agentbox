// This file defines structured AgentBox container runtime error codes. Runtime
// migration gaps use stable bizerr codes so callers never receive raw Docker or
// host environment strings.

package container

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeContainerInvalidInput reports invalid container runtime input.
	CodeContainerInvalidInput = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CONTAINER_INVALID_INPUT",
		"AgentBox container input is invalid.",
		gcode.CodeInvalidParameter,
	)
	// CodeContainerNotFound reports a missing or invisible plugin-managed
	// container without leaking Docker IDs, names, status, or labels.
	CodeContainerNotFound = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CONTAINER_NOT_FOUND",
		"AgentBox container was not found.",
		gcode.CodeNotFound,
	)
	// CodeContainerRuntimeUnavailable reports Docker health failures or runtime-backed
	// container lifecycle actions unavailable in the current migration slice.
	CodeContainerRuntimeUnavailable = bizerr.MustDefine(
		"JOHN_AI_AGENTBOX_CONTAINER_RUNTIME_UNAVAILABLE",
		"AgentBox container runtime is temporarily unavailable.",
		gcode.CodeNotSupported,
	)
)

func newInvalidInputError() error {
	return bizerr.NewCode(CodeContainerInvalidInput)
}

func newContainerNotFoundError() error {
	return bizerr.NewCode(CodeContainerNotFound)
}

func newRuntimeUnavailableError() error {
	return bizerr.NewCode(CodeContainerRuntimeUnavailable)
}

func wrapLifecycleError(cause error) error {
	if cause == nil {
		return nil
	}
	if bizerr.Is(cause, CodeContainerNotFound) ||
		bizerr.Is(cause, CodeContainerInvalidInput) ||
		bizerr.Is(cause, CodeContainerRuntimeUnavailable) {
		return cause
	}
	return wrapRuntimeUnavailable(cause)
}
