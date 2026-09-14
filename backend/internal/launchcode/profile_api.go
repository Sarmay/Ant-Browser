package launchcode

import (
	"net/http"

	"ant-chrome/backend/internal/browser"
)

// ProfileWriteRequest 用于创建/更新实例配置。
// profile 为持久化配置；start 为本次自动启动的临时参数。
type ProfileWriteRequest struct {
	Profile    *browser.ProfileInput `json:"profile"`
	LaunchCode string                `json:"launchCode"`
	AutoLaunch bool                  `json:"autoLaunch"`
	Start      *LaunchRequestParams  `json:"start"`
}

// ProfileCopyRequest 用于从已有实例复制配置。
// 空请求体合法：名称自动生成，指纹模式默认 auto_fingerprint。
// profileName 是新实例名称；源实例用 selector / code / profileId 选择。
type ProfileCopyRequest struct {
	ProfileName       string               `json:"profileName"`
	Mode              string               `json:"mode"`
	AutomationTargets []string             `json:"automationTargets"`
	LaunchCode        string               `json:"launchCode"`
	AutoLaunch        bool                 `json:"autoLaunch"`
	Start             *LaunchRequestParams `json:"start"`
	ProxyId           string               `json:"proxyId"`
	ProxyConfig       string               `json:"proxyConfig"`
	Selector          *LaunchSelector      `json:"selector"`
	Code              string               `json:"code"`
	Key               string               `json:"key"`
	ProfileID         string               `json:"profileId"`
	Keyword           string               `json:"keyword"`
	Keywords          []string             `json:"keywords"`
	Tag               string               `json:"tag"`
	Tags              []string             `json:"tags"`
	GroupID           string               `json:"groupId"`
	MatchMode         string               `json:"matchMode"`
}

type profileCreator interface {
	CreateProfile(input browser.ProfileInput) (*browser.Profile, error)
}

type profileCopier interface {
	CopyProfileWithOptions(profileID string, newName string, options browser.ProfileCopyOptions) (*browser.Profile, error)
}

type profileUpdater interface {
	UpdateProfile(profileID string, input browser.ProfileInput) (*browser.Profile, error)
}

type profileDeleter interface {
	DeleteProfile(profileID string) error
}

func (s *LaunchServer) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListProfiles(w, r)
	case http.MethodPost:
		s.handleCreateProfile(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"ok":    false,
			"error": "method not allowed",
		})
	}
}

func (s *LaunchServer) handleProfilesCopy(w http.ResponseWriter, r *http.Request) {
	s.handleCopyProfile(w, r, "")
}

func (s *LaunchServer) handleProfileByID(w http.ResponseWriter, r *http.Request) {
	profileID, action, ok := parseProfilePath(r.URL.Path)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"ok":    false,
			"error": "profile not found",
		})
		return
	}

	switch action {
	case "status":
		s.handleProfileStatus(w, r, profileID)
		return
	case "stop":
		s.handleStopProfile(w, r, profileID)
		return
	case "copy":
		s.handleCopyProfile(w, r, profileID)
		return
	case "restart":
		s.handleRestartProfile(w, r, profileID)
		return
	case "open-url":
		s.handleOpenProfileURL(w, r, profileID)
		return
	case "cookies":
		s.handleProfileCookies(w, r, profileID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetProfile(w, r, profileID)
	case http.MethodPut:
		s.handleUpdateProfile(w, r, profileID)
	case http.MethodDelete:
		s.handleDeleteProfile(w, r, profileID)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"ok":    false,
			"error": "method not allowed",
		})
	}
}
