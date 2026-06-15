// This file implements DAO-backed prompt-template override persistence. It
// reads and writes only the plugin-owned john_ai_agentbox_system_prompt_overrides
// table so prompt data remains isolated from host AI modules and other plugins.

package prompt

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

type daoStore struct{}

var _ OverrideStore = (*daoStore)(nil)

// NewDAOStore creates the plugin-owned prompt override store.
func NewDAOStore() OverrideStore {
	return &daoStore{}
}

// GetPromptTemplateOverride returns one persisted override by code.
func (s *daoStore) GetPromptTemplateOverride(ctx context.Context, code string) (OverrideInfo, error) {
	code = normalizePromptCode(code)
	if code == "" {
		return OverrideInfo{}, bizerr.NewCode(CodePromptInvalidInput)
	}
	var row *entity.SystemPromptOverrides
	err := dao.SystemPromptOverrides.Ctx(ctx).
		Where(do.SystemPromptOverrides{Code: code}).
		Scan(&row)
	if err != nil {
		return OverrideInfo{}, bizerr.WrapCode(err, CodePromptStoreUnavailable)
	}
	if row == nil {
		return OverrideInfo{}, bizerr.NewCode(CodePromptNotFound)
	}
	return overrideFromEntity(row), nil
}

// UpsertPromptTemplateOverride stores content for one registered template.
func (s *daoStore) UpsertPromptTemplateOverride(ctx context.Context, code string, input UpdateInput) (OverrideInfo, error) {
	code = normalizePromptCode(code)
	if code == "" {
		return OverrideInfo{}, bizerr.NewCode(CodePromptInvalidInput)
	}
	_, err := dao.SystemPromptOverrides.Ctx(ctx).Data(do.SystemPromptOverrides{
		Code:    code,
		Content: input.Content,
	}).Save()
	if err != nil {
		return OverrideInfo{}, bizerr.WrapCode(err, CodePromptStoreUnavailable)
	}
	return s.GetPromptTemplateOverride(ctx, code)
}

// RestorePromptTemplateOverride removes persisted content for one template.
func (s *daoStore) RestorePromptTemplateOverride(ctx context.Context, code string) (OverrideInfo, error) {
	code = normalizePromptCode(code)
	if code == "" {
		return OverrideInfo{}, bizerr.NewCode(CodePromptInvalidInput)
	}
	_, err := dao.SystemPromptOverrides.Ctx(ctx).
		Where(do.SystemPromptOverrides{Code: code}).
		Delete()
	if err != nil {
		return OverrideInfo{}, bizerr.WrapCode(err, CodePromptStoreUnavailable)
	}
	return OverrideInfo{Code: code}, nil
}

func overrideFromEntity(row *entity.SystemPromptOverrides) OverrideInfo {
	if row == nil {
		return OverrideInfo{}
	}
	return OverrideInfo{
		Code:      row.Code,
		Content:   row.Content,
		CreatedAt: promptUnixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt: promptUnixMilliFromTimePtr(row.UpdatedAt),
	}
}

func promptUnixMilliFromTimePtr(value *time.Time) int64 {
	if value == nil || value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}
