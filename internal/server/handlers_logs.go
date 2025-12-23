package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) logsHandler(w http.ResponseWriter, r *http.Request) {
	// Preserving query params for the HTMX request
	query := r.URL.Query().Encode()
	targetURL := "/api/full-logs-page"
	if query != "" {
		targetURL += "?" + query
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Cluster Logs</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    <div id="main-content" hx-get="%s" hx-trigger="load" hx-select="body > *" hx-swap="innerHTML">
        <nav>
            <a href="/">Home</a>
            <a href="/logs" class="active">Logs</a>
            <div class="theme-switch-wrapper">
                <div class="theme-selection-container">
                    <button class="theme-toggle-btn">Themes</button>
                    <div class="theme-menu">
                        <button class="theme-option" data-theme="light">Light</button>
                        <button class="theme-option" data-theme="dark">Dark</button>
                        <button class="theme-option" data-theme="ultra-dark">Ultra Dark</button>
                        <button class="theme-option" data-theme="retro">Retro</button>
                        <button class="theme-option" data-theme="compact">Compact</button>
                        <button class="theme-option" data-theme="powershell">Powershell</button>
                    </div>
                </div>
            </div>
        </nav>
        <div class="container" style="height: 70vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%%;">
                <h2 style="margin-bottom: 1rem;">Loading Logs...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Filtering and retrieving logs from the bundle. This may take a moment for large datasets.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0%% { transform: rotate(0deg); } 100%% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
</body>
</html>`, targetURL)
}

func (s *Server) apiFullLogsPageHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query()

	search := query.Get("search")
	ignore := query.Get("ignore")

	// Merge "ignore" into "search" for the unified UI
	if ignore != "" {
		if search != "" {
			search = fmt.Sprintf("(%s) NOT (%s)", search, ignore)
		} else {
			search = fmt.Sprintf("NOT (%s)", ignore)
		}
		// Clear ignore so it isn't applied twice if the store logic were to change,
		// and so the template doesn't see it if we were to use it.
		ignore = ""
	}

	filter := models.LogFilter{
		Search:    search,
		Ignore:    ignore,         // Should be empty now
		Level:     query["level"], // Get all "level" parameters as a slice
		Node:      query["node"],
		Component: query["component"],
		StartTime: query.Get("startTime"),
		EndTime:   query.Get("endTime"),
		Sort:      query.Get("sort"),
	}

	// Calculate global min and max times for the UI date pickers from the store
	minTime, maxTime, err := s.store.GetMinMaxLogTime()
	if err != nil {
		s.logger.Error("Error getting min/max log times", "error", err)
		http.Error(w, "Failed to retrieve log time range", http.StatusInternalServerError)
		return
	}
	var minTimeStr, maxTimeStr string
	if !minTime.IsZero() && !maxTime.IsZero() {
		minTimeStr = minTime.UTC().Format("2006-01-02T15:04:05")
		maxTimeStr = maxTime.UTC().Format("2006-01-02T15:04:05")
	}

	// Collect unique nodes and components for dropdowns from the store
	nodes, err := s.store.GetDistinctLogNodes()
	if err != nil {
		s.logger.Error("Error getting distinct log nodes", "error", err)
		http.Error(w, "Failed to retrieve distinct log nodes", http.StatusInternalServerError)
		return
	}
	components, err := s.store.GetDistinctLogComponents()
	if err != nil {
		s.logger.Error("Error getting distinct log components", "error", err)
		http.Error(w, "Failed to retrieve distinct log components", http.StatusInternalServerError)
		return
	}

	type LogsPageData struct {
		Filter     models.LogFilter
		Nodes      []string
		Components []string
		MinTime    string
		MaxTime    string
		LogsOnly   bool
		Sessions   map[string]*BundleSession
		ActivePath string
	}

	data := LogsPageData{
		Filter:     filter,
		Nodes:      nodes,
		Components: components,
		MinTime:    minTimeStr,
		MaxTime:    maxTimeStr,
		LogsOnly:   s.logsOnly,
		Sessions:   s.sessions,
		ActivePath: s.activePath,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.logsTemplate.Execute(buf, data)
	if err != nil {
		s.logger.Error("Error executing logs template", "error", err)
		http.Error(w, "Failed to execute logs template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write logs response", "error", err)
	}
}

// apiLogsHandler serves filtered and paginated logs as JSON for infinite scrolling
func (s *Server) apiLogsHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	query := r.URL.Query()

	filter := models.LogFilter{
		Search:    query.Get("search"),
		Ignore:    query.Get("ignore"),
		Level:     query["level"], // Get all "level" parameters as a slice
		Node:      query["node"],
		Component: query["component"],
		StartTime: query.Get("startTime"),
		EndTime:   query.Get("endTime"),
		Sort:      query.Get("sort"),
	}

	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 1000 { // Enforce a reasonable limit
		limit = 100
	}

	filter.Limit = limit
	filter.Offset = offset

	pagedLogs, totalFilteredLogs, err := s.store.GetLogs(&filter)
	if err != nil {
		s.logger.Error("Error getting filtered logs from store", "error", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	type apiLogsResponse struct {
		Logs    []models.LogEntry `json:"logs"`
		Total   int               `json:"total"`
		Offset  int               `json:"offset"`
		Limit   int               `json:"limit"`
		Count   int               `json:"count"`
		HasMore bool              `json:"hasMore"`
	}

	// Convert []*models.LogEntry to []models.LogEntry
	logsForResponse := make([]models.LogEntry, len(pagedLogs))
	for i, logEntry := range pagedLogs {
		logsForResponse[i] = *logEntry
	}

	response := apiLogsResponse{
		Logs:    logsForResponse,
		Total:   totalFilteredLogs,
		Offset:  offset,
		Limit:   limit,
		Count:   len(logsForResponse),
		HasMore: (offset + len(logsForResponse)) < totalFilteredLogs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode logs response", "error", err)
	}
}

// logAnalysisHandler renders the initial page with a loading state
func (s *Server) logAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	// Render the full page with "Loading..." state
	err := s.logAnalysisTemplate.Execute(buf, map[string]interface{}{
		"Partial":    false,
		"LogsOnly":   s.logsOnly,
		"Sessions":   s.sessions,
		"ActivePath": s.activePath,
	})
	if err != nil {
		s.logger.Error("Error executing log analysis template", "error", err)
		http.Error(w, "Failed to execute log analysis template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write log analysis response", "error", err)
	}
}

// apiLogAnalysisDataHandler performs the heavy log analysis and returns the HTML partial
func (s *Server) apiLogAnalysisDataHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	patterns, totalLogs, err := s.store.GetLogPatterns()
	if err != nil {
		s.logger.Error("Error getting log patterns from store", "error", err)
		http.Error(w, "Failed to retrieve log patterns", http.StatusInternalServerError)
		return
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	err = s.logAnalysisTemplate.Execute(buf, map[string]interface{}{
		"Partial":       true,
		"Patterns":      patterns,
		"TotalLogs":     totalLogs,
		"TotalPatterns": len(patterns),
		"Sessions":      s.sessions,
		"ActivePath":    s.activePath,
		"LogsOnly":      s.logsOnly,
	})
	if err != nil {
		s.logger.Error("Error executing log analysis template (partial)", "error", err)
		http.Error(w, "Failed to execute log analysis template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write log analysis data response", "error", err)
	}
}
func (s *Server) apiLogContextHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	requestedPath := query.Get("filePath")
	lineNumberStr := query.Get("lineNumber")
	contextSizeStr := query.Get("contextSize")

	if requestedPath == "" || lineNumberStr == "" {
		http.Error(w, "Missing required parameters: filePath and lineNumber", http.StatusBadRequest)
		return
	}

	// Security: Validate that the file is within the bundle directory
	// Clean the requested path to resolve .. and symlinks roughly
	cleanRequested := filepath.Clean(requestedPath)
	
	// Determine the absolute path of the bundle for comparison
	absBundlePath, err := filepath.Abs(s.bundlePath)
	if err != nil {
		s.logger.Error("Failed to resolve absolute bundle path", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// The requestedPath might be absolute or relative. 
	// If it's absolute, check prefix. If relative, join with bundle path first.
	var absRequested string
	if filepath.IsAbs(cleanRequested) {
		absRequested = cleanRequested
	} else {
		absRequested = filepath.Join(absBundlePath, cleanRequested)
	}

	// Final Clean
	absRequested = filepath.Clean(absRequested)

	// Check if the resolved path starts with the bundle directory
	// We use the volume separator logic to be robust across OSs
	if !strings.HasPrefix(absRequested, absBundlePath) {
		s.logger.Warn("Security: Attempted path traversal", "requested", requestedPath, "resolved", absRequested, "base", absBundlePath)
		http.Error(w, "Forbidden: File access denied", http.StatusForbidden)
		return
	}

	lineNumber, err := strconv.Atoi(lineNumberStr)
	if err != nil {
		http.Error(w, "Invalid lineNumber", http.StatusBadRequest)
		return
	}

	contextSize := 10 // Default context size
	if contextSizeStr != "" {
		contextSize, err = strconv.Atoi(contextSizeStr)
		if err != nil {
			http.Error(w, "Invalid contextSize", http.StatusBadRequest)
			return
		}
	}

	file, err := os.Open(absRequested)
	if err != nil {
		s.logger.Error("Error opening file for context", "path", absRequested, "error", err)
		http.Error(w, "Failed to open log file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("Error scanning file for context", "error", err)
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}

	start := lineNumber - contextSize - 1
	if start < 0 {
		start = 0
	}

	end := lineNumber + contextSize
	if end > len(lines) {
		end = len(lines)
	}

	contextLines := lines[start:end]

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contextLines); err != nil {
		s.logger.Error("Failed to encode log context", "error", err)
	}
}
