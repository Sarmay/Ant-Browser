package launchcode

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"ant-chrome/backend/internal/browser"
)

func decodeProfileWriteRequest(r *http.Request) (ProfileWriteRequest, int, string) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return ProfileWriteRequest{}, http.StatusMethodNotAllowed, "method not allowed"
	}

	var req ProfileWriteRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return ProfileWriteRequest{}, http.StatusBadRequest, "invalid request body"
	}
	if req.Profile == nil {
		return ProfileWriteRequest{}, http.StatusBadRequest, "profile is required"
	}
	return req, http.StatusOK, ""
}

func decodeProfileCopyRequest(r *http.Request) (ProfileCopyRequest, int, string) {
	if r.Method != http.MethodPost {
		return ProfileCopyRequest{}, http.StatusMethodNotAllowed, "method not allowed"
	}
	if r.Body == nil {
		return ProfileCopyRequest{}, http.StatusOK, ""
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return ProfileCopyRequest{}, http.StatusBadRequest, "invalid request body"
	}
	if strings.TrimSpace(string(body)) == "" {
		return ProfileCopyRequest{}, http.StatusOK, ""
	}

	var req ProfileCopyRequest
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return ProfileCopyRequest{}, http.StatusBadRequest, "invalid request body"
	}
	return req, http.StatusOK, ""
}

func profileCopyOptionsFromRequest(req ProfileCopyRequest, mode string) browser.ProfileCopyOptions {
	return browser.ProfileCopyOptions{
		Mode:              mode,
		AutomationTargets: req.AutomationTargets,
		ProxyId:           strings.TrimSpace(req.ProxyId),
		ProxyConfig:       strings.TrimSpace(req.ProxyConfig),
	}
}

func mergeCopySelector(req ProfileCopyRequest) LaunchSelector {
	var nested LaunchSelector
	if req.Selector != nil {
		nested = *req.Selector
	}
	return normalizeRuntimeSelector(buildMergedSelector(selectorMergeInput{
		Code:        firstNonEmpty(nested.Code, req.Code),
		Key:         firstNonEmpty(nested.Key, req.Key),
		ProfileID:   firstNonEmpty(nested.ProfileID, req.ProfileID),
		ProfileName: nested.ProfileName,
		Keywords:    appendSelectorTerms(nil, "", nested.Keywords, nested.Keyword, req.Keyword, req.Keywords),
		Tags:        appendSelectorTerms(nil, nested.Tag, nested.Tags, req.Tag, req.Tags),
		GroupID:     firstNonEmpty(nested.GroupID, req.GroupID),
		MatchMode:   firstNonEmpty(nested.MatchMode, req.MatchMode),
	}))
}

func normalizeProfileCopyMode(mode string) (string, int, string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto_fingerprint":
		return "auto_fingerprint", http.StatusOK, ""
	case "regular":
		return "regular", http.StatusOK, ""
	default:
		return "", http.StatusBadRequest, "mode must be auto_fingerprint or regular"
	}
}

func normalizeProfileInput(input browser.ProfileInput) browser.ProfileInput {
	return browser.ProfileInput{
		ProfileName:        strings.TrimSpace(input.ProfileName),
		UserDataDir:        strings.TrimSpace(input.UserDataDir),
		CoreId:             strings.TrimSpace(input.CoreId),
		RestoreLastSession: browser.NormalizeRestoreLastSessionMode(input.RestoreLastSession),
		FingerprintArgs:    normalizeStringSlice(input.FingerprintArgs),
		ProxyId:            strings.TrimSpace(input.ProxyId),
		ProxyConfig:        strings.TrimSpace(input.ProxyConfig),
		MemoryLimitMB:      input.MemoryLimitMB,
		LaunchArgs:         normalizeStringSlice(input.LaunchArgs),
		Tags:               normalizeStringSlice(input.Tags),
		Keywords:           normalizeStringSlice(input.Keywords),
		GroupId:            strings.TrimSpace(input.GroupId),
	}
}

