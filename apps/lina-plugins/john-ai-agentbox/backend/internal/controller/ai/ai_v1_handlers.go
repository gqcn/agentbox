// This file implements version-one AI capability request handlers. The route
// group middleware applies AgentBox authentication before these methods run.

package ai

import (
	"context"

	v1 "john-ai-agentbox/backend/api/ai/v1"
	aisvc "john-ai-agentbox/backend/internal/service/ai"
)

// CapabilityTiers returns all fixed AgentBox AI capability tiers.
func (c *ControllerV1) CapabilityTiers(ctx context.Context, _ *v1.CapabilityTiersReq) (res *v1.CapabilityTiersRes, err error) {
	items, err := c.aiSvc.ListTiers(ctx)
	if err != nil {
		return nil, err
	}
	out := toTierListResponse(items)
	return (*v1.CapabilityTiersRes)(&out), nil
}

// UpdateCapabilityTier updates one AI capability tier and optional binding.
func (c *ControllerV1) UpdateCapabilityTier(ctx context.Context, req *v1.UpdateCapabilityTierReq) (res *v1.UpdateCapabilityTierRes, err error) {
	item, err := c.aiSvc.UpdateTier(ctx, req.Code, aisvc.UpdateTierInput{
		Enabled:         req.Enabled,
		ProviderID:      req.ProviderID,
		ProviderModelID: req.ProviderModelID,
		Protocol:        req.Protocol,
	})
	if err != nil {
		return nil, err
	}
	out := toTierResponse(*item)
	return (*v1.UpdateCapabilityTierRes)(&out), nil
}

// TestCapabilityTier runs a lightweight provider connectivity test.
func (c *ControllerV1) TestCapabilityTier(ctx context.Context, req *v1.TestCapabilityTierReq) (res *v1.TestCapabilityTierRes, err error) {
	item, err := c.aiSvc.TestTier(ctx, req.Code, aisvc.TestTierInput{
		ProviderID:      req.ProviderID,
		ProviderModelID: req.ProviderModelID,
		Protocol:        req.Protocol,
	})
	if err != nil {
		return nil, err
	}
	out := toTestResponse(*item)
	return (*v1.TestCapabilityTierRes)(&out), nil
}

// Invocations returns sanitized AI invocation logs.
func (c *ControllerV1) Invocations(ctx context.Context, req *v1.InvocationsReq) (res *v1.InvocationsRes, err error) {
	items, err := c.aiSvc.ListInvocations(ctx, aisvc.InvocationLogFilter{
		Purpose:  req.Purpose,
		TierCode: req.Tier,
		Status:   req.Status,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := toInvocationListResponse(items)
	return (*v1.InvocationsRes)(&out), nil
}
