// This file implements AgentBox coding-agent persistence with AgentBox-user
// ownership checks. All list, detail, update, image switch, and delete paths
// include user_id constraints before returning or mutating resource data.

package catalog

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"lina-core/pkg/bizerr"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
)

var (
	agentImageLostPaths      = []string{"/etc", "/opt", "/usr", "/var"}
	agentImagePreservedPaths = []string{"/home", "/root", "/home/agent/workspace", "/home/agent/shared"}
)

// ListUserAgents returns non-deleted coding agents owned by one AgentBox user.
func (s *serviceImpl) ListUserAgents(ctx context.Context, userID string) ([]AgentInfo, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	var rows []*agentJoinRow
	err := dao.CodingAgents.Ctx(ctx).
		As("a").
		Fields(agentSelectFields()).
		LeftJoin(dao.AiProviders.Table()+" p", "p.id = a.provider_id").
		LeftJoin(dao.CodingImages.Table()+" i", "i.id = a.image_id").
		LeftJoin(dao.AgentRuntimes.Table()+" r", "r.agent_id = a.id").
		Where("a.user_id", userID).
		OrderDesc("a.updated_at").
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	items := make([]AgentInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, agentInfoFromJoinRow(row))
	}
	return items, nil
}

// CreateUserAgent creates one coding agent owned by one AgentBox user.
func (s *serviceImpl) CreateUserAgent(ctx context.Context, userID string, input AgentInput) (*AgentInfo, error) {
	userID = strings.TrimSpace(userID)
	normalized := normalizeAgentInput(input)
	if userID == "" || normalized.ImageID <= 0 {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	image, err := s.validateAgentInput(ctx, normalized)
	if err != nil {
		return nil, err
	}
	agentID := newAgentID()
	_, err = dao.CodingAgents.Ctx(ctx).Data(do.CodingAgents{
		Id:            agentID,
		UserId:        userID,
		Name:          normalized.Name,
		ProviderId:    normalized.ProviderID,
		ModelName:     normalized.ModelName,
		ModelProtocol: normalized.ModelProtocol,
		ImageId:       normalized.ImageID,
		AgentType:     image.AgentType,
		IconKey:       normalized.IconKey,
		Notes:         normalized.Notes,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	return s.GetUserAgent(ctx, userID, agentID)
}

// GetUserAgent returns one coding agent owned by one AgentBox user.
func (s *serviceImpl) GetUserAgent(ctx context.Context, userID string, agentID string) (*AgentInfo, error) {
	userID = strings.TrimSpace(userID)
	agentID = strings.TrimSpace(agentID)
	if userID == "" || agentID == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	var row *agentJoinRow
	err := dao.CodingAgents.Ctx(ctx).
		As("a").
		Fields(agentSelectFields()).
		LeftJoin(dao.AiProviders.Table()+" p", "p.id = a.provider_id").
		LeftJoin(dao.CodingImages.Table()+" i", "i.id = a.image_id").
		LeftJoin(dao.AgentRuntimes.Table()+" r", "r.agent_id = a.id").
		Where("a.id", agentID).
		Where("a.user_id", userID).
		Scan(&row)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	item := agentInfoFromJoinRow(row)
	return &item, nil
}

// UpdateUserAgent updates one coding agent owned by one AgentBox user.
func (s *serviceImpl) UpdateUserAgent(ctx context.Context, userID string, agentID string, input AgentInput) (*AgentInfo, error) {
	current, err := s.GetUserAgent(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	normalized := normalizeAgentInput(input)
	normalized.ImageID = current.ImageID
	image, err := s.validateAgentInput(ctx, normalized)
	if err != nil {
		return nil, err
	}
	result, err := dao.CodingAgents.Ctx(ctx).
		Where(do.CodingAgents{Id: strings.TrimSpace(agentID), UserId: strings.TrimSpace(userID)}).
		Data(do.CodingAgents{
			Name:          normalized.Name,
			ProviderId:    normalized.ProviderID,
			ModelName:     normalized.ModelName,
			ModelProtocol: normalized.ModelProtocol,
			AgentType:     image.AgentType,
			IconKey:       normalized.IconKey,
			Notes:         normalized.Notes,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	return s.GetUserAgent(ctx, userID, agentID)
}

// SetUserAgentImage switches one coding agent's image after ownership validation.
func (s *serviceImpl) SetUserAgentImage(ctx context.Context, userID string, agentID string, imageID int64) (*ChangeAgentImageOutput, error) {
	if _, err := s.GetUserAgent(ctx, userID, agentID); err != nil {
		return nil, err
	}
	image, err := s.getImage(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if !image.Enabled {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	result, err := dao.CodingAgents.Ctx(ctx).
		Where(do.CodingAgents{Id: strings.TrimSpace(agentID), UserId: strings.TrimSpace(userID)}).
		Data(do.CodingAgents{
			ImageId:   imageID,
			AgentType: image.AgentType,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return nil, bizerr.NewCode(CodeCatalogNotFound)
	}
	agent, err := s.GetUserAgent(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	return &ChangeAgentImageOutput{
		Agent:          *agent,
		LostPaths:      append([]string(nil), agentImageLostPaths...),
		PreservedPaths: append([]string(nil), agentImagePreservedPaths...),
	}, nil
}

// DeleteUserAgent soft-deletes one coding agent after ownership validation.
func (s *serviceImpl) DeleteUserAgent(ctx context.Context, userID string, agentID string) error {
	userID = strings.TrimSpace(userID)
	agentID = strings.TrimSpace(agentID)
	if userID == "" || agentID == "" {
		return bizerr.NewCode(CodeCatalogInvalidInput)
	}
	result, err := dao.CodingAgents.Ctx(ctx).
		Where(do.CodingAgents{Id: agentID, UserId: userID}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr == nil && affected == 0 {
		return bizerr.NewCode(CodeCatalogNotFound)
	}
	return nil
}

// validateAgentInput checks provider, model, and image references.
func (s *serviceImpl) validateAgentInput(ctx context.Context, input AgentInput) (*CodingImageInfo, error) {
	if input.Name == "" || input.ProviderID <= 0 || input.ModelName == "" || input.ModelProtocol == "" || input.ImageID <= 0 || input.AgentType == "" {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	if _, err := s.getProviderRecord(ctx, input.ProviderID); err != nil {
		return nil, err
	}
	modelCount, err := dao.ProviderModels.Ctx(ctx).
		Where(do.ProviderModels{
			ProviderId: input.ProviderID,
			Name:       input.ModelName,
			Protocol:   input.ModelProtocol,
		}).
		Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeCatalogStoreUnavailable)
	}
	if modelCount == 0 {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	image, err := s.getImage(ctx, input.ImageID)
	if err != nil {
		return nil, err
	}
	if !image.Enabled {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	if image.AgentType != input.AgentType {
		return nil, bizerr.NewCode(CodeCatalogInvalidInput)
	}
	return image, nil
}

// normalizeAgentInput trims agent fields and normalizes enum-like values.
func normalizeAgentInput(input AgentInput) AgentInput {
	return AgentInput{
		Name:          strings.TrimSpace(input.Name),
		ProviderID:    input.ProviderID,
		ModelName:     strings.TrimSpace(input.ModelName),
		ModelProtocol: normalizeProtocol(input.ModelProtocol),
		ImageID:       input.ImageID,
		AgentType:     normalizeAgentType(input.AgentType),
		IconKey:       strings.TrimSpace(input.IconKey),
		Notes:         strings.TrimSpace(input.Notes),
	}
}

// newAgentID creates an opaque AgentBox public identifier.
func newAgentID() string {
	return "agt-" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
