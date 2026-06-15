// This file implements workspace path normalization and current migration
// behavior. It deliberately checks ownership before reporting runtime
// availability so invisible Agents do not leak path, status, or runtime state.

package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
	"lina-core/pkg/bizerr"
)

// PathSuggestions returns bounded path suggestions for one visible Agent.
func (s *serviceImpl) PathSuggestions(ctx context.Context, userID string, agentID string, query string, fallbackPath string) ([]PathSuggestion, error) {
	userID, agentID = strings.TrimSpace(userID), strings.TrimSpace(agentID)
	workspacePath, err := s.normalizeWorkspacePath(firstNonEmpty(query, fallbackPath))
	if err != nil {
		return nil, err
	}
	if err := s.accessSvc.EnsureWorkspaceResourceVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	searchRoot := workspacePath
	if stat, statErr := s.runtimeBackend.WorkspacePathStat(ctx, userID, agentID, searchRoot); statErr != nil || stat.Type != WorkspaceNodeDirectory {
		searchRoot = path.Dir(searchRoot)
	}
	entries, err := s.runtimeBackend.WorkspaceDirectoryEntries(ctx, userID, agentID, searchRoot, false)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	queryText := strings.ToLower(path.Base(strings.TrimRight(workspacePath, "/")))
	items := make([]PathSuggestion, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != WorkspaceNodeDirectory {
			continue
		}
		if queryText != "." && queryText != "/" && queryText != "" && !strings.Contains(strings.ToLower(entry.Name), queryText) {
			continue
		}
		items = append(items, PathSuggestion{Name: entry.Name, Path: entry.Path})
	}
	return items, nil
}

