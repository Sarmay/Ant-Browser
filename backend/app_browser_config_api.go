package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"ant-chrome/backend/internal/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetBrowserSettings() BrowserSettings {
	return BrowserSettings{
		UserDataRoot:           a.config.Browser.UserDataRoot,
		DefaultFingerprintArgs: append([]string{}, a.config.Browser.DefaultFingerprintArgs...),
		DefaultLaunchArgs:      append([]string{}, a.config.Browser.DefaultLaunchArgs...),
		DefaultStartURLs:       append([]string{}, a.config.Browser.DefaultStartURLs...),
		LightStartEnabled:      browserLightStartEnabled(a.config),
		RestoreLastSession:     a.config.Browser.RestoreLastSession,
		StartReadyTimeoutMs:    browserStartReadyTimeoutMillis(a.config),
		StartStableWindowMs:    browserStartStableWindowMillis(a.config),
		DefaultConnectorType:   config.NormalizeBrowserConnectorType(a.config.Browser.DefaultConnectorType),
	}
}

func (a *App) SaveBrowserSettings(settings BrowserSettings) error {
	log := logger.New("Browser")
	a.config.Browser.UserDataRoot = strings.TrimSpace(settings.UserDataRoot)
	a.config.Browser.DefaultFingerprintArgs = append([]string{}, settings.DefaultFingerprintArgs...)
	a.config.Browser.DefaultLaunchArgs = append([]string{}, settings.DefaultLaunchArgs...)
	if settings.DefaultStartURLs != nil {
		a.config.Browser.DefaultStartURLs = normalizeNonEmptyStrings(settings.DefaultStartURLs)
	} else if a.config.Browser.DefaultStartURLs == nil {
		a.config.Browser.DefaultStartURLs = config.DefaultBrowserStartURLs()
	}
	lightStartEnabled := settings.LightStartEnabled
	a.config.Browser.LightStartEnabled = &lightStartEnabled
	a.config.Browser.RestoreLastSession = settings.RestoreLastSession
	a.config.Browser.DefaultConnectorType = config.NormalizeBrowserConnectorType(settings.DefaultConnectorType)
	if settings.StartReadyTimeoutMs > 0 {
		a.config.Browser.StartReadyTimeoutMs = settings.StartReadyTimeoutMs
	} else if a.config.Browser.StartReadyTimeoutMs <= 0 {
		a.config.Browser.StartReadyTimeoutMs = browserStartReadyTimeoutMillis(nil)
	}
	if settings.StartStableWindowMs > 0 {
		a.config.Browser.StartStableWindowMs = settings.StartStableWindowMs
	} else if a.config.Browser.StartStableWindowMs <= 0 {
		a.config.Browser.StartStableWindowMs = browserStartStableWindowMillis(nil)
	}
	if err := a.config.Save(a.resolveAppPath("config.yaml")); err != nil {
		log.Error("浏览器配置保存失败", logger.F("error", err))
		return err
	}
	return nil
}

func (a *App) BrowserCoreList() []BrowserCore {
	return a.browserMgr.ListCores()
}

func (a *App) BrowserCoreSave(input BrowserCoreInput) error {
	return a.browserMgr.SaveCore(input)
}

func (a *App) BrowserCoreDelete(coreId string) error {
	return a.browserMgr.DeleteCore(coreId)
}

func (a *App) BrowserCoreSetDefault(coreId string) error {
	return a.browserMgr.SetDefaultCore(coreId)
}

func (a *App) BrowserCoreValidate(corePath string) BrowserCoreValidateResult {
	return a.browserMgr.ValidateCorePath(corePath)
}

func (a *App) BrowserCoreExtendedInfo() []BrowserCoreExtendedInfo {
	return a.browserMgr.GetCoresExtendedInfo()
}

// BrowserCoreScan 重新扫描 chrome 目录，自动注册新内核
func (a *App) BrowserCoreScan() []BrowserCore {
	a.autoDetectCores()
	return a.browserMgr.ListCores()
}

