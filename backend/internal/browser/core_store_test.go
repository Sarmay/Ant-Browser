package browser

import (
	"path/filepath"
	"strings"
	"testing"

	"ant-chrome/backend/internal/config"
	"ant-chrome/backend/internal/database"
)

func TestDeleteCoreGuardsAreConsistentAcrossStores(t *testing.T) {
	stores := []struct {
		name  string
		setup func(*testing.T, []Core) *Manager
	}{
		{name: "config", setup: newConfigCoreStoreManager},
		{name: "sqlite", setup: newSQLiteCoreStoreManager},
	}

	tests := []struct {
		name        string
		cores       []Core
		profiles    []config.BrowserProfileConfig
		deleteID    string
		wantError   string
		wantCoreIDs []string
	}{
		{
			name:        "rejects missing core",
			cores:       testCores(),
			deleteID:    "missing",
			wantError:   "内核不存在",
			wantCoreIDs: []string{"default", "removable"},
		},
		{
			name: "rejects only core",
			cores: []Core{
				{CoreId: "only", CoreName: "唯一内核", CorePath: "/cores/only", IsDefault: true},
			},
			deleteID:    "only",
			wantError:   "唯一内核",
			wantCoreIDs: []string{"only"},
		},
		{
			name:        "rejects default core",
			cores:       testCores(),
			deleteID:    "default",
			wantError:   "默认内核",
			wantCoreIDs: []string{"default", "removable"},
		},
		{
			name:        "rejects referenced core",
			cores:       testCores(),
			profiles:    []config.BrowserProfileConfig{{ProfileId: "profile-1", CoreId: "removable"}},
			deleteID:    "removable",
			wantError:   "1 个实例引用",
			wantCoreIDs: []string{"default", "removable"},
		},
		{
			name:        "deletes unreferenced non-default core",
			cores:       testCores(),
			deleteID:    "removable",
			wantCoreIDs: []string{"default"},
		},
	}

	for _, store := range stores {
		for _, test := range tests {
			t.Run(store.name+"/"+test.name, func(t *testing.T) {
				manager := store.setup(t, test.cores)
				manager.Config.Browser.Profiles = test.profiles

				err := manager.DeleteCore(test.deleteID)
				if test.wantError == "" {
					if err != nil {
						t.Fatalf("DeleteCore returned error: %v", err)
					}
				} else {
					if err == nil || !strings.Contains(err.Error(), test.wantError) {
						t.Fatalf("DeleteCore error = %v, want message containing %q", err, test.wantError)
					}
				}

				assertCoreIDs(t, manager.ListCores(), test.wantCoreIDs)
			})
		}
	}
}

func newConfigCoreStoreManager(t *testing.T, cores []Core) *Manager {
	t.Helper()
	manager := NewManager(&config.Config{}, t.TempDir())
	manager.Config.Browser.Cores = append([]Core(nil), cores...)
	return manager
}

func newSQLiteCoreStoreManager(t *testing.T, cores []Core) *Manager {
	t.Helper()
	appRoot := t.TempDir()
	db, err := database.NewDB(filepath.Join(appRoot, "cores.db"))
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	manager := NewManager(&config.Config{}, appRoot)
	manager.CoreDAO = NewSQLiteCoreDAO(db.GetConn())
	for _, core := range cores {
		if err := manager.CoreDAO.Upsert(core); err != nil {
			t.Fatalf("Upsert returned error: %v", err)
		}
	}
	return manager
}

func testCores() []Core {
	return []Core{
		{CoreId: "default", CoreName: "默认内核", CorePath: "/cores/default", IsDefault: true},
		{CoreId: "removable", CoreName: "可删除内核", CorePath: "/cores/removable"},
	}
}

func assertCoreIDs(t *testing.T, cores []Core, want []string) {
	t.Helper()
	if len(cores) != len(want) {
		t.Fatalf("core count = %d, want %d (%v)", len(cores), len(want), want)
	}
	for i, core := range cores {
		if core.CoreId != want[i] {
			t.Fatalf("core[%d].CoreId = %q, want %q", i, core.CoreId, want[i])
		}
	}
}
