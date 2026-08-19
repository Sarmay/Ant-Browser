package browser

import (
	"archive/tar"
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type tarArchiveTestEntry struct {
	name     string
	typeflag byte
	linkName string
	contents string
}

func TestExtractZipArchiveKeepsAppBundleSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows 创建符号链接需要额外权限")
	}

	archivePath := filepath.Join(t.TempDir(), "core.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("创建 ZIP 压缩包失败: %v", err)
	}
	writer := zip.NewWriter(archive)
	entry, err := writer.Create("Fingerprint Browser.app/Contents/Versions/A/marker")
	if err != nil {
		t.Fatalf("创建 ZIP 文件条目失败: %v", err)
	}
	if _, err := entry.Write([]byte("ok")); err != nil {
		t.Fatalf("写入 ZIP 文件条目失败: %v", err)
	}
	linkHeader := &zip.FileHeader{Name: "Fingerprint Browser.app/Contents/Versions/Current", Method: zip.Store}
	linkHeader.SetMode(os.ModeSymlink | 0o777)
	linkEntry, err := writer.CreateHeader(linkHeader)
	if err != nil {
		t.Fatalf("创建 ZIP 符号链接条目失败: %v", err)
	}
	if _, err := linkEntry.Write([]byte("A")); err != nil {
		t.Fatalf("写入 ZIP 符号链接条目失败: %v", err)
	}
	aliasHeader := &zip.FileHeader{Name: "Fingerprint Browser.app/Contents/Versions/Alias", Method: zip.Store}
	aliasHeader.SetMode(os.ModeSymlink | 0o777)
	aliasEntry, err := writer.CreateHeader(aliasHeader)
	if err != nil {
		t.Fatalf("创建 ZIP 符号链接链条目失败: %v", err)
	}
	if _, err := aliasEntry.Write([]byte("Current")); err != nil {
		t.Fatalf("写入 ZIP 符号链接链条目失败: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 ZIP 写入器失败: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("关闭 ZIP 文件失败: %v", err)
	}

	dest := filepath.Join(t.TempDir(), "dest")
	if err := extractZipArchiveAndStripRoot(archivePath, dest, func(int, string) {}); err != nil {
		t.Fatalf("解压 ZIP 应用包失败: %v", err)
	}

	appPath := filepath.Join(dest, "Fingerprint Browser.app")
	linkPath := filepath.Join(appPath, "Contents", "Versions", "Current")
	if target, err := os.Readlink(linkPath); err != nil || target != "A" {
		t.Fatalf("应用包内符号链接未保留: target=%q err=%v", target, err)
	}
	contents, err := os.ReadFile(filepath.Join(linkPath, "marker"))
	if err != nil || string(contents) != "ok" {
		t.Fatalf("应用包内符号链接不可用: contents=%q err=%v", contents, err)
	}
	aliasContents, err := os.ReadFile(filepath.Join(appPath, "Contents", "Versions", "Alias", "marker"))
	if err != nil || string(aliasContents) != "ok" {
		t.Fatalf("应用包内符号链接链不可用: contents=%q err=%v", aliasContents, err)
	}
	if _, err := os.Stat(filepath.Join(dest, "Contents")); !os.IsNotExist(err) {
		t.Fatalf("顶层 .app 不应被剥离: %v", err)
	}
}

