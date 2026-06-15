// Package prompt implements AgentBox prompt-template registry, override
// management, preview rendering, and runtime prompt rendering. Persisted
// overrides live only in plugin-owned john_ai_agentbox_ tables.
package prompt

import (
	"context"
	"sort"
	"strings"
)

import "lina-core/pkg/bizerr"

const (
	// SystemPromptCodeGitCommitMessage identifies the Git commit message template.
	SystemPromptCodeGitCommitMessage = "git_commit_message"
	// AIPurposeGitCommitMessage identifies Git commit message AI invocations.
	AIPurposeGitCommitMessage = "git_commit_message"
	// AICapabilityTierBasic is the low-cost tier for simple text generation.
	AICapabilityTierBasic = "basic"
)

// VariableInfo describes one variable accepted by a prompt template.
type VariableInfo struct {
	Name        string
	Description string
	Required    bool
	SampleValue string
}

// TemplateInfo describes one registered prompt template after merging code
// registry defaults with any persisted content override.
type TemplateInfo struct {
	Code           string
	DisplayName    string
	Description    string
	Purpose        string
	TierCode       string
	DefaultContent string
	Content        string
	Variables      []VariableInfo
	CreatedAt      int64
	UpdatedAt      int64
}

// OverrideInfo describes one persisted content override row.
type OverrideInfo struct {
	Code      string
	Content   string
	CreatedAt int64
	UpdatedAt int64
}

// UpdateInput carries editable content for a registered prompt template.
type UpdateInput struct {
	Content string
}

// PreviewInput carries draft content and optional variables for preview rendering.
type PreviewInput struct {
	Content   string
	Variables map[string]string
}

// PreviewOutput returns the rendered prompt preview.
type PreviewOutput struct {
	RenderedPrompt string
}

// OverrideStore persists content overrides for registered prompt templates.
type OverrideStore interface {
	// GetPromptTemplateOverride returns one override by prompt template code. A
	// CodePromptNotFound error means callers should fall back to registry defaults.
	GetPromptTemplateOverride(ctx context.Context, code string) (OverrideInfo, error)
	// UpsertPromptTemplateOverride stores editable content for a registered
	// prompt template and returns the persisted row.
	UpsertPromptTemplateOverride(ctx context.Context, code string, input UpdateInput) (OverrideInfo, error)
	// RestorePromptTemplateOverride clears saved content for a prompt template.
	// The returned projection may have zero timestamps when the row is removed.
	RestorePromptTemplateOverride(ctx context.Context, code string) (OverrideInfo, error)
}

// Renderer renders runtime prompts from a registered template and variables.
type Renderer interface {
	// Render returns the current prompt text for code after applying persisted
	// override fallback rules. Missing required variables or template execution
	// errors are returned and callers must not proceed to AI invocation.
	Render(ctx context.Context, code string, variables map[string]string) (string, error)
}

// Service manages AgentBox prompt templates and also renders runtime prompts.
type Service interface {
	Renderer
	// ListTemplates returns all registered prompt templates in stable order.
	ListTemplates(ctx context.Context) ([]TemplateInfo, error)
	// GetTemplate returns one registered prompt template by code.
	GetTemplate(ctx context.Context, code string) (TemplateInfo, error)
	// UpdateTemplate validates and saves content for one template.
	UpdateTemplate(ctx context.Context, code string, input UpdateInput) (TemplateInfo, error)
	// RestoreTemplate restores one template to default behavior.
	RestoreTemplate(ctx context.Context, code string) (TemplateInfo, error)
	// PreviewTemplate renders a template draft with sample or caller-supplied
	// variables without persisting content or calling an AI provider.
	PreviewTemplate(ctx context.Context, code string, input PreviewInput) (PreviewOutput, error)
}

// serviceImpl combines immutable registry definitions with persisted overrides.
type serviceImpl struct {
	store    OverrideStore
	registry map[string]templateDefinition
	order    []string
}

var _ Service = (*serviceImpl)(nil)

// New creates a prompt service with an explicitly injected override store.
func New(store OverrideStore) (Service, error) {
	if store == nil {
		return nil, bizerr.NewCode(CodePromptStoreUnavailable)
	}
	registry := defaultRegistry()
	order := make([]string, 0, len(registry))
	for code := range registry {
		order = append(order, code)
	}
	sort.Strings(order)
	return &serviceImpl{
		store:    store,
		registry: registry,
		order:    order,
	}, nil
}

