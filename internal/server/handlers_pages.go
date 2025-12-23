package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

type RackData struct {
	Count   int
	NodeIDs []int
}

type HomePageData struct {
	GroupedFiles              map[string][]parser.ParsedFile
	NodeHostname              string
	TotalBrokers              int
	Version                   string
	IsHealthy                 bool
	UnderReplicatedPartitions int
	LeaderlessPartitions      int
	NodesInMaintenanceMode    int
	MaintenanceModeNodeIDs    []int
	NodesDown                 int
	RackAwarenessEnabled      bool
	RackInfo                  map[string]RackData
	StartupTime               time.Time
	Sessions                  map[string]*BundleSession
	ActivePath                string
	LogsOnly                  bool
}

func (s *Server) setupHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	data := map[string]interface{}{
		"CanCancel": s.bundlePath != "",
	}

	if err := s.setupTemplate.Execute(buf, data); err != nil {
		http.Error(w, "Failed to execute setup template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write setup response", "error", err)
	}
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		buf := builderPool.Get().(*strings.Builder)
		buf.Reset()
		defer builderPool.Put(buf)
		data := map[string]interface{}{
			"CanCancel": false,
		}
		if err := s.setupTemplate.Execute(buf, data); err != nil {
			http.Error(w, "Failed to execute setup template", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, buf.String())
		return
	}

	if s.logsOnly {
		http.Redirect(w, r, "/logs", http.StatusFound)
		return
	}

	pageData := s.buildHomePageData()
	pageData.Sessions = s.sessions
	pageData.ActivePath = s.activePath

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.homeTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write home response", "error", err)
	}
}

func (s *Server) partitionsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	pageData := s.buildPartitionsPageData()

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.partitionsTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute partitions template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write partitions response", "error", err)
	}
}

func (s *Server) kafkaHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	kafkaMetadata := s.cachedData.KafkaMetadata
	var rawJSON string
	for _, file := range s.cachedData.GroupedFiles["Root"] {
		if file.FileName == "kafka.json" {
			if data, ok := file.Data.(string); ok {
				rawJSON = data
			} else {
				jsonBytes, err := json.MarshalIndent(file.Data, "", "  ")
				if err == nil {
					rawJSON = string(jsonBytes)
				}
			}
			break
		}
	}

	topicConfigsMap := make(map[string]models.TopicConfig)
	for _, tc := range s.cachedData.TopicConfigs {
		topicConfigsMap[tc.Name] = tc
	}

	type KafkaPageData struct {
		Metadata       models.KafkaMetadataResponse
		RawJSON        string
		TopicConfigs   map[string]models.TopicConfig
		ConsumerGroups map[string]models.ConsumerGroup
		Sessions       map[string]*BundleSession
		ActivePath     string
		LogsOnly       bool
	}

	pageData := KafkaPageData{
		Metadata:       kafkaMetadata,
		RawJSON:        rawJSON,
		TopicConfigs:   topicConfigsMap,
		ConsumerGroups: s.cachedData.ConsumerGroups,
		Sessions:       s.sessions,
		ActivePath:     s.activePath,
		LogsOnly:       s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.kafkaTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute kafka template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write kafka response", "error", err)
	}
}

func (s *Server) groupsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type GroupsPageData struct {
		ConsumerGroups map[string]models.ConsumerGroup
		Sessions       map[string]*BundleSession
		ActivePath     string
		LogsOnly       bool
	}

	pageData := GroupsPageData{
		ConsumerGroups: s.cachedData.ConsumerGroups,
		Sessions:       s.sessions,
		ActivePath:     s.activePath,
		LogsOnly:       s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	err := s.groupsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing groups template", "error", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write groups response", "error", err)
	}
}

func (s *Server) k8sHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type K8sPageData struct {
		NodeHostname string
		Store        models.K8sStore
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := K8sPageData{
		NodeHostname: s.nodeHostname,
		Store:        s.cachedData.K8sStore,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.k8sTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing k8s template", "error", err)
		http.Error(w, "Failed to execute k8s template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write k8s response", "error", err)
	}
}

func (s *Server) systemHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type SystemPageData struct {
		NodeHostname string
		System       models.SystemState
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := SystemPageData{
		NodeHostname: s.nodeHostname,
		System:       s.cachedData.System,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.systemTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing system template", "error", err)
		http.Error(w, "Failed to execute system template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write system response", "error", err)
	}
}

// Helpers

type TopicPartitions struct {
	Name       string
	Partitions []models.PartitionInfo
}

type PartitionsPageData struct {
	TotalPartitions   int
	PartitionsPerNode map[int]int
	LeadersPerNode    map[int]int
	Topics            map[string]*TopicPartitions
	NodeHostname      string
	Sessions          map[string]*BundleSession
	ActivePath        string
	LogsOnly          bool
}

