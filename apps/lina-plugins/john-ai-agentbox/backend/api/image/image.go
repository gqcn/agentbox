// This file defines the public coding-image API contract surface for the
// AgentBox plugin. DTOs live under versioned subpackages while this package
// exposes the GoFrame controller interface used by route registration.

package image

import (
	"context"

	v1 "john-ai-agentbox/backend/api/image/v1"
)

// IImageV1 defines AgentBox coding-image HTTP handlers.
type IImageV1 interface {
	// List returns configured coding images.
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	// Create creates one coding image profile.
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	// Update updates one coding image profile.
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	// Delete deletes one coding image profile when it is unused.
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
}
