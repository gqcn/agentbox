// Package workspace owns AgentBox workspace JSON behavior. It enforces
// AgentBox user ownership before delegating to a narrow runtime backend for
// tree, preview, upload, download, resource streaming, isolated HTML preview,
// basic file mutation, read-only skill listing, and runtime Git status,
// diff, index, discard, and commit operations; skill upload operations remain
// unavailable until their runtime slice is migrated.
package workspace

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

const (
	// DefaultWorkspacePreviewLimit caps text previews returned through workspace APIs.
	DefaultWorkspacePreviewLimit = int64(1024 * 1024)
	// DefaultWorkspaceUploadFileLimit caps one multipart workspace upload file.
	DefaultWorkspaceUploadFileLimit = int64(10 * 1024 * 1024)
	// DefaultWorkspaceUploadCountLimit caps uploaded files in one request.
	DefaultWorkspaceUploadCountLimit = 16
	// DefaultWorkspaceSkillListLimit caps one skills list response.
	DefaultWorkspaceSkillListLimit = 200
	// DefaultWorkspaceSkillManifestLimit caps one SKILL.md metadata read.
	DefaultWorkspaceSkillManifestLimit = int64(256 * 1024)
	// WorkspaceNodeDirectory marks a directory tree node.
	WorkspaceNodeDirectory = "directory"
	// WorkspaceNodeFile marks a file tree node.
	WorkspaceNodeFile = "file"
	// WorkspacePreviewText marks a text preview.
	WorkspacePreviewText = "text"
	// WorkspacePreviewImage marks an image preview.
	WorkspacePreviewImage = "image"
	// WorkspacePreviewUnsupported marks unsupported previews.
	WorkspacePreviewUnsupported = "unsupported"
	// DefaultWorkspaceRootPath is the default container-side Agent workspace root.
	DefaultWorkspaceRootPath = "/home/agent/workspace"
	// DefaultSharedRootPath is the container-side shared directory root for one user.
	DefaultSharedRootPath = "/home/agent/shared"
	// ResourceDispositionInline renders safe workspace resources in the browser.
	ResourceDispositionInline = "inline"
	// ResourceDispositionAttachment forces workspace resources to download.
	ResourceDispositionAttachment = "attachment"
	// SkillScopeGlobal marks user-global AgentBox skills.
	SkillScopeGlobal = "global"
	// SkillScopeProject marks project-local AgentBox skills.
	SkillScopeProject = "project"
	// SkillManifestName is the metadata file name inside one skill directory.
	SkillManifestName = "SKILL.md"
	// GitRepositoryStateOK marks an available Git repository.
	GitRepositoryStateOK = "ok"
	// GitRepositoryStateClean marks a clean Git repository.
	GitRepositoryStateClean = "clean"
	// GitRepositoryStateNotRepo marks a workspace path outside a Git repository.
	GitRepositoryStateNotRepo = "not_repo"
	// GitRepositoryStateError marks a Git inspection failure.
	GitRepositoryStateError = "error"
	// GitDiffScopeStaged selects staged Git diff content.
	GitDiffScopeStaged = "staged"
	// GitDiffScopeUnstaged selects unstaged Git diff content.
	GitDiffScopeUnstaged = "unstaged"
	// GitChangeScopeStaged marks index changes.
	GitChangeScopeStaged = "staged"
	// GitChangeScopeUnstaged marks worktree changes.
	GitChangeScopeUnstaged = "unstaged"
	// GitChangeTypeModified marks modified files.
	GitChangeTypeModified = "modified"
	// GitChangeTypeAdded marks added files.
	GitChangeTypeAdded = "added"
	// GitChangeTypeDeleted marks deleted files.
	GitChangeTypeDeleted = "deleted"
	// GitChangeTypeRenamed marks renamed files.
	GitChangeTypeRenamed = "renamed"
	// GitChangeTypeCopied marks copied files.
	GitChangeTypeCopied = "copied"
	// GitChangeTypeTypeChanged marks type-changed files.
	GitChangeTypeTypeChanged = "type_changed"
	// GitChangeTypeUnmerged marks unmerged files.
	GitChangeTypeUnmerged = "unmerged"
	// GitChangeTypeUntracked marks untracked files.
	GitChangeTypeUntracked = "untracked"
	// GitChangeTypeUnknown marks unknown Git states.
	GitChangeTypeUnknown = "unknown"
)

// PathSuggestion is the service-level workspace path candidate projection.
type PathSuggestion struct {
	Name string
	Path string
}

// TreeNode is the service-level workspace tree projection.
type TreeNode struct {
	Name       string
	Path       string
	Type       string
	Size       int64
	ModifiedAt int64
	Expandable bool
	Children   []TreeNode
}

