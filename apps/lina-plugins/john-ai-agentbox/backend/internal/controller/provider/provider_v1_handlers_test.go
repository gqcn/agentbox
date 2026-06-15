// This file verifies provider controller DTO projection and service error
// propagation without depending on plugin database state.

package provider

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/provider/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// TestProviderListProjectsModels verifies model projections stay attached to providers.
func TestProviderListProjectsModels(t *testing.T) {
	controller := NewV1(&fakeCatalogService{
		providers: []catalogsvc.ProviderInfo{{
			ID:               7,
			Name:             "OpenAI",
			APIKeyMasked:     "sk-****1234",
			APIKeyConfigured: true,
			CreatedAt:        1718000000000,
			UpdatedAt:        1718000001000,
			Models: []catalogsvc.ProviderModelInfo{{
				ID:         9,
				ProviderID: 7,
				Name:       "gpt-5",
				Protocol:   catalogsvc.ProtocolOpenAI,
				Source:     catalogsvc.ModelSourceManual,
			}},
		}},
	})

	res, err := controller.List(context.Background(), &v1.ListReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(*res) != 1 || (*res)[0].Models[0].Name != "gpt-5" {
		t.Fatalf("unexpected provider response: %#v", *res)
	}
	if !(*res)[0].APIKeyConfigured || (*res)[0].CreatedAt != 1718000000000 {
		t.Fatalf("provider projection lost fields: %#v", (*res)[0])
	}
}

// TestProviderCreatePropagatesStructuredErrors verifies controller does not rewrite bizerr failures.
func TestProviderCreatePropagatesStructuredErrors(t *testing.T) {
	expectedErr := bizerr.NewCode(catalogsvc.CodeCatalogInvalidInput)
	controller := NewV1(&fakeCatalogService{err: expectedErr})

	_, err := controller.Create(context.Background(), &v1.CreateReq{})
	if !bizerr.Is(err, catalogsvc.CodeCatalogInvalidInput) {
		t.Fatalf("expected invalid-input bizerr, got %v", err)
	}
}

type fakeCatalogService struct {
	catalogsvc.Service
	err       error
	providers []catalogsvc.ProviderInfo
	model     *catalogsvc.ProviderModelInfo
	syncOut   *catalogsvc.SyncProviderModelsOutput
}

func (s *fakeCatalogService) ListProviders(context.Context) ([]catalogsvc.ProviderInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.providers, nil
}

func (s *fakeCatalogService) CreateProvider(context.Context, catalogsvc.ProviderInput) (*catalogsvc.ProviderInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.providers[0], nil
}

func (s *fakeCatalogService) GetProvider(context.Context, int64) (*catalogsvc.ProviderInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.providers[0], nil
}

func (s *fakeCatalogService) UpdateProvider(context.Context, int64, catalogsvc.ProviderInput) (*catalogsvc.ProviderInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.providers[0], nil
}

func (s *fakeCatalogService) DeleteProvider(context.Context, int64) error {
	return s.err
}

func (s *fakeCatalogService) CreateProviderModel(context.Context, int64, catalogsvc.ProviderModelInput) (*catalogsvc.ProviderModelInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.model, nil
}

func (s *fakeCatalogService) DeleteProviderModel(context.Context, int64, int64) error {
	return s.err
}

func (s *fakeCatalogService) SyncProviderModels(context.Context, int64, string) (*catalogsvc.SyncProviderModelsOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.syncOut, nil
}
