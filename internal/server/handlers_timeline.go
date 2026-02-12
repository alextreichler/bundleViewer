package server

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/timeline"
)

func (s *Server) timelineHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	// Render the full page with "Loading..." state
	data := s.newBasePageData("Timeline")
	pageData := struct {
		BasePageData
		NodeHostname string
		Partial      bool
	}{
		BasePageData: data,
		NodeHostname: s.nodeHostname,
		Partial:      false,
	}

	err := s.timelineTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing timeline template", "error", err)
		http.Error(w, "Failed to execute timeline template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write timeline response", "error", err)
	}
}

type AggregatedTimelinePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) apiTimelineAggregatedHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		respondJSON(w, []AggregatedTimelinePoint{})
		return
	}

	// Parse bucket size (default 2m)
	bucketStr := r.URL.Query().Get("bucket")
	bucketSize := 2 * time.Minute
	if bucketStr != "" {
		if d, err := time.ParseDuration(bucketStr); err == nil {
			bucketSize = d
		}
	}

	// 1. Get Log Counts (Errors/Warns)
	logCounts, err := s.store.GetLogCountsByTime(bucketSize)
	if err != nil {
		s.logger.Error("Error getting log counts for timeline", "error", err)
		http.Error(w, "Failed to get log counts", http.StatusInternalServerError)
		return
	}

	// 2. Get K8s Event Counts (Warnings)
	k8sCounts, err := s.store.GetK8sEventCountsByTime(bucketSize)
	if err != nil {
		s.logger.Warn("Error getting k8s event counts for timeline", "error", err)
		// Non-fatal
	}

	// Merge counts
	merged := make(map[time.Time]int)
	for t, c := range logCounts {
		merged[t] += c
	}
	for t, c := range k8sCounts {
		merged[t] += c
	}

	// Convert to sorted slice
	var result []AggregatedTimelinePoint
	for t, c := range merged {
		result = append(result, AggregatedTimelinePoint{Timestamp: t, Count: c})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	respondJSON(w, result)
}

func (s *Server) apiTimelineDataHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		return
	}

	// Prepare source data
	// For now, we mainly use logs. We can expand to K8s events later if we parse them into a usable format.

	// Retrieve logs from the store (filtered for performance)
	// We only want ERROR and WARN logs for the timeline, and we limit the count to prevent OOM
	filter := models.LogFilter{
		Level: []string{"ERROR", "WARN"},
		Limit: 10000, 
	}
	allLogsPtr, _, err := s.store.GetLogs(&filter)
	if err != nil {
		s.logger.Error("Error getting all logs for timeline", "error", err)
		http.Error(w, "Failed to retrieve logs for timeline", http.StatusInternalServerError)
		return
	}

	// Use pointers directly
	sourceData := &models.TimelineSourceData{
		Logs:      allLogsPtr,
		K8sPods:   s.cachedData.K8sStore.Pods,
		K8sEvents: s.cachedData.K8sStore.Events,
	}

	events := timeline.GenerateTimeline(sourceData)

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	type TimelinePageData struct {
		BasePageData
		NodeHostname string
		Events       []timeline.TimelineEvent
		Partial      bool
	}

	data := TimelinePageData{
		BasePageData: s.newBasePageData("Timeline"),
		NodeHostname: s.nodeHostname,
		Events:       events,
		Partial:      true,
	}

	err = s.timelineTemplate.Execute(buf, data)
	if err != nil {
		s.logger.Error("Error executing timeline template (partial)", "error", err)
		http.Error(w, "Failed to execute timeline template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write timeline data response", "error", err)
	}
}
