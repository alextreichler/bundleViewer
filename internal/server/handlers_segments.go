package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/parser"
)

func (s *Server) segmentViewHandler(w http.ResponseWriter, r *http.Request) {
	// Expect query param ?path=...
	relativePath := r.URL.Query().Get("path")
	if relativePath == "" {
		http.Error(w, "Missing 'path' parameter", http.StatusBadRequest)
		return
	}
	
	// Security: basic directory traversal check
	if strings.Contains(relativePath, "..") {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.bundlePath, relativePath)
	
	// Determine type
	ext := filepath.Ext(fullPath)
	
	type PageData struct {
		BasePageData
		NodeHostname string
		FilePath     string
		FileName     string
		Type         string
		Data         interface{}
		Error        string
	}

	data := PageData{
		BasePageData: s.newBasePageData("Segments"),
		NodeHostname: s.nodeHostname,
		FilePath:     relativePath,
		FileName:     filepath.Base(fullPath),
	}

	if ext == ".log" {
		data.Type = "Segment"
		info, err := parser.ParseLogSegment(fullPath)
		if err != nil {
			data.Error = fmt.Sprintf("Failed to parse log segment: %v", err)
			// Still return info if we have partial data (Batches might be populated)
			data.Data = info
		} else {
			data.Data = info
		}
	} else if strings.HasSuffix(fullPath, ".snapshot") {
		data.Type = "Snapshot"
		info, err := parser.ParseSnapshot(fullPath)
		if err != nil {
			data.Error = fmt.Sprintf("Failed to parse snapshot: %v", err)
		}
		data.Data = info
	} else if strings.Contains(fullPath, "index") {
		// Index file: TODO impl parser
		data.Type = "Index"
		data.Error = "Index parsing not yet implemented"
	} else {
		data.Type = "Unknown"
		data.Error = "Unsupported file type"
	}
	
	// Handle JSON request (for AJAX/HTMX potentially)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	// Render Template
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)
	
	if err := s.segmentViewTemplate.Execute(buf, data); err != nil {
		s.logger.Error("Failed to render segment view", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}