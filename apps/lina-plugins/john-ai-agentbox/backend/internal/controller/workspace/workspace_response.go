// This file maps AgentBox workspace service projections to public DTOs while
// keeping runtime and service internals out of HTTP response contracts.

package workspace

import (
	v1 "john-ai-agentbox/backend/api/workspace/v1"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

func toPathSuggestionListResponse(items []workspacesvc.PathSuggestion) []v1.WorkspacePathSuggestion {
	out := make([]v1.WorkspacePathSuggestion, 0, len(items))
	for _, item := range items {
		out = append(out, v1.WorkspacePathSuggestion{
			Name: item.Name,
			Path: item.Path,
		})
	}
	return out
}

func toTreeNodeListResponse(items []workspacesvc.TreeNode) []v1.WorkspaceTreeNode {
	out := make([]v1.WorkspaceTreeNode, 0, len(items))
	for _, item := range items {
		out = append(out, toTreeNodeResponse(item))
	}
	return out
}

func toTreeNodeResponse(item workspacesvc.TreeNode) v1.WorkspaceTreeNode {
	return v1.WorkspaceTreeNode{
		Name:       item.Name,
		Path:       item.Path,
		Type:       item.Type,
		Size:       item.Size,
		ModifiedAt: item.ModifiedAt,
		Expandable: item.Expandable,
		Children:   toTreeNodeListResponse(item.Children),
	}
}

func toFileInfoResponse(item workspacesvc.FileInfo) v1.WorkspaceFileInfo {
	return v1.WorkspaceFileInfo{
		Name:        item.Name,
		Path:        item.Path,
		Type:        item.Type,
		Size:        item.Size,
		ModifiedAt:  item.ModifiedAt,
		ContentType: item.ContentType,
	}
}

func toFilePreviewResponse(item *workspacesvc.FilePreview) *v1.WorkspaceFilePreview {
	if item == nil {
		return nil
	}
	return &v1.WorkspaceFilePreview{
		File:        toFileInfoResponse(item.File),
		PreviewType: item.PreviewType,
		Content:     item.Content,
		Encoding:    item.Encoding,
		ContentHash: item.ContentHash,
		TooLarge:    item.TooLarge,
		DownloadURL: item.DownloadURL,
	}
}

func toUploadResponse(item *workspacesvc.UploadResponse) *v1.WorkspaceUploadResponse {
	if item == nil {
		return nil
	}
	out := make([]v1.WorkspaceFileInfo, 0, len(item.Files))
	for _, file := range item.Files {
		out = append(out, toFileInfoResponse(file))
	}
	return &v1.WorkspaceUploadResponse{Files: out}
}

func toSkillListResponse(item *workspacesvc.SkillListResponse) *v1.WorkspaceSkillListResponse {
	if item == nil {
		return nil
	}
	out := make([]v1.WorkspaceSkillInfo, 0, len(item.Items))
	for _, skill := range item.Items {
		out = append(out, toSkillInfoResponse(skill))
	}
	return &v1.WorkspaceSkillListResponse{
		Scope: item.Scope,
		Path:  item.Path,
		Items: out,
	}
}

func toSkillUploadResponse(item *workspacesvc.SkillUploadResponse) *v1.WorkspaceSkillUploadResponse {
	if item == nil {
		return nil
	}
	out := make([]v1.WorkspaceSkillInfo, 0, len(item.Skills))
	for _, skill := range item.Skills {
		out = append(out, toSkillInfoResponse(skill))
	}
	return &v1.WorkspaceSkillUploadResponse{Skills: out}
}

func toSkillInfoResponse(item workspacesvc.SkillInfo) v1.WorkspaceSkillInfo {
	return v1.WorkspaceSkillInfo{
		Name:        item.Name,
		Description: item.Description,
		Scope:       item.Scope,
		Path:        item.Path,
		Source:      item.Source,
		HasManifest: item.HasManifest,
	}
}

func toGitStatusResponse(item *workspacesvc.GitStatusResponse) *v1.GitStatusResponse {
	if item == nil {
		return nil
	}
	return &v1.GitStatusResponse{
		State:         item.State,
		Path:          item.Path,
		Root:          item.Root,
		Message:       item.Message,
		Changes:       toGitChangeListResponse(item.Changes),
		StagedChanges: toGitChangeListResponse(item.StagedChanges),
		ChangeTree:    toGitTreeNodeListResponse(item.ChangeTree),
		StagedTree:    toGitTreeNodeListResponse(item.StagedTree),
	}
}

func toGitFileResponse(item *workspacesvc.GitFileResponse) *v1.GitFileResponse {
	if item == nil {
		return nil
	}
	file := toFilePreviewResponse(&item.File)
	return &v1.GitFileResponse{
		File:    *file,
		Status:  item.Status,
		Message: item.Message,
	}
}

func toGitDiffResponse(item *workspacesvc.GitDiffResponse) *v1.GitDiffResponse {
	if item == nil {
		return nil
	}
	return &v1.GitDiffResponse{
		Path:            item.Path,
		Status:          item.Status,
		Scope:           item.Scope,
		Diff:            item.Diff,
		OriginalContent: item.OriginalContent,
		ModifiedContent: item.ModifiedContent,
		OriginalPath:    item.OriginalPath,
		ModifiedPath:    item.ModifiedPath,
		Language:        item.Language,
		Message:         item.Message,
	}
}

func toGitCommitMessageSuggestionResponse(item *workspacesvc.GitCommitMessageSuggestionResponse) *v1.GitCommitMessageSuggestionResponse {
	if item == nil {
		return nil
	}
	return &v1.GitCommitMessageSuggestionResponse{
		Message:         item.Message,
		DiffScope:       item.DiffScope,
		TierCode:        item.TierCode,
		ProviderID:      item.ProviderID,
		ProviderName:    item.ProviderName,
		ProviderModelID: item.ProviderModelID,
		ModelName:       item.ModelName,
		Protocol:        item.Protocol,
		Truncated:       item.Truncated,
		GeneratedAt:     item.GeneratedAt,
	}
}

func toGitCommitResponse(item *workspacesvc.GitCommitResponse) *v1.GitCommitResponse {
	if item == nil {
		return nil
	}
	status := toGitStatusResponse(&item.Status)
	return &v1.GitCommitResponse{
		CommitHash: item.CommitHash,
		Pushed:     item.Pushed,
		Status:     *status,
	}
}

func toGitChangeListResponse(items []workspacesvc.GitChange) []v1.GitChange {
	out := make([]v1.GitChange, 0, len(items))
	for _, item := range items {
		out = append(out, v1.GitChange{
			Path:        item.Path,
			OldPath:     item.OldPath,
			Status:      item.Status,
			IndexState:  item.IndexState,
			WorkState:   item.WorkState,
			ChangeScope: item.ChangeScope,
		})
	}
	return out
}

func toGitTreeNodeListResponse(items []workspacesvc.GitTreeNode) []v1.GitTreeNode {
	out := make([]v1.GitTreeNode, 0, len(items))
	for _, item := range items {
		out = append(out, v1.GitTreeNode{
			Name:        item.Name,
			Path:        item.Path,
			OldPath:     item.OldPath,
			Type:        item.Type,
			Status:      item.Status,
			ChangeScope: item.ChangeScope,
			Children:    toGitTreeNodeListResponse(item.Children),
		})
	}
	return out
}