func TestExtractZipArchiveRejectsUnsafeSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows 创建符号链接需要额外权限")
	}

	archivePath := filepath.Join(t.TempDir(), "unsafe.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("创建 ZIP 压缩包失败: %v", err)
	}
	writer := zip.NewWriter(archive)
	linkHeader := &zip.FileHeader{Name: "link", Method: zip.Store}
	linkHeader.SetMode(os.ModeSymlink | 0o777)
	linkEntry, err := writer.CreateHeader(linkHeader)
	if err != nil {
		t.Fatalf("创建 ZIP 符号链接条目失败: %v", err)
	}
	if _, err := linkEntry.Write([]byte("../outside")); err != nil {
		t.Fatalf("写入 ZIP 符号链接条目失败: %v", err)
	}
	otherEntry, err := writer.Create("other")
	if err != nil {
		t.Fatalf("创建 ZIP 普通文件条目失败: %v", err)
	}
	if _, err := otherEntry.Write([]byte("ok")); err != nil {
		t.Fatalf("写入 ZIP 普通文件条目失败: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 ZIP 写入器失败: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("关闭 ZIP 文件失败: %v", err)
	}

	if err := extractZipArchiveAndStripRoot(archivePath, filepath.Join(t.TempDir(), "dest"), func(int, string) {}); err == nil {
		t.Fatal("包含越界符号链接的 ZIP 应被拒绝")
	}
}

func TestExtractTarArchiveRejectsUnsafeEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []tarArchiveTestEntry
	}{
		{
			name: "path traversal",
			entries: []tarArchiveTestEntry{{
				name:     "../outside",
				typeflag: tar.TypeReg,
				contents: "unsafe",
			}},
		},
		{
			name: "absolute path",
			entries: []tarArchiveTestEntry{{
				name:     "/outside",
				typeflag: tar.TypeReg,
				contents: "unsafe",
			}},
		},
		{
			name: "symlink outside destination",
			entries: []tarArchiveTestEntry{{
				name:     "link",
				typeflag: tar.TypeSymlink,
				linkName: "../outside",
			}},
		},
		{
			name: "hardlink outside destination",
			entries: []tarArchiveTestEntry{{
				name:     "link",
				typeflag: tar.TypeLink,
				linkName: "../outside",
			}},
		},
		{
			name: "write through created symlink",
			entries: []tarArchiveTestEntry{
				{name: "root/target", typeflag: tar.TypeDir},
				{name: "root/link", typeflag: tar.TypeSymlink, linkName: "target"},
				{name: "root/link/escaped", typeflag: tar.TypeReg, contents: "unsafe"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			archivePath := writeTarArchive(t, test.entries)
			dest := filepath.Join(t.TempDir(), "dest")
			if err := extractTarArchiveAndStripRoot(archivePath, dest, func(int, string) {}); err == nil {
				t.Fatal("包含不安全条目的 TAR 应被拒绝")
			}
			if _, err := os.Stat(filepath.Join(dest, "root", "target", "escaped")); !os.IsNotExist(err) {
				t.Fatalf("不应通过符号链接写入文件: %v", err)
			}
		})
	}
}

func TestExtractTarArchiveCreatesSafeHardlink(t *testing.T) {
	archivePath := writeTarArchive(t, []tarArchiveTestEntry{
		{name: "core/copy", typeflag: tar.TypeLink, linkName: "core/original"},
		{name: "core/original", typeflag: tar.TypeReg, contents: "ok"},
	})
	dest := filepath.Join(t.TempDir(), "dest")

	if err := extractTarArchiveAndStripRoot(archivePath, dest, func(int, string) {}); err != nil {
		t.Fatalf("解压包含安全硬链接的 TAR 失败: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(dest, "copy"))
	if err != nil || string(contents) != "ok" {
		t.Fatalf("安全硬链接未保留: contents=%q err=%v", contents, err)
	}
}

func writeTarArchive(t *testing.T, entries []tarArchiveTestEntry) string {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "archive.tar")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("创建 TAR 压缩包失败: %v", err)
	}
	writer := tar.NewWriter(archive)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Linkname: entry.linkName,
			Mode:     0o755,
			Size:     int64(len(entry.contents)),
		}
		if entry.typeflag != tar.TypeReg && entry.typeflag != tar.TypeRegA {
			header.Size = 0
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("写入 TAR 头失败: %v", err)
		}
		if entry.contents != "" {
			if _, err := writer.Write([]byte(entry.contents)); err != nil {
				t.Fatalf("写入 TAR 内容失败: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 TAR 写入器失败: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("关闭 TAR 文件失败: %v", err)
	}
	return archivePath
}