// BrowserCoreImportLocal 选择一个已解压内核目录或归档文件并注册。
func (a *App) BrowserCoreImportLocal() (*BrowserCore, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app context is nil")
	}
	if a.browserMgr == nil {
		return nil, fmt.Errorf("browser manager is nil")
	}

	var selectedPath string
	var err error
	if goruntime.GOOS == "darwin" {
		const (
			chooseAppBundle = "选择应用包或目录"
			chooseArchive   = "选择压缩归档"
			cancelImport    = "取消"
		)
		choice, dialogErr := wailsruntime.MessageDialog(a.ctx, wailsruntime.MessageDialogOptions{
			Type:          wailsruntime.QuestionDialog,
			Title:         "导入本地内核",
			Message:       "请选择 macOS 内核来源。DMG 无需解压：请先双击挂载，再选择其中的 Chromium.app；应用包会自动复制到本地内核目录。",
			Buttons:       []string{chooseAppBundle, chooseArchive, cancelImport},
			DefaultButton: chooseAppBundle,
			CancelButton:  cancelImport,
		})
		if dialogErr != nil {
			return nil, dialogErr
		}
		switch choice {
		case chooseAppBundle:
			selectedPath, err = wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
				Title:                      "选择 Chromium.app 或已解压内核目录",
				TreatPackagesAsDirectories: true,
			})
		case chooseArchive:
			// Wails v2.12 converts every pattern to a UTType on macOS. Composite
			// extensions such as *.tar.gz and wildcards such as *.* produce nil
			// UTTypes and abort the host process, so use only valid single extensions.
			selectedPath, err = wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
				Title: "选择 Chrome 内核压缩归档",
				Filters: []wailsruntime.FileFilter{
					{DisplayName: "Chrome 内核归档", Pattern: "*.zip;*.tar;*.gz;*.tgz;*.xz;*.txz;*.bz2;*.tbz2"},
				},
			})
		default:
			return nil, nil
		}
	} else {
		selectedPath, err = wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
			Title: "选择 Chrome 内核归档文件",
			Filters: []wailsruntime.FileFilter{
				{DisplayName: "Chrome 内核归档 (" + browser.SupportedCoreArchiveDescription() + ")", Pattern: browser.SupportedCoreArchivePattern()},
				{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
			},
		})
	}
	if err != nil {
		return nil, err
	}
	selectedPath = strings.TrimSpace(selectedPath)
	if selectedPath == "" {
		return nil, nil
	}

	absPath, err := filepath.Abs(selectedPath)
	if err != nil {
		return nil, err
	}
	if info, statErr := os.Stat(absPath); statErr == nil && info.IsDir() {
		return a.importLocalBrowserCoreDirectory(absPath)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return nil, statErr
	}
	if goruntime.GOOS == "darwin" && strings.EqualFold(filepath.Ext(absPath), ".dmg") {
		return nil, fmt.Errorf("macOS .dmg 无需解压：请先双击挂载，再返回选择其中的 Chromium.app；应用会自动复制该内核")
	}
	return a.importLocalBrowserCoreArchive(absPath)
}

