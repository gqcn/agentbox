// This file implements version-one coding-agent request handlers. Every
// handler derives the AgentBox user ID from authentication context before
// delegating to user-scoped service methods.

package agent

import (
	"context"

	v1 "john-ai-agentbox/backend/api/agent/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// List returns non-deleted agents owned by the current AgentBox user.
func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (res *v1.ListRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.catalogSvc.ListUserAgents(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := toAgentListResponse(items)
	return (*v1.ListRes)(&out), nil
}

// Create creates one agent owned by the current AgentBox user.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.CreateUserAgent(ctx, userID, catalogsvc.AgentInput{
		Name:          req.Name,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		ModelProtocol: req.ModelProtocol,
		ImageID:       req.ImageID,
		AgentType:     req.AgentType,
		IconKey:       req.IconKey,
		Notes:         req.Notes,
	})
	if err != nil {
		return nil, err
	}
	out := toAgentResponse(*item)
	return (*v1.CreateRes)(&out), nil
}

// Detail returns one agent owned by the current AgentBox user.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.GetUserAgent(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toAgentResponse(*item)
	return (*v1.DetailRes)(&out), nil
}

// Update updates one agent owned by the current AgentBox user.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.UpdateUserAgent(ctx, userID, req.ID, catalogsvc.AgentInput{
		Name:          req.Name,
		ProviderID:    req.ProviderID,
		ModelName:     req.ModelName,
		ModelProtocol: req.ModelProtocol,
		AgentType:     req.AgentType,
		IconKey:       req.IconKey,
		Notes:         req.Notes,
	})
	if err != nil {
		return nil, err
	}
	out := toAgentResponse(*item)
	return (*v1.UpdateRes)(&out), nil
}

// ChangeImage switches one agent image after current-user ownership validation.
func (c *ControllerV1) ChangeImage(ctx context.Context, req *v1.ChangeImageReq) (res *v1.ChangeImageRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.catalogSvc.SetUserAgentImage(ctx, userID, req.ID, req.ImageID)
	if err != nil {
		return nil, err
	}
	response := toChangeImageResponse(out)
	return (*v1.ChangeImageRes)(response), nil
}

// Start validates current-user ownership before starting an Agent runtime when available.
func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.StartUserAgentRuntime(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toAgentResponse(*item)
	return (*v1.StartRes)(&out), nil
}

// Stop validates current-user ownership before stopping an Agent runtime when available.
func (c *ControllerV1) Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.StopUserAgentRuntime(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toAgentResponse(*item)
	return (*v1.StopRes)(&out), nil
}

// Logs validates current-user ownership before reading Agent runtime logs when available.
func (c *ControllerV1) Logs(ctx context.Context, req *v1.LogsReq) (res *v1.LogsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.catalogSvc.UserAgentRuntimeLogs(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.LogsRes{Logs: item.Logs}, nil
}

// Delete soft-deletes one agent owned by the current AgentBox user.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.catalogSvc.DeleteUserAgent(ctx, userID, req.ID); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: true}, nil
}
