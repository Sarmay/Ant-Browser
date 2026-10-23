package backend

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// HostFingerprintPreset 是当前这台电脑上可以写进指纹配置、并且浏览器内能核对的设备信息。
type HostFingerprintPreset struct {
	Brand               string `json:"brand"`
	BrandVersion        string `json:"brandVersion"`
	Platform            string `json:"platform"`
	PlatformVersion     string `json:"platformVersion"`
	Language            string `json:"language"`
	AcceptLanguage      string `json:"acceptLanguage"`
	Timezone            string `json:"timezone"`
	HardwareConcurrency string `json:"hardwareConcurrency"`
}

func (a *App) BrowserHostFingerprintPreset() HostFingerprintPreset {
	preset := detectHostFingerprintPreset()
	preset.Brand = "Chrome"
	if version := a.detectInstalledChromeVersion(); version != "" {
		preset.BrandVersion = version
	}
	if preset.Language != "" && preset.AcceptLanguage == "" {
		preset.AcceptLanguage = buildAcceptLanguage(preset.Language)
	}
	return preset
}

func detectHostFingerprintPreset() HostFingerprintPreset {
	language := detectHostLanguage()
	acceptLanguage := ""
	if language != "" {
		acceptLanguage = buildAcceptLanguage(language)
	}
	return HostFingerprintPreset{
		Platform:            detectHostPlatformName(),
		PlatformVersion:     detectHostPlatformVersion(),
		Language:            language,
		AcceptLanguage:      acceptLanguage,
		Timezone:            detectHostTimezone(),
		HardwareConcurrency: strconv.Itoa(runtime.NumCPU()),
	}
}

func detectHostPlatformName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func detectHostPlatformVersion() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err != nil {
			return ""
		}
		return normalizeHostPlatformVersion(string(out))
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "[Environment]::OSVersion.Version.ToString()").Output()
		if err != nil {
			return ""
		}
		return normalizeHostPlatformVersion(string(out))
	default:
		return normalizeHostPlatformVersion(linuxVersionID())
	}
}

func detectHostLanguage() string {
	switch runtime.GOOS {
	case "darwin":
		if tag := localeFromCommand("defaults", "read", "-g", "AppleLocale"); tag != "" {
			return tag
		}
	case "windows":
		if tag := localeFromCommand("powershell", "-NoProfile", "-Command", "(Get-Culture).Name"); tag != "" {
			return tag
		}
	}
	lang := strings.TrimSpace(os.Getenv("LANG"))
	if lang == "" {
		lang = strings.TrimSpace(os.Getenv("LC_ALL"))
	}
	if idx := strings.Index(lang, "."); idx > 0 {
		lang = lang[:idx]
	}
	return normalizeLocaleTag(lang)
}

func detectHostTimezone() string {
	if tz := strings.TrimSpace(os.Getenv("TZ")); looksLikeIANATimezone(tz) {
		return tz
	}
	link, err := os.Readlink("/etc/localtime")
	if err != nil {
		return ""
	}
	const marker = "zoneinfo/"
	index := strings.LastIndex(link, marker)
	if index < 0 {
		return ""
	}
	zone := strings.Trim(link[index+len(marker):], "/")
	if looksLikeIANATimezone(zone) {
		return zone
	}
	return ""
}

func (a *App) detectInstalledChromeVersion() string {
	if a == nil || a.browserMgr == nil {
		return ""
	}
	core, ok := a.browserMgr.GetDefaultCore()
	if !ok {
		return ""
	}
	if version := strings.TrimSpace(a.browserMgr.GetChromeVersion(core.CorePath)); version != "" {
		return version
	}
	root := a.browserMgr.ResolveRelativePath(core.CorePath)
	return chromeVersionFromCoreTree(root)
}

func chromeVersionFromCoreTree(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return ""
	}
	patterns := []string{
		filepath.Join(root, "Chromium.app", "Contents", "Frameworks", "Chromium Framework.framework", "Versions", "*"),
		filepath.Join(root, "Google Chrome.app", "Contents", "Frameworks", "Google Chrome Framework.framework", "Versions", "*"),
		filepath.Join(root, "Chrome.app", "Contents", "Frameworks", "Google Chrome Framework.framework", "Versions", "*"),
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			name := filepath.Base(match)
			if name == "Current" || strings.HasPrefix(name, ".") {
				continue
			}
			if dottedVersionPattern.MatchString(name) {
				return name
			}
		}
	}
	return ""
}

var dottedVersionPattern = regexp.MustCompile(`^\d+(?:\.\d+){1,4}$`)
var hostVersionTokenPattern = regexp.MustCompile(`\d+(?:\.\d+){0,4}`)

func localeFromCommand(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return normalizeLocaleTag(string(out))
}

func linuxVersionID() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "VERSION_ID=") {
			continue
		}
		value := strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"'`)
		return value
	}
	return ""
}

func normalizeHostPlatformVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	match := hostVersionTokenPattern.FindString(strings.ReplaceAll(value, "_", "."))
	if match == "" {
		return ""
	}
	parts := strings.Split(match, ".")
	switch len(parts) {
	case 1:
		return parts[0] + ".0.0"
	case 2:
		return match + ".0"
	default:
		return match
	}
}

func normalizeLocaleTag(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "_", "-")
	if index := strings.Index(value, "."); index > 0 {
		value = value[:index]
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '-' })
	if len(parts) == 0 {
		return ""
	}
	if len(parts) >= 3 && (strings.EqualFold(parts[1], "Hans") || strings.EqualFold(parts[1], "Hant") || strings.EqualFold(parts[1], "Latn")) && len(parts[2]) == 2 {
		return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[2])
	}
	if len(parts) >= 2 && len(parts[1]) == 2 {
		return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[1])
	}
	if len(parts[0]) < 2 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func looksLikeIANATimezone(value string) bool {
	value = strings.TrimSpace(value)
	return value == "UTC" || strings.Contains(value, "/")
}
