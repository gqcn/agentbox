// This file implements AgentBox workspace Git entry-point validation and
// read-only runtime projections. It verifies Agent ownership and
// repository-relative file inputs before delegating to the runtime backend.

package workspace

import (
	"context"
	"path"
	"sort"
	"strings"

	"lina-core/pkg/bizerr"
)

// GitStatus validates repository path visibility before returning runtime-backed Git status.
func (s *serviceImpl) GitStatus(ctx context.Context, userID string, agentID string, inputPath string) (*GitStatusResponse, error) {
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	status, err := s.runtimeBackend.WorkspaceGitStatus(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return gitStatusFromRuntime(workspacePath, status), nil
}

// GitFile validates repository and file paths before returning a runtime-backed Git file preview.
func (s *serviceImpl) GitFile(ctx context.Context, userID string, agentID string, inputPath string, file string) (*GitFileResponse, error) {
	if err := validateGitFilePath(file); err != nil {
		return nil, err
	}
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	item, err := s.runtimeBackend.WorkspaceGitFile(ctx, userID, agentID, workspacePath, cleanGitFilePath(file))
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return s.gitFileFromRuntime(agentID, item), nil
}

// GitDiff validates repository and file paths before returning runtime-backed Git diff content.
func (s *serviceImpl) GitDiff(ctx context.Context, userID string, agentID string, inputPath string, file string, scope string) (*GitDiffResponse, error) {
	if err := validateGitFilePath(file); err != nil {
		return nil, err
	}
	scope, err := normalizeGitDiffScope(scope)
	if err != nil {
		return nil, err
	}
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	item, err := s.runtimeBackend.WorkspaceGitDiff(ctx, userID, agentID, workspacePath, cleanGitFilePath(file), scope)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return gitDiffFromRuntime(cleanGitFilePath(file), scope, item), nil
}

// GitCommitMessageSuggestion validates repository path visibility before generating a runtime-backed suggestion.
func (s *serviceImpl) GitCommitMessageSuggestion(ctx context.Context, userID string, agentID string, inputPath string) (*GitCommitMessageSuggestionResponse, error) {
	if err := s.ensureWorkspacePathVisible(ctx, userID, agentID, inputPath); err != nil {
		return nil, err
	}
	return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
}

// GitIndexStage validates repository path visibility before staging files.
func (s *serviceImpl) GitIndexStage(ctx context.Context, userID string, agentID string, input GitIndexInput) (*GitStatusResponse, error) {
	if err := validateGitFileSelection(input.Files, input.All); err != nil {
		return nil, err
	}
	workspacePath, err := s.normalizeWorkspacePath(input.Path)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	status, err := s.runtimeBackend.WorkspaceGitStage(ctx, userID, agentID, RuntimeGitIndexInput{
		Path:  workspacePath,
		Files: cleanGitFilePaths(input.Files),
		All:   input.All,
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return gitStatusFromRuntime(workspacePath, status), nil
}

// GitIndexUnstage validates repository path visibility before unstaging files.
func (s *serviceImpl) GitIndexUnstage(ctx context.Context, userID string, agentID string, input GitIndexInput) (*GitStatusResponse, error) {
	if err := validateGitFileSelection(input.Files, input.All); err != nil {
		return nil, err
	}
	workspacePath, err := s.normalizeWorkspacePath(input.Path)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	status, err := s.runtimeBackend.WorkspaceGitUnstage(ctx, userID, agentID, RuntimeGitIndexInput{
		Path:  workspacePath,
		Files: cleanGitFilePaths(input.Files),
		All:   input.All,
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return gitStatusFromRuntime(workspacePath, status), nil
}

// GitChangesDiscard validates repository path visibility before discarding worktree changes.
func (s *serviceImpl) GitChangesDiscard(ctx context.Context, userID string, agentID string, input GitDiscardInput) (*GitStatusResponse, error) {
	if err := validateGitFileSelection(input.Files, false); err != nil {
		return nil, err
	}
	workspacePath, err := s.normalizeWorkspacePath(input.Path)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	status, err := s.runtimeBackend.WorkspaceGitDiscard(ctx, userID, agentID, RuntimeGitDiscardInput{
		Path:  workspacePath,
		Files: cleanGitFilePaths(input.Files),
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return gitStatusFromRuntime(workspacePath, status), nil
}

// GitCommit validates repository path visibility before creating a commit.
func (s *serviceImpl) GitCommit(ctx context.Context, userID string, agentID string, input GitCommitInput) (*GitCommitResponse, error) {
	commitMessage := strings.TrimSpace(input.Message)
	if commitMessage == "" {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	workspacePath, err := s.normalizeWorkspacePath(input.Path)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	commit, err := s.runtimeBackend.WorkspaceGitCommit(ctx, userID, agentID, RuntimeGitCommitInput{
		Path:    workspacePath,
		Message: commitMessage,
		Push:    input.Push,
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if commit == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	return &GitCommitResponse{
		CommitHash: commit.CommitHash,
		Pushed:     commit.Pushed,
		Status:     *gitStatusFromRuntime(workspacePath, &commit.Status),
	}, nil
}

func normalizeGitDiffScope(scope string) (string, error) {
	value := strings.TrimSpace(scope)
	if value == "" {
		return GitDiffScopeUnstaged, nil
	}
	if value == GitDiffScopeStaged || value == GitDiffScopeUnstaged {
		return value, nil
	}
	return "", bizerr.NewCode(CodeWorkspaceInvalidInput)
}

func validateGitFileSelection(files []string, all bool) error {
	if all {
		return nil
	}
	if len(files) == 0 {
		return bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	for _, file := range files {
		if err := validateGitFilePath(file); err != nil {
			return err
		}
	}
	return nil
}

func validateGitFilePath(file string) error {
	value := strings.TrimSpace(file)
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") {
		return bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	return nil
}

func cleanGitFilePath(file string) string {
	return path.Clean(strings.TrimSpace(file))
}

func cleanGitFilePaths(files []string) []string {
	out := make([]string, 0, len(files))
	for _, file := range files {
		out = append(out, cleanGitFilePath(file))
	}
	return out
}

func gitStatusFromRuntime(workspacePath string, status *RuntimeGitStatus) *GitStatusResponse {
	response := &GitStatusResponse{
		State: GitRepositoryStateNotRepo,
		Path:  workspacePath,
	}
	if status == nil {
		response.Message = "Selected path is not a Git repository."
		return response
	}
	response.Path = firstNonEmpty(status.Path, workspacePath)
	if status.NotRepository {
		response.Message = "Selected path is not a Git repository."
		return response
	}
	response.Root = status.RepositoryRoot
	changes := parseGitStatusPorcelain(status.Porcelain)
	unstagedChanges, stagedChanges := splitGitChanges(changes)
	response.Changes = unstagedChanges
	response.StagedChanges = stagedChanges
	response.ChangeTree = buildGitTree(unstagedChanges)
	response.StagedTree = buildGitTree(stagedChanges)
	response.State = GitRepositoryStateOK
	if len(changes) == 0 {
		response.State = GitRepositoryStateClean
		response.Message = "Repository is clean."
	}
	return response
}

func (s *serviceImpl) gitFileFromRuntime(agentID string, item *RuntimeGitFile) *GitFileResponse {
	response := &GitFileResponse{}
	if item == nil || item.NotRepository {
		response.Message = "Selected path is not a Git repository."
		return response
	}
	response.Status = item.Status
	response.Message = item.Message
	if item.Deleted {
		response.Status = GitChangeTypeDeleted
		if response.Message == "" {
			response.Message = "File is deleted from the workspace."
		}
		return response
	}
	fileInfo := fileInfoFromRuntime(item.File)
	fileInfo.ContentType = detectContentTypeBytes(item.File.Path, item.Content)
	preview := FilePreview{
		File:        fileInfo,
		PreviewType: WorkspacePreviewUnsupported,
		DownloadURL: workspaceDownloadURL(agentID, item.File.Path),
	}
	if isImageContentType(fileInfo.ContentType) {
		preview.PreviewType = WorkspacePreviewImage
	} else if item.File.Size > s.config.PreviewLimitBytes {
		preview.PreviewType = WorkspacePreviewText
		preview.TooLarge = true
	} else if isLikelyText(item.Content, preview.File.ContentType) {
		preview.PreviewType = WorkspacePreviewText
		preview.Content = string(item.Content)
		preview.Encoding = "utf-8"
		preview.ContentHash = workspaceContentHash(item.Content)
	}
	response.File = preview
	return response
}

func gitDiffFromRuntime(file string, scope string, item *RuntimeGitDiff) *GitDiffResponse {
	response := &GitDiffResponse{
		Path:         file,
		Scope:        scope,
		OriginalPath: file,
		ModifiedPath: file,
		Language:     gitDiffLanguage(file),
	}
	if item == nil || item.NotRepository {
		response.Message = "Selected path is not a Git repository."
		return response
	}
	response.Path = firstNonEmpty(item.Path, file)
	response.Status = item.Status
	response.Scope = firstNonEmpty(item.Scope, scope)
	response.Diff = item.Diff
	response.OriginalContent = item.OriginalContent
	response.ModifiedContent = item.ModifiedContent
	response.OriginalPath = firstNonEmpty(item.OriginalPath, response.Path)
	response.ModifiedPath = firstNonEmpty(item.ModifiedPath, response.Path)
	response.Language = firstNonEmpty(item.Language, gitDiffLanguage(response.Path))
	response.Message = item.Message
	return response
}

func parseGitStatusPorcelain(output string) []GitChange {
	if output == "" {
		return nil
	}
	parts := strings.Split(output, "\x00")
	changes := make([]GitChange, 0, len(parts))
	for index := 0; index < len(parts); index++ {
		part := parts[index]
		if part == "" || len(part) < 3 {
			continue
		}
		indexState := string(part[0])
		workState := string(part[1])
		filePath := strings.TrimSpace(part[3:])
		oldPath := ""
		if indexState == "R" || indexState == "C" {
			if index+1 < len(parts) {
				oldPath = parts[index+1]
				index++
			}
		}
		changes = append(changes, GitChange{
			Path:       filePath,
			OldPath:    oldPath,
			Status:     gitChangeType(indexState, workState),
			IndexState: indexState,
			WorkState:  workState,
		})
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	return changes
}

func splitGitChanges(changes []GitChange) ([]GitChange, []GitChange) {
	unstaged := make([]GitChange, 0, len(changes))
	staged := make([]GitChange, 0, len(changes))
	for _, change := range changes {
		if hasGitWorktreeChange(change) {
			item := change
			item.ChangeScope = GitChangeScopeUnstaged
			item.Status = scopedGitChangeStatus(change, GitDiffScopeUnstaged)
			unstaged = append(unstaged, item)
		}
		if hasGitIndexChange(change) {
			item := change
			item.ChangeScope = GitChangeScopeStaged
			item.Status = scopedGitChangeStatus(change, GitDiffScopeStaged)
			staged = append(staged, item)
		}
	}
	return unstaged, staged
}

func hasGitWorktreeChange(change GitChange) bool {
	return change.WorkState != "" && change.WorkState != " "
}

func hasGitIndexChange(change GitChange) bool {
	return change.IndexState != "" && change.IndexState != " " && change.IndexState != "?"
}

func scopedGitChangeStatus(change GitChange, scope string) string {
	if scope == GitDiffScopeStaged {
		return gitStateChangeType(change.IndexState)
	}
	if change.IndexState == "?" && change.WorkState == "?" {
		return GitChangeTypeUntracked
	}
	return gitStateChangeType(change.WorkState)
}

func gitStateChangeType(state string) string {
	switch state {
	case "M":
		return GitChangeTypeModified
	case "A":
		return GitChangeTypeAdded
	case "D":
		return GitChangeTypeDeleted
	case "R":
		return GitChangeTypeRenamed
	case "C":
		return GitChangeTypeCopied
	case "T":
		return GitChangeTypeTypeChanged
	case "U":
		return GitChangeTypeUnmerged
	case "?":
		return GitChangeTypeUntracked
	default:
		return GitChangeTypeUnknown
	}
}

func gitChangeType(indexState string, workState string) string {
	if indexState == "?" && workState == "?" {
		return GitChangeTypeUntracked
	}
	if indexState == "U" || workState == "U" {
		return GitChangeTypeUnmerged
	}
	for _, state := range []string{workState, indexState} {
		switch state {
		case "M":
			return GitChangeTypeModified
		case "A":
			return GitChangeTypeAdded
		case "D":
			return GitChangeTypeDeleted
		case "R":
			return GitChangeTypeRenamed
		case "C":
			return GitChangeTypeCopied
		case "T":
			return GitChangeTypeTypeChanged
		}
	}
	return GitChangeTypeUnknown
}

func buildGitTree(changes []GitChange) []GitTreeNode {
	type mutableNode struct {
		node     GitTreeNode
		children map[string]*mutableNode
	}
	root := &mutableNode{children: map[string]*mutableNode{}}
	for _, change := range changes {
		parts := strings.Split(change.Path, "/")
		current := root
		for index, part := range parts {
			if part == "" {
				continue
			}
			child, ok := current.children[part]
			if !ok {
				nodeType := WorkspaceNodeDirectory
				if index == len(parts)-1 {
					nodeType = WorkspaceNodeFile
				}
				child = &mutableNode{
					node: GitTreeNode{
						Name: part,
						Path: strings.Join(parts[:index+1], "/"),
						Type: nodeType,
					},
					children: map[string]*mutableNode{},
				}
				current.children[part] = child
			}
			if index == len(parts)-1 {
				child.node.Status = change.Status
				child.node.OldPath = change.OldPath
				child.node.ChangeScope = change.ChangeScope
			}
			current = child
		}
	}
	var convert func(node *mutableNode) []GitTreeNode
	convert = func(node *mutableNode) []GitTreeNode {
		keys := make([]string, 0, len(node.children))
		for key := range node.children {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		items := make([]GitTreeNode, 0, len(keys))
		for _, key := range keys {
			child := node.children[key]
			child.node.Children = convert(child)
			items = append(items, child.node)
		}
		return items
	}
	return convert(root)
}

func gitDiffLanguage(pathValue string) string {
	lower := strings.ToLower(path.Base(pathValue))
	ext := strings.ToLower(path.Ext(lower))
	switch {
	case lower == "dockerfile" || lower == "containerfile" || strings.HasSuffix(lower, ".dockerfile"):
		return "dockerfile"
	case lower == "makefile":
		return "shell"
	case lower == "go.mod" || lower == "go.sum" || ext == ".go":
		return "go"
	case ext == ".ts" || ext == ".tsx" || ext == ".mts" || ext == ".cts":
		return "typescript"
	case ext == ".js" || ext == ".jsx" || ext == ".mjs" || ext == ".cjs":
		return "javascript"
	case ext == ".json" || ext == ".jsonc":
		return "json"
	case ext == ".md" || ext == ".markdown" || ext == ".mdx":
		return "markdown"
	case ext == ".css":
		return "css"
	case ext == ".scss":
		return "scss"
	case ext == ".less":
		return "less"
	case ext == ".html" || ext == ".htm":
		return "html"
	case ext == ".xml":
		return "xml"
	case ext == ".yaml" || ext == ".yml":
		return "yaml"
	case ext == ".sql":
		return "sql"
	case ext == ".sh" || ext == ".bash" || ext == ".zsh":
		return "shell"
	case ext == ".py":
		return "python"
	case ext == ".rs":
		return "rust"
	case ext == ".java":
		return "java"
	case ext == ".c" || ext == ".h":
		return "c"
	case ext == ".cc" || ext == ".cpp" || ext == ".cxx" || ext == ".hpp" || ext == ".hh" || ext == ".hxx":
		return "cpp"
	case ext == ".cs":
		return "csharp"
	case ext == ".php":
		return "php"
	case ext == ".rb":
		return "ruby"
	case ext == ".swift":
		return "swift"
	case ext == ".kt" || ext == ".kts":
		return "kotlin"
	case ext == ".lua":
		return "lua"
	case ext == ".graphql" || ext == ".gql":
		return "graphql"
	case ext == ".tf" || ext == ".tfvars" || ext == ".hcl":
		return "hcl"
	case ext == ".proto":
		return "proto"
	case ext == ".r":
		return "r"
	case ext == ".scala" || ext == ".sbt":
		return "scala"
	case ext == ".toml" || ext == ".ini" || ext == ".properties" || lower == ".env":
		return "ini"
	default:
		return "plaintext"
	}
}
