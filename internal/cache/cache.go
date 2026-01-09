package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
	"github.com/alextreichler/bundleViewer/internal/store"
)

type CachedData struct {
	GroupedFiles      map[string][]parser.ParsedFile
	DataDiskFiles     []string
	CacheDiskFiles    []string
	Partitions        []models.ClusterPartition
	Leaders           []models.PartitionLeader
	KafkaMetadata     models.KafkaMetadataResponse
	TopicConfigs      models.TopicConfigsResponse
	ConsumerGroups    map[string]models.ConsumerGroup
	HealthOverview    models.HealthOverview
	ResourceUsage     models.ResourceUsage
	DataDiskStats     models.DiskStats
	CacheDiskStats    models.DiskStats
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
	var leaders []models.PartitionLeader
	var kafkaMetadataData models.KafkaMetadataResponse
	var topicConfigsData models.TopicConfigsResponse
	var consumerGroupsData map[string]models.ConsumerGroup
	var healthOverviewData models.HealthOverview
	var resourceUsageData models.ResourceUsage
	var dataDiskStatsData models.DiskStats
	var cacheDiskStatsData models.DiskStats
	var duEntriesData []models.DuEntry
	var k8sStoreData models.K8sStore
	var redpandaDataDir string
	var systemData models.SystemState
	var sarData models.SarData
	var metricDefinitionsData map[string]string
	var redpandaConfig map[string]interface{}
	var partitionBalancerData models.PartitionBalancerStatus

	// Define tasks
	tasks := []struct {
		name string
		fn   func()
	}{
		{
			name: "SAR data",
			fn: func() {
				p.Update(0, "Parsing SAR data...")
				sarData, _ = parser.ParseSar(bundlePath, logsOnly)
				p.Update(1, "SAR data parsed")
			},
		},
		{
			name: "System state",
			fn: func() {
				p.Update(0, "Parsing system state...")
				df, _ := parser.ParseDF(bundlePath)
				free, _ := parser.ParseFree(bundlePath)
				top, _ := parser.ParseTop(bundlePath)
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
					Uname: uname, DMI: dmi, Dig: dig, Syslog: syslog,
					VMStatAnalysis: vmStatTS, LSPCI: lspci, CPU: cpuInfo, Sysctl: sysctl,
					NTP: ntp, Interfaces: ips, Connections: conns, ConnSummary: connSummary,
					VMStat: vmStat, MDStat: mdStat, CmdLine: cmdLine, CoreCount: coreCount,
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
				if len(metricsFiles) > 0 {
					if defs, err := parser.ParseMetricDefinitions(metricsFiles[0]); err == nil {
						for k, v := range defs {
							metricDefinitionsData[k] = v
						}
					}
					for _, file := range metricsFiles {
						metricTimestamp := time.Now()
						if fi, err := os.Stat(file); err == nil {
							metricTimestamp = fi.ModTime()
						}
						_ = parser.ParsePrometheusMetrics(file, nil, s, metricTimestamp)
					}
				}
				p.Update(1, "Prometheus metrics parsed")
			},
		},
		{
			name: "Log ingestion",
			fn: func() {
				p.Update(0, "Ingesting logs into SQLite...")
				hasLogs, _ := s.HasLogs()
				if !hasLogs {
					// Optimization: Drop indexes for faster bulk insert
					if sqlite, ok := s.(*store.SQLiteStore); ok {
						_ = sqlite.DropIndexes()
					}

					_ = parser.ParseLogs(bundlePath, s, logsOnly, p)

					// Restore indexes and rebuild FTS
					if sqlite, ok := s.(*store.SQLiteStore); ok {
						p.SetStatus("Restoring indexes (this may take a minute)...")
						_ = sqlite.RestoreIndexes()
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
				k8sStoreData, _ = parser.ParseK8sResources(bundlePath, logsOnly, s)
				p.Update(1, "Kubernetes resources parsed")
			},
		},
		{
			name: "Partitions",
			fn: func() {
				partitions, _ = parser.ParseClusterPartitions(bundlePath)
				p.Update(1, "Partition metadata parsed")
			},
		},
		{
			name: "Partition leaders",
			fn: func() {
				leaders, _ = parser.ParsePartitionLeaders(bundlePath, logsOnly)
				p.Update(1, "Leader metadata parsed")
			},
		},
		{
			name: "Kafka metadata",
			fn: func() {
				p.Update(0, "Parsing Kafka metadata, configs, and groups...")
				fullData, _ := parser.ParseKafkaJSON(bundlePath)
				kafkaMetadataData = fullData.Metadata
				topicConfigsData = fullData.TopicConfigs
				consumerGroupsData = fullData.ConsumerGroups
				p.Update(1, "Kafka data parsed")
			},
		},
		{
			name: "Health overview",
			fn: func() {
				healthOverviewData, _ = parser.ParseHealthOverview(bundlePath)
				p.Update(1, "Health overview parsed")
			},
		},
		{
			name: "Resource usage",
			fn: func() {
				resourceUsageFile := filepath.Join(bundlePath, "resource-usage.json")
				if data, err := os.ReadFile(resourceUsageFile); err == nil {
					_ = json.Unmarshal(data, &resourceUsageData)
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
					_ = s.BulkInsertCpuProfiles(profiles)
				}
				p.Update(1, "CPU profiles parsed")
			},
		},
		{
			name: "Data disk stats",
			fn: func() {
				if len(dataDiskFiles) > 0 {
					if data, err := os.ReadFile(dataDiskFiles[0]); err == nil {
						_ = json.Unmarshal(data, &dataDiskStatsData)
					}
				}
				p.Update(1, "Data disk stats parsed")
			},
		},
		{
			name: "Cache disk stats",
			fn: func() {
				if len(cacheDiskFiles) > 0 {
					if data, err := os.ReadFile(cacheDiskFiles[0]); err == nil {
						_ = json.Unmarshal(data, &cacheDiskStatsData)
					}
				}
				p.Update(1, "Cache disk stats parsed")
			},
		},
		{
			name: "Disk usage (DU)",
			fn: func() {
				duEntriesData, _ = parser.ParseDuOutput(bundlePath, logsOnly)
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

	p.SetStatus("Optimizing database and building indexes...")
	_ = s.Optimize()
	p.Update(1, "Database optimized")

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

	return &CachedData{
		GroupedFiles: groupedFiles, DataDiskFiles: dataDiskFiles, CacheDiskFiles: cacheDiskFiles,
		Partitions: partitions, Leaders: leaders, KafkaMetadata: kafkaMetadataData,
		TopicConfigs: topicConfigsData, ConsumerGroups: consumerGroupsData, HealthOverview: healthOverviewData, ResourceUsage: resourceUsageData,
		DataDiskStats: dataDiskStatsData, CacheDiskStats: cacheDiskStatsData, DuEntries: duEntriesData,
		K8sStore: k8sStoreData, RedpandaDataDir: redpandaDataDir, RedpandaConfig: redpandaConfig, PartitionBalancer: partitionBalancerData,
		MetricDefinitions: metricDefinitionsData, System: systemData, SarData: sarData,
		TopicDetails: topicDetails, Store: s, CoreDumps: coreDumps, LogsOnly: logsOnly,
	}, nil
}
