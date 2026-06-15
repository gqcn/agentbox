// This file defines the public settings API contract surface for the AgentBox
// plugin. Settings are user-scoped plugin data and are published through the
// source-plugin API namespace.

package setting

import (
	"context"

	v1 "john-ai-agentbox/backend/api/setting/v1"
)

// ISettingV1 defines AgentBox user-scoped setting HTTP handlers.
type ISettingV1 interface {
	// Detail returns one current-user setting by key.
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	// Update creates or updates one current-user setting by key.
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
