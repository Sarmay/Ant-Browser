package launchcode

import (
	"net/http"
	"strings"
	"time"

	"ant-chrome/backend/internal/logger"
)

// handleCreateProfile POST /api/profiles
func (s *LaunchServer) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.New("LaunchServer")
	startAt := time.Now()

	req, status, errMsg := decodeProfileWriteRequest(r)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	input := normalizeProfileInput(*req.Profile)
	profile, launchCode, status, errMsg := s.createProfile(input, req.LaunchCode)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	launchedProfile, launched, launchErr := s.maybeAutoLaunchProfile(profile, req)
	if launchErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":          false,
			"created":     true,
			"updated":     false,
			"launched":    false,
			"profileId":   profile.ProfileId,
			"profileName": profile.ProfileName,
			"launchCode":  launchCode,
			"profile":     profile,
			"error":       launchErr.Error(),
		})
		log.Warn("Profile API 创建后自动启动失败",
			logger.F("profile_id", profile.ProfileId),
			logger.F("profile_name", profile.ProfileName),
			logger.F("launch_code", launchCode),
			logger.F("duration_ms", time.Since(startAt).Milliseconds()),
			logger.F("error", launchErr.Error()),
		)
		return
	}
	if launched {
		mergeProfileRuntime(profile, launchedProfile)
		s.SetActiveProfile(profile)
	}

	writeJSON(w, http.StatusCreated, s.profileWriteSuccessPayload(profile, launchCode, true, false, launched))
	log.Info("Profile API 创建实例",
		logger.F("profile_id", profile.ProfileId),
		logger.F("profile_name", profile.ProfileName),
		logger.F("launch_code", launchCode),
		logger.F("auto_launch", launched),
		logger.F("duration_ms", time.Since(startAt).Milliseconds()),
	)
}

func (s *LaunchServer) handleListProfiles(w http.ResponseWriter, _ *http.Request) {
	items, status, errMsg := s.listProfiles()
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"count": len(items),
		"items": items,
	})
}

func (s *LaunchServer) handleGetProfile(w http.ResponseWriter, _ *http.Request, profileID string) {
	profile, status, errMsg := s.profileSnapshotByID(profileID)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"profileId":   profile.ProfileId,
		"profileName": profile.ProfileName,
		"launchCode":  profile.LaunchCode,
		"profile":     profile,
	})
}

func (s *LaunchServer) handleUpdateProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	log := logger.New("LaunchServer")
	startAt := time.Now()

	previous, status, errMsg := s.profileSnapshotByID(profileID)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	req, status, errMsg := decodeProfileWriteRequest(r)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	input := normalizeProfileInput(*req.Profile)
	profile, launchCode, status, errMsg := s.updateProfile(profileID, input, req.LaunchCode, previous)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	launchedProfile, launched, launchErr := s.maybeAutoLaunchProfile(profile, req)
	if launchErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":          false,
			"created":     false,
			"updated":     true,
			"launched":    false,
			"profileId":   profile.ProfileId,
			"profileName": profile.ProfileName,
			"launchCode":  launchCode,
			"profile":     profile,
			"error":       launchErr.Error(),
		})
		log.Warn("Profile API 更新后自动启动失败",
			logger.F("profile_id", profile.ProfileId),
			logger.F("profile_name", profile.ProfileName),
			logger.F("launch_code", launchCode),
			logger.F("duration_ms", time.Since(startAt).Milliseconds()),
			logger.F("error", launchErr.Error()),
		)
		return
	}
	if launched {
		mergeProfileRuntime(profile, launchedProfile)
		s.SetActiveProfile(profile)
	}

	writeJSON(w, http.StatusOK, s.profileWriteSuccessPayload(profile, launchCode, false, true, launched))
	log.Info("Profile API 更新实例",
		logger.F("profile_id", profile.ProfileId),
		logger.F("profile_name", profile.ProfileName),
		logger.F("launch_code", launchCode),
		logger.F("auto_launch", launched),
		logger.F("duration_ms", time.Since(startAt).Milliseconds()),
	)
}

func (s *LaunchServer) handleDeleteProfile(w http.ResponseWriter, _ *http.Request, profileID string) {
	profile, status, errMsg := s.profileSnapshotByID(profileID)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}
	if profile.Running {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"ok":    false,
			"error": "running profile cannot be deleted",
		})
		return
	}

	if err := s.deleteProfileInternal(profileID); err != nil {
		writeJSON(w, mapProfileWriteErrorStatus(err), map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	if s.service != nil {
		_ = s.service.Remove(profileID)
	}
	s.ClearActiveProfile(profileID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"deleted":     true,
		"profileId":   profileID,
		"profileName": profile.ProfileName,
		"launchCode":  profile.LaunchCode,
	})
}

// handleCopyProfile POST /api/profiles/{profileId}/copy
func (s *LaunchServer) handleCopyProfile(w http.ResponseWriter, r *http.Request, profileID string) {
	log := logger.New("LaunchServer")
	startAt := time.Now()

	req, status, errMsg := decodeProfileCopyRequest(r)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	mode, status, errMsg := normalizeProfileCopyMode(req.Mode)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	sourceID, status, errMsg := s.resolveCopySource(profileID, req)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	profile, launchCode, status, errMsg := s.copyProfile(sourceID, strings.TrimSpace(req.ProfileName), profileCopyOptionsFromRequest(req, mode), req.LaunchCode)
	if errMsg != "" {
		writeJSON(w, status, map[string]interface{}{
			"ok":    false,
			"error": errMsg,
		})
		return
	}

	launchedProfile, launched, launchErr := s.maybeAutoLaunchProfile(profile, ProfileWriteRequest{
		AutoLaunch: req.AutoLaunch,
		Start:      req.Start,
	})
	if launchErr != nil {
		writeJSON(w, http.StatusInternalServerError, s.profileCopyFailurePayload(profile, launchCode, sourceID, launchErr.Error()))
		log.Warn("Profile API 复制后自动启动失败",
			logger.F("source_profile_id", sourceID),
			logger.F("profile_id", profile.ProfileId),
			logger.F("profile_name", profile.ProfileName),
			logger.F("launch_code", launchCode),
			logger.F("duration_ms", time.Since(startAt).Milliseconds()),
			logger.F("error", launchErr.Error()),
		)
		return
	}
	if launched {
		mergeProfileRuntime(profile, launchedProfile)
		s.SetActiveProfile(profile)
	}

	writeJSON(w, http.StatusCreated, s.profileCopySuccessPayload(profile, launchCode, sourceID, launched))
	log.Info("Profile API 复制实例",
		logger.F("source_profile_id", sourceID),
		logger.F("profile_id", profile.ProfileId),
		logger.F("profile_name", profile.ProfileName),
		logger.F("launch_code", launchCode),
		logger.F("mode", mode),
		logger.F("auto_launch", launched),
		logger.F("duration_ms", time.Since(startAt).Milliseconds()),
	)
}
