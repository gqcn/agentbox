// This file implements read-only Agent workspace runtime helpers on top of the
// plugin-labelled Docker runtime. Every operation first resolves a container
// that belongs to the current AgentBox user and Agent, then constrains paths to
// the managed workspace or shared roots inside that container.

package container

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/moby/moby/api/pkg/stdcopy"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"

	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// WorkspacePathStat resolves and stats one workspace path for a visible Agent.
func (b *dockerRuntimeBackend) WorkspacePathStat(ctx context.Context, userID string, agentID string, workspacePath string) (*workspacesvc.RuntimePathEntry, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	return b.agentPathStat(ctx, userID, agentID, resolvedPath)
}

// WorkspaceDirectoryEntries returns immediate entries under one workspace directory.
func (b *dockerRuntimeBackend) WorkspaceDirectoryEntries(ctx context.Context, userID string, agentID string, workspacePath string, includeFiles bool) ([]workspacesvc.RuntimePathEntry, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
dir=$1
if [ ! -d "$dir" ]; then
  echo "path is not a directory" >&2
  exit 2
fi
find "$dir" -mindepth 1 -maxdepth 1 -printf '%f\t%p\t%y\t%s\t%T@\n'`,
		"agentbox-list-dir",
		resolvedPath,
	})
	if err != nil {
		return nil, err
	}
	entries, err := parseAgentPathEntries(output)
	if err != nil {
		return nil, err
	}
	items := make([]workspacesvc.RuntimePathEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != workspacesvc.WorkspaceNodeDirectory && !includeFiles {
			continue
		}
		items = append(items, entry)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == workspacesvc.WorkspaceNodeDirectory
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

// WorkspaceReadFile reads one non-directory workspace file for a visible Agent.
func (b *dockerRuntimeBackend) WorkspaceReadFile(ctx context.Context, userID string, agentID string, workspacePath string) ([]byte, *workspacesvc.RuntimePathEntry, error) {
	runtimeFile, err := b.WorkspaceOpenFile(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, nil, err
	}
	defer runtimeFile.Reader.Close()
	data, err := io.ReadAll(runtimeFile.Reader)
	return data, &runtimeFile.Entry, err
}

// WorkspaceOpenFile opens one non-directory workspace file stream for a visible Agent.
func (b *dockerRuntimeBackend) WorkspaceOpenFile(ctx context.Context, userID string, agentID string, workspacePath string) (*workspacesvc.RuntimeFile, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	stat, err := b.agentPathStat(ctx, userID, agentID, resolvedPath)
	if err != nil {
		return nil, err
	}
	if stat.Type == workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("workspace path is a directory")
	}
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	reader, err := cli.CopyFromContainer(ctx, inspected.ID, dockerclient.CopyFromContainerOptions{SourcePath: resolvedPath})
	if err != nil {
		return nil, dockerActionError(err, "read agentbox agent workspace file")
	}
	fileReader := newSingleFileTarReadCloser(reader.Content)
	return &workspacesvc.RuntimeFile{Entry: *stat, Reader: fileReader}, nil
}

// WorkspaceWriteFile writes one non-directory workspace file for a visible Agent.
func (b *dockerRuntimeBackend) WorkspaceWriteFile(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeWriteFileInput) (*workspacesvc.RuntimePathEntry, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	stat, err := b.agentPathStat(ctx, userID, agentID, resolvedPath)
	if err != nil {
		return nil, err
	}
	if stat.Type == workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("workspace path is a directory")
	}
	parentPath := path.Dir(resolvedPath)
	filename := path.Base(resolvedPath)
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	archive, err := fileContentTar(filename, input.Content)
	if err != nil {
		return nil, err
	}
	if _, err := cli.CopyToContainer(ctx, inspected.ID, dockerclient.CopyToContainerOptions{
		DestinationPath: parentPath,
		Content:         archive,
	}); err != nil {
		return nil, dockerActionError(err, "write agentbox agent workspace file")
	}
	return b.agentPathStat(ctx, userID, agentID, resolvedPath)
}

// WorkspaceUploadFile creates or overwrites one file under a visible workspace directory.
func (b *dockerRuntimeBackend) WorkspaceUploadFile(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeUploadFileInput) (*workspacesvc.RuntimePathEntry, error) {
	directoryPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, input.DirectoryPath)
	if err != nil {
		return nil, err
	}
	directoryStat, err := b.agentPathStat(ctx, userID, agentID, directoryPath)
	if err != nil {
		return nil, err
	}
	if directoryStat.Type != workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("workspace upload target is not a directory")
	}
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	archive, err := fileContentTar(input.Name, input.Content)
	if err != nil {
		return nil, err
	}
	if _, err := cli.CopyToContainer(ctx, inspected.ID, dockerclient.CopyToContainerOptions{
		DestinationPath: directoryPath,
		Content:         archive,
	}); err != nil {
		return nil, dockerActionError(err, "upload agentbox agent workspace file")
	}
	return b.agentPathStat(ctx, userID, agentID, path.Join(directoryPath, input.Name))
}

// WorkspaceCreateEntry creates one file or directory under a visible workspace directory.
func (b *dockerRuntimeBackend) WorkspaceCreateEntry(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeCreateEntryInput) (*workspacesvc.RuntimePathEntry, error) {
	parentPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, input.ParentPath)
	if err != nil {
		return nil, err
	}
	parentStat, err := b.agentPathStat(ctx, userID, agentID, parentPath)
	if err != nil {
		return nil, err
	}
	if parentStat.Type != workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("workspace parent path is not a directory")
	}
	targetPath := path.Join(parentPath, input.Name)
	if input.Type == workspacesvc.WorkspaceNodeDirectory {
		if _, err := b.runAgentCommand(ctx, userID, agentID, parentPath, []string{
			"/bin/sh",
			"-c",
			`set -eu
target=$1
if [ -e "$target" ]; then
  echo "workspace entry already exists" >&2
  exit 3
fi
mkdir -- "$target"`,
			"agentbox-mkdir",
			targetPath,
		}); err != nil {
			return nil, err
		}
		return b.agentPathStat(ctx, userID, agentID, targetPath)
	}
	if input.Type != workspacesvc.WorkspaceNodeFile {
		return nil, gerror.New("workspace entry type is invalid")
	}
	if _, err := b.runAgentCommand(ctx, userID, agentID, parentPath, []string{
		"/bin/sh",
		"-c",
		`set -eu
target=$1
if [ -e "$target" ]; then
  echo "workspace entry already exists" >&2
  exit 3
fi`,
		"agentbox-check-create-file",
		targetPath,
	}); err != nil {
		return nil, err
	}
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	archive, err := fileContentTar(input.Name, nil)
	if err != nil {
		return nil, err
	}
	if _, err := cli.CopyToContainer(ctx, inspected.ID, dockerclient.CopyToContainerOptions{
		DestinationPath: parentPath,
		Content:         archive,
	}); err != nil {
		return nil, dockerActionError(err, "create agentbox agent workspace file")
	}
	return b.agentPathStat(ctx, userID, agentID, targetPath)
}

// WorkspaceGitStatus reads Git status for a visible workspace path without mutating the repository.
func (b *dockerRuntimeBackend) WorkspaceGitStatus(ctx context.Context, userID string, agentID string, workspacePath string) (*workspacesvc.RuntimeGitStatus, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, resolvedPath, []string{
		"/bin/sh",
		"-c",
		`set -eu
target=$1
repo_root=$(git -C "$target" rev-parse --show-toplevel 2>/dev/null || true)
if [ -z "$repo_root" ]; then
  printf 'not_repo\n%s\n' "$target"
  exit 0
fi
workspace_root=$2
shared_root=$3
case "$repo_root" in
  "$workspace_root"|"$workspace_root"/*|"$shared_root"|"$shared_root"/*) ;;
  *) echo "git repository is outside workspace" >&2; exit 2 ;;
esac
printf 'repo\n%s\n' "$repo_root"
git -C "$repo_root" status --porcelain=v1 -z`,
		"agentbox-git-status",
		resolvedPath,
		b.workspaceRootPath(),
		b.sharedRootPath(),
	})
	if err != nil {
		return nil, err
	}
	return parseAgentGitStatus(output, resolvedPath)
}

// WorkspaceGitFile reads one repository-relative file preview without mutating the repository.
func (b *dockerRuntimeBackend) WorkspaceGitFile(ctx context.Context, userID string, agentID string, workspacePath string, file string) (*workspacesvc.RuntimeGitFile, error) {
	repoRoot, notRepo, err := b.resolveAgentGitRepository(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	if notRepo {
		return &workspacesvc.RuntimeGitFile{NotRepository: true}, nil
	}
	targetPath, err := agentGitFileTarget(repoRoot, file)
	if err != nil {
		return nil, err
	}
	targetPath, err = b.resolveAgentWorkspacePath(ctx, userID, agentID, targetPath)
	if err != nil {
		if isAgentWorkspaceMissingPathError(err) {
			return &workspacesvc.RuntimeGitFile{
				Status:  workspacesvc.GitChangeTypeDeleted,
				Message: "File is deleted from the workspace.",
				Deleted: true,
			}, nil
		}
		return nil, err
	}
	stat, err := b.agentPathStat(ctx, userID, agentID, targetPath)
	if err != nil {
		if isAgentWorkspaceMissingPathError(err) {
			return &workspacesvc.RuntimeGitFile{
				Status:  workspacesvc.GitChangeTypeDeleted,
				Message: "File is deleted from the workspace.",
				Deleted: true,
			}, nil
		}
		return nil, err
	}
	if stat.Type == workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("git file path is a directory")
	}
	status := b.agentGitFileStatus(ctx, userID, agentID, repoRoot, file, workspacesvc.GitDiffScopeUnstaged)
	if stat.Size > b.config.Workspace.PreviewLimitBytes {
		return &workspacesvc.RuntimeGitFile{
			File:   *stat,
			Status: status,
		}, nil
	}
	data, refreshedStat, err := b.WorkspaceReadFile(ctx, userID, agentID, targetPath)
	if err != nil {
		return nil, err
	}
	if refreshedStat != nil {
		stat = refreshedStat
	}
	return &workspacesvc.RuntimeGitFile{
		File:    *stat,
		Content: data,
		Status:  status,
	}, nil
}

// WorkspaceGitDiff reads one repository-relative file diff without mutating the repository.
func (b *dockerRuntimeBackend) WorkspaceGitDiff(ctx context.Context, userID string, agentID string, workspacePath string, file string, scope string) (*workspacesvc.RuntimeGitDiff, error) {
	repoRoot, notRepo, err := b.resolveAgentGitRepository(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	if notRepo {
		return &workspacesvc.RuntimeGitDiff{Path: file, Scope: scope, NotRepository: true}, nil
	}
	targetPath, err := agentGitFileTarget(repoRoot, file)
	if err != nil {
		return nil, err
	}
	targetPath, err = b.resolveAgentWorkspacePath(ctx, userID, agentID, targetPath)
	if err != nil && !isAgentWorkspaceMissingPathError(err) {
		return nil, err
	}
	relFile := strings.TrimSpace(file)
	status := b.agentGitFileStatus(ctx, userID, agentID, repoRoot, relFile, scope)
	diff, err := b.agentGitDiff(ctx, userID, agentID, repoRoot, relFile, status, scope)
	if err != nil {
		return nil, err
	}
	response := &workspacesvc.RuntimeGitDiff{
		Path:         relFile,
		Status:       status,
		Scope:        scope,
		Diff:         diff,
		OriginalPath: relFile,
		ModifiedPath: relFile,
	}
	original, err := b.agentGitOriginalContent(ctx, userID, agentID, repoRoot, relFile, status)
	if err != nil {
		response.Message = err.Error()
	} else {
		response.OriginalContent = original
	}
	modified, err := b.agentGitModifiedContent(ctx, userID, agentID, targetPath, repoRoot, relFile, status, scope)
	if err != nil && response.Message == "" {
		response.Message = err.Error()
	} else if err == nil {
		response.ModifiedContent = modified
	}
	return response, nil
}

// WorkspaceGitStage stages selected or all repository changes for a visible Agent runtime.
func (b *dockerRuntimeBackend) WorkspaceGitStage(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeGitIndexInput) (*workspacesvc.RuntimeGitStatus, error) {
	repoRoot, err := b.resolveAgentGitRepositoryRequired(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	if input.All {
		if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "add", "-A", "--", "."); err != nil {
			return nil, err
		}
		return b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
	}
	if err := b.ensureAgentGitFiles(ctx, userID, agentID, repoRoot, input.Files); err != nil {
		return nil, err
	}
	args := append([]string{"add", "--"}, input.Files...)
	if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, args...); err != nil {
		return nil, err
	}
	return b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
}

// WorkspaceGitUnstage removes selected or all repository changes from the Git index.
func (b *dockerRuntimeBackend) WorkspaceGitUnstage(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeGitIndexInput) (*workspacesvc.RuntimeGitStatus, error) {
	repoRoot, err := b.resolveAgentGitRepositoryRequired(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	if input.All {
		if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "restore", "--staged", "--", "."); err != nil {
			return nil, err
		}
		return b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
	}
	if err := b.ensureAgentGitFiles(ctx, userID, agentID, repoRoot, input.Files); err != nil {
		return nil, err
	}
	args := append([]string{"restore", "--staged", "--"}, input.Files...)
	if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, args...); err != nil {
		return nil, err
	}
	return b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
}

// WorkspaceGitDiscard discards selected unstaged worktree changes for a visible Agent runtime.
func (b *dockerRuntimeBackend) WorkspaceGitDiscard(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeGitDiscardInput) (*workspacesvc.RuntimeGitStatus, error) {
	repoRoot, err := b.resolveAgentGitRepositoryRequired(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	if err := b.ensureAgentGitFiles(ctx, userID, agentID, repoRoot, input.Files); err != nil {
		return nil, err
	}
	for _, relFile := range input.Files {
		status := b.agentGitFileStatus(ctx, userID, agentID, repoRoot, relFile, workspacesvc.GitDiffScopeUnstaged)
		if status == workspacesvc.GitChangeTypeUntracked {
			if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "clean", "-fd", "--", relFile); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "restore", "--worktree", "--", relFile); err != nil {
			return nil, err
		}
	}
	return b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
}

// WorkspaceGitCommit creates a Git commit from staged changes for a visible Agent runtime.
func (b *dockerRuntimeBackend) WorkspaceGitCommit(ctx context.Context, userID string, agentID string, input workspacesvc.RuntimeGitCommitInput) (*workspacesvc.RuntimeGitCommitResult, error) {
	repoRoot, err := b.resolveAgentGitRepositoryRequired(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return nil, gerror.New("git commit message is required")
	}
	staged, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(staged) == "" {
		return nil, gerror.New("no staged changes to commit")
	}
	if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "commit", "-m", message); err != nil {
		return nil, err
	}
	hash, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	if input.Push {
		if _, err := b.runAgentGit(ctx, userID, agentID, repoRoot, "push"); err != nil {
			return nil, err
		}
	}
	status, err := b.WorkspaceGitStatus(ctx, userID, agentID, input.Path)
	if err != nil {
		return nil, err
	}
	return &workspacesvc.RuntimeGitCommitResult{
		CommitHash: strings.TrimSpace(hash),
		Pushed:     input.Push,
		Status:     *status,
	}, nil
}

// WorkspaceSkills lists bounded skill metadata for one visible Agent runtime.
func (b *dockerRuntimeBackend) WorkspaceSkills(ctx context.Context, userID string, agentID string, scope string, workspacePath string) ([]workspacesvc.SkillInfo, error) {
	if scope == workspacesvc.SkillScopeGlobal {
		return b.agentGlobalSkills(ctx, userID, agentID)
	}
	projectPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, err
	}
	stat, err := b.agentPathStat(ctx, userID, agentID, projectPath)
	if err != nil {
		return nil, err
	}
	if stat.Type != workspacesvc.WorkspaceNodeDirectory {
		return nil, gerror.New("project path is not a directory")
	}
	return b.agentSkillDir(ctx, userID, agentID, path.Join(projectPath, ".agents", "skills"), workspacesvc.SkillScopeProject, projectPath)
}

func (b *dockerRuntimeBackend) resolveAgentWorkspacePath(ctx context.Context, userID string, agentID string, workspacePath string) (string, error) {
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
workspace_root=$1
shared_root=$2
target=$3
case "$target" in
  "$workspace_root"|"$workspace_root"/*|"$shared_root"|"$shared_root"/*) ;;
  *) echo "path is outside workspace" >&2; exit 2 ;;
esac
resolved=$(readlink -f -- "$target") || { echo "workspace path does not exist" >&2; exit 2; }
case "$resolved" in
  "$workspace_root"|"$workspace_root"/*|"$shared_root"|"$shared_root"/*) printf '%s\n' "$resolved" ;;
  *) echo "path is outside workspace" >&2; exit 2 ;;
esac`,
		"agentbox-resolve-workspace",
		b.workspaceRootPath(),
		b.sharedRootPath(),
		workspacePath,
	})
	if err != nil {
		return "", err
	}
	resolved := strings.TrimSpace(output)
	if resolved == "" {
		return "", gerror.New("workspace path does not exist")
	}
	return resolved, nil
}

func (b *dockerRuntimeBackend) agentPathStat(ctx context.Context, userID string, agentID string, containerPath string) (*workspacesvc.RuntimePathEntry, error) {
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
target=$1
name=$(basename -- "$target")
mode=$(stat -c '%f' -- "$target")
size=$(stat -c '%s' -- "$target")
mtime=$(stat -c '%Y' -- "$target")
link_target=""
if [ -L "$target" ]; then
  link_target=$(readlink -- "$target" || true)
fi
printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$name" "$target" "$mode" "$size" "$mtime" "$link_target"`,
		"agentbox-stat-path",
		containerPath,
	})
	if err != nil {
		return nil, err
	}
	return parseAgentPathStat(output, containerPath)
}

func (b *dockerRuntimeBackend) runAgentCommand(ctx context.Context, userID string, agentID string, workingDir string, command []string) (string, error) {
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return "", err
	}
	cli, err := b.requireClient()
	if err != nil {
		return "", err
	}
	resp, err := cli.ExecCreate(ctx, inspected.ID, dockerclient.ExecCreateOptions{
		TTY:          false,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   workingDir,
		Cmd:          command,
	})
	if err != nil {
		return "", dockerActionError(err, "create agentbox agent workspace exec")
	}
	attach, err := cli.ExecAttach(ctx, resp.ID, dockerclient.ExecAttachOptions{TTY: false})
	if err != nil {
		return "", dockerActionError(err, "attach agentbox agent workspace exec")
	}
	defer attach.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, attach.Reader); err != nil {
		return "", gerror.Wrap(err, "read agentbox agent workspace exec output")
	}
	inspectedExec, err := cli.ExecInspect(ctx, resp.ID, dockerclient.ExecInspectOptions{})
	if err != nil {
		return "", dockerActionError(err, "inspect agentbox agent workspace exec")
	}
	if inspectedExec.ExitCode != 0 {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = "agentbox agent workspace exec failed"
		}
		return stdout.String(), gerror.New(message)
	}
	return stdout.String(), nil
}

func (b *dockerRuntimeBackend) inspectRunningAgentContainer(ctx context.Context, userID string, agentID string) (dockercontainer.InspectResponse, error) {
	cli, err := b.requireClient()
	if err != nil {
		return dockercontainer.InspectResponse{}, err
	}
	inspected, err := b.inspectVisibleAgentContainer(ctx, cli, userID, agentID, "")
	if err != nil {
		return dockercontainer.InspectResponse{}, err
	}
	if inspected.State == nil || inspected.State.Status != dockercontainer.StateRunning {
		return dockercontainer.InspectResponse{}, gerror.New("agent container is not running")
	}
	return inspected, nil
}

func (b *dockerRuntimeBackend) resolveAgentGitRepository(ctx context.Context, userID string, agentID string, workspacePath string) (string, bool, error) {
	resolvedPath, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, workspacePath)
	if err != nil {
		return "", false, err
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, resolvedPath, []string{
		"/bin/sh",
		"-c",
		`set -eu
target=$1
repo_root=$(git -C "$target" rev-parse --show-toplevel 2>/dev/null || true)
if [ -z "$repo_root" ]; then
  printf 'not_repo\n%s\n' "$target"
  exit 0
fi
workspace_root=$2
shared_root=$3
case "$repo_root" in
  "$workspace_root"|"$workspace_root"/*|"$shared_root"|"$shared_root"/*) ;;
  *) echo "git repository is outside workspace" >&2; exit 2 ;;
esac
printf 'repo\n%s\n' "$repo_root"`,
		"agentbox-git-root",
		resolvedPath,
		b.workspaceRootPath(),
		b.sharedRootPath(),
	})
	if err != nil {
		return "", false, err
	}
	return parseAgentGitRepositoryRoot(output, resolvedPath)
}

func (b *dockerRuntimeBackend) resolveAgentGitRepositoryRequired(ctx context.Context, userID string, agentID string, workspacePath string) (string, error) {
	repoRoot, notRepo, err := b.resolveAgentGitRepository(ctx, userID, agentID, workspacePath)
	if err != nil {
		return "", err
	}
	if notRepo {
		return "", gerror.New("selected path is not a git repository")
	}
	return repoRoot, nil
}

func (b *dockerRuntimeBackend) ensureAgentGitFiles(ctx context.Context, userID string, agentID string, repoRoot string, files []string) error {
	for _, file := range files {
		targetPath, err := agentGitFileTarget(repoRoot, file)
		if err != nil {
			return err
		}
		if _, err := b.resolveAgentWorkspacePath(ctx, userID, agentID, targetPath); err != nil && !isAgentWorkspaceMissingPathError(err) {
			return err
		}
	}
	return nil
}

func (b *dockerRuntimeBackend) agentGitFileStatus(ctx context.Context, userID string, agentID string, repoRoot string, relFile string, scope string) string {
	output, err := b.runAgentCommand(ctx, userID, agentID, repoRoot, []string{"git", "-C", repoRoot, "status", "--porcelain=v1", "-z", "--", relFile})
	if err != nil {
		return workspacesvc.GitChangeTypeUnknown
	}
	return agentGitFileStatusFromPorcelain(output, scope)
}

func (b *dockerRuntimeBackend) runAgentGit(ctx context.Context, userID string, agentID string, repoRoot string, args ...string) (string, error) {
	command := append([]string{"git", "-C", repoRoot}, args...)
	return b.runAgentCommand(ctx, userID, agentID, repoRoot, command)
}

func (b *dockerRuntimeBackend) agentGitDiff(ctx context.Context, userID string, agentID string, repoRoot string, relFile string, status string, scope string) (string, error) {
	var command []string
	if status == workspacesvc.GitChangeTypeUntracked && scope == workspacesvc.GitDiffScopeUnstaged {
		command = []string{"git", "-C", repoRoot, "diff", "--no-index", "--", "/dev/null", relFile}
	} else if scope == workspacesvc.GitDiffScopeStaged {
		command = []string{"git", "-C", repoRoot, "diff", "--cached", "--", relFile}
	} else {
		command = []string{"git", "-C", repoRoot, "diff", "--", relFile}
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, repoRoot, command)
	if err != nil && strings.TrimSpace(output) == "" {
		return "", err
	}
	return output, nil
}

func (b *dockerRuntimeBackend) agentGitOriginalContent(ctx context.Context, userID string, agentID string, repoRoot string, relFile string, status string) (string, error) {
	if relFile == "" || status == workspacesvc.GitChangeTypeUntracked || status == workspacesvc.GitChangeTypeAdded {
		return "", nil
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, repoRoot, []string{"git", "-C", repoRoot, "show", "HEAD:" + relFile})
	if err != nil {
		return "", err
	}
	return b.editableAgentGitText(output)
}

func (b *dockerRuntimeBackend) agentGitModifiedContent(ctx context.Context, userID string, agentID string, targetPath string, repoRoot string, relFile string, status string, scope string) (string, error) {
	if relFile == "" || status == workspacesvc.GitChangeTypeDeleted {
		return "", nil
	}
	if scope == workspacesvc.GitDiffScopeStaged {
		output, err := b.runAgentCommand(ctx, userID, agentID, repoRoot, []string{"git", "-C", repoRoot, "show", ":" + relFile})
		if err != nil {
			return "", err
		}
		return b.editableAgentGitText(output)
	}
	stat, err := b.agentPathStat(ctx, userID, agentID, targetPath)
	if err != nil {
		return "", err
	}
	if stat.Type == workspacesvc.WorkspaceNodeDirectory {
		return "", gerror.New("git diff path is a directory")
	}
	if stat.Size > b.config.Workspace.PreviewLimitBytes {
		return "", gerror.New("git diff content exceeds editable size limit")
	}
	data, stat, err := b.WorkspaceReadFile(ctx, userID, agentID, targetPath)
	if err != nil {
		return "", err
	}
	if stat != nil && stat.Type == workspacesvc.WorkspaceNodeDirectory {
		return "", gerror.New("git diff path is a directory")
	}
	if int64(len(data)) > b.config.Workspace.PreviewLimitBytes {
		return "", gerror.New("git diff content exceeds editable size limit")
	}
	if !utf8.Valid(data) || bytes.Contains(data, []byte{0}) {
		return "", gerror.New("git diff content is not editable text")
	}
	return string(data), nil
}

func (b *dockerRuntimeBackend) agentGlobalSkills(ctx context.Context, userID string, agentID string) ([]workspacesvc.SkillInfo, error) {
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
for dir in "${CODEX_HOME:-}/skills" "$HOME/.codex/skills" "$HOME/.agents/skills"; do
  if [ "$dir" = "/skills" ]; then continue; fi
  if [ -d "$dir" ]; then printf '%s\n' "$dir"; fi
done`,
		"agentbox-global-skills",
	})
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	items := make([]workspacesvc.SkillInfo, 0)
	for _, line := range strings.Split(output, "\n") {
		dir := strings.TrimSpace(line)
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		skills, err := b.agentGlobalSkillDir(ctx, userID, agentID, dir)
		if err != nil {
			return nil, err
		}
		items = append(items, skills...)
		if len(items) >= b.config.Workspace.SkillListLimit {
			return items[:b.config.Workspace.SkillListLimit], nil
		}
	}
	return items, nil
}

func (b *dockerRuntimeBackend) agentGlobalSkillDir(ctx context.Context, userID string, agentID string, skillsRoot string) ([]workspacesvc.SkillInfo, error) {
	entries, err := b.agentDirectoryEntries(ctx, userID, agentID, skillsRoot, false)
	if err != nil {
		return nil, err
	}
	items := make([]workspacesvc.SkillInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != workspacesvc.WorkspaceNodeDirectory {
			continue
		}
		item := workspacesvc.SkillInfo{
			Name:   entry.Name,
			Scope:  workspacesvc.SkillScopeGlobal,
			Path:   entry.Path,
			Source: skillsRoot,
		}
		raw, ok, err := b.agentOptionalTextFile(ctx, userID, agentID, path.Join(entry.Path, workspacesvc.SkillManifestName), b.config.Workspace.SkillManifestLimitBytes)
		if err != nil {
			return nil, err
		}
		if ok {
			item.HasManifest = true
			name, description := workspacesvc.ParseSkillMarkdown(raw)
			if name != "" {
				item.Name = name
			}
			item.Description = description
		}
		items = append(items, item)
		if len(items) >= b.config.Workspace.SkillListLimit {
			break
		}
	}
	return items, nil
}

func (b *dockerRuntimeBackend) agentSkillDir(ctx context.Context, userID string, agentID string, skillsRoot string, scope string, source string) ([]workspacesvc.SkillInfo, error) {
	entries, err := b.WorkspaceDirectoryEntries(ctx, userID, agentID, skillsRoot, false)
	if err != nil {
		if isAgentWorkspaceMissingPathError(err) || strings.Contains(strings.ToLower(err.Error()), "path is not a directory") {
			return []workspacesvc.SkillInfo{}, nil
		}
		return nil, err
	}
	items := make([]workspacesvc.SkillInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != workspacesvc.WorkspaceNodeDirectory {
			continue
		}
		item := workspacesvc.SkillInfo{
			Name:   entry.Name,
			Scope:  scope,
			Path:   entry.Path,
			Source: source,
		}
		raw, ok, err := b.agentOptionalTextFile(ctx, userID, agentID, path.Join(entry.Path, workspacesvc.SkillManifestName), b.config.Workspace.SkillManifestLimitBytes)
		if err != nil {
			return nil, err
		}
		if ok {
			item.HasManifest = true
			name, description := workspacesvc.ParseSkillMarkdown(raw)
			if name != "" {
				item.Name = name
			}
			item.Description = description
		}
		items = append(items, item)
		if len(items) >= b.config.Workspace.SkillListLimit {
			break
		}
	}
	return items, nil
}

func (b *dockerRuntimeBackend) agentOptionalTextFile(ctx context.Context, userID string, agentID string, containerPath string, limit int64) ([]byte, bool, error) {
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
file=$1
limit=$2
if [ ! -f "$file" ]; then
  printf 'missing\n'
  exit 0
fi
size=$(stat -c '%s' -- "$file")
if [ "$limit" -gt 0 ] && [ "$size" -gt "$limit" ]; then
  printf 'missing\n'
  exit 0
fi
printf 'ok\n'
cat -- "$file"`,
		"agentbox-read-optional-skill",
		containerPath,
		strconv.FormatInt(limit, 10),
	})
	if err != nil {
		if isAgentWorkspaceMissingPathError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	firstBreak := strings.IndexByte(output, '\n')
	if firstBreak < 0 {
		return nil, false, gerror.New("skill manifest output is invalid")
	}
	state := strings.TrimSpace(output[:firstBreak])
	if state == "missing" {
		return nil, false, nil
	}
	if state != "ok" {
		return nil, false, gerror.New("skill manifest state is invalid")
	}
	data := []byte(output[firstBreak+1:])
	if !utf8.Valid(data) || bytes.Contains(data, []byte{0}) {
		return nil, false, nil
	}
	return data, true, nil
}

func (b *dockerRuntimeBackend) agentDirectoryEntries(ctx context.Context, userID string, agentID string, directoryPath string, includeFiles bool) ([]workspacesvc.RuntimePathEntry, error) {
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		`set -eu
dir=$1
if [ ! -d "$dir" ]; then
  exit 0
fi
find "$dir" -mindepth 1 -maxdepth 1 -printf '%f\t%p\t%y\t%s\t%T@\n'`,
		"agentbox-list-skill-dir",
		directoryPath,
	})
	if err != nil {
		return nil, err
	}
	entries, err := parseAgentPathEntries(output)
	if err != nil {
		return nil, err
	}
	items := make([]workspacesvc.RuntimePathEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != workspacesvc.WorkspaceNodeDirectory && !includeFiles {
			continue
		}
		items = append(items, entry)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == workspacesvc.WorkspaceNodeDirectory
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items, nil
}

func parseAgentPathStat(output string, fallbackPath string) (*workspacesvc.RuntimePathEntry, error) {
	line := strings.TrimRight(output, "\r\n")
	if line == "" {
		return nil, gerror.New("agent path stat output is empty")
	}
	parts := strings.Split(line, "\t")
	if len(parts) != 6 {
		return nil, gerror.New("agent path stat output is invalid")
	}
	mode, err := strconv.ParseUint(parts[2], 16, 64)
	if err != nil {
		return nil, gerror.Wrap(err, "parse agent path mode")
	}
	size, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, gerror.Wrap(err, "parse agent path size")
	}
	mtimeUnix, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return nil, gerror.Wrap(err, "parse agent path mtime")
	}
	entryType := workspacesvc.WorkspaceNodeFile
	if mode&0170000 == 0040000 {
		entryType = workspacesvc.WorkspaceNodeDirectory
	}
	return &workspacesvc.RuntimePathEntry{
		Name:       defaultString(parts[0], pathBase(fallbackPath)),
		Path:       defaultString(parts[1], fallbackPath),
		Type:       entryType,
		Size:       size,
		ModifiedAt: time.Unix(mtimeUnix, 0),
		LinkTarget: parts[5],
	}, nil
}

func parseAgentPathEntries(output string) ([]workspacesvc.RuntimePathEntry, error) {
	lines := strings.Split(output, "\n")
	items := make([]workspacesvc.RuntimePathEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 5 {
			return nil, gerror.New("agent directory output is invalid")
		}
		size, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return nil, gerror.Wrap(err, "parse agent path size")
		}
		mtimeUnix, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			return nil, gerror.Wrap(err, "parse agent path mtime")
		}
		seconds := int64(mtimeUnix)
		nanos := int64((mtimeUnix - float64(seconds)) * float64(time.Second))
		entryType := workspacesvc.WorkspaceNodeFile
		if parts[2] == "d" {
			entryType = workspacesvc.WorkspaceNodeDirectory
		}
		items = append(items, workspacesvc.RuntimePathEntry{
			Name:       parts[0],
			Path:       parts[1],
			Type:       entryType,
			Size:       size,
			ModifiedAt: time.Unix(seconds, nanos),
		})
	}
	return items, nil
}

func parseAgentGitStatus(output string, fallbackPath string) (*workspacesvc.RuntimeGitStatus, error) {
	firstBreak := strings.IndexByte(output, '\n')
	if firstBreak < 0 {
		return nil, gerror.New("agent git status output is invalid")
	}
	secondBreak := strings.IndexByte(output[firstBreak+1:], '\n')
	if secondBreak < 0 {
		return nil, gerror.New("agent git status output is invalid")
	}
	secondBreak += firstBreak + 1
	state := strings.TrimSpace(output[:firstBreak])
	rootOrPath := strings.TrimSpace(output[firstBreak+1 : secondBreak])
	porcelain := output[secondBreak+1:]
	if state == "not_repo" {
		return &workspacesvc.RuntimeGitStatus{
			Path:          defaultString(rootOrPath, fallbackPath),
			NotRepository: true,
		}, nil
	}
	if state != "repo" {
		return nil, gerror.New("agent git status state is invalid")
	}
	return &workspacesvc.RuntimeGitStatus{
		Path:           fallbackPath,
		RepositoryRoot: defaultString(rootOrPath, fallbackPath),
		Porcelain:      porcelain,
	}, nil
}

func parseAgentGitRepositoryRoot(output string, fallbackPath string) (string, bool, error) {
	firstBreak := strings.IndexByte(output, '\n')
	if firstBreak < 0 {
		return "", false, gerror.New("agent git repository output is invalid")
	}
	secondBreak := strings.IndexByte(output[firstBreak+1:], '\n')
	if secondBreak < 0 {
		return "", false, gerror.New("agent git repository output is invalid")
	}
	secondBreak += firstBreak + 1
	state := strings.TrimSpace(output[:firstBreak])
	rootOrPath := strings.TrimSpace(output[firstBreak+1 : secondBreak])
	if state == "not_repo" {
		return defaultString(rootOrPath, fallbackPath), true, nil
	}
	if state != "repo" {
		return "", false, gerror.New("agent git repository state is invalid")
	}
	return defaultString(rootOrPath, fallbackPath), false, nil
}

func agentGitFileTarget(repoRoot string, file string) (string, error) {
	relFile := path.Clean(strings.TrimSpace(file))
	if relFile == "." || relFile == ".." || strings.HasPrefix(relFile, "../") || strings.HasPrefix(relFile, "/") {
		return "", gerror.New("git file path is invalid")
	}
	targetPath := path.Join(repoRoot, relFile)
	if targetPath != repoRoot && strings.HasPrefix(targetPath, repoRoot+"/") {
		return targetPath, nil
	}
	return "", gerror.New("git file path is invalid")
}

func agentGitFileStatusFromPorcelain(output string, scope string) string {
	parts := strings.Split(output, "\x00")
	for _, part := range parts {
		if part == "" || len(part) < 3 {
			continue
		}
		indexState := string(part[0])
		workState := string(part[1])
		if indexState == "?" && workState == "?" {
			return workspacesvc.GitChangeTypeUntracked
		}
		if scope == workspacesvc.GitDiffScopeStaged {
			return agentGitStateChangeType(indexState)
		}
		return agentGitStateChangeType(workState)
	}
	return ""
}

func agentGitStateChangeType(state string) string {
	switch state {
	case "M":
		return workspacesvc.GitChangeTypeModified
	case "A":
		return workspacesvc.GitChangeTypeAdded
	case "D":
		return workspacesvc.GitChangeTypeDeleted
	case "R":
		return workspacesvc.GitChangeTypeRenamed
	case "C":
		return workspacesvc.GitChangeTypeCopied
	case "T":
		return workspacesvc.GitChangeTypeTypeChanged
	case "U":
		return workspacesvc.GitChangeTypeUnmerged
	default:
		return workspacesvc.GitChangeTypeUnknown
	}
}

func (b *dockerRuntimeBackend) editableAgentGitText(content string) (string, error) {
	if int64(len(content)) > b.config.Workspace.PreviewLimitBytes {
		return "", gerror.New("git diff content exceeds editable size limit")
	}
	if !utf8.ValidString(content) || strings.Contains(content, "\x00") {
		return "", gerror.New("git diff content is not editable text")
	}
	return content, nil
}

func isAgentWorkspaceMissingPathError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "workspace path does not exist") ||
		strings.Contains(message, "no such file") ||
		strings.Contains(message, "not found")
}

func readSingleFileFromTar(reader io.Reader) ([]byte, error) {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil, gerror.New("file archive is empty")
		}
		if err != nil {
			return nil, gerror.Wrap(err, "read file archive")
		}
		if header.FileInfo().IsDir() {
			continue
		}
		data, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, gerror.Wrap(err, "read archived file")
		}
		return data, nil
	}
}

func fileContentTar(filename string, content []byte) (io.Reader, error) {
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	header := &tar.Header{
		Name: path.Base(filename),
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := writer.WriteHeader(header); err != nil {
		return nil, gerror.Wrap(err, "write workspace file archive header")
	}
	if len(content) > 0 {
		if _, err := writer.Write(content); err != nil {
			return nil, gerror.Wrap(err, "write workspace file archive content")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, gerror.Wrap(err, "close workspace file archive")
	}
	return bytes.NewReader(buffer.Bytes()), nil
}

type singleFileTarReadCloser struct {
	source    io.ReadCloser
	tarReader *tar.Reader
	current   io.Reader
	ready     bool
	done      bool
}

func newSingleFileTarReadCloser(source io.ReadCloser) io.ReadCloser {
	return &singleFileTarReadCloser{
		source:    source,
		tarReader: tar.NewReader(source),
	}
}

func (r *singleFileTarReadCloser) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	if !r.ready {
		if err := r.openFirstFile(); err != nil {
			return 0, err
		}
	}
	n, err := r.current.Read(p)
	if err == io.EOF {
		r.done = true
	}
	return n, err
}

func (r *singleFileTarReadCloser) Close() error {
	return r.source.Close()
}

func (r *singleFileTarReadCloser) openFirstFile() error {
	for {
		header, err := r.tarReader.Next()
		if err == io.EOF {
			return gerror.New("file archive is empty")
		}
		if err != nil {
			return gerror.Wrap(err, "read file archive")
		}
		if header.FileInfo().IsDir() {
			continue
		}
		r.current = r.tarReader
		r.ready = true
		return nil
	}
}

func pathBase(value string) string {
	value = strings.TrimRight(value, "/")
	if value == "" {
		return "/"
	}
	return path.Base(value)
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
