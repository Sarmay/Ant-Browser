package launchcode

import (
	"net/http"
	"strings"

	"ant-chrome/backend/internal/browser"
)

type proxyCatalogItem struct {
	ProxyId     string `json:"proxyId"`
	ProxyName   string `json:"proxyName"`
	ProxyConfig string `json:"proxyConfig"`
	GroupName   string `json:"groupName,omitempty"`
}

type groupCatalogItem struct {
	GroupId       string `json:"groupId"`
	GroupName     string `json:"groupName"`
	ParentId      string `json:"parentId,omitempty"`
	SortOrder     int    `json:"sortOrder"`
	InstanceCount int    `json:"instanceCount"`
}

type coreCatalogItem struct {
	CoreId    string `json:"coreId"`
	CoreName  string `json:"coreName"`
	IsDefault bool   `json:"isDefault"`
}

func (s *LaunchServer) handleProxies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"ok":    false,
			"error": "method not allowed",
		})
		return
	}
	if s.browserMgr == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"error": "proxy catalog is not available",
		})
		return
	}

	fallback := []browser.Proxy{}
	if s.browserMgr.Config != nil {
		fallback = s.browserMgr.Config.Browser.Proxies
	}
	items := make([]proxyCatalogItem, 0)
	for _, proxy := range browser.ListProxiesWithFallback(s.browserMgr.ProxyDAO, fallback) {
		items = append(items, proxyCatalogItem{
			ProxyId:     strings.TrimSpace(proxy.ProxyId),
			ProxyName:   strings.TrimSpace(proxy.ProxyName),
			ProxyConfig: strings.TrimSpace(proxy.ProxyConfig),
			GroupName:   strings.TrimSpace(proxy.GroupName),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"count": len(items),
		"items": items,
	})
}

func (s *LaunchServer) handleGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"ok":    false,
			"error": "method not allowed",
		})
		return
	}
	if s.browserMgr == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"error": "group catalog is not available",
		})
		return
	}

	countMap := map[string]int{}
	for _, profile := range s.profileSnapshots() {
		groupID := strings.TrimSpace(profile.GroupId)
		if groupID == "" {
			continue
		}
		countMap[groupID]++
	}

	items := make([]groupCatalogItem, 0)
	if s.browserMgr.GroupDAO != nil {
		groups, err := s.browserMgr.GroupDAO.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}
		for _, group := range groups {
			if group == nil {
				continue
			}
			items = append(items, groupCatalogItem{
				GroupId:       strings.TrimSpace(group.GroupId),
				GroupName:     strings.TrimSpace(group.GroupName),
				ParentId:      strings.TrimSpace(group.ParentId),
				SortOrder:     group.SortOrder,
				InstanceCount: countMap[strings.TrimSpace(group.GroupId)],
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"count": len(items),
		"items": items,
	})
}

func (s *LaunchServer) handleCores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"ok":    false,
			"error": "method not allowed",
		})
		return
	}
	if s.browserMgr == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"error": "core catalog is not available",
		})
		return
	}

	items := make([]coreCatalogItem, 0)
	for _, core := range s.browserMgr.ListCores() {
		items = append(items, coreCatalogItem{
			CoreId:    strings.TrimSpace(core.CoreId),
			CoreName:  strings.TrimSpace(core.CoreName),
			IsDefault: core.IsDefault,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"count": len(items),
		"items": items,
	})
}
