// This file defines version-one workspace DTOs for the AgentBox plugin. Paths
// are plugin-relative and are published under /x/john-ai-agentbox/api/v1 by
// source-plugin route registration.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PathSuggestionsReq lists workspace path suggestions for one Agent.
type PathSuggestionsReq struct {
	g.Meta `path:"/agents/{id}/workspace/paths" method:"get" tags:"AgentBox Workspace" summary:"List AgentBox workspace path suggestions" dc:"List bounded workspace directory suggestions for one authenticated-user-owned Agent. Runtime-backed suggestions become available after workspace runtime migration."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Query  string `json:"query" dc:"Path search query; omitted means using path parameter if supplied" eg:"project"`
	Path   string `json:"path" dc:"Fallback workspace path query retained for frontend compatibility" eg:"/home/agent/workspace/project"`
}

// PathSuggestionsRes returns workspace path suggestions.
type PathSuggestionsRes = []WorkspacePathSuggestion

// DirectoryTreeReq lists immediate workspace directory entries.
type DirectoryTreeReq struct {
	g.Meta       `path:"/agents/{id}/workspace/tree" method:"get" tags:"AgentBox Workspace" summary:"List AgentBox workspace tree" dc:"List immediate workspace entries for one authenticated-user-owned Agent. Runtime-backed tree data becomes available after workspace runtime migration."`
	ID           string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path         string `json:"path" dc:"Workspace path to list; empty defaults to /home/agent/workspace" eg:"/home/agent/workspace"`
	IncludeFiles bool   `json:"includeFiles" dc:"Whether file entries should be included in the tree" eg:"true"`
}

// DirectoryTreeRes returns workspace tree nodes.
type DirectoryTreeRes = []WorkspaceTreeNode

// FilePreviewReq reads one workspace file preview.
type FilePreviewReq struct {
	g.Meta `path:"/agents/{id}/workspace/file" method:"get" tags:"AgentBox Workspace" summary:"Preview AgentBox workspace file" dc:"Read a safe preview for one file inside an authenticated-user-owned Agent workspace. Runtime-backed file previews become available after workspace runtime migration."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" v:"required" dc:"Workspace file path" eg:"/home/agent/workspace/README.md"`
}

// FilePreviewRes returns workspace file preview details.
type FilePreviewRes = WorkspaceFilePreview

// FileSaveReq updates one text file inside an Agent workspace.
type FileSaveReq struct {
	g.Meta  `path:"/agents/{id}/workspace/file" method:"put" tags:"AgentBox Workspace" summary:"Save AgentBox workspace file" dc:"Update one editable UTF-8 text file inside an authenticated-user-owned Agent workspace or shared directory."`
	ID      string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path    string `json:"path" v:"required" dc:"Workspace file path" eg:"/home/agent/workspace/README.md"`
	Content string `json:"content" dc:"Text content to write into the file" eg:"# README"`
	// Encoding is currently limited to utf-8 to keep editor writes predictable.
	Encoding string `json:"encoding" dc:"Text encoding; omitted defaults to utf-8; only utf-8 is currently accepted" eg:"utf-8"`
	BaseHash string `json:"baseHash" dc:"Content hash returned by preview; omitted disables stale-write protection" eg:"sha256:abc123"`
}

// FileSaveRes returns updated workspace file preview details.
type FileSaveRes = WorkspaceFilePreview

// FileCreateReq creates one empty file inside an Agent workspace directory.
type FileCreateReq struct {
	g.Meta     `path:"/agents/{id}/workspace/files" method:"post" tags:"AgentBox Workspace" summary:"Create AgentBox workspace file" dc:"Create one empty file in a target directory inside an authenticated-user-owned Agent workspace or shared directory."`
	ID         string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	ParentPath string `json:"parentPath" v:"required" dc:"Workspace parent directory where the file should be created" eg:"/home/agent/workspace/project-a"`
	Name       string `json:"name" v:"required" dc:"File name to create as a single safe path segment" eg:"notes.md"`
}

// FileCreateRes returns the created file preview.
type FileCreateRes = WorkspaceFilePreview

