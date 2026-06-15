// This file implements version-one provider request handlers. Authentication
// is applied by the route group middleware; these handlers keep to DTO
// translation and catalog service calls.

package provider

import (
	"context"

	v1 "john-ai-agentbox/backend/api/provider/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// List returns configured AgentBox AI providers.
func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (res *v1.ListRes, err error) {
	items, err := c.catalogSvc.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	out := toProviderListResponse(items)
	return (*v1.ListRes)(&out), nil
}

// Create creates one AgentBox AI provider.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	item, err := c.catalogSvc.CreateProvider(ctx, catalogsvc.ProviderInput{
		Name:             req.Name,
		HomepageURL:      req.HomepageURL,
		Notes:            req.Notes,
		APIKey:           req.APIKey,
		OpenAIBaseURL:    req.OpenAIBaseURL,
		AnthropicBaseURL: req.AnthropicBaseURL,
	})
	if err != nil {
		return nil, err
	}
	out := toProviderResponse(*item)
	return (*v1.CreateRes)(&out), nil
}

// Detail returns one AgentBox AI provider.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, err := c.catalogSvc.GetProvider(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	out := toProviderResponse(*item)
	return (*v1.DetailRes)(&out), nil
}

// Update updates one AgentBox AI provider.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	item, err := c.catalogSvc.UpdateProvider(ctx, req.ID, catalogsvc.ProviderInput{
		Name:             req.Name,
		HomepageURL:      req.HomepageURL,
		Notes:            req.Notes,
		APIKey:           req.APIKey,
		OpenAIBaseURL:    req.OpenAIBaseURL,
		AnthropicBaseURL: req.AnthropicBaseURL,
	})
	if err != nil {
		return nil, err
	}
	out := toProviderResponse(*item)
	return (*v1.UpdateRes)(&out), nil
}

// Delete deletes one unused AgentBox AI provider.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err := c.catalogSvc.DeleteProvider(ctx, req.ID); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: true}, nil
}

// CreateModel creates one provider model.
func (c *ControllerV1) CreateModel(ctx context.Context, req *v1.CreateModelReq) (res *v1.CreateModelRes, err error) {
	item, err := c.catalogSvc.CreateProviderModel(ctx, req.ID, catalogsvc.ProviderModelInput{
		Name:     req.Name,
		Protocol: req.Protocol,
	})
	if err != nil {
		return nil, err
	}
	out := toProviderModelResponse(*item)
	return (*v1.CreateModelRes)(&out), nil
}

// DeleteModel deletes one unused provider model.
func (c *ControllerV1) DeleteModel(ctx context.Context, req *v1.DeleteModelReq) (res *v1.DeleteModelRes, err error) {
	if err := c.catalogSvc.DeleteProviderModel(ctx, req.ID, req.ModelID); err != nil {
		return nil, err
	}
	return &v1.DeleteModelRes{Deleted: true}, nil
}

// SyncModels synchronizes provider models from the remote provider API.
func (c *ControllerV1) SyncModels(ctx context.Context, req *v1.SyncModelsReq) (res *v1.SyncModelsRes, err error) {
	out, err := c.catalogSvc.SyncProviderModels(ctx, req.ID, req.Protocol)
	if err != nil {
		return nil, err
	}
	response := toSyncModelsResponse(out)
	return (*v1.SyncModelsRes)(response), nil
}
