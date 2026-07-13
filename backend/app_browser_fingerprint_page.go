package backend

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const fingerprintCheckBookmarkURL = "ant://fingerprint-check"

// BrowserInstanceOpenFingerprintCheck 启动或复用实例，并在目标浏览器内打开本地指纹检测页。
func (a *App) BrowserInstanceOpenFingerprintCheck(profileId string) (*BrowserProfile, error) {
	profileId = strings.TrimSpace(profileId)
	if profileId == "" {
		return nil, fmt.Errorf("实例 ID 不能为空")
	}
	return a.browserInstanceStartInternal(profileId, nil, []string{fingerprintCheckBookmarkURL}, true, true, false, "", "")
}

func (a *App) ensureFingerprintCheckPageURL(profileId string) (string, error) {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return "", err
	}
	return a.ensureFingerprintCheckPageURLForExpectedArgs(profileId, expectedArgs, true)
}

func (a *App) ensureFingerprintCheckPageBookmarkURL(profileId string) (string, error) {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return "", err
	}
	return a.ensureFingerprintCheckPageURLForExpectedArgs(profileId, expectedArgs, false)
}

func (a *App) ensureFingerprintCheckPageURLForProfile(profileId string, coreId string, fingerprintArgs []string, withTimestamp bool) (string, error) {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.ensureFingerprintCheckPageURLForExpectedArgs(profileId, expectedArgs, withTimestamp)
}

func (a *App) ensureFingerprintCheckPageURLForExpectedArgs(profileId string, expectedArgs []string, withTimestamp bool) (string, error) {
	pagePath, err := a.writeFingerprintCheckPageForExpectedArgs(profileId, expectedArgs)
	if err != nil {
		return "", err
	}
	return fingerprintCheckPageFileURL(pagePath, profileId, withTimestamp), nil
}

func (a *App) writeFingerprintCheckPageForProfile(profileId string, coreId string, fingerprintArgs []string) (string, error) {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.writeFingerprintCheckPageForExpectedArgs(profileId, expectedArgs)
}

func (a *App) writeFingerprintCheckPageForExpectedArgs(profileId string, expectedArgs []string) (string, error) {
	pageDir := a.resolveAppPath(filepath.ToSlash(filepath.Join("data", "fingerprint-check", safeFingerprintCheckProfilePath(profileId))))
	if err := os.MkdirAll(pageDir, 0o755); err != nil {
		return "", fmt.Errorf("创建指纹检测页目录失败: %w", err)
	}
	contextData, err := a.buildFingerprintCheckPageContextForExpectedArgs(profileId, expectedArgs)
	if err != nil {
		return "", err
	}
	pagePath := filepath.Join(pageDir, "index.html")
	pageHTML := strings.Replace(fingerprintCheckHTML, "__FINGERPRINT_CHECK_CONTEXT__", string(contextData), 1)
	if err := os.WriteFile(pagePath, []byte(pageHTML), 0o644); err != nil {
		return "", fmt.Errorf("写入指纹检测页失败: %w", err)
	}
	return pagePath, nil
}

func fingerprintCheckPageFileURL(pagePath string, profileId string, withTimestamp bool) string {
	urlPath := filepath.ToSlash(pagePath)
	if len(urlPath) >= 2 && urlPath[1] == ':' {
		urlPath = "/" + urlPath
	}
	fileURL := url.URL{Scheme: "file", Path: urlPath}
	query := fileURL.Query()
	query.Set("profileId", profileId)
	if withTimestamp {
		query.Set("ts", fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	fileURL.RawQuery = query.Encode()
	return fileURL.String()
}

func safeFingerprintCheckProfilePath(profileId string) string {
	profileId = strings.TrimSpace(profileId)
	if profileId == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, char := range profileId {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' {
			builder.WriteRune(char)
		} else {
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "unknown"
	}
	return builder.String()
}

func (a *App) resolveFingerprintCheckStartURL(profileId string, targetURL string) string {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return targetURL
	}
	return a.resolveFingerprintCheckStartURLForExpectedArgs(profileId, expectedArgs, targetURL)
}

func (a *App) resolveFingerprintCheckStartURLForProfile(profileId string, coreId string, fingerprintArgs []string, targetURL string) string {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.resolveFingerprintCheckStartURLForExpectedArgs(profileId, expectedArgs, targetURL)
}

func (a *App) resolveFingerprintCheckStartURLForExpectedArgs(profileId string, expectedArgs []string, targetURL string) string {
	if !strings.EqualFold(strings.TrimSpace(targetURL), fingerprintCheckBookmarkURL) {
		return targetURL
	}
	pageURL, err := a.ensureFingerprintCheckPageURLForExpectedArgs(profileId, expectedArgs, true)
	if err != nil {
		return targetURL
	}
	return pageURL
}

func (a *App) resolveFingerprintCheckStartURLs(profileId string, urls []string) []string {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return append([]string{}, urls...)
	}
	return a.resolveFingerprintCheckStartURLsForExpectedArgs(profileId, expectedArgs, urls)
}

func (a *App) resolveFingerprintCheckStartURLsForProfile(profileId string, coreId string, fingerprintArgs []string, urls []string) []string {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.resolveFingerprintCheckStartURLsForExpectedArgs(profileId, expectedArgs, urls)
}

func (a *App) resolveFingerprintCheckStartURLsForExpectedArgs(profileId string, expectedArgs []string, urls []string) []string {
	if len(urls) == 0 {
		return urls
	}
	out := append([]string{}, urls...)
	for index, item := range out {
		out[index] = a.resolveFingerprintCheckStartURLForExpectedArgs(profileId, expectedArgs, item)
	}
	return out
}

func (a *App) runtimeBookmarksForProfile(profileId string, bookmarks []BrowserBookmark) ([]BrowserBookmark, string, error) {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return nil, "", err
	}
	return a.runtimeBookmarksForProfileExpectedArgs(profileId, expectedArgs, bookmarks)
}