func (a *App) importLocalBrowserCoreArchive(archivePath string) (*BrowserCore, error) {
	archiveName := strings.TrimSpace(filepath.Base(archivePath))
	coreName := strings.TrimSpace(coreNameFromArchiveName(archiveName))
	if coreName == "" {
		coreName = "本地内核"
	}

	targetCorePath := filepath.Join("chrome", coreName)
	targetDir := a.browserMgr.ResolveRelativePath(targetCorePath)
	if _, err := os.Stat(targetDir); err == nil {
		return nil, fmt.Errorf("同名内核目录已存在：%s", targetCorePath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, err
	}
	tempExtractDir, err := os.MkdirTemp(parentDir, coreName+"_import_*")
	if err != nil {
		return nil, err
	}
	cleanupTempExtract := true
	defer func() {
		if cleanupTempExtract {
			_ = os.RemoveAll(tempExtractDir)
		}
	}()

	a.emitBrowserCoreImportProgress("extracting", 0, "开始解压本地内核包...")
	if err := browser.ExtractCoreArchiveAndStripRootForImport(archivePath, tempExtractDir, func(progress int, message string) {
		a.emitBrowserCoreImportProgress("extracting", progress, message)
	}); err != nil {
		a.emitBrowserCoreImportProgress("error", 0, "解压失败: "+err.Error())
		return nil, fmt.Errorf("解压失败: %w", err)
	}
	a.emitBrowserCoreImportProgress("validating", 90, "正在校验内核可执行文件...")
	if _, _, ok := browser.FindCoreExecutable(tempExtractDir); !ok {
		err := fmt.Errorf("所选归档不是当前平台可用的内核包：当前平台 %s，未找到浏览器可执行文件（候选：%s）", browser.CoreExecutablePlatform(), strings.Join(browser.CoreExecutableCandidates(), ", "))
		a.emitBrowserCoreImportProgress("error", 0, err.Error())
		return nil, err
	}
	a.emitBrowserCoreImportProgress("saving", 95, "正在保存内核配置...")
	if err := os.Rename(tempExtractDir, targetDir); err != nil {
		a.emitBrowserCoreImportProgress("error", 0, "保存内核目录失败: "+err.Error())
		return nil, err
	}
	cleanupTempExtract = false

	input := browser.CoreInput{
		CoreName:  coreName,
		CorePath:  targetCorePath,
		IsDefault: len(a.browserMgr.ListCores()) == 0,
	}
	core, err := a.savePublishedImportedCore(input, targetDir)
	if err != nil {
		a.emitBrowserCoreImportProgress("error", 0, "保存配置失败: "+err.Error())
		return nil, err
	}
	a.emitBrowserCoreImportProgress("done", 100, "导入完成")
	return core, nil
}

func (a *App) emitBrowserCoreImportProgress(phase string, progress int, message string) {
	if a == nil || a.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(a.ctx, "core-import:progress", map[string]interface{}{
		"phase":    phase,
		"progress": progress,
		"message":  message,
	})
}

func coreNameFromArchiveName(name string) string {
	name = strings.TrimSpace(name)
	for _, suffix := range []string{".tar.gz", ".tar.xz", ".tar.bz2", ".tgz", ".txz", ".tbz2", ".zip", ".tar"} {
		if strings.HasSuffix(strings.ToLower(name), suffix) {
			return strings.TrimSpace(name[:len(name)-len(suffix)])
		}
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// BrowserCoreImportLocalDirectory 选择一个已解压内核目录并注册。
func (a *App) BrowserCoreImportLocalDirectory() (*BrowserCore, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app context is nil")
	}
	if a.browserMgr == nil {
		return nil, fmt.Errorf("browser manager is nil")
	}

	selectedDir, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:                      "选择已解压的 Chrome 内核目录",
		TreatPackagesAsDirectories: goruntime.GOOS == "darwin",
	})
	if err != nil {
		return nil, err
	}
	selectedDir = strings.TrimSpace(selectedDir)
	if selectedDir == "" {
		return nil, nil
	}

	absDir, err := filepath.Abs(selectedDir)
	if err != nil {
		return nil, err
	}
	return a.importLocalBrowserCoreDirectory(absDir)
}

