package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alextreichler/bundleViewer/internal/downloader"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) cpuProfilesHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"ActivePath": s.activePath,
		"LogsOnly":   s.logsOnly,
	}

	if err := s.cpuProfilesTemplate.Execute(w, data); err != nil {
		s.logger.Error("Failed to execute cpu profiles template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
	if err := json.NewEncoder(w).Encode(details); err != nil {
		s.logger.Error("Failed to encode cpu profile details", "error", err)
	}
}

func (s *Server) getProfileVersionAndArch() (string, string, error) {
	// Find a cpu profile file
	pattern := filepath.Join(s.bundlePath, "admin", "cpu_profile_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return "", "", fmt.Errorf("no cpu profile files found")
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		return "", "", err
	}

	var profile models.CpuProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return "", "", err
	}

	return profile.Version, profile.Arch, nil
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists":  exists,
		"path":    path,
		"version": version,
		"arch":    arch,
	})
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"path":   path,
	})
}
