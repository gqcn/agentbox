// This file converts AgentBox authentication service projections into public
// API DTOs without leaking service internals into controller methods.

package auth

import (
	v1 "john-ai-agentbox/backend/api/auth/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
)

// toAuthSessionResponse maps a service session projection into a public DTO.
func toAuthSessionResponse(out *authsvc.SessionOutput) *v1.AuthSessionResponse {
	if out == nil || out.User == nil {
		return &v1.AuthSessionResponse{}
	}
	return &v1.AuthSessionResponse{
		User: &v1.AuthUser{
			ID:          out.User.ID,
			Username:    out.User.Username,
			DisplayName: out.User.DisplayName,
			LastLoginAt: out.User.LastLoginAt,
		},
	}
}
