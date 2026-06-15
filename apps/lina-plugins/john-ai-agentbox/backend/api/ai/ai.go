// This file defines the public AI capability API contract surface for the
// AgentBox plugin. DTOs live under versioned subpackages while this package
// exposes the GoFrame controller interface used by route registration.

package ai

import (
	"context"

	v1 "john-ai-agentbox/backend/api/ai/v1"
)

// IAiV1 defines AgentBox AI capability HTTP handlers.
type IAiV1 interface {
	// CapabilityTiers returns all fixed AgentBox AI capability tiers.
	CapabilityTiers(ctx context.Context, req *v1.CapabilityTiersReq) (res *v1.CapabilityTiersRes, err error)
	// UpdateCapabilityTier updates one capability tier and optional model binding.
	UpdateCapabilityTier(ctx context.Context, req *v1.UpdateCapabilityTierReq) (res *v1.UpdateCapabilityTierRes, err error)
	// TestCapabilityTier runs a lightweight provider connectivity test.
	TestCapabilityTier(ctx context.Context, req *v1.TestCapabilityTierReq) (res *v1.TestCapabilityTierRes, err error)
	// Invocations returns sanitized AI invocation logs.
	Invocations(ctx context.Context, req *v1.InvocationsReq) (res *v1.InvocationsRes, err error)
}