func (a *App) runtimeBookmarksForProfileData(profileId string, coreId string, fingerprintArgs []string, bookmarks []BrowserBookmark) ([]BrowserBookmark, string, error) {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.runtimeBookmarksForProfileExpectedArgs(profileId, expectedArgs, bookmarks)
}

func (a *App) runtimeBookmarksForProfileExpectedArgs(profileId string, expectedArgs []string, bookmarks []BrowserBookmark) ([]BrowserBookmark, string, error) {
	if len(bookmarks) == 0 {
		return bookmarks, "", nil
	}
	needsFingerprintURL := false
	for _, item := range bookmarks {
		if strings.EqualFold(strings.TrimSpace(item.URL), fingerprintCheckBookmarkURL) {
			needsFingerprintURL = true
			break
		}
	}
	if !needsFingerprintURL {
		return append([]BrowserBookmark{}, bookmarks...), "", nil
	}
	pageURL, err := a.ensureFingerprintCheckPageURLForExpectedArgs(profileId, expectedArgs, false)
	if err != nil {
		return nil, "", err
	}
	out := append([]BrowserBookmark{}, bookmarks...)
	for index := range out {
		if strings.EqualFold(strings.TrimSpace(out[index].URL), fingerprintCheckBookmarkURL) {
			out[index].URL = pageURL
		}
	}
	return out, pageURL, nil
}

type fingerprintCheckPageContext struct {
	ProfileId string                         `json:"profileId"`
	Expected  BrowserFingerprintExpectedInfo `json:"expected"`
}

func (a *App) buildFingerprintCheckPageContext(profileId string) ([]byte, error) {
	expectedArgs, err := a.fingerprintCheckProfileExpectedArgs(profileId)
	if err != nil {
		return nil, err
	}
	return a.buildFingerprintCheckPageContextForExpectedArgs(profileId, expectedArgs)
}

func (a *App) buildFingerprintCheckPageContextForProfile(profileId string, coreId string, fingerprintArgs []string) ([]byte, error) {
	expectedArgs := a.buildFingerprintCheckExpectedArgs(profileId, coreId, fingerprintArgs, nil)
	return a.buildFingerprintCheckPageContextForExpectedArgs(profileId, expectedArgs)
}

func (a *App) buildFingerprintCheckPageContextForExpectedArgs(profileId string, expectedArgs []string) ([]byte, error) {
	expected := buildBrowserFingerprintExpected(expectedArgs)
	data, err := json.MarshalIndent(fingerprintCheckPageContext{ProfileId: profileId, Expected: expected}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("生成指纹检测上下文失败: %w", err)
	}
	return append(data, '\n'), nil
}

func (a *App) fingerprintCheckProfileExpectedArgs(profileId string) ([]string, error) {
	profile, err := a.fingerprintCheckProfileSnapshot(profileId)
	if err != nil {
		return nil, err
	}
	return a.fingerprintCheckExpectedArgsFromProfile(profile), nil
}

func (a *App) buildFingerprintCheckExpectedArgs(profileId string, coreId string, fingerprintArgs []string, launchArgs []string) []string {
	fingerprintLaunchArgs := a.buildBrowserFingerprintCapabilityReport(profileId, coreId, fingerprintArgs).LaunchArgs
	sanitizedLaunchArgs, _ := sanitizeManagedLaunchArgs(launchArgs)
	return combineFingerprintExpectedArgs(fingerprintLaunchArgs, sanitizedLaunchArgs)
}

func combineFingerprintExpectedArgs(argGroups ...[]string) []string {
	result := make([]string, 0)
	for _, args := range argGroups {
		result = append(result, normalizeNonEmptyStrings(args)...)
	}
	return result
}

func (a *App) fingerprintCheckExpectedArgsFromProfile(profile *BrowserProfile) []string {
	if profile == nil {
		return nil
	}
	if browserProfileMayHaveRuntimeLaunchArgs(profile) {
		if args := normalizeNonEmptyStrings(profile.LastLaunchArgs); len(args) > 0 {
			return args
		}
		if args := a.recoverBrowserLaunchArgsForProfile(profile); len(args) > 0 {
			profile.LastLaunchArgs = append([]string{}, args...)
			return args
		}
	}
	return a.buildFingerprintCheckExpectedArgs(profile.ProfileId, profile.CoreId, profile.FingerprintArgs, profile.LaunchArgs)
}

func (a *App) fingerprintCheckExpectedArgsFromLockedProfile(profile *BrowserProfile) []string {
	if profile == nil {
		return nil
	}
	if browserProfileMayHaveRuntimeLaunchArgs(profile) {
		if args := normalizeNonEmptyStrings(profile.LastLaunchArgs); len(args) > 0 {
			return args
		}
		if args := a.recoverBrowserLaunchArgsForProfile(profile); len(args) > 0 {
			profile.LastLaunchArgs = append([]string{}, args...)
			return args
		}
	}
	return a.buildFingerprintCheckExpectedArgs(profile.ProfileId, profile.CoreId, profile.FingerprintArgs, profile.LaunchArgs)
}

func browserProfileMayHaveRuntimeLaunchArgs(profile *BrowserProfile) bool {
	return profile != nil && (profile.Running || profile.Pid > 0 || profile.DebugPort > 0)
}

func (a *App) recoverBrowserLaunchArgsForProfile(profile *BrowserProfile) []string {
	if a == nil || a.browserMgr == nil || profile == nil {
		return nil
	}
	if a.browserMgr.Config == nil || (!profile.Running && profile.Pid <= 0 && profile.DebugPort <= 0) {
		return nil
	}
	userDataDir := a.browserMgr.ResolveUserDataDir(profile)
	processes, err := findBrowserUserDataProcesses(userDataDir)
	if err != nil || len(processes) == 0 {
		return nil
	}
	if profile.Pid > 0 {
		for _, process := range processes {
			if process.PID == profile.Pid {
				return parseBrowserProcessCommandLineArgs(process.CommandLine)
			}
		}
	}
	if profile.DebugPort > 0 {
		for _, process := range processes {
			debugPort := process.DebugPort
			if debugPort <= 0 {
				debugPort = parseRemoteDebuggingPort(process.CommandLine)
			}
			if debugPort == profile.DebugPort {
				return parseBrowserProcessCommandLineArgs(process.CommandLine)
			}
		}
	}
	return parseBrowserProcessCommandLineArgs(processes[0].CommandLine)
}

