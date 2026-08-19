package browser

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

type archiveEntryMeta struct {
	Name string
}

type archiveProgress struct {
	index int
	total int
}

type pendingArchiveHardlink struct {
	name       string
	targetPath string
	linkPath   string
}

func SupportedCoreArchivePattern() string {
	return "*.zip;*.tar;*.tar.gz;*.tgz;*.tar.xz;*.txz;*.tar.bz2;*.tbz2"
}

func SupportedCoreArchiveDescription() string {
	return "支持 ZIP、TAR、TAR.GZ、TAR.XZ、TAR.BZ2"
}

func coreArchiveTempPattern(rawURL string) string {
	lowerName := strings.ToLower(strings.TrimSpace(rawURL))
	if parsed, err := filepathFromURLPath(lowerName); err == nil && parsed != "" {
		lowerName = parsed
	}
	for _, suffix := range coreArchiveSuffixes() {
		if strings.HasSuffix(lowerName, suffix) {
			return "download_*" + suffix
		}
	}
	return "download_*"
}

func filepathFromURLPath(raw string) (string, error) {
	parts := strings.SplitN(raw, "?", 2)
	parts = strings.SplitN(parts[0], "#", 2)
	return filepath.Base(parts[0]), nil
}

func extractCoreArchiveAndStripRoot(archivePath, dest string, progressCb func(int, string)) error {
	lower := strings.ToLower(archivePath)
	if strings.HasSuffix(lower, ".zip") {
		return extractZipArchiveAndStripRoot(archivePath, dest, progressCb)
	}
	if isTarArchivePath(lower) {
		return extractTarArchiveAndStripRoot(archivePath, dest, progressCb)
	}
	if err := extractZipArchiveAndStripRoot(archivePath, dest, progressCb); err == nil {
		return nil
	}
	return extractTarArchiveAndStripRoot(archivePath, dest, progressCb)
}

func ExtractCoreArchiveAndStripRootForImport(archivePath, dest string, progressCb func(int, string)) error {
	return extractCoreArchiveAndStripRoot(archivePath, dest, progressCb)
}

func extractZipArchiveAndStripRoot(archivePath, dest string, progressCb func(int, string)) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("空的压缩包")
	}
	metas := make([]archiveEntryMeta, 0, len(reader.File))
	for _, file := range reader.File {
		metas = append(metas, archiveEntryMeta{Name: file.Name})
	}
	rootPrefix, hasCommonRoot := detectCommonArchiveRoot(metas)
	if err := prepareArchiveDestination(dest); err != nil {
		return err
	}

	progress := archiveProgress{total: len(reader.File)}
	for _, file := range reader.File {
		progress.report(progressCb)
		cleanName := strippedArchiveName(file.Name, rootPrefix, hasCommonRoot)
		if cleanName == "" {
			continue
		}
		targetPath, err := safeArchiveTargetPath(dest, cleanName)
		if err != nil {
			return err
		}
		if file.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := readArchiveSymlinkTarget(file)
			if err != nil {
				return fmt.Errorf("读取压缩包符号链接失败 %s: %w", file.Name, err)
			}
			if err := createArchiveSymlink(dest, targetPath, linkTarget); err != nil {
				return fmt.Errorf("创建符号链接失败 %s: %w", cleanName, err)
			}
			continue
		}
		if file.FileInfo().IsDir() {
			if err := ensureArchiveDirectories(dest, targetPath, file.Mode().Perm()); err != nil {
				return err
			}
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("读取压缩包文件失败 %s: %w", file.Name, err)
		}
		if err := writeReaderToFile(dest, targetPath, rc, file.Mode().Perm()); err != nil {
			_ = rc.Close()
			return err
		}
		if err := rc.Close(); err != nil {
			return err
		}
	}
	progressCb(100, "解压完成！")
	return nil
}

