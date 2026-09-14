package launchcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
)

func newProfileCopyTestHandler(t *testing.T) (*LaunchServer, *browser.Manager, http.Handler) {
	t.Helper()

	cfg := &config.Config{}
	cfg.Browser.DefaultFingerprintArgs = []string{"--fingerprint=default-seed"}
	manager := browser.NewManager(cfg, t.TempDir())
	service := NewLaunchCodeService(NewMemoryLaunchCodeDAO())
	manager.CodeProvider = service
	server := NewLaunchServer(service, nil, manager, 19876)
	return server, manager, NewTestHandler(server)
}

func createCopySourceProfile(t *testing.T, manager *browser.Manager) *browser.Profile {
	t.Helper()

	profile, err := manager.Create(browser.ProfileInput{
		ProfileName:     "main",
		FingerprintArgs: []string{"--fingerprint=111", "--lang=en-US"},
		LaunchArgs:      []string{"--no-first-run"},
		Tags:            []string{"电商"},
		Keywords:        []string{"main"},
		GroupId:         "group-sales",
	})
	if err != nil {
		t.Fatalf("Create source profile: %v", err)
	}
	return profile
}

func decodeCopyResponse(t *testing.T, response *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var payload map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, response.Body.String())
	}
	return payload
}

func TestParseProfilePathAcceptsCopyAction(t *testing.T) {
	profileID, action, ok := parseProfilePath("/api/profiles/abc-123/copy")
	if !ok {
		t.Fatal("parseProfilePath rejected copy path")
	}
	if profileID != "abc-123" {
		t.Fatalf("profileID = %q, want abc-123", profileID)
	}
	if action != "copy" {
		t.Fatalf("action = %q, want copy", action)
	}
}

func TestCopyProfileEmptyBodyCreatesNewInstance(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	payload := decodeCopyResponse(t, response)
	if payload["ok"] != true {
		t.Fatalf("ok = %#v, want true", payload["ok"])
	}
	if payload["copied"] != true {
		t.Fatalf("copied = %#v, want true", payload["copied"])
	}
	if payload["created"] != true {
		t.Fatalf("created = %#v, want true", payload["created"])
	}
	if payload["sourceProfileId"] != source.ProfileId {
		t.Fatalf("sourceProfileId = %#v, want %s", payload["sourceProfileId"], source.ProfileId)
	}

	profileID, _ := payload["profileId"].(string)
	if strings.TrimSpace(profileID) == "" || profileID == source.ProfileId {
		t.Fatalf("profileId = %q, want a new id", profileID)
	}
	profileName, _ := payload["profileName"].(string)
	if !strings.Contains(profileName, "副本") {
		t.Fatalf("profileName = %q, want generated copy name", profileName)
	}

	copied, _, errMsg := NewLaunchServer(nil, nil, manager, 19876).profileSnapshotByID(profileID)
	if errMsg != "" {
		t.Fatalf("load copied profile: %s", errMsg)
	}
	if copied.UserDataDir != copied.ProfileId {
		t.Fatalf("userDataDir = %q, want new profileId", copied.UserDataDir)
	}
	if copied.Running {
		t.Fatal("copied profile should not be running")
	}
	if got := strings.Join(copied.Tags, ","); got != "电商" {
		t.Fatalf("tags = %q, want 电商", got)
	}
	if got := strings.Join(copied.LaunchArgs, ","); got != "--no-first-run" {
		t.Fatalf("launchArgs = %q, want --no-first-run", got)
	}
	if copied.GroupId != "group-sales" {
		t.Fatalf("groupId = %q, want group-sales", copied.GroupId)
	}
}

func TestCopyProfileAutoFingerprintRegeneratesSeed(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"profileName": "task-001",
		"mode": "auto_fingerprint"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	payload := decodeCopyResponse(t, response)
	if payload["profileName"] != "task-001" {
		t.Fatalf("profileName = %#v, want task-001", payload["profileName"])
	}

	profileMap, _ := payload["profile"].(map[string]interface{})
	args := jsonStringSlice(profileMap["fingerprintArgs"])
	if containsString(args, "--fingerprint=111") {
		t.Fatalf("fingerprintArgs = %#v, still contains source seed", args)
	}
	if !containsString(args, "--fingerprint=default-seed") {
		t.Fatalf("fingerprintArgs = %#v, missing default seed", args)
	}
	if !containsString(args, "--lang=en-US") {
		t.Fatalf("fingerprintArgs = %#v, missing preserved lang", args)
	}
}