func (a *App) importLocalBrowserCoreDirectory(absDir string) (*BrowserCore, error) {
	executablePath, _, ok := browser.FindCoreExecutable(absDir)
	if !ok {
		return nil, fmt.Errorf("所选目录不是当前平台可用的内核目录：当前平台 %s，未找到浏览器可执行文件（候选：%s）", browser.CoreExecutablePlatform(), strings.Join(browser.CoreExecutableCandidates(), ", "))
	}
	if goruntime.GOOS == "darwin" {
		appBundlePath, shouldCopy, resolveErr := darwinMountedAppBundleForImport(absDir, executablePath)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if shouldCopy {
			return a.importMountedDarwinAppBundle(appBundlePath)
		}
	}

	corePath := a.relativeCorePathIfPossible(absDir)
	coreName := strings.TrimSpace(filepath.Base(absDir))
	if coreName == "" || coreName == "." || coreName == string(filepath.Separator) {
		coreName = "本地内核"
	}

	for _, existing := range a.browserMgr.ListCores() {
		if normalizeCorePathForCompare(existing.CorePath) == normalizeCorePathForCompare(corePath) {
			return &existing, nil
		}
	}

	input := browser.CoreInput{
		CoreName:  coreName,
		CorePath:  corePath,
		IsDefault: len(a.browserMgr.ListCores()) == 0,
	}
	if err := a.browserMgr.SaveCore(input); err != nil {
		return nil, err
	}

	for _, saved := range a.browserMgr.ListCores() {
		if normalizeCorePathForCompare(saved.CorePath) == normalizeCorePathForCompare(corePath) {
			return &saved, nil
		}
	}
	return nil, fmt.Errorf("本地内核已保存但未能读取结果")
}

func (a *App) importMountedDarwinAppBundle(appBundlePath string) (*BrowserCore, error) {
	appBundlePath = strings.TrimSpace(appBundlePath)
	appBundleName := filepath.Base(appBundlePath)
	coreName := strings.TrimSpace(strings.TrimSuffix(appBundleName, filepath.Ext(appBundleName)))
	if coreName == "" {
		coreName = "本地内核"
	}

	targetCorePath := filepath.Join("chrome", coreName)
	targetDir := a.browserMgr.ResolveRelativePath(targetCorePath)
	if _, err := os.Stat(targetDir); err == nil {
		return nil, fmt.Errorf("同名内核目录已存在：%s", targetCorePath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, err
	}
	tempDir, err := os.MkdirTemp(parentDir, coreName+"_import_*")
	if err != nil {
		return nil, err
	}
	cleanupTempDir := true
	defer func() {
		if cleanupTempDir {
			_ = os.RemoveAll(tempDir)
		}
	}()

	a.emitBrowserCoreImportProgress("copying", 10, "正在复制 macOS 应用包...")
	targetAppBundlePath := filepath.Join(tempDir, appBundleName)
	output, err := exec.Command("/usr/bin/ditto", appBundlePath, targetAppBundlePath).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			err = fmt.Errorf("%w: %s", err, message)
		}
		a.emitBrowserCoreImportProgress("error", 0, "复制 macOS 应用包失败: "+err.Error())
		return nil, fmt.Errorf("复制 macOS 应用包失败: %w", err)
	}
	a.emitBrowserCoreImportProgress("validating", 90, "正在校验复制后的内核...")
	if _, _, ok := browser.FindCoreExecutable(tempDir); !ok {
		err := fmt.Errorf("复制后的应用包不是当前平台可用的内核：当前平台 %s，未找到浏览器可执行文件（候选：%s）", browser.CoreExecutablePlatform(), strings.Join(browser.CoreExecutableCandidates(), ", "))
		a.emitBrowserCoreImportProgress("error", 0, err.Error())
		return nil, err
	}
	if err := os.Rename(tempDir, targetDir); err != nil {
		a.emitBrowserCoreImportProgress("error", 0, "保存内核目录失败: "+err.Error())
		return nil, err
	}
	cleanupTempDir = false

	input := browser.CoreInput{
		CoreName:  coreName,
		CorePath:  targetCorePath,
		IsDefault: len(a.browserMgr.ListCores()) == 0,
	}
	core, err := a.savePublishedImportedCore(input, targetDir)
	if err != nil {
		a.emitBrowserCoreImportProgress("error", 0, "保存配置失败: "+err.Error())
		return nil, err
	}
	a.emitBrowserCoreImportProgress("done", 100, "导入完成")
	return core, nil
}

