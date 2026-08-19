package browser

import (
	"ant-chrome/backend/internal/database"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestSQLiteExtensionDAODeleteRemovesExtensionAndProfileAssociations(t *testing.T) {
	db, dao := newExtensionDAODeleteTestDB(t)
	const extensionID = "extension-delete-success"
	seedExtensionDAODeleteTestData(t, dao, extensionID, "profile-delete-success")

	if err := dao.Delete(extensionID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := dao.Get(extensionID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Get after delete error = %v, want sql.ErrNoRows", err)
	}
	if count := extensionDAOProfileAssociationCount(t, db, extensionID); count != 0 {
		t.Errorf("profile association count after delete = %d, want 0", count)
	}
}

func TestSQLiteExtensionDAODeleteRollsBackWhenProfileAssociationDeletionFails(t *testing.T) {
	db, dao := newExtensionDAODeleteTestDB(t)
	const extensionID = "extension-association-failure"
	seedExtensionDAODeleteTestData(t, dao, extensionID, "profile-association-failure")
	if _, err := db.GetConn().Exec(`
		CREATE TRIGGER fail_extension_profile_association_delete
		BEFORE DELETE ON browser_profile_extensions
		WHEN OLD.extension_id = 'extension-association-failure'
		BEGIN
			SELECT RAISE(ABORT, 'forced association deletion failure');
		END`); err != nil {
		t.Fatalf("create association deletion trigger returned error: %v", err)
	}

	if err := dao.Delete(extensionID); err == nil {
		t.Fatal("Delete returned nil error when profile association deletion failed")
	}
	if _, err := dao.Get(extensionID); err != nil {
		t.Fatalf("extension record was deleted despite association deletion failure: %v", err)
	}
	if count := extensionDAOProfileAssociationCount(t, db, extensionID); count != 1 {
		t.Errorf("profile association count after failed delete = %d, want 1", count)
	}
}

func TestSQLiteExtensionDAODeleteRollsBackWhenExtensionDeletionFails(t *testing.T) {
	db, dao := newExtensionDAODeleteTestDB(t)
	const extensionID = "extension-record-failure"
	seedExtensionDAODeleteTestData(t, dao, extensionID, "profile-record-failure")
	if _, err := db.GetConn().Exec(`
		CREATE TRIGGER fail_extension_record_delete
		BEFORE DELETE ON browser_extensions
		WHEN OLD.extension_id = 'extension-record-failure'
		BEGIN
			SELECT RAISE(ABORT, 'forced extension deletion failure');
		END`); err != nil {
		t.Fatalf("create extension deletion trigger returned error: %v", err)
	}

	if err := dao.Delete(extensionID); err == nil {
		t.Fatal("Delete returned nil error when extension deletion failed")
	}
	if _, err := dao.Get(extensionID); err != nil {
		t.Fatalf("extension record was deleted despite deletion failure: %v", err)
	}
	if count := extensionDAOProfileAssociationCount(t, db, extensionID); count != 1 {
		t.Errorf("profile association count after failed delete = %d, want 1", count)
	}
}

func newExtensionDAODeleteTestDB(t *testing.T) (*database.DB, *SQLiteExtensionDAO) {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), "extensions.db"))
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	return db, NewSQLiteExtensionDAO(db.GetConn())
}

func seedExtensionDAODeleteTestData(t *testing.T, dao *SQLiteExtensionDAO, extensionID string, profileID string) {
	t.Helper()
	if err := dao.Upsert(Extension{
		ExtensionID: extensionID,
		Name:        "test extension",
		InstallDir:  t.TempDir(),
		Enabled:     true,
	}); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	if _, err := dao.SetProfileSettings(profileID, []string{extensionID}, true); err != nil {
		t.Fatalf("SetProfileSettings returned error: %v", err)
	}
}

func extensionDAOProfileAssociationCount(t *testing.T, db *database.DB, extensionID string) int {
	t.Helper()
	var count int
	if err := db.GetConn().QueryRow(`SELECT COUNT(*) FROM browser_profile_extensions WHERE extension_id = ?`, extensionID).Scan(&count); err != nil {
		t.Fatalf("count profile associations returned error: %v", err)
	}
	return count
}
