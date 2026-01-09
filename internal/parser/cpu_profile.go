package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// ParseCpuProfiles reads and parses all cpu_profile_*.json files in the admin directory.
func ParseCpuProfiles(bundlePath string) ([]models.CpuProfileEntry, error) {
	pattern := filepath.Join(bundlePath, "admin", "cpu_profile_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob cpu profile files: %w", err)
	}

	var allEntries []models.CpuProfileEntry

	for _, file := range files {
		// Extract node name from filename
		// Format: cpu_profile_<node_name>.json or cpu_profile_https---<node_name>.json
		// The example file was: cpu_profile_https---redpanda-observability-0.messaging-redpanda-observability.backend.ppe.gaincapital.cloud.json
		baseName := filepath.Base(file)
		nodeName := strings.TrimPrefix(baseName, "cpu_profile_")
		nodeName = strings.TrimSuffix(nodeName, ".json")
		
		// Clean up node name if it has common prefixes/suffixes from the bundle collection
		// e.g. "https---"
		if strings.HasPrefix(nodeName, "https---") {
			nodeName = strings.TrimPrefix(nodeName, "https---")
		}
		// e.g. port suffix like ".-9644" if present (seen in errors.txt)
		if idx := strings.LastIndex(nodeName, ".-"); idx != -1 {
			nodeName = nodeName[:idx]
		}


		data, err := os.ReadFile(file)
		if err != nil {
			slog.Warn("Failed to read cpu profile file", "path", file, "error", err)
			continue
		}

		var profile models.CpuProfile
		if err := json.Unmarshal(data, &profile); err != nil {
			slog.Warn("Failed to unmarshal cpu profile", "path", file, "error", err)
			continue
		}

		for _, shardProfile := range profile.Profile {
			for _, sample := range shardProfile.Samples {
				entry := models.CpuProfileEntry{
					Node:            nodeName,
					ShardID:         shardProfile.ShardID,
					SchedulingGroup: sample.SchedulingGroup,
					UserBacktrace:   sample.UserBacktrace,
					Occurrences:     sample.Occurrences,
				}
				allEntries = append(allEntries, entry)
			}
		}
	}

	return allEntries, nil
}
