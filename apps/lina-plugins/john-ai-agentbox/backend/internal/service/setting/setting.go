// Package setting owns AgentBox user-scoped key/value preferences. Settings use
// the plugin-local AgentBox user boundary and never read LinaPro workbench user
// preferences or host AI configuration tables.
package setting

import "context"

// SettingInfo is the service-level user setting projection.
type SettingInfo struct {
	Key       string
	Value     string
	CreatedAt int64
	UpdatedAt int64
}

// Service defines AgentBox user-scoped setting persistence.
type Service interface {
	// GetUserSetting returns one setting owned by userID and key. Missing rows
	// return CodeSettingNotFound so callers do not accidentally share defaults
	// across AgentBox users.
	GetUserSetting(ctx context.Context, userID string, key string) (*SettingInfo, error)
	// UpsertUserSetting creates or updates one setting owned by userID and key.
	// Empty values are valid because callers may serialize empty preference
	// objects; blank user IDs or keys return CodeSettingInvalidInput.
	UpsertUserSetting(ctx context.Context, userID string, key string, value string) (*SettingInfo, error)
}

// serviceImpl is the default AgentBox setting service implementation.
type serviceImpl struct{}

var _ Service = (*serviceImpl)(nil)

// New creates the AgentBox setting service.
func New() (Service, error) {
	return &serviceImpl{}, nil
}