func profileToInput(profile *browser.Profile) browser.ProfileInput {
	if profile == nil {
		return browser.ProfileInput{}
	}
	return browser.ProfileInput{
		ProfileName:        strings.TrimSpace(profile.ProfileName),
		UserDataDir:        strings.TrimSpace(profile.UserDataDir),
		CoreId:             strings.TrimSpace(profile.CoreId),
		RestoreLastSession: browser.NormalizeRestoreLastSessionMode(profile.RestoreLastSession),
		FingerprintArgs:    append([]string{}, profile.FingerprintArgs...),
		ProxyId:            strings.TrimSpace(profile.ProxyId),
		ProxyConfig:        strings.TrimSpace(profile.ProxyConfig),
		MemoryLimitMB:      profile.MemoryLimitMB,
		LaunchArgs:         append([]string{}, profile.LaunchArgs...),
		Tags:               append([]string{}, profile.Tags...),
		Keywords:           append([]string{}, profile.Keywords...),
		GroupId:            strings.TrimSpace(profile.GroupId),
	}
}

func mergeProfileRuntime(target, runtimeProfile *browser.Profile) {
	if target == nil || runtimeProfile == nil {
		return
	}
	target.Running = runtimeProfile.Running
	target.DebugPort = runtimeProfile.DebugPort
	target.DebugReady = runtimeProfile.DebugReady
	target.Pid = runtimeProfile.Pid
	target.RuntimeWarning = runtimeProfile.RuntimeWarning
	target.LastError = runtimeProfile.LastError
	target.LastStartAt = runtimeProfile.LastStartAt
	target.LastStopAt = runtimeProfile.LastStopAt
}

func parseProfilePath(path string) (string, string, bool) {
	path = strings.TrimPrefix(path, "/api/profiles/")
	path = strings.Trim(path, "/")
	path = strings.TrimSpace(path)
	if path == "" {
		return "", "", false
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), "", strings.TrimSpace(parts[0]) != ""
	}
	if len(parts) == 2 {
		profileID := strings.TrimSpace(parts[0])
		action := strings.ToLower(strings.TrimSpace(parts[1]))
		if profileID == "" || action == "" {
			return "", "", false
		}
		switch action {
		case "status", "stop", "copy", "restart", "open-url", "cookies":
			return profileID, action, true
		default:
			return "", "", false
		}
	}
	return "", "", false
}

func parseProfilePathID(path string) (string, bool) {
	profileID, action, ok := parseProfilePath(path)
	if !ok || action != "" {
		return "", false
	}
	return profileID, true
}

func mapProfileWriteErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case msg == strings.ToLower(strings.TrimSpace(http.ErrNotSupported.Error())):
		return http.StatusServiceUnavailable
	case strings.Contains(msg, "profile not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "running profile cannot be deleted"):
		return http.StatusConflict
	case strings.Contains(msg, "launch code already exists"):
		return http.StatusConflict
	case strings.Contains(msg, "launch code format invalid"),
		strings.Contains(msg, "launch code must be"):
		return http.StatusBadRequest
	case strings.Contains(msg, "proxy id not found"),
		strings.Contains(msg, "代理id不存在"):
		return http.StatusBadRequest
	case strings.Contains(msg, "复制实例失败"):
		return http.StatusBadRequest
	case strings.Contains(msg, "目标地址不能为空"),
		strings.Contains(msg, "url is required"),
		strings.Contains(msg, "cookies is required"),
		strings.Contains(msg, "invalid netscape"):
		return http.StatusBadRequest
	case strings.Contains(msg, "实例当前未运行"),
		strings.Contains(msg, "实例未运行"),
		strings.Contains(msg, "调试接口尚未就绪"):
		return http.StatusConflict
	case strings.Contains(msg, "实例数量已达上限"):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
