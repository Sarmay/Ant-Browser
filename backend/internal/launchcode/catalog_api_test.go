package launchcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ant-chrome/backend/internal/config"
)

func TestCatalogEndpointsListLocalResources(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	manager.Config.Browser.Proxies = []config.BrowserProxy{
		{ProxyId: "proxy-us", ProxyName: "US", ProxyConfig: "http://127.0.0.1:7890", GroupName: "海外"},
	}
	manager.Config.Browser.Cores = []config.BrowserCore{
		{CoreId: "core-1", CoreName: "Chrome", IsDefault: true},
	}

	tests := []struct {
		path string
		id   string
	}{
		{path: "/api/proxies", id: "proxy-us"},
		{path: "/api/cores", id: "core-1"},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodGet, tt.path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body=%s", tt.path, response.Code, response.Body.String())
		}
		var payload struct {
			OK    bool                     `json:"ok"`
			Count int                      `json:"count"`
			Items []map[string]interface{} `json:"items"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode %s: %v", tt.path, err)
		}
		if !payload.OK || payload.Count < 1 {
			t.Fatalf("%s payload = %#v", tt.path, payload)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/groups", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("groups status = %d, body=%s", response.Code, response.Body.String())
	}
}

func TestParseNetscapeCookies(t *testing.T) {
	raw := "# Netscape HTTP Cookie File\n\n.example.com\tTRUE\t/\tTRUE\t0\tsession\tabc123\n"
	cookies, err := parseNetscapeCookies(raw)
	if err != nil {
		t.Fatalf("parseNetscapeCookies: %v", err)
	}
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "abc123" {
		t.Fatalf("cookies = %#v", cookies)
	}
	if !cookies[0].Secure {
		t.Fatal("expected secure cookie")
	}
}
