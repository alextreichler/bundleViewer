package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Server) storageHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// List all .log files in the bundle
	var logFiles []string
	
	// We want to scan the whole bundle for .log files, but maybe prioritize "kafka" dir
	// Walker function
	err := filepath.WalkDir(s.bundlePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".log") {
			// Make relative path
			rel, _ := filepath.Rel(s.bundlePath, path)
			logFiles = append(logFiles, rel)
		}
		return nil
	})
	if err != nil {
		s.logger.Warn("Failed to walk bundle for log files", "error", err)
	}
	
	sort.Strings(logFiles)

	type StoragePageData struct {
		Files        []string
		NodeHostname string
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := StoragePageData{
		Files:        logFiles,
		NodeHostname: s.nodeHostname,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	if err := s.segmentsTemplate.Execute(buf, pageData); err != nil {
		s.logger.Error("Failed to render storage template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}
