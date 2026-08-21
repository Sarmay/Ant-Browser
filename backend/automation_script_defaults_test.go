package backend

import (
	"os"
	"testing"
)

func TestEnsureAutomationScriptDefaults(t *testing.T) {
	tempRoot := t.TempDir()
	app := NewApp(tempRoot)

	// Step 1: Initial list on fresh store should seed all 7 default scripts and create marker file
	scripts, err := app.AutomationScriptList()
	if err != nil {
		t.Fatalf("AutomationScriptList() failed: %v", err)
	}

	if len(scripts) != 7 {
		t.Fatalf("expected 7 default scripts, got %d", len(scripts))
	}

	markerPath := app.automationScriptDefaultsMarkerPath(automationScriptDefaultsMarkerName)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Fatalf("expected defaults marker file %q to exist", markerPath)
	}

	// Step 2: Calling again with marker present should return the same list without error
	scriptsAgain, err := app.AutomationScriptList()
	if err != nil {
		t.Fatalf("AutomationScriptList() second call failed: %v", err)
	}
	if len(scriptsAgain) != 7 {
		t.Fatalf("expected 7 scripts on second call, got %d", len(scriptsAgain))
	}

	// Step 3: Verify script details
	expectedIDs := map[string]bool{
		"dual-instance-runtime-switch": true,
		"news-query-txt":               true,
		"proton-mail-first-message":    true,
		"web-image-generate-download":  true,
		"lianjia-wh-home-step1":        true,
		"lianjia-wh-cookie-prepare":    true,
		"beike-house-price-extract":    true,
	}

	for _, s := range scripts {
		if !expectedIDs[s.ID] {
			t.Errorf("unexpected script ID %q", s.ID)
		}
		if s.ScriptText == "" {
			t.Errorf("script %q has empty ScriptText", s.ID)
		}
	}
}

func TestEnsureAutomationScriptDefaultsMigrationAddsMissing(t *testing.T) {
	tempRoot := t.TempDir()
	app := NewApp(tempRoot)

	store := app.automationScriptStore()

	// Pretend only one legacy script was seeded, with a legacy marker
	defaults, err := app.AutomationScriptList()
	if err != nil {
		t.Fatalf("initial list failed: %v", err)
	}
	if len(defaults) != 7 {
		t.Fatalf("expected 7 defaults, got %d", len(defaults))
	}

	// Remove current marker and set a legacy marker
	currMarker := app.automationScriptDefaultsMarkerPath(automationScriptDefaultsMarkerName)
	_ = os.Remove(currMarker)

	legacyMarker := app.automationScriptDefaultsMarkerPath("defaults-seeded-v11")
	_ = os.WriteFile(legacyMarker, []byte("ok\n"), 0o644)

	// Delete one of the scripts from the store
	if err := store.Delete("beike-house-price-extract"); err != nil {
		t.Fatalf("failed to delete script: %v", err)
	}

	// Call ensureAutomationScriptDefaults again
	if err := app.ensureAutomationScriptDefaults(store); err != nil {
		t.Fatalf("ensureAutomationScriptDefaults failed: %v", err)
	}

	// The missing script should be re-imported
	reloaded, err := store.Get("beike-house-price-extract")
	if err != nil {
		t.Fatalf("expected beike-house-price-extract to be restored, got err: %v", err)
	}
	if reloaded.ID != "beike-house-price-extract" {
		t.Fatalf("unexpected reloaded ID %q", reloaded.ID)
	}
}
