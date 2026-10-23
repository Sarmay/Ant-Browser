package backend

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestNormalizeHostPlatformVersion(t *testing.T) {
	cases := map[string]string{
		"26.5":          "26.5.0",
		"26.5.1":        "26.5.1",
		"10.0.26100.1":  "10.0.26100.1",
		"15":            "15.0.0",
		"not-a-version": "",
	}
	for input, want := range cases {
		if got := normalizeHostPlatformVersion(input); got != want {
			t.Fatalf("normalizeHostPlatformVersion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeLocaleTag(t *testing.T) {
	cases := map[string]string{
		"zh_CN":       "zh-CN",
		"zh-Hans-CN":  "zh-CN",
		"en_US.UTF-8": "en-US",
		"ja-JP":       "ja-JP",
	}
	for input, want := range cases {
		if got := normalizeLocaleTag(input); got != want {
			t.Fatalf("normalizeLocaleTag(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDetectHostFingerprintPresetUsesCurrentDevice(t *testing.T) {
	preset := detectHostFingerprintPreset()
	switch runtime.GOOS {
	case "darwin":
		if preset.Platform != "macos" {
			t.Fatalf("platform = %q, want macos", preset.Platform)
		}
	case "windows":
		if preset.Platform != "windows" {
			t.Fatalf("platform = %q, want windows", preset.Platform)
		}
	default:
		if preset.Platform != "linux" {
			t.Fatalf("platform = %q, want linux", preset.Platform)
		}
	}
	if preset.PlatformVersion == "" || !dottedVersionPattern.MatchString(preset.PlatformVersion) {
		t.Fatalf("platform version = %q, want a dotted version", preset.PlatformVersion)
	}
	if preset.HardwareConcurrency != strconv.Itoa(runtime.NumCPU()) {
		t.Fatalf("cpu = %q, want %d", preset.HardwareConcurrency, runtime.NumCPU())
	}
	if preset.Language == "" {
		t.Fatal("language is empty")
	}
	if preset.AcceptLanguage == "" {
		t.Fatal("accept language is empty")
	}
	if runtime.GOOS == "darwin" && preset.Timezone == "" {
		t.Fatal("timezone is empty")
	}
}

func TestChromeVersionFromInstalledCoreTree(t *testing.T) {
	root := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ant-browser", "chrome", "Chromium")
	if _, err := os.Stat(root); err != nil {
		t.Skip(root)
	}
	got := chromeVersionFromCoreTree(root)
	if got == "" || !dottedVersionPattern.MatchString(got) {
		t.Fatalf("chrome version = %q", got)
	}
}
