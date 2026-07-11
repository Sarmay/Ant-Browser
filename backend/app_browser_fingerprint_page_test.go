package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"encoding/json"
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
	if !strings.Contains(text, "Ant 指纹检测") || !strings.Contains(text, "指纹项") || !strings.Contains(text, "期望来源") || !strings.Contains(text, "结果") {
		t.Fatalf("generated page missing expected content")
	}
	if strings.Contains(text, "是否命中") {
		t.Fatalf("generated page still contains old ambiguous result header")
	}
	if !strings.Contains(text, "指纹对比") || !strings.Contains(text, "\"language\": \"ja-JP\"") || !strings.Contains(text, "\"acceptLanguage\": \"ja-JP,ja\"") || !strings.Contains(text, "\"timezone\": \"Asia/Tokyo\"") {
		t.Fatalf("generated page missing embedded expected context")
	}
	if !strings.Contains(text, "outerWidth") || !strings.Contains(text, "比对 window.outerWidth/outerHeight") {
		t.Fatalf("generated page missing window outer size comparison")
	}
	if strings.Contains(text, "比对 window.innerWidth/innerHeight") {
		t.Fatalf("generated page still compares window inner size for --window-size")
	}
	if !strings.Contains(text, "formatLocalDateTime") || !strings.Contains(text, "检测时间：本地") || !strings.Contains(text, "UTC ") {
		t.Fatalf("generated page missing local and UTC time display")
	}
	if !strings.Contains(text, "FINGERPRINT_AUTO_REFRESH_MS = 60 * 60 * 1000") || !strings.Contains(text, "maybeAutoRefresh") || !strings.Contains(text, "setInterval(maybeAutoRefresh") {
		t.Fatalf("generated page missing one-hour auto refresh")
	}
}

func TestFingerprintCheckPageBuildsRuntimeBaseline(t *testing.T) {
	app := NewApp(t.TempDir())
	app.browserMgr = browser.NewManager(nil, app.appRoot)
	app.browserMgr.Profiles["profile-seed"] = &browser.Profile{
		ProfileId:       "profile-seed",
		FingerprintArgs: []string{"--fingerprint=676448312360042767", "--fingerprint-brand=Chrome", "--fingerprint-platform=windows"},
	}

	pageURL, err := app.ensureFingerprintCheckPageURL("profile-seed")
	if err != nil {
		t.Fatalf("ensureFingerprintCheckPageURL error = %v", err)
	}
	parsed, err := url.Parse(pageURL)
	if err != nil {
		t.Fatalf("parse page url error = %v", err)
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
	if !strings.Contains(text, "baselineStorageKey") || !strings.Contains(text, "latestBaselineCreated") {
		t.Fatalf("generated page missing runtime baseline storage")
	}
	if !strings.Contains(text, "resetBaselineBtn") || !strings.Contains(text, "resetFingerprintBaseline") || !strings.Contains(text, "重建基线") {
		t.Fatalf("generated page missing runtime baseline reset action")
	}
	if !strings.Contains(text, "首次采集并保存为运行基线") || !strings.Contains(text, "检测页已用首次实际采集值建立运行基线") {
		t.Fatalf("generated page missing runtime baseline marker")
	}
	if !strings.Contains(text, "基线一致") || !strings.Contains(text, "基线变化") || !strings.Contains(text, "不代表配置未生效") {
		t.Fatalf("generated page missing runtime baseline change labels")
	}
	if !strings.Contains(text, "效果观测") || !strings.Contains(text, "观测变化") || !strings.Contains(text, "不等于配置失败") {
		t.Fatalf("generated page missing fingerprint effect observation labels")
	}
	if !strings.Contains(text, "未建立观测基线") || !strings.Contains(text, "改过指纹配置或 Seed 后，先重建基线再刷新验证稳定性") {
		t.Fatalf("generated page missing effect observation baseline guidance")
	}
	if !strings.Contains(text, "配置期望") || !strings.Contains(text, "运行基线") || !strings.Contains(text, "已建基线") {
		t.Fatalf("generated page missing expected source labels")
	}
	if !strings.Contains(text, "实测无效") || !strings.Contains(text, "不可配置") || !strings.Contains(text, "不作为期望") {
		t.Fatalf("generated page missing unsupported fingerprint labels")
	}
	if !strings.Contains(text, "已配置") || !strings.Contains(text, "配置下发") || !strings.Contains(text, "效果看 Canvas Hash") || !strings.Contains(text, "效果看 ClientRects Hash") {
		t.Fatalf("generated page missing unreadable launch switch labels")
	}
	if strings.Contains(text, "无期望") {
		t.Fatalf("generated page still contains misleading unknown expectation text")
	}
}

func TestFingerprintCheckPageContextUsesAdaptedArgs(t *testing.T) {
	appRoot := t.TempDir()
	app := NewApp(appRoot)
	app.browserMgr = browser.NewManager(app.config, app.appRoot)
	coreDir := filepath.Join(appRoot, "core-144")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir core dir error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "manifest.json"), []byte(`{"version":"144.0.7559.132"}`), 0o644); err != nil {
		t.Fatalf("write manifest error = %v", err)
	}
	app.config = &config.Config{}
	app.browserMgr.Config = app.config
	app.config.Browser.Cores = []browser.Core{{CoreId: "core-144", CoreName: "Chrome 144", CorePath: coreDir, IsDefault: true}}
	app.browserMgr.Profiles["profile-144"] = &browser.Profile{
		ProfileId: "profile-144",
		CoreId:    "core-144",
		FingerprintArgs: []string{
			"--fingerprint=123",
			"--fingerprint-gpu-vendor=NVIDIA",
			"--disable-gpu-fingerprint",
		},
	}

	data, err := app.buildFingerprintCheckPageContext("profile-144")
	if err != nil {
		t.Fatalf("buildFingerprintCheckPageContext error = %v", err)
	}
	var context fingerprintCheckPageContext
	if err := json.Unmarshal(data, &context); err != nil {
		t.Fatalf("unmarshal context error = %v", err)
	}
	if context.Expected.Seed != "123" {
		t.Fatalf("seed = %q, want 123", context.Expected.Seed)
	}
	if context.Expected.DisableSpoofing != "gpu" {
		t.Fatalf("disableSpoofing = %q, want gpu", context.Expected.DisableSpoofing)
	}
}

