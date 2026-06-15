// This file implements user-scoped AgentBox key/value setting persistence. All
// queries include user_id in the database condition so settings with the same
// key never cross AgentBox user boundaries.

package setting

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// GetUserSetting returns one setting owned by userID and key.
func (s *serviceImpl) GetUserSetting(ctx context.Context, userID string, key string) (*SettingInfo, error) {
	userID, key = normalizeSettingScope(userID, key)
	if userID == "" || key == "" {
		return nil, bizerr.NewCode(CodeSettingInvalidInput)
	}
	var row *entity.UserSettings
	err := dao.UserSettings.Ctx(ctx).
		Where(do.UserSettings{UserId: userID, Key: key}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettingStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeSettingNotFound)
	}
	return settingInfoFromEntity(row), nil
}

// UpsertUserSetting creates or updates one setting owned by userID and key.
func (s *serviceImpl) UpsertUserSetting(ctx context.Context, userID string, key string, value string) (*SettingInfo, error) {
	userID, key = normalizeSettingScope(userID, key)
	if userID == "" || key == "" {
		return nil, bizerr.NewCode(CodeSettingInvalidInput)
	}
	_, err := dao.UserSettings.Ctx(ctx).Data(do.UserSettings{
		UserId: userID,
		Key:    key,
		Value:  value,
	}).
		OnConflict(dao.UserSettings.Columns().UserId, dao.UserSettings.Columns().Key).
		OnDuplicate(dao.UserSettings.Columns().Value).
		Save()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettingStoreUnavailable)
	}
	return s.GetUserSetting(ctx, userID, key)
}

func normalizeSettingScope(userID string, key string) (string, string) {
	return strings.TrimSpace(userID), strings.ToLower(strings.TrimSpace(key))
}

func settingInfoFromEntity(row *entity.UserSettings) *SettingInfo {
	if row == nil {
		return nil
	}
	return &SettingInfo{
		Key:       row.Key,
		Value:     row.Value,
		CreatedAt: unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt: unixMilliFromTimePtr(row.UpdatedAt),
	}
}

func unixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}