// DirectoryTree returns immediate tree nodes for one visible Agent.
func (s *serviceImpl) DirectoryTree(ctx context.Context, userID string, agentID string, inputPath string, includeFiles bool) ([]TreeNode, error) {
	userID, agentID = strings.TrimSpace(userID), strings.TrimSpace(agentID)
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return nil, err
	}
	if err := s.accessSvc.EnsureWorkspaceResourceVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	stat, err := s.runtimeBackend.WorkspacePathStat(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if stat.Type != WorkspaceNodeDirectory {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	entries, err := s.runtimeBackend.WorkspaceDirectoryEntries(ctx, userID, agentID, workspacePath, includeFiles)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	nodes := make([]TreeNode, 0, len(entries))
	for _, entry := range entries {
		nodes = append(nodes, treeNodeFromRuntime(entry))
	}
	return nodes, nil
}

// FilePreview validates file path visibility before returning a runtime-backed preview.
func (s *serviceImpl) FilePreview(ctx context.Context, userID string, agentID string, inputPath string) (*FilePreview, error) {
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
	stat, err := s.runtimeBackend.WorkspacePathStat(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if stat.Type == WorkspaceNodeDirectory {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	fileInfo := fileInfoFromRuntime(*stat)
	fileInfo.ContentType = detectContentTypeByPath(stat.Path)
	preview := &FilePreview{
		File:        fileInfo,
		PreviewType: WorkspacePreviewUnsupported,
		DownloadURL: workspaceDownloadURL(agentID, stat.Path),
	}
	if isImageContentType(fileInfo.ContentType) {
		preview.PreviewType = WorkspacePreviewImage
		return preview, nil
	}
	if stat.Size > s.config.PreviewLimitBytes {
		preview.PreviewType = WorkspacePreviewText
		preview.TooLarge = true
		return preview, nil
	}
	data, stat, err := s.runtimeBackend.WorkspaceReadFile(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if stat != nil {
		preview.File = fileInfoFromRuntime(*stat)
		preview.File.ContentType = detectContentTypeBytes(stat.Path, data)
	}
	if !isLikelyText(data, preview.File.ContentType) {
		return preview, nil
	}
	preview.PreviewType = WorkspacePreviewText
	preview.Content = string(data)
	preview.Encoding = "utf-8"
	preview.ContentHash = workspaceContentHash(data)
	return preview, nil
}

// FileSave validates file path visibility before saving runtime-backed file content.
func (s *serviceImpl) FileSave(ctx context.Context, userID string, agentID string, input FileSaveInput) (*FilePreview, error) {
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
	if encoding := strings.TrimSpace(input.Encoding); encoding != "" && !strings.EqualFold(encoding, "utf-8") {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if strings.TrimSpace(input.BaseHash) != "" {
		data, _, err := s.runtimeBackend.WorkspaceReadFile(ctx, userID, agentID, workspacePath)
		if err != nil {
			return nil, wrapWorkspaceRuntimeUnavailable(err)
		}
		if workspaceContentHash(data) != strings.TrimSpace(input.BaseHash) {
			return nil, bizerr.NewCode(CodeWorkspaceStateConflict)
		}
	}
	entry, err := s.runtimeBackend.WorkspaceWriteFile(ctx, userID, agentID, RuntimeWriteFileInput{
		Path:    workspacePath,
		Content: []byte(input.Content),
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return s.filePreviewFromRuntime(ctx, userID, agentID, workspacePath, entry)
}

// FileCreate validates parent directory visibility before creating a runtime-backed file.
func (s *serviceImpl) FileCreate(ctx context.Context, userID string, agentID string, input CreateEntryInput) (*FilePreview, error) {
	if err := validateWorkspaceEntryName(input.Name); err != nil {
		return nil, err
	}
	parentPath, err := s.normalizeWorkspacePath(input.ParentPath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, parentPath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	entry, err := s.runtimeBackend.WorkspaceCreateEntry(ctx, userID, agentID, RuntimeCreateEntryInput{
		ParentPath: parentPath,
		Name:       strings.TrimSpace(input.Name),
		Type:       WorkspaceNodeFile,
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	return s.filePreviewFromRuntime(ctx, userID, agentID, entry.Path, entry)
}

// DirectoryCreate validates parent directory visibility before creating a runtime-backed directory.
func (s *serviceImpl) DirectoryCreate(ctx context.Context, userID string, agentID string, input CreateEntryInput) (*FileInfo, error) {
	if err := validateWorkspaceEntryName(input.Name); err != nil {
		return nil, err
	}
	parentPath, err := s.normalizeWorkspacePath(input.ParentPath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, parentPath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	entry, err := s.runtimeBackend.WorkspaceCreateEntry(ctx, userID, agentID, RuntimeCreateEntryInput{
		ParentPath: parentPath,
		Name:       strings.TrimSpace(input.Name),
		Type:       WorkspaceNodeDirectory,
	})
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	fileInfo := fileInfoFromRuntime(*entry)
	fileInfo.ContentType = detectContentTypeByPath(entry.Path)
	return &fileInfo, nil
}

// FileUpload validates target directory visibility before uploading runtime-backed files.
func (s *serviceImpl) FileUpload(ctx context.Context, userID string, agentID string, input FileUploadInput) (*UploadResponse, error) {
	if len(input.Files) == 0 || len(input.Files) > s.config.UploadCountLimit {
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
	stat, err := s.runtimeBackend.WorkspacePathStat(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if stat.Type != WorkspaceNodeDirectory {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	uploaded := make([]FileInfo, 0, len(input.Files))
	for _, file := range input.Files {
		name, err := safeUploadFileName(file.Name)
		if err != nil {
			return nil, err
		}
		if file.Size > s.config.UploadFileLimitBytes {
			return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
		}
		content, err := readBoundedUploadFile(file.Reader, s.config.UploadFileLimitBytes)
		if err != nil {
			return nil, err
		}
		entry, err := s.runtimeBackend.WorkspaceUploadFile(ctx, userID, agentID, RuntimeUploadFileInput{
			DirectoryPath: workspacePath,
			Name:          name,
			Content:       content,
		})
		if err != nil {
			return nil, wrapWorkspaceRuntimeUnavailable(err)
		}
		fileInfo := fileInfoFromRuntime(*entry)
		fileInfo.ContentType = detectContentTypeBytes(entry.Path, content)
		uploaded = append(uploaded, fileInfo)
	}
	return &UploadResponse{Files: uploaded}, nil
}

// FileDownload validates file visibility before returning a runtime-backed attachment stream.
func (s *serviceImpl) FileDownload(ctx context.Context, userID string, agentID string, inputPath string) (*FileStream, error) {
	return s.openWorkspaceFile(ctx, userID, agentID, inputPath, ResourceDispositionAttachment)
}

// Resource validates file visibility before returning a runtime-backed resource stream.
func (s *serviceImpl) Resource(ctx context.Context, userID string, agentID string, inputPath string, disposition string) (*FileStream, error) {
	disposition = strings.TrimSpace(disposition)
	if disposition != "" && disposition != ResourceDispositionInline && disposition != ResourceDispositionAttachment {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if disposition == "" {
		disposition = ResourceDispositionInline
	}
	return s.openWorkspaceFile(ctx, userID, agentID, inputPath, disposition)
}

// HtmlPreview validates HTML file visibility before returning a runtime-backed isolated preview stream.
func (s *serviceImpl) HtmlPreview(ctx context.Context, userID string, agentID string, inputPath string) (*FileStream, error) {
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return nil, err
	}
	lowerPath := strings.ToLower(workspacePath)
	if !isHTMLWorkspacePath(lowerPath) {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if err := s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	runtimeFile, err := s.runtimeBackend.WorkspaceOpenFile(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if !isHTMLWorkspacePath(runtimeFile.Entry.Path) {
		if closeErr := runtimeFile.Reader.Close(); closeErr != nil {
			return nil, wrapWorkspaceRuntimeUnavailable(closeErr)
		}
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	fileInfo := fileInfoFromRuntime(runtimeFile.Entry)
	fileInfo.ContentType = "text/html; charset=utf-8"
	return &FileStream{
		File:        fileInfo,
		Reader:      runtimeFile.Reader,
		Disposition: ResourceDispositionInline,
	}, nil
}

// Skills validates Agent and workspace visibility before listing runtime-backed skills.
func (s *serviceImpl) Skills(ctx context.Context, userID string, agentID string, input SkillListInput) (*SkillListResponse, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope != "" && scope != SkillScopeGlobal && scope != SkillScopeProject {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if scope == "" {
		scope = SkillScopeGlobal
	}
	workspacePath := input.Path
	if scope != SkillScopeGlobal {
		workspacePath = firstNonEmpty(input.Path, s.config.WorkspaceRootPath)
	}
	if err := s.ensureWorkspacePathVisible(ctx, userID, agentID, workspacePath); err != nil {
		return nil, err
	}
	if s.runtimeBackend == nil {
		return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	items, err := s.runtimeBackend.WorkspaceSkills(ctx, strings.TrimSpace(userID), strings.TrimSpace(agentID), scope, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	items = filterSkills(items, input.Query)
	sortSkills(items)
	if len(items) > s.config.SkillListLimit {
		items = items[:s.config.SkillListLimit]
	}
	return &SkillListResponse{
		Scope: scope,
		Path:  s.normalizeSkillResponsePath(scope, workspacePath),
		Items: items,
	}, nil
}

// SkillUpload validates Agent and workspace visibility before uploading project skills.
func (s *serviceImpl) SkillUpload(ctx context.Context, userID string, agentID string, input SkillUploadInput) (*SkillUploadResponse, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope != "" && scope != SkillScopeProject {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if err := s.ensureWorkspacePathVisible(ctx, userID, agentID, input.Path); err != nil {
		return nil, err
	}
	return nil, bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
}

func isHTMLWorkspacePath(workspacePath string) bool {
	lowerPath := strings.ToLower(strings.TrimSpace(workspacePath))
	return strings.HasSuffix(lowerPath, ".html") || strings.HasSuffix(lowerPath, ".htm")
}

func (s *serviceImpl) ensureWorkspacePathVisible(ctx context.Context, userID string, agentID string, inputPath string) error {
	workspacePath, err := s.normalizeWorkspacePath(inputPath)
	if err != nil {
		return err
	}
	return s.ensureNormalizedWorkspacePathVisible(ctx, userID, agentID, workspacePath)
}

func (s *serviceImpl) ensureNormalizedWorkspacePathVisible(ctx context.Context, userID string, agentID string, workspacePath string) error {
	userID, agentID = strings.TrimSpace(userID), strings.TrimSpace(agentID)
	return s.accessSvc.EnsureWorkspaceResourceVisible(ctx, userID, agentID, workspacePath)
}

func (s *serviceImpl) openWorkspaceFile(ctx context.Context, userID string, agentID string, inputPath string, disposition string) (*FileStream, error) {
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
	runtimeFile, err := s.runtimeBackend.WorkspaceOpenFile(ctx, userID, agentID, workspacePath)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	fileInfo := fileInfoFromRuntime(runtimeFile.Entry)
	fileInfo.ContentType = detectContentTypeByPath(fileInfo.Path)
	if disposition == "" {
		disposition = ResourceDispositionInline
	}
	if disposition == ResourceDispositionInline && !isSafeInlineContentType(fileInfo.ContentType) {
		disposition = ResourceDispositionAttachment
	}
	return &FileStream{
		File:        fileInfo,
		Reader:      runtimeFile.Reader,
		Disposition: disposition,
	}, nil
}

func (s *serviceImpl) filePreviewFromRuntime(ctx context.Context, userID string, agentID string, workspacePath string, entry *RuntimePathEntry) (*FilePreview, error) {
	if entry == nil {
		return s.FilePreview(ctx, userID, agentID, workspacePath)
	}
	if entry.Type == WorkspaceNodeDirectory {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	fileInfo := fileInfoFromRuntime(*entry)
	fileInfo.ContentType = detectContentTypeByPath(entry.Path)
	preview := &FilePreview{
		File:        fileInfo,
		PreviewType: WorkspacePreviewUnsupported,
		DownloadURL: workspaceDownloadURL(agentID, entry.Path),
	}
	if isImageContentType(fileInfo.ContentType) {
		preview.PreviewType = WorkspacePreviewImage
		return preview, nil
	}
	if entry.Size > s.config.PreviewLimitBytes {
		preview.PreviewType = WorkspacePreviewText
		preview.TooLarge = true
		return preview, nil
	}
	data, stat, err := s.runtimeBackend.WorkspaceReadFile(ctx, userID, agentID, entry.Path)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if stat != nil {
		preview.File = fileInfoFromRuntime(*stat)
		preview.File.ContentType = detectContentTypeBytes(stat.Path, data)
	}
	if !isLikelyText(data, preview.File.ContentType) {
		return preview, nil
	}
	preview.PreviewType = WorkspacePreviewText
	preview.Content = string(data)
	preview.Encoding = "utf-8"
	preview.ContentHash = workspaceContentHash(data)
	return preview, nil
}

func (s *serviceImpl) normalizeWorkspacePath(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" || value == "." {
		return s.config.WorkspaceRootPath, nil
	}
	if !strings.HasPrefix(value, "/") {
		value = path.Join(s.config.WorkspaceRootPath, value)
	}
	cleaned := path.Clean(value)
	if cleaned == s.config.WorkspaceRootPath || strings.HasPrefix(cleaned, s.config.WorkspaceRootPath+"/") {
		return cleaned, nil
	}
	if cleaned == s.config.SharedRootPath || strings.HasPrefix(cleaned, s.config.SharedRootPath+"/") {
		return cleaned, nil
	}
	return "", bizerr.NewCode(CodeWorkspaceInvalidInput)
}

func validateWorkspaceEntryName(name string) error {
	value := strings.TrimSpace(name)
	if value == "" || value == "." || value == ".." {
		return bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	return nil
}

func safeUploadFileName(name string) (string, error) {
	value := strings.TrimSpace(name)
	if value == "" || value == "." || value == ".." || strings.Contains(value, "\x00") {
		return "", bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	if err := validateWorkspaceEntryName(value); err != nil {
		return "", err
	}
	return value, nil
}

func readBoundedUploadFile(reader io.Reader, limit int64) ([]byte, error) {
	if reader == nil {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	limited := io.LimitReader(reader, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, wrapWorkspaceRuntimeUnavailable(err)
	}
	if int64(len(data)) > limit {
		return nil, bizerr.NewCode(CodeWorkspaceInvalidInput)
	}
	return data, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func treeNodeFromRuntime(entry RuntimePathEntry) TreeNode {
	return TreeNode{
		Name:       entry.Name,
		Path:       entry.Path,
		Type:       entry.Type,
		Size:       entry.Size,
		ModifiedAt: entry.ModifiedAt.UnixMilli(),
		Expandable: entry.Type == WorkspaceNodeDirectory,
	}
}

func fileInfoFromRuntime(entry RuntimePathEntry) FileInfo {
	return FileInfo{
		Name:       entry.Name,
		Path:       entry.Path,
		Type:       entry.Type,
		Size:       entry.Size,
		ModifiedAt: entry.ModifiedAt.UnixMilli(),
	}
}

func detectContentTypeByPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if typ := mime.TypeByExtension(ext); typ != "" {
		return typ
	}
	switch ext {
	case ".md", ".markdown", ".txt", ".log", ".json", ".yaml", ".yml", ".toml", ".xml", ".html", ".htm", ".css", ".js", ".ts", ".tsx", ".jsx", ".go", ".py", ".rs", ".java", ".c", ".h", ".cpp", ".hpp", ".sh", ".sql":
		return "text/plain; charset=utf-8"
	}
	return "application/octet-stream"
}

func detectContentTypeBytes(filePath string, data []byte) string {
	if typ := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath))); typ != "" {
		return typ
	}
	if len(data) > 0 {
		return http.DetectContentType(data)
	}
	return "application/octet-stream"
}

func isImageContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), "image/")
}

func isSafeInlineContentType(contentType string) bool {
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	if lowerType == "" || lowerType == "application/octet-stream" {
		return false
	}
	if strings.HasPrefix(lowerType, "image/") {
		return !strings.Contains(lowerType, "svg")
	}
	return strings.HasPrefix(lowerType, "text/plain") ||
		strings.HasPrefix(lowerType, "application/json") ||
		strings.HasPrefix(lowerType, "application/pdf")
}

func isLikelyText(data []byte, contentType string) bool {
	lowerType := strings.ToLower(contentType)
	if strings.HasPrefix(lowerType, "text/") ||
		strings.Contains(lowerType, "json") ||
		strings.Contains(lowerType, "xml") ||
		strings.Contains(lowerType, "yaml") ||
		strings.Contains(lowerType, "toml") ||
		strings.Contains(lowerType, "javascript") ||
		strings.Contains(lowerType, "typescript") {
		return utf8.Valid(data)
	}
	if len(data) == 0 {
		return true
	}
	return utf8.Valid(data) && !bytesContainZero(data)
}

func bytesContainZero(data []byte) bool {
	for _, item := range data {
		if item == 0 {
			return true
		}
	}
	return false
}

func workspaceContentHash(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func workspaceDownloadURL(agentID string, workspacePath string) string {
	return "/x/john-ai-agentbox/api/v1/agents/" + url.PathEscape(agentID) + "/workspace/download?path=" + url.QueryEscape(workspacePath)
}

func wrapWorkspaceRuntimeUnavailable(cause error) error {
	if cause == nil {
		return bizerr.NewCode(CodeWorkspaceRuntimeUnavailable)
	}
	return bizerr.WrapCode(cause, CodeWorkspaceRuntimeUnavailable)
}

type skillMarkdownFrontMatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ParseSkillMarkdown extracts display metadata from a SKILL.md file and falls
// back to the first meaningful markdown body line when no description exists.
func ParseSkillMarkdown(raw []byte) (string, string) {
	text := string(raw)
	name := ""
	description := ""
	body := text
	if frontMatter, markdownBody, ok := splitSkillFrontMatter(text); ok {
		body = markdownBody
		metadata := skillMarkdownFrontMatter{}
		if err := yaml.Unmarshal([]byte(frontMatter), &metadata); err == nil {
			name = metadata.Name
			description = metadata.Description
		}
	}
	if strings.TrimSpace(description) == "" {
		description = firstSkillMarkdownDescription(body)
	}
	return strings.TrimSpace(name), strings.TrimSpace(description)
}

func splitSkillFrontMatter(text string) (string, string, bool) {
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return "", text, false
	}
	for index := 1; index < len(lines); index++ {
		if strings.TrimRight(lines[index], " \t") == "---" {
			return strings.Join(lines[1:index], "\n"), strings.Join(lines[index+1:], "\n"), true
		}
	}
	return "", text, false
}

func firstSkillMarkdownDescription(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			continue
		}
		return line
	}
	return ""
}

func filterSkills(items []SkillInfo, query string) []SkillInfo {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return items
	}
	filtered := make([]SkillInfo, 0, len(items))
	for _, item := range items {
		haystack := strings.ToLower(item.Name + " " + item.Description + " " + item.Path + " " + item.Source)
		if strings.Contains(haystack, query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func sortSkills(items []SkillInfo) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Scope != items[j].Scope {
			return items[i].Scope < items[j].Scope
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

func (s *serviceImpl) normalizeSkillResponsePath(scope string, workspacePath string) string {
	if scope != SkillScopeProject {
		return ""
	}
	return firstNonEmpty(strings.TrimSpace(workspacePath), s.config.WorkspaceRootPath)
}
