package launchcode

import (
	"net/http"
	"strings"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/logger"
)

func (s *LaunchServer) createProfile(input browser.ProfileInput, requestedCode string) (*browser.Profile, string, int, string) {
	profile, err := s.createProfileInternal(input)
	if err != nil {
		return nil, "", mapProfileWriteErrorStatus(err), err.Error()
	}
	if profile == nil {
		return nil, "", http.StatusInternalServerError, "profile creation returned nil profile"
	}

	launchCode, status, errMsg := s.applyRequestedLaunchCode(profile.ProfileId, strings.TrimSpace(profile.LaunchCode), requestedCode)
	if errMsg != "" {
		_ = s.deleteCreatedProfile(profile.ProfileId)
		return nil, "", status, errMsg
	}
	profile.LaunchCode = launchCode
	return profile, launchCode, http.StatusCreated, ""
}

func (s *LaunchServer) updateProfile(profileID string, input browser.ProfileInput, requestedCode string, previous *browser.Profile) (*browser.Profile, string, int, string) {
	profile, err := s.updateProfileInternal(profileID, input)
	if err != nil {
		return nil, "", mapProfileWriteErrorStatus(err), err.Error()
	}
	if profile == nil {
		return nil, "", http.StatusInternalServerError, "profile update returned nil profile"
	}

	currentCode := ""
	if previous != nil {
		currentCode = strings.TrimSpace(previous.LaunchCode)
	}
	launchCode, status, errMsg := s.applyRequestedLaunchCode(profile.ProfileId, currentCode, requestedCode)
	if errMsg != "" {
		if rollbackErr := s.rollbackProfileUpdate(profileID, previous); rollbackErr != nil {
			logger.New("LaunchServer").Warn("Profile API 更新回滚失败",
				logger.F("profile_id", profileID),
				logger.F("error", rollbackErr.Error()),
			)
		}
		return nil, "", status, errMsg
	}
	profile.LaunchCode = launchCode
	return profile, launchCode, http.StatusOK, ""
}

func (s *LaunchServer) maybeAutoLaunchProfile(profile *browser.Profile, req ProfileWriteRequest) (*browser.Profile, bool, error) {
	if profile == nil || !req.AutoLaunch {
		return nil, false, nil
	}

	params := LaunchRequestParams{}
	if req.Start != nil {
		params = normalizeLaunchRequestParams(*req.Start)
	}

	launchedProfile, err := s.launchProfile(profile.ProfileId, params)
	if err != nil {
		return nil, false, err
	}
	return launchedProfile, true, nil
}

func (s *LaunchServer) resolveCopySource(pathProfileID string, req ProfileCopyRequest) (string, int, string) {
	pathProfileID = strings.TrimSpace(pathProfileID)
	if pathProfileID != "" {
		return pathProfileID, http.StatusOK, ""
	}

	selector := mergeCopySelector(req)
	if selector.IsEmpty() {
		return "", http.StatusBadRequest, "selector is required"
	}
	if err := selector.Validate(); err != nil {
		return "", http.StatusBadRequest, err.Error()
	}
	if selector.MatchMode == launchMatchModeAll {
		return "", http.StatusBadRequest, "matchMode must be unique or first for profile copy"
	}

	profile, status, errMsg := s.findProfileBySelector(selector)
	if errMsg != "" {
		return "", status, errMsg
	}
	return profile.ProfileId, http.StatusOK, ""
}

func (s *LaunchServer) copyProfile(sourceID, newName string, options browser.ProfileCopyOptions, requestedCode string) (*browser.Profile, string, int, string) {
	profile, err := s.copyProfileInternal(sourceID, newName, options)
	if err != nil {
		return nil, "", mapProfileWriteErrorStatus(err), err.Error()
	}
	if profile == nil {
		return nil, "", http.StatusInternalServerError, "profile copy returned nil profile"
	}

	launchCode, status, errMsg := s.applyRequestedLaunchCode(profile.ProfileId, strings.TrimSpace(profile.LaunchCode), requestedCode)
	if errMsg != "" {
		_ = s.deleteCreatedProfile(profile.ProfileId)
		return nil, "", status, errMsg
	}
	profile.LaunchCode = launchCode
	return profile, launchCode, http.StatusCreated, ""
}

func (s *LaunchServer) copyProfileInternal(sourceID, newName string, options browser.ProfileCopyOptions) (*browser.Profile, error) {
	if copier, ok := s.starter.(profileCopier); ok {
		return copier.CopyProfileWithOptions(sourceID, newName, options)
	}
	if s.browserMgr != nil {
		return s.browserMgr.CopyWithOptions(sourceID, newName, options)
	}
	return nil, http.ErrNotSupported
}

func (s *LaunchServer) createProfileInternal(input browser.ProfileInput) (*browser.Profile, error) {
	if creator, ok := s.starter.(profileCreator); ok {
		return creator.CreateProfile(input)
	}
	if s.browserMgr != nil {
		return s.browserMgr.Create(input)
	}
	return nil, http.ErrNotSupported
}

