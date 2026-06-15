// This file verifies AgentBox protected-route authentication middleware. Tests
// use a fake auth service and a local GoFrame server so request context and
// middleware chaining are exercised through the real HTTP path.

package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/pkg/bizerr"

	authsvc "john-ai-agentbox/backend/internal/service/auth"
	"john-ai-agentbox/backend/internal/service/authctx"
)

// TestAuthRejectsUnauthenticatedRequest verifies missing sessions stop the handler chain.
func TestAuthRejectsUnauthenticatedRequest(t *testing.T) {
	var handlerHits atomic.Int32
	baseURL, shutdown := startAuthMiddlewareServer(t, &fakeAuthService{}, &handlerHits)
	defer shutdown()

	body, status := authMiddlewareGet(t, baseURL+"/protected", "")
	if status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d body=%s", status, body)
	}
	if handlerHits.Load() != 0 {
		t.Fatalf("protected handler executed %d times", handlerHits.Load())
	}
}

// TestAuthPropagatesAuthenticatedUser verifies authenticated users are stored in context.
func TestAuthPropagatesAuthenticatedUser(t *testing.T) {
	var handlerHits atomic.Int32
	baseURL, shutdown := startAuthMiddlewareServer(t, &fakeAuthService{validToken: "valid-token"}, &handlerHits)
	defer shutdown()

	body, status := authMiddlewareGet(t, baseURL+"/protected", "agent_box_session=valid-token")
	if status != http.StatusOK {
		t.Fatalf("expected ok status, got %d body=%s", status, body)
	}
	json, err := gjson.DecodeToJson(body)
	if err != nil {
		t.Fatalf("decode response body: %v body=%s", err, body)
	}
	if json.Get("userId").String() != "usr-test" {
		t.Fatalf("unexpected body %s", body)
	}
	if handlerHits.Load() != 1 {
		t.Fatalf("expected protected handler once, got %d", handlerHits.Load())
	}
}

// TestAuthRejectsMissingService verifies nil auth service returns a structured auth error.
func TestAuthRejectsMissingService(t *testing.T) {
	var handlerHits atomic.Int32
	baseURL, shutdown := startAuthMiddlewareServer(t, nil, &handlerHits)
	defer shutdown()

	body, status := authMiddlewareGet(t, baseURL+"/protected", "agent_box_session=valid-token")
	if status != http.StatusInternalServerError {
		t.Fatalf("expected internal error status, got %d body=%s", status, body)
	}
	if handlerHits.Load() != 0 {
		t.Fatalf("protected handler executed %d times", handlerHits.Load())
	}
}

func startAuthMiddlewareServer(t *testing.T, service authsvc.Service, handlerHits *atomic.Int32) (string, func()) {
	t.Helper()
	server := g.Server(guid.S())
	server.SetAddr(ghttp.FreePortAddress)
	server.SetLogger(nil)
	server.SetDumpRouterMap(false)
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(simpleErrorResponse, Auth(service))
		group.GET("/protected", func(r *ghttp.Request) {
			userID, ok := authctx.UserID(r.Context())
			if !ok {
				r.SetError(bizerr.NewCode(authsvc.CodeAuthRequired))
				return
			}
			handlerHits.Add(1)
			r.Response.WriteJsonExit(map[string]string{"userId": userID})
		})
	})
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			if err := server.Shutdown(); err != nil {
				t.Errorf("shutdown server: %v", err)
			}
		})
	}
	t.Cleanup(shutdown)
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("http://127.0.0.1:%d", server.GetListenedPort()), shutdown
}

func simpleErrorResponse(r *ghttp.Request) {
	r.Middleware.Next()
	if err := r.GetError(); err != nil {
		status := http.StatusInternalServerError
		if bizerr.Is(err, authsvc.CodeAuthRequired) {
			status = http.StatusUnauthorized
		}
		r.Response.WriteStatus(status)
	}
}

func authMiddlewareGet(t *testing.T, url string, cookie string) ([]byte, int) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body, response.StatusCode
}

type fakeAuthService struct {
	validToken string
}

func (s *fakeAuthService) Login(context.Context, authsvc.LoginInput) (*authsvc.SessionOutput, error) {
	return nil, bizerr.NewCode(authsvc.CodeAuthInvalidCredentials)
}

func (s *fakeAuthService) CurrentSession(_ context.Context, token string) (*authsvc.SessionOutput, error) {
	if token == "" || token != s.validToken {
		return nil, bizerr.NewCode(authsvc.CodeAuthRequired)
	}
	return &authsvc.SessionOutput{
		User: &authsvc.UserOutput{
			ID:          "usr-test",
			Username:    "tester",
			DisplayName: "tester",
		},
		ExpiresAt: 1718000000000,
	}, nil
}

func (s *fakeAuthService) Logout(context.Context, string) (*authsvc.LogoutOutput, error) {
	return &authsvc.LogoutOutput{LoggedOut: true}, nil
}
