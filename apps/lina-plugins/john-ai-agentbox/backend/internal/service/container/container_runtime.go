// This file implements container runtime behavior for the current migration
// slice. Docker health and label-scoped lifecycle actions delegate to injected
// backends after validating authenticated AgentBox user context.

package container

import "context"

// DockerHealth reports runtime health for the authenticated AgentBox user.
func (s *serviceImpl) DockerHealth(ctx context.Context, userID string) (*DockerHealthResponse, error) {
	if _, _, err := normalizeRuntimeUserAndID(userID); err != nil {
		return nil, err
	}
	item, err := s.healthBackend.Ping(ctx)
	if err != nil {
		return nil, wrapRuntimeUnavailable(err)
	}
	if item == nil || !item.OK {
		return nil, newRuntimeUnavailableError()
	}
	return item, nil
}

// List lists runtime containers for the authenticated AgentBox user.
func (s *serviceImpl) List(ctx context.Context, userID string) ([]ContainerInfo, error) {
	normalizedUserID, _, err := normalizeRuntimeUserAndID(userID)
	if err != nil {
		return nil, err
	}
	items, err := s.lifecycleBackend.List(ctx, normalizedUserID)
	if err != nil {
		return nil, wrapLifecycleError(err)
	}
	return items, nil
}

// Create creates a runtime container for the authenticated AgentBox user.
func (s *serviceImpl) Create(_ context.Context, userID string, _ string) (*ContainerInfo, error) {
	if _, _, err := normalizeRuntimeUserAndID(userID); err != nil {
		return nil, err
	}
	return nil, newRuntimeUnavailableError()
}

// Detail gets one runtime container visible to the authenticated AgentBox user.
func (s *serviceImpl) Detail(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	normalizedUserID, values, err := normalizeRuntimeUserAndID(userID, containerID)
	if err != nil {
		return nil, err
	}
	item, err := s.lifecycleBackend.Inspect(ctx, normalizedUserID, values[0])
	if err != nil {
		return nil, wrapLifecycleError(err)
	}
	return item, nil
}

// Start starts one runtime container visible to the authenticated AgentBox user.
func (s *serviceImpl) Start(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	normalizedUserID, values, err := normalizeRuntimeUserAndID(userID, containerID)
	if err != nil {
		return nil, err
	}
	item, err := s.lifecycleBackend.Start(ctx, normalizedUserID, values[0])
	if err != nil {
		return nil, wrapLifecycleError(err)
	}
	return item, nil
}

// Stop stops one runtime container visible to the authenticated AgentBox user.
func (s *serviceImpl) Stop(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	normalizedUserID, values, err := normalizeRuntimeUserAndID(userID, containerID)
	if err != nil {
		return nil, err
	}
	item, err := s.lifecycleBackend.Stop(ctx, normalizedUserID, values[0])
	if err != nil {
		return nil, wrapLifecycleError(err)
	}
	return item, nil
}

// Delete deletes one runtime container visible to the authenticated AgentBox user.
func (s *serviceImpl) Delete(ctx context.Context, userID string, containerID string) (bool, error) {
	normalizedUserID, values, err := normalizeRuntimeUserAndID(userID, containerID)
	if err != nil {
		return false, err
	}
	deleted, err := s.lifecycleBackend.Delete(ctx, normalizedUserID, values[0])
	if err != nil {
		return false, wrapLifecycleError(err)
	}
	return deleted, nil
}

// Logs reads logs for one runtime container visible to the authenticated AgentBox user.
func (s *serviceImpl) Logs(ctx context.Context, userID string, containerID string) (*LogsResponse, error) {
	normalizedUserID, values, err := normalizeRuntimeUserAndID(userID, containerID)
	if err != nil {
		return nil, err
	}
	item, err := s.lifecycleBackend.Logs(ctx, normalizedUserID, values[0])
	if err != nil {
		return nil, wrapLifecycleError(err)
	}
	return item, nil
}
