package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/database"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserExtensionDeleteRemovesDirectoryBeforeDatabaseRecords(t *testing.T) {
	app, db, dao := newExtensionDeleteTestApp(t)
	const extensionID = "extension-delete-success"
	const profileID = "profile-delete-success"
	installDir := seedExtensionDeleteTestData(t, app.appRoot, dao, extensionID, profileID)

	originalRemoveAll := removeExtensionInstallDir
	removeExtensionInstallDir = func(path string) error {
		if path != installDir {
			t.Errorf("remove path = %q, want %q", path, installDir)
		}
		if _, err := dao.Get(extensionID); err != nil {
			t.Errorf("extension was deleted before its directory: %v", err)
		}
		if count := extensionProfileAssociationCount(t, db, extensionID); count != 1 {
			t.Errorf("profile association count before directory deletion = %d, want 1", count)
		}
		return originalRemoveAll(path)
	}
	t.Cleanup(func() { removeExtensionInstallDir = originalRemoveAll })

	if err := app.BrowserExtensionDelete(extensionID); err != nil {
		t.Fatalf("BrowserExtensionDelete returned error: %v", err)
	}
	if _, err := os.Stat(installDir); !os.IsNotExist(err) {
		t.Fatalf("install directory still exists or cannot be checked: %v", err)
	}
	if _, err := dao.Get(extensionID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Get after delete error = %v, want sql.ErrNoRows", err)
	}
	if count := extensionProfileAssociationCount(t, db, extensionID); count != 0 {
		t.Errorf("profile association count after delete = %d, want 0", count)
	}
}

func TestBrowserExtensionDeleteKeepsDatabaseRecordsWhenDirectoryDeletionFails(t *testing.T) {
	app, db, dao := newExtensionDeleteTestApp(t)
	const extensionID = "extension-delete-failure"
	const profileID = "profile-delete-failure"
	installDir := seedExtensionDeleteTestData(t, app.appRoot, dao, extensionID, profileID)

	originalRemoveAll := removeExtensionInstallDir
	removeExtensionInstallDir = func(path string) error {
		if path != installDir {
			t.Errorf("remove path = %q, want %q", path, installDir)
		}
		return errors.New("forced extension directory deletion failure")
	}
	t.Cleanup(func() { removeExtensionInstallDir = originalRemoveAll })

	if err := app.BrowserExtensionDelete(extensionID); err == nil {
		t.Fatal("BrowserExtensionDelete returned nil error when directory deletion failed")
	}
	if _, err := os.Stat(installDir); err != nil {
		t.Fatalf("install directory was removed despite deletion failure: %v", err)
	}
	if _, err := dao.Get(extensionID); err != nil {
		t.Fatalf("extension record was deleted despite directory deletion failure: %v", err)
	}
	if count := extensionProfileAssociationCount(t, db, extensionID); count != 1 {
		t.Errorf("profile association count after directory deletion failure = %d, want 1", count)
	}
}

func newExtensionDeleteTestApp(t *testing.T) (*App, *database.DB, *browser.SQLiteExtensionDAO) {
	t.Helper()
	appRoot := t.TempDir()
	db, err := database.NewDB(filepath.Join(appRoot, "extensions.db"))
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	dao := browser.NewSQLiteExtensionDAO(db.GetConn())
	app := NewApp(appRoot)
	app.browserMgr = browser.NewManager(nil, appRoot)
	app.browserMgr.ExtensionDAO = dao
	return app, db, dao
}

func seedExtensionDeleteTestData(t *testing.T, appRoot string, dao *browser.SQLiteExtensionDAO, extensionID string, profileID string) string {
	t.Helper()
	installDir := filepath.Join(appRoot, "data", "extensions", extensionID)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatalf("MkdirAll install directory returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "manifest.json"), []byte(`{"name":"test extension"}`), 0o644); err != nil {
		t.Fatalf("WriteFile manifest returned error: %v", err)
	}
	if err := dao.Upsert(browser.Extension{
		ExtensionID: extensionID,
		Name:        "test extension",
		InstallDir:  installDir,
		Enabled:     true,
	}); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	if _, err := dao.SetProfileSettings(profileID, []string{extensionID}, true); err != nil {
		t.Fatalf("SetProfileSettings returned error: %v", err)
	}
	return installDir
}

func extensionProfileAssociationCount(t *testing.T, db *database.DB, extensionID string) int {
	t.Helper()
	var count int
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profile_extensions WHERE extension_id = ?`, extensionID).Scan(&count); err != nil {
		t.Fatalf("count profile associations returned error: %v", err)
	}
	return count
}
