package launchcode

import (
	"io/fs"
	"net/http"
	"strings"
)

const docsContentSecurityPolicy = "default-src 'self'; base-uri 'none'; connect-src 'self'; form-action 'none'; frame-ancestors 'none'; img-src 'self' data:; object-src 'none'; script-src 'self'; style-src 'self'"

func (s *LaunchServer) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	s.registerDocsRoutes(mux)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/automation/scripts", s.handleAutomationScripts)
	mux.HandleFunc("/api/automation/scripts/", s.handleAutomationScriptByID)
	mux.HandleFunc("/api/automation/scripts/run", s.handleAutomationScriptRun)
	mux.HandleFunc("/api/automation/scripts/runs", s.handleAutomationScriptRuns)
	mux.HandleFunc("/api/automation/hooks/", s.handleAutomationPublicHook)
	mux.HandleFunc("/api/profiles", s.handleProfiles)
	mux.HandleFunc("/api/profiles/", s.handleProfileByID)
	mux.HandleFunc("/api/runtime/active", s.handleRuntimeActive)
	mux.HandleFunc("/api/runtime/session", s.handleRuntimeSession)
	mux.HandleFunc("/api/runtime/status", s.handleRuntimeStatus)
	mux.HandleFunc("/api/runtime/stop", s.handleRuntimeStop)
	mux.HandleFunc("/api/launch", s.handleLaunchWithBody)
	mux.HandleFunc("/api/launch/logs", s.handleLaunchLogs)
	mux.HandleFunc("/api/launch/", s.handleLaunch)
	mux.HandleFunc("/", s.handleCDPProxy)
	return mux
}

func (s *LaunchServer) registerDocsRoutes(mux *http.ServeMux) {
	docsFS := s.docsFileSystem()
	if docsFS == nil {
		mux.HandleFunc("/docs/", handleDocsUnavailable)
		mux.HandleFunc("/docs", handleDocsUnavailable)
		mux.HandleFunc("/system/api/docs", handleDocsUnavailable)
		return
	}

	docsRoot, err := fs.Sub(docsFS, "docs")
	if err != nil {
		return
	}

	if pageRoot, pageErr := fs.Sub(docsRoot, "docs"); pageErr == nil {
		fileServer := docsSecurityHeaders(http.FileServer(http.FS(pageRoot)))
		mux.Handle("/docs/", http.StripPrefix("/docs/", fileServer))
	}
	if assetRoot, assetErr := fs.Sub(docsRoot, "assets"); assetErr == nil {
		assetServer := docsSecurityHeaders(http.FileServer(http.FS(assetRoot)))
		mux.Handle("/docs/assets/", http.StripPrefix("/docs/assets/", assetServer))
	}
	mux.HandleFunc("/docs/context.json", s.handleDocsContext)
	mux.HandleFunc("/docs", redirectToDocs)
	mux.HandleFunc("/system/api/docs", redirectToDocs)
}

func handleDocsUnavailable(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
		"ok":    false,
		"error": "documentation frontend is unavailable",
	})
}

func docsSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", docsContentSecurityPolicy)
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func (s *LaunchServer) handleDocsContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", docsContentSecurityPolicy)
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"baseUrl":    s.CDPURL(),
		"authHeader": s.APIAuthHeader(),
	})
}

func redirectToDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	target := "/docs/"
	if query := strings.TrimSpace(r.URL.RawQuery); query != "" {
		target += "?" + query
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func (s *LaunchServer) buildHandler(includeLocalhost bool) http.Handler {
	var handler http.Handler = s.buildMux()
	handler = s.apiAuthMiddleware(handler)
	if includeLocalhost {
		handler = s.localhostMiddleware(handler)
	}
	return handler
}

// handleHealth GET /api/health
func (s *LaunchServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// NewTestHandler 返回不含 localhost 限制的 handler，仅供测试使用
func NewTestHandler(s *LaunchServer) http.Handler {
	return s.buildHandler(false)
}
