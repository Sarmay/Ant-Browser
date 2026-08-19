package backend

import (
	"ant-chrome/backend/internal/browser"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func newSnapshotTestApp(t *testing.T) (*App, string) {
	t.Helper()

	root := t.TempDir()
	app := NewApp(root)
	app.browserMgr = browser.NewManager(nil, root)
	profileID := "profile-1"
	app.browserMgr.Profiles[profileID] = &browser.Profile{ProfileId: profileID}
	return app, profileID
}

func writeAppSnapshotFiles(t *testing.T, app *App, profileID, snapshotID, name string) (string, string) {
	t.Helper()

	dir, err := app.snapshotDir(profileID)
	if err != nil {
		t.Fatalf("create snapshot directory: %v", err)
	}
	metaPath := filepath.Join(dir, snapshotID+"_"+name+".meta.json")
	zipPath := filepath.Join(dir, snapshotID+"_"+name+".zip")
	if err := os.WriteFile(metaPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write meta file: %v", err)
	}
	if err := os.WriteFile(zipPath, []byte("zip"), 0o644); err != nil {
		t.Fatalf("write zip file: %v", err)
	}
	return metaPath, zipPath
}

func TestBrowserSnapshotDeleteRemovesOnlyRequestedSnapshot(t *testing.T) {
	app, profileID := newSnapshotTestApp(t)
	metaPath, zipPath := writeAppSnapshotFiles(t, app, profileID, "snapshot-a", "current")
	otherMetaPath, otherZipPath := writeAppSnapshotFiles(t, app, profileID, "snapshot-a-extra", "other")

	if err := app.BrowserSnapshotDelete(profileID, "snapshot-a"); err != nil {
		t.Fatalf("BrowserSnapshotDelete() error = %v", err)
	}
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Fatalf("requested meta file still exists or cannot be checked: %v", err)
	}
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Fatalf("requested zip file still exists or cannot be checked: %v", err)
	}
	if _, err := os.Stat(otherMetaPath); err != nil {
		t.Fatalf("other meta file was changed: %v", err)
	}
	if _, err := os.Stat(otherZipPath); err != nil {
		t.Fatalf("other zip file was changed: %v", err)
	}
}

func TestBrowserSnapshotDeleteRejectsUnsafeSnapshotIDs(t *testing.T) {
	app, profileID := newSnapshotTestApp(t)
	absID := filepath.Join(t.TempDir(), "snapshot")
	for _, snapshotID := range []string{"", absID, "..", "snapshot/child", `snapshot\child`, "snapshot..id"} {
		t.Run(snapshotID, func(t *testing.T) {
			if err := app.BrowserSnapshotDelete(profileID, snapshotID); err == nil {
				t.Fatal("BrowserSnapshotDelete() error = nil, want invalid snapshot ID to be rejected")
			}
		})
	}
}

func TestBrowserSnapshotDeleteRequiresExistingProfile(t *testing.T) {
	app, profileID := newSnapshotTestApp(t)
	metaPath, zipPath := writeAppSnapshotFiles(t, app, profileID, "snapshot-a", "current")

	if err := app.BrowserSnapshotDelete("missing-profile", "snapshot-a"); err == nil {
		t.Fatal("BrowserSnapshotDelete() error = nil, want missing profile to be rejected")
	}
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("meta file was changed for missing profile: %v", err)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("zip file was changed for missing profile: %v", err)
	}
}

func TestBrowserSnapshotDeletePropagatesRemoveError(t *testing.T) {
	app, profileID := newSnapshotTestApp(t)
	metaPath, zipPath := writeAppSnapshotFiles(t, app, profileID, "snapshot-a", "current")
	if err := os.Remove(zipPath); err != nil {
		t.Fatalf("remove zip file before setup: %v", err)
	}
	if err := os.Mkdir(zipPath, 0o755); err != nil {
		t.Fatalf("create zip directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(zipPath, "contents"), []byte("locked"), 0o644); err != nil {
		t.Fatalf("write zip directory contents: %v", err)
	}

	err := app.BrowserSnapshotDelete(profileID, "snapshot-a")
	if err == nil {
		t.Fatal("BrowserSnapshotDelete() error = nil, want os.Remove failure")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || pathErr.Path != zipPath {
		t.Fatalf("BrowserSnapshotDelete() error = %v, want os.Remove error for %q", err, zipPath)
	}
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("meta file was removed after zip removal failed: %v", err)
	}
}