// DirectoryCreateReq creates one directory inside an Agent workspace directory.
type DirectoryCreateReq struct {
	g.Meta     `path:"/agents/{id}/workspace/directories" method:"post" tags:"AgentBox Workspace" summary:"Create AgentBox workspace directory" dc:"Create one directory in a target directory inside an authenticated-user-owned Agent workspace or shared directory."`
	ID         string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	ParentPath string `json:"parentPath" v:"required" dc:"Workspace parent directory where the directory should be created" eg:"/home/agent/workspace/project-a"`
	Name       string `json:"name" v:"required" dc:"Directory name to create as a single safe path segment" eg:"docs"`
}

// DirectoryCreateRes returns the created directory details.
type DirectoryCreateRes = WorkspaceFileInfo

// FileUploadReq uploads files into an Agent workspace directory.
type FileUploadReq struct {
	g.Meta `path:"/agents/{id}/workspace/upload" method:"post" mime:"multipart/form-data" tags:"AgentBox Workspace" summary:"Upload AgentBox workspace files" dc:"Upload up to sixteen bounded files into a target directory inside an authenticated-user-owned Agent workspace or shared directory."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" dc:"Workspace target directory; empty defaults to /home/agent/workspace" eg:"/home/agent/workspace/project-a"`
}

// FileUploadRes returns uploaded file details.
type FileUploadRes = WorkspaceUploadResponse

// FileDownloadReq downloads one workspace file.
type FileDownloadReq struct {
	g.Meta `path:"/agents/{id}/workspace/download" method:"get" tags:"AgentBox Workspace" summary:"Download AgentBox workspace file" dc:"Download one runtime file from an authenticated-user-owned Agent workspace or shared directory as an attachment."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" v:"required" dc:"Workspace file path" eg:"/home/agent/workspace/README.md"`
}

// FileDownloadRes is a placeholder because the controller writes binary content directly.
type FileDownloadRes struct{}

// ResourceReq streams one workspace file resource for inline preview or download.
type ResourceReq struct {
	g.Meta      `path:"/agents/{id}/workspace/resources" method:"get" tags:"AgentBox Workspace" summary:"Open AgentBox workspace resource" dc:"Open one runtime file inside an authenticated-user-owned Agent workspace or shared directory as an inline resource or attachment download."`
	ID          string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path        string `json:"path" v:"required" dc:"Workspace or shared file path to open" eg:"/home/agent/workspace/report/run.log"`
	Disposition string `json:"disposition" dc:"Response disposition: inline or attachment; omitted defaults to inline" eg:"inline"`
}

// ResourceRes is a placeholder because the controller writes binary content directly.
type ResourceRes struct{}

// HtmlPreviewReq redirects one workspace HTML file to an isolated preview URL.
type HtmlPreviewReq struct {
	g.Meta `path:"/agents/{id}/workspace/html-previews" method:"get" tags:"AgentBox Workspace" summary:"Open AgentBox workspace HTML preview" dc:"Stream one HTML file inside an authenticated-user-owned Agent workspace or shared directory with restrictive sandbox headers. Linked runtime services, scripts, forms, and same-origin access are intentionally unavailable in this isolated file preview."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" v:"required" dc:"Workspace or shared HTML file path to preview" eg:"/home/agent/workspace/report/index.html"`
}

// HtmlPreviewRes is written directly by the controller after runtime migration.
type HtmlPreviewRes struct{}

// SkillsReq lists global or project skills visible to one Agent.
type SkillsReq struct {
	g.Meta `path:"/agents/{id}/skills" method:"get" tags:"AgentBox Workspace Skills" summary:"List AgentBox workspace skills" dc:"List bounded global or project skills visible from one authenticated-user-owned Agent runtime workspace."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Scope  string `json:"scope" dc:"Skill scope: global or project; omitted means all scopes" eg:"project"`
	Path   string `json:"path" dc:"Workspace project path for project skills" eg:"/home/agent/workspace/project-a"`
	Query  string `json:"query" dc:"Skill name or description filter" eg:"frontend"`
}

// SkillsRes returns visible skills.
type SkillsRes = WorkspaceSkillListResponse

// SkillUploadReq uploads project skills into one Agent workspace.
type SkillUploadReq struct {
	g.Meta `path:"/agents/{id}/skills/upload" method:"post" mime:"multipart/form-data" tags:"AgentBox Workspace Skills" summary:"Upload AgentBox project skills" dc:"Upload one or more Markdown files or ZIP archives as project-level skills for one authenticated-user-owned Agent workspace. Runtime-backed uploads become available after workspace runtime migration."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Scope  string `json:"scope" dc:"Skill scope; currently project is expected after runtime migration" eg:"project"`
	Path   string `json:"path" dc:"Workspace project path for uploaded skills" eg:"/home/agent/workspace/project-a"`
}

