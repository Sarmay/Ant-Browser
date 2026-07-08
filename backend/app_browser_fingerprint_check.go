package backend

import (
	"encoding/json"
	"fmt"
	"strings"
)

type BrowserFingerprintCheckResult struct {
	ProfileId string                         `json:"profileId"`
	Runtime   BrowserFingerprintRuntimeInfo  `json:"runtime"`
	Expected  BrowserFingerprintExpectedInfo `json:"expected"`
}

type BrowserFingerprintRuntimeInfo struct {
	Language            string   `json:"language"`
	Languages           []string `json:"languages"`
	Timezone            string   `json:"timezone"`
	HardwareConcurrency int      `json:"hardwareConcurrency"`
	DeviceMemory        float64  `json:"deviceMemory"`
	Platform            string   `json:"platform"`
	UserAgent           string   `json:"userAgent"`
	UserAgentData       string   `json:"userAgentData"`
	Webdriver           bool     `json:"webdriver"`
	ScreenWidth         int      `json:"screenWidth"`
	ScreenHeight        int      `json:"screenHeight"`
	ColorDepth          int      `json:"colorDepth"`
	InnerWidth          int      `json:"innerWidth"`
	InnerHeight         int      `json:"innerHeight"`
	DevicePixelRatio    float64  `json:"devicePixelRatio"`
	WebGLVendor         string   `json:"webglVendor"`
	WebGLRenderer       string   `json:"webglRenderer"`
	CanvasHash          string   `json:"canvasHash"`
	AudioHash           string   `json:"audioHash"`
	ClientRectsHash     string   `json:"clientRectsHash"`
	Plugins             []string `json:"plugins"`
}

type BrowserFingerprintExpectedInfo struct {
	Language            string `json:"language"`
	AcceptLanguage      string `json:"acceptLanguage"`
	Timezone            string `json:"timezone"`
	HardwareConcurrency string `json:"hardwareConcurrency"`
	WindowSize          string `json:"windowSize"`
	Brand               string `json:"brand"`
	BrandVersion        string `json:"brandVersion"`
	Platform            string `json:"platform"`
	PlatformVersion     string `json:"platformVersion"`
	Seed                string `json:"seed"`
	DisableSpoofing     string `json:"disableSpoofing"`
	WebRTCPolicy        string `json:"webrtcPolicy"`
}

func (a *App) BrowserProfileFingerprintCheck(profileId string) (*BrowserFingerprintCheckResult, error) {
	profileId = strings.TrimSpace(profileId)
	if profileId == "" {
		return nil, fmt.Errorf("实例 ID 不能为空")
	}

	var profile *BrowserProfile
	for _, item := range a.browserMgr.List() {
		if item.ProfileId == profileId {
			current := item
			profile = &current
			break
		}
	}
	if profile == nil {
		return nil, fmt.Errorf("实例不存在: %s", profileId)
	}
	if !profile.Running || !profile.DebugReady || profile.DebugPort <= 0 {
		return nil, fmt.Errorf("实例未处于可自测状态，请先启动实例并等待调试端口就绪")
	}
	if err := probeBrowserDebugPort(profile.DebugPort, browserDebugProbeTimeout); err != nil {
		return nil, fmt.Errorf("实例调试端口不可用，请重启实例后再试: %w", err)
	}

	runtimeInfo, err := evaluateBrowserFingerprintRuntime(profile.DebugPort)
	if err != nil {
		return nil, err
	}
	return &BrowserFingerprintCheckResult{
		ProfileId: profile.ProfileId,
		Runtime:   runtimeInfo,
		Expected:  buildBrowserFingerprintExpected(profile.FingerprintArgs),
	}, nil
}

