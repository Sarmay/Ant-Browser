package launchcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ant-chrome/backend/internal/automation"
	"ant-chrome/backend/internal/browser"
)

type fakeAutomationStarter struct {
	mu      sync.Mutex
	runs    map[string]automation.ScriptRunRecord
	scripts map[string]automation.ScriptRecord
}

func (f *fakeAutomationStarter) StartInstance(string) (*browser.Profile, error) {
	return nil, http.ErrNotSupported
}

func (f *fakeAutomationStarter) AutomationScriptList() ([]automation.ScriptRecord, error) {
	return nil, nil
}

func (f *fakeAutomationStarter) AutomationScriptGet(scriptID string) (*automation.ScriptRecord, error) {
	item, ok := f.scripts[scriptID]
	if !ok {
		return nil, osNotExist("script not found")
	}
	return &item, nil
}

func (f *fakeAutomationStarter) AutomationScriptRunWithOptions(input automation.ScriptRunRequest) (*automation.ScriptRunRecord, error) {
	record := automation.ScriptRunRecord{
		ID:         "sync-1",
		ScriptID:   input.ScriptID,
		Status:     "success",
		Summary:    "done",
		StartedAt:  time.Now().Format(time.RFC3339),
		FinishedAt: time.Now().Format(time.RFC3339),
	}
	return &record, nil
}

func (f *fakeAutomationStarter) AutomationScriptRunList(int) ([]automation.ScriptRunRecord, error) {
	return nil, nil
}

func (f *fakeAutomationStarter) AutomationScriptRunGet(runID string) (*automation.ScriptRunRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	record, ok := f.runs[runID]
	if !ok {
		return nil, osNotExist("run not found")
	}
	copy := record
	return &copy, nil
}

func (f *fakeAutomationStarter) AutomationScriptStartRun(input automation.ScriptRunRequest) (*automation.ScriptRunRecord, error) {
	if _, ok := f.scripts[input.ScriptID]; !ok {
		return nil, osNotExist("script not found")
	}
	record := automation.ScriptRunRecord{
		ID:         "async-1",
		ScriptID:   input.ScriptID,
		ScriptName: "demo",
		Status:     "running",
		Summary:    "脚本执行中",
		StartedAt:  time.Now().Format(time.RFC3339),
	}
	f.mu.Lock()
	if f.runs == nil {
		f.runs = map[string]automation.ScriptRunRecord{}
	}
	f.runs[record.ID] = record
	f.mu.Unlock()
	return &record, nil
}

type osNotExist string

func (e osNotExist) Error() string { return string(e) }
func (e osNotExist) Is(target error) bool {
	return target != nil && target.Error() == "script not found" || false
}

func TestAutomationScriptRunAsyncAccepted(t *testing.T) {
	starter := &fakeAutomationStarter{
		scripts: map[string]automation.ScriptRecord{
			"news-query-txt": {ID: "news-query-txt", Name: "demo", Type: "playwright-cdp"},
		},
	}
	server := NewLaunchServer(nil, starter, nil, 19876)
	handler := NewTestHandler(server)

	request := httptest.NewRequest(http.MethodPost, "/api/automation/scripts/run", strings.NewReader(`{
		"scriptId": "news-query-txt",
		"async": true
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202; body=%s", response.Code, response.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/automation/scripts/runs/async-1", nil)
	getRes := httptest.NewRecorder()
	handler.ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("get run status = %d; body=%s", getRes.Code, getRes.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(getRes.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["ok"] != true {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestParseAutomationScriptRunPathID(t *testing.T) {
	runID, ok := parseAutomationScriptRunPathID("/api/automation/scripts/runs/async-1")
	if !ok || runID != "async-1" {
		t.Fatalf("runID = %q ok=%v", runID, ok)
	}
}
