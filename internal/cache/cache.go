package cache

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
	"github.com/alextreichler/bundleViewer/internal/store"
)

type CachedData struct {
	GroupedFiles      map[string][]parser.ParsedFile
	DataDiskFiles     []string
	CacheDiskFiles    []string
	Partitions        []models.ClusterPartition
	PartitionDebug    []models.PartitionDebug
	Leaders           []models.PartitionLeader
	KafkaMetadata     models.KafkaMetadataResponse
	TopicConfigs      models.TopicConfigsResponse
	ConsumerGroups    map[string]models.ConsumerGroup
	HealthOverview    models.HealthOverview
	ResourceUsage     models.ResourceUsage
	DataDiskStats     map[int]models.DiskStats
	CacheDiskStats    map[int]models.DiskStats
	DuEntries         []models.DuEntry
	K8sStore          models.K8sStore
	RedpandaDataDir   string
	RedpandaConfig    map[string]interface{}
	PartitionBalancer models.PartitionBalancerStatus
	MetricDefinitions map[string]string
	System            models.SystemState
	SarData           models.SarData
	TopicDetails      map[string]models.TopicConfig
	Store             store.Store
	CoreDumps         []string
	LogsOnly          bool
	StorageAnalysis   *models.StorageAnalysis
	TopicThroughput   map[string]float64
	RedpandaVersion   string
	SelfTestResults   []models.SelfTestNodeResult
}