// FileInfo is the service-level workspace file metadata projection.
type FileInfo struct {
	Name        string
	Path        string
	Type        string
	Size        int64
	ModifiedAt  int64
	ContentType string
}

// FilePreview is the service-level workspace file preview projection.
type FilePreview struct {
	File        FileInfo
	PreviewType string
	Content     string
	Encoding    string
	ContentHash string
	TooLarge    bool
	DownloadURL string
}

// FileStream is a runtime-backed workspace file stream. Callers own Reader and
// must close it after copying response content.
type FileStream struct {
	File        FileInfo
	Reader      io.ReadCloser
	Disposition string
}

// FileSaveInput carries workspace file save parameters.
type FileSaveInput struct {
	Path     string
	Content  string
	Encoding string
	BaseHash string
}

// RuntimePathEntry is the runtime-level workspace path projection.
type RuntimePathEntry struct {
	Name       string
	Path       string
	Type       string
	Size       int64
	ModifiedAt time.Time
	LinkTarget string
}

// RuntimeFile is a runtime-level workspace file stream projection.
type RuntimeFile struct {
	Entry  RuntimePathEntry
	Reader io.ReadCloser
}

// RuntimeWriteFileInput carries runtime file write parameters.
type RuntimeWriteFileInput struct {
	Path    string
	Content []byte
}

// RuntimeUploadFileInput carries one runtime upload file into a workspace directory.
type RuntimeUploadFileInput struct {
	DirectoryPath string
	Name          string
	Content       []byte
}

// RuntimeCreateEntryInput carries runtime file or directory creation parameters.
type RuntimeCreateEntryInput struct {
	ParentPath string
	Name       string
	Type       string
}

// RuntimeGitStatus is the runtime-level Git status projection.
type RuntimeGitStatus struct {
	Path           string
	RepositoryRoot string
	Porcelain      string
	NotRepository  bool
}

// RuntimeGitFile is the runtime-level Git file preview projection.
type RuntimeGitFile struct {
	File          RuntimePathEntry
	Content       []byte
	Status        string
	Message       string
	NotRepository bool
	Deleted       bool
}

// RuntimeGitDiff is the runtime-level per-file Git diff projection.
type RuntimeGitDiff struct {
	Path            string
	Status          string
	Scope           string
	Diff            string
	OriginalContent string
	ModifiedContent string
	OriginalPath    string
	ModifiedPath    string
	Language        string
	Message         string
	NotRepository   bool
}

// RuntimeGitIndexInput carries runtime Git index mutation parameters.
type RuntimeGitIndexInput struct {
	Path  string
	Files []string
	All   bool
}

// RuntimeGitDiscardInput carries runtime Git discard mutation parameters.
type RuntimeGitDiscardInput struct {
	Path  string
	Files []string
}

// RuntimeGitCommitInput carries runtime Git commit mutation parameters.
type RuntimeGitCommitInput struct {
	Path    string
	Message string
	Push    bool
}

// RuntimeGitCommitResult describes a created runtime Git commit.
type RuntimeGitCommitResult struct {
	CommitHash string
	Pushed     bool
	Status     RuntimeGitStatus
}

