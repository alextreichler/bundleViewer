package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/downloader"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) cpuProfilesHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	pageData := s.newBasePageData("CPU Profiles")

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	if err := s.cpuProfilesTemplate.Execute(buf, pageData); err != nil {
		s.logger.Error("Failed to execute cpu profiles template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write cpu profiles response", "error", err)
	}
}

func (s *Server) apiCpuProfilesHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "Store not initialized", http.StatusInternalServerError)
		return
	}

	profiles, err := s.store.GetCpuProfiles()
	if err != nil {
		s.logger.Error("Failed to get cpu profiles", "error", err)
		http.Error(w, "Failed to get cpu profiles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(profiles); err != nil {
		s.logger.Error("Failed to encode cpu profiles", "error", err)
	}
}

func (s *Server) apiCpuProfileDetailsHandler(w http.ResponseWriter, r *http.Request) {
	node := r.URL.Query().Get("node")
	shard := r.URL.Query().Get("shard")
	group := r.URL.Query().Get("group")

	if node == "" {
		http.Error(w, "Missing node parameter", http.StatusBadRequest)
		return
	}
	if shard == "" {
		http.Error(w, "Missing shard parameter", http.StatusBadRequest)
		return
	}
	if group == "" {
		http.Error(w, "Missing group parameter", http.StatusBadRequest)
		return
	}

	// Parse shard to int
	var shardID int
	if _, err := fmt.Sscanf(shard, "%d", &shardID); err != nil {
		http.Error(w, "Invalid shard parameter", http.StatusBadRequest)
		return
	}

	details, err := s.store.GetCpuProfileDetails(node, shardID, group)
	if err != nil {
		s.logger.Error("Failed to get cpu profile details", "error", err)
		http.Error(w, "Failed to get cpu profile details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	// Add insights to each detail
	for i := range details {
		details[i].Insight = getProfileInsight(details[i].UserBacktrace, group)
	}

	if err := json.NewEncoder(w).Encode(details); err != nil {
		s.logger.Error("Failed to encode cpu profile details", "error", err)
	}
}

func getProfileInsight(backtrace, group string) string {
	bt := strings.ToLower(backtrace)

	// 1. Schema Registry Bottleneck (Stone-X case)
	if strings.Contains(bt, "schema_registry") || strings.Contains(bt, "build_file_with_refs") {
		return "⚠️ Schema Registry Bottleneck: Heavy activity detected in schema normalization or reference building. Known issue in v25.1.1 (Fixed in v25.1.2)."
	}

	// 2. State Machine Recovery (Akamai case)
	if strings.Contains(bt, "state_machine_manager::apply") || strings.Contains(bt, "raft::state_machine") {
		return "⚠️ Raft State Machine Loop: Time spent in STM apply may indicate a recovery loop or high state machine churn."
	}

	// 3. Segment Handling (Crane case)
	if strings.Contains(bt, "persistent_size") && group == "main" {
		return "⚠️ Segment Fragmentation: High overhead in main group during segment size calculation. Often indicates too many small segment files."
	}

	// 4. SSL/TLS Handshake churn
	if strings.Contains(bt, "ssl_do_handshake") || strings.Contains(bt, "ssl_accept") {
		return "ℹ️ SSL/TLS Churn: High frequency of TLS handshakes. Check for clients not using connection pooling."
	}

	// 5. Seastar Reactor spinning (WorldQuant case)
	if strings.Contains(bt, "epoll_wait") || strings.Contains(bt, "reactor::run") {
		return "ℹ️ Seastar Reactor Activity: If this shard is showing 100% utilization while idle, it may be hit by a known Seastar reactor spinning bug."
	}

	return ""
}

func (s *Server) getProfileVersionAndArch() (string, string, error) {
	// Find a cpu profile file
	pattern := filepath.Join(s.bundlePath, "admin", "cpu_profile_*.json")
	files, err := filepath.Glob(pattern)
	if err == nil && len(files) > 0 {
		data, err := os.ReadFile(files[0])
		if err == nil {
			// Try single object first
			var profile models.CpuProfile
			if err := json.Unmarshal(data, &profile); err == nil && profile.Version != "" {
				return profile.Version, profile.Arch, nil
			}

			// Try array
			var profiles []models.CpuProfile
			if err := json.Unmarshal(data, &profiles); err == nil && len(profiles) > 0 && profiles[0].Version != "" {
				return profiles[0].Version, profiles[0].Arch, nil
			}
		}
	}

	// Fallback 1: Detect from brokers.json
	version := s.detectVersionFromBrokers()
	if version == "" {
		// Fallback 2: Detect from logs
		version, _ = s.detectVersionFromLogs()
	}

	if version != "" {
		// Arch fallback
		arch := "amd64"
		if s.cachedData != nil && s.cachedData.System.Uname.Machine != "" {
			machine := s.cachedData.System.Uname.Machine
			if machine == "aarch64" || machine == "arm64" {
				arch = "arm64"
			} else {
				arch = "amd64" // default to amd64 for x86_64
			}
		}
		return version, arch, nil
	}

	return "", "", fmt.Errorf("failed to parse cpu profile version and could not detect from brokers or logs")
}

func (s *Server) detectVersionFromBrokers() string {
	if s.cachedData == nil {
		return ""
	}

	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "brokers.json" {
			if brokers, ok := file.Data.([]interface{}); ok {
				for _, brokerData := range brokers {
					if broker, ok := brokerData.(map[string]interface{}); ok {
						if v, ok := broker["version"].(string); ok && v != "" {
							return v
						}
					}
				}
			}
			break
		}
	}
	return ""
}

func (s *Server) detectVersionFromLogs() (string, error) {
	if s.store == nil {
		return "", fmt.Errorf("store not available")
	}

	// Search for "Welcome to Redpanda v"
	filter := &models.LogFilter{
		Search: "\"Welcome to Redpanda v\"",
		Limit:  1,
	}

	logs, _, err := s.store.GetLogs(filter)
	if err != nil {
		return "", err
	}

	if len(logs) == 0 {
		return "", fmt.Errorf("version string not found in logs")
	}

	// Message: "Welcome to Redpanda v22.3.11 - 1234..."
	msg := logs[0].Message
	re := regexp.MustCompile(`Welcome to Redpanda v([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("version pattern mismatch")
}

func (s *Server) apiCpuBinaryStatusHandler(w http.ResponseWriter, r *http.Request) {
	version, arch, err := s.getProfileVersionAndArch()
	if err != nil {
		http.Error(w, "Failed to determine version: "+err.Error(), http.StatusNotFound)
		return
	}

	path := downloader.GetBinaryPath(s.dataDir, version, arch)
	_, err = os.Stat(path)
	exists := err == nil

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"exists":  exists,
		"path":    path,
		"version": version,
		"arch":    arch,
	}); err != nil {
		s.logger.Error("Failed to encode cpu binary status", "error", err)
	}
}

func (s *Server) apiCpuDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	version, arch, err := s.getProfileVersionAndArch()
	if err != nil {
		http.Error(w, "Failed to determine version: "+err.Error(), http.StatusNotFound)
		return
	}

	path, err := downloader.DownloadRedpanda(s.dataDir, version, arch)
	if err != nil {
		s.logger.Error("Download failed", "error", err)
		http.Error(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"path":   path,
	}); err != nil {
		s.logger.Error("Failed to encode cpu download response", "error", err)
	}
}