// This file implements version-one workspace request handlers. Every handler
// reads the authenticated AgentBox user ID from context before delegating to
// service methods that enforce Agent ownership.

package workspace

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/workspace/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// PathSuggestions returns workspace path candidates for one visible Agent.
func (c *ControllerV1) PathSuggestions(ctx context.Context, req *v1.PathSuggestionsReq) (res *v1.PathSuggestionsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.workspaceSvc.PathSuggestions(ctx, userID, req.ID, req.Query, req.Path)
	if err != nil {
		return nil, err
	}
	out := toPathSuggestionListResponse(items)
	return (*v1.PathSuggestionsRes)(&out), nil
}

// DirectoryTree returns workspace tree nodes for one visible Agent.
func (c *ControllerV1) DirectoryTree(ctx context.Context, req *v1.DirectoryTreeReq) (res *v1.DirectoryTreeRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.workspaceSvc.DirectoryTree(ctx, userID, req.ID, req.Path, req.IncludeFiles)
	if err != nil {
		return nil, err
	}
	out := toTreeNodeListResponse(items)
	return (*v1.DirectoryTreeRes)(&out), nil
}

// FilePreview returns a workspace file preview for one visible Agent.
func (c *ControllerV1) FilePreview(ctx context.Context, req *v1.FilePreviewReq) (res *v1.FilePreviewRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.FilePreview(ctx, userID, req.ID, req.Path)
	if err != nil {
		return nil, err
	}
	return toFilePreviewResponse(item), nil
}

// FileSave saves a workspace file for one visible Agent.
func (c *ControllerV1) FileSave(ctx context.Context, req *v1.FileSaveReq) (res *v1.FileSaveRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.FileSave(ctx, userID, req.ID, workspacesvc.FileSaveInput{
		Path:     req.Path,
		Content:  req.Content,
		Encoding: req.Encoding,
		BaseHash: req.BaseHash,
	})
	if err != nil {
		return nil, err
	}
	return toFilePreviewResponse(item), nil
}

// FileCreate creates a workspace file for one visible Agent.
func (c *ControllerV1) FileCreate(ctx context.Context, req *v1.FileCreateReq) (res *v1.FileCreateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.FileCreate(ctx, userID, req.ID, workspacesvc.CreateEntryInput{
		ParentPath: req.ParentPath,
		Name:       req.Name,
	})
	if err != nil {
		return nil, err
	}
	return toFilePreviewResponse(item), nil
}

// DirectoryCreate creates a workspace directory for one visible Agent.
func (c *ControllerV1) DirectoryCreate(ctx context.Context, req *v1.DirectoryCreateReq) (res *v1.DirectoryCreateRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.DirectoryCreate(ctx, userID, req.ID, workspacesvc.CreateEntryInput{
		ParentPath: req.ParentPath,
		Name:       req.Name,
	})
	if err != nil {
		return nil, err
	}
	out := toFileInfoResponse(*item)
	return (*v1.DirectoryCreateRes)(&out), nil
}

// FileUpload validates workspace upload ownership for one visible Agent.
func (c *ControllerV1) FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	files, cleanup, err := workspaceUploadFilesFromRequest(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cleanupErr := cleanup(); err == nil && cleanupErr != nil {
			err = bizerr.WrapCode(cleanupErr, workspacesvc.CodeWorkspaceRuntimeUnavailable)
		}
	}()
	item, err := c.workspaceSvc.FileUpload(ctx, userID, req.ID, workspacesvc.FileUploadInput{
		Path:  req.Path,
		Files: files,
	})
	if err != nil {
		return nil, err
	}
	return toUploadResponse(item), nil
}

func workspaceUploadFilesFromRequest(ctx context.Context) ([]workspacesvc.UploadFile, func() error, error) {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return nil, func() error { return nil }, bizerr.NewCode(workspacesvc.CodeWorkspaceInvalidInput)
	}
	uploads := request.GetUploadFiles("file")
	if len(uploads) == 0 {
		uploads = request.GetUploadFiles("files")
	}
	return workspaceUploadFilesFromUploads(uploads)
}

func workspaceUploadFilesFromUploads(uploads ghttp.UploadFiles) ([]workspacesvc.UploadFile, func() error, error) {
	if len(uploads) == 0 {
		return nil, func() error { return nil }, bizerr.NewCode(workspacesvc.CodeWorkspaceInvalidInput)
	}
	files := make([]workspacesvc.UploadFile, 0, len(uploads))
	closers := make([]func() error, 0, len(uploads))
	for _, upload := range uploads {
		if upload == nil || upload.FileHeader == nil {
			if closeErr := closeUploadFiles(closers); closeErr != nil {
				return nil, func() error { return nil }, bizerr.WrapCode(closeErr, workspacesvc.CodeWorkspaceRuntimeUnavailable)
			}
			return nil, func() error { return nil }, bizerr.NewCode(workspacesvc.CodeWorkspaceInvalidInput)
		}
		opened, err := upload.Open()
		if err != nil {
			if closeErr := closeUploadFiles(closers); closeErr != nil {
				return nil, func() error { return nil }, bizerr.WrapCode(closeErr, workspacesvc.CodeWorkspaceRuntimeUnavailable)
			}
			return nil, func() error { return nil }, bizerr.WrapCode(err, workspacesvc.CodeWorkspaceRuntimeUnavailable)
		}
		closers = append(closers, func() error {
			return opened.Close()
		})
		files = append(files, workspacesvc.UploadFile{
			Name:   upload.Filename,
			Reader: opened,
			Size:   upload.Size,
		})
	}
	return files, func() error { return closeUploadFiles(closers) }, nil
}

