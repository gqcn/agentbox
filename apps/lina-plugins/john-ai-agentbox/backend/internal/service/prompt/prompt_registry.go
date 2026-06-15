// This file declares immutable AgentBox prompt-template definitions. Defaults
// are code-owned product strategy and are merged with database overrides only
// at service read/render boundaries.

package prompt

const gitCommitMessageDefaultTemplate = `Generate one concise Git commit message for the following {{.DiffScope}} diff.
Rules:
- Output exactly one line.
- Do not include quotes, markdown, bullet points, or explanations.
- Use imperative mood and keep it under 72 characters when possible.
{{.TruncatedNotice}}

Diff:
{{.Diff}}`

// templateDefinition is a code registry entry for one prompt template.
type templateDefinition struct {
	Code           string
	DisplayName    string
	Description    string
	Purpose        string
	TierCode       string
	DefaultContent string
	Variables      []VariableInfo
}

// defaultRegistry returns all built-in prompt template definitions.
func defaultRegistry() map[string]templateDefinition {
	gitCommit := templateDefinition{
		Code:           SystemPromptCodeGitCommitMessage,
		DisplayName:    "Git Commit Message",
		Description:    "Generates one concise Git commit message from staged, unstaged, or untracked repository changes.",
		Purpose:        AIPurposeGitCommitMessage,
		TierCode:       AICapabilityTierBasic,
		DefaultContent: gitCommitMessageDefaultTemplate,
		Variables: []VariableInfo{
			{
				Name:        "DiffScope",
				Description: "Git diff scope selected for generation, usually staged or unstaged.",
				Required:    false,
				SampleValue: "staged",
			},
			{
				Name:        "Diff",
				Description: "Git diff content or safe untracked-file summary used to generate the commit message.",
				Required:    true,
				SampleValue: "diff --git a/web/src/App.tsx b/web/src/App.tsx\n+Add prompt management navigation",
			},
			{
				Name:        "TruncatedNotice",
				Description: "One-line notice that the diff was truncated, empty when the full diff is visible.",
				Required:    false,
				SampleValue: "- The diff is truncated, so summarize only visible changes.",
			},
		},
	}
	return map[string]templateDefinition{
		gitCommit.Code: gitCommit,
	}
}

// templateInfo merges a registry entry and an optional override row.
func templateInfo(def templateDefinition, override OverrideInfo) TemplateInfo {
	content := def.DefaultContent
	if override.Content != "" {
		content = override.Content
	}
	return TemplateInfo{
		Code:           def.Code,
		DisplayName:    def.DisplayName,
		Description:    def.Description,
		Purpose:        def.Purpose,
		TierCode:       def.TierCode,
		DefaultContent: def.DefaultContent,
		Content:        content,
		Variables:      append([]VariableInfo(nil), def.Variables...),
		CreatedAt:      override.CreatedAt,
		UpdatedAt:      override.UpdatedAt,
	}
}