func (a *App) fingerprintCheckProfileSnapshot(profileId string) (*BrowserProfile, error) {
	if a == nil || a.browserMgr == nil {
		return nil, fmt.Errorf("浏览器管理器未初始化")
	}
	a.browserMgr.Mutex.Lock()
	defer a.browserMgr.Mutex.Unlock()
	profile := a.browserMgr.Profiles[profileId]
	if profile == nil {
		return nil, fmt.Errorf("实例不存在: %s", profileId)
	}
	snapshot := *profile
	snapshot.FingerprintArgs = append([]string{}, profile.FingerprintArgs...)
	snapshot.LaunchArgs = append([]string{}, profile.LaunchArgs...)
	snapshot.LastLaunchArgs = append([]string{}, profile.LastLaunchArgs...)
	return &snapshot, nil
}

const fingerprintCheckHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Ant 指纹检测</title>
  <style>
    :root { color-scheme: light; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; background: #f6f7f9; color: #111827; }
    main { max-width: 1480px; margin: 0 auto; padding: 24px; }
    header { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
    h1 { margin: 0; font-size: 22px; }
    .meta { color: #6b7280; font-size: 13px; margin-top: 4px; }
    .actions { display: flex; gap: 8px; }
    button { border: 1px solid #111827; background: #111827; color: #fff; border-radius: 8px; height: 34px; padding: 0 12px; cursor: pointer; }
    button.secondary { background: #fff; color: #111827; }
    .grid { display: block; }
    section { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; overflow-x: auto; }
    h2 { margin: 0; padding: 11px 13px; font-size: 14px; border-bottom: 1px solid #e5e7eb; background: #fafafa; }
    table { width: 100%; min-width: 1280px; border-collapse: collapse; }
    th, td { padding: 9px 12px; border-bottom: 1px solid #f1f5f9; vertical-align: top; font-size: 13px; text-align: left; }
    th { background: #f8fafc; color: #475569; font-weight: 600; white-space: nowrap; }
    tr:last-child td { border-bottom: 0; }
    td.item { width: 160px; color: #111827; font-weight: 600; white-space: nowrap; }
    td.source { width: 108px; color: #334155; font-weight: 600; white-space: nowrap; }
    td.hit { width: 92px; font-weight: 700; white-space: nowrap; }
    td.reason { width: 380px; color: #475569; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; word-break: break-all; }
    .value-pair { display: grid; gap: 5px; }
    .value-line { display: grid; grid-template-columns: 42px minmax(0, 1fr); gap: 8px; align-items: start; }
    .value-label { color: #64748b; }
    .ok { color: #047857; }
    .warn { color: #b45309; }
    .bad { color: #b91c1c; }
    .muted { color: #64748b; }
    .summary { display: none; }
    pre { margin: 0; padding: 12px; white-space: pre-wrap; word-break: break-all; font-size: 12px; }
  </style>
</head>
<body>
<main>
  <header>
    <div>
      <h1>Ant 指纹检测</h1>
      <div class="meta" id="meta">正在检测当前浏览器真实指纹...</div>
    </div>
    <div class="actions">
      <button class="secondary" id="resetBaselineBtn">重建基线</button>
      <button class="secondary" id="refreshBtn">重新检测</button>
      <button id="copyBtn">复制 JSON</button>
    </div>
  </header>
  <div class="summary" id="summary"></div>
  <div class="grid" id="app"></div>
</main>
<script>
var latestReport = null;
var latestContext = __FINGERPRINT_CHECK_CONTEXT__;
var fingerprintCheckRunning = false;
var FINGERPRINT_AUTO_REFRESH_MS = 60 * 60 * 1000;
var FINGERPRINT_AUTO_REFRESH_CHECK_MS = 60 * 1000;
function hashString(input) {
  var hash = 2166136261;
  var text = String(input || '');
  for (var i = 0; i < text.length; i++) {
    hash ^= text.charCodeAt(i);
    hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24);
  }
  return ('00000000' + (hash >>> 0).toString(16)).slice(-8);
}
function safe(fn, fallback) { try { return fn(); } catch (e) { return fallback; } }
function canvasHash() {
  return safe(function () {
    var canvas = document.createElement('canvas');
    canvas.width = 320; canvas.height = 96;
    var ctx = canvas.getContext('2d');
    ctx.textBaseline = 'top';
    ctx.font = '16px Arial';
    ctx.fillStyle = '#f60'; ctx.fillRect(4, 4, 150, 36);
    ctx.fillStyle = '#069'; ctx.fillText('Ant fingerprint 检测', 9, 12);
    ctx.strokeStyle = 'rgba(120,60,200,.85)'; ctx.beginPath(); ctx.arc(210, 44, 30, 0, Math.PI * 2); ctx.stroke();
    return hashString(canvas.toDataURL());
  }, '');
}
async function audioHash() {
  return await safe(async function () {
    var Ctor = window.OfflineAudioContext || window.webkitOfflineAudioContext;
    if (!Ctor) return '';
    var ctx = new Ctor(1, 44100, 44100);
    var osc = ctx.createOscillator();
    var comp = ctx.createDynamicsCompressor();
    osc.type = 'triangle'; osc.frequency.value = 10000;
    comp.threshold.value = -50; comp.knee.value = 40; comp.ratio.value = 12; comp.attack.value = 0; comp.release.value = .25;
    osc.connect(comp); comp.connect(ctx.destination); osc.start(0);
    var buffer = await ctx.startRendering();
    var data = buffer.getChannelData(0).slice(4500, 5000);
    return hashString(Array.prototype.map.call(data, function (v) { return v.toFixed(6); }).join(','));
  }, '');
}
function clientRectsHash() {
  return safe(function () {
    var node = document.createElement('div');
    node.style.cssText = 'position:absolute;left:-9999px;top:-9999px;width:180px;font:13px Arial;line-height:17px;';
    node.textContent = 'Ant fingerprint client rects check';
    document.body.appendChild(node);
    var rects = Array.prototype.map.call(node.getClientRects(), function (r) {
      return [r.x, r.y, r.width, r.height].map(function (v) { return Number(v).toFixed(3); }).join(':');
    }).join('|');
    document.body.removeChild(node);
    return hashString(rects);
  }, '');
}
function fontProbe() {
  return safe(function () {
    var baseFonts = ['monospace', 'sans-serif', 'serif'];
    var candidates = ['Arial', 'Calibri', 'Cambria', 'Consolas', 'Courier New', 'Georgia', 'Helvetica', 'Microsoft YaHei', 'PingFang SC', 'Roboto', 'Segoe UI', 'Times New Roman'];
    var text = 'mmmmmmmmmmlli';
    var size = '72px';
    var canvas = document.createElement('canvas');
    var ctx = canvas.getContext('2d');
    var base = {};
    baseFonts.forEach(function (font) { ctx.font = size + ' ' + font; base[font] = ctx.measureText(text).width; });
    var detected = candidates.filter(function (font) {
      return baseFonts.some(function (baseFont) {
        ctx.font = size + ' "' + font + '",' + baseFont;
        return ctx.measureText(text).width !== base[baseFont];
      });
    });
    return { detected: detected, hash: hashString(detected.join('|')) };
  }, { detected: [], hash: '' });
}
function webglInfo() {
  return safe(function () {
    var canvas = document.createElement('canvas');
    var gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
    if (!gl) return { vendor: '', renderer: '', hash: '' };
    var debug = gl.getExtension('WEBGL_debug_renderer_info');
    var vendor = debug ? gl.getParameter(debug.UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR);
    var renderer = debug ? gl.getParameter(debug.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER);
    var params = [vendor, renderer, gl.getParameter(gl.VERSION), gl.getParameter(gl.SHADING_LANGUAGE_VERSION)].join('|');
    return { vendor: vendor || '', renderer: renderer || '', hash: hashString(params) };
  }, { vendor: '', renderer: '', hash: '' });
}
async function webrtcCandidates() {
  return await safe(function () {
    return new Promise(function (resolve) {
      if (!window.RTCPeerConnection) return resolve([]);
      var pc = new RTCPeerConnection({ iceServers: [] });
      var candidates = [];
      pc.createDataChannel('ant');
      pc.onicecandidate = function (event) {
        if (event && event.candidate && event.candidate.candidate) candidates.push(event.candidate.candidate);
      };
      pc.createOffer().then(function (offer) { return pc.setLocalDescription(offer); }).catch(function () {});
      setTimeout(function () { try { pc.close(); } catch (e) {} resolve(candidates); }, 1600);
    });
  }, []);
}
async function collect() {
  var uaData = navigator.userAgentData ? {
    brands: navigator.userAgentData.brands || [],
    mobile: navigator.userAgentData.mobile,
    platform: navigator.userAgentData.platform || ''
  } : null;
  var fonts = fontProbe();
  var gl = webglInfo();
  var candidates = await webrtcCandidates();
  return {
    generatedAt: new Date().toISOString(),
    urlProfileId: new URLSearchParams(location.search).get('profileId') || '',
    identity: {
      userAgent: navigator.userAgent || '',
      platform: navigator.platform || '',
      userAgentData: uaData,
      webdriver: navigator.webdriver === true
    },
    locale: {
      language: navigator.language || '',
      languages: Array.prototype.slice.call(navigator.languages || []),
      timezone: (Intl.DateTimeFormat().resolvedOptions() || {}).timeZone || '',
      timezoneOffset: new Date().getTimezoneOffset()
    },
    hardware: {
      hardwareConcurrency: navigator.hardwareConcurrency || 0,
      deviceMemory: navigator.deviceMemory || 0,
      maxTouchPoints: navigator.maxTouchPoints || 0,
      cookieEnabled: navigator.cookieEnabled === true,
      doNotTrack: navigator.doNotTrack || ''
    },
    screen: {
      width: screen.width || 0,
      height: screen.height || 0,
      availWidth: screen.availWidth || 0,
      availHeight: screen.availHeight || 0,
      colorDepth: screen.colorDepth || 0,
      pixelDepth: screen.pixelDepth || 0,
      devicePixelRatio: window.devicePixelRatio || 0,
      innerWidth: window.innerWidth || 0,
      innerHeight: window.innerHeight || 0,
      outerWidth: window.outerWidth || 0,
      outerHeight: window.outerHeight || 0
    },
    advanced: {
      canvasHash: canvasHash(),
      audioHash: await audioHash(),
      clientRectsHash: clientRectsHash(),
      fontHash: fonts.hash,
      detectedFonts: fonts.detected,
      webglVendor: gl.vendor,
      webglRenderer: gl.renderer,
      webglHash: gl.hash,
      plugins: Array.prototype.map.call(navigator.plugins || [], function (p) { return p.name || ''; }),
      mimeTypes: Array.prototype.map.call(navigator.mimeTypes || [], function (m) { return m.type || ''; })
    },
    network: {
      webrtcCandidates: candidates,
      localCandidateCount: candidates.filter(function (item) { return / typ host /.test(item); }).length
    }
  };
}
function csvItems(value) { return String(value || '').split(',').map(function (item) { return item.trim(); }).filter(Boolean); }
function matchExact(expected, actual) {
  if (expected === undefined || expected === null || expected === '') return 'unknown';
  return String(expected) === String(actual) ? 'match' : 'mismatch';
}
function matchArrayPrefix(expectedCsv, actualArray) {
  var expected = csvItems(expectedCsv);
  if (!expected.length) return 'unknown';
  var actual = (actualArray || []).map(function (item) { return String(item).trim(); });
  return expected.every(function (item, index) { return actual[index] === item; }) ? 'match' : 'mismatch';
}
function matchContains(expected, actual) {
  if (expected === undefined || expected === null || expected === '') return 'unknown';
  return String(actual || '').indexOf(String(expected)) >= 0 ? 'match' : 'mismatch';
}
function displayValue(value) {
  if (value === undefined || value === null || value === '') return '-';
  if (Array.isArray(value)) return value.length ? value.join(', ') : '-';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}
function hasExpected(value) {
  return !(value === undefined || value === null || value === '');
}
function pairValue(expected, actual) {
  return '<div class="value-pair"><div class="value-line"><span class="value-label">期望</span><code>' + escapeHtml(displayValue(expected)) + '</code></div><div class="value-line"><span class="value-label">实际</span><code>' + escapeHtml(displayValue(actual)) + '</code></div></div>';
}
function normalizeDisplay(value) { return displayValue(value); }
function compareExactStatus(expected, actual) {
  if (!hasExpected(expected)) return 'unknown';
  return String(expected) === String(actual) ? 'match' : 'mismatch';
}
function compareArrayPrefixStatus(expectedCsv, actualArray) {
  var expected = csvItems(expectedCsv);
  if (!expected.length) return 'unknown';
  var actual = (actualArray || []).map(function (item) { return String(item).trim(); });
  return expected.every(function (item, index) { return actual[index] === item; }) ? 'match' : 'mismatch';
}
function compareArrayContainsAllStatus(expectedCsv, actualArray) {
  var expected = csvItems(expectedCsv);
  if (!expected.length) return 'unknown';
  var actual = (actualArray || []).map(function (item) { return String(item).trim().toLowerCase(); });
  return expected.every(function (item) { return actual.indexOf(String(item).trim().toLowerCase()) >= 0; }) ? 'match' : 'mismatch';
}
function compareContainsStatus(expected, actual) {
  if (!hasExpected(expected)) return 'unknown';
  return String(actual || '').indexOf(String(expected)) >= 0 ? 'match' : 'mismatch';
}
function versionParts(value) {
  var normalized = String(value || '').replace(/_/g, '.');
  var match = normalized.match(/\d+(?:\.\d+)*/);
  return match ? match[0].split('.').map(function (item) { return parseInt(item, 10); }).filter(function (item) { return !isNaN(item); }) : [];
}
function versionList(value, patterns) {
  var text = String(value || '').replace(/_/g, '.');
  var list = [];
  patterns.forEach(function (pattern) {
    var match;
    while ((match = pattern.exec(text)) !== null) {
      var parts = versionParts(match[1]);
      if (parts.length) list.push(parts);
    }
  });
  return list;
}
function sameVersionPrefix(left, right) {
  if (!left.length || !right.length) return false;
  var size = Math.min(left.length, right.length);
  for (var index = 0; index < size; index += 1) {
    if (left[index] !== right[index]) return false;
  }
  return true;
}
function compareBrowserVersionStatus(expected, actual) {
  if (!hasExpected(expected)) return 'unknown';
  if (compareContainsStatus(expected, actual) === 'match') return 'match';
  var expectedParts = versionParts(expected);
  if (!expectedParts.length) return 'mismatch';
  var browserVersions = versionList(actual, [/(?:Chrome|Chromium|Edg|OPR|Vivaldi)\/([0-9]+(?:[._][0-9]+)*)/g, /"version"\s*:\s*"([0-9]+(?:[._][0-9]+)*)"/g]);
  return browserVersions.some(function (actualParts) { return actualParts[0] === expectedParts[0]; }) ? 'compatible' : 'mismatch';
}
function comparePlatformVersionStatus(expected, actual) {
  if (!hasExpected(expected)) return 'unknown';
  if (compareContainsStatus(expected, actual) === 'match') return 'match';
  var expectedParts = versionParts(expected);
  if (!expectedParts.length) return 'mismatch';
  var platformVersions = versionList(actual, [/Windows NT\s+([0-9]+(?:[._][0-9]+)*)/g, /Mac OS X\s+([0-9]+(?:[._][0-9]+)*)/g, /Android\s+([0-9]+(?:[._][0-9]+)*)/g, /(?:CPU (?:iPhone )?OS|iPhone OS)\s+([0-9]+(?:[._][0-9]+)*)/g, /"platformVersion"\s*:\s*"([0-9]+(?:[._][0-9]+)*)"/g]);
  return platformVersions.some(function (actualParts) { return sameVersionPrefix(expectedParts, actualParts); }) ? 'compatible' : 'mismatch';
}
function normalizePlatformForCompare(value) {
  var normalized = String(value || '').trim().toLowerCase();
  if (!normalized) return '';
  if (['windows', 'win', 'win32', 'win64', 'wince'].indexOf(normalized) >= 0) return 'windows';
  if (normalized.indexOf('linux') === 0 || normalized === 'x11') return 'linux';
  if (normalized === 'mac' || normalized === 'macos' || normalized.indexOf('mac') >= 0) return 'macos';
  return normalized;
}
function comparePlatformStatus(expected, actual) {
  if (!hasExpected(expected)) return 'unknown';
  var expectedPlatform = normalizePlatformForCompare(expected);
  var actualPlatform = normalizePlatformForCompare(actual);
  return expectedPlatform && expectedPlatform === actualPlatform ? 'match' : 'mismatch';
}
function statusText(status, source) {
  if (source === '效果观测' && status === 'baseline') return '已建观测基线';
  if (source === '效果观测' && status === 'match') return '观测一致';
  if (source === '效果观测' && status === 'mismatch') return '观测变化';
  if (source === '运行基线' && status === 'match') return '基线一致';
  if (source === '运行基线' && status === 'mismatch') return '基线变化';
  if (status === 'match') return '命中';
  if (status === 'compatible') return '口径匹配';
  if (status === 'mismatch') return '未命中';
  if (status === 'warning') return '风险';
  if (status === 'unreadable') return '已配置';
  if (status === 'unsupported') return '不可配置';
  if (status === 'baseline') return '已建基线';
  if (source === '未配置') return '未配置';
  return '未采集';
}
function statusClass(status, source) {
  if (status === 'match') return 'ok';
  if (status === 'compatible') return 'ok';
  if (source === '效果观测' && status === 'mismatch') return 'warn';
  if (source === '运行基线' && status === 'mismatch') return 'warn';
  if (status === 'mismatch') return 'bad';
  if (status === 'warning') return 'warn';
  return 'muted';
}
function reasonFor(status, reason) {
  if (reason) return reason;
  if (status === 'unknown') return '未显式配置，且本次没有可用实际值，无法建立运行基线';
  if (status === 'match') return '实际值与期望值一致';
  if (status === 'mismatch') return '实际值与期望值不一致';
  if (status === 'unreadable') return '浏览器 JS 无法读取该启动配置，只展示期望值';
  if (status === 'unsupported') return '当前内核本地实测该独立参数无效，未作为可配置期望';
  if (status === 'baseline') return '首次采集并保存为运行基线';
  return '';
}
function sourceFor(status, expected, actual) {
  if (hasExpected(expected)) return '配置期望';
  if (status === 'baseline' || status === 'match' || status === 'mismatch') return '运行基线';
  if (status === 'unreadable') return '配置下发';
  if (hasObservedValue(actual)) return '未配置';
  return '未采集';
}
function fingerprintRow(name, expected, actual, status, reason, source) {
  return {
    name: name,
    expected: expected,
    actual: actual,
    status: status,
    source: source || sourceFor(status, expected, actual),
    reason: reasonFor(status, reason)
  };
}
var latestBaseline = {};
var latestBaselineCreated = {};
function baselineStorageKey(context) {
  var expected = context && context.expected ? context.expected : {};
  var profileId = context && context.profileId ? context.profileId : 'unknown';
  var seed = expected.seed || 'no-seed';
  return 'ant:fingerprint-check:baseline:' + profileId + ':' + seed;
}
function loadFingerprintBaseline(context) {
  latestBaselineCreated = {};
  try {
    latestBaseline = JSON.parse(localStorage.getItem(baselineStorageKey(context)) || '{}') || {};
  } catch (e) {
    latestBaseline = {};
  }
}
function saveFingerprintBaseline(context) {
  try {
    localStorage.setItem(baselineStorageKey(context), JSON.stringify(latestBaseline));
  } catch (e) {}
}
function hasObservedValue(value) {
  if (value === undefined || value === null || value === '') return false;
  if (Array.isArray(value)) return value.length > 0;
  return true;
}
function cloneObservedValue(value) {
  if (Array.isArray(value)) return value.slice();
  if (value && typeof value === 'object') return JSON.parse(JSON.stringify(value));
  return value;
}
function sameObservedValue(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}
function baselineValue(context, key, actual) {
  if (Object.prototype.hasOwnProperty.call(latestBaseline, key)) return latestBaseline[key];
  if (!hasObservedValue(actual)) return '';
  latestBaseline[key] = cloneObservedValue(actual);
  latestBaselineCreated[key] = true;
  saveFingerprintBaseline(context);
  return latestBaseline[key];
}
function observedExpected(context, key, explicitValue, actual, emptyExpectedText) {
  if (hasExpected(explicitValue)) return explicitValue;
  var value = baselineValue(context, key, actual);
  return hasObservedValue(value) ? value : (emptyExpectedText || '未建立运行基线');
}
function observedStatus(context, key, explicitValue, actual, explicitCompare) {
  if (hasExpected(explicitValue)) return explicitCompare(explicitValue, actual);
  var value = baselineValue(context, key, actual);
  if (!hasObservedValue(value)) return 'unknown';
  if (latestBaselineCreated[key]) return 'baseline';
  return sameObservedValue(value, actual) ? 'match' : 'mismatch';
}
function observedReason(name, explicitValue, explicitReason, status, source) {
  if (hasExpected(explicitValue)) return explicitReason;
  if (source === '效果观测') {
    var prefix = explicitReason || (name + ' 是效果观测值，不是配置期望');
    if (status === 'baseline') return prefix + '；已保存首次实际值为观测基线。改过指纹配置或 Seed 后，先重建基线再刷新验证稳定性';
    if (status === 'mismatch') return prefix + '；当前实际值与观测基线不同，说明输出已变化，不等于配置失败。若未改配置且重建基线后仍变化，才是稳定性问题';
    if (status === 'match') return prefix + '；当前实际值与观测基线一致';
    return prefix + '；本次没有可用实际值，无法建立观测基线';
  }
  if (status === 'baseline') return name + ' 没有显式配置值；检测页已用首次实际采集值建立运行基线，后续按基线比对';
  if (status === 'mismatch') return name + ' 没有显式配置期望；当前实际值与运行基线不同，这是观测值变化，不代表配置未生效';
  if (status === 'match') return name + ' 没有显式配置期望；当前实际值与运行基线一致';
  return name + ' 没有显式配置期望，且本次没有可用实际值，无法建立运行基线';
}
function observedSource(context, key, explicitValue, actual, observedSourceName) {
  if (hasExpected(explicitValue)) return '配置期望';
  var value = baselineValue(context, key, actual);
  return hasObservedValue(value) ? (observedSourceName || '运行基线') : '未采集';
}
function observedFingerprintRow(name, context, key, explicitValue, actual, explicitCompare, explicitReason, options) {
  options = options || {};
  var expectedValue = observedExpected(context, key, explicitValue, actual, options.emptyExpectedText);
  var status = observedStatus(context, key, explicitValue, actual, explicitCompare);
  var source = observedSource(context, key, explicitValue, actual, options.observedSourceName);
  var reason = observedReason(name, explicitValue, explicitReason, status, source);
  return fingerprintRow(name, expectedValue, actual, status, reason, source);
}
function effectObservedFingerprintRow(name, context, key, actual, reason) {
  return observedFingerprintRow(name, context, key, '', actual, compareExactStatus, reason, {
    observedSourceName: '效果观测',
    emptyExpectedText: '未建立观测基线'
  });
}
function resetFingerprintBaseline(context) {
  latestBaseline = {};
  latestBaselineCreated = {};
  try {
    localStorage.removeItem(baselineStorageKey(context));
  } catch (e) {}
}
function buildFingerprintRows(report, context) {
  loadFingerprintBaseline(context);
  var expected = context && context.expected ? context.expected : {};
  var uaVersion = report.identity.userAgent + (report.identity.userAgentData ? ' / ' + JSON.stringify(report.identity.userAgentData) : '');
  var rows = [];
  rows.push(fingerprintRow('指纹 Seed', expected.seed, 'JS 不可读取', expected.seed ? 'unreadable' : 'unknown', 'Seed 是启动参数，页面无法从 JS 反读'));
  rows.push(fingerprintRow('禁用原生暴露', expected.disableSpoofing, 'JS 不可读取', expected.disableSpoofing ? 'unreadable' : 'unknown', '该项是启动保护策略，页面无法从 JS 反读'));
  rows.push(observedFingerprintRow('语言', context, 'language', expected.language, report.locale.language, compareExactStatus, '比对 navigator.language'));
  rows.push(observedFingerprintRow('语言列表', context, 'languages', expected.acceptLanguage, report.locale.languages, compareArrayPrefixStatus, '比对 navigator.languages 前缀'));
  rows.push(observedFingerprintRow('时区', context, 'timezone', expected.timezone, report.locale.timezone, compareExactStatus, '比对 Intl.DateTimeFormat().resolvedOptions().timeZone'));
  rows.push(observedFingerprintRow('CPU 核心', context, 'hardwareConcurrency', expected.hardwareConcurrency, report.hardware.hardwareConcurrency, compareExactStatus, '比对 navigator.hardwareConcurrency'));
  rows.push(observedFingerprintRow('设备内存', context, 'deviceMemory', expected.deviceMemory, report.hardware.deviceMemory, compareExactStatus, '比对 navigator.deviceMemory'));
  rows.push(observedFingerprintRow('触控点', context, 'maxTouchPoints', expected.touchPoints, report.hardware.maxTouchPoints, compareExactStatus, '比对 navigator.maxTouchPoints'));
  rows.push(observedFingerprintRow('Do Not Track', context, 'doNotTrack', expected.doNotTrack, report.hardware.doNotTrack, compareExactStatus, '比对 navigator.doNotTrack'));
  rows.push(observedFingerprintRow('窗口大小', context, 'windowSize', expected.windowSize, report.screen.outerWidth + ',' + report.screen.outerHeight, compareExactStatus, '比对 window.outerWidth/outerHeight'));
  rows.push(observedFingerprintRow('颜色深度', context, 'colorDepth', expected.colorDepth, report.screen.colorDepth, compareExactStatus, '比对 screen.colorDepth'));
  rows.push(fingerprintRow('品牌', expected.brand, report.identity.userAgent, compareContainsStatus(expected.brand, report.identity.userAgent), '比对 User-Agent 是否包含期望品牌'));
  rows.push(observedFingerprintRow('品牌版本', context, 'brandVersion', expected.brandVersion, report.identity.userAgent, compareBrowserVersionStatus, '优先比对完整浏览器版本；User-Agent 只暴露主版本时按主版本口径匹配'));
  rows.push(fingerprintRow('平台', expected.platform, report.identity.platform, comparePlatformStatus(expected.platform, report.identity.platform), '比对 navigator.platform'));
  rows.push(observedFingerprintRow('平台版本', context, 'platformVersion', expected.platformVersion, uaVersion, comparePlatformVersionStatus, '优先比对完整系统版本；User-Agent / UA-CH 只暴露短版本时按可见版本前缀匹配'));
  rows.push(fingerprintRow('Webdriver', 'false', report.identity.webdriver ? 'true' : 'false', report.identity.webdriver ? 'mismatch' : 'match', '期望 navigator.webdriver 不暴露自动化', '内置期望'));
  var expectsWebRTCBlocked = hasExpected(expected.webrtcPolicy);
  rows.push(fingerprintRow('WebRTC Host', expectsWebRTCBlocked ? '0 host candidates' : '', report.network.localCandidateCount + ' host candidates', expectsWebRTCBlocked ? (report.network.localCandidateCount > 0 ? 'mismatch' : 'match') : 'unknown', expectsWebRTCBlocked ? '比对是否暴露本机 host candidate' : '未配置 WebRTC 期望，只展示实际采集值'));
  rows.push(fingerprintRow('媒体设备数量', '不作为期望', 'JS 不可读取', 'unsupported', '媒体设备数量独立参数本地 Chrom-144 实测无效，未作为运行参数传递', '实测无效'));
  rows.push(fingerprintRow('Canvas 噪声', expected.canvasNoise, 'JS 不可读取', expected.canvasNoise ? 'unreadable' : 'unknown', 'Canvas 噪声开关已作为启动参数下发；页面不能直接反读开关，效果看 Canvas Hash 是否稳定变化'));
  rows.push(fingerprintRow('Audio 噪声', '不作为期望', 'JS 不可读取', 'unsupported', 'Audio 独立噪声参数本地 Chrom-144 实测无效；音频变化通过 Seed 和 Audio Hash 观察', '实测无效'));
  rows.push(fingerprintRow('ClientRects 噪声', expected.clientRectsNoise, 'JS 不可读取', expected.clientRectsNoise ? 'unreadable' : 'unknown', 'ClientRects 噪声开关已作为启动参数下发；页面不能直接反读开关，效果看 ClientRects Hash 是否稳定变化'));
  rows.push(effectObservedFingerprintRow('Canvas Hash', context, 'canvasHash', report.advanced.canvasHash, expected.canvasNoise ? 'Canvas Hash 是 Canvas 噪声效果观测值，不是配置期望' : 'Canvas Hash 是检测输出哈希，不是配置期望'));
  rows.push(effectObservedFingerprintRow('Audio Hash', context, 'audioHash', report.advanced.audioHash, 'Audio Hash 是检测输出哈希，不是配置期望'));
  rows.push(effectObservedFingerprintRow('ClientRects Hash', context, 'clientRectsHash', report.advanced.clientRectsHash, expected.clientRectsNoise ? 'ClientRects Hash 是 ClientRects 噪声效果观测值，不是配置期望' : 'ClientRects Hash 是检测输出哈希，不是配置期望'));
  rows.push(effectObservedFingerprintRow('Fonts Hash', context, 'fontHash', report.advanced.fontHash, 'Fonts Hash 是检测输出哈希，不是配置期望'));
  rows.push(observedFingerprintRow('Detected Fonts', context, 'detectedFonts', expected.fontList, report.advanced.detectedFonts, compareArrayContainsAllStatus, '比对检测到的字体是否包含配置列表'));
  rows.push(observedFingerprintRow('WebGL Vendor', context, 'webglVendor', expected.webGLVendor || expected.webglVendor, report.advanced.webglVendor, compareExactStatus, '比对 WebGL Vendor'));
  rows.push(observedFingerprintRow('WebGL Renderer', context, 'webglRenderer', expected.webGLRenderer || expected.webglRenderer, report.advanced.webglRenderer, compareExactStatus, '比对 WebGL Renderer'));
  rows.push(effectObservedFingerprintRow('WebGL Hash', context, 'webglHash', report.advanced.webglHash, 'WebGL Hash 是检测输出哈希，不是配置期望'));
  rows.push(observedFingerprintRow('Plugins', context, 'plugins', '', report.advanced.plugins, compareExactStatus, '比对插件列表'));
  rows.push(observedFingerprintRow('MIME Types', context, 'mimeTypes', '', report.advanced.mimeTypes, compareExactStatus, '比对 MIME 列表'));
  rows.push(observedFingerprintRow('屏幕尺寸', context, 'screenSize', '', report.screen.width + 'x' + report.screen.height, compareExactStatus, '比对屏幕尺寸'));
  rows.push(observedFingerprintRow('DPR', context, 'devicePixelRatio', '', report.screen.devicePixelRatio, compareExactStatus, '比对 DPR'));
  return rows;
}
function shouldShowReason(status) {
  return status === 'mismatch' || status === 'warning' || status === 'unreadable' || status === 'unsupported' || status === 'baseline' || status === 'unknown';
}
function renderFingerprintTable(rows) {
  return '<section><h2>指纹对比</h2><table><thead><tr><th>指纹项</th><th>值</th><th>期望来源</th><th>结果</th><th>原因</th></tr></thead><tbody>' + rows.map(function (item) {
    var reason = shouldShowReason(item.status) ? item.reason : '';
    return '<tr><td class="item">' + escapeHtml(item.name) + '</td><td>' + pairValue(item.expected, item.actual) + '</td><td class="source">' + escapeHtml(item.source) + '</td><td class="hit ' + statusClass(item.status, item.source) + '">' + statusText(item.status, item.source) + '</td><td class="reason">' + escapeHtml(reason) + '</td></tr>';
  }).join('') + '</tbody></table></section>';
}
function padDatePart(value) { return String(value).padStart(2, '0'); }
function formatLocalDateTime(isoText) {
  var date = new Date(isoText);
  if (Number.isNaN(date.getTime())) return isoText || '-';
  return date.getFullYear() + '-' + padDatePart(date.getMonth() + 1) + '-' + padDatePart(date.getDate()) + ' ' + padDatePart(date.getHours()) + ':' + padDatePart(date.getMinutes()) + ':' + padDatePart(date.getSeconds());
}
function reportAgeMs(report) {
  var time = Date.parse(report && report.generatedAt ? report.generatedAt : '');
  if (Number.isNaN(time)) return 0;
  return Date.now() - time;
}
function renderMeta(report) {
  var parts = ['检测时间：本地 ' + formatLocalDateTime(report.generatedAt), 'UTC ' + report.generatedAt];
  if (report.urlProfileId) parts.push('Profile: ' + report.urlProfileId);
  document.getElementById('meta').textContent = parts.join(' / ');
}
function escapeHtml(text) { return String(text).replace(/[&<>"']/g, function (m) { return ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'})[m]; }); }
function render(report) {
  var rows = buildFingerprintRows(report, latestContext);
  report.rows = rows;
  renderMeta(report);
  document.getElementById('summary').innerHTML = '';
  document.getElementById('app').innerHTML = renderFingerprintTable(rows);
}
async function run() {
  if (fingerprintCheckRunning) return;
  fingerprintCheckRunning = true;
  try {
    latestReport = await collect();
    render(latestReport);
  } finally {
    fingerprintCheckRunning = false;
  }
}
function maybeAutoRefresh() {
  if (!latestReport || fingerprintCheckRunning) return;
  if (reportAgeMs(latestReport) >= FINGERPRINT_AUTO_REFRESH_MS) run();
}
document.getElementById('resetBaselineBtn').onclick = function () { resetFingerprintBaseline(latestContext); run(); };
document.getElementById('refreshBtn').onclick = run;
document.getElementById('copyBtn').onclick = function () {
  if (!latestReport) return;
  navigator.clipboard.writeText(JSON.stringify(latestReport, null, 2)).catch(function () {});
};
setInterval(maybeAutoRefresh, FINGERPRINT_AUTO_REFRESH_CHECK_MS);
run();
</script>
</body>
</html>`