// ListTemplates returns every registry template merged with persisted overrides.
func (s *serviceImpl) ListTemplates(ctx context.Context) ([]TemplateInfo, error) {
	items := make([]TemplateInfo, 0, len(s.order))
	for _, code := range s.order {
		item, err := s.GetTemplate(ctx, code)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetTemplate returns one registry template merged with any persisted override.
func (s *serviceImpl) GetTemplate(ctx context.Context, code string) (TemplateInfo, error) {
	def, err := s.definition(code)
	if err != nil {
		return TemplateInfo{}, err
	}
	override, exists, err := s.loadOverride(ctx, def.Code)
	if err != nil {
		return TemplateInfo{}, err
	}
	if !exists {
		return templateInfo(def, OverrideInfo{}), nil
	}
	return templateInfo(def, override), nil
}

// UpdateTemplate validates and saves content for one template.
func (s *serviceImpl) UpdateTemplate(ctx context.Context, code string, input UpdateInput) (TemplateInfo, error) {
	def, err := s.definition(code)
	if err != nil {
		return TemplateInfo{}, err
	}
	if strings.TrimSpace(input.Content) == "" {
		return TemplateInfo{}, bizerr.NewCode(CodePromptInvalidInput)
	}
	if err := validateTemplateContent(def, input.Content, true); err != nil {
		return TemplateInfo{}, err
	}
	override, err := s.store.UpsertPromptTemplateOverride(ctx, def.Code, input)
	if err != nil {
		return TemplateInfo{}, err
	}
	return templateInfo(def, override), nil
}

// RestoreTemplate clears saved content and returns the default projection.
func (s *serviceImpl) RestoreTemplate(ctx context.Context, code string) (TemplateInfo, error) {
	def, err := s.definition(code)
	if err != nil {
		return TemplateInfo{}, err
	}
	override, err := s.store.RestorePromptTemplateOverride(ctx, def.Code)
	if err != nil {
		return TemplateInfo{}, err
	}
	return templateInfo(def, override), nil
}

// PreviewTemplate renders a draft or current template without persistence.
func (s *serviceImpl) PreviewTemplate(ctx context.Context, code string, input PreviewInput) (PreviewOutput, error) {
	def, err := s.definition(code)
	if err != nil {
		return PreviewOutput{}, err
	}
	content := input.Content
	if strings.TrimSpace(content) == "" {
		current, err := s.GetTemplate(ctx, def.Code)
		if err != nil {
			return PreviewOutput{}, err
		}
		content = current.Content
	}
	rendered, err := renderTemplateContent(def, content, previewVariables(def, input.Variables), false)
	if err != nil {
		return PreviewOutput{}, err
	}
	return PreviewOutput{RenderedPrompt: rendered}, nil
}

// Render resolves a template's current content and renders it with variables.
func (s *serviceImpl) Render(ctx context.Context, code string, variables map[string]string) (string, error) {
	def, err := s.definition(code)
	if err != nil {
		return "", err
	}
	info, err := s.GetTemplate(ctx, def.Code)
	if err != nil {
		return "", err
	}
	return renderTemplateContent(def, info.Content, variables, true)
}

// definition resolves one template definition by normalized code.
func (s *serviceImpl) definition(code string) (templateDefinition, error) {
	code = normalizePromptCode(code)
	if code == "" {
		return templateDefinition{}, bizerr.NewCode(CodePromptInvalidInput)
	}
	def, ok := s.registry[code]
	if !ok {
		return templateDefinition{}, bizerr.NewCode(CodePromptNotFound)
	}
	return def, nil
}

// loadOverride treats missing override rows as a default-template fallback.
func (s *serviceImpl) loadOverride(ctx context.Context, code string) (OverrideInfo, bool, error) {
	override, err := s.store.GetPromptTemplateOverride(ctx, code)
	if err != nil {
		if bizerr.Is(err, CodePromptNotFound) {
			return OverrideInfo{}, false, nil
		}
		return OverrideInfo{}, false, err
	}
	return override, true, nil
}

func normalizePromptCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