// SkillUploadRes returns uploaded skill details.
type SkillUploadRes = WorkspaceSkillUploadResponse

// GitStatusReq gets Git status for one workspace path.
type GitStatusReq struct {
	g.Meta `path:"/agents/{id}/git/status" method:"get" tags:"AgentBox Workspace Git" summary:"Get AgentBox workspace Git status" dc:"Get read-only Git repository status and change tree for one authenticated-user-owned Agent workspace path."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
}

// GitStatusRes returns Git status details.
type GitStatusRes = GitStatusResponse

// GitFileReq previews a file with Git status metadata.
type GitFileReq struct {
	g.Meta `path:"/agents/{id}/git/file" method:"get" tags:"AgentBox Workspace Git" summary:"Preview AgentBox workspace Git file" dc:"Read a runtime-backed file preview from a Git repository and include per-file status metadata for one authenticated-user-owned Agent workspace."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	File   string `json:"file" v:"required" dc:"Repository-relative file path" eg:"src/main.tsx"`
}

// GitFileRes returns Git file preview details.
type GitFileRes = GitFileResponse

// GitDiffReq reads a Git diff for one file.
type GitDiffReq struct {
	g.Meta `path:"/agents/{id}/git/diff" method:"get" tags:"AgentBox Workspace Git" summary:"Get AgentBox workspace Git diff" dc:"Read a runtime-backed unified diff and editable side-by-side text models for one changed file in an authenticated-user-owned Agent workspace repository."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	File   string `json:"file" v:"required" dc:"Repository-relative file path" eg:"src/main.tsx"`
	Scope  string `json:"scope" dc:"Diff scope: staged or unstaged; omitted defaults to unstaged" eg:"unstaged"`
}

// GitDiffRes returns Git diff details.
type GitDiffRes = GitDiffResponse

// GitCommitMessageSuggestionReq generates a commit message suggestion from Git diff.
type GitCommitMessageSuggestionReq struct {
	g.Meta `path:"/agents/{id}/git/commit-message-suggestions" method:"post" tags:"AgentBox Workspace Git" summary:"Generate AgentBox Git commit message suggestion" dc:"Generate one editable commit message suggestion from staged or unstaged Git diff for one authenticated-user-owned Agent workspace. Runtime-backed suggestions become available after workspace runtime migration."`
	ID     string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string `json:"path" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
}

// GitCommitMessageSuggestionRes returns the generated commit message suggestion.
type GitCommitMessageSuggestionRes = GitCommitMessageSuggestionResponse

// GitIndexStageReq stages repository files into the Git index.
type GitIndexStageReq struct {
	g.Meta `path:"/agents/{id}/git/index" method:"put" tags:"AgentBox Workspace Git" summary:"Stage AgentBox workspace Git files" dc:"Add one or more repository-relative files from an authenticated-user-owned Agent workspace Git repository into the Git index. Runtime-backed staging becomes available after workspace runtime migration."`
	ID     string   `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string   `json:"path" v:"required" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	Files  []string `json:"files" dc:"Repository-relative file paths to stage; required unless all is true" eg:"src/main.tsx"`
	All    bool     `json:"all" dc:"Whether to stage all repository changes after runtime migration; false means stage selected files" eg:"false"`
}

// GitIndexStageRes returns refreshed Git status after staging.
type GitIndexStageRes = GitStatusResponse

