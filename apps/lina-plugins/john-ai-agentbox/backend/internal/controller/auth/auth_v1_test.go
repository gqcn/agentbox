// This file verifies the AgentBox auth controller at the HTTP boundary. It
// covers the independent agent_box_session cookie without using the production
// database-backed authentication store.

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	authsvc "john-ai-agentbox/backend/internal/service/auth"
)

func TestAuthControllerLoginSessionAndLogoutCookies(t *testing.T) {
	ctx := context.Background()
	service := &fakeAuthControllerService{}
	baseURL, shutdown := startAuthControllerServer(t, service)
	defer shutdown()

	loginBody, loginStatus, loginHeader := authControllerRequestJSON(t, ctx, http.MethodPost, baseURL+"/auth/sessions", map[string]string{
		"username": "admin",
		"password": "admin123",
	}, "")
	if loginStatus != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginStatus, loginBody)
	}
	if loginJSON := mustAuthControllerJSON(t, loginBody); loginJSON.Get("user.id").String() != "usr-admin" {
		t.Fatalf("login body = %s", loginBody)
	}
	loginCookie := strings.Join(loginHeader.Values("Set-Cookie"), "; ")
	if !strings.Contains(loginCookie, agentBoxSessionCookieName+"=issued-token") ||
		!strings.Contains(loginCookie, "HttpOnly") ||
		!strings.Contains(loginCookie, "SameSite=Lax") {
		t.Fatalf("login Set-Cookie = %q", loginCookie)
	}

	sessionBody, sessionStatus, _ := authControllerRequestJSON(t, ctx, http.MethodGet, baseURL+"/auth/session", nil, agentBoxSessionCookieName+"=issued-token")
	if sessionStatus != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionStatus, sessionBody)
	}
	if sessionJSON := mustAuthControllerJSON(t, sessionBody); sessionJSON.Get("user.username").String() != "admin" {
		t.Fatalf("session body = %s", sessionBody)
	}
	if service.currentToken != "issued-token" {
		t.Fatalf("current session token = %q", service.currentToken)
	}

	logoutBody, logoutStatus, logoutHeader := authControllerRequestJSON(t, ctx, http.MethodDelete, baseURL+"/auth/session", nil, agentBoxSessionCookieName+"=issued-token")
	if logoutStatus != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logoutStatus, logoutBody)
	}
	if logoutJSON := mustAuthControllerJSON(t, logoutBody); !logoutJSON.Get("loggedOut").Bool() {
		t.Fatalf("logout body = %s", logoutBody)
	}
	if service.logoutToken != "issued-token" {
		t.Fatalf("logout token = %q", service.logoutToken)
	}
	clearCookie := strings.Join(logoutHeader.Values("Set-Cookie"), "; ")
	if !strings.Contains(clearCookie, agentBoxSessionCookieName+"=") ||
		!strings.Contains(clearCookie, "HttpOnly") ||
		!strings.Contains(clearCookie, "SameSite=Lax") {
		t.Fatalf("logout Set-Cookie = %q", clearCookie)
	}
}

func TestAuthControllerDoesNotSetCookieOnLoginError(t *testing.T) {
	ctx := context.Background()
	service := &fakeAuthControllerService{loginErr: errors.New("invalid credentials")}
	baseURL, shutdown := startAuthControllerServer(t, service)
	defer shutdown()

	_, _, header := authControllerRequestJSON(t, ctx, http.MethodPost, baseURL+"/auth/sessions", map[string]string{
		"username": "admin",
		"password": "wrong",
	}, "")
	if cookies := strings.Join(header.Values("Set-Cookie"), "; "); strings.Contains(cookies, agentBoxSessionCookieName+"=") {
		t.Fatalf("invalid login set cookie = %q", cookies)
	}
}

func startAuthControllerServer(t *testing.T, service authsvc.Service) (string, func()) {
	t.Helper()
	server := g.Server(guid.S())
	server.SetAddr(ghttp.FreePortAddress)
	server.SetLogger(nil)
	server.SetDumpRouterMap(false)
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(authControllerTestResponse)
		group.Bind(NewV1(service))
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

func authControllerTestResponse(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}
	if err := r.GetError(); err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJson(map[string]string{"error": err.Error()})
		return
	}
	r.Response.WriteJson(r.GetHandlerResponse())
}

func authControllerRequestJSON(t *testing.T, ctx context.Context, method string, url string, payload any, cookie string) ([]byte, int, http.Header) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return raw, response.StatusCode, response.Header.Clone()
}

func mustAuthControllerJSON(t *testing.T, body []byte) *gjson.Json {
	t.Helper()
	value, err := gjson.DecodeToJson(body)
	if err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, body)
	}
	return value
}

type fakeAuthControllerService struct {
	loginErr     error
	currentToken string
	logoutToken  string
}

func (s *fakeAuthControllerService) Login(_ context.Context, in authsvc.LoginInput) (*authsvc.SessionOutput, error) {
	if s.loginErr != nil {
		return nil, s.loginErr
	}
	if in.Username != "admin" || in.Password != "admin123" {
		return nil, errors.New("invalid credentials")
	}
	return &authsvc.SessionOutput{
		User: &authsvc.UserOutput{
			ID:          "usr-admin",
			Username:    "admin",
			DisplayName: "admin",
			LastLoginAt: time.Now().UnixMilli(),
		},
		Token:     "issued-token",
		ExpiresAt: time.Now().Add(time.Hour).UnixMilli(),
	}, nil
}

func (s *fakeAuthControllerService) CurrentSession(_ context.Context, token string) (*authsvc.SessionOutput, error) {
	s.currentToken = token
	if token != "issued-token" {
		return nil, errors.New("authentication required")
	}
	return &authsvc.SessionOutput{
		User: &authsvc.UserOutput{
			ID:          "usr-admin",
			Username:    "admin",
			DisplayName: "admin",
			LastLoginAt: time.Now().UnixMilli(),
		},
		ExpiresAt: time.Now().Add(time.Hour).UnixMilli(),
	}, nil
}

func (s *fakeAuthControllerService) Logout(_ context.Context, token string) (*authsvc.LogoutOutput, error) {
	s.logoutToken = token
	return &authsvc.LogoutOutput{LoggedOut: true}, nil
}
