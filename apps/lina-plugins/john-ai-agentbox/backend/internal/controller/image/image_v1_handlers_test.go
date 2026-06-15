// This file verifies image controller DTO projection and service error
// propagation without depending on plugin database state.

package image

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/image/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// TestImageListProjectsFields verifies coding-image fields are returned unchanged.
func TestImageListProjectsFields(t *testing.T) {
	controller := NewV1(&fakeImageCatalogService{
		images: []catalogsvc.CodingImageInfo{{
			ID:           3,
			Name:         "Codex Ubuntu",
			ImageRef:     "ghcr.io/example/codex:latest",
			AgentType:    catalogsvc.AgentTypeCodex,
			DefaultShell: "/bin/bash",
			Enabled:      true,
			IsDefault:    true,
			CreatedAt:    1718000000000,
			UpdatedAt:    1718000001000,
		}},
	})

	res, err := controller.List(context.Background(), &v1.ListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(*res) != 1 || (*res)[0].ImageRef != "ghcr.io/example/codex:latest" {
		t.Fatalf("unexpected image response: %#v", *res)
	}
	if !(*res)[0].IsDefault || (*res)[0].UpdatedAt != 1718000001000 {
		t.Fatalf("image projection lost fields: %#v", (*res)[0])
	}
}

// TestImageDeletePropagatesStructuredErrors verifies controller does not rewrite bizerr failures.
func TestImageDeletePropagatesStructuredErrors(t *testing.T) {
	expectedErr := bizerr.NewCode(catalogsvc.CodeCatalogResourceInUse)
	controller := NewV1(&fakeImageCatalogService{err: expectedErr})

	_, err := controller.Delete(context.Background(), &v1.DeleteReq{ID: 3})
	if !bizerr.Is(err, catalogsvc.CodeCatalogResourceInUse) {
		t.Fatalf("expected resource-in-use bizerr, got %v", err)
	}
}

type fakeImageCatalogService struct {
	catalogsvc.Service
	err    error
	images []catalogsvc.CodingImageInfo
}

func (s *fakeImageCatalogService) ListImages(context.Context) ([]catalogsvc.CodingImageInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.images, nil
}

func (s *fakeImageCatalogService) CreateImage(context.Context, catalogsvc.CodingImageInput) (*catalogsvc.CodingImageInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.images[0], nil
}

func (s *fakeImageCatalogService) UpdateImage(context.Context, int64, catalogsvc.CodingImageInput) (*catalogsvc.CodingImageInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.images[0], nil
}

func (s *fakeImageCatalogService) DeleteImage(context.Context, int64) error {
	return s.err
}
