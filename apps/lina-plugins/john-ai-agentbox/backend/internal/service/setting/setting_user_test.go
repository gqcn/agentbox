// This file verifies AgentBox setting input and timestamp projection helpers.
// Database integration coverage is provided by route and package compile tests;
// this unit test remains self-contained and avoids shared database state.

package setting

import (
	"testing"
	"time"

	"john-ai-agentbox/backend/internal/model/entity"
)

// TestNormalizeSettingScope verifies user-scoped setting key normalization.
func TestNormalizeSettingScope(t *testing.T) {
	userID, key := normalizeSettingScope(" usr-1 ", " Workbench ")
	if userID != "usr-1" || key != "workbench" {
		t.Fatalf("normalized scope user=%q key=%q", userID, key)
	}
}

// TestSettingInfoFromEntityTimeProjection verifies Unix millisecond projection
// for API response boundaries.
func TestSettingInfoFromEntityTimeProjection(t *testing.T) {
	created := time.UnixMilli(1704067200000)
	updated := time.UnixMilli(1704067201000)
	info := settingInfoFromEntity(testSettingEntity{
		Key:       "workbench",
		Value:     "{}",
		CreatedAt: &created,
		UpdatedAt: &updated,
	}.toEntity())
	if info == nil {
		t.Fatal("expected setting info")
	}
	if info.CreatedAt != 1704067200000 || info.UpdatedAt != 1704067201000 {
		t.Fatalf("timestamps = %d %d", info.CreatedAt, info.UpdatedAt)
	}
}

type testSettingEntity struct {
	Key       string
	Value     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (e testSettingEntity) toEntity() *entity.UserSettings {
	return &entity.UserSettings{
		Key:       e.Key,
		Value:     e.Value,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
