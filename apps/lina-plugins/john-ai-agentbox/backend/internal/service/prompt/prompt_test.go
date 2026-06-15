// This file verifies prompt registry fallback, content override precedence,
// validation failures, preview rendering, restore behavior, and strict runtime
// variable handling without requiring a database.

package prompt

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestPromptServiceDefaultFallback verifies a missing override row falls back
// to the code registry default template.
func TestPromptServiceDefaultFallback(t *testing.T) {
	ctx := context.Background()
	service := newPromptTestService(t, &fakePromptStore{})

	item, err := service.GetTemplate(ctx, SystemPromptCodeGitCommitMessage)
	if err != nil {
		t.Fatal(err)
	}
	if item.Code != SystemPromptCodeGitCommitMessage {
		t.Fatalf("template fallback = %#v", item)
	}
	if item.Content != item.DefaultContent || !strings.Contains(item.Content, "{{.Diff}}") {
		t.Fatalf("template content = %q", item.Content)
	}
}

// TestPromptServiceSavedTemplateRender verifies saved content is used for
// runtime rendering.
func TestPromptServiceSavedTemplateRender(t *testing.T) {
	ctx := context.Background()
	store := &fakePromptStore{
		override: OverrideInfo{
			Code:      SystemPromptCodeGitCommitMessage,
			Content:   "saved {{.DiffScope}} {{.Diff}}",
			CreatedAt: 1704067200000,
			UpdatedAt: 1704067201000,
		},
	}
	service := newPromptTestService(t, store)

	rendered, err := service.Render(ctx, SystemPromptCodeGitCommitMessage, map[string]string{
		"DiffScope":       "staged",
		"Diff":            "diff content",
		"TruncatedNotice": "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "saved staged diff content" {
		t.Fatalf("rendered = %q", rendered)
	}
}

// TestPromptServiceUpdateValidationRejectsInvalidTemplates verifies custom
// template validation runs before persistence.
func TestPromptServiceUpdateValidationRejectsInvalidTemplates(t *testing.T) {
	ctx := context.Background()
	store := &fakePromptStore{}
	service := newPromptTestService(t, store)

	cases := []struct {
		name    string
		content string
	}{
		{name: "empty", content: " "},
		{name: "syntax", content: "{{.Diff"},
		{name: "unknown", content: "{{.Unknown}} {{.Diff}}"},
		{name: "missing diff", content: "scope {{.DiffScope}}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.UpdateTemplate(ctx, SystemPromptCodeGitCommitMessage, UpdateInput{
				Content: tc.content,
			})
			if err == nil || !bizerr.Is(err, CodePromptInvalidInput) {
				t.Fatalf("error = %v, want CodePromptInvalidInput", err)
			}
			if store.upsertCount != 0 {
				t.Fatalf("invalid template was persisted %d time(s)", store.upsertCount)
			}
		})
	}
}

// TestPromptServicePreviewDraftDoesNotPersist verifies preview renders a draft
// template with sample variables and does not write through the store.
func TestPromptServicePreviewDraftDoesNotPersist(t *testing.T) {
	ctx := context.Background()
	store := &fakePromptStore{}
	service := newPromptTestService(t, store)

	preview, err := service.PreviewTemplate(ctx, SystemPromptCodeGitCommitMessage, PreviewInput{
		Content: "preview {{.DiffScope}} {{.Diff}}",
		Variables: map[string]string{
			"DiffScope": "unstaged",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preview.RenderedPrompt, "preview unstaged") || !strings.Contains(preview.RenderedPrompt, "diff --git") {
		t.Fatalf("preview = %q", preview.RenderedPrompt)
	}
	if store.upsertCount != 0 || store.restoreCount != 0 {
		t.Fatalf("preview persisted state: upsert=%d restore=%d", store.upsertCount, store.restoreCount)
	}
}

// TestPromptServiceUpdateRestoreAndRenderFailures verifies valid updates,
// restoring defaults, and strict missing runtime variables.
func TestPromptServiceUpdateRestoreAndRenderFailures(t *testing.T) {
	ctx := context.Background()
	store := &fakePromptStore{}
	service := newPromptTestService(t, store)

	updated, err := service.UpdateTemplate(ctx, SystemPromptCodeGitCommitMessage, UpdateInput{
		Content: "commit {{.Diff}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Content != "commit {{.Diff}}" {
		t.Fatalf("updated template = %#v", updated)
	}

	if _, err := service.Render(ctx, SystemPromptCodeGitCommitMessage, map[string]string{}); err == nil || !bizerr.Is(err, CodePromptInvalidInput) {
		t.Fatalf("missing runtime variable error = %v", err)
	}

	restored, err := service.RestoreTemplate(ctx, SystemPromptCodeGitCommitMessage)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Content != restored.DefaultContent {
		t.Fatalf("restored template = %#v", restored)
	}
}

// TestReferencedTemplateVariablesIncludesBranches verifies validation catches
// unknown fields even when they are inside a branch that might not execute.
func TestReferencedTemplateVariablesIncludesBranches(t *testing.T) {
	tmpl, err := parsePromptTemplate("test", `{{if false}}{{.Hidden}}{{end}}{{.Diff}}`)
	if err != nil {
		t.Fatal(err)
	}
	refs := referencedTemplateVariables(tmpl)
	if _, ok := refs["Hidden"]; !ok {
		t.Fatalf("refs = %s, want Hidden", formatVariableNames(refs))
	}
	if _, ok := refs["Diff"]; !ok {
		t.Fatalf("refs = %s, want Diff", formatVariableNames(refs))
	}
}

// newPromptTestService creates a prompt service for tests.
func newPromptTestService(t *testing.T, store OverrideStore) Service {
	t.Helper()
	service, err := New(store)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

type fakePromptStore struct {
	override     OverrideInfo
	hasOverride  bool
	upsertCount  int
	restoreCount int
}

// GetPromptTemplateOverride returns a configured fake override or not-found.
func (s *fakePromptStore) GetPromptTemplateOverride(_ context.Context, _ string) (OverrideInfo, error) {
	if s.hasOverride || s.override.Code != "" {
		return s.override, nil
	}
	return OverrideInfo{}, bizerr.NewCode(CodePromptNotFound)
}

// UpsertPromptTemplateOverride records an override in memory.
func (s *fakePromptStore) UpsertPromptTemplateOverride(_ context.Context, code string, input UpdateInput) (OverrideInfo, error) {
	s.upsertCount++
	s.hasOverride = true
	s.override = OverrideInfo{
		Code:      code,
		Content:   input.Content,
		CreatedAt: 1704067200000,
		UpdatedAt: 1704067201000,
	}
	return s.override, nil
}

// RestorePromptTemplateOverride disables and clears the fake override.
func (s *fakePromptStore) RestorePromptTemplateOverride(_ context.Context, code string) (OverrideInfo, error) {
	s.restoreCount++
	s.hasOverride = true
	s.override = OverrideInfo{
		Code: code,
	}
	return s.override, nil
}
