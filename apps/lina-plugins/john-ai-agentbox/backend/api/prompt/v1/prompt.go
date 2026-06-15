// This file defines version-one prompt-template DTOs for the AgentBox plugin.
// Paths are plugin-relative and are published under /x/john-ai-agentbox/api/v1
// by source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq lists registered AgentBox prompt templates.
type ListReq struct {
	g.Meta `path:"/prompt-templates" method:"get" tags:"AgentBox Prompts" summary:"List AgentBox prompt templates" dc:"List all registered AgentBox prompt templates by merging code registry defaults with persisted plugin-owned content overrides. Time-point fields are Unix timestamp in milliseconds."`
}

// ListRes returns prompt template records.
type ListRes = []PromptTemplateInfo

// DetailReq reads one registered prompt template.
type DetailReq struct {
	g.Meta `path:"/prompt-templates/{code}" method:"get" tags:"AgentBox Prompts" summary:"Get AgentBox prompt template" dc:"Get one AgentBox prompt template by code, including default content, current content, declared variables, and override timestamps as Unix timestamp in milliseconds."`
	Code   string `json:"code" v:"required" dc:"Prompt template code, currently git_commit_message" eg:"git_commit_message"`
}

// DetailRes returns one prompt template record.
type DetailRes = PromptTemplateInfo

// UpdateReq updates one prompt template content override.
type UpdateReq struct {
	g.Meta  `path:"/prompt-templates/{code}" method:"put" tags:"AgentBox Prompts" summary:"Update AgentBox prompt template" dc:"Save content for one registered AgentBox prompt template after syntax and required-variable validation. Saved content is used directly by the owning plugin feature."`
	Code    string `json:"code" v:"required" dc:"Prompt template code, currently git_commit_message" eg:"git_commit_message"`
	Content string `json:"content" dc:"Go text/template content to save. Content must be non-empty and reference required variables such as Diff for git_commit_message." eg:"Write one Conventional Commit for {{.Diff}}"`
}

// UpdateRes returns the updated prompt template.
type UpdateRes = PromptTemplateInfo

// RestoreReq restores one prompt template to default behavior.
type RestoreReq struct {
	g.Meta `path:"/prompt-templates/{code}/restore" method:"post" tags:"AgentBox Prompts" summary:"Restore AgentBox prompt template" dc:"Clear the saved content override for one registered AgentBox prompt template, returning the default-content template projection with timestamps as Unix timestamp in milliseconds."`
	Code   string `json:"code" v:"required" dc:"Prompt template code, currently git_commit_message" eg:"git_commit_message"`
}

// RestoreRes returns the restored prompt template.
type RestoreRes = PromptTemplateInfo

// PreviewReq renders a prompt template preview without persistence.
type PreviewReq struct {
	g.Meta    `path:"/prompt-templates/{code}/previews" method:"post" tags:"AgentBox Prompts" summary:"Preview AgentBox prompt template" dc:"Render a current or draft AgentBox prompt template with supplied variables or safe sample variables. This action does not persist content and does not invoke an AI model."`
	Code      string            `json:"code" v:"required" dc:"Prompt template code, currently git_commit_message" eg:"git_commit_message"`
	Content   string            `json:"content" dc:"Draft Go text/template content to preview; omitted or empty means preview the current template content" eg:"Write a concise commit message for {{.Diff}}"`
	Variables map[string]string `json:"variables" dc:"Optional template variables for preview; omitted keys fall back to safe sample values" eg:"{}"`
}

// PreviewRes returns the rendered prompt preview.
type PreviewRes struct {
	RenderedPrompt string `json:"renderedPrompt" dc:"Rendered prompt text produced without invoking an AI model" eg:"Generate one concise Git commit message for the staged diff"`
}

// PromptTemplateVariableInfo describes one template variable.
type PromptTemplateVariableInfo struct {
	Name        string `json:"name" dc:"Template variable name used with Go text/template dot notation" eg:"Diff"`
	Description string `json:"description" dc:"Human-readable variable meaning and expected content" eg:"Git diff content used to generate a commit message"`
	Required    bool   `json:"required" dc:"Whether this variable must be referenced and supplied when rendering the template" eg:"true"`
	SampleValue string `json:"sampleValue" dc:"Safe sample value used by preview when the caller does not provide a value" eg:"diff --git a/app.go b/app.go"`
}

// PromptTemplateInfo describes one registered prompt template.
type PromptTemplateInfo struct {
	Code           string                       `json:"code" dc:"Prompt template code" eg:"git_commit_message"`
	DisplayName    string                       `json:"displayName" dc:"Human-readable template display name" eg:"Git Commit Message"`
	Description    string                       `json:"description" dc:"Template purpose and management scope" eg:"Generates one concise Git commit message from repository diff"`
	Purpose        string                       `json:"purpose" dc:"AI invocation purpose associated with this template" eg:"git_commit_message"`
	TierCode       string                       `json:"tierCode" dc:"Default AI capability tier used by the owning AgentBox feature" eg:"basic"`
	DefaultContent string                       `json:"defaultContent" dc:"Built-in default template content from code registry" eg:"Generate one concise Git commit message for {{.DiffScope}} diff"`
	Content        string                       `json:"content" dc:"Template content currently used for rendering; falls back to defaultContent when no saved override exists" eg:"Write a Conventional Commit for {{.Diff}}"`
	Variables      []PromptTemplateVariableInfo `json:"variables" dc:"Declared variables available to this template" eg:"[]"`
	CreatedAt      int64                        `json:"createdAt" dc:"Override creation time as Unix timestamp in milliseconds; 0 when no override row exists" eg:"1704067200000"`
	UpdatedAt      int64                        `json:"updatedAt" dc:"Override update time as Unix timestamp in milliseconds; 0 when no override row exists" eg:"1704067201000"`
}
