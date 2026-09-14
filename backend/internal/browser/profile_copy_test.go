package browser

import (
	"ant-chrome/backend/internal/config"
	"testing"
)

func TestNormalizeCopyAutomationTargetsMapsLegacyTargetsToSeed(t *testing.T) {
	targets, err := normalizeCopyAutomationTargets([]string{"render", "fonts", "devices"})
	if err != nil {
		t.Fatalf("normalizeCopyAutomationTargets returned error: %v", err)
	}
	if len(targets) != 1 || targets[0] != copyAutomationTargetSeed {
		t.Fatalf("targets = %#v, want seed alias", targets)
	}
}

func TestBuildAutoFingerprintArgsKeepsLegacyFineGrainedArgs(t *testing.T) {
	sourceArgs := []string{
		"--fingerprint=111",
		"--fingerprint-canvas-noise=true",
		"--fingerprint-webgl-vendor=Intel",
	}
	defaultArgs := []string{"--fingerprint=222"}

	got := buildAutoFingerprintArgs(sourceArgs, defaultArgs, []string{copyAutomationTargetSeed})
	want := []string{"--fingerprint=222", "--fingerprint-canvas-noise=true", "--fingerprint-webgl-vendor=Intel"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d; got=%#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got[%d] = %q, want %q; got=%#v", index, got[index], want[index], got)
		}
	}
}

func TestCopyWithOptionsOverridesProxyConfig(t *testing.T) {
	manager := NewManager(&config.Config{}, t.TempDir())
	source, err := manager.Create(ProfileInput{
		ProfileName: "main",
		ProxyConfig: "socks5://127.0.0.1:1080",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	copied, err := manager.CopyWithOptions(source.ProfileId, "task", ProfileCopyOptions{
		Mode:        copyModeRegular,
		ProxyConfig: "http://127.0.0.1:18080",
	})
	if err != nil {
		t.Fatalf("CopyWithOptions: %v", err)
	}
	if copied.ProxyConfig != "http://127.0.0.1:18080" {
		t.Fatalf("proxyConfig = %q, want custom http proxy", copied.ProxyConfig)
	}
	if copied.ProxyId != "" {
		t.Fatalf("proxyId = %q, want empty custom proxy", copied.ProxyId)
	}
}

func TestCopyWithOptionsInheritsProxyWhenOmitted(t *testing.T) {
	manager := NewManager(&config.Config{}, t.TempDir())
	source, err := manager.Create(ProfileInput{
		ProfileName: "main",
		ProxyConfig: "socks5://127.0.0.1:1080",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	copied, err := manager.CopyWithOptions(source.ProfileId, "task", ProfileCopyOptions{Mode: copyModeRegular})
	if err != nil {
		t.Fatalf("CopyWithOptions: %v", err)
	}
	if copied.ProxyConfig != source.ProxyConfig {
		t.Fatalf("proxyConfig = %q, want inherited %q", copied.ProxyConfig, source.ProxyConfig)
	}
}