func (s *Server) buildHomePageData() HomePageData {
	totalBrokers := len(s.cachedData.KafkaMetadata.Brokers)
	if totalBrokers == 0 {
		totalBrokers = len(s.cachedData.HealthOverview.AllNodes)
	}
	isHealthy := s.cachedData.HealthOverview.IsHealthy
	underReplicatedPartitions := s.cachedData.HealthOverview.UnderReplicatedCount
	leaderlessPartitions := s.cachedData.HealthOverview.LeaderlessCount

	versionSet := make(map[string]struct{})
	versions := []string{}
	nodesInMaintenanceMode := 0
	maintenanceModeNodeIDs := []int{}
	rackInfo := make(map[string]RackData)

	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "brokers.json" {
			if brokers, ok := file.Data.([]interface{}); ok {
				for _, brokerData := range brokers {
					if broker, ok := brokerData.(map[string]interface{}); ok {
						nodeID := -1
						if id, ok := broker["node_id"].(float64); ok {
							nodeID = int(id)
						}

						if v, ok := broker["version"].(string); ok {
							if _, exists := versionSet[v]; !exists {
								versionSet[v] = struct{}{}
								versions = append(versions, v)
							}
						}

						inMaintenance := false
						if ms, ok := broker["maintenance_status"].(map[string]interface{}); ok {
							for _, v := range ms {
								if status, ok := v.(bool); ok && status {
									inMaintenance = true
									break
								}
							}
						}
						// Fallback to draining if maintenance_status didn't trigger it
						if !inMaintenance {
							if draining, ok := broker["draining"].(bool); ok && draining {
								inMaintenance = true
							}
						}

						if inMaintenance {
							nodesInMaintenanceMode++
							if nodeID != -1 {
								maintenanceModeNodeIDs = append(maintenanceModeNodeIDs, nodeID)
							}
						}

						if rack, ok := broker["rack"].(string); ok {
							rackData := rackInfo[rack]
							rackData.Count++
							if nodeID != -1 {
								rackData.NodeIDs = append(rackData.NodeIDs, nodeID)
							}
							rackInfo[rack] = rackData
						}
					}
				}
			}
			break
		}
	}

	versionStr := "N/A"
	if len(versions) > 0 {
		versionStr = strings.Join(versions, ", ")
	}

	rackAwarenessEnabled := false
	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "cluster_config.json" {
			if configs, ok := file.Data.([]models.ClusterConfigEntry); ok {
				for _, config := range configs {
					if config.Key == "enable_rack_awareness" {
						if enabled, ok := config.Value.(bool); ok {
							rackAwarenessEnabled = enabled
						}
						break
					}
				}
			}
			break
		}
	}

	nodesDown := len(s.cachedData.HealthOverview.NodesDown)

	return HomePageData{
		GroupedFiles:              s.cachedData.GroupedFiles,
		NodeHostname:              s.nodeHostname,
		TotalBrokers:              totalBrokers,
		Version:                   versionStr,
		IsHealthy:                 isHealthy,
		UnderReplicatedPartitions: underReplicatedPartitions,
		LeaderlessPartitions:      leaderlessPartitions,
		NodesInMaintenanceMode:    nodesInMaintenanceMode,
		MaintenanceModeNodeIDs:    maintenanceModeNodeIDs,
		NodesDown:                 nodesDown,
		RackAwarenessEnabled:      rackAwarenessEnabled,
		RackInfo:                  rackInfo,
		StartupTime:               time.Now(),
		Sessions:                  s.sessions,
		ActivePath:                s.activePath,
		LogsOnly:                  s.logsOnly,
	}
}

func (s *Server) buildPartitionsPageData() PartitionsPageData {
	partitions := s.cachedData.Partitions
	leaders := s.cachedData.Leaders

	leaderMap := make(map[string]int)
	for _, l := range leaders {
		key := fmt.Sprintf("%s-%d", l.Topic, l.PartitionID)
		leaderMap[key] = l.Leader
	}

	partitionsPerNode := make(map[int]int)
	leadersPerNode := make(map[int]int)
	topics := make(map[string]*TopicPartitions, len(partitions)/10)

	for _, p := range partitions {
		if _, ok := topics[p.Topic]; !ok {
			topics[p.Topic] = &TopicPartitions{
				Name:       p.Topic,
				Partitions: []models.PartitionInfo{},
			}
		}

		replicas := []int{}
		for _, r := range p.Replicas {
			replicas = append(replicas, r.NodeID)
			partitionsPerNode[r.NodeID]++
		}

		key := fmt.Sprintf("%s-%d", p.Topic, p.PartitionID)
		leader := leaderMap[key]
		if leader != -1 {
			leadersPerNode[leader]++
		}

		topics[p.Topic].Partitions = append(topics[p.Topic].Partitions, models.PartitionInfo{
			ID:       p.PartitionID,
			Replicas: replicas,
			Leader:   leader,
		})
	}

	return PartitionsPageData{
		TotalPartitions:   len(partitions),
		PartitionsPerNode: partitionsPerNode,
		LeadersPerNode:    leadersPerNode,
		Topics:            topics,
		NodeHostname:      s.nodeHostname,
		Sessions:          s.sessions,
		ActivePath:        s.activePath,
		LogsOnly:          s.logsOnly,
	}
}