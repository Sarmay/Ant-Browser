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
	pageDir := a.resolveAppPath(filepath.ToSlash(filepath.Join("data", "fingerprint-check")))
	if err := os.MkdirAll(pageDir, 0o755); err != nil {
		return "", fmt.Errorf("创建指纹检测页目录失败: %w", err)
	}
	contextData, err := a.buildFingerprintCheckPageContext(profileId)
	if err != nil {
		return "", err
	}
	pagePath := filepath.Join(pageDir, "index.html")
	pageHTML := strings.Replace(fingerprintCheckHTML, "__FINGERPRINT_CHECK_CONTEXT__", string(contextData), 1)
	if err := os.WriteFile(pagePath, []byte(pageHTML), 0o644); err != nil {
		return "", fmt.Errorf("写入指纹检测页失败: %w", err)
	}
	urlPath := filepath.ToSlash(pagePath)
	if len(urlPath) >= 2 && urlPath[1] == ':' {
		urlPath = "/" + urlPath
	}
	fileURL := url.URL{Scheme: "file", Path: urlPath}
	query := fileURL.Query()
	query.Set("profileId", profileId)
	query.Set("ts", fmt.Sprintf("%d", time.Now().UnixNano()))
	fileURL.RawQuery = query.Encode()
	return fileURL.String(), nil
}

func (a *App) resolveFingerprintCheckStartURL(profileId string, targetURL string) string {
	if !strings.EqualFold(strings.TrimSpace(targetURL), fingerprintCheckBookmarkURL) {
		return targetURL
	}
	pageURL, err := a.ensureFingerprintCheckPageURL(profileId)
	if err != nil {
		return targetURL
	}
	return pageURL
}

func (a *App) resolveFingerprintCheckStartURLs(profileId string, urls []string) []string {
	if len(urls) == 0 {
		return urls
	}
	out := append([]string{}, urls...)
	for index, item := range out {
		out[index] = a.resolveFingerprintCheckStartURL(profileId, item)
	}
	return out
}

type fingerprintCheckPageContext struct {
	ProfileId string                         `json:"profileId"`
	Expected  BrowserFingerprintExpectedInfo `json:"expected"`
}