func New(bundlePath string, s store.Store, logsOnly bool, p *models.ProgressTracker) (*CachedData, error) {
	adminPath := bundlePath + "/admin"
	utilsPath := bundlePath + "/utils"
	procPath := bundlePath + "/proc"

	dataDiskPattern := filepath.Join(bundlePath, "admin", "disk_stat_data_*.json")
	dataDiskFiles, _ := filepath.Glob(dataDiskPattern)

	cacheDiskPattern := filepath.Join(bundlePath, "admin", "disk_stat_cache_*.json")
	cacheDiskFiles, _ := filepath.Glob(cacheDiskPattern)

	var coreDumps []string
	if matches, err := filepath.Glob(filepath.Join(bundlePath, "core.*")); err == nil {
		for _, m := range matches {
			coreDumps = append(coreDumps, filepath.Base(m))
		}
	}

	var wg sync.WaitGroup
	var adminFiles, utilsFiles, procFiles, rootFiles []parser.ParsedFile
	var clusterConfigData []models.ClusterConfigEntry
	var clusterConfigError error
	var partitions []models.ClusterPartition
	var partitionDebug []models.PartitionDebug
	var leaders []models.PartitionLeader
	var kafkaMetadataData models.KafkaMetadataResponse
	var topicConfigsData models.TopicConfigsResponse
	var consumerGroupsData map[string]models.ConsumerGroup
	var healthOverviewData models.HealthOverview
	var resourceUsageData models.ResourceUsage
	var dataDiskStatsData = make(map[int]models.DiskStats)
	var cacheDiskStatsData = make(map[int]models.DiskStats)
	var duEntriesData []models.DuEntry
	var storageAnalysisData *models.StorageAnalysis
	var k8sStoreData models.K8sStore
	var redpandaDataDir string
	var systemData models.SystemState
	var sarData models.SarData
	var metricDefinitionsData map[string]string
	var redpandaConfig map[string]interface{}
	var partitionBalancerData models.PartitionBalancerStatus
	var topicThroughput map[string]float64
	var selfTestResults []models.SelfTestNodeResult

	// Define tasks
	tasks := []struct {
		name string
		fn   func()
	}{
		{
			name: "SAR data",
			fn: func() {
				p.Update(0, "Parsing SAR data...")
				var err error
				sarData, err = parser.ParseSar(bundlePath, logsOnly)
				if err != nil {
					slog.Warn("Failed to parse SAR data", "error", err)
				}
				p.Update(1, "SAR data parsed")
			},
		},
		{
			name: "System state",
			fn: func() {
				p.Update(0, "Parsing system state...")
				df, _ := parser.ParseDF(bundlePath)
				free, _ := parser.ParseFree(bundlePath)
				top, procs, _ := parser.ParseTop(bundlePath)
				sysctl, _ := parser.ParseSysctl(bundlePath)
				ntp, _ := parser.ParseNTP(bundlePath)
				ips, _ := parser.ParseIP(bundlePath)
				conns, _ := parser.ParseSS(bundlePath)
				uname, _ := parser.ParseUnameInfo(bundlePath)
				memInfo, _ := parser.ParseMemInfo(bundlePath)
				vmStat, _ := parser.ParseVMStat(bundlePath)
				cpuInfo, _ := parser.ParseCPUInfo(bundlePath)
				mdStat, _ := parser.ParseMDStat(bundlePath)
				cmdLine, _ := parser.ParseCmdLine(bundlePath)
				dmi, _ := parser.ParseDMI(bundlePath)
				dig, _ := parser.ParseDig(bundlePath)
				syslog, _ := parser.ParseSyslog(bundlePath)
				vmStatTS, _ := parser.ParseVMStatTimeSeries(bundlePath)
				lspci, _ := parser.ParseLSPCI(bundlePath)
				thp, _ := parser.ParseTHP(bundlePath)
				interrupts, _ := parser.ParseInterrupts(bundlePath)
				softirqs, _ := parser.ParseSoftIRQs(bundlePath)

				connSummary := models.ConnectionSummary{
					Total:   len(conns),
					ByState: make(map[string]int),
					ByPort:  make(map[string]int),
				}
				for _, c := range conns {
					connSummary.ByState[c.State]++
					if idx := strings.LastIndex(c.LocalAddr, ":"); idx != -1 {
						port := c.LocalAddr[idx+1:]
						connSummary.ByPort[port]++
					}
				}
				for port, count := range connSummary.ByPort {
					connSummary.SortedByPort = append(connSummary.SortedByPort, models.PortCount{Port: port, Count: count})
				}
				sort.Slice(connSummary.SortedByPort, func(i, j int) bool {
					return connSummary.SortedByPort[i].Count > connSummary.SortedByPort[j].Count
				})

				procPattern := filepath.Join(bundlePath, "proc", "cpuinfo-processor-*")
				coreFiles, _ := filepath.Glob(procPattern)
				coreCount := len(coreFiles)

				systemData = models.SystemState{
					FileSystems: df, Memory: free, MemInfo: memInfo, Load: top,
					Processes: procs,
					Uname: uname, DMI: dmi, Dig: dig, Syslog: syslog,
					VMStatAnalysis: vmStatTS, LSPCI: lspci, CPU: cpuInfo, Sysctl: sysctl,
					NTP: ntp, Interfaces: ips, Connections: conns, ConnSummary: connSummary,
					VMStat: vmStat, MDStat: mdStat, CmdLine: cmdLine, CoreCount: coreCount,
					TransparentHugePages: thp,
					Interrupts: interrupts,
					SoftIRQs: softirqs,
				}
				p.Update(1, "System state parsed")
			},
		},
		{
			name: "Prometheus metrics",
			fn: func() {
				p.Update(0, "Parsing Prometheus metrics...")
				patterns := []string{
					filepath.Join(bundlePath, "metrics", "*", "*_metrics.txt"),
					filepath.Join(bundlePath, "metrics", "*_metrics.txt"),
				}
				var metricsFiles []string
				for _, ptrn := range patterns {
					matches, _ := filepath.Glob(ptrn)
					if len(matches) > 0 {
						metricsFiles = matches
						break
					}
				}
				metricDefinitionsData = make(map[string]string)
				topicThroughput = make(map[string]float64) // Initialize outer variable
				
				if len(metricsFiles) > 0 {
					if defs, err := parser.ParseMetricDefinitions(metricsFiles[0]); err == nil {
						for k, v := range defs {
							metricDefinitionsData[k] = v
						}
					}

					const maxMetricFileSize = 500 * 1024 * 1024 // 500MB limit

					for _, file := range metricsFiles {
						fi, err := os.Stat(file)
						if err == nil {
							if fi.Size() > maxMetricFileSize {
								continue
							}
						}

						metricTimestamp := time.Now()
						if err == nil {
							metricTimestamp = fi.ModTime()
						}
						
						// Create a temporary store capture to get metrics for throughput calculation
						// We can't easily hook into ParsePrometheusMetrics without changing its signature significantly
						// OR we can rely on the Store. 
						// BETTER APPROACH: Since we already parse them into the DB, let's just query the DB 
						// for this specific aggregation *after* this loop if possible?
						// Actually, `s.BulkInsertMetrics` puts them in DB. 
						// Let's query the DB for the aggregation in a separate step or just do it here if we want to parse twice (bad).
						
						// Alternative: The cache *has* the store. We can just run a query at the end of this block.
						if err := parser.ParsePrometheusMetrics(file, nil, s, metricTimestamp); err != nil {
							slog.Warn("Failed to parse prometheus metrics", "file", file, "error", err)
						}
					}

					// Now query the store for per-topic throughput
					// We want SUM(value) of vectorized_storage_log_written_bytes GROUP BY label->>topic
					// Note: This relies on the store having JSON support for labels
					if sqlite, ok := s.(*store.SQLiteStore); ok {
						rows, err := sqlite.Query("SELECT json_extract(labels, '$.topic') as topic, SUM(value) FROM metrics WHERE name='vectorized_storage_log_written_bytes' GROUP BY topic")
						if err == nil {
							defer rows.Close()
							for rows.Next() {
								var topic string
								var val float64
								if err := rows.Scan(&topic, &val); err == nil && topic != "" {
									topicThroughput[topic] = val
								}
							}
						}
					}
				}
				p.Update(1, "Prometheus metrics parsed")
			},
		},
		{
			name: "Self-test results",
			fn: func() {
				p.Update(0, "Parsing self-test results...")
				var err error
				selfTestResults, err = parser.ParseSelfTestResults(bundlePath)
				if err != nil {
					slog.Warn("Failed to parse self-test results", "error", err)
				}
				p.Update(1, "Self-test results parsed")
			},
		},
		{
			name: "Log ingestion",
			fn: func() {
				p.Update(0, "Ingesting logs into SQLite...")
				hasLogs, err := s.HasLogs()
				if err != nil {
					slog.Warn("Failed to check if logs exist in store", "error", err)
				}
				if !hasLogs {
					// Optimization: Drop indexes for faster bulk insert
					if sqlite, ok := s.(*store.SQLiteStore); ok {
						if err := sqlite.DropIndexes(); err != nil {
							slog.Warn("Failed to drop indexes for bulk insert", "error", err)
						}
					}

					if err := parser.ParseLogs(bundlePath, s, logsOnly, p); err != nil {
						slog.Warn("Failed to parse logs", "error", err)
					}

					// Restore indexes and rebuild FTS
					if sqlite, ok := s.(*store.SQLiteStore); ok {
						p.SetStatus("Restoring indexes (this may take a minute)...")
						if err := sqlite.RestoreIndexes(); err != nil {
							slog.Warn("Failed to restore indexes after bulk insert", "error", err)
						}
					}
				}
				p.Update(1, "Logs ingested")
			},
		},
		{
			name: "Data directory",
			fn: func() {
				redpandaDataDir, _ = parser.ParseRedpandaDataDirectory(bundlePath)
				p.Update(1, "Data directory identified")
			},
		},
		{
			name: "Configuration",
			fn: func() {
				redpandaConfig, _ = parser.ParseRedpandaConfig(bundlePath)
				p.Update(1, "Configuration parsed")
			},
		},
		{
			name: "Partition Balancer Status",
			fn: func() {
				partitionBalancerData, _ = parser.ParsePartitionBalancerStatus(bundlePath)
				p.Update(1, "Partition balancer status parsed")
			},
		},
		{
			name: "Kubernetes resources",
			fn: func() {
				p.Update(0, "Parsing Kubernetes resources...")
				var err error
				k8sStoreData, err = parser.ParseK8sResources(bundlePath, logsOnly, s)
				if err != nil {
					slog.Warn("Failed to parse K8s resources", "error", err)
				}
				p.Update(1, "Kubernetes resources parsed")
			},
		},
		{
			name: "Partitions",
			fn: func() {
				var err error
				partitions, err = parser.ParseClusterPartitions(bundlePath)
				if err != nil {
					slog.Warn("Failed to parse cluster partitions", "error", err)
				}
				p.Update(1, "Partition metadata parsed")
			},
		},
		{
			name: "Partition debug",
			fn: func() {
				var err error
				partitionDebug, err = parser.ParsePartitionDebug(bundlePath)
				if err != nil {
					slog.Warn("Failed to parse partition debug", "error", err)
				}
				p.Update(1, "Partition debug parsed")
			},
		},
		{
			name: "Partition leaders",
			fn: func() {
				var err error
				leaders, err = parser.ParsePartitionLeaders(bundlePath, logsOnly)
				if err != nil {
					slog.Warn("Failed to parse partition leaders", "error", err)
				}
				p.Update(1, "Leader metadata parsed")
			},
		},
		{
			name: "Kafka metadata",
			fn: func() {
				p.Update(0, "Parsing Kafka metadata, configs, and groups...")
				fullData, err := parser.ParseKafkaJSON(bundlePath)
				if err != nil {
					slog.Warn("Failed to parse Kafka JSON", "error", err)
				}
				kafkaMetadataData = fullData.Metadata
				topicConfigsData = fullData.TopicConfigs
				consumerGroupsData = fullData.ConsumerGroups
				p.Update(1, "Kafka data parsed")
			},
		},
		{
			name: "Health overview",
			fn: func() {
				var err error
				healthOverviewData, err = parser.ParseHealthOverview(bundlePath)
				if err != nil {
					slog.Warn("Failed to parse health overview", "error", err)
				}
				p.Update(1, "Health overview parsed")
			},
		},
		{
			name: "Resource usage",
			fn: func() {
				resourceUsageFile := filepath.Join(bundlePath, "resource-usage.json")
				if data, err := os.ReadFile(resourceUsageFile); err == nil {
					if err := json.Unmarshal(data, &resourceUsageData); err != nil {
						slog.Warn("Failed to unmarshal resource usage", "error", err)
					}
				}
				p.Update(1, "Resource usage parsed")
			},
		},
		{
			name: "CPU Profiles",
			fn: func() {
				p.Update(0, "Parsing CPU profiles...")
				profiles, err := parser.ParseCpuProfiles(bundlePath)
				if err == nil && len(profiles) > 0 {
					if err := s.BulkInsertCpuProfiles(profiles); err != nil {
						slog.Warn("Failed to bulk insert CPU profiles", "error", err)
					}
				} else if err != nil {
					slog.Warn("Failed to parse CPU profiles", "error", err)
				}
				p.Update(1, "CPU profiles parsed")
			},
		},
		{
			name: "Storage Analysis",
			fn: func() {
				var err error
				storageAnalysisData, err = analysis.AnalyzeStorage(bundlePath)
				if err != nil {
					slog.Warn("Failed to analyze storage", "error", err)
				}
				p.Update(1, "Storage analysis completed")
			},
		},
		{
			name: "Disk usage (DU)",
			fn: func() {
				var err error
				duEntriesData, err = parser.ParseDuOutput(bundlePath, logsOnly)
				if err != nil {
					slog.Warn("Failed to parse DU output", "error", err)
				}
				p.Update(1, "Disk usage (DU) parsed")
			},
		},
		{
			name: "Admin directory",
			fn: func() {
				allAdminFiles, _ := parser.ParseAllFiles(adminPath, []string{".json"})
				for _, file := range allAdminFiles {
					if file.FileName == "cluster_config.json" {
						clusterConfigData, clusterConfigError = parser.ParseClusterConfig(bundlePath)
					} else {
						adminFiles = append(adminFiles, file)
					}
				}
				p.Update(1, "Admin directory parsed")
			},
		},
		{
			name: "Utils directory",
			fn: func() {
				utilsFiles, _ = parser.ParseAllFiles(utilsPath, []string{".txt"})
				p.Update(1, "Utils directory parsed")
			},
		},
		{
			name: "Proc directory",
			fn: func() {
				allProcFiles, _ := parser.ParseAllFiles(procPath, []string{""})
				for _, file := range allProcFiles {
					if !strings.HasPrefix(file.FileName, "cpuinfo-processor-") {
						procFiles = append(procFiles, file)
					}
				}
				p.Update(1, "Proc directory parsed")
			},
		},
		{
			name: "Root files",
			fn: func() {
				rootFilesToParse := []string{"errors.txt", "kafka.json", "redpanda.yaml", "resource-usage.json"}
				rootFiles, _ = parser.ParseSpecificFiles(bundlePath, rootFilesToParse)
				p.Update(1, "Root files parsed")
			},
		},
	}

	// Dispatch tasks
	wg.Add(len(tasks))
	for _, t := range tasks {
		go func(taskFunc func()) {
			defer wg.Done()
			taskFunc()
		}(t.fn)
	}

	wg.Wait()

	if err := s.Optimize(p); err != nil {
		slog.Warn("Failed to optimize database", "error", err)
	}
	p.Update(1, "Database optimized")

	// Post-process: Map disk stats to nodes using brokers.json
	hostToNodeID := make(map[string]int)
	for _, file := range adminFiles {
		if file.FileName == "brokers.json" {
			if data, ok := file.Data.(string); ok {
				// Should be []interface{} but parser might return string if generic parse
				var brokers []interface{}
				if err := json.Unmarshal([]byte(data), &brokers); err == nil {
					for _, b := range brokers {
						if bMap, ok := b.(map[string]interface{}); ok {
							if id, ok := bMap["node_id"].(float64); ok {
								if addr, ok := bMap["internal_rpc_address"].(string); ok {
									hostToNodeID[addr] = int(id)
								}
							}
						}
					}
				}
			} else if data, ok := file.Data.([]interface{}); ok {
				// Already parsed as JSON
				for _, b := range data {
					if bMap, ok := b.(map[string]interface{}); ok {
						if id, ok := bMap["node_id"].(float64); ok {
							if addr, ok := bMap["internal_rpc_address"].(string); ok {
								hostToNodeID[addr] = int(id)
							}
						}
					}
				}
			}
		}
	}

	// Process disk stats from admin files
	for _, file := range adminFiles {
		var targetMap map[int]models.DiskStats
		if strings.HasPrefix(file.FileName, "disk_stat_data_") {
			targetMap = dataDiskStatsData
		} else if strings.HasPrefix(file.FileName, "disk_stat_cache_") {
			targetMap = cacheDiskStatsData
		} else {
			continue
		}

		// Parse stats
		var stats models.DiskStats
		var err error
		if sData, ok := file.Data.(string); ok {
			err = json.Unmarshal([]byte(sData), &stats)
		} else if mapData, ok := file.Data.(map[string]interface{}); ok {
			// Convert map to struct manually or re-marshal
			b, _ := json.Marshal(mapData)
			err = json.Unmarshal(b, &stats)
		}

		if err == nil {
			// Match filename to node
			// Filename: disk_stat_data_<hostname>-<port>.json or similar
			// Hostname might strictly match internal_rpc_address

			// Brute force check against known hosts
			for host, id := range hostToNodeID {
				if strings.Contains(file.FileName, host) {
					targetMap[id] = stats
					break
				}
			}

			// Fallback: Try to find NodeID in filename directly?
			// No, bundle filenames are messy. Relying on brokers.json is best.
		}
	}

	sort.Slice(adminFiles, func(i, j int) bool { return adminFiles[i].FileName < adminFiles[j].FileName })
	sort.Slice(utilsFiles, func(i, j int) bool { return utilsFiles[i].FileName < utilsFiles[j].FileName })
	sort.Slice(procFiles, func(i, j int) bool { return procFiles[i].FileName < procFiles[j].FileName })
	sort.Slice(rootFiles, func(i, j int) bool { return rootFiles[i].FileName < rootFiles[j].FileName })

	groupedFiles := map[string][]parser.ParsedFile{
		"Root":  rootFiles,
		"Admin": adminFiles,
		"Proc":  procFiles,
		"Utils": utilsFiles,
	}
	if clusterConfigError != nil {
		groupedFiles["Admin"] = append(groupedFiles["Admin"], parser.ParsedFile{FileName: "cluster_config.json", Error: clusterConfigError})
	} else if len(clusterConfigData) > 0 {
		groupedFiles["Admin"] = append(groupedFiles["Admin"], parser.ParsedFile{FileName: "cluster_config.json", Data: clusterConfigData})
	}

	topicDetails := make(map[string]models.TopicConfig)
	for _, tc := range topicConfigsData {
		topicDetails[tc.Name] = tc
	}

	// Detect Redpanda Version
	var redpandaVersion string
	for _, file := range adminFiles {
		if file.FileName == "brokers.json" {
			if brokers, ok := file.Data.([]interface{}); ok {
				for _, brokerData := range brokers {
					if broker, ok := brokerData.(map[string]interface{}); ok {
						if v, ok := broker["version"].(string); ok && v != "" {
							redpandaVersion = v
							break
						}
					}
				}
			}
			break
		}
	}

	if redpandaVersion == "" {
		// Try from logs - search only the very beginning of the logs (asc)
		filter := &models.LogFilter{
			Search: "\"Welcome to Redpanda v\"",
			Sort:   "asc",
			Limit:  1,
		}
		logs, _, err := s.GetLogs(filter)
		if err == nil && len(logs) > 0 {
			re := regexp.MustCompile(`Welcome to Redpanda v([0-9]+\.[0-9]+\.[0-9]+)`)
			matches := re.FindStringSubmatch(logs[0].Message)
			if len(matches) > 1 {
				redpandaVersion = matches[1]
			}
		}
	}

	return &CachedData{
		GroupedFiles: groupedFiles, DataDiskFiles: dataDiskFiles, CacheDiskFiles: cacheDiskFiles,
		Partitions: partitions, PartitionDebug: partitionDebug, Leaders: leaders, KafkaMetadata: kafkaMetadataData,
		TopicConfigs: topicConfigsData, ConsumerGroups: consumerGroupsData, HealthOverview: healthOverviewData, ResourceUsage: resourceUsageData,
		DataDiskStats: dataDiskStatsData, CacheDiskStats: cacheDiskStatsData, DuEntries: duEntriesData,
		K8sStore: k8sStoreData, RedpandaDataDir: redpandaDataDir, RedpandaConfig: redpandaConfig, PartitionBalancer: partitionBalancerData,
		MetricDefinitions: metricDefinitionsData, System: systemData, SarData: sarData,
		TopicDetails: topicDetails, Store: s, CoreDumps: coreDumps, LogsOnly: logsOnly,
		StorageAnalysis: storageAnalysisData, TopicThroughput: topicThroughput,
		RedpandaVersion: redpandaVersion,
		SelfTestResults: selfTestResults,
	}, nil
}
