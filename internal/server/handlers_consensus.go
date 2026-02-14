package server

import (
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NodeConsensusStat struct {
	Node  string
	Count int
}

type HotspotStat struct {
	Name  string
	Count int
}

type ConsensusSummary struct {
	TotalEvents     int
	ElectedCount    int
	StepdownCount   int
	TimeoutCount    int
	MoveCount       int
	HealthCount     int
	MostActiveNode  string
	MostActiveNTP   string
	StepdownReasons map[string]int
	StartTime       time.Time
	EndTime         time.Time
	NodeStats       []NodeConsensusStat
	TopPartitions   []HotspotStat
	TopTopics       []HotspotStat
}

func (s *Server) consensusHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	events, err := s.store.GetGlobalRaftEvents()
	if err != nil {
		s.logger.Error("Failed to get global raft events", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	summary := ConsensusSummary{
		TotalEvents:     len(events),
		StepdownReasons: make(map[string]int),
	}

	if len(events) > 0 {
		summary.EndTime = events[0].Timestamp
		summary.StartTime = events[len(events)-1].Timestamp
	}

	nodeCounts := make(map[string]int)
	ntpCounts := make(map[string]int)
	topicCounts := make(map[string]int)

	for _, ev := range events {
		switch ev.Type {
		case "Elected":
			summary.ElectedCount++
		case "Stepdown":
			summary.StepdownCount++
			if ev.Reason != "" {
				summary.StepdownReasons[ev.Reason]++
			}
		case "Timeout":
			summary.TimeoutCount++
		case "Move":
			summary.MoveCount++
		case "Health":
			summary.HealthCount++
		}

		nodeCounts[ev.Node]++
		if ev.NTP != "" {
			ntpCounts[ev.NTP]++
			// Extract topic from NTP (format: ns/topic/partition)
			parts := strings.Split(ev.NTP, "/")
			if len(parts) >= 2 {
				topicCounts[parts[1]]++
			}
		}
	}

	// Prepare Node Stats
	for node, count := range nodeCounts {
		summary.NodeStats = append(summary.NodeStats, NodeConsensusStat{Node: node, Count: count})
	}
	sort.Slice(summary.NodeStats, func(i, j int) bool {
		return summary.NodeStats[i].Count > summary.NodeStats[j].Count
	})

	if len(summary.NodeStats) > 0 {
		summary.MostActiveNode = summary.NodeStats[0].Node
	}

	// Prepare Hotspot Stats
	for ntp, count := range ntpCounts {
		summary.TopPartitions = append(summary.TopPartitions, HotspotStat{Name: ntp, Count: count})
	}
	sort.Slice(summary.TopPartitions, func(i, j int) bool {
		return summary.TopPartitions[i].Count > summary.TopPartitions[j].Count
	})
	if len(summary.TopPartitions) > 10 {
		summary.TopPartitions = summary.TopPartitions[:10]
	}

	for topic, count := range topicCounts {
		summary.TopTopics = append(summary.TopTopics, HotspotStat{Name: topic, Count: count})
	}
	sort.Slice(summary.TopTopics, func(i, j int) bool {
		return summary.TopTopics[i].Count > summary.TopTopics[j].Count
	})
	if len(summary.TopTopics) > 10 {
		summary.TopTopics = summary.TopTopics[:10]
	}

	if len(summary.TopPartitions) > 0 {
		summary.MostActiveNTP = summary.TopPartitions[0].Name
	}

	type ConsensusPageData struct {
		BasePageData
		Events  []models.RaftTimelineEvent
		Summary ConsensusSummary
	}

	data := ConsensusPageData{
		BasePageData: s.newBasePageData("Consensus"),
		Events:       events,
		Summary:      summary,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	err = s.consensusTemplate.Execute(buf, data)
	if err != nil {
		s.logger.Error("Error executing consensus template", "error", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write consensus response", "error", err)
	}
}
