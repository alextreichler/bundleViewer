package server

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func getSMPCount(bundlePath string) int {
	// Try to find redpanda.yaml in the bundle root
	// We might need to look in other locations if the structure varies
	paths := []string{
		filepath.Join(bundlePath, "redpanda.yaml"),
		filepath.Join(bundlePath, "etc", "redpanda", "redpanda.yaml"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			// Look for "- --smp=14" or "smp: 14"
			// The example showed:
			//     additional_start_flags:
			//         - --smp=14
			
			reSMP := regexp.MustCompile(`--smp=(\d+)`)
			
			for scanner.Scan() {
				line := scanner.Text()
				if matches := reSMP.FindStringSubmatch(line); len(matches) > 1 {
					if val, err := strconv.Atoi(matches[1]); err == nil {
						return val
					}
				}
			}
		}
	}
	return 0
}

func (s *Server) skewHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var brokers []interface{}
	var clusterConfig map[string]interface{}

	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "brokers.json" {
			if bList, ok := file.Data.([]interface{}); ok {
				brokers = bList
			}
		}
		if file.FileName == "cluster_config.json" {
			if configEntries, ok := file.Data.([]models.ClusterConfigEntry); ok {
				clusterConfig = make(map[string]interface{})
				for _, entry := range configEntries {
					clusterConfig[entry.Key] = entry.Value
				}
			} else if config, ok := file.Data.(map[string]interface{}); ok {
				// Fallback in case it was somehow stored as a map (though parser returns slice)
				clusterConfig = config
			}
		}
	}
	
	smpCount := getSMPCount(s.bundlePath)

	skewData := analysis.AnalyzePartitionSkew(
		s.cachedData.Partitions, 
		brokers, 
		clusterConfig, 
		s.cachedData.DataDiskStats, 
		s.cachedData.PartitionBalancer,
		s.cachedData.TopicThroughput,
		smpCount,
	)

	// Sort NodeSkews and balancing nodes by ID for consistent ordering
	sort.Slice(skewData.NodeSkews, func(i, j int) bool {
		return skewData.NodeSkews[i].NodeID < skewData.NodeSkews[j].NodeID
	})
	sort.Slice(skewData.Balancing.Nodes, func(i, j int) bool {
		return skewData.Balancing.Nodes[i].NodeID < skewData.Balancing.Nodes[j].NodeID
	})

	type SkewPageData struct {
		BasePageData
		NodeHostname string
		Analysis     analysis.SkewAnalysis
	}

	pageData := SkewPageData{
		BasePageData: s.newBasePageData("Skew"),
		NodeHostname: s.nodeHostname,
		Analysis:     skewData,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(16384) // Increased buffer size for larger page

	err := s.skewTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing skew template", "error", err)
		http.Error(w, "Failed to execute skew template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write skew response", "error", err)
	}
}