func closeUploadFiles(closers []func() error) error {
	var firstErr error
	for _, closer := range closers {
		if err := closer(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// FileDownload streams a workspace file download for one visible Agent.
func (c *ControllerV1) FileDownload(ctx context.Context, req *v1.FileDownloadReq) (res *v1.FileDownloadRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.FileDownload(ctx, userID, req.ID, req.Path)
	if err != nil {
		return nil, err
	}
	return nil, writeWorkspaceFileStream(ctx, item)
}

// Resource streams a workspace resource for one visible Agent.
func (c *ControllerV1) Resource(ctx context.Context, req *v1.ResourceReq) (res *v1.ResourceRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.Resource(ctx, userID, req.ID, req.Path, req.Disposition)
	if err != nil {
		return nil, err
	}
	return nil, writeWorkspaceFileStream(ctx, item)
}

// HtmlPreview streams an isolated workspace HTML preview for one visible Agent.
func (c *ControllerV1) HtmlPreview(ctx context.Context, req *v1.HtmlPreviewReq) (res *v1.HtmlPreviewRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.HtmlPreview(ctx, userID, req.ID, req.Path)
	if err != nil {
		return nil, err
	}
	return nil, writeWorkspaceHTMLPreviewStream(ctx, item)
}

// Skills lists runtime-backed skills for one visible Agent.
func (c *ControllerV1) Skills(ctx context.Context, req *v1.SkillsReq) (res *v1.SkillsRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.Skills(ctx, userID, req.ID, workspacesvc.SkillListInput{
		Scope: req.Scope,
		Path:  req.Path,
		Query: req.Query,
	})
	if err != nil {
		return nil, err
	}
	return toSkillListResponse(item), nil
}

// SkillUpload validates project skill upload ownership for one visible Agent.
func (c *ControllerV1) SkillUpload(ctx context.Context, req *v1.SkillUploadReq) (res *v1.SkillUploadRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.SkillUpload(ctx, userID, req.ID, workspacesvc.SkillUploadInput{
		Scope: req.Scope,
		Path:  req.Path,
	})
	if err != nil {
		return nil, err
	}
	return toSkillUploadResponse(item), nil
}

// GitStatus returns Git status for one visible Agent workspace path.
func (c *ControllerV1) GitStatus(ctx context.Context, req *v1.GitStatusReq) (res *v1.GitStatusRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitStatus(ctx, userID, req.ID, req.Path)
	if err != nil {
		return nil, err
	}
	return toGitStatusResponse(item), nil
}

// GitFile returns a Git file preview for one visible Agent workspace path.
func (c *ControllerV1) GitFile(ctx context.Context, req *v1.GitFileReq) (res *v1.GitFileRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitFile(ctx, userID, req.ID, req.Path, req.File)
	if err != nil {
		return nil, err
	}
	return toGitFileResponse(item), nil
}

// GitDiff returns a Git diff for one visible Agent workspace path.
func (c *ControllerV1) GitDiff(ctx context.Context, req *v1.GitDiffReq) (res *v1.GitDiffRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitDiff(ctx, userID, req.ID, req.Path, req.File, req.Scope)
	if err != nil {
		return nil, err
	}
	return toGitDiffResponse(item), nil
}

// GitCommitMessageSuggestion generates a commit message suggestion for one visible Agent workspace path.
func (c *ControllerV1) GitCommitMessageSuggestion(ctx context.Context, req *v1.GitCommitMessageSuggestionReq) (res *v1.GitCommitMessageSuggestionRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitCommitMessageSuggestion(ctx, userID, req.ID, req.Path)
	if err != nil {
		return nil, err
	}
	return toGitCommitMessageSuggestionResponse(item), nil
}

// GitIndexStage stages Git files for one visible Agent workspace path.
func (c *ControllerV1) GitIndexStage(ctx context.Context, req *v1.GitIndexStageReq) (res *v1.GitIndexStageRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitIndexStage(ctx, userID, req.ID, workspacesvc.GitIndexInput{
		Path:  req.Path,
		Files: req.Files,
		All:   req.All,
	})
	if err != nil {
		return nil, err
	}
	return toGitStatusResponse(item), nil
}

// GitIndexUnstage unstages Git files for one visible Agent workspace path.
func (c *ControllerV1) GitIndexUnstage(ctx context.Context, req *v1.GitIndexUnstageReq) (res *v1.GitIndexUnstageRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitIndexUnstage(ctx, userID, req.ID, workspacesvc.GitIndexInput{
		Path:  req.Path,
		Files: req.Files,
		All:   req.All,
	})
	if err != nil {
		return nil, err
	}
	return toGitStatusResponse(item), nil
}

// GitChangesDiscard discards Git changes for one visible Agent workspace path.
func (c *ControllerV1) GitChangesDiscard(ctx context.Context, req *v1.GitChangesDiscardReq) (res *v1.GitChangesDiscardRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitChangesDiscard(ctx, userID, req.ID, workspacesvc.GitDiscardInput{
		Path:  req.Path,
		Files: req.Files,
	})
	if err != nil {
		return nil, err
	}
	return toGitStatusResponse(item), nil
}

// GitCommit creates a Git commit for one visible Agent workspace path.
func (c *ControllerV1) GitCommit(ctx context.Context, req *v1.GitCommitReq) (res *v1.GitCommitRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.workspaceSvc.GitCommit(ctx, userID, req.ID, workspacesvc.GitCommitInput{
		Path:    req.Path,
		Message: req.Message,
		Push:    req.Push,
	})
	if err != nil {
		return nil, err
	}
	return toGitCommitResponse(item), nil
}
