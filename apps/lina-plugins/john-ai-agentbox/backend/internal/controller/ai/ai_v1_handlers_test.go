// This file verifies AI capability controller DTO projection and structured
// service error propagation without depending on plugin database state.

package ai

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/ai/v1"
	aisvc "john-ai-agentbox/backend/internal/service/ai"
)

// TestCapabilityTiersProjectsBindingAndLastTest verifies nested AI projections.
func TestCapabilityTiersProjectsBindingAndLastTest(t *testing.T) {
	controller := NewV1(&fakeAIService{
		tiers: []aisvc.CapabilityTierInfo{{
			Code:        aisvc.TierBasic,
			DisplayName: "基础",
			Enabled:     true,
			Configured:  true,
			Available:   true,
			Binding: &aisvc.CapabilityBindingInfo{
				ID:              7,
				TierCode:        aisvc.TierBasic,
				ProviderID:      1,
				ProviderName:    "OpenAI",
				ProviderModelID: 9,
				ModelName:       "gpt-test",
				Protocol:        "openai",
				Enabled:         true,
				CreatedAt:       1718000000000,
				UpdatedAt:       1718000001000,
			},
			LastTest: &aisvc.InvocationLogInfo{
				ID:        11,
				Purpose:   aisvc.PurposeCapabilityTest,
				TierCode:  aisvc.TierBasic,
				Status:    aisvc.InvocationStatusSuccess,
				CreatedAt: 1718000002000,
			},
		}},
	})

	res, err := controller.CapabilityTiers(context.Background(), &v1.CapabilityTiersReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(*res) != 1 || (*res)[0].Binding == nil || (*res)[0].Binding.ModelName != "gpt-test" {
		t.Fatalf("unexpected tier response: %#v", *res)
	}
	if (*res)[0].LastTest == nil || (*res)[0].LastTest.Status != aisvc.InvocationStatusSuccess {
		t.Fatalf("unexpected last test: %#v", (*res)[0].LastTest)
	}
}

// TestAIControllerPropagatesStructuredErrors verifies handlers do not rewrite bizerr failures.
func TestAIControllerPropagatesStructuredErrors(t *testing.T) {
	controller := NewV1(&fakeAIService{err: bizerr.NewCode(aisvc.CodeAIInvalidInput)})

	_, err := controller.UpdateCapabilityTier(context.Background(), &v1.UpdateCapabilityTierReq{
		Code: aisvc.TierBasic,
	})
	if !bizerr.Is(err, aisvc.CodeAIInvalidInput) {
		t.Fatalf("error = %v, want CodeAIInvalidInput", err)
	}
}

type fakeAIService struct {
	aisvc.Service
	err   error
	tiers []aisvc.CapabilityTierInfo
}

func (s *fakeAIService) ListTiers(context.Context) ([]aisvc.CapabilityTierInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tiers, nil
}

func (s *fakeAIService) UpdateTier(context.Context, string, aisvc.UpdateTierInput) (*aisvc.CapabilityTierInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.tiers[0], nil
}

func (s *fakeAIService) TestTier(context.Context, string, aisvc.TestTierInput) (*aisvc.CapabilityTestResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &aisvc.CapabilityTestResult{Status: aisvc.InvocationStatusSuccess}, nil
}

func (s *fakeAIService) ListInvocations(context.Context, aisvc.InvocationLogFilter) ([]aisvc.InvocationLogInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []aisvc.InvocationLogInfo{}, nil
}