func extractTarArchiveAndStripRoot(archivePath, dest string, progressCb func(int, string)) error {
	if err := prepareArchiveDestination(dest); err != nil {
		return err
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	stream, closeStream, err := tarStreamReader(archivePath, file)
	if err != nil {
		return err
	}
	defer closeStream()

	reader := tar.NewReader(stream)
	entryCount := 0
	topLevels := make(map[string]struct{})
	pendingHardlinks := make([]pendingArchiveHardlink, 0)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		entryCount++
		if entryCount == 1 || entryCount%50 == 0 {
			progressCb(0, fmt.Sprintf("正在解压文件 %d...", entryCount))
		}
		cleanName := strippedArchiveName(header.Name, "", false)
		if cleanName == "" {
			continue
		}
		if top := topLevelArchiveName(cleanName); top != "" {
			topLevels[top] = struct{}{}
		}
		targetPath, err := safeArchiveTargetPath(dest, cleanName)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := ensureArchiveDirectories(dest, targetPath, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := createArchiveSymlink(dest, targetPath, header.Linkname); err != nil {
				return fmt.Errorf("创建符号链接失败 %s: %w", cleanName, err)
			}
		case tar.TypeLink:
			linkName := strippedArchiveName(header.Linkname, "", false)
			linkPath, err := safeArchiveTargetPath(dest, linkName)
			if err != nil {
				return fmt.Errorf("非法硬链接目标 %s: %w", header.Linkname, err)
			}
			if err := ensureArchivePathHasNoSymlink(dest, linkPath); err != nil {
				return err
			}
			if err := ensureArchiveDirectories(dest, filepath.Dir(targetPath), 0o755); err != nil {
				return err
			}
			if err := ensureArchivePathHasNoSymlink(dest, targetPath); err != nil {
				return err
			}
			pendingHardlinks = append(pendingHardlinks, pendingArchiveHardlink{name: cleanName, targetPath: linkPath, linkPath: targetPath})
		case tar.TypeReg, tar.TypeRegA:
			if err := writeReaderToFile(dest, targetPath, reader, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		}
	}
	if entryCount == 0 {
		return fmt.Errorf("空的压缩包")
	}
	for _, hardlink := range pendingHardlinks {
		if err := ensureArchivePathHasNoSymlink(dest, hardlink.targetPath); err != nil {
			return err
		}
		if err := os.Link(hardlink.targetPath, hardlink.linkPath); err != nil {
			return fmt.Errorf("创建硬链接失败 %s: %w", hardlink.name, err)
		}
	}
	if err := stripSingleExtractedRoot(dest, topLevels); err != nil {
		return err
	}
	progressCb(100, "解压完成！")
	return nil
}

func tarStreamReader(archivePath string, file *os.File) (io.Reader, func(), error) {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		reader, err := gzip.NewReader(file)
		if err != nil {
			return nil, func() {}, err
		}
		return reader, func() { _ = reader.Close() }, nil
	case strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz"):
		reader, err := xz.NewReader(file)
		return reader, func() {}, err
	case strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2"):
		return bzip2.NewReader(file), func() {}, nil
	case strings.HasSuffix(lower, ".tar"):
		return file, func() {}, nil
	default:
		return file, func() {}, nil
	}
}

