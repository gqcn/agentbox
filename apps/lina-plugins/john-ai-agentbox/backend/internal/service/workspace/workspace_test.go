// This file verifies workspace service access boundaries and the read-only
// runtime-backed tree/file preview slice. Invisible Agents must still be
// rejected before runtime state is considered.

package workspace

import (
	"bytes"
	"context"
	"io"
	"path"
	"testing"
	"time"

	"lina-core/pkg/bizerr"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

// TestNewRequiresAccessService verifies constructor dependency validation.
func TestNewRequiresAccessService(t *testing.T) {
	if _, err := New(nil, nil); err == nil {
		t.Fatal("expected constructor to reject nil access service")
	}
}

// TestDirectoryTreeRejectsInvisibleAgentBeforeRuntime verifies ownership is checked first.
func TestDirectoryTreeRejectsInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.DirectoryTree(context.Background(), "usr-owner", "agt-other", "/home/agent/workspace/private", true)
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestPathSuggestionsRejectsInvisibleAgentBeforeRuntime verifies suggestions use ownership checks.
func TestPathSuggestionsRejectsInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.PathSuggestions(context.Background(), "usr-owner", "agt-other", "project", "")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestVisibleAgentWithoutRuntimeReturnsRuntimeUnavailable verifies explicit nil runtime degradation.
func TestVisibleAgentWithoutRuntimeReturnsRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.DirectoryTree(context.Background(), "usr-owner", "agt-owned", "", false)
	if !bizerr.Is(err, CodeWorkspaceRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
}

// TestFileDownloadRejectsInvisibleAgentBeforeRuntime verifies downloads check ownership first.
func TestFileDownloadRejectsInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.FileDownload(context.Background(), "usr-owner", "agt-other", "/home/agent/workspace/private.log")
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestNilRuntimeWorkspaceEntrypointsReturnRuntimeUnavailable verifies visible runtime-backed paths do not fake data.
func TestNilRuntimeWorkspaceEntrypointsReturnRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		run  func() error
	}{
		{
			name: "resource",
			run: func() error {
				_, err := service.Resource(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md", ResourceDispositionInline)
				return err
			},
		},
		{
			name: "download",
			run: func() error {
				_, err := service.FileDownload(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md")
				return err
			},
		},
		{
			name: "html-preview",
			run: func() error {
				_, err := service.HtmlPreview(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/index.html")
				return err
			},
		},
		{
			name: "skills",
			run: func() error {
				_, err := service.Skills(context.Background(), "usr-owner", "agt-owned", SkillListInput{Scope: SkillScopeProject, Path: "/home/agent/workspace"})
				return err
			},
		},
	}
	for _, check := range checks {
		if err := check.run(); !bizerr.Is(err, CodeWorkspaceRuntimeUnavailable) {
			t.Fatalf("%s: expected runtime unavailable error, got %v", check.name, err)
		}
	}
}

// TestDirectoryTreeUsesRuntimeBackend verifies visible tree reads delegate to the injected runtime.
func TestDirectoryTreeUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		entries: []RuntimePathEntry{
			{Name: "src", Path: "/home/agent/workspace/src", Type: WorkspaceNodeDirectory, ModifiedAt: time.Unix(20, 0)},
			{Name: "README.md", Path: "/home/agent/workspace/README.md", Type: WorkspaceNodeFile, Size: 12, ModifiedAt: time.Unix(10, 0)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := service.DirectoryTree(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 || nodes[0].Name != "src" || !nodes[0].Expandable || nodes[1].Name != "README.md" {
		t.Fatalf("unexpected tree nodes: %#v", nodes)
	}
}

// TestFilePreviewUsesRuntimeBackend verifies text previews return content and a stable hash.
func TestFilePreviewUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		fileData: []byte("# README\n"),
	})
	if err != nil {
		t.Fatal(err)
	}

	preview, err := service.FilePreview(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if preview.PreviewType != WorkspacePreviewText || preview.Content != "# README\n" || preview.Encoding != "utf-8" || preview.ContentHash == "" {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	if preview.File.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q", preview.File.ContentType)
	}
}

// TestFileDownloadUsesRuntimeBackend verifies downloads return a bounded runtime stream.
func TestFileDownloadUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		fileData: []byte("download body"),
	})
	if err != nil {
		t.Fatal(err)
	}

	stream, err := service.FileDownload(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Reader.Close()
	data, err := io.ReadAll(stream.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "download body" || stream.Disposition != ResourceDispositionAttachment {
		t.Fatalf("unexpected download stream: data=%q stream=%#v", string(data), stream)
	}
	if stream.File.Name != "README.md" || stream.File.ContentType == "" {
		t.Fatalf("unexpected file metadata: %#v", stream.File)
	}
}

// TestResourceUsesRuntimeBackend verifies inline resources share runtime file streaming.
func TestResourceUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		fileData: []byte("resource body"),
	})
	if err != nil {
		t.Fatal(err)
	}

	stream, err := service.Resource(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md", ResourceDispositionInline)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Reader.Close()
	if stream.Disposition != ResourceDispositionInline {
		t.Fatalf("unexpected disposition: %q", stream.Disposition)
	}
}

