package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindCoreExecutableFindsBrandDarwinAppBundle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	appPath := filepath.Join(root, "Fingerprint Chromium.app")
	executablePath := filepath.Join(appPath, "Contents", "MacOS", "Fingerprint Chromium")
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatalf("创建测试应用包失败: %v", err)
	}
	if err := os.WriteFile(executablePath, []byte("stub"), 0o755); err != nil {
		t.Fatalf("写入测试可执行文件失败: %v", err)
	}

	gotPath, _, ok := FindCoreExecutable(appPath)
	if !ok {
		t.Fatal("品牌化 macOS 应用包应能识别")
	}
	if gotPath != executablePath {
		t.Fatalf("识别到的可执行文件路径错误: got=%q want=%q", gotPath, executablePath)
	}
}

func TestStripSingleExtractedRootKeepsAppBundle(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "Fingerprint Chromium.app")
	if err := os.MkdirAll(filepath.Join(appPath, "Contents", "MacOS"), 0o755); err != nil {
		t.Fatalf("创建应用包目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appPath, "Contents", "MacOS", "Fingerprint Chromium"), []byte("stub"), 0o755); err != nil {
		t.Fatalf("写入应用包文件失败: %v", err)
	}

	if err := stripSingleExtractedRoot(root, map[string]struct{}{"Fingerprint Chromium.app": {}}); err != nil {
		t.Fatalf("处理应用包顶层目录失败: %v", err)
	}
	if _, err := os.Stat(appPath); err != nil {
		t.Fatalf("应用包顶层目录不应被剥离: %v", err)
	}
}

func TestFindCoreExecutableRejectsUnrelatedDarwinAppBundle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	appPath := filepath.Join(root, "Ant Browser.app")
	executablePath := filepath.Join(appPath, "Contents", "MacOS", "ant-chrome")
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatalf("创建测试应用包失败: %v", err)
	}
	if err := os.WriteFile(executablePath, []byte("stub"), 0o755); err != nil {
		t.Fatalf("写入测试可执行文件失败: %v", err)
	}

	if gotPath, _, ok := FindCoreExecutable(appPath); ok {
		t.Fatalf("不相关应用包不应被识别为内核: %q", gotPath)
	}
}

func TestFindCoreExecutableFindsNonExecutableDarwinBundleEntry(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	appPath := filepath.Join(root, "Fingerprint Chromium.app")
	executablePath := filepath.Join(appPath, "Contents", "MacOS", "Fingerprint Chromium")
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatalf("创建测试应用包失败: %v", err)
	}
	if err := os.WriteFile(executablePath, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	if gotPath, _, ok := FindCoreExecutable(appPath); !ok || gotPath != executablePath {
		t.Fatalf("导入识别应允许由运行时修复权限的应用包文件: got=%q ok=%v", gotPath, ok)
	}
}

func TestFindCoreExecutableFindsNonExecutableKnownDarwinBundleEntry(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	appPath := filepath.Join(root, "Google Chrome.app")
	executablePath := filepath.Join(appPath, "Contents", "MacOS", "Google Chrome")
	if err := os.MkdirAll(filepath.Dir(executablePath), 0o755); err != nil {
		t.Fatalf("创建测试应用包失败: %v", err)
	}
	if err := os.WriteFile(executablePath, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	if gotPath, _, ok := FindCoreExecutable(appPath); !ok || gotPath != executablePath {
		t.Fatalf("已知应用包文件应先被识别，再由运行时修复权限: got=%q ok=%v", gotPath, ok)
	}
}

func TestFindCoreExecutableAcceptsExecutableDarwinBundleSymlink(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin app bundle behavior")
	}

	root := t.TempDir()
	appPath := filepath.Join(root, "Fingerprint Chromium.app")
	macOSDir := filepath.Join(appPath, "Contents", "MacOS")
	targetPath := filepath.Join(macOSDir, "Fingerprint Chromium.bin")
	linkPath := filepath.Join(macOSDir, "Fingerprint Chromium")
	if err := os.MkdirAll(macOSDir, 0o755); err != nil {
		t.Fatalf("创建测试应用包失败: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("stub"), 0o755); err != nil {
		t.Fatalf("写入测试可执行文件失败: %v", err)
	}
	if err := os.Symlink(filepath.Base(targetPath), linkPath); err != nil {
		t.Fatalf("创建可执行文件符号链接失败: %v", err)
	}

	gotPath, _, ok := FindCoreExecutable(appPath)
	if !ok || gotPath != linkPath {
		t.Fatalf("应用包内可执行文件符号链接应被识别: got=%q ok=%v", gotPath, ok)
	}
}
