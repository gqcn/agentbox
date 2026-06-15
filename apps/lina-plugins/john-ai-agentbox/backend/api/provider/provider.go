// This file defines the public provider API contract surface for the
// AgentBox plugin. DTOs live under versioned subpackages while this package
// exposes the GoFrame controller interface used by route registration.

package provider

import (
	"context"

	v1 "john-ai-agentbox/backend/api/provider/v1"
)

// IProviderV1 defines AgentBox AI provider and provider-model HTTP handlers.
type IProviderV1 interface {
	// List returns configured AI providers with model projections.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Create creates one AI provider configuration.
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	// Detail returns one AI provider configuration.
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	// Update updates one AI provider configuration.
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	// Delete deletes one AI provider configuration when it is unused.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	// CreateModel creates or updates one manually managed provider model.
	CreateModel(ctx context.Context, req *v1.CreateModelReq) (res *v1.CreateModelRes, err error)
	// DeleteModel deletes one provider model when it is unused.
	DeleteModel(ctx context.Context, req *v1.DeleteModelReq) (res *v1.DeleteModelRes, err error)
	// SyncModels synchronizes provider models from the remote provider API.
	SyncModels(ctx context.Context, req *v1.SyncModelsReq) (res *v1.SyncModelsRes, err error)
}