// RuntimeBackend reads workspace state from one visible Agent runtime. The
// backend must scope every operation to the authenticated AgentBox user and the
// requested Agent, and must reject or hide containers that do not carry the
// plugin's ownership labels.
type RuntimeBackend interface {
	// WorkspacePathStat resolves and stats one workspace path for a visible Agent.
	WorkspacePathStat(ctx context.Context, userID string, agentID string, workspacePath string) (*RuntimePathEntry, error)
	// WorkspaceDirectoryEntries returns immediate entries under one workspace directory.
	WorkspaceDirectoryEntries(ctx context.Context, userID string, agentID string, workspacePath string, includeFiles bool) ([]RuntimePathEntry, error)
	// WorkspaceReadFile reads one non-directory workspace file for a visible Agent.
	WorkspaceReadFile(ctx context.Context, userID string, agentID string, workspacePath string) ([]byte, *RuntimePathEntry, error)
	// WorkspaceOpenFile opens one non-directory workspace file stream for a visible Agent.
	WorkspaceOpenFile(ctx context.Context, userID string, agentID string, workspacePath string) (*RuntimeFile, error)
	// WorkspaceWriteFile writes one non-directory workspace file for a visible Agent.
	WorkspaceWriteFile(ctx context.Context, userID string, agentID string, input RuntimeWriteFileInput) (*RuntimePathEntry, error)
	// WorkspaceUploadFile creates or overwrites one file under a visible workspace directory.
	WorkspaceUploadFile(ctx context.Context, userID string, agentID string, input RuntimeUploadFileInput) (*RuntimePathEntry, error)
	// WorkspaceCreateEntry creates one file or directory under a visible workspace directory.
	WorkspaceCreateEntry(ctx context.Context, userID string, agentID string, input RuntimeCreateEntryInput) (*RuntimePathEntry, error)
	// WorkspaceGitStatus reads Git status for a visible workspace path without mutating the repository.
	WorkspaceGitStatus(ctx context.Context, userID string, agentID string, workspacePath string) (*RuntimeGitStatus, error)
	// WorkspaceGitFile reads one repository-relative file preview without mutating the repository.
	WorkspaceGitFile(ctx context.Context, userID string, agentID string, workspacePath string, file string) (*RuntimeGitFile, error)
	// WorkspaceGitDiff reads one repository-relative file diff without mutating the repository.
	WorkspaceGitDiff(ctx context.Context, userID string, agentID string, workspacePath string, file string, scope string) (*RuntimeGitDiff, error)
	// WorkspaceGitStage stages selected or all repository changes for a visible Agent runtime.
	WorkspaceGitStage(ctx context.Context, userID string, agentID string, input RuntimeGitIndexInput) (*RuntimeGitStatus, error)
	// WorkspaceGitUnstage unstages selected or all repository changes for a visible Agent runtime.
	WorkspaceGitUnstage(ctx context.Context, userID string, agentID string, input RuntimeGitIndexInput) (*RuntimeGitStatus, error)
	// WorkspaceGitDiscard discards selected worktree changes for a visible Agent runtime.
	WorkspaceGitDiscard(ctx context.Context, userID string, agentID string, input RuntimeGitDiscardInput) (*RuntimeGitStatus, error)
	// WorkspaceGitCommit creates a commit from staged changes for a visible Agent runtime.
	WorkspaceGitCommit(ctx context.Context, userID string, agentID string, input RuntimeGitCommitInput) (*RuntimeGitCommitResult, error)
	// WorkspaceSkills lists bounded skill metadata for one visible Agent runtime.
	WorkspaceSkills(ctx context.Context, userID string, agentID string, scope string, workspacePath string) ([]SkillInfo, error)
}

// CreateEntryInput carries workspace file or directory creation parameters.
type CreateEntryInput struct {
	ParentPath string
	Name       string
}

// UploadFile carries one workspace upload file stream. Callers retain
// ownership of Reader and are responsible for closing it when applicable.
type UploadFile struct {
	Name   string
	Reader io.Reader
	Size   int64
}

// FileUploadInput carries workspace upload parameters.
type FileUploadInput struct {
	Path  string
	Files []UploadFile
}

// UploadResponse is the service-level workspace upload projection.
type UploadResponse struct {
	Files []FileInfo
}

// SkillListInput carries AgentBox skill list parameters.
type SkillListInput struct {
	Scope string
	Path  string
	Query string
}

// SkillInfo is the service-level AgentBox skill projection.
type SkillInfo struct {
	Name        string
	Description string
	Scope       string
	Path        string
	Source      string
	HasManifest bool
}

// SkillListResponse is the service-level AgentBox skill list projection.
type SkillListResponse struct {
	Scope string
	Path  string
	Items []SkillInfo
}

// SkillUploadInput carries project skill upload parameters.
type SkillUploadInput struct {
	Scope string
	Path  string
}

// SkillUploadResponse is the service-level AgentBox skill upload projection.
type SkillUploadResponse struct {
	Skills []SkillInfo
}

// GitChange is the service-level Git path change projection.
type GitChange struct {
	Path        string
	OldPath     string
	Status      string
	IndexState  string
	WorkState   string
	ChangeScope string
}

// GitTreeNode is the service-level Git changed-path tree projection.
type GitTreeNode struct {
	Name        string
	Path        string
	OldPath     string
	Type        string
	Status      string
	ChangeScope string
	Children    []GitTreeNode
}

// GitStatusResponse is the service-level Git status projection.
type GitStatusResponse struct {
	State         string
	Path          string
	Root          string
	Message       string
	Changes       []GitChange
	StagedChanges []GitChange
	ChangeTree    []GitTreeNode
	StagedTree    []GitTreeNode
}

// GitFileResponse is the service-level Git file preview projection.
type GitFileResponse struct {
	File    FilePreview
	Status  string
	Message string
}

