// This file implements version-one container request handlers. Every handler
// reads the authenticated AgentBox user ID from context before delegating to
// runtime service methods.

package container

import (
	"context"

	v1 "john-ai-agentbox/backend/api/container/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
)

// DockerHealth reports AgentBox runtime backend health for the current user.
func (c *ControllerV1) DockerHealth(ctx context.Context, _ *v1.DockerHealthReq) (res *v1.DockerHealthRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.DockerHealth(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toDockerHealthResponse(item), nil
}

// List lists AgentBox runtime containers for the current user.
func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (res *v1.ListRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.containerSvc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := toContainerListResponse(items)
	return (*v1.ListRes)(&out), nil
}

// Create creates an AgentBox runtime container for the current user.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.Create(ctx, userID, req.Name)
	if err != nil {
		return nil, err
	}
	out := toContainerResponse(*item)
	return (*v1.CreateRes)(&out), nil
}

// Detail gets one AgentBox runtime container for the current user.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.Detail(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toContainerResponse(*item)
	return (*v1.DetailRes)(&out), nil
}

// Start starts one AgentBox runtime container for the current user.
func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.Start(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toContainerResponse(*item)
	return (*v1.StartRes)(&out), nil
}

// Stop stops one AgentBox runtime container for the current user.
func (c *ControllerV1) Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.Stop(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toContainerResponse(*item)
	return (*v1.StopRes)(&out), nil
}

// Delete deletes one AgentBox runtime container for the current user.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	deleted, err := c.containerSvc.Delete(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: deleted}, nil
}

// Logs reads one AgentBox runtime container log stream for the current user.
func (c *ControllerV1) Logs(ctx context.Context, req *v1.LogsReq) (res *v1.LogsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.containerSvc.Logs(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	return toLogsResponse(item), nil
}