func TestCopyProfileRegularKeepsFingerprintArgs(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"profileName": "same-fp",
		"mode": "regular"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	payload := decodeCopyResponse(t, response)
	profileMap, _ := payload["profile"].(map[string]interface{})
	args := jsonStringSlice(profileMap["fingerprintArgs"])
	if !containsString(args, "--fingerprint=111") {
		t.Fatalf("fingerprintArgs = %#v, want source seed", args)
	}
	if containsString(args, "--fingerprint=default-seed") {
		t.Fatalf("fingerprintArgs = %#v, should keep source seed", args)
	}
}

func TestCopyProfileCustomLaunchCode(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"profileName": "coded",
		"launchCode": "TASK_001"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	payload := decodeCopyResponse(t, response)
	if payload["launchCode"] != "TASK_001" {
		t.Fatalf("launchCode = %#v, want TASK_001", payload["launchCode"])
	}
}

func TestCopyProfileLaunchCodeConflict(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	first := createCopySourceProfile(t, manager)
	second, err := manager.Create(browser.ProfileInput{ProfileName: "other"})
	if err != nil {
		t.Fatalf("Create second profile: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+second.ProfileId+"/copy", strings.NewReader(`{
		"launchCode": "`+first.LaunchCode+`"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
}

func TestCopyProfileRejectsUnknownMode(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"mode": "defaults"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	payload := decodeCopyResponse(t, response)
	if payload["error"] != "mode must be auto_fingerprint or regular" {
		t.Fatalf("error = %#v", payload["error"])
	}
}

func TestCopyProfileRejectsUnknownAutomationTarget(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"automationTargets": ["not-a-target"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestCopyProfileNotFound(t *testing.T) {
	_, _, handler := newProfileCopyTestHandler(t)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/missing-id/copy", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestCopyProfileMethodNotAllowed(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodGet, "/api/profiles/"+source.ProfileId+"/copy", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusMethodNotAllowed, response.Body.String())
	}
}

func TestCopyProfileRejectsUnknownFields(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"newName": "task-001"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestCopyProfileBySelectorCode(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/copy", strings.NewReader(`{
		"selector": { "code": "`+source.LaunchCode+`" },
		"profileName": "from-code"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	payload := decodeCopyResponse(t, response)
	if payload["profileName"] != "from-code" {
		t.Fatalf("profileName = %#v", payload["profileName"])
	}
	if payload["sourceProfileId"] != source.ProfileId {
		t.Fatalf("sourceProfileId = %#v", payload["sourceProfileId"])
	}
}

func TestCopyProfileCustomProxy(t *testing.T) {
	_, manager, handler := newProfileCopyTestHandler(t)
	source := createCopySourceProfile(t, manager)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/"+source.ProfileId+"/copy", strings.NewReader(`{
		"profileName": "with-proxy",
		"proxyConfig": "http://127.0.0.1:18080"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	payload := decodeCopyResponse(t, response)
	profileMap, _ := payload["profile"].(map[string]interface{})
	if profileMap["proxyConfig"] != "http://127.0.0.1:18080" {
		t.Fatalf("proxyConfig = %#v", profileMap["proxyConfig"])
	}
}

func TestCopyProfileSelectorRequired(t *testing.T) {
	_, _, handler := newProfileCopyTestHandler(t)
	request := httptest.NewRequest(http.MethodPost, "/api/profiles/copy", strings.NewReader(`{
		"profileName": "missing-source"
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestCopyProfileCatalogUnavailable(t *testing.T) {
	server := NewLaunchServer(nil, nil, nil, 19876)
	handler := NewTestHandler(server)

	request := httptest.NewRequest(http.MethodPost, "/api/profiles/any-id/copy", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusServiceUnavailable, response.Body.String())
	}
}

func jsonStringSlice(value interface{}) []string {
	raw, ok := value.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		text, _ := item.(string)
		out = append(out, text)
	}
	return out
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