// TestResourceForUnsafeInlineContentBecomesAttachment verifies active content is not served inline.
func TestResourceForUnsafeInlineContentBecomesAttachment(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		filePath: "/home/agent/workspace/index.html",
		fileData: []byte("<html></html>"),
	})
	if err != nil {
		t.Fatal(err)
	}

	stream, err := service.Resource(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/index.html", ResourceDispositionInline)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Reader.Close()
	if stream.Disposition != ResourceDispositionAttachment {
		t.Fatalf("expected unsafe inline content to become attachment, got %q", stream.Disposition)
	}
}

// TestHtmlPreviewUsesRuntimeBackend verifies HTML previews stream runtime files with sandbox-safe content metadata.
func TestHtmlPreviewUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		filePath: "/home/agent/workspace/report/index.html",
		fileData: []byte("<!doctype html><title>Report</title>"),
	})
	if err != nil {
		t.Fatal(err)
	}

	stream, err := service.HtmlPreview(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/report/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Reader.Close()

	data, err := io.ReadAll(stream.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "<!doctype html><title>Report</title>" {
		t.Fatalf("expected html stream content, got %q", string(data))
	}
	if stream.File.ContentType != "text/html; charset=utf-8" || stream.Disposition != ResourceDispositionInline {
		t.Fatalf("expected html inline stream metadata, got %+v", stream)
	}
}

// TestHtmlPreviewRejectsNonHTMLPath verifies isolated previews only accept HTML file paths.
func TestHtmlPreviewRejectsNonHTMLPath(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{filePath: "/home/agent/workspace/README.md"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.HtmlPreview(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/README.md")
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

// TestFileSaveUsesRuntimeBackend verifies file writes return a refreshed preview.
func TestFileSaveUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{fileData: []byte("old content")}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	preview, err := service.FileSave(context.Background(), "usr-owner", "agt-owned", FileSaveInput{
		Path:     "/home/agent/workspace/README.md",
		Content:  "new content",
		Encoding: "utf-8",
		BaseHash: workspaceContentHash([]byte("old content")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Content != "new content" || preview.ContentHash != workspaceContentHash([]byte("new content")) {
		t.Fatalf("unexpected saved preview: %#v", preview)
	}
	if string(backend.fileData) != "new content" {
		t.Fatalf("runtime content was not updated: %q", string(backend.fileData))
	}
}

// TestFileSaveRejectsStaleBaseHash verifies stale writes return a conflict.
func TestFileSaveRejectsStaleBaseHash(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{fileData: []byte("current")})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.FileSave(context.Background(), "usr-owner", "agt-owned", FileSaveInput{
		Path:     "/home/agent/workspace/README.md",
		Content:  "new content",
		BaseHash: workspaceContentHash([]byte("old")),
	})
	if !bizerr.Is(err, CodeWorkspaceStateConflict) {
		t.Fatalf("expected state conflict, got %v", err)
	}
}

// TestFileCreateUsesRuntimeBackend verifies creating files returns an editable preview.
func TestFileCreateUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{})
	if err != nil {
		t.Fatal(err)
	}

	preview, err := service.FileCreate(context.Background(), "usr-owner", "agt-owned", CreateEntryInput{
		ParentPath: "/home/agent/workspace",
		Name:       "notes.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.File.Path != "/home/agent/workspace/notes.md" || preview.Content != "" || preview.ContentHash == "" {
		t.Fatalf("unexpected created file preview: %#v", preview)
	}
}

// TestDirectoryCreateUsesRuntimeBackend verifies creating directories returns metadata.
func TestDirectoryCreateUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{})
	if err != nil {
		t.Fatal(err)
	}

	info, err := service.DirectoryCreate(context.Background(), "usr-owner", "agt-owned", CreateEntryInput{
		ParentPath: "/home/agent/workspace",
		Name:       "docs",
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Path != "/home/agent/workspace/docs" || info.Type != WorkspaceNodeDirectory {
		t.Fatalf("unexpected directory info: %#v", info)
	}
}

// TestFileUploadUsesRuntimeBackend verifies multipart uploads write bounded files into the target directory.
func TestFileUploadUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	response, err := service.FileUpload(context.Background(), "usr-owner", "agt-owned", FileUploadInput{
		Path: "/home/agent/workspace",
		Files: []UploadFile{{
			Name:   "notes.md",
			Reader: bytes.NewReader([]byte("uploaded content")),
			Size:   int64(len("uploaded content")),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Files) != 1 || response.Files[0].Path != "/home/agent/workspace/notes.md" {
		t.Fatalf("unexpected upload response: %#v", response)
	}
	if string(backend.fileData) != "uploaded content" {
		t.Fatalf("runtime content was not uploaded: %q", string(backend.fileData))
	}
}

// TestFileUploadRejectsInvalidName verifies uploads never preserve path segments from client filenames.
func TestFileUploadRejectsInvalidName(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.FileUpload(context.Background(), "usr-owner", "agt-owned", FileUploadInput{
		Path: "/home/agent/workspace",
		Files: []UploadFile{{
			Name:   "\x00",
			Reader: bytes.NewReader([]byte("content")),
			Size:   int64(len("content")),
		}},
	})
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}

	_, err = service.FileUpload(context.Background(), "usr-owner", "agt-owned", FileUploadInput{
		Path: "/home/agent/workspace",
		Files: []UploadFile{{
			Name:   "nested/secret.md",
			Reader: bytes.NewReader([]byte("content")),
			Size:   int64(len("content")),
		}},
	})
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error for path segment, got %v", err)
	}
}

// TestFileUploadRejectsOversizedFile verifies service-level upload size bounds.
func TestFileUploadRejectsOversizedFile(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.FileUpload(context.Background(), "usr-owner", "agt-owned", FileUploadInput{
		Path: "/home/agent/workspace",
		Files: []UploadFile{{
			Name:   "large.bin",
			Reader: bytes.NewReader([]byte("content")),
			Size:   DefaultWorkspaceUploadFileLimit + 1,
		}},
	})
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

// TestPathSuggestionsUsesRuntimeBackend verifies suggestions are filtered after runtime listing.
func TestPathSuggestionsUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		entries: []RuntimePathEntry{
			{Name: "project-a", Path: "/home/agent/workspace/project-a", Type: WorkspaceNodeDirectory, ModifiedAt: time.Unix(20, 0)},
			{Name: "notes", Path: "/home/agent/workspace/notes", Type: WorkspaceNodeDirectory, ModifiedAt: time.Unix(10, 0)},
			{Name: "project.txt", Path: "/home/agent/workspace/project.txt", Type: WorkspaceNodeFile, ModifiedAt: time.Unix(5, 0)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := service.PathSuggestions(context.Background(), "usr-owner", "agt-owned", "project", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "project-a" {
		t.Fatalf("unexpected suggestions: %#v", items)
	}
}

// TestRejectsInvalidWorkspaceRoot verifies unsupported roots are rejected.
func TestRejectsInvalidWorkspaceRoot(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.DirectoryTree(context.Background(), "usr-owner", "agt-owned", "/workspace", false)
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

// TestRejectsUnsafeWorkspaceEntryName verifies create operations reject path separators.
func TestRejectsUnsafeWorkspaceEntryName(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.FileCreate(context.Background(), "usr-owner", "agt-owned", CreateEntryInput{
		ParentPath: "/home/agent/workspace",
		Name:       "../secret.md",
	})
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

// TestSkillsUsesRuntimeBackend verifies project skill listings are filtered and projected from runtime metadata.
func TestSkillsUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		skills: []SkillInfo{
			{
				Name:        "frontend-review",
				Description: "Review frontend code",
				Scope:       SkillScopeProject,
				Path:        "/home/agent/workspace/project/.agents/skills/frontend-review",
				Source:      "/home/agent/workspace/project",
				HasManifest: true,
			},
			{
				Name:   "backend-audit",
				Scope:  SkillScopeProject,
				Path:   "/home/agent/workspace/project/.agents/skills/backend-audit",
				Source: "/home/agent/workspace/project",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	response, err := service.Skills(context.Background(), "usr-owner", "agt-owned", SkillListInput{
		Scope: SkillScopeProject,
		Path:  "/home/agent/workspace/project",
		Query: "front",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Scope != SkillScopeProject || response.Path != "/home/agent/workspace/project" || len(response.Items) != 1 {
		t.Fatalf("unexpected skills response: %#v", response)
	}
	if response.Items[0].Name != "frontend-review" || !response.Items[0].HasManifest {
		t.Fatalf("unexpected skill item: %#v", response.Items[0])
	}
}

// TestParseSkillMarkdown verifies SKILL.md YAML metadata and folded descriptions are parsed for display only.
func TestParseSkillMarkdown(t *testing.T) {
	name, description := ParseSkillMarkdown([]byte("---\nname: demo\ndescription: >-\n  Demo skill\n  parses folded YAML descriptions\n---\n# Ignored\n"))
	if name != "demo" || description != "Demo skill parses folded YAML descriptions" {
		t.Fatalf("name=%q description=%q", name, description)
	}

	name, description = ParseSkillMarkdown([]byte("# Title\nUseful body summary\n"))
	if name != "" || description != "Useful body summary" {
		t.Fatalf("fallback name=%q description=%q", name, description)
	}
}

// TestRejectsUnsafeGitFilePath verifies Git file paths are repository-relative.
func TestRejectsUnsafeGitFilePath(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.GitFile(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace", "../secret.go")
	if !bizerr.Is(err, CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

// TestGitEntrypointsRejectInvisibleAgentBeforeRuntime verifies Git operations check ownership first.
func TestGitEntrypointsRejectInvisibleAgentBeforeRuntime(t *testing.T) {
	service, err := New(&fakeAccessService{err: bizerr.NewCode(accesssvc.CodeAccessResourceUnavailable)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.GitIndexStage(context.Background(), "usr-owner", "agt-other", GitIndexInput{
		Path:  "/home/agent/workspace",
		Files: []string{"src/main.go"},
	})
	if !bizerr.Is(err, accesssvc.CodeAccessResourceUnavailable) {
		t.Fatalf("expected invisible resource error, got %v", err)
	}
}

// TestNilRuntimeGitEntrypointsReturnRuntimeUnavailable verifies visible Git operations do not fake runtime data.
func TestNilRuntimeGitEntrypointsReturnRuntimeUnavailable(t *testing.T) {
	service, err := New(&fakeAccessService{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.GitDiff(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace", "src/main.go", GitDiffScopeUnstaged)
	if !bizerr.Is(err, CodeWorkspaceRuntimeUnavailable) {
		t.Fatalf("expected runtime unavailable error, got %v", err)
	}
}

// TestGitStatusUsesRuntimeBackend verifies read-only Git status projections from runtime output.
func TestGitStatusUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		gitStatus: &RuntimeGitStatus{
			Path:           "/home/agent/workspace/project",
			RepositoryRoot: "/home/agent/workspace/project",
			Porcelain:      " M tracked.txt\x00?? new.txt\x00A  staged.txt\x00",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	status, err := service.GitStatus(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/project")
	if err != nil {
		t.Fatal(err)
	}
	if status.State != GitRepositoryStateOK || len(status.Changes) != 2 || len(status.StagedChanges) != 1 {
		t.Fatalf("unexpected git status: %#v", status)
	}
	if status.Changes[0].ChangeScope != GitChangeScopeUnstaged || status.StagedChanges[0].ChangeScope != GitChangeScopeStaged {
		t.Fatalf("unexpected git scopes: unstaged=%#v staged=%#v", status.Changes, status.StagedChanges)
	}
	if len(status.ChangeTree) == 0 || len(status.StagedTree) == 0 {
		t.Fatalf("expected git trees, got %#v", status)
	}
}

// TestGitStatusNotRepositoryReturnsProjection verifies non-repositories are not treated as runtime failures.
func TestGitStatusNotRepositoryReturnsProjection(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		gitStatus: &RuntimeGitStatus{
			Path:          "/home/agent/workspace",
			NotRepository: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	status, err := service.GitStatus(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace")
	if err != nil {
		t.Fatal(err)
	}
	if status.State != GitRepositoryStateNotRepo || status.Path != "/home/agent/workspace" {
		t.Fatalf("unexpected not-repo status: %#v", status)
	}
}

// TestGitFileUsesRuntimeBackend verifies Git file previews are read from the injected runtime.
func TestGitFileUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		filePath: "/home/agent/workspace/project/src/main.ts",
		fileData: []byte("console.log('hello')\n"),
	})
	if err != nil {
		t.Fatal(err)
	}

	file, err := service.GitFile(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/project", "src/main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if file.Status != GitChangeTypeModified || file.File.Content != "console.log('hello')\n" || file.File.PreviewType != WorkspacePreviewText {
		t.Fatalf("unexpected git file: %#v", file)
	}
	if file.File.File.Path != "/home/agent/workspace/project/src/main.ts" || file.File.ContentHash == "" {
		t.Fatalf("unexpected file preview metadata: %#v", file.File)
	}
}

// TestGitDiffUsesRuntimeBackend verifies Git diff projection defaults scope and includes text models.
func TestGitDiffUsesRuntimeBackend(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{})
	if err != nil {
		t.Fatal(err)
	}

	diff, err := service.GitDiff(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace/project", "src/main.ts", "")
	if err != nil {
		t.Fatal(err)
	}
	if diff.Scope != GitDiffScopeUnstaged || diff.Status != GitChangeTypeModified {
		t.Fatalf("unexpected git diff status/scope: %#v", diff)
	}
	if diff.OriginalContent != "one\n" || diff.ModifiedContent != "two\n" || diff.Language != "typescript" {
		t.Fatalf("unexpected git diff content: %#v", diff)
	}
}

// TestGitStageUsesRuntimeBackend verifies selected files are staged through the runtime backend.
func TestGitStageUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	status, err := service.GitIndexStage(context.Background(), "usr-owner", "agt-owned", GitIndexInput{
		Path:  "/home/agent/workspace/project",
		Files: []string{"src/main.ts"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if backend.lastGitStage == nil || backend.lastGitStage.Path != "/home/agent/workspace/project" || backend.lastGitStage.Files[0] != "src/main.ts" {
		t.Fatalf("runtime stage input not captured: %#v", backend.lastGitStage)
	}
	if status.State != GitRepositoryStateOK || len(status.StagedChanges) != 1 {
		t.Fatalf("unexpected staged status: %#v", status)
	}
}

// TestGitUnstageAllUsesRuntimeBackend verifies all-index mutation is delegated to runtime.
func TestGitUnstageAllUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.GitIndexUnstage(context.Background(), "usr-owner", "agt-owned", GitIndexInput{
		Path: "/home/agent/workspace/project",
		All:  true,
	}); err != nil {
		t.Fatal(err)
	}
	if backend.lastGitUnstage == nil || !backend.lastGitUnstage.All {
		t.Fatalf("runtime unstage-all input not captured: %#v", backend.lastGitUnstage)
	}
}

// TestGitDiscardUsesRuntimeBackend verifies worktree discard is delegated after validation.
func TestGitDiscardUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.GitChangesDiscard(context.Background(), "usr-owner", "agt-owned", GitDiscardInput{
		Path:  "/home/agent/workspace/project",
		Files: []string{"src/main.ts"},
	}); err != nil {
		t.Fatal(err)
	}
	if backend.lastGitDiscard == nil || backend.lastGitDiscard.Files[0] != "src/main.ts" {
		t.Fatalf("runtime discard input not captured: %#v", backend.lastGitDiscard)
	}
}

// TestGitCommitUsesRuntimeBackend verifies commits are created through the runtime backend.
func TestGitCommitUsesRuntimeBackend(t *testing.T) {
	backend := &fakeRuntimeBackend{}
	service, err := New(&fakeAccessService{}, backend)
	if err != nil {
		t.Fatal(err)
	}

	commit, err := service.GitCommit(context.Background(), "usr-owner", "agt-owned", GitCommitInput{
		Path:    "/home/agent/workspace/project",
		Message: "Update workspace source control",
		Push:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if backend.lastGitCommit == nil || backend.lastGitCommit.Message != "Update workspace source control" || !backend.lastGitCommit.Push {
		t.Fatalf("runtime commit input not captured: %#v", backend.lastGitCommit)
	}
	if commit.CommitHash != "abc123" || !commit.Pushed || commit.Status.State != GitRepositoryStateClean {
		t.Fatalf("unexpected commit response: %#v", commit)
	}
}

// TestGitFileNotRepositoryReturnsProjection verifies non-repositories do not fail as runtime errors.
func TestGitFileNotRepositoryReturnsProjection(t *testing.T) {
	service, err := New(&fakeAccessService{}, &fakeRuntimeBackend{
		gitFile: &RuntimeGitFile{NotRepository: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	file, err := service.GitFile(context.Background(), "usr-owner", "agt-owned", "/home/agent/workspace", "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if file.Message == "" {
		t.Fatalf("expected not-repository message, got %#v", file)
	}
}

type fakeAccessService struct {
	accesssvc.Service
	err error
}

func (s *fakeAccessService) EnsureWorkspaceResourceVisible(context.Context, string, string, string) error {
	return s.err
}

type fakeRuntimeBackend struct {
	entries   []RuntimePathEntry
	filePath  string
	fileData  []byte
	dirs      map[string]RuntimePathEntry
	gitStatus *RuntimeGitStatus
	gitFile   *RuntimeGitFile
	gitDiff   *RuntimeGitDiff
	skills    []SkillInfo
	err       error

	lastGitStage   *RuntimeGitIndexInput
	lastGitUnstage *RuntimeGitIndexInput
	lastGitDiscard *RuntimeGitDiscardInput
	lastGitCommit  *RuntimeGitCommitInput
}

func (b *fakeRuntimeBackend) WorkspacePathStat(_ context.Context, _ string, _ string, workspacePath string) (*RuntimePathEntry, error) {
	if b.err != nil {
		return nil, b.err
	}
	filePath := b.filePath
	if filePath == "" {
		filePath = "/home/agent/workspace/README.md"
	}
	if workspacePath == filePath {
		return &RuntimePathEntry{
			Name:       fakePathBase(filePath),
			Path:       workspacePath,
			Type:       WorkspaceNodeFile,
			Size:       int64(len(b.fileData)),
			ModifiedAt: time.Unix(2, 0),
		}, nil
	}
	return &RuntimePathEntry{
		Name:       "workspace",
		Path:       DefaultWorkspaceRootPath,
		Type:       WorkspaceNodeDirectory,
		ModifiedAt: time.Unix(1, 0),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceDirectoryEntries(context.Context, string, string, string, bool) ([]RuntimePathEntry, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.entries, nil
}

func (b *fakeRuntimeBackend) WorkspaceReadFile(context.Context, string, string, string) ([]byte, *RuntimePathEntry, error) {
	if b.err != nil {
		return nil, nil, b.err
	}
	data := b.fileData
	if data == nil {
		data = []byte("workspace file")
	}
	filePath := b.filePath
	if filePath == "" {
		filePath = "/home/agent/workspace/README.md"
	}
	return data, &RuntimePathEntry{
		Name:       fakePathBase(filePath),
		Path:       filePath,
		Type:       WorkspaceNodeFile,
		Size:       int64(len(data)),
		ModifiedAt: time.Unix(2, 0),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceOpenFile(context.Context, string, string, string) (*RuntimeFile, error) {
	if b.err != nil {
		return nil, b.err
	}
	data := b.fileData
	if data == nil {
		data = []byte("workspace file")
	}
	filePath := b.filePath
	if filePath == "" {
		filePath = "/home/agent/workspace/README.md"
	}
	return &RuntimeFile{
		Entry: RuntimePathEntry{
			Name:       fakePathBase(filePath),
			Path:       filePath,
			Type:       WorkspaceNodeFile,
			Size:       int64(len(data)),
			ModifiedAt: time.Unix(2, 0),
		},
		Reader: io.NopCloser(bytes.NewReader(data)),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceWriteFile(_ context.Context, _ string, _ string, input RuntimeWriteFileInput) (*RuntimePathEntry, error) {
	if b.err != nil {
		return nil, b.err
	}
	b.filePath = input.Path
	b.fileData = append([]byte(nil), input.Content...)
	return &RuntimePathEntry{
		Name:       fakePathBase(input.Path),
		Path:       input.Path,
		Type:       WorkspaceNodeFile,
		Size:       int64(len(input.Content)),
		ModifiedAt: time.Unix(3, 0),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceUploadFile(_ context.Context, _ string, _ string, input RuntimeUploadFileInput) (*RuntimePathEntry, error) {
	if b.err != nil {
		return nil, b.err
	}
	targetPath := input.DirectoryPath + "/" + input.Name
	b.filePath = targetPath
	b.fileData = append([]byte(nil), input.Content...)
	return &RuntimePathEntry{
		Name:       input.Name,
		Path:       targetPath,
		Type:       WorkspaceNodeFile,
		Size:       int64(len(input.Content)),
		ModifiedAt: time.Unix(4, 0),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceCreateEntry(_ context.Context, _ string, _ string, input RuntimeCreateEntryInput) (*RuntimePathEntry, error) {
	if b.err != nil {
		return nil, b.err
	}
	targetPath := input.ParentPath + "/" + input.Name
	if input.Type == WorkspaceNodeDirectory {
		if b.dirs == nil {
			b.dirs = map[string]RuntimePathEntry{}
		}
		entry := RuntimePathEntry{
			Name:       input.Name,
			Path:       targetPath,
			Type:       WorkspaceNodeDirectory,
			ModifiedAt: time.Unix(4, 0),
		}
		b.dirs[targetPath] = entry
		return &entry, nil
	}
	b.filePath = targetPath
	b.fileData = []byte{}
	return &RuntimePathEntry{
		Name:       input.Name,
		Path:       targetPath,
		Type:       WorkspaceNodeFile,
		ModifiedAt: time.Unix(4, 0),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitStatus(_ context.Context, _ string, _ string, workspacePath string) (*RuntimeGitStatus, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.gitStatus != nil {
		return b.gitStatus, nil
	}
	return &RuntimeGitStatus{
		Path:           workspacePath,
		RepositoryRoot: workspacePath,
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitFile(_ context.Context, _ string, _ string, _ string, file string) (*RuntimeGitFile, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.gitFile != nil {
		return b.gitFile, nil
	}
	data := b.fileData
	if data == nil {
		data = []byte("workspace file")
	}
	filePath := b.filePath
	if filePath == "" {
		filePath = DefaultWorkspaceRootPath + "/" + file
	}
	return &RuntimeGitFile{
		File: RuntimePathEntry{
			Name:       fakePathBase(filePath),
			Path:       filePath,
			Type:       WorkspaceNodeFile,
			Size:       int64(len(data)),
			ModifiedAt: time.Unix(2, 0),
		},
		Content: data,
		Status:  GitChangeTypeModified,
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitDiff(_ context.Context, _ string, _ string, _ string, file string, scope string) (*RuntimeGitDiff, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.gitDiff != nil {
		return b.gitDiff, nil
	}
	return &RuntimeGitDiff{
		Path:            file,
		Status:          GitChangeTypeModified,
		Scope:           scope,
		Diff:            "diff --git a/" + file + " b/" + file + "\n-one\n+two\n",
		OriginalContent: "one\n",
		ModifiedContent: "two\n",
		OriginalPath:    file,
		ModifiedPath:    file,
		Language:        gitDiffLanguage(file),
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitStage(_ context.Context, _ string, _ string, input RuntimeGitIndexInput) (*RuntimeGitStatus, error) {
	if b.err != nil {
		return nil, b.err
	}
	copied := input
	copied.Files = append([]string(nil), input.Files...)
	b.lastGitStage = &copied
	return &RuntimeGitStatus{
		Path:           input.Path,
		RepositoryRoot: input.Path,
		Porcelain:      "A  src/main.ts\x00",
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitUnstage(_ context.Context, _ string, _ string, input RuntimeGitIndexInput) (*RuntimeGitStatus, error) {
	if b.err != nil {
		return nil, b.err
	}
	copied := input
	copied.Files = append([]string(nil), input.Files...)
	b.lastGitUnstage = &copied
	return &RuntimeGitStatus{
		Path:           input.Path,
		RepositoryRoot: input.Path,
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitDiscard(_ context.Context, _ string, _ string, input RuntimeGitDiscardInput) (*RuntimeGitStatus, error) {
	if b.err != nil {
		return nil, b.err
	}
	copied := input
	copied.Files = append([]string(nil), input.Files...)
	b.lastGitDiscard = &copied
	return &RuntimeGitStatus{
		Path:           input.Path,
		RepositoryRoot: input.Path,
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceGitCommit(_ context.Context, _ string, _ string, input RuntimeGitCommitInput) (*RuntimeGitCommitResult, error) {
	if b.err != nil {
		return nil, b.err
	}
	copied := input
	b.lastGitCommit = &copied
	return &RuntimeGitCommitResult{
		CommitHash: "abc123",
		Pushed:     input.Push,
		Status: RuntimeGitStatus{
			Path:           input.Path,
			RepositoryRoot: input.Path,
		},
	}, nil
}

func (b *fakeRuntimeBackend) WorkspaceSkills(_ context.Context, _ string, _ string, scope string, workspacePath string) ([]SkillInfo, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.skills != nil {
		return b.skills, nil
	}
	return []SkillInfo{{
		Name:        "frontend-review",
		Description: "Review frontend changes",
		Scope:       scope,
		Path:        workspacePath + "/.agents/skills/frontend-review",
		Source:      workspacePath,
		HasManifest: true,
	}}, nil
}

func fakePathBase(value string) string {
	name := path.Base(value)
	if name == "." || name == "/" {
		return "workspace"
	}
	return name
}