// GitDiffResponse is the service-level per-file Git diff projection.
type GitDiffResponse struct {
	Path            string
	Status          string
	Scope           string
	Diff            string
	OriginalContent string
	ModifiedContent string
	OriginalPath    string
	ModifiedPath    string
	Language        string
	Message         string
}

// GitCommitMessageSuggestionResponse is the service-level AI commit message projection.
type GitCommitMessageSuggestionResponse struct {
	Message         string
	DiffScope       string
	TierCode        string
	ProviderID      int64
	ProviderName    string
	ProviderModelID int64
	ModelName       string
	Protocol        string
	Truncated       bool
	GeneratedAt     int64
}

// GitIndexInput carries Git index mutation parameters.
type GitIndexInput struct {
	Path  string
	Files []string
	All   bool
}

// GitDiscardInput carries Git discard mutation parameters.
type GitDiscardInput struct {
	Path  string
	Files []string
}

// GitCommitInput carries Git commit mutation parameters.
type GitCommitInput struct {
	Path    string
	Message string
	Push    bool
}

// GitCommitResponse is the service-level Git commit projection.
type GitCommitResponse struct {
	CommitHash string
	Pushed     bool
	Status     GitStatusResponse
}

// Config contains pure value settings for workspace validation and bounded
// runtime reads.
type Config struct {
	WorkspaceRootPath       string
	SharedRootPath          string
	PreviewLimitBytes       int64
	UploadFileLimitBytes    int64
	UploadCountLimit        int
	SkillListLimit          int
	SkillManifestLimitBytes int64
}

// Service defines AgentBox workspace JSON behavior. Methods first validate the
// current AgentBox user's Agent ownership and then delegate to runtime-backed
// workspace implementations when available.
type Service interface {
	// PathSuggestions returns bounded path suggestions for one visible Agent.
	// When a runtime backend is injected it reads immediate runtime directory
	// entries; otherwise visible Agents receive a structured runtime-unavailable
	// error.
	PathSuggestions(ctx context.Context, userID string, agentID string, query string, fallbackPath string) ([]PathSuggestion, error)
	// DirectoryTree returns immediate tree nodes for one visible Agent. When a
	// runtime backend is injected it reads runtime directory entries; otherwise
	// visible Agents receive a structured runtime-unavailable error.
	DirectoryTree(ctx context.Context, userID string, agentID string, path string, includeFiles bool) ([]TreeNode, error)
	// FilePreview validates file path visibility before returning a runtime-backed preview.
	FilePreview(ctx context.Context, userID string, agentID string, path string) (*FilePreview, error)
	// FileSave validates file path visibility before saving runtime-backed file content.
	FileSave(ctx context.Context, userID string, agentID string, input FileSaveInput) (*FilePreview, error)
	// FileCreate validates parent directory visibility before creating a runtime-backed file.
	FileCreate(ctx context.Context, userID string, agentID string, input CreateEntryInput) (*FilePreview, error)
	// DirectoryCreate validates parent directory visibility before creating a runtime-backed directory.
	DirectoryCreate(ctx context.Context, userID string, agentID string, input CreateEntryInput) (*FileInfo, error)
	// FileUpload validates target directory visibility before uploading bounded runtime-backed files.
	FileUpload(ctx context.Context, userID string, agentID string, input FileUploadInput) (*UploadResponse, error)
	// FileDownload validates file visibility before returning a runtime-backed attachment stream.
	FileDownload(ctx context.Context, userID string, agentID string, path string) (*FileStream, error)
	// Resource validates file visibility before returning a runtime-backed inline or attachment stream.
	Resource(ctx context.Context, userID string, agentID string, path string, disposition string) (*FileStream, error)
	// HtmlPreview validates HTML file visibility before returning a runtime-backed isolated preview stream.
	HtmlPreview(ctx context.Context, userID string, agentID string, path string) (*FileStream, error)
	// Skills validates Agent and workspace visibility before listing runtime-backed skills.
	Skills(ctx context.Context, userID string, agentID string, input SkillListInput) (*SkillListResponse, error)
	// SkillUpload validates Agent and workspace visibility before uploading runtime-backed project skills.
	SkillUpload(ctx context.Context, userID string, agentID string, input SkillUploadInput) (*SkillUploadResponse, error)
	// GitStatus validates repository path visibility before returning runtime-backed Git status.
	GitStatus(ctx context.Context, userID string, agentID string, path string) (*GitStatusResponse, error)
	// GitFile validates repository and file paths before returning a runtime-backed Git file preview.
	GitFile(ctx context.Context, userID string, agentID string, path string, file string) (*GitFileResponse, error)
	// GitDiff validates repository and file paths before returning runtime-backed Git diff content.
	GitDiff(ctx context.Context, userID string, agentID string, path string, file string, scope string) (*GitDiffResponse, error)
	// GitCommitMessageSuggestion validates repository path visibility before generating a runtime-backed suggestion.
	GitCommitMessageSuggestion(ctx context.Context, userID string, agentID string, path string) (*GitCommitMessageSuggestionResponse, error)
	// GitIndexStage validates repository path visibility before staging files.
	GitIndexStage(ctx context.Context, userID string, agentID string, input GitIndexInput) (*GitStatusResponse, error)
	// GitIndexUnstage validates repository path visibility before unstaging files.
	GitIndexUnstage(ctx context.Context, userID string, agentID string, input GitIndexInput) (*GitStatusResponse, error)
	// GitChangesDiscard validates repository path visibility before discarding worktree changes.
	GitChangesDiscard(ctx context.Context, userID string, agentID string, input GitDiscardInput) (*GitStatusResponse, error)
	// GitCommit validates repository path visibility before creating a commit.
	GitCommit(ctx context.Context, userID string, agentID string, input GitCommitInput) (*GitCommitResponse, error)
}

