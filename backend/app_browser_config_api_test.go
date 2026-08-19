package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDarwinAppBundleRoot(t *testing.T) {
	root := t.TempDir()
	appBundle := filepath.Join(root, "Fingerprint Chromium.app")
	executable := filepath.Join(appBundle, "Contents", "MacOS", "Chromium")

	if got := darwinAppBundleRoot(executable); got != appBundle {
		t.Fatalf("应用包根目录错误: got=%q want=%q", got, appBundle)
	}
	if got := darwinAppBundleRoot(filepath.Join(root, "chrome")); got != "" {
		t.Fatalf("普通目录不应识别为应用包: %q", got)
	}
}

func TestPathWithinDirectory(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "Chromium.app")
	if !pathWithinDirectory(inside, root) {
		t.Fatal("子目录应被识别为根目录内路径")
	}
	if pathWithinDirectory(filepath.Dir(root), root) {
		t.Fatal("父目录不应被识别为根目录内路径")
	}
}

func TestValidateDarwinMountedAppBundleRejectsMountEscapes(t *testing.T) {
	appBundle := "/Volumes/Installer/Fingerprint Chromium.app"
	executable := filepath.Join(appBundle, "Contents", "MacOS", "Fingerprint Chromium")

	if got, shouldCopy, err := validateDarwinMountedAppBundle("/Volumes/Installer/source", executable, appBundle); err != nil || !shouldCopy || got != appBundle {
		t.Fatalf("同一挂载卷中的应用包应复制: got=%q shouldCopy=%v err=%v", got, shouldCopy, err)
	}
	if _, _, err := validateDarwinMountedAppBundle("/Volumes/Installer/source", "/private/tmp/Chromium.app/Contents/MacOS/Chromium", "/private/tmp/Chromium.app"); err == nil {
		t.Fatal("从挂载卷逃逸到本地路径的符号链接应被拒绝")
	}
	if _, _, err := validateDarwinMountedAppBundle("/Volumes/Installer/source", executable, "/Volumes/Other/Fingerprint Chromium.app"); err == nil {
		t.Fatal("可执行文件和应用包位于不同挂载卷时应被拒绝")
	}
	otherAppBundle := "/Volumes/Other/Fingerprint Chromium.app"
	otherExecutable := filepath.Join(otherAppBundle, "Contents", "MacOS", "Fingerprint Chromium")
	if _, _, err := validateDarwinMountedAppBundle("/Volumes/Installer/source", otherExecutable, otherAppBundle); err == nil {
		t.Fatal("可执行文件和应用包离开所选挂载卷时应被拒绝")
	}
}

func TestDarwinMountedAppBundleForImportRejectsExecutableSymlinkEscape(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	selectedApp := filepath.Join(root, "Selected Chromium.app")
	otherApp := filepath.Join(root, "Other Chromium.app")
	selectedExecutable := filepath.Join(selectedApp, "Contents", "MacOS", "Selected Chromium")
	otherExecutable := filepath.Join(otherApp, "Contents", "MacOS", "Other Chromium")
	if err := os.MkdirAll(filepath.Dir(selectedExecutable), 0o755); err != nil {
		t.Fatalf("创建所选应用包失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(otherExecutable), 0o755); err != nil {
		t.Fatalf("创建目标应用包失败: %v", err)
	}
	if err := os.WriteFile(otherExecutable, []byte("stub"), 0o755); err != nil {
		t.Fatalf("写入目标可执行文件失败: %v", err)
	}
	if err := os.Symlink(otherExecutable, selectedExecutable); err != nil {
		t.Fatalf("创建越界可执行文件链接失败: %v", err)
	}

	if _, _, err := darwinMountedAppBundleForImport(selectedApp, selectedExecutable); err == nil {
		t.Fatal("离开所选应用包的可执行文件链接应被拒绝")
	}
}

func TestDarwinVolumesMountNormalizesPaths(t *testing.T) {
	if got, want := darwinVolumesMount("/Volumes/Installer/../Installer/Fingerprint Chromium.app"), "/Volumes/Installer"; got != want {
		t.Fatalf("挂载点归一化错误: got=%q want=%q", got, want)
	}
	if got := darwinVolumesMount("/Volumes"); got != "" {
		t.Fatalf("/Volumes 本身不是挂载点: %q", got)
	}
}

type importCoreDAOStub struct {
	cores     []browser.Core
	upsertErr error
	listErr   error
}

func (s *importCoreDAOStub) List() ([]browser.Core, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return append([]browser.Core{}, s.cores...), nil
}

func (s *importCoreDAOStub) Upsert(core browser.Core) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	s.cores = append(s.cores, core)
	return nil
}