// GitIndexUnstageReq removes repository files from the Git index.
type GitIndexUnstageReq struct {
	g.Meta `path:"/agents/{id}/git/index" method:"delete" tags:"AgentBox Workspace Git" summary:"Unstage AgentBox workspace Git files" dc:"Remove one or more repository-relative files from an authenticated-user-owned Agent workspace Git repository index. Runtime-backed unstaging becomes available after workspace runtime migration."`
	ID     string   `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string   `json:"path" v:"required" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	Files  []string `json:"files" dc:"Repository-relative file paths to unstage; required unless all is true" eg:"src/main.tsx"`
	All    bool     `json:"all" dc:"Whether to unstage all changes after runtime migration; false means unstage selected files" eg:"false"`
}

// GitIndexUnstageRes returns refreshed Git status after unstaging.
type GitIndexUnstageRes = GitStatusResponse

// GitChangesDiscardReq discards unstaged Git changes.
type GitChangesDiscardReq struct {
	g.Meta `path:"/agents/{id}/git/changes" method:"delete" tags:"AgentBox Workspace Git" summary:"Discard AgentBox workspace Git changes" dc:"Discard unstaged repository-relative worktree changes from an authenticated-user-owned Agent workspace repository. Runtime-backed discard becomes available after workspace runtime migration."`
	ID     string   `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path   string   `json:"path" v:"required" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	Files  []string `json:"files" v:"required" dc:"Repository-relative file paths whose unstaged changes should be discarded" eg:"src/main.tsx"`
}

// GitChangesDiscardRes returns refreshed Git status after discarding changes.
type GitChangesDiscardRes = GitStatusResponse

// GitCommitReq creates a Git commit from staged changes and optionally pushes it.
type GitCommitReq struct {
	g.Meta  `path:"/agents/{id}/git/commits" method:"post" tags:"AgentBox Workspace Git" summary:"Create AgentBox workspace Git commit" dc:"Create a Git commit from currently staged changes in an authenticated-user-owned Agent workspace repository. Runtime-backed commits become available after workspace runtime migration."`
	ID      string `json:"id" v:"required" dc:"Agent ID" eg:"agt-1234567890abcdef"`
	Path    string `json:"path" v:"required" dc:"Workspace path inside a Git repository" eg:"/home/agent/workspace/project-a"`
	Message string `json:"message" v:"required" dc:"Git commit message used after runtime migration" eg:"Update workspace source control"`
	Push    bool   `json:"push" dc:"Whether to push the created commit after runtime migration" eg:"false"`
}

// GitCommitRes returns the created commit hash and refreshed Git status.
type GitCommitRes = GitCommitResponse

// WorkspacePathSuggestion describes one workspace path candidate.
type WorkspacePathSuggestion struct {
	Name string `json:"name" dc:"Suggested directory name" eg:"project-a"`
	Path string `json:"path" dc:"Suggested workspace path" eg:"/home/agent/workspace/project-a"`
}

// WorkspaceTreeNode describes one workspace tree node.
type WorkspaceTreeNode struct {
	Name       string              `json:"name" dc:"Workspace entry name" eg:"README.md"`
	Path       string              `json:"path" dc:"Workspace entry path" eg:"/home/agent/workspace/README.md"`
	Type       string              `json:"type" dc:"Workspace entry type: directory or file" eg:"file"`
	Size       int64               `json:"size,omitempty" dc:"File size in bytes; omitted for directories" eg:"1024"`
	ModifiedAt int64               `json:"modifiedAt,omitempty" dc:"Entry modification time as Unix timestamp in milliseconds" eg:"1704067200000"`
	Expandable bool                `json:"expandable" dc:"Whether the entry can be expanded as a directory" eg:"true"`
	Children   []WorkspaceTreeNode `json:"children,omitempty" dc:"Optional child entries for preloaded directories" eg:"[]"`
}

// WorkspaceFileInfo describes one workspace file or directory.
type WorkspaceFileInfo struct {
	Name        string `json:"name" dc:"Workspace entry name" eg:"README.md"`
	Path        string `json:"path" dc:"Workspace entry path" eg:"/home/agent/workspace/README.md"`
	Type        string `json:"type" dc:"Workspace entry type: directory or file" eg:"file"`
	Size        int64  `json:"size" dc:"File size in bytes; zero for directories or unknown runtime values" eg:"1024"`
	ModifiedAt  int64  `json:"modifiedAt,omitempty" dc:"Entry modification time as Unix timestamp in milliseconds" eg:"1704067200000"`
	ContentType string `json:"contentType,omitempty" dc:"Detected file content type" eg:"text/markdown"`
}