// savePublishedImportedCore registers a directory that has already been made
// visible with os.Rename. A failed save must not strand a same-name directory
// that cannot be selected again. A successful save with a transient list
// failure is deliberately left in place because its configuration may already
// be durable even though it cannot be read back immediately.
func (a *App) savePublishedImportedCore(input browser.CoreInput, publishedDir string) (*BrowserCore, error) {
	var coresBeforeSave []browser.Core
	if a.browserMgr.CoreDAO == nil && a.browserMgr.Config != nil {
		coresBeforeSave = append([]browser.Core(nil), a.browserMgr.Config.Browser.Cores...)
	}
	if err := a.browserMgr.SaveCore(input); err != nil {
		// A DAO upsert can be durable even when its return path reports an
		// error. The config-file fallback mutates memory before Save returns,
		// so its in-memory match is not evidence of persistence.
		if a.browserMgr.CoreDAO != nil {
			if saved, ok := a.browserCorePersistedByPath(input.CorePath); ok {
				return &saved, nil
			}
		} else if a.browserMgr.Config != nil {
			a.browserMgr.Config.Browser.Cores = coresBeforeSave
		}
		if cleanupErr := os.RemoveAll(publishedDir); cleanupErr != nil {
			return nil, fmt.Errorf("保存内核配置失败: %w；清理未注册内核目录失败: %v", err, cleanupErr)
		}
		return nil, err
	}
	if saved, ok := a.browserCoreByPath(input.CorePath); ok {
		return &saved, nil
	}
	return nil, fmt.Errorf("本地内核已保存但未能读取结果")
}

func (a *App) browserCoreByPath(corePath string) (BrowserCore, bool) {
	for _, saved := range a.browserMgr.ListCores() {
		if normalizeCorePathForCompare(saved.CorePath) == normalizeCorePathForCompare(corePath) {
			return saved, true
		}
	}
	return BrowserCore{}, false
}

func (a *App) browserCorePersistedByPath(corePath string) (BrowserCore, bool) {
	if a.browserMgr.CoreDAO == nil {
		return BrowserCore{}, false
	}
	cores, err := a.browserMgr.CoreDAO.List()
	if err != nil {
		return BrowserCore{}, false
	}
	for _, saved := range cores {
		if normalizeCorePathForCompare(saved.CorePath) == normalizeCorePathForCompare(corePath) {
			return BrowserCore(saved), true
		}
	}
	return BrowserCore{}, false
}

func darwinAppBundleRoot(path string) string {
	for path = filepath.Clean(strings.TrimSpace(path)); path != ""; {
		if strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".app") {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return ""
}

// darwinMountedAppBundleForImport resolves both the selected source and its
// executable before deciding whether a mounted app bundle should be copied.
// Mounted sources may not escape their /Volumes mount through symlinks.
func darwinMountedAppBundleForImport(sourceDir, executablePath string) (string, bool, error) {
	sourceDir = filepath.Clean(strings.TrimSpace(sourceDir))
	sourceAppBundlePath := darwinAppBundleRoot(executablePath)
	resolvedSourceDir := sourceDir
	if resolvedPath, err := filepath.EvalSymlinks(resolvedSourceDir); err == nil {
		resolvedSourceDir = resolvedPath
	} else if darwinVolumesMount(resolvedSourceDir) != "" {
		return "", false, fmt.Errorf("无法解析挂载卷中的内核目录: %w", err)
	}

	resolvedExecutablePath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return "", false, fmt.Errorf("无法解析内核可执行文件: %w", err)
	}
	appBundlePath := darwinAppBundleRoot(resolvedExecutablePath)
	if appBundlePath == "" {
		if darwinVolumesMount(resolvedSourceDir) != "" || darwinVolumesMount(resolvedExecutablePath) != "" {
			return "", false, fmt.Errorf("挂载卷中的内核可执行文件不在 .app 应用包内")
		}
		return "", false, nil
	}
	resolvedAppBundlePath, err := filepath.EvalSymlinks(appBundlePath)
	if err != nil {
		return "", false, fmt.Errorf("无法解析 macOS 应用包: %w", err)
	}
	if sourceAppBundlePath != "" {
		resolvedSourceAppBundlePath, resolveErr := filepath.EvalSymlinks(sourceAppBundlePath)
		if resolveErr != nil {
			return "", false, fmt.Errorf("无法解析所选 macOS 应用包: %w", resolveErr)
		}
		if filepath.Clean(resolvedSourceAppBundlePath) != filepath.Clean(resolvedAppBundlePath) {
			return "", false, fmt.Errorf("内核可执行文件通过符号链接离开了所选 .app 应用包")
		}
	}
	mountCheckSourceDir := resolvedSourceDir
	if darwinVolumesMount(sourceDir) != "" {
		mountCheckSourceDir = sourceDir
	}
	return validateDarwinMountedAppBundle(mountCheckSourceDir, resolvedExecutablePath, resolvedAppBundlePath)
}