func TestFingerprintCheckPageContextUsesLaunchArgsAsExpected(t *testing.T) {
	app := NewApp(t.TempDir())
	app.browserMgr = browser.NewManager(nil, app.appRoot)
	app.browserMgr.Profiles["profile-launch-args"] = &browser.Profile{
		ProfileId: "profile-launch-args",
		FingerprintArgs: []string{
			"--fingerprint-brand=Chrome",
			"--fingerprint-platform=windows",
		},
		LaunchArgs: []string{
			"--lang=zh-CN",
			"--timezone=Etc/GMT-8",
			"--fingerprint-hardware-concurrency=26",
			"--window-size=1689,1243",
			"--disable-non-proxied-udp",
		},
	}

	data, err := app.buildFingerprintCheckPageContext("profile-launch-args")
	if err != nil {
		t.Fatalf("buildFingerprintCheckPageContext error = %v", err)
	}
	var context fingerprintCheckPageContext
	if err := json.Unmarshal(data, &context); err != nil {
		t.Fatalf("unmarshal context error = %v", err)
	}
	if context.Expected.Language != "zh-CN" {
		t.Fatalf("language = %q, want zh-CN", context.Expected.Language)
	}
	if context.Expected.AcceptLanguage != "zh-CN,zh" {
		t.Fatalf("acceptLanguage = %q, want zh-CN,zh", context.Expected.AcceptLanguage)
	}
	if context.Expected.Timezone != "Etc/GMT-8" {
		t.Fatalf("timezone = %q, want Etc/GMT-8", context.Expected.Timezone)
	}
	if context.Expected.HardwareConcurrency != "26" {
		t.Fatalf("hardwareConcurrency = %q, want 26", context.Expected.HardwareConcurrency)
	}
	if context.Expected.WindowSize != "1689,1243" {
		t.Fatalf("windowSize = %q, want 1689,1243", context.Expected.WindowSize)
	}
	if context.Expected.WebRTCPolicy != "disable_non_proxied_udp" {
		t.Fatalf("webrtcPolicy = %q, want disable_non_proxied_udp", context.Expected.WebRTCPolicy)
	}
}

func TestFingerprintCheckPageContextIgnoresLastLaunchArgsWhenStopped(t *testing.T) {
	app := NewApp(t.TempDir())
	app.browserMgr = browser.NewManager(nil, app.appRoot)
	app.browserMgr.Profiles["profile-stopped"] = &browser.Profile{
		ProfileId: "profile-stopped",
		Running:   false,
		FingerprintArgs: []string{
			"--fingerprint-brand=Chrome",
			"--fingerprint-platform=windows",
		},
		LaunchArgs: []string{
			"--timezone=Asia/Shanghai",
		},
		LastLaunchArgs: []string{
			"--timezone=Asia/Tokyo",
		},
	}

	data, err := app.buildFingerprintCheckPageContext("profile-stopped")
	if err != nil {
		t.Fatalf("buildFingerprintCheckPageContext error = %v", err)
	}
	var context fingerprintCheckPageContext
	if err := json.Unmarshal(data, &context); err != nil {
		t.Fatalf("unmarshal context error = %v", err)
	}
	if context.Expected.Timezone != "Asia/Shanghai" {
		t.Fatalf("timezone = %q, want stopped profile config timezone", context.Expected.Timezone)
	}
}

func TestFingerprintCheckPageContextUsesLastLaunchArgsWhenRunning(t *testing.T) {
	app := NewApp(t.TempDir())
	app.browserMgr = browser.NewManager(nil, app.appRoot)
	app.browserMgr.Profiles["profile-running"] = &browser.Profile{
		ProfileId: "profile-running",
		Running:   true,
		FingerprintArgs: []string{
			"--fingerprint-brand=Chrome",
			"--fingerprint-platform=windows",
		},
		LaunchArgs: []string{
			"--timezone=Asia/Shanghai",
		},
		LastLaunchArgs: []string{
			"--timezone=Asia/Tokyo",
		},
	}

	data, err := app.buildFingerprintCheckPageContext("profile-running")
	if err != nil {
		t.Fatalf("buildFingerprintCheckPageContext error = %v", err)
	}
	var context fingerprintCheckPageContext
	if err := json.Unmarshal(data, &context); err != nil {
		t.Fatalf("unmarshal context error = %v", err)
	}
	if context.Expected.Timezone != "Asia/Tokyo" {
		t.Fatalf("timezone = %q, want running profile last launch timezone", context.Expected.Timezone)
	}
}
