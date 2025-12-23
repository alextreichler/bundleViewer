package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/diagnostics"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) diagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	report := diagnostics.Audit(s.cachedData, s.store)

	type DiagnosticsPageData struct {
		NodeHostname string
		Report       diagnostics.DiagnosticsReport
		Syslog       models.SyslogAnalysis
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := DiagnosticsPageData{
		NodeHostname: s.nodeHostname,
		Report:       report,
		Syslog:       s.cachedData.System.Syslog,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.diagnosticsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing diagnostics template", "error", err)
		http.Error(w, "Failed to execute diagnostics template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write diagnostics response", "error", err)
	}
}
