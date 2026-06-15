// This file defines the public container API contract surface for the
// AgentBox plugin. Versioned DTOs stay in subpackages while this package keeps
// the controller interface used by source-plugin route registration.

package container

import (
	"context"

	v1 "john-ai-agentbox/backend/api/container/v1"
)

// IContainerV1 defines AgentBox container HTTP handlers.
type IContainerV1 interface {
	// DockerHealth checks whether the runtime backend is available.
	DockerHealth(ctx context.Context, req *v1.DockerHealthReq) (res *v1.DockerHealthRes, err error)
	// List lists plugin-owned runtime containers for the current AgentBox user.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Create creates a runtime container for the current AgentBox user.
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	// Detail gets one plugin-owned runtime container for the current AgentBox user.
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	// Start starts one plugin-owned runtime container.
	Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error)
	// Stop stops one plugin-owned runtime container.
	Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error)
	// Delete deletes one plugin-owned runtime container.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// Logs reads logs for one plugin-owned runtime container.
	Logs(ctx context.Context, req *v1.LogsReq) (res *v1.LogsRes, err error)
}
