package browser

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

// CoreExecutableCandidates 返回当前平台可接受的浏览器可执行文件候选名。
func CoreExecutableCandidates() []string {
	switch goruntime.GOOS {
	case "windows":
		return []string{"chrome.exe"}
	case "linux":
		return []string{"chrome", "chrome-bin", "chromium", "chromium-browser", "ungoogled-chromium", "chrome.exe"}
	case "darwin":
		return []string{
			"Google Chrome.app/Contents/MacOS/Google Chrome",
			"Chromium.app/Contents/MacOS/Chromium",
			"chrome",
		}
	default:
		return []string{"chrome"}
	}
}

func CoreExecutablePlatform() string {
	return goruntime.GOOS + "/" + goruntime.GOARCH
}

// FindCoreExecutable 在指定目录查找可执行文件，返回绝对路径和命中的候选名。
func FindCoreExecutable(baseDir string) (string, string, bool) {
	if directPath, directCandidate, ok := FindCoreExecutableShallow(baseDir); ok {
		return directPath, directCandidate, true
	}
	if recursivePath, recursiveCandidate, ok := findNestedCoreExecutable(baseDir); ok {
		return recursivePath, recursiveCandidate, true
	}
	return "", "", false
}

func FindCoreExecutableShallow(baseDir string) (string, string, bool) {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		return "", "", false
	}
	if directPath, directCandidate, ok := findDirectCoreExecutable(baseDir); ok {
		return directPath, directCandidate, true
	}
	if bundlePath, bundleCandidate, ok := findAppBundleExecutable(baseDir); ok {
		return bundlePath, bundleCandidate, true
	}
	for _, candidate := range CoreExecutableCandidates() {
		p := filepath.Join(baseDir, filepath.FromSlash(candidate))
		if isRegularExecutable(p) {
			return p, candidate, true
		}
	}
	return "", "", false
}

func findNestedCoreExecutable(baseDir string) (string, string, bool) {
	info, err := os.Stat(baseDir)
	if err != nil || !info.IsDir() {
		return "", "", false
	}
	if goruntime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(filepath.Base(filepath.Clean(baseDir))), ".app") {
		// The shallow bundle check is authoritative. Do not walk into a rejected
		// bundle and accidentally accept a non-executable known-name file.
		return "", "", false
	}
	baseDepth := strings.Count(filepath.ToSlash(filepath.Clean(baseDir)), "/")
	candidateNames := make(map[string]string)
	for _, candidate := range CoreExecutableCandidates() {
		candidateNames[strings.ToLower(filepath.Base(candidate))] = candidate
	}

	var matchedPath string
	var matchedCandidate string
	_ = filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || path == baseDir || matchedPath != "" {
			return nil
		}
		if entry.IsDir() {
			if goruntime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(entry.Name()), ".app") {
				if appPath, appCandidate, ok := findAppBundleExecutable(path); ok {
					matchedPath = appPath
					matchedCandidate = appCandidate
				}
				return filepath.SkipDir
			}
			depth := strings.Count(filepath.ToSlash(filepath.Clean(path)), "/") - baseDepth
			if depth > 5 {
				return filepath.SkipDir
			}
			return nil
		}
		candidate, ok := candidateNames[strings.ToLower(entry.Name())]
		if !ok {
			return nil
		}
		if !isRegularExecutable(path) {
			return nil
		}
		matchedPath = path
		matchedCandidate = candidate
		return nil
	})
	if matchedPath == "" {
		return "", "", false
	}
	return matchedPath, matchedCandidate, true
}

func findDirectCoreExecutable(path string) (string, string, bool) {
	if !isRegularExecutable(path) {
		return "", "", false
	}

	normalized := filepath.ToSlash(filepath.Clean(path))
	for _, candidate := range CoreExecutableCandidates() {
		candidatePath := filepath.ToSlash(candidate)
		if strings.HasSuffix(normalized, candidatePath) || filepath.Base(normalized) == filepath.Base(candidatePath) {
			return path, candidate, true
		}
	}

	return "", "", false
}

func findGenericDarwinAppExecutable(appPath string) (string, string, bool) {
	macOSDir := filepath.Join(appPath, "Contents", "MacOS")
	entries, err := os.ReadDir(macOSDir)
	if err != nil {
		return "", "", false
	}

	bundleName := strings.TrimSuffix(filepath.Base(appPath), ".app")
	if !isLikelyChromiumBundle(bundleName) {
		return "", "", false
	}
	for _, entry := range entries {
		executablePath := appBundleExecutablePath(appPath, entry.Name())
		if !isRegularExecutable(executablePath) {
			continue
		}
		if strings.EqualFold(entry.Name(), bundleName) && isLikelyChromiumBundle(bundleName) {
			return executablePath, appBundleExecutableCandidate(appPath, entry.Name()), true
		}
		if isLikelyChromiumExecutable(entry.Name()) {
			return executablePath, appBundleExecutableCandidate(appPath, entry.Name()), true
		}
	}
	return "", "", false
}

// isRegularExecutable is a discovery predicate, not a launch predicate.
// Runtime launch calls fsutil.EnsureExecutable, which repairs missing bits on
// Unix after an archive extractor has stripped them. Windows has no execute
// bit in the Go file mode model, so regularity is sufficient there as well.
func isRegularExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	return true
}

func appBundleExecutablePath(appPath, executableName string) string {
	return filepath.Join(appPath, "Contents", "MacOS", executableName)
}

func appBundleExecutableCandidate(appPath, executableName string) string {
	return filepath.ToSlash(filepath.Join(filepath.Base(appPath), "Contents", "MacOS", executableName))
}

func isLikelyChromiumExecutable(name string) bool {
	lower := strings.ToLower(name)
	for _, excluded := range []string{"helper", "crashpad", "renderer", "utility", "sandbox", "nacl"} {
		if strings.Contains(lower, excluded) {
			return false
		}
	}
	return strings.Contains(lower, "chrom") || strings.Contains(lower, "fingerprint")
}

func isLikelyChromiumBundle(name string) bool {
	lower := strings.ToLower(strings.TrimSuffix(name, ".app"))
	return strings.Contains(lower, "chrom") || strings.Contains(lower, "fingerprint")
}

func findAppBundleExecutable(path string) (string, string, bool) {
	if goruntime.GOOS != "darwin" {
		return "", "", false
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", "", false
	}

	normalized := filepath.ToSlash(filepath.Clean(path))
	if !strings.HasSuffix(strings.ToLower(normalized), ".app") {
		return "", "", false
	}

	for _, candidate := range CoreExecutableCandidates() {
		candidatePath := filepath.ToSlash(candidate)
		appMarker := ".app/"
		index := strings.Index(strings.ToLower(candidatePath), appMarker)
		if index < 0 {
			continue
		}
		if !strings.EqualFold(filepath.Base(normalized), filepath.Base(candidatePath[:index+len(".app")])) {
			continue
		}

		relativeExecutable := candidatePath[index+len(appMarker):]
		if relativeExecutable == "" {
			continue
		}

		p := filepath.Join(path, filepath.FromSlash(relativeExecutable))
		if isRegularExecutable(p) {
			return p, candidate, true
		}
	}

	return findGenericDarwinAppExecutable(path)
}
