package browser

import (
	"ant-chrome/backend/internal/database"
	"path/filepath"
	"testing"
)

func TestSQLiteGroupDAODeleteRollsBackWhenMovingProfilesFails(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "groups.db"))
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	dao := NewSQLiteGroupDAO(db.GetConn())
	parent, err := dao.Create(GroupInput{GroupName: "parent"})
	if err != nil {
		t.Fatalf("Create parent returned error: %v", err)
	}
	target, err := dao.Create(GroupInput{GroupName: "target", ParentId: parent.GroupId})
	if err != nil {
		t.Fatalf("Create target returned error: %v", err)
	}
	child, err := dao.Create(GroupInput{GroupName: "child", ParentId: target.GroupId})
	if err != nil {
		t.Fatalf("Create child returned error: %v", err)
	}

	if _, err := db.GetConn().Exec(`
		INSERT INTO browser_profiles (profile_id, profile_name, group_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		"profile-in-target", "profile", target.GroupId, "2026-08-08T00:00:00Z", "2026-08-08T00:00:00Z"); err != nil {
		t.Fatalf("insert profile returned error: %v", err)
	}
	if _, err := db.GetConn().Exec(`
		CREATE TRIGGER fail_profile_group_move
		BEFORE UPDATE OF group_id ON browser_profiles
		BEGIN
			SELECT RAISE(ABORT, 'forced profile group move failure');
		END`); err != nil {
		t.Fatalf("create trigger returned error: %v", err)
	}

	if err := dao.Delete(target.GroupId); err == nil {
		t.Fatal("Delete returned nil error when moving profiles fails")
	}

	storedChild, err := dao.GetById(child.GroupId)
	if err != nil {
		t.Fatalf("GetById child returned error: %v", err)
	}
	if storedChild.ParentId != target.GroupId {
		t.Errorf("child ParentId = %q, want %q after rollback", storedChild.ParentId, target.GroupId)
	}

	var profileGroupID string
	if err := db.GetConn().QueryRow(`SELECT group_id FROM browser_profiles WHERE profile_id = ?`, "profile-in-target").Scan(&profileGroupID); err != nil {
		t.Fatalf("query profile group returned error: %v", err)
	}
	if profileGroupID != target.GroupId {
		t.Errorf("profile group_id = %q, want %q after rollback", profileGroupID, target.GroupId)
	}

	if _, err := dao.GetById(target.GroupId); err != nil {
		t.Errorf("target group was deleted despite rollback: %v", err)
	}
}
