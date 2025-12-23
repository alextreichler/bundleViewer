package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	var path string
	var err error

	switch runtime.GOOS {
	case "darwin":
		// macOS: Use AppleScript to open folder picker and return POSIX path
		script := `POSIX path of (choose folder with prompt "Select Redpanda Bundle Directory:")`
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err == nil {
			path = strings.TrimSpace(string(output))
		}
	case "windows":
		// Windows: Use PowerShell to open folder picker
		script := `
		Add-Type -AssemblyName System.Windows.Forms
		$FolderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog
		$FolderBrowser.Description = "Select Redpanda Bundle Directory"
		$result = $FolderBrowser.ShowDialog()
		if ($result -eq "OK") {
			Write-Host $FolderBrowser.SelectedPath
		}
		`
		cmd := exec.Command("powershell", "-Command", script)
		output, err := cmd.Output()
		if err == nil {
			path = strings.TrimSpace(string(output))
		}
	default:
		http.Error(w, "Folder picker not supported on this platform. Please enter path manually.", http.StatusNotImplemented)
		return
	}

	if err != nil {
		// User likely cancelled or something went wrong
		s.logger.Debug("Folder picker closed or failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"path": ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": path})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	bundlePath := r.FormValue("bundlePath")
	logsOnly := r.FormValue("logsOnly") == "on"

	if bundlePath == "" {
		http.Error(w, "Bundle path is required", http.StatusBadRequest)
		return
	}

	// 1. Initialize Progress
	s.progress = models.NewProgressTracker(21)

	// 1. Generate unique DB name for this bundle
	bundleName := filepath.Base(bundlePath)
	sanitizedName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, bundleName)
	dbName := fmt.Sprintf("bundle_%s.db", sanitizedName)
	dbPath := filepath.Join(s.dataDir, dbName)

	cleanDB := !s.persist
	if cleanDB {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("Failed to remove old database", "db", dbPath, "error", err)
		}
	}

	sqliteStore, err := store.NewSQLiteStore(dbPath, bundlePath, cleanDB)
	if err != nil {
		s.logger.Error("Failed to initialize SQLite store", "error", err)
		http.Error(w, "Failed to initialize database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Initialize Cache with progress tracker
	cachedData, err := cache.New(bundlePath, sqliteStore, logsOnly, s.progress)
	if err != nil {
		s.logger.Error("Failed to create cache", "error", err)
		http.Error(w, "Failed to analyze bundle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Create and Store Session
	session := &BundleSession{
		Path:         bundlePath,
		Name:         bundleName,
		CachedData:   cachedData,
		Store:        sqliteStore,
		NodeHostname: getNodeHostname(bundlePath, cachedData),
		LogsOnly:     logsOnly,
	}
	s.sessions[bundlePath] = session

	// 4. Update Active Server State
	s.activePath = bundlePath
	s.bundlePath = session.Path
	s.logsOnly = session.LogsOnly
	s.store = session.Store
	s.cachedData = session.CachedData
	s.nodeHostname = session.NodeHostname

	s.progress.Finish()

	// 5. Signal success and redirect via HTMX
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSwitchBundle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	session, exists := s.sessions[path]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update active fields
	s.activePath = path
	s.bundlePath = session.Path
	s.logsOnly = session.LogsOnly
	s.store = session.Store
	s.cachedData = session.CachedData
	s.nodeHostname = session.NodeHostname

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handlePartitionLeadersAPI(w http.ResponseWriter, r *http.Request) {
	leaders := s.cachedData.Leaders

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(leaders); err != nil {
		s.logger.Error("Error encoding partition leaders", "error", err)
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) fullDataHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("fileName")
	jsonPath := r.URL.Query().Get("jsonPath")

	if fileName == "" || jsonPath == "" {
		http.Error(w, "fileName and jsonPath parameters required", http.StatusBadRequest)
		return
	}

	var targetData interface{}
	found := false
	for _, group := range s.cachedData.GroupedFiles {
		for _, file := range group {
			if file.FileName == fileName {
				targetData = file.Data
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		http.Error(w, "File not found in cache", http.StatusNotFound)
		return
	}

	currentData := traverseJSONPath(targetData, jsonPath)
	if currentData == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(currentData); err != nil {
		s.logger.Error("Error encoding full data", "error", err)
	}
}

func (s *Server) progressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	for {
		// Check if connection is still alive
		select {
		case <-r.Context().Done():
			return // Connection closed by client
		default:
			percentage, status, finished := s.progress.Get()
			if _, err := fmt.Fprintf(w, "data: {\"progress\": %d, \"status\": \"%s\", \"finished\": %v}\n\n", percentage, status, finished); err != nil {
				return
			}
			flusher.Flush()
			if finished {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (s *Server) topicDetailsHandler(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	if topicName == "" {
		http.Error(w, "topic parameter required", http.StatusBadRequest)
		return
	}

	partitions := s.cachedData.Partitions
	leaders := s.cachedData.Leaders
	leaderMap := make(map[string]int, len(leaders))
	for _, l := range leaders {
		key := fmt.Sprintf("%s-%d", l.Topic, l.PartitionID)
		leaderMap[key] = l.Leader
	}

	var topicPartitions []models.PartitionInfo
	for _, p := range partitions {
		if p.Topic == topicName {
			key := fmt.Sprintf("%s-%d", p.Topic, p.PartitionID)
			var replicas []int
			for _, r := range p.Replicas {
				replicas = append(replicas, r.NodeID)
			}
			topicPartitions = append(topicPartitions, models.PartitionInfo{
				ID:       p.PartitionID,
				Leader:   leaderMap[key],
				Replicas: replicas,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(topicPartitions); err != nil {
		s.logger.Error("Error encoding topic details", "error", err)
	}
}

// Lazy loading handler for large datasets
func (s *Server) lazyLoadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("fileName")
	jsonPath := r.URL.Query().Get("jsonPath")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	if fileName == "" || jsonPath == "" {
		http.Error(w, "fileName and jsonPath parameters required", http.StatusBadRequest)
		return
	}

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > maxChunkSize {
		limit = maxChunkSize
	}

	// Set timeout for slow operations
	ctx := r.Context()
	done := make(chan bool)
	var targetData interface{}
	var found bool

	go func() {
		// Find the file in cached data
		for _, group := range s.cachedData.GroupedFiles {
			for _, file := range group {
				if file.FileName == fileName {
					targetData = file.Data
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		done <- true
	}()

	select {
	case <-done:
		// Continue processing
	case <-ctx.Done():
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	}

	if !found {
		http.Error(w, "File not found in cache", http.StatusNotFound)
		return
	}

	// Traverse the JSON path
	currentData := traverseJSONPath(targetData, jsonPath)
	if currentData == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	// Extract the requested slice
	var result interface{}
	var total int

	switch v := currentData.(type) {
	case []interface{}:
		total = len(v)
		end := offset + limit
		if end > total {
			end = total
		}
		if offset < total {
			result = v[offset:end]
		} else {
			result = []interface{}{}
		}
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		total = len(keys)
		end := offset + limit
		if end > total {
			end = total
		}
		if offset < total {
			resultMap := make(map[string]interface{})
			for i := offset; i < end; i++ {
				resultMap[keys[i]] = v[keys[i]]
			}
			result = resultMap
		} else {
			result = map[string]interface{}{}
		}
	default:
		result = currentData
		// total = 1 (unused)
	}

	// Render HTML with timeout protection
	htmlChan := make(chan template.HTML, 1)
	go func() {
		htmlChan <- renderDataLazy(fileName, jsonPath, result, 0, offset)
	}()

	var html template.HTML
	select {
	case html = <-htmlChan:
		// Got the HTML
	case <-time.After(10 * time.Second):
		html = template.HTML("<tr><td colspan='2'>Rendering timeout - data too complex</td></tr>")
		s.logger.Warn("Rendering timeout", "fileName", fileName, "jsonPath", jsonPath)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, string(html)); err != nil {
		s.logger.Error("Failed to write lazy load response", "error", err)
	}
}
