package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/models"
)

type ShardCPUMetric struct {
	ShardID string
	Usage   float64 // Percentage (0-100) based on average over uptime
}

type IOMetric struct {
	ShardID  string
	Class    string
	ReadOps  float64
	WriteOps float64
}

type KafkaLatencyStats struct {
	RequestType string
	AvgLatency  float64 // Seconds
	Count       float64
}

type NetworkMetrics struct {
	KafkaLatencies    []KafkaLatencyStats
	ActiveConnections float64
	RPCErrors         float64
}

type MetricsPageData struct {
	BasePageData
	NodeHostname      string
	ResourceUsage     models.ResourceUsage
	ShardCPU          []ShardCPUMetric
	IOMetrics         []IOMetric
	MaxReadOps        float64 // For bar calculation
	MaxWriteOps       float64 // For bar calculation
	SarData           models.SarData
	PerformanceReport analysis.PerformanceReport
	CoreCount         int
	Network           NetworkMetrics
	RaftRecovery      RaftRecoveryMetrics
	MetricNames       []string // Added for interactive explorer
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Render a lightweight loading page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Metrics Overview</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
    <script src="/static/chart.js"></script>
</head>
<body>
    <div id="main-content" hx-get="/api/full-metrics" hx-trigger="load" hx-select="body > *" hx-swap="innerHTML">
        <nav>
            <a href="/">Home</a>
            <a href="/metrics" class="active">Metrics</a>
            <div class="theme-switch-wrapper">
                <div class="theme-selection-container">
                    <button class="theme-toggle-btn">Themes</button>
                    <div class="theme-menu">
                        <button class="theme-option" data-theme="light">Light</button>
                        <button class="theme-option" data-theme="dark">Dark</button>
                        <button class="theme-option" data-theme="ultra-dark">Ultra Dark</button>
                        <button class="theme-option" data-theme="retro">Retro</button>
                        <button class="theme-option" data-theme="compact">Compact</button>
                        <button class="theme-option" data-theme="powershell">Powershell</button>
                    </div>
                </div>
            </div>
        </nav>
        <div class="version-info" style="position: absolute; top: 10px; right: 10px; font-size: 0.7rem; color: var(--text-color-muted);">
            v%s
        </div>
        <div class="container" style="height: 70vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%%;">
                <h2 style="margin-bottom: 1rem;">Loading Metrics...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Aggregating cluster-wide metrics and preparing visualizations. This may take a moment.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0%% { transform: rotate(0deg); } 100%% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
</body>
</html>`, s.currentVersion)
}

func (s *Server) apiFullMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	// 1. Get Resource Usage
	resourceUsage := s.cachedData.ResourceUsage

	// 2. Get Metric Names for Explorer
	metricNames, err := s.store.GetDistinctMetricNames()
	if err != nil {
		s.logger.Warn("Failed to get distinct metric names", "error", err)
	}

	// 3. Process Prometheus Metrics from the store (Existing logic)
	var shardCPU []ShardCPUMetric
	var ioMetrics []IOMetric
	var uptime float64

	// Determine time range for metrics
	startTime := time.Time{} // Zero time means no start filter
	endTime := time.Time{}   // Zero time means no end filter
	limit := 0               // 0 means no limit (all results)
	offset := 0

	// Check for direct utilization metric
	useDirectUtilization := false
	shardCPUMap := make(map[string]float64)

	utilizationMetrics, err := s.store.GetMetrics("vectorized_reactor_utilization", nil, startTime, endTime, limit, offset)
	if err != nil {
		s.logger.Warn("Failed to get vectorized_reactor_utilization metrics", "error", err)
	} else if len(utilizationMetrics) > 0 {
		useDirectUtilization = true
		for _, m := range utilizationMetrics {
			shardID := m.Labels["shard"]
			if shardID != "" {
				val := m.Value
				// If value is <= 1.0, treat as ratio and convert to percentage
				if val <= 1.0 {
					val *= 100
				}
				shardCPUMap[shardID] = val
			} else {
				s.logger.Debug("vectorized_reactor_utilization metric missing shard label", "labels", m.Labels)
			}
		}
	}

	// If direct utilization is not available, calculate from busy seconds
	if !useDirectUtilization {
		busyMetrics, err := s.store.GetMetrics("vectorized_reactor_busy_seconds_total", nil, startTime, endTime, 0, 0)
		if err != nil {
			s.logger.Warn("Failed to get vectorized_reactor_busy_seconds_total metrics", "error", err)
		} else {
			uptimeMetrics, err := s.store.GetMetrics("redpanda_application_uptime_seconds_total", nil, startTime, endTime, 1, 0)
			if err != nil {
				s.logger.Warn("Failed to get uptime metrics", "error", err)
			} else if len(uptimeMetrics) > 0 {
				uptime = uptimeMetrics[0].Value
				if uptime > 0 {
					for _, m := range busyMetrics {
						shardID := m.Labels["shard"]
						if shardID != "" {
							usage := (m.Value / uptime) * 100
							shardCPUMap[shardID] = usage
						}
					}
				}
			}
		}
	}

	// Convert map to slice for template
	for shardID, usage := range shardCPUMap {
		shardCPU = append(shardCPU, ShardCPUMetric{
			ShardID: shardID,
			Usage:   usage,
		})
	}

	// Sort by shard ID
	sort.Slice(shardCPU, func(i, j int) bool {
		iNum, _ := strconv.Atoi(shardCPU[i].ShardID)
		jNum, _ := strconv.Atoi(shardCPU[j].ShardID)
		return iNum < jNum
	})

	// 4. Process IO Metrics
	ioMap := make(map[string]map[string]*IOMetric) // shard -> class -> metric

	readMetrics, err := s.store.GetMetrics("vectorized_io_queue_total_read_ops", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get read ops metrics", "error", err)
	}
	writeMetrics, err := s.store.GetMetrics("vectorized_io_queue_total_write_ops", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get write ops metrics", "error", err)
	}

	for _, m := range readMetrics {
		shardID := m.Labels["shard"]
		class := m.Labels["class"]
		if shardID != "" && class != "" {
			if ioMap[shardID] == nil {
				ioMap[shardID] = make(map[string]*IOMetric)
			}
			if ioMap[shardID][class] == nil {
				ioMap[shardID][class] = &IOMetric{
					ShardID: shardID,
					Class:   class,
				}
			}
			ioMap[shardID][class].ReadOps = m.Value
		}
	}

	for _, m := range writeMetrics {
		shardID := m.Labels["shard"]
		class := m.Labels["class"]
		if shardID != "" && class != "" {
			if ioMap[shardID] == nil {
				ioMap[shardID] = make(map[string]*IOMetric)
			}
			if ioMap[shardID][class] == nil {
				ioMap[shardID][class] = &IOMetric{
					ShardID: shardID,
					Class:   class,
				}
			}
			ioMap[shardID][class].WriteOps = m.Value
		}
	}

	// Convert to slice and find max values
	var maxReadOps, maxWriteOps float64
	for _, classes := range ioMap {
		for _, metric := range classes {
			if metric.ReadOps > maxReadOps {
				maxReadOps = metric.ReadOps
			}
			if metric.WriteOps > maxWriteOps {
				maxWriteOps = metric.WriteOps
			}
			ioMetrics = append(ioMetrics, *metric)
		}
	}

	// Sort by shard and class
	sort.Slice(ioMetrics, func(i, j int) bool {
		if ioMetrics[i].ShardID != ioMetrics[j].ShardID {
			iNum, _ := strconv.Atoi(ioMetrics[i].ShardID)
			jNum, _ := strconv.Atoi(ioMetrics[j].ShardID)
			return iNum < jNum
		}
		return ioMetrics[i].Class < ioMetrics[j].Class
	})

	// 5. Network & Kafka Metrics
	var networkMetrics NetworkMetrics

	// Kafka Latency (Sum / Count)
	kafkaSum, err := s.store.GetMetrics("redpanda_kafka_request_latency_seconds_sum", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get kafka latency sum", "error", err)
	}
	kafkaCount, err := s.store.GetMetrics("redpanda_kafka_request_latency_seconds_count", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get kafka latency count", "error", err)
	}

	type latPair struct {
		sum   float64
		count float64
	}
	latMap := make(map[string]*latPair)

	for _, m := range kafkaSum {
		reqType := m.Labels["request"]
		if reqType == "" {
			continue
		}
		if _, ok := latMap[reqType]; !ok {
			latMap[reqType] = &latPair{}
		}
		latMap[reqType].sum += m.Value
	}
	for _, m := range kafkaCount {
		reqType := m.Labels["request"]
		if reqType == "" {
			continue
		}
		if _, ok := latMap[reqType]; !ok {
			latMap[reqType] = &latPair{}
		}
		latMap[reqType].count += m.Value
	}

	for reqType, pair := range latMap {
		if pair.count > 0 {
			networkMetrics.KafkaLatencies = append(networkMetrics.KafkaLatencies, KafkaLatencyStats{
				RequestType: reqType,
				AvgLatency:  pair.sum / pair.count,
				Count:       pair.count,
			})
		}
	}
	// Sort by latency desc
	sort.Slice(networkMetrics.KafkaLatencies, func(i, j int) bool {
		return networkMetrics.KafkaLatencies[i].AvgLatency > networkMetrics.KafkaLatencies[j].AvgLatency
	})

	// Active Connections (Sum across shards)
	connMetrics, err := s.store.GetMetrics("redpanda_rpc_active_connections", nil, startTime, endTime, 0, 0)
	if err == nil {
		for _, m := range connMetrics {
			networkMetrics.ActiveConnections += m.Value
		}
	}

	// RPC Errors (Sum across shards)
	rpcErrMetrics, err := s.store.GetMetrics("redpanda_rpc_request_errors_total", nil, startTime, endTime, 0, 0)
	if err == nil {
		for _, m := range rpcErrMetrics {
			networkMetrics.RPCErrors += m.Value
		}
	}

	// 6. Performance Analysis (Integrated)
	// We gather key metrics for AnalyzePerformance
	perfMetrics := &models.MetricsBundle{
		Files: make(map[string][]models.PrometheusMetric),
	}
	var allMetricList []models.PrometheusMetric
	// Fetch common perf metrics
	perfNames := []string{
		"vectorized_reactor_utilization",
		"vectorized_network_bytes_received",
		"vectorized_network_bytes_sent",
		"vectorized_raft_leadership_changes",
		"redpanda_rpc_active_connections",
	}
	for _, name := range perfNames {
		ms, _ := s.store.GetMetrics(name, nil, time.Time{}, time.Time{}, 1000, 0)
		for _, m := range ms {
			allMetricList = append(allMetricList, *m)
		}
	}
	perfMetrics.Files["snapshot"] = allMetricList
	perfReport := analysis.AnalyzePerformance(perfMetrics, s.cachedData.SarData)

	baseData := s.newBasePageData("Metrics")

	// 7. Build page data
	pageData := MetricsPageData{
		BasePageData:      baseData,
		NodeHostname:      s.nodeHostname,
		ResourceUsage:     resourceUsage,
		ShardCPU:          shardCPU,
		IOMetrics:         ioMetrics,
		MaxReadOps:        maxReadOps,
		MaxWriteOps:       maxWriteOps,
		SarData:           s.cachedData.SarData,
		PerformanceReport: perfReport,
		CoreCount:         s.cachedData.System.CoreCount,
		Network:           networkMetrics,
		RaftRecovery:      baseData.RaftRecovery,
		MetricNames:       metricNames,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.metricsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing metrics template", "error", err)
		http.Error(w, "Failed to execute metrics template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write metrics response", "error", err)
	}
}

// API Handler for Metric List
func (s *Server) handleMetricsList(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	names, err := s.store.GetDistinctMetricNames()
	if err != nil {
		s.logger.Error("Failed to get metric names", "error", err)
		http.Error(w, "Failed to get metric names", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		s.logger.Error("Failed to encode metric names", "error", err)
	}
}

// API Handler for Metric Data (Chart.js format)
func (s *Server) handleMetricsData(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Missing metric name", http.StatusBadRequest)
		return
	}

	// Fetch all data for this metric (no time range for now, fetch all)
	metrics, err := s.store.GetMetrics(metricName, nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		s.logger.Error("Failed to fetch metric data", "metric", metricName, "error", err)
		http.Error(w, "Failed to fetch metric data", http.StatusInternalServerError)
		return
	}

	// Group by labels
	type SeriesPoint struct {
		X string  `json:"x"` // Timestamp ISO
		Y float64 `json:"y"`
	}
	type Series struct {
		Label string        `json:"label"`
		Data  []SeriesPoint `json:"data"`
	}

	seriesMap := make(map[string]*Series)

	for _, m := range metrics {
		// Construct label string (e.g. "shard=0, class=kafka")
		var keys []string
		for k := range m.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var labelParts []string
		for _, k := range keys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, m.Labels[k]))
		}
		label := strings.Join(labelParts, ", ")
		if label == "" {
			label = "value"
		}

		if _, ok := seriesMap[label]; !ok {
			seriesMap[label] = &Series{
				Label: label,
				Data:  []SeriesPoint{},
			}
		}

		seriesMap[label].Data = append(seriesMap[label].Data, SeriesPoint{
			X: m.Timestamp.Format(time.RFC3339),
			Y: m.Value,
		})
	}

	// Convert map to slice
	var result []Series
	for _, series := range seriesMap {
		// Sort data points by time
		sort.Slice(series.Data, func(i, j int) bool {
			return series.Data[i].X < series.Data[j].X
		})
		result = append(result, *series)
	}

	// Sort series by label
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		s.logger.Error("Failed to encode metric data", "error", err)
	}
}
