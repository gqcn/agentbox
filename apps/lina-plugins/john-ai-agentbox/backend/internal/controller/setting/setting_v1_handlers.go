// This file implements version-one setting handlers. Authentication middleware
// has already validated agent_box_session before these handlers run.

package setting

import (
	"context"

	v1 "john-ai-agentbox/backend/api/setting/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
)

// Detail returns one persisted AgentBox setting for the current user by key.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.settingSvc.GetUserSetting(ctx, userID, req.Key)
	if err != nil {
		return nil, err
	}
	out := toSettingResponse(*item)
	return (*v1.DetailRes)(&out), nil
}

// Update creates or updates one persisted AgentBox setting for the current user.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.settingSvc.UpsertUserSetting(ctx, userID, req.Key, req.Value)
	if err != nil {
		return nil, err
	}
	out := toSettingResponse(*item)
	return (*v1.UpdateRes)(&out), nil
}
