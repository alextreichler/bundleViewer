package server

import (
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/models"
)

type TopicDiskInfo struct {
	Name         string
	TotalSize    int64
	Partitions   map[string]int64
	Active       bool
	SegmentCount int
}

func (s *Server) diskOverviewHandler(w http.ResponseWriter, r *http.Request) {
	resourceUsage := s.cachedData.ResourceUsage
	
	// Aggregate disk stats across all nodes
	var dataDiskStats models.DiskStats
	for _, stats := range s.cachedData.DataDiskStats {
		dataDiskStats.TotalBytes += stats.TotalBytes
		dataDiskStats.FreeBytes += stats.FreeBytes
	}
	
	var cacheDiskStats models.DiskStats
	for _, stats := range s.cachedData.CacheDiskStats {
		cacheDiskStats.TotalBytes += stats.TotalBytes
		cacheDiskStats.FreeBytes += stats.FreeBytes
	}

	duEntries := s.cachedData.DuEntries

	topicDiskUsage := make(map[string]int64)
	partitionDiskUsage := make(map[string]map[string]int64)

	kafkaDataPathPrefix := "/usrdata/redpanda/data/kafka/"
	if s.cachedData.RedpandaDataDir != "" {
		kafkaDataPathPrefix = filepath.Join(s.cachedData.RedpandaDataDir, "kafka")
		if !strings.HasSuffix(kafkaDataPathPrefix, "/") {
			kafkaDataPathPrefix += "/"
		}
	}

	// Auto-detect prefix if duEntries don't match the expected prefix
	hasMatch := false
	if len(duEntries) > 0 {
		for _, entry := range duEntries {
			if strings.HasPrefix(entry.Path, kafkaDataPathPrefix) {
				hasMatch = true
				break
			}
		}

		if !hasMatch {
			// Try to find a common pattern
			for _, entry := range duEntries {
				if idx := strings.Index(entry.Path, "/kafka/"); idx != -1 {
					possiblePrefix := entry.Path[:idx+7]
					kafkaDataPathPrefix = possiblePrefix
					break
				}
			}
		}
	}

	storageAnalysis, err := analysis.AnalyzeStorage(s.bundlePath)
	if err != nil {
		s.logger.Warn("Failed to analyze storage", "error", err)
	}

	// Create a map for fast lookup of storage stats
	storageStatsMap := make(map[string]models.TopicStorageStats)
	if storageAnalysis != nil {
		for _, stat := range storageAnalysis.Topics {
			storageStatsMap[stat.TopicName] = stat
		}
	}

	for _, entry := range duEntries {
		relativePath := ""
		if strings.HasPrefix(entry.Path, kafkaDataPathPrefix) {
			relativePath = strings.TrimPrefix(entry.Path, kafkaDataPathPrefix)
		} else if idx := strings.Index(entry.Path, "/data/kafka/"); idx != -1 {
			relativePath = entry.Path[idx+12:]
		}

		if relativePath != "" {
			parts := strings.Split(relativePath, "/")

			if len(parts) >= 1 {
				topicName := parts[0]
				topicDiskUsage[topicName] += entry.Size

				if len(parts) >= 2 {
					partitionID := parts[1]
					if _, ok := partitionDiskUsage[topicName]; !ok {
						partitionDiskUsage[topicName] = make(map[string]int64)
					}
					partitionDiskUsage[topicName][partitionID] += entry.Size
				}
			}
		}
	}

	// Ensure all topics from storage analysis are included, even if missing from du.txt
	for topicName, stat := range storageStatsMap {
		if _, exists := topicDiskUsage[topicName]; !exists {
			topicDiskUsage[topicName] = stat.SizeBytes // Use size from analysis as fallback
		}
	}

	var topicsDiskInfo []TopicDiskInfo
	for topicName, totalSize := range topicDiskUsage {
		info := TopicDiskInfo{
			Name:       topicName,
			TotalSize:  totalSize,
			Partitions: partitionDiskUsage[topicName],
		}
		
		// Enrich with storage analysis data
		if stat, ok := storageStatsMap[topicName]; ok {
			info.Active = stat.Active
			info.SegmentCount = stat.SegmentCount
		} else {
			// Default values if not in storage analysis (assume active? or unknown)
			// Generally, if it's on disk but not in partitions list, it's likely orphaned/inactive, 
			// but strict "Active" logic depends on cluster_partitions.json.
			// If we didn't find it in storageStatsMap, we didn't process it there.
			// Let's default to false for safety if analysis ran.
			info.Active = false 
		}

		topicsDiskInfo = append(topicsDiskInfo, info)
	}
	sort.Slice(topicsDiskInfo, func(i, j int) bool {
		// Sort by size descending (matching the old "Detailed Analysis" behavior which is often more useful)
		// or Name? The previous "Disk Usage" table was sorted by Name.
		// The user asked to combine them. "Disk Usage" table was Name sorted. 
		// "Detailed" was Size sorted.
		// Let's stick to Name sorting as it's cleaner for a main list, or maybe Size is better for disk analysis.
		// I'll keep the original sorting (Name) of the first table to minimize disruption, unless I change it.
		// Original code: return topicsDiskInfo[i].Name < topicsDiskInfo[j].Name
		return topicsDiskInfo[i].Name < topicsDiskInfo[j].Name
	})

	type DiskPageData struct {
		ResourceUsage   models.ResourceUsage
		DataDiskStats   models.DiskStats
		CacheDiskStats  models.DiskStats
		TopicsDiskInfo  []TopicDiskInfo
		StorageAnalysis *models.StorageAnalysis
		NodeHostname    string
		Sessions        map[string]*BundleSession
		ActivePath      string
		LogsOnly        bool
	}

	// storageAnalysis is already computed above

	pageData := DiskPageData{
		ResourceUsage:   resourceUsage,
		DataDiskStats:   dataDiskStats,
		CacheDiskStats:  cacheDiskStats,
		TopicsDiskInfo:  topicsDiskInfo,
		StorageAnalysis: storageAnalysis,
		NodeHostname:    s.nodeHostname,
		Sessions:        s.sessions,
		ActivePath:      s.activePath,
		LogsOnly:        s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.diskTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute disk template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write disk overview response", "error", err)
	}
}