// serviceImpl is the default workspace service implementation.
type serviceImpl struct {
	accessSvc      accesssvc.Service
	runtimeBackend RuntimeBackend
	config         Config
}

var _ Service = (*serviceImpl)(nil)

// New creates a workspace service using explicit dependency injection. Passing
// a nil runtimeBackend keeps runtime-backed operations structurally unavailable
// after ownership checks, which is useful for controlled startup degradation and
// focused unit tests.
func New(accessSvc accesssvc.Service, runtimeBackend RuntimeBackend, configs ...Config) (Service, error) {
	if accessSvc == nil {
		return nil, gerror.New("agentbox workspace access service is required")
	}
	config := Config{}
	if len(configs) > 0 {
		config = configs[0]
	}
	config = NormalizeConfig(config)
	return &serviceImpl{accessSvc: accessSvc, runtimeBackend: runtimeBackend, config: config}, nil
}

// DefaultConfig returns conservative workspace runtime limits.
func DefaultConfig() Config {
	return Config{
		WorkspaceRootPath:       DefaultWorkspaceRootPath,
		SharedRootPath:          DefaultSharedRootPath,
		PreviewLimitBytes:       DefaultWorkspacePreviewLimit,
		UploadFileLimitBytes:    DefaultWorkspaceUploadFileLimit,
		UploadCountLimit:        DefaultWorkspaceUploadCountLimit,
		SkillListLimit:          DefaultWorkspaceSkillListLimit,
		SkillManifestLimitBytes: DefaultWorkspaceSkillManifestLimit,
	}
}

// NormalizeConfig applies workspace defaults and validates container root paths.
func NormalizeConfig(config Config) Config {
	defaults := DefaultConfig()
	if config.WorkspaceRootPath == "" {
		config.WorkspaceRootPath = defaults.WorkspaceRootPath
	}
	if config.SharedRootPath == "" {
		config.SharedRootPath = defaults.SharedRootPath
	}
	if config.PreviewLimitBytes <= 0 {
		config.PreviewLimitBytes = defaults.PreviewLimitBytes
	}
	if config.UploadFileLimitBytes <= 0 {
		config.UploadFileLimitBytes = defaults.UploadFileLimitBytes
	}
	if config.UploadCountLimit <= 0 {
		config.UploadCountLimit = defaults.UploadCountLimit
	}
	if config.SkillListLimit <= 0 {
		config.SkillListLimit = defaults.SkillListLimit
	}
	if config.SkillManifestLimitBytes <= 0 {
		config.SkillManifestLimitBytes = defaults.SkillManifestLimitBytes
	}
	workspaceRoot, err := normalizeContainerRoot(config.WorkspaceRootPath)
	if err != nil {
		config.WorkspaceRootPath = defaults.WorkspaceRootPath
	} else {
		config.WorkspaceRootPath = workspaceRoot
	}
	sharedRoot, err := normalizeContainerRoot(config.SharedRootPath)
	if err != nil {
		config.SharedRootPath = defaults.SharedRootPath
	} else {
		config.SharedRootPath = sharedRoot
	}
	if config.WorkspaceRootPath == config.SharedRootPath {
		config.SharedRootPath = defaults.SharedRootPath
	}
	return config
}

func normalizeContainerRoot(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !strings.HasPrefix(trimmed, "/") {
		return "", gerror.New("agentbox workspace root must be an absolute container path")
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "/" || cleaned == "." {
		return "", gerror.New("agentbox workspace root cannot be the container root")
	}
	return cleaned, nil
}