func (s *LaunchServer) updateProfileInternal(profileID string, input browser.ProfileInput) (*browser.Profile, error) {
	if updater, ok := s.starter.(profileUpdater); ok {
		return updater.UpdateProfile(profileID, input)
	}
	if s.browserMgr != nil {
		return s.browserMgr.Update(profileID, input)
	}
	return nil, http.ErrNotSupported
}

func (s *LaunchServer) deleteCreatedProfile(profileID string) error {
	if deleter, ok := s.starter.(profileDeleter); ok {
		return deleter.DeleteProfile(profileID)
	}
	if s.browserMgr != nil {
		return s.browserMgr.Delete(profileID)
	}
	return nil
}

func (s *LaunchServer) deleteProfileInternal(profileID string) error {
	return s.deleteCreatedProfile(profileID)
}

func (s *LaunchServer) rollbackProfileUpdate(profileID string, previous *browser.Profile) error {
	if previous == nil {
		return nil
	}
	_, err := s.updateProfileInternal(profileID, profileToInput(previous))
	return err
}

func (s *LaunchServer) listProfiles() ([]browser.Profile, int, string) {
	if s.browserMgr == nil {
		return nil, http.StatusServiceUnavailable, "profile catalog is not available"
	}

	items := s.browserMgr.List()
	for i := range items {
		items[i].LaunchCode = s.resolveProfileLaunchCode(items[i].ProfileId, items[i].LaunchCode)
	}
	return items, http.StatusOK, ""
}

func (s *LaunchServer) profileSnapshotByID(profileID string) (*browser.Profile, int, string) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil, http.StatusNotFound, "profile not found"
	}
	if s.browserMgr == nil {
		return nil, http.StatusServiceUnavailable, "profile catalog is not available"
	}

	s.browserMgr.Mutex.Lock()
	profile, ok := s.browserMgr.Profiles[profileID]
	var snapshot browser.Profile
	if ok && profile != nil {
		snapshot = *profile
	}
	s.browserMgr.Mutex.Unlock()
	if !ok {
		return nil, http.StatusNotFound, "profile not found"
	}

	snapshot.LaunchCode = s.resolveProfileLaunchCode(snapshot.ProfileId, snapshot.LaunchCode)
	return &snapshot, http.StatusOK, ""
}

func (s *LaunchServer) applyRequestedLaunchCode(profileID, currentCode, requestedCode string) (string, int, string) {
	currentCode = strings.TrimSpace(currentCode)
	requestedCode = strings.TrimSpace(requestedCode)
	if requestedCode == "" {
		return s.resolveProfileLaunchCode(profileID, currentCode), http.StatusOK, ""
	}
	if s.service == nil {
		return "", http.StatusServiceUnavailable, "launch code service is unavailable"
	}

	code, err := s.service.SetCode(profileID, requestedCode)
	if err != nil {
		return "", mapProfileWriteErrorStatus(err), err.Error()
	}
	return code, http.StatusOK, ""
}

func (s *LaunchServer) resolveProfileLaunchCode(profileID, currentCode string) string {
	if trimmed := strings.TrimSpace(currentCode); trimmed != "" {
		return trimmed
	}
	if s.service == nil || strings.TrimSpace(profileID) == "" {
		return ""
	}
	code, err := s.service.EnsureCode(profileID)
	if err != nil {
		return ""
	}
	return code
}

func (s *LaunchServer) profileWriteSuccessPayload(profile *browser.Profile, launchCode string, created bool, updated bool, launched bool) map[string]interface{} {
	payload := map[string]interface{}{
		"ok":          true,
		"created":     created,
		"updated":     updated,
		"launched":    launched,
		"profileId":   profile.ProfileId,
		"profileName": profile.ProfileName,
		"launchCode":  launchCode,
		"profile":     profile,
	}

	if !launched {
		return payload
	}

	for key, value := range s.launchSuccessPayload(profile, launchCode) {
		payload[key] = value
	}
	payload["created"] = created
	payload["updated"] = updated
	payload["launched"] = true
	payload["profile"] = profile
	return payload
}

func (s *LaunchServer) profileCopySuccessPayload(profile *browser.Profile, launchCode, sourceProfileID string, launched bool) map[string]interface{} {
	payload := s.profileWriteSuccessPayload(profile, launchCode, true, false, launched)
	payload["copied"] = true
	payload["sourceProfileId"] = sourceProfileID
	return payload
}

func (s *LaunchServer) profileCopyFailurePayload(profile *browser.Profile, launchCode, sourceProfileID, errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"ok":              false,
		"created":         true,
		"updated":         false,
		"copied":          true,
		"launched":        false,
		"profileId":       profile.ProfileId,
		"profileName":     profile.ProfileName,
		"launchCode":      launchCode,
		"sourceProfileId": sourceProfileID,
		"profile":         profile,
		"error":           errMsg,
	}
}
