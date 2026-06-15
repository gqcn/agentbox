// This file implements version-one coding-image request handlers.
// Authentication is applied by the route group middleware; these handlers keep
// to DTO translation and catalog service calls.

package image

import (
	"context"

	v1 "john-ai-agentbox/backend/api/image/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// List returns configured AgentBox coding images.
func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (res *v1.ListRes, err error) {
	items, err := c.catalogSvc.ListImages(ctx)
	if err != nil {
		return nil, err
	}
	out := toImageListResponse(items)
	return (*v1.ListRes)(&out), nil
}

// Create creates one AgentBox coding image.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	item, err := c.catalogSvc.CreateImage(ctx, catalogsvc.CodingImageInput{
		Name:         req.Name,
		ImageRef:     req.ImageRef,
		AgentType:    req.AgentType,
		DefaultShell: req.DefaultShell,
		Notes:        req.Notes,
		Enabled:      req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	out := toImageResponse(*item)
	return (*v1.CreateRes)(&out), nil
}

// Update updates one AgentBox coding image.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	item, err := c.catalogSvc.UpdateImage(ctx, req.ID, catalogsvc.CodingImageInput{
		Name:         req.Name,
		ImageRef:     req.ImageRef,
		AgentType:    req.AgentType,
		DefaultShell: req.DefaultShell,
		Notes:        req.Notes,
		Enabled:      req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	out := toImageResponse(*item)
	return (*v1.UpdateRes)(&out), nil
}

// Delete deletes one unused AgentBox coding image.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err := c.catalogSvc.DeleteImage(ctx, req.ID); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Deleted: true}, nil
}
