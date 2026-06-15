// This file defines the public workspace API contract surface for the
// AgentBox plugin. Versioned DTOs stay in subpackages while this package keeps
// the controller interface used by source-plugin route registration.

package workspace

import (
	"context"

	v1 "john-ai-agentbox/backend/api/workspace/v1"
)

// IWorkspaceV1 defines AgentBox workspace HTTP handlers.
type IWorkspaceV1 interface {
	// PathSuggestions lists bounded workspace path suggestions for one visible Agent.
	PathSuggestions(ctx context.Context, req *v1.PathSuggestionsReq) (res *v1.PathSuggestionsRes, err error)
	// DirectoryTree lists workspace tree nodes for one visible Agent.
	DirectoryTree(ctx context.Context, req *v1.DirectoryTreeReq) (res *v1.DirectoryTreeRes, err error)
	// FilePreview reads one workspace file preview for one visible Agent.
	FilePreview(ctx context.Context, req *v1.FilePreviewReq) (res *v1.FilePreviewRes, err error)
	// FileSave saves one workspace file for one visible Agent.
	FileSave(ctx context.Context, req *v1.FileSaveReq) (res *v1.FileSaveRes, err error)
	// FileCreate creates one workspace file for one visible Agent.
	FileCreate(ctx context.Context, req *v1.FileCreateReq) (res *v1.FileCreateRes, err error)
	// DirectoryCreate creates one workspace directory for one visible Agent.
	DirectoryCreate(ctx context.Context, req *v1.DirectoryCreateReq) (res *v1.DirectoryCreateRes, err error)
	// FileUpload uploads files into one visible Agent workspace.
	FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error)
	// FileDownload downloads one workspace file after Agent ownership validation.
	FileDownload(ctx context.Context, req *v1.FileDownloadReq) (res *v1.FileDownloadRes, err error)
	// Resource opens one workspace resource after Agent ownership validation.
	Resource(ctx context.Context, req *v1.ResourceReq) (res *v1.ResourceRes, err error)
	// HtmlPreview streams one sandboxed workspace HTML preview after Agent ownership validation.
	HtmlPreview(ctx context.Context, req *v1.HtmlPreviewReq) (res *v1.HtmlPreviewRes, err error)
	// Skills lists global or project skills for one visible Agent.
	Skills(ctx context.Context, req *v1.SkillsReq) (res *v1.SkillsRes, err error)
	// SkillUpload uploads project skills for one visible Agent.
	SkillUpload(ctx context.Context, req *v1.SkillUploadReq) (res *v1.SkillUploadRes, err error)
	// GitStatus returns Git status for one visible Agent workspace path.
	GitStatus(ctx context.Context, req *v1.GitStatusReq) (res *v1.GitStatusRes, err error)
	// GitFile previews a Git file for one visible Agent workspace path.
	GitFile(ctx context.Context, req *v1.GitFileReq) (res *v1.GitFileRes, err error)
	// GitDiff reads a Git diff for one visible Agent workspace path.
	GitDiff(ctx context.Context, req *v1.GitDiffReq) (res *v1.GitDiffRes, err error)
	// GitCommitMessageSuggestion generates a Git commit message suggestion.
	GitCommitMessageSuggestion(ctx context.Context, req *v1.GitCommitMessageSuggestionReq) (res *v1.GitCommitMessageSuggestionRes, err error)
	// GitIndexStage stages repository files for one visible Agent workspace path.
	GitIndexStage(ctx context.Context, req *v1.GitIndexStageReq) (res *v1.GitIndexStageRes, err error)
	// GitIndexUnstage unstages repository files for one visible Agent workspace path.
	GitIndexUnstage(ctx context.Context, req *v1.GitIndexUnstageReq) (res *v1.GitIndexUnstageRes, err error)
	// GitChangesDiscard discards unstaged repository changes for one visible Agent workspace path.
	GitChangesDiscard(ctx context.Context, req *v1.GitChangesDiscardReq) (res *v1.GitChangesDiscardRes, err error)
	// GitCommit creates a Git commit for one visible Agent workspace path.
	GitCommit(ctx context.Context, req *v1.GitCommitReq) (res *v1.GitCommitRes, err error)
}