// WorkspaceFilePreview describes one safe workspace file preview.
type WorkspaceFilePreview struct {
	File        WorkspaceFileInfo `json:"file" dc:"Workspace file metadata" eg:"{}"`
	PreviewType string            `json:"previewType" dc:"Preview type: text, image, or unsupported" eg:"text"`
	Content     string            `json:"content,omitempty" dc:"Preview content for text files" eg:"# README"`
	Encoding    string            `json:"encoding,omitempty" dc:"Preview text encoding" eg:"utf-8"`
	ContentHash string            `json:"contentHash,omitempty" dc:"Content hash for stale-write protection" eg:"sha256:abc123"`
	TooLarge    bool              `json:"tooLarge" dc:"Whether the file is too large for inline preview" eg:"false"`
	DownloadURL string            `json:"downloadUrl,omitempty" dc:"Plugin API download URL for the file" eg:"/x/john-ai-agentbox/api/v1/agents/agt-123/workspace/download?path=/home/agent/workspace/README.md"`
}

// WorkspaceUploadResponse describes uploaded workspace files.
type WorkspaceUploadResponse struct {
	Files []WorkspaceFileInfo `json:"files" dc:"Uploaded file metadata" eg:"[]"`
}

// WorkspaceSkillInfo describes one AgentBox skill.
type WorkspaceSkillInfo struct {
	Name        string `json:"name" dc:"Skill name" eg:"frontend-review"`
	Description string `json:"description,omitempty" dc:"Optional skill description" eg:"Review frontend code"`
	Scope       string `json:"scope" dc:"Skill scope: global or project" eg:"project"`
	Path        string `json:"path" dc:"Skill file or directory path" eg:"/home/agent/workspace/.agents/skills/frontend-review"`
	Source      string `json:"source" dc:"Skill source label" eg:"project"`
	HasManifest bool   `json:"hasManifest" dc:"Whether the skill has a manifest file" eg:"true"`
}

// WorkspaceSkillListResponse describes visible AgentBox skills for one scope.
type WorkspaceSkillListResponse struct {
	Scope string               `json:"scope" dc:"Returned skill scope: global or project" eg:"project"`
	Path  string               `json:"path,omitempty" dc:"Workspace project path used for project skills" eg:"/home/agent/workspace/project-a"`
	Items []WorkspaceSkillInfo `json:"items" dc:"Visible skills" eg:"[]"`
}

// WorkspaceSkillUploadResponse describes uploaded AgentBox project skills.
type WorkspaceSkillUploadResponse struct {
	Skills []WorkspaceSkillInfo `json:"skills" dc:"Uploaded skills" eg:"[]"`
}

// GitChange describes one Git path change.
type GitChange struct {
	Path        string `json:"path" dc:"Repository-relative file path" eg:"src/main.tsx"`
	OldPath     string `json:"oldPath,omitempty" dc:"Previous path for renamed files" eg:"src/old.tsx"`
	Status      string `json:"status" dc:"Git status code or label" eg:"modified"`
	IndexState  string `json:"indexState,omitempty" dc:"Index status code" eg:"M"`
	WorkState   string `json:"workState,omitempty" dc:"Worktree status code" eg:"M"`
	ChangeScope string `json:"changeScope,omitempty" dc:"Change scope: staged or unstaged" eg:"unstaged"`
}

// GitTreeNode describes one changed path in a Git tree.
type GitTreeNode struct {
	Name        string        `json:"name" dc:"Tree node name" eg:"main.tsx"`
	Path        string        `json:"path" dc:"Repository-relative node path" eg:"src/main.tsx"`
	OldPath     string        `json:"oldPath,omitempty" dc:"Previous path for renamed files" eg:"src/old.tsx"`
	Type        string        `json:"type" dc:"Workspace node type: directory or file" eg:"file"`
	Status      string        `json:"status,omitempty" dc:"Git status code or label" eg:"modified"`
	ChangeScope string        `json:"changeScope,omitempty" dc:"Change scope: staged or unstaged" eg:"unstaged"`
	Children    []GitTreeNode `json:"children,omitempty" dc:"Child changed paths" eg:"[]"`
}