func validateDarwinMountedAppBundle(sourceDir, resolvedExecutablePath, resolvedAppBundlePath string) (string, bool, error) {
	resolvedAppBundlePath = filepath.Clean(strings.TrimSpace(resolvedAppBundlePath))
	if resolvedAppBundlePath != "" && !strings.HasSuffix(strings.ToLower(filepath.Base(resolvedAppBundlePath)), ".app") {
		return "", false, fmt.Errorf("解析后的 macOS 应用包不是 .app 目录: %s", resolvedAppBundlePath)
	}

	sourceMount := darwinVolumesMount(sourceDir)
	executableMount := darwinVolumesMount(resolvedExecutablePath)
	bundleMount := darwinVolumesMount(resolvedAppBundlePath)
	if sourceMount == "" && executableMount == "" && bundleMount == "" {
		return "", false, nil
	}
	if executableMount == "" || bundleMount == "" || executableMount != bundleMount {
		return "", false, fmt.Errorf("挂载卷中的内核可执行文件或应用包通过符号链接越过了 /Volumes 挂载边界")
	}
	if sourceMount != "" && sourceMount != executableMount {
		return "", false, fmt.Errorf("挂载卷中的内核可执行文件通过符号链接离开了所选挂载卷")
	}
	if !pathWithinDirectory(resolvedExecutablePath, resolvedAppBundlePath) {
		return "", false, fmt.Errorf("挂载卷中的内核可执行文件不在解析后的应用包内")
	}
	return resolvedAppBundlePath, true, nil
}

// darwinVolumesMount returns the normalized /Volumes/<mount-name> ancestor.
// It intentionally does not treat /Volumes itself as a mount.
func darwinVolumesMount(path string) string {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || !filepath.IsAbs(path) || !pathWithinDirectory(path, "/Volumes") {
		return ""
	}
	rel, err := filepath.Rel("/Volumes", path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return ""
	}
	mountName := strings.Split(filepath.ToSlash(rel), "/")[0]
	if mountName == "" || mountName == "." || mountName == ".." {
		return ""
	}
	return filepath.Join("/Volumes", mountName)
}

func pathWithinDirectory(path, root string) bool {
	path = strings.TrimSpace(path)
	root = strings.TrimSpace(root)
	if path == "" || root == "" {
		return false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	return rel != ".." && !strings.HasPrefix(rel, "../") && !filepath.IsAbs(rel)
}

func (a *App) relativeCorePathIfPossible(absDir string) string {
	for _, root := range []string{a.appRootAbs(), a.appStateRootAbs()} {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(rootAbs, absDir)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(filepath.ToSlash(rel), "../") && !filepath.IsAbs(rel) {
			return filepath.ToSlash(rel)
		}
	}
	return absDir
}

// BrowserCoreDownload 在线下载并自动解压配置内核
func (a *App) BrowserCoreDownload(coreName, url, proxyConfig string) error {
	if a.ctx == nil {
		return fmt.Errorf("app context is nil")
	}
	go a.browserMgr.DownloadAndExtractCore(a.ctx, coreName, url, proxyConfig)
	return nil
}

// BrowserCoreRedownload 重新下载并替换指定内核目录
func (a *App) BrowserCoreRedownload(coreId, url, proxyConfig string) error {
	if a.ctx == nil {
		return fmt.Errorf("app context is nil")
	}
	go a.browserMgr.RedownloadCore(a.ctx, coreId, url, proxyConfig)
	return nil
}
