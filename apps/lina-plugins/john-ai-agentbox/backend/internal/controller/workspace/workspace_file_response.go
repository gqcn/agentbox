// This file centralizes workspace binary response writing so handlers can stay
// focused on authentication and service delegation.

package workspace

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/closeutil"

	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// writeWorkspaceFileStream copies one service-owned workspace stream into the HTTP response.
func writeWorkspaceFileStream(ctx context.Context, item *workspacesvc.FileStream) (err error) {
	if item == nil || item.Reader == nil {
		return nil
	}
	defer closeutil.Close(ctx, item.Reader, &err, "close agentbox workspace file stream failed")
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return nil
	}

	contentType := strings.TrimSpace(item.File.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	request.Response.Header().Set("Content-Type", contentType)
	request.Response.Header().Set("X-Content-Type-Options", "nosniff")
	if item.File.Size > 0 {
		request.Response.Header().Set("Content-Length", strconv.FormatInt(item.File.Size, 10))
	}
	disposition := strings.TrimSpace(item.Disposition)
	if disposition == "" {
		disposition = workspacesvc.ResourceDispositionInline
	}
	request.Response.Header().Set("Content-Disposition", disposition+"; filename=\""+sanitizeResponseFilename(item.File.Name)+"\"")

	if err = writeResponseBody(request, item.Reader); err != nil {
		return err
	}
	request.ExitAll()
	return nil
}

// writeWorkspaceHTMLPreviewStream renders one workspace HTML file in an
// isolated document context. The CSP sandbox intentionally prevents scripts,
// forms, same-origin access, and object embeds; richer live service previews
// belong to the service proxy/tunnel runtime.
func writeWorkspaceHTMLPreviewStream(ctx context.Context, item *workspacesvc.FileStream) (err error) {
	if item == nil || item.Reader == nil {
		return nil
	}
	defer closeutil.Close(ctx, item.Reader, &err, "close agentbox workspace html preview stream failed")
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return nil
	}

	request.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	request.Response.Header().Set("X-Content-Type-Options", "nosniff")
	request.Response.Header().Set("Content-Disposition", "inline; filename=\""+sanitizeResponseFilename(item.File.Name)+"\"")
	request.Response.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; img-src data: blob:; media-src data: blob:; style-src 'unsafe-inline'; font-src data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; object-src 'none'")
	request.Response.Header().Set("Referrer-Policy", "no-referrer")
	request.Response.Header().Set("Cache-Control", "no-store")
	if item.File.Size > 0 {
		request.Response.Header().Set("Content-Length", strconv.FormatInt(item.File.Size, 10))
	}

	if err = writeResponseBody(request, item.Reader); err != nil {
		return err
	}
	request.ExitAll()
	return nil
}

// writeResponseBody copies the stream through GoFrame's raw response writer.
func writeResponseBody(request *ghttp.Request, reader io.Reader) error {
	_, err := io.Copy(request.Response.RawWriter(), reader)
	return err
}

// sanitizeResponseFilename keeps Content-Disposition filenames header-safe.
func sanitizeResponseFilename(filename string) string {
	value := strings.TrimSpace(filename)
	if value == "" {
		value = "download"
	}
	replacer := strings.NewReplacer("\\", "_", "\"", "_", "\r", "_", "\n", "_")
	return replacer.Replace(value)
}