// GitStatusResponse describes Git repository status.
type GitStatusResponse struct {
	State         string        `json:"state" dc:"Repository state: ok, clean, not_repo, or error" eg:"ok"`
	Path          string        `json:"path" dc:"Workspace path used for the Git command" eg:"/home/agent/workspace/project-a"`
	Root          string        `json:"root,omitempty" dc:"Repository root path" eg:"/home/agent/workspace/project-a"`
	Message       string        `json:"message,omitempty" dc:"Optional Git status message" eg:"Repository is clean"`
	Changes       []GitChange   `json:"changes,omitempty" dc:"Unstaged and untracked changes" eg:"[]"`
	StagedChanges []GitChange   `json:"stagedChanges,omitempty" dc:"Staged changes" eg:"[]"`
	ChangeTree    []GitTreeNode `json:"changeTree,omitempty" dc:"Unstaged change tree" eg:"[]"`
	StagedTree    []GitTreeNode `json:"stagedTree,omitempty" dc:"Staged change tree" eg:"[]"`
}

// GitFileResponse describes a Git file preview.
type GitFileResponse struct {
	File    WorkspaceFilePreview `json:"file" dc:"Workspace file preview" eg:"{}"`
	Status  string               `json:"status,omitempty" dc:"Git file status" eg:"modified"`
	Message string               `json:"message,omitempty" dc:"Optional Git file message" eg:"File has unstaged changes"`
}

// GitDiffResponse describes a per-file Git diff.
type GitDiffResponse struct {
	Path            string `json:"path" dc:"Repository-relative file path" eg:"src/main.tsx"`
	Status          string `json:"status,omitempty" dc:"Git file status" eg:"modified"`
	Scope           string `json:"scope,omitempty" dc:"Diff scope: staged or unstaged" eg:"unstaged"`
	Diff            string `json:"diff" dc:"Unified diff text" eg:"diff --git a/src/main.tsx b/src/main.tsx"`
	OriginalContent string `json:"originalContent" dc:"Original file content" eg:"old content"`
	ModifiedContent string `json:"modifiedContent" dc:"Modified file content" eg:"new content"`
	OriginalPath    string `json:"originalPath" dc:"Original repository-relative file path" eg:"src/main.tsx"`
	ModifiedPath    string `json:"modifiedPath" dc:"Modified repository-relative file path" eg:"src/main.tsx"`
	Language        string `json:"language" dc:"Detected language for diff viewer" eg:"typescript"`
	Message         string `json:"message,omitempty" dc:"Optional Git diff message" eg:"No diff available"`
}

// GitCommitMessageSuggestionResponse describes a generated commit message.
type GitCommitMessageSuggestionResponse struct {
	Message         string `json:"message" dc:"Suggested Git commit message" eg:"Update workspace source control"`
	DiffScope       string `json:"diffScope" dc:"Diff scope used for generation: staged or unstaged" eg:"staged"`
	TierCode        string `json:"tierCode" dc:"AgentBox AI capability tier used for generation" eg:"basic"`
	ProviderID      int64  `json:"providerId" dc:"Provider ID used for generation" eg:"1"`
	ProviderName    string `json:"providerName" dc:"Provider name used for generation" eg:"OpenAI"`
	ProviderModelID int64  `json:"providerModelId" dc:"Provider model ID used for generation" eg:"1"`
	ModelName       string `json:"modelName" dc:"Model name used for generation" eg:"gpt-4.1"`
	Protocol        string `json:"protocol" dc:"Provider protocol: openai or anthropic" eg:"openai"`
	Truncated       bool   `json:"truncated" dc:"Whether the diff was truncated before generation" eg:"false"`
	GeneratedAt     int64  `json:"generatedAt" dc:"Generation time as Unix timestamp in milliseconds" eg:"1704067200000"`
}

// GitCommitResponse describes a created Git commit.
type GitCommitResponse struct {
	CommitHash string            `json:"commitHash" dc:"Created Git commit hash" eg:"abc123def456"`
	Pushed     bool              `json:"pushed" dc:"Whether the commit was pushed" eg:"false"`
	Status     GitStatusResponse `json:"status" dc:"Refreshed Git status after commit" eg:"{}"`
}