func (s *importCoreDAOStub) Delete(coreID string) error     { return nil }
func (s *importCoreDAOStub) SetDefault(coreID string) error { return nil }

func TestSavePublishedImportedCoreRemovesDirectoryWhenSaveFails(t *testing.T) {
	root := t.TempDir()
	publishedDir := filepath.Join(root, "chrome", "Fingerprint Chromium")
	if err := os.MkdirAll(publishedDir, 0o755); err != nil {
		t.Fatalf("创建已发布目录失败: %v", err)
	}

	cfg := config.DefaultConfig()
	app := NewApp(root)
	app.config = cfg
	app.browserMgr = browser.NewManager(cfg, root)
	app.browserMgr.CoreDAO = &importCoreDAOStub{upsertErr: errors.New("database unavailable")}

	_, err := app.savePublishedImportedCore(browser.CoreInput{CoreName: "Fingerprint Chromium", CorePath: filepath.Join("chrome", "Fingerprint Chromium")}, publishedDir)
	if err == nil {
		t.Fatal("保存失败时应返回错误")
	}
	if _, statErr := os.Stat(publishedDir); !os.IsNotExist(statErr) {
		t.Fatalf("未注册的已发布目录应被清理: %v", statErr)
	}
}

func TestSavePublishedImportedCoreKeepsDirectoryWhenReadbackIsTransient(t *testing.T) {
	root := t.TempDir()
	publishedDir := filepath.Join(root, "chrome", "Fingerprint Chromium")
	if err := os.MkdirAll(publishedDir, 0o755); err != nil {
		t.Fatalf("创建已发布目录失败: %v", err)
	}

	cfg := config.DefaultConfig()
	app := NewApp(root)
	app.config = cfg
	app.browserMgr = browser.NewManager(cfg, root)
	app.browserMgr.CoreDAO = &importCoreDAOStub{listErr: errors.New("temporary read failure")}

	_, err := app.savePublishedImportedCore(browser.CoreInput{CoreName: "Fingerprint Chromium", CorePath: filepath.Join("chrome", "Fingerprint Chromium")}, publishedDir)
	if err == nil {
		t.Fatal("回读失败时应返回错误")
	}
	if _, statErr := os.Stat(publishedDir); statErr != nil {
		t.Fatalf("持久化成功但回读暂时失败时不应删除目录: %v", statErr)
	}
}

func TestSavePublishedImportedCoreDoesNotTrustFailedConfigSave(t *testing.T) {
	root := t.TempDir()
	publishedDir := filepath.Join(root, "chrome", "Fingerprint Chromium")
	if err := os.MkdirAll(publishedDir, 0o755); err != nil {
		t.Fatalf("创建已发布目录失败: %v", err)
	}
	// Make the config target a directory so the file fallback fails after it
	// has already updated the in-memory core slice.
	if err := os.Mkdir(filepath.Join(root, "config.yaml"), 0o755); err != nil {
		t.Fatalf("创建故障配置目标失败: %v", err)
	}

	cfg := config.DefaultConfig()
	app := NewApp(root)
	app.config = cfg
	app.browserMgr = browser.NewManager(cfg, root)

	_, err := app.savePublishedImportedCore(browser.CoreInput{CoreName: "Fingerprint Chromium", CorePath: filepath.Join("chrome", "Fingerprint Chromium")}, publishedDir)
	if err == nil {
		t.Fatal("配置文件保存失败时应返回错误")
	}
	if _, statErr := os.Stat(publishedDir); !os.IsNotExist(statErr) {
		t.Fatalf("配置未持久化时应清理已发布目录: %v", statErr)
	}
	if len(app.browserMgr.Config.Browser.Cores) != 0 {
		t.Fatalf("配置保存失败后不应保留内存中的内核记录: %#v", app.browserMgr.Config.Browser.Cores)
	}
}
