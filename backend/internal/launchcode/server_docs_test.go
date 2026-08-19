package launchcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func newDocsTestHandler(t *testing.T) (*LaunchServer, http.Handler) {
	t.Helper()

	server := NewLaunchServer(nil, nil, nil, 19876)
	server.SetDocsFS(fstest.MapFS{
		"docs/docs/index.html": {Data: []byte("<!doctype html><title>Ant Browser Docs</title>")},
		"docs/assets/app.js":   {Data: []byte("document.body.dataset.ready = 'true'")},
	})
	return server, NewTestHandler(server)
}

func TestDocsRoutesServeStaticFrontend(t *testing.T) {
	_, handler := newDocsTestHandler(t)

	tests := []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/docs/", contentType: "text/html", body: "Ant Browser Docs"},
		{path: "/docs/assets/app.js", contentType: "javascript", body: "dataset.ready"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); !strings.Contains(got, tt.contentType) {
				t.Fatalf("Content-Type = %q, want substring %q", got, tt.contentType)
			}
			if !strings.Contains(response.Body.String(), tt.body) {
				t.Fatalf("body = %q, want substring %q", response.Body.String(), tt.body)
			}
			if got := response.Header().Get("Content-Security-Policy"); got != docsContentSecurityPolicy {
				t.Fatalf("Content-Security-Policy = %q, want %q", got, docsContentSecurityPolicy)
			}
			if got := response.Header().Get("X-Frame-Options"); got != "DENY" {
				t.Fatalf("X-Frame-Options = %q, want DENY", got)
			}
		})
	}
}

func TestDocsRedirectsPreserveQuery(t *testing.T) {
	_, handler := newDocsTestHandler(t)

	for _, path := range []string{"/docs?doc=api-health", "/system/api/docs?doc=api-health"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusTemporaryRedirect {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusTemporaryRedirect)
			}
			if got := response.Header().Get("Location"); got != "/docs/?doc=api-health" {
				t.Fatalf("Location = %q, want /docs/?doc=api-health", got)
			}
		})
	}
}

func TestDocsContextDoesNotExposeAPIKey(t *testing.T) {
	server, handler := newDocsTestHandler(t)
	server.SetAPIAuthConfig(APIAuthConfig{
		Enabled: true,
		APIKey:  "secret-value",
		Header:  "X-Custom-Key",
	})

	request := httptest.NewRequest(http.MethodGet, "/docs/context.json", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "secret-value") {
		t.Fatal("docs context exposed the configured API key")
	}

	var payload struct {
		BaseURL    string `json:"baseUrl"`
		AuthHeader string `json:"authHeader"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.BaseURL != "http://127.0.0.1:19876" {
		t.Fatalf("baseUrl = %q, want http://127.0.0.1:19876", payload.BaseURL)
	}
	if payload.AuthHeader != "X-Custom-Key" {
		t.Fatalf("authHeader = %q, want X-Custom-Key", payload.AuthHeader)
	}
}

func TestDocsRoutesRejectWrites(t *testing.T) {
	_, handler := newDocsTestHandler(t)

	for _, path := range []string{"/docs", "/system/api/docs", "/docs/context.json"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, path, strings.NewReader("ignored"))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
			}
			if got := response.Header().Get("Allow"); got != "GET, HEAD" {
				t.Fatalf("Allow = %q, want GET, HEAD", got)
			}
		})
	}
}

func TestDocsRouteDoesNotFallThroughToCDPProxy(t *testing.T) {
	server := NewLaunchServer(nil, nil, nil, 19876)
	handler := NewTestHandler(server)

	request := httptest.NewRequest(http.MethodGet, "/system/api/docs", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if strings.Contains(response.Body.String(), "no active browser debug target") {
		t.Fatal("docs route fell through to the CDP proxy")
	}
}
