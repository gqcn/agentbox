// This file defines the public prompt-template API contract surface for the
// AgentBox plugin. Prompt templates are plugin-owned AI data and are published
// through the source-plugin API namespace.

package prompt

import (
	"context"

	v1 "john-ai-agentbox/backend/api/prompt/v1"
)

// IPromptV1 defines AgentBox prompt-template HTTP handlers.
type IPromptV1 interface {
	// List returns registered prompt templates.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Detail returns one registered prompt template by code.
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	// Update stores one prompt-template content override.
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	// Restore clears one prompt-template content override.
	Restore(ctx context.Context, req *v1.RestoreReq) (res *v1.RestoreRes, err error)
	// Preview renders a prompt-template draft without persisting it.
	Preview(ctx context.Context, req *v1.PreviewReq) (res *v1.PreviewRes, err error)
}
