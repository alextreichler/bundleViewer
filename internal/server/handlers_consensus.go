package server

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NodeConsensusStat struct {
	Node      string
	Elections int
	Stepdowns int
	Timeouts  int
	Moves     int
	Total     int
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
	StepdownByNode  map[string]map[string]int // Reason -> Node -> Count
	StartTime       time.Time
	EndTime         time.Time
	NodeStats       []NodeConsensusStat
	TopPartitions   []HotspotStat
	TopTopics       []HotspotStat
	Insights        []string
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
		StepdownByNode:  make(map[string]map[string]int),
	}

	if len(events) > 0 {
		summary.EndTime = events[0].Timestamp
		summary.StartTime = events[len(events)-1].Timestamp
	}

	nodeData := make(map[string]*NodeConsensusStat)
	ntpCounts := make(map[string]int)
	topicCounts := make(map[string]int)

	for _, ev := range events {
		if _, ok := nodeData[ev.Node]; !ok {
			nodeData[ev.Node] = &NodeConsensusStat{Node: ev.Node}
		}
		stat := nodeData[ev.Node]
		stat.Total++

		switch ev.Type {
		case "Elected":
			summary.ElectedCount++
			stat.Elections++
		case "Stepdown":
			summary.StepdownCount++
			stat.Stepdowns++
			if ev.Reason != "" {
				summary.StepdownReasons[ev.Reason]++
				
				if summary.StepdownByNode[ev.Reason] == nil {
					summary.StepdownByNode[ev.Reason] = make(map[string]int)
				}
				summary.StepdownByNode[ev.Reason][ev.Node]++
			}
		case "Timeout":
			summary.TimeoutCount++
			stat.Timeouts++
		case "Move":
			summary.MoveCount++
			stat.Moves++
		case "Health":
			summary.HealthCount++
		}

		if ev.NTP != "" {
			ntpCounts[ev.NTP]++
			// Extract topic from NTP (format: ns/topic/partition)
			parts := strings.Split(ev.NTP, "/")
			if len(parts) >= 2 {
				topicCounts[parts[1]]++
			}
		}
	}

	// Finalize Node Stats
	for _, stat := range nodeData {
		summary.NodeStats = append(summary.NodeStats, *stat)
	}
	sort.Slice(summary.NodeStats, func(i, j int) bool {
		return summary.NodeStats[i].Total > summary.NodeStats[j].Total
	})

	if len(summary.NodeStats) > 0 {
		summary.MostActiveNode = summary.NodeStats[0].Node
		
		// Logic 1: The Node Hotspot
		if summary.NodeStats[0].Timeouts > 5 || summary.NodeStats[0].Elections > 10 {
			summary.Insights = append(summary.Insights, fmt.Sprintf("🚩 **Node Hotspot (%s):** This node is experiencing significantly more consensus activity than others. High timeouts and elections suggest it may be the Cluster Controller or is under extreme disk/CPU pressure.", summary.MostActiveNode))
		}
	}

	// Logic 2: Hostile Election Loop
	if summary.ElectedCount > 0 && summary.StepdownCount > 0 {
		diff := summary.ElectedCount - summary.StepdownCount
		if diff < 5 && diff > -5 && summary.ElectedCount > 10 {
			summary.Insights = append(summary.Insights, "🔄 **Election/Stepdown Loop:** The number of elections nearly matches stepdowns. This indicates 'Leadership Churn'—where nodes lose leadership due to transient stalls and immediately re-elect themselves once the stall clears.")
		}
	}

	// Logic 3: RPC Timeouts
	if summary.TimeoutCount > 0 {
		summary.Insights = append(summary.Insights, fmt.Sprintf("⏰ **RPC Timeouts detected:** Found %d raft timeouts. These are leading indicators of metadata pressure or 'Stability Storms' caused by aggressive timeout settings (e.g. 100ms node_status_interval).", summary.TimeoutCount))
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