func isTarArchivePath(path string) bool {
	for _, suffix := range coreArchiveSuffixes() {
		if suffix == ".zip" {
			continue
		}
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func coreArchiveSuffixes() []string {
	return []string{".tar.gz", ".tar.xz", ".tar.bz2", ".tgz", ".txz", ".tbz2", ".zip", ".tar"}
}

func detectCommonArchiveRoot(entries []archiveEntryMeta) (string, bool) {
	var rootPrefix string
	for _, entry := range entries {
		cleanName := normalizeArchiveEntryName(entry.Name)
		parts := strings.SplitN(cleanName, "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		if rootPrefix == "" {
			rootPrefix = parts[0] + "/"
			continue
		}
		if !strings.HasPrefix(cleanName, rootPrefix) && cleanName != strings.TrimSuffix(rootPrefix, "/") {
			return "", false
		}
	}
	if isAppBundleRoot(rootPrefix) {
		return "", false
	}
	return rootPrefix, rootPrefix != ""
}

func strippedArchiveName(name string, rootPrefix string, hasCommonRoot bool) string {
	cleanName := normalizeArchiveEntryName(name)
	if hasCommonRoot {
		if cleanName == rootPrefix || cleanName == strings.TrimSuffix(rootPrefix, "/") {
			return ""
		}
		cleanName = strings.TrimPrefix(cleanName, rootPrefix)
	}
	if cleanName == "" || cleanName == "." || cleanName == "/" {
		return ""
	}
	return cleanName
}

func topLevelArchiveName(name string) string {
	cleanName := normalizeArchiveEntryName(name)
	if cleanName == "" || cleanName == "." {
		return ""
	}
	parts := strings.SplitN(cleanName, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func stripSingleExtractedRoot(dest string, topLevels map[string]struct{}) error {
	if len(topLevels) != 1 {
		return nil
	}
	var rootName string
	for name := range topLevels {
		rootName = name
	}
	if isAppBundleRoot(rootName) {
		return nil
	}
	rootPath, err := safeArchiveTargetPath(dest, rootName)
	if err != nil {
		return err
	}
	info, err := os.Lstat(rootPath)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		source := filepath.Join(rootPath, entry.Name())
		target := filepath.Join(dest, entry.Name())
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("剥离顶层目录失败，目标已存在: %s", target)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(source, target); err != nil {
			return err
		}
	}
	return os.Remove(rootPath)
}

func isAppBundleRoot(name string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSuffix(strings.TrimSpace(name), "/")), ".app")
}

func normalizeArchiveEntryName(name string) string {
	cleanName := strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	return path.Clean(cleanName)
}

func safeArchiveTargetPath(dest, cleanName string) (string, error) {
	entryPath := filepath.FromSlash(cleanName)
	if cleanName == "." || strings.HasPrefix(cleanName, "../") || cleanName == ".." || filepath.IsAbs(entryPath) || filepath.VolumeName(entryPath) != "" {
		return "", fmt.Errorf("非法文件路径: %s", cleanName)
	}
	targetPath := filepath.Join(dest, entryPath)
	if err := ensureArchivePathWithinDestination(dest, targetPath); err != nil {
		return "", fmt.Errorf("非法文件路径: %s", cleanName)
	}
	return targetPath, nil
}

func prepareArchiveDestination(dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	info, err := os.Lstat(dest)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("解压目标目录无效: %s", dest)
	}
	return nil
}

func ensureArchivePathWithinDestination(dest, targetPath string) error {
	destClean := canonicalArchivePath(filepath.Clean(dest))
	targetClean := canonicalArchivePath(filepath.Clean(targetPath))
	relativePath, err := filepath.Rel(destClean, targetClean)
	if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relativePath) {
		return fmt.Errorf("路径越过解压目标目录: %s", targetPath)
	}
	return nil
}

func canonicalArchivePath(path string) string {
	path = filepath.Clean(path)
	currentPath := path
	missingParts := make([]string, 0)
	for {
		if resolvedPath, err := filepath.EvalSymlinks(currentPath); err == nil {
			for index := len(missingParts) - 1; index >= 0; index-- {
				resolvedPath = filepath.Join(resolvedPath, missingParts[index])
			}
			return filepath.Clean(resolvedPath)
		} else if !os.IsNotExist(err) {
			return path
		}
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return path
		}
		missingParts = append(missingParts, filepath.Base(currentPath))
		currentPath = parentPath
	}
}

