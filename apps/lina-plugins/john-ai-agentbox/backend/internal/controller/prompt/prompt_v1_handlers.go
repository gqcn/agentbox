// This file implements version-one prompt-template handlers. Authentication is
// applied by the route group middleware; handlers only translate DTOs and call
// the prompt service.

package prompt

import (
	"context"

	v1 "john-ai-agentbox/backend/api/prompt/v1"
	promptsvc "john-ai-agentbox/backend/internal/service/prompt"
)

// List returns registered AgentBox prompt templates.
func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (res *v1.ListRes, err error) {
	items, err := c.promptSvc.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	out := toPromptListResponse(items)
	return (*v1.ListRes)(&out), nil
}

// Detail returns one AgentBox prompt template.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, err := c.promptSvc.GetTemplate(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	out := toPromptResponse(item)
	return (*v1.DetailRes)(&out), nil
}

// Update stores one AgentBox prompt template content override.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	item, err := c.promptSvc.UpdateTemplate(ctx, req.Code, promptsvc.UpdateInput{
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}
	out := toPromptResponse(item)
	return (*v1.UpdateRes)(&out), nil
}

// Restore clears one AgentBox prompt template content override.
func (c *ControllerV1) Restore(ctx context.Context, req *v1.RestoreReq) (res *v1.RestoreRes, err error) {
	item, err := c.promptSvc.RestoreTemplate(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	out := toPromptResponse(item)
	return (*v1.RestoreRes)(&out), nil
}

// Preview renders one AgentBox prompt template draft without persistence.
func (c *ControllerV1) Preview(ctx context.Context, req *v1.PreviewReq) (res *v1.PreviewRes, err error) {
	out, err := c.promptSvc.PreviewTemplate(ctx, req.Code, promptsvc.PreviewInput{
		Content:   req.Content,
		Variables: req.Variables,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PreviewRes{RenderedPrompt: out.RenderedPrompt}, nil
}
