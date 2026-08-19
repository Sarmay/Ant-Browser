package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSnapshotFiles(t *testing.T, dir, snapshotID, name string) (string, string) {
	t.Helper()

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

func TestFindFilesMatchesExactSnapshotID(t *testing.T) {
	dir := t.TempDir()
	wantMetaPath, wantZipPath := writeSnapshotFiles(t, dir, "snapshot-a", "current")
	writeSnapshotFiles(t, dir, "snapshot-a-extra", "other")

	metaPath, zipPath, err := FindFiles(dir, "snapshot-a")
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}
	if metaPath != wantMetaPath || zipPath != wantZipPath {
		t.Fatalf("FindFiles() = (%q, %q), want (%q, %q)", metaPath, zipPath, wantMetaPath, wantZipPath)
	}
}

func TestFindFilesRejectsPrefixOnlyMatch(t *testing.T) {
	dir := t.TempDir()
	metaPath, zipPath := writeSnapshotFiles(t, dir, "snapshot-a-extra", "other")

	gotMetaPath, gotZipPath, err := FindFiles(dir, "snapshot-a")
	if err == nil {
		t.Fatal("FindFiles() error = nil, want prefix-only match to be rejected")
	}
	if gotMetaPath != "" || gotZipPath != "" {
		t.Fatalf("FindFiles() = (%q, %q), want empty paths", gotMetaPath, gotZipPath)
	}
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("prefix meta file was changed: %v", err)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("prefix zip file was changed: %v", err)
	}
}

func TestFindFilesRejectsUnsafeSnapshotIDs(t *testing.T) {
	dir := t.TempDir()
	absID := filepath.Join(dir, "snapshot")
	for _, snapshotID := range []string{"", absID, "..", "snapshot/child", `snapshot\child`, "snapshot..id"} {
		t.Run(snapshotID, func(t *testing.T) {
			metaPath, zipPath, err := FindFiles(dir, snapshotID)
			if err == nil {
				t.Fatal("FindFiles() error = nil, want invalid snapshot ID to be rejected")
			}
			if metaPath != "" || zipPath != "" {
				t.Fatalf("FindFiles() = (%q, %q), want empty paths", metaPath, zipPath)
			}
		})
	}
}
