package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ant-chrome/backend/internal/config"
)

func TestDeleteRejectsRunningProfileBeforeSoftDelete(t *testing.T) {
	manager := NewManager(&config.Config{}, t.TempDir())
	profile := &Profile{ProfileId: "running-profile", Running: true}
	dao := &profileDeleteMemoryDAO{profiles: map[string]*Profile{profile.ProfileId: profile}}
	manager.Profiles[profile.ProfileId] = profile
	manager.ProfileDAO = dao

	err := manager.Delete(profile.ProfileId)
	if err == nil {
		t.Fatal("Delete returned nil error for a running profile")
	}
	if !strings.Contains(err.Error(), "正在运行") || !strings.Contains(err.Error(), "先停止") {
		t.Fatalf("Delete error = %q, want a clear running-profile message", err)
	}
	if dao.softDeleteCalls != 0 {
		t.Fatalf("SoftDelete calls = %d, want 0", dao.softDeleteCalls)
	}
	if profile.DeletedAt != "" {
		t.Fatalf("profile.DeletedAt = %q, want empty", profile.DeletedAt)
	}
	if _, exists := manager.Profiles[profile.ProfileId]; !exists {
		t.Fatal("running profile was removed from manager state")
	}
}

func TestDeleteSoftDeletesStoppedProfileWithoutRemovingUserData(t *testing.T) {
	appRoot := t.TempDir()
	manager := NewManager(&config.Config{}, appRoot)
	manager.Config.Browser.UserDataRoot = "data"
	profile := &Profile{ProfileId: "stopped-profile", UserDataDir: "stopped-profile"}
	dao := &profileDeleteMemoryDAO{profiles: map[string]*Profile{profile.ProfileId: profile}}
	manager.Profiles[profile.ProfileId] = profile
	manager.ProfileDAO = dao

	userDataDir := filepath.Join(appRoot, "data", profile.UserDataDir)
	userDataFile := filepath.Join(userDataDir, "Default", "Preferences")
	if err := os.MkdirAll(filepath.Dir(userDataFile), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(userDataFile, []byte("user data"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := manager.Delete(profile.ProfileId); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if dao.softDeleteCalls != 1 {
		t.Fatalf("SoftDelete calls = %d, want 1", dao.softDeleteCalls)
	}
	if profile.DeletedAt == "" {
		t.Fatal("stopped profile was not soft deleted")
	}
	if _, exists := manager.Profiles[profile.ProfileId]; exists {
		t.Fatal("soft-deleted profile remains in manager state")
	}
	data, err := os.ReadFile(userDataFile)
	if err != nil {
		t.Fatalf("ReadFile user data returned error: %v", err)
	}
	if string(data) != "user data" {
		t.Fatalf("user data = %q, want %q", data, "user data")
	}
}

type profileDeleteMemoryDAO struct {
	profiles        map[string]*Profile
	softDeleteCalls int
}

func (d *profileDeleteMemoryDAO) List() ([]*Profile, error) { return nil, nil }

func (d *profileDeleteMemoryDAO) ListDeleted() ([]*Profile, error) { return nil, nil }

func (d *profileDeleteMemoryDAO) GetById(profileID string) (*Profile, error) {
	profile, exists := d.profiles[profileID]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}
	return profile, nil
}

func (d *profileDeleteMemoryDAO) Upsert(profile *Profile) error {
	d.profiles[profile.ProfileId] = profile
	return nil
}

func (d *profileDeleteMemoryDAO) Delete(profileID string) error {
	delete(d.profiles, profileID)
	return nil
}

func (d *profileDeleteMemoryDAO) SoftDelete(profileID string, deletedAt string) error {
	profile, exists := d.profiles[profileID]
	if !exists {
		return fmt.Errorf("profile not found: %s", profileID)
	}
	d.softDeleteCalls++
	profile.DeletedAt = deletedAt
	profile.UpdatedAt = deletedAt
	return nil
}

func (d *profileDeleteMemoryDAO) Restore(profileID string) error {
	profile, exists := d.profiles[profileID]
	if !exists {
		return fmt.Errorf("profile not found: %s", profileID)
	}
	profile.DeletedAt = ""
	return nil
}

func (d *profileDeleteMemoryDAO) ListExpiredDeleted(string) ([]*Profile, error) { return nil, nil }
