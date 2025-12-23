package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

func (s *Server) segmentsHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Find all .log files in controller-logs recursively
	var segmentFiles []string
	searchDir := filepath.Join(s.bundlePath, "controller-logs")
	
	// Check if directory exists
	if _, err := os.Stat(searchDir); err == nil {
		err := filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !d.IsDir() {
				if strings.HasSuffix(d.Name(), ".log") || d.Name() == "snapshot" {
					rel, _ := filepath.Rel(s.bundlePath, path)
					segmentFiles = append(segmentFiles, rel)
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Warn("Error walking controller-logs", "error", err)
		}
	} else {
		// Fallback: Try finding any .log file if controller-logs doesn't exist
		// This is useful for bundles that might just have logs at root
		filepath.WalkDir(s.bundlePath, func(path string, d os.DirEntry, err error) error {
			if err != nil { return nil }
			if !d.IsDir() {
				if strings.HasSuffix(d.Name(), ".log") || d.Name() == "snapshot" {
					// Avoid adding thousands of data segment logs if we are scanning root
					// Just add a few as a fallback or restrict to specific patterns
					if len(segmentFiles) < 50 {
						rel, _ := filepath.Rel(s.bundlePath, path)
						segmentFiles = append(segmentFiles, rel)
					}
				}
			}
			return nil
		})
	}

	type SegmentsPageData struct {
		NodeHostname string
		Files        []string
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	data := SegmentsPageData{
		NodeHostname: s.nodeHostname,
		Files:        segmentFiles,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	s.segmentsTemplate.Execute(buf, data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}

func (s *Server) segmentViewHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	relPath := r.URL.Query().Get("file")
	if relPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	// Security: Prevent breaking out of bundle
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.bundlePath, cleanPath)
	
	var info models.SegmentInfo
	var err error

	if filepath.Base(cleanPath) == "snapshot" {
		info, err = parser.ParseSnapshot(fullPath)
	} else {
		info, err = parser.ParseLogSegment(fullPath)
	}

	if err != nil && info.Batches == nil && info.Snapshot == nil { // Partial results ok
		http.Error(w, fmt.Sprintf("Failed to parse segment: %v", err), http.StatusInternalServerError)
		return
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	s.segmentViewTemplate.Execute(buf, info) // We'll need a separate template or partial for this
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}
