package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeBrowserLanguageArgsCompletesAcceptLanguage(t *testing.T) {
	args := normalizeBrowserLanguageArgs([]string{"--fingerprint-brand=Chrome", "--lang=ja-JP"})
	if got := browserArgValue(args, browserAcceptLangArg); got != "ja-JP,ja" {
		t.Fatalf("accept language = %q, want %q", got, "ja-JP,ja")
	}
	if got := browserArgValue(args, browserLangArg); got != "ja-JP" {
		t.Fatalf("lang = %q, want %q", got, "ja-JP")
	}
}

func TestNormalizeBrowserLanguageArgsInfersLang(t *testing.T) {
	args := normalizeBrowserLanguageArgs([]string{"--accept-lang=en-US,en"})
	if got := browserArgValue(args, browserLangArg); got != "en-US" {
		t.Fatalf("lang = %q, want %q", got, "en-US")
	}
}

func TestWriteBrowserLanguagePreferences(t *testing.T) {
	root := t.TempDir()
	prefsDir := filepath.Join(root, "Default")
	if err := os.MkdirAll(prefsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prefsDir, "Preferences"), []byte(`{"profile":{"name":"Default"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := writeBrowserLanguagePreferences(root, []string{"--lang=zh-CN"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(prefsDir, "Preferences"))
	if err != nil {
		t.Fatal(err)
	}
	var prefs map[string]interface{}
	if err := json.Unmarshal(data, &prefs); err != nil {
		t.Fatal(err)
	}
	intl, ok := prefs["intl"].(map[string]interface{})
	if !ok {
		t.Fatalf("intl missing: %#v", prefs)
	}
	want := map[string]interface{}{
		"accept_languages": "zh-CN,zh",
	}
	for key, value := range want {
		if !reflect.DeepEqual(intl[key], value) {
			t.Fatalf("intl[%s] = %#v, want %#v", key, intl[key], value)
		}
	}
	if _, ok := prefs["profile"].(map[string]interface{}); !ok {
		t.Fatalf("existing preferences were not preserved: %#v", prefs)
	}
}
