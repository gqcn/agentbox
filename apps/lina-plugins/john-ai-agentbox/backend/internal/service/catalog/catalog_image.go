// This file implements AgentBox coding-image persistence. Delete and update
// checks preserve default images and referenced images so existing agents and
// historical sessions do not lose their configuration projections.

package catalog

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
	"john-ai-agentbox/backend/internal/model/entity"
)

// ListImages returns all coding images ordered for selector presentation.
func (s *serviceImpl) ListImages(ctx context.Context) ([]CodingImageInfo, error) {
	cols := dao.CodingImages.Columns()
	var rows []*entity.CodingImages
	err := dao.CodingImages.Ctx(ctx).
		OrderDesc(cols.IsDefault).
		OrderAsc(cols.Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	items := make([]CodingImageInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, codingImageFromEntity(row))
	}
	return items, nil
}

// CreateImage creates one coding-image profile.
func (s *serviceImpl) CreateImage(ctx context.Context, input CodingImageInput) (*CodingImageInfo, error) {
	normalized := normalizeImageInput(input)
	if err := validateImageInput(normalized); err != nil {
		return nil, err
	}
	id, err := dao.CodingImages.Ctx(ctx).Data(do.CodingImages{
		Name:         normalized.Name,
		ImageRef:     normalized.ImageRef,
		AgentType:    normalized.AgentType,
		DefaultShell: normalized.DefaultShell,
		Notes:        normalized.Notes,
		Enabled:      normalized.Enabled,
		IsDefault:    false,
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return s.getImage(ctx, id)
}

// UpdateImage updates one coding-image profile.
func (s *serviceImpl) UpdateImage(ctx context.Context, id int64, input CodingImageInput) (*CodingImageInfo, error) {
	normalized := normalizeImageInput(input)
	if id <= 0 {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	if err := validateImageInput(normalized); err != nil {
		return nil, err
	}
	current, err := s.getImage(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.IsDefault && current.ImageRef != normalized.ImageRef {
		return nil, bizerr.NewCode(CodeCatalogResourceInUse)
	}
	result, err := dao.CodingImages.Ctx(ctx).
		Where(do.CodingImages{Id: id}).
		Data(do.CodingImages{
			Name:         normalized.Name,
			ImageRef:     normalized.ImageRef,
			AgentType:    normalized.AgentType,
			DefaultShell: normalized.DefaultShell,
			Notes:        normalized.Notes,
			Enabled:      normalized.Enabled,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	return s.getImage(ctx, id)
}

// DeleteImage deletes one unused non-default coding-image profile.
func (s *serviceImpl) DeleteImage(ctx context.Context, id int64) error {
	if id <= 0 {
		return bizerr.NewCode(CodeCatalogInvalidInput)
	}
	current, err := s.getImage(ctx, id)
	if err != nil {
		return err
	}
	if current.IsDefault {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	agentCount, err := dao.CodingAgents.Ctx(ctx).
		Unscoped().
		Where(do.CodingAgents{ImageId: id}).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if agentCount > 0 {
		return bizerr.NewCode(CodeCatalogResourceInUse)
	}
	result, err := dao.CodingImages.Ctx(ctx).
		Where(do.CodingImages{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return bizerr.NewCode(CodeCatalogNotFound)
	}
	return nil
}

// getImage returns one coding-image projection.
func (s *serviceImpl) getImage(ctx context.Context, id int64) (*CodingImageInfo, error) {
	var row *entity.CodingImages
	err := dao.CodingImages.Ctx(ctx).
		Where(do.CodingImages{Id: id}).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	item := codingImageFromEntity(row)
	return &item, nil
}

// normalizeImageInput trims image strings and applies stable defaults.
func normalizeImageInput(input CodingImageInput) CodingImageInput {
	agentType := normalizeAgentType(input.AgentType)
	if agentType == "" {
		agentType = AgentTypeCustom
	}
	defaultShell := strings.TrimSpace(input.DefaultShell)
	if defaultShell == "" {
		defaultShell = DefaultShell
	}
	return CodingImageInput{
		Name:         strings.TrimSpace(input.Name),
		ImageRef:     strings.TrimSpace(input.ImageRef),
		AgentType:    agentType,
		DefaultShell: defaultShell,
		Notes:        strings.TrimSpace(input.Notes),
		Enabled:      input.Enabled,
	}
}

// validateImageInput checks required image fields.
func validateImageInput(input CodingImageInput) error {
	if input.Name == "" || input.ImageRef == "" || input.AgentType == "" {
		return bizerr.NewCode(CodeCatalogInvalidInput)
	}
	return nil
}

// normalizeAgentType constrains image agent types to supported values.
func normalizeAgentType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case AgentTypeClaudeCode:
		return AgentTypeClaudeCode
	case AgentTypeCodex:
		return AgentTypeCodex
	case AgentTypeCustom:
		return AgentTypeCustom
	default:
		return ""
	}
}

// codingImageFromEntity maps generated image entities into service projections.
func codingImageFromEntity(row *entity.CodingImages) CodingImageInfo {
	if row == nil {
		return CodingImageInfo{}
	}
	return CodingImageInfo{
		ID:           row.Id,
		Name:         row.Name,
		ImageRef:     row.ImageRef,
		AgentType:    row.AgentType,
		DefaultShell: row.DefaultShell,
		Notes:        row.Notes,
		Enabled:      row.Enabled,
		IsDefault:    row.IsDefault,
		CreatedAt:    unixMilliFromTimePtr(row.CreatedAt),
		UpdatedAt:    unixMilliFromTimePtr(row.UpdatedAt),
	}
}
