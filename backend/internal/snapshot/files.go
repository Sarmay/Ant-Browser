package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDir(appDataRoot, profileID string) (string, error) {
	dir := filepath.Join(appDataRoot, "snapshots", profileID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// ValidateID rejects values that cannot safely identify a snapshot file pair.
func ValidateID(snapshotID string) error {
	if strings.TrimSpace(snapshotID) == "" ||
		filepath.IsAbs(snapshotID) ||
		strings.Contains(snapshotID, "..") ||
		strings.ContainsAny(snapshotID, `/\\`) {
		return fmt.Errorf("非法快照 ID: %q", snapshotID)
	}
	return nil
}

func FindFiles(snapDir, snapshotID string) (metaPath, zipPath string, err error) {
	if err := ValidateID(snapshotID); err != nil {
		return "", "", err
	}

	entries, err := os.ReadDir(snapDir)
	if err != nil {
		return "", "", err
	}
	prefix := snapshotID + "_"
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".meta.json") {
			metaPath = filepath.Join(snapDir, entry.Name())
			zipPath = strings.TrimSuffix(metaPath, ".meta.json") + ".zip"
			if _, err := os.Stat(zipPath); err != nil {
				return "", "", fmt.Errorf("快照文件不存在: %s", zipPath)
			}
			return metaPath, zipPath, nil
		}
	}
	return "", "", fmt.Errorf("快照不存在: %s", snapshotID)
}
