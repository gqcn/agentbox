// This file verifies workspace controller ownership plumbing. The controller
// must use the authenticated AgentBox user from context for all workspace
// service calls and reject calls without plugin authentication context.

package workspace

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"

	v1 "john-ai-agentbox/backend/api/workspace/v1"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// TestDirectoryTreeUsesAuthenticatedUser verifies tree calls are user scoped.
func TestDirectoryTreeUsesAuthenticatedUser(t *testing.T) {
	service := &fakeWorkspaceService{
		nodes: []workspacesvc.TreeNode{{
			Name:       "project",
			Path:       "/home/agent/workspace/project",
			Type:       workspacesvc.WorkspaceNodeDirectory,
			Expandable: true,
		}},
	}
	controller := newTestController(t, service)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	res, err := controller.DirectoryTree(ctx, &v1.DirectoryTreeReq{ID: "agt-test", Path: "/home/agent/workspace", IncludeFiles: true})
	if err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" || service.lastPath != "/home/agent/workspace" || !service.lastIncludeFiles {
		t.Fatalf("unexpected service call: user=%q agent=%q path=%q include=%v", service.lastUserID, service.lastAgentID, service.lastPath, service.lastIncludeFiles)
	}
	if len(*res) != 1 || (*res)[0].Name != "project" {
		t.Fatalf("unexpected tree response: %#v", *res)
	}
}

// TestPathSuggestionsRequiresAuthenticatedUser verifies missing auth is rejected.
func TestPathSuggestionsRequiresAuthenticatedUser(t *testing.T) {
	controller := newTestController(t, &fakeWorkspaceService{})

	_, err := controller.PathSuggestions(context.Background(), &v1.PathSuggestionsReq{ID: "agt-test", Query: "project"})
	if !bizerr.Is(err, authsvc.CodeAuthRequired) {
		t.Fatalf("expected auth required error, got %v", err)
	}
}

// TestFileDownloadUsesAuthenticatedUser verifies download calls are user scoped.
func TestFileDownloadUsesAuthenticatedUser(t *testing.T) {
	service := &fakeWorkspaceService{
		fileStream: &workspacesvc.FileStream{
			File: workspacesvc.FileInfo{
				Name:        "README.md",
				Path:        "/home/agent/workspace/README.md",
				Type:        workspacesvc.WorkspaceNodeFile,
				Size:        4,
				ContentType: "text/plain; charset=utf-8",
			},
			Reader:      io.NopCloser(bytes.NewReader([]byte("body"))),
			Disposition: workspacesvc.ResourceDispositionAttachment,
		},
	}
	controller := newTestController(t, service)
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	if _, err := controller.FileDownload(ctx, &v1.FileDownloadReq{ID: "agt-test", Path: "/home/agent/workspace/README.md"}); err != nil {
		t.Fatal(err)
	}
	if service.lastUserID != "usr-owner" || service.lastAgentID != "agt-test" || service.lastPath != "/home/agent/workspace/README.md" {
		t.Fatalf("unexpected service call: user=%q agent=%q path=%q", service.lastUserID, service.lastAgentID, service.lastPath)
	}
}

// TestFileUploadRequiresMultipartRequest verifies upload handlers require request file context.
func TestFileUploadRequiresMultipartRequest(t *testing.T) {
	controller := newTestController(t, &fakeWorkspaceService{})
	ctx := authctx.WithUser(context.Background(), authsvc.UserOutput{ID: "usr-owner", Username: "owner"})

	_, err := controller.FileUpload(ctx, &v1.FileUploadReq{ID: "agt-test", Path: "/home/agent/workspace"})
	if !bizerr.Is(err, workspacesvc.CodeWorkspaceInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func newTestController(t *testing.T, workspaceSvc workspacesvc.Service) *ControllerV1 {
	t.Helper()
	controller, err := NewV1(workspaceSvc)
	if err != nil {
		t.Fatal(err)
	}
	typed, ok := controller.(*ControllerV1)
	if !ok {
		t.Fatalf("unexpected controller type %T", controller)
	}
	return typed
}

// TestWorkspaceUploadFilesFromUploads verifies multipart file objects are converted to service streams.
func TestWorkspaceUploadFilesFromUploads(t *testing.T) {
	uploads := ghttp.UploadFiles{newTestUploadFile(t, "notes.md", "body")}

	files, cleanup, err := workspaceUploadFilesFromUploads(uploads)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "notes.md" || files[0].Size != int64(len("body")) {
		t.Fatalf("unexpected upload files: %#v", files)
	}
	data, err := io.ReadAll(files[0].Reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "body" {
		t.Fatalf("unexpected upload content: %q", string(data))
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}

func newTestUploadFile(t *testing.T, filename string, body string) *ghttp.UploadFile {
	t.Helper()
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(&requestBody, writer.Boundary())
	form, err := reader.ReadForm(int64(len(body)) + 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := form.RemoveAll(); err != nil {
			t.Fatal(err)
		}
	})
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one multipart file, got %d", len(files))
	}
	return &ghttp.UploadFile{FileHeader: files[0]}
}

type fakeWorkspaceService struct {
	workspacesvc.Service
	lastUserID       string
	lastAgentID      string
	lastPath         string
	lastQuery        string
	lastIncludeFiles bool
	suggestions      []workspacesvc.PathSuggestion
	nodes            []workspacesvc.TreeNode
	fileStream       *workspacesvc.FileStream
	uploadResponse   *workspacesvc.UploadResponse
	lastUploadFiles  []workspacesvc.UploadFile
	err              error
}

func (s *fakeWorkspaceService) PathSuggestions(_ context.Context, userID string, agentID string, query string, fallbackPath string) ([]workspacesvc.PathSuggestion, error) {
	s.lastUserID, s.lastAgentID, s.lastQuery, s.lastPath = userID, agentID, query, fallbackPath
	if s.err != nil {
		return nil, s.err
	}
	return s.suggestions, nil
}

func (s *fakeWorkspaceService) DirectoryTree(_ context.Context, userID string, agentID string, path string, includeFiles bool) ([]workspacesvc.TreeNode, error) {
	s.lastUserID, s.lastAgentID, s.lastPath, s.lastIncludeFiles = userID, agentID, path, includeFiles
	if s.err != nil {
		return nil, s.err
	}
	return s.nodes, nil
}

func (s *fakeWorkspaceService) FileDownload(_ context.Context, userID string, agentID string, path string) (*workspacesvc.FileStream, error) {
	s.lastUserID, s.lastAgentID, s.lastPath = userID, agentID, path
	if s.err != nil {
		return nil, s.err
	}
	return s.fileStream, nil
}

func (s *fakeWorkspaceService) Resource(_ context.Context, userID string, agentID string, path string, disposition string) (*workspacesvc.FileStream, error) {
	s.lastUserID, s.lastAgentID, s.lastPath = userID, agentID, path
	if s.err != nil {
		return nil, s.err
	}
	if s.fileStream != nil {
		s.fileStream.Disposition = disposition
	}
	return s.fileStream, nil
}

func (s *fakeWorkspaceService) FileUpload(_ context.Context, userID string, agentID string, input workspacesvc.FileUploadInput) (*workspacesvc.UploadResponse, error) {
	s.lastUserID, s.lastAgentID, s.lastPath, s.lastUploadFiles = userID, agentID, input.Path, input.Files
	if s.err != nil {
		return nil, s.err
	}
	return s.uploadResponse, nil
}
