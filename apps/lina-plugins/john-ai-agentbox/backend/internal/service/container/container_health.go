// This file adapts the Docker SDK health probe to the AgentBox container
// service. The adapter deliberately exposes only ping metadata so runtime
// health can be verified without migrating container lifecycle operations.

package container

import (
	"context"

	"john-ai-agentbox/backend/internal/service/container/internal/dockerhealth"
)

// dockerHealthBackend adapts Docker runtime health to the service boundary.
type dockerHealthBackend struct {
	runtime dockerhealth.Runtime
}

var _ DockerHealthBackend = (*dockerHealthBackend)(nil)

// Ping verifies the Docker daemon and returns public health metadata.
func (b *dockerHealthBackend) Ping(ctx context.Context) (*DockerHealthResponse, error) {
	if b == nil || b.runtime == nil {
		return nil, newRuntimeUnavailableError()
	}
	info, err := b.runtime.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &DockerHealthResponse{
		OK:         true,
		APIVersion: info.APIVersion,
		OSType:     info.OSType,
	}, nil
}