func ensureArchiveDirectories(dest, directoryPath string, mode os.FileMode) error {
	if err := ensureArchivePathWithinDestination(dest, directoryPath); err != nil {
		return err
	}
	relativePath, err := filepath.Rel(filepath.Clean(dest), filepath.Clean(directoryPath))
	if err != nil {
		return err
	}
	if relativePath == "." {
		return nil
	}

	currentPath := filepath.Clean(dest)
	parts := strings.Split(relativePath, string(os.PathSeparator))
	for index, part := range parts {
		if part == "" || part == "." {
			continue
		}
		currentPath = filepath.Join(currentPath, part)
		info, err := os.Lstat(currentPath)
		switch {
		case err == nil:
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("解压路径包含非目录或符号链接: %s", currentPath)
			}
		case os.IsNotExist(err):
			directoryMode := os.FileMode(0o755)
			if index == len(parts)-1 {
				directoryMode = mode
			}
			if err := os.Mkdir(currentPath, directoryMode); err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

func ensureArchivePathHasNoSymlink(dest, targetPath string) error {
	if err := ensureArchivePathWithinDestination(dest, targetPath); err != nil {
		return err
	}
	relativePath, err := filepath.Rel(filepath.Clean(dest), filepath.Clean(targetPath))
	if err != nil {
		return err
	}
	if relativePath == "." {
		return nil
	}

	currentPath := filepath.Clean(dest)
	for _, part := range strings.Split(relativePath, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		currentPath = filepath.Join(currentPath, part)
		info, err := os.Lstat(currentPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("解压路径包含符号链接: %s", currentPath)
		}
	}
	return nil
}

func createArchiveSymlink(dest, targetPath, linkTarget string) error {
	if err := validateArchiveSymlinkTarget(dest, targetPath, linkTarget); err != nil {
		return err
	}
	if err := ensureArchiveDirectories(dest, filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Symlink(linkTarget, targetPath); err != nil {
		return err
	}
	return nil
}

func validateArchiveSymlinkTarget(dest, targetPath, linkTarget string) error {
	if linkTarget == "" || strings.ContainsRune(linkTarget, '\x00') {
		return fmt.Errorf("非法符号链接目标: %q", linkTarget)
	}
	normalizedTarget := strings.ReplaceAll(linkTarget, "\\", "/")
	linkPath := filepath.FromSlash(normalizedTarget)
	if path.IsAbs(normalizedTarget) || filepath.IsAbs(linkPath) || filepath.VolumeName(linkPath) != "" {
		return fmt.Errorf("非法符号链接目标: %s", linkTarget)
	}
	resolvedTarget := filepath.Join(filepath.Dir(targetPath), linkPath)
	if err := ensureArchivePathWithinDestination(dest, resolvedTarget); err != nil {
		return fmt.Errorf("非法符号链接目标: %s: %w", linkTarget, err)
	}
	if err := ensureArchiveResolvedPathWithinDestination(dest, resolvedTarget); err != nil {
		return fmt.Errorf("非法符号链接目标: %s: %w", linkTarget, err)
	}
	return nil
}

func ensureArchiveResolvedPathWithinDestination(dest, targetPath string) error {
	if err := ensureArchivePathWithinDestination(dest, targetPath); err != nil {
		return err
	}
	resolvedPath, err := filepath.EvalSymlinks(targetPath)
	if err == nil {
		if pathErr := ensureArchivePathWithinDestination(dest, resolvedPath); pathErr != nil {
			return pathErr
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	// The target may be a forward reference that does not exist yet. Walk to
	// the nearest existing ancestor and resolve any links there before allowing
	// the new link to be created.
	currentPath := filepath.Clean(targetPath)
	for {
		if _, statErr := os.Lstat(currentPath); statErr == nil {
			resolvedAncestor, resolveErr := filepath.EvalSymlinks(currentPath)
			if resolveErr != nil {
				return resolveErr
			}
			return ensureArchivePathWithinDestination(dest, resolvedAncestor)
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return ensureArchivePathWithinDestination(dest, currentPath)
		}
		currentPath = parentPath
	}
}

func readArchiveSymlinkTarget(file *zip.File) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()
	const maxSymlinkTargetSize = 4 << 10
	contents, err := io.ReadAll(io.LimitReader(reader, maxSymlinkTargetSize+1))
	if err != nil {
		return "", err
	}
	if len(contents) > maxSymlinkTargetSize {
		return "", fmt.Errorf("符号链接目标过长")
	}
	return string(contents), nil
}

func writeReaderToFile(dest, targetPath string, reader io.Reader, mode os.FileMode) error {
	if err := ensureArchiveDirectories(dest, filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	if err := ensureArchivePathHasNoSymlink(dest, targetPath); err != nil {
		return err
	}
	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("打开解压文件写入失败 %s: %w", targetPath, err)
	}
	_, copyErr := io.Copy(outFile, reader)
	closeErr := outFile.Close()
	if copyErr != nil {
		return fmt.Errorf("写入文件流失败 %s: %w", targetPath, copyErr)
	}
	return closeErr
}

func (p *archiveProgress) report(progressCb func(int, string)) {
	p.index++
	if p.total <= 0 {
		progressCb(0, "正在解压...")
		return
	}
	percent := int((float64(p.index-1) / float64(p.total)) * 100)
	if p.index == 1 || p.index%50 == 0 {
		progressCb(percent, fmt.Sprintf("正在解压文件 %d / %d...", p.index, p.total))
	}
}