func (a *App) buildFingerprintCheckPageContext(profileId string) ([]byte, error) {
	if a == nil || a.browserMgr == nil {
		return nil, fmt.Errorf("浏览器管理器未初始化")
	}
	a.browserMgr.Mutex.Lock()
	profile := a.browserMgr.Profiles[profileId]
	if profile == nil {
		a.browserMgr.Mutex.Unlock()
		return nil, fmt.Errorf("实例不存在: %s", profileId)
	}
	fingerprintArgs := append([]string{}, profile.FingerprintArgs...)
	a.browserMgr.Mutex.Unlock()

	expected := buildBrowserFingerprintExpected(fingerprintArgs)
	if expected.Seed == "" {
		expected.Seed = fmt.Sprintf("%d", browserFingerprintSeedForProfileID(profileId))
	}
	data, err := json.MarshalIndent(fingerprintCheckPageContext{ProfileId: profileId, Expected: expected}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("生成指纹检测上下文失败: %w", err)
	}
	return append(data, '\n'), nil
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
    main { max-width: 1120px; margin: 0 auto; padding: 24px; }
    header { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
    h1 { margin: 0; font-size: 22px; }
    .meta { color: #6b7280; font-size: 13px; margin-top: 4px; }
    .actions { display: flex; gap: 8px; }
    button { border: 1px solid #111827; background: #111827; color: #fff; border-radius: 8px; height: 34px; padding: 0 12px; cursor: pointer; }
    button.secondary { background: #fff; color: #111827; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 14px; }
    section { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; overflow: hidden; }
    h2 { margin: 0; padding: 11px 13px; font-size: 14px; border-bottom: 1px solid #e5e7eb; background: #fafafa; }
    table { width: 100%; border-collapse: collapse; }
    td { padding: 8px 12px; border-bottom: 1px solid #f1f5f9; vertical-align: top; font-size: 13px; }
    tr:last-child td { border-bottom: 0; }
    td:first-child { width: 132px; color: #6b7280; white-space: nowrap; }
    code { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; word-break: break-all; }
    .ok { color: #047857; }
    .warn { color: #b45309; }
    .bad { color: #b91c1c; }
    .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px; margin-bottom: 14px; }
    .summary-card { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; padding: 12px; }
    .summary-card strong { display: block; font-size: 20px; margin-top: 4px; }
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
      innerHeight: window.innerHeight || 0
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
function evaluate(report, context) {
  var expected = context && context.expected ? context.expected : {};
  var checks = [
    { name: '语言', status: matchExact(expected.language, report.locale.language) },
    { name: '语言列表', status: matchArrayPrefix(expected.acceptLanguage, report.locale.languages) },
    { name: '时区', status: matchExact(expected.timezone, report.locale.timezone) },
    { name: 'CPU 核心', status: matchExact(expected.hardwareConcurrency, report.hardware.hardwareConcurrency) },
    { name: '窗口大小', status: matchExact(expected.windowSize, report.screen.innerWidth + ',' + report.screen.innerHeight) },
    { name: '品牌', status: matchContains(expected.brand, report.identity.userAgent) },
    { name: '品牌版本', status: matchContains(expected.brandVersion, report.identity.userAgent) },
    { name: '平台版本', status: matchContains(expected.platformVersion, report.identity.userAgent) },
    { name: 'Webdriver', status: report.identity.webdriver ? 'mismatch' : 'match' },
    { name: 'WebRTC Host', status: report.network.localCandidateCount > 0 ? 'warning' : 'match' }
  ];
  checks.push({ name: 'Canvas Hash', status: report.advanced.canvasHash ? 'observed' : 'unknown' });
  checks.push({ name: 'Audio Hash', status: report.advanced.audioHash ? 'observed' : 'unknown' });
  checks.push({ name: 'ClientRects Hash', status: report.advanced.clientRectsHash ? 'observed' : 'unknown' });
  checks.push({ name: 'Fonts Hash', status: report.advanced.fontHash ? 'observed' : 'unknown' });
  checks.push({ name: 'WebGL Hash', status: report.advanced.webglHash ? 'observed' : 'unknown' });
  var mismatch = checks.filter(function (item) { return item.status === 'mismatch'; }).length;
  var warning = checks.filter(function (item) { return item.status === 'warning'; }).length;
  var match = checks.filter(function (item) { return item.status === 'match'; }).length;
  var observed = checks.filter(function (item) { return item.status === 'observed'; }).length;
  var status = mismatch > 0 ? '部分未生效' : warning > 0 ? '有风险' : '已生效';
  return { status: status, match: match, mismatch: mismatch, warning: warning, observed: observed, checks: checks, expected: expected };
}
function row(name, value, cls) {
  var display = Array.isArray(value) ? value.join(', ') : (typeof value === 'object' && value !== null ? JSON.stringify(value) : String(value));
  return '<tr><td>' + escapeHtml(name) + '</td><td class="' + (cls || '') + '"><code>' + escapeHtml(display || '-') + '</code></td></tr>';
}
function section(title, rows) { return '<section><h2>' + escapeHtml(title) + '</h2><table>' + rows.join('') + '</table></section>'; }
function escapeHtml(text) { return String(text).replace(/[&<>"']/g, function (m) { return ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'})[m]; }); }
function statusClass(status) {
  if (status === 'match' || status === 'observed') return 'ok';
  if (status === 'warning') return 'warn';
  if (status === 'mismatch') return 'bad';
  return '';
}
function statusText(status) {
  if (status === 'match') return '已生效';
  if (status === 'mismatch') return '未匹配';
  if (status === 'warning') return '有风险';
  if (status === 'observed') return '已采集';
  return '仅展示';
}
function render(report) {
  var verdict = evaluate(report, latestContext);
  report.verdict = verdict;
  var webdriverClass = report.identity.webdriver ? 'bad' : 'ok';
  var webrtcClass = report.network.localCandidateCount > 0 ? 'warn' : 'ok';
  document.getElementById('meta').textContent = '检测时间：' + report.generatedAt + (report.urlProfileId ? ' / Profile: ' + report.urlProfileId : '');
  document.getElementById('summary').innerHTML = [
    '<div class="summary-card">整体结论<strong class="' + (verdict.mismatch ? 'bad' : verdict.warning ? 'warn' : 'ok') + '">' + escapeHtml(verdict.status) + '</strong></div>',
    '<div class="summary-card">匹配项<strong class="ok">' + verdict.match + '</strong></div>',
    '<div class="summary-card">风险/不匹配<strong class="' + (verdict.mismatch ? 'bad' : verdict.warning ? 'warn' : 'ok') + '">' + (verdict.warning + verdict.mismatch) + '</strong></div>',
    '<div class="summary-card">Seed<strong>' + escapeHtml(verdict.expected.seed || '-') + '</strong></div>'
  ].join('');
  document.getElementById('app').innerHTML = [
    section('生效判定', verdict.checks.map(function (item) { return row(item.name, statusText(item.status), statusClass(item.status)); })),
    section('配置期望', [
      row('Language', verdict.expected.language || '-'), row('Accept-Language', verdict.expected.acceptLanguage || '-'), row('Timezone', verdict.expected.timezone || '-'), row('CPU', verdict.expected.hardwareConcurrency || '-'), row('Window', verdict.expected.windowSize || '-'), row('Brand', verdict.expected.brand || '-'), row('Platform', verdict.expected.platform || '-'), row('WebRTC', verdict.expected.webrtcPolicy || '-')
    ]),
    section('基础身份', [
      row('User-Agent', report.identity.userAgent), row('Platform', report.identity.platform), row('UA-CH', report.identity.userAgentData), row('Webdriver', report.identity.webdriver, webdriverClass)
    ]),
    section('语言与时区', [
      row('Language', report.locale.language), row('Languages', report.locale.languages), row('Timezone', report.locale.timezone), row('TZ Offset', report.locale.timezoneOffset)
    ]),
    section('硬件与屏幕', [
      row('CPU Cores', report.hardware.hardwareConcurrency), row('Device Memory', report.hardware.deviceMemory), row('Touch Points', report.hardware.maxTouchPoints), row('Screen', report.screen.width + 'x' + report.screen.height), row('Inner', report.screen.innerWidth + 'x' + report.screen.innerHeight), row('Color Depth', report.screen.colorDepth), row('DPR', report.screen.devicePixelRatio)
    ]),
    section('高级指纹', [
      row('Canvas Hash', report.advanced.canvasHash), row('Audio Hash', report.advanced.audioHash), row('ClientRects Hash', report.advanced.clientRectsHash), row('Font Hash', report.advanced.fontHash), row('Detected Fonts', report.advanced.detectedFonts), row('WebGL Vendor', report.advanced.webglVendor), row('WebGL Renderer', report.advanced.webglRenderer), row('WebGL Hash', report.advanced.webglHash)
    ]),
    section('插件与 MIME', [
      row('Plugins', report.advanced.plugins), row('MIME Types', report.advanced.mimeTypes)
    ]),
    section('WebRTC', [
      row('Host Candidates', report.network.localCandidateCount, webrtcClass), row('Candidates', report.network.webrtcCandidates)
    ])
  ].join('');
}
async function run() { latestReport = await collect(); render(latestReport); }
document.getElementById('refreshBtn').onclick = run;
document.getElementById('copyBtn').onclick = function () {
  if (!latestReport) return;
  navigator.clipboard.writeText(JSON.stringify(latestReport, null, 2)).catch(function () {});
};
run();
</script>
</body>
</html>`