func evaluateBrowserFingerprintRuntime(debugPort int) (BrowserFingerprintRuntimeInfo, error) {
	const expression = `(async function(){
  function hashString(input) {
    var hash = 2166136261;
    var text = String(input || '');
    for (var i = 0; i < text.length; i++) {
      hash ^= text.charCodeAt(i);
      hash += (hash << 1) + (hash << 4) + (hash << 7) + (hash << 8) + (hash << 24);
    }
    return ('00000000' + (hash >>> 0).toString(16)).slice(-8);
  }
  function canvasHash() {
    var canvas = document.createElement('canvas');
    canvas.width = 280;
    canvas.height = 80;
    var ctx = canvas.getContext('2d');
    if (!ctx) return '';
    ctx.textBaseline = 'top';
    ctx.font = '16px Arial';
    ctx.fillStyle = '#f60';
    ctx.fillRect(4, 4, 120, 32);
    ctx.fillStyle = '#069';
    ctx.fillText('ant fingerprint 自测', 8, 10);
    ctx.strokeStyle = 'rgba(120, 60, 200, .8)';
    ctx.arc(160, 40, 28, 0, Math.PI * 2, true);
    ctx.stroke();
    return hashString(canvas.toDataURL());
  }
  async function audioHash() {
    try {
      var AudioContextCtor = window.OfflineAudioContext || window.webkitOfflineAudioContext;
      if (!AudioContextCtor) return '';
      var context = new AudioContextCtor(1, 44100, 44100);
      var oscillator = context.createOscillator();
      var compressor = context.createDynamicsCompressor();
      oscillator.type = 'triangle';
      oscillator.frequency.value = 10000;
      compressor.threshold.value = -50;
      compressor.knee.value = 40;
      compressor.ratio.value = 12;
      compressor.attack.value = 0;
      compressor.release.value = 0.25;
      oscillator.connect(compressor);
      compressor.connect(context.destination);
      oscillator.start(0);
      var buffer = await context.startRendering();
      var data = buffer.getChannelData(0).slice(4500, 5000);
      return hashString(Array.prototype.map.call(data, function (value) { return value.toFixed(6); }).join(','));
    } catch (e) {
      return '';
    }
  }
  function clientRectsHash() {
    var node = document.createElement('div');
    node.style.cssText = 'position:absolute;left:-9999px;top:-9999px;width:180px;font:13px Arial;line-height:17px;';
    node.textContent = 'ant fingerprint client rects check';
    document.body.appendChild(node);
    var rects = Array.prototype.map.call(node.getClientRects(), function (rect) {
      return [rect.x, rect.y, rect.width, rect.height].map(function (value) { return Number(value).toFixed(3); }).join(':');
    }).join('|');
    document.body.removeChild(node);
    return hashString(rects);
  }
  var gl = document.createElement('canvas').getContext('webgl') || document.createElement('canvas').getContext('experimental-webgl');
  var debug = gl && gl.getExtension('WEBGL_debug_renderer_info');
  return JSON.stringify({
    language: navigator.language || '',
    languages: Array.prototype.slice.call(navigator.languages || []),
    timezone: (Intl.DateTimeFormat().resolvedOptions() || {}).timeZone || '',
    hardwareConcurrency: navigator.hardwareConcurrency || 0,
    deviceMemory: navigator.deviceMemory || 0,
    platform: navigator.platform || '',
    userAgent: navigator.userAgent || '',
    userAgentData: navigator.userAgentData ? JSON.stringify({brands: navigator.userAgentData.brands || [], mobile: navigator.userAgentData.mobile, platform: navigator.userAgentData.platform || ''}) : '',
    webdriver: navigator.webdriver === true,
    screenWidth: screen.width || 0,
    screenHeight: screen.height || 0,
    colorDepth: screen.colorDepth || 0,
    innerWidth: window.innerWidth || 0,
    innerHeight: window.innerHeight || 0,
    devicePixelRatio: window.devicePixelRatio || 0,
    webglVendor: debug ? gl.getParameter(debug.UNMASKED_VENDOR_WEBGL) : '',
    webglRenderer: debug ? gl.getParameter(debug.UNMASKED_RENDERER_WEBGL) : '',
    canvasHash: canvasHash(),
    audioHash: await audioHash(),
    clientRectsHash: clientRectsHash(),
    plugins: Array.prototype.map.call(navigator.plugins || [], function (plugin) { return plugin.name || ''; })
  });
})()`

	result, err := cdpCall(debugPort, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":   true,
		"timeout":       3000,
	})
	if err != nil {
		return BrowserFingerprintRuntimeInfo{}, fmt.Errorf("指纹自测失败: %w", err)
	}

	value, ok := nestedString(result, "result", "value")
	if !ok || strings.TrimSpace(value) == "" {
		return BrowserFingerprintRuntimeInfo{}, fmt.Errorf("指纹自测失败：CDP 未返回有效结果")
	}
	var runtimeInfo BrowserFingerprintRuntimeInfo
	if err := json.Unmarshal([]byte(value), &runtimeInfo); err != nil {
		return BrowserFingerprintRuntimeInfo{}, fmt.Errorf("指纹自测结果解析失败: %w", err)
	}
	return runtimeInfo, nil
}

func buildBrowserFingerprintExpected(args []string) BrowserFingerprintExpectedInfo {
	normalizedArgs := normalizeBrowserLanguageArgs(args)
	return BrowserFingerprintExpectedInfo{
		Language:            browserArgValue(normalizedArgs, "--lang"),
		AcceptLanguage:      browserArgValue(normalizedArgs, "--accept-lang"),
		Timezone:            browserArgValue(normalizedArgs, "--timezone"),
		HardwareConcurrency: browserArgValue(normalizedArgs, "--fingerprint-hardware-concurrency"),
		WindowSize:          browserArgValue(normalizedArgs, "--window-size"),
		Brand:               browserArgValue(normalizedArgs, "--fingerprint-brand"),
		BrandVersion:        browserArgValue(normalizedArgs, "--fingerprint-brand-version"),
		Platform:            browserArgValue(normalizedArgs, "--fingerprint-platform"),
		PlatformVersion:     browserArgValue(normalizedArgs, "--fingerprint-platform-version"),
		Seed:                browserArgValue(normalizedArgs, "--fingerprint"),
		DisableSpoofing:     browserArgValue(normalizedArgs, "--disable-spoofing"),
		WebRTCPolicy:        browserWebRTCPolicy(normalizedArgs),
	}
}

func browserWebRTCPolicy(args []string) string {
	for _, arg := range args {
		if strings.TrimSpace(arg) == "--disable-non-proxied-udp" {
			return "disable_non_proxied_udp"
		}
	}
	return browserArgValue(args, "--webrtc-ip-handling-policy")
}

func nestedString(root map[string]any, keys ...string) (string, bool) {
	var current any = root
	for _, key := range keys {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = object[key]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}
