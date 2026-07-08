package backend

import (
	"ant-chrome/backend/internal/browser"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureFingerprintCheckPageURL(t *testing.T) {
	app := NewApp(t.TempDir())
	app.browserMgr = browser.NewManager(nil, app.appRoot)
	app.browserMgr.Profiles["profile-123"] = &browser.Profile{
		ProfileId: "profile-123",
		FingerprintArgs: []string{
			"--lang=ja-JP",
			"--timezone=Asia/Tokyo",
			"--fingerprint-hardware-concurrency=8",
		},
	}
	pageURL, err := app.ensureFingerprintCheckPageURL("profile-123")
	if err != nil {
		t.Fatalf("ensureFingerprintCheckPageURL error = %v", err)
	}
	parsed, err := url.Parse(pageURL)
	if err != nil {
		t.Fatalf("parse page url error = %v", err)
	}
	if parsed.Scheme != "file" {
		t.Fatalf("scheme = %q", parsed.Scheme)
	}
	if got := parsed.Query().Get("profileId"); got != "profile-123" {
		t.Fatalf("profileId query = %q", got)
	}
	pagePath := parsed.Path
	if runtime.GOOS == "windows" && len(pagePath) >= 3 && pagePath[0] == '/' && pagePath[2] == ':' {
		pagePath = strings.TrimPrefix(pagePath, "/")
	}
	content, err := os.ReadFile(filepath.FromSlash(pagePath))
	if err != nil {
		t.Fatalf("read generated page error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Ant 指纹检测") || !strings.Contains(text, "Canvas Hash") || !strings.Contains(text, "WebRTC") {
		t.Fatalf("generated page missing expected content")
	}
	if !strings.Contains(text, "生效判定") || !strings.Contains(text, "\"language\": \"ja-JP\"") || !strings.Contains(text, "\"timezone\": \"Asia/Tokyo\"") {
		t.Fatalf("generated page missing embedded expected context")
	}
}
