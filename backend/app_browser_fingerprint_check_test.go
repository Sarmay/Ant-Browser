package backend

import "testing"

func TestBuildBrowserFingerprintExpected(t *testing.T) {
	expected := buildBrowserFingerprintExpected([]string{
		"--fingerprint=12345",
		"--fingerprint-brand=Chrome",
		"--fingerprint-brand-version=144.0.7559.132",
		"--fingerprint-platform=windows",
		"--fingerprint-platform-version=10.0.0",
		"--lang=ja-JP",
		"--timezone=Asia/Tokyo",
		"--fingerprint-hardware-concurrency=3",
		"--window-size=1111,777",
		"--disable-non-proxied-udp",
		"--disable-spoofing=font,gpu",
	})

	if expected.Language != "ja-JP" {
		t.Fatalf("language = %q", expected.Language)
	}
	if expected.AcceptLanguage != "ja-JP,ja" {
		t.Fatalf("acceptLanguage = %q", expected.AcceptLanguage)
	}
	if expected.Timezone != "Asia/Tokyo" {
		t.Fatalf("timezone = %q", expected.Timezone)
	}
	if expected.HardwareConcurrency != "3" {
		t.Fatalf("hardwareConcurrency = %q", expected.HardwareConcurrency)
	}
	if expected.WindowSize != "1111,777" {
		t.Fatalf("windowSize = %q", expected.WindowSize)
	}
	if expected.Brand != "Chrome" || expected.Platform != "windows" {
		t.Fatalf("identity = %q / %q", expected.Brand, expected.Platform)
	}
	if expected.BrandVersion != "144.0.7559.132" || expected.PlatformVersion != "10.0.0" {
		t.Fatalf("identity versions = %q / %q", expected.BrandVersion, expected.PlatformVersion)
	}
	if expected.Seed != "12345" {
		t.Fatalf("seed = %q", expected.Seed)
	}
	if expected.WebRTCPolicy != "disable_non_proxied_udp" {
		t.Fatalf("webrtcPolicy = %q", expected.WebRTCPolicy)
	}
	if expected.DisableSpoofing != "font,gpu" {
		t.Fatalf("disableSpoofing = %q", expected.DisableSpoofing)
	}
}
