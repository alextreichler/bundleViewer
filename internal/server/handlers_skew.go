package server

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/analysis"
)

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
			if config, ok := file.Data.(map[string]interface{}); ok {
				clusterConfig = config
			}
		}
	}

	skewData := analysis.AnalyzePartitionSkew(s.cachedData.Partitions, brokers, clusterConfig)

	// Sort NodeSkews and balancing nodes by ID for consistent ordering
	sort.Slice(skewData.NodeSkews, func(i, j int) bool {
		return skewData.NodeSkews[i].NodeID < skewData.NodeSkews[j].NodeID
	})
	sort.Slice(skewData.Balancing.Nodes, func(i, j int) bool {
		return skewData.Balancing.Nodes[i].NodeID < skewData.Balancing.Nodes[j].NodeID
	})

	type SkewPageData struct {
		NodeHostname string
		Analysis     analysis.SkewAnalysis
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := SkewPageData{
		NodeHostname: s.nodeHostname,
		Analysis:     skewData,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
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
