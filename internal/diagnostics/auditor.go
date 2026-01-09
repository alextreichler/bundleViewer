package diagnostics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityWarning  Severity = "Warning"
	SeverityInfo     Severity = "Info"
	SeverityPass     Severity = "Pass"
)

type CheckResult struct {
	Category      string   `json:"category"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	CurrentValue  string   `json:"current_value"`
	ExpectedValue string   `json:"expected_value"`
	Status        Severity `json:"status"`
	Remediation   string   `json:"remediation"`
}

type DiagnosticsReport struct {
	Results []CheckResult
}

// Audit performs all diagnostic checks on the cached data
func Audit(data *cache.CachedData, s store.Store) DiagnosticsReport {
	var results []CheckResult

	// 1. OS Tuning Checks (sysctl)
	results = append(results, checkSysctl(data.System.Sysctl)...)

	// 2. Redpanda Configuration Checks
	results = append(results, checkRedpandaConfig(data.RedpandaConfig)...)

	// 3. Disk Checks
	results = append(results, checkDisks(data)...)

	// 4. Resource Checks
	results = append(results, checkResources(data)...)

	// 5. Network State Checks
	results = append(results, checkNetwork(data)...)

	// 6. Bottleneck Checks
	results = append(results, checkBottlenecks(s)...)

	// 7. Crash Dump Checks (Critical)
	results = append(results, checkCrashDumps(data)...)

	// 8. K8s Pod Lifecycle (OOM/Crash)
	results = append(results, checkK8sEvents(data)...)

	// 9. Hardware & System Integrity (RAID, Virt, OOMs)
	results = append(results, checkHardware(data)...)

	// 10. Log-based Stall Detection
	results = append(results, checkStalls(s)...)

	// 11. Cluster Balance & Partition Distribution
	results = append(results, checkClusterBalance(data)...)

	// 12. Controller Health
	results = append(results, checkControllerHealth(data, s)...)

	return DiagnosticsReport{Results: results}
}

func checkControllerHealth(data *cache.CachedData, s store.Store) []CheckResult {
	var results []CheckResult

	// 1. Controller Leadership Stability
	// Find redpanda/controller/0
	var controllerLeader int = -1
	var controllerTerm int = -1

	for _, l := range data.Leaders {
		if l.Ns == "redpanda" && l.Topic == "controller" && l.PartitionID == 0 {
			controllerLeader = l.Leader
			controllerTerm = l.LastStableLeaderTerm
			break
		}
	}

	if controllerLeader == -1 {
		results = append(results, CheckResult{
			Category:      "Controller Health",
			Name:          "Controller Leader",
			Description:   "Leadership of redpanda/controller/0",
			CurrentValue:  "None",
			ExpectedValue: "Valid Node ID",
			Status:        SeverityCritical,
			Remediation:   "Controller has no leader. Cluster management operations will fail. Check logs.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Controller Health",
			Name:          "Controller Leader",
			Description:   "Leadership of redpanda/controller/0",
			CurrentValue:  fmt.Sprintf("Node %d (Term %d)", controllerLeader, controllerTerm),
			ExpectedValue: "Valid Node ID",
			Status:        SeverityPass,
		})
	}

	// 2. Leader Imbalance
	// Count leaders per node
	leaderCounts := make(map[int]int)
	totalLeaders := 0
	for _, l := range data.Leaders {
		leaderCounts[l.Leader]++
		totalLeaders++
	}

	if totalLeaders > 0 {
		avgLeaders := float64(totalLeaders) / float64(len(data.HealthOverview.AllNodes))
		// Check for severe imbalance (> 50% deviation from average)
		for nodeID, count := range leaderCounts {
			// Skip negative leader IDs (no leader)
			if nodeID < 0 {
				continue
			}
			deviation := (float64(count) - avgLeaders) / avgLeaders
			if deviation > 0.5 {
				results = append(results, CheckResult{
					Category:      "Cluster Balance",
					Name:          fmt.Sprintf("Leader Imbalance (Node %d)", nodeID),
					Description:   "Node holds disproportionate number of partition leaderships",
					CurrentValue:  fmt.Sprintf("%d leaders (Avg: %.0f)", count, avgLeaders),
					ExpectedValue: "Balanced",
					Status:        SeverityWarning,
					Remediation:   "Leader Balancer may be disabled or failing. Check logs for 'leader balancer'.",
				})
			}
		}
	}

	// 3. Controller Metrics (Pending Operations)
	// We need to fetch metrics for this.
	// Since we don't have the full metric store easily accessible as a map here (it's in the store),
	// we would ideally query it. But Audit takes the store.
	// Let's do a quick check for `vectorized_cluster_controller_pending_partition_operations`
	
	// Use a recent window
	metrics, err := s.GetMetrics("vectorized_cluster_controller_pending_partition_operations", nil, time.Now().Add(-1*time.Hour), time.Now(), 1, 0)
	if err == nil && len(metrics) > 0 {
		if metrics[0].Value > 10 { // Arbitrary threshold for "backlog"
			results = append(results, CheckResult{
				Category:      "Controller Health",
				Name:          "Pending Partition Operations",
				Description:   "Controller has a backlog of partition operations",
				CurrentValue:  fmt.Sprintf("%.0f", metrics[0].Value),
				ExpectedValue: "0",
				Status:        SeverityWarning,
				Remediation:   "Controller is busy or stuck processing partition moves/creations.",
			})
		}
	}

	return results
}

func checkClusterBalance(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// 1. Check Partition Balancer Status
	if data.PartitionBalancer.Status == "stalled" {
		results = append(results, CheckResult{
			Category:      "Cluster Balance",
			Name:          "Partition Balancer Stalled",
			Description:   "Balancer cannot schedule required moves",
			CurrentValue:  "Stalled",
			ExpectedValue: "Ready / In Progress",
			Status:        SeverityCritical,
			Remediation:   "Check for insufficient disk space, high partition counts, or quorum issues.",
		})
	}

	// 2. Check for Empty Nodes
	// We need to count partitions per node.
	// We can use data.Partitions (ClusterPartition list)
	nodePartitionCounts := make(map[int]int)
	
	// Initialize counts for all known brokers to 0
	// We can get broker IDs from HealthOverview.AllNodes or ResourceUsage
	for _, id := range data.HealthOverview.AllNodes {
		nodePartitionCounts[id] = 0
	}

	for _, p := range data.Partitions {
		for _, r := range p.Replicas {
			nodePartitionCounts[r.NodeID]++
		}
	}

	for nodeID, count := range nodePartitionCounts {
		if count == 0 {
			results = append(results, CheckResult{
				Category:      "Cluster Balance",
				Name:          fmt.Sprintf("Empty Node (ID %d)", nodeID),
				Description:   "Node has 0 partitions",
				CurrentValue:  "0 partitions",
				ExpectedValue: "> 0",
				Status:        SeverityWarning,
				Remediation:   "Node is empty. Run `rpk cluster partitions balance` to redistribute load.",
			})
		}
	}

	// 3. Rack Awareness
	// Check if racks are configured in kafka.json (metadata) or we can infer from partition replicas
	// If any partition has all replicas in the same rack (and we have > 1 rack), that's a violation?
	// The text says "partitions with violated rack constraint present" is a trigger.
	// We don't have explicit "rack violation" metric exposed easily without calculating it.
	// But we can check if `rack` is set on brokers.
	
	racks := make(map[string]bool)
	brokersWithRack := 0
	for _, b := range data.KafkaMetadata.Brokers {
		if b.Rack != nil && *b.Rack != "" {
			racks[*b.Rack] = true
			brokersWithRack++
		}
	}

	if brokersWithRack > 0 && brokersWithRack < len(data.KafkaMetadata.Brokers) {
		results = append(results, CheckResult{
			Category:      "Configuration",
			Name:          "Inconsistent Rack Configuration",
			Description:   "Some brokers have rack IDs, others do not",
			CurrentValue:  fmt.Sprintf("%d/%d brokers have racks", brokersWithRack, len(data.KafkaMetadata.Brokers)),
			ExpectedValue: "All or None",
			Status:        SeverityWarning,
			Remediation:   "Ensure all brokers have `rack` configured in redpanda.yaml for rack awareness.",
		})
	}

	// 4. Disk Usage Threshold (80%)
	// This is already checked in checkDisks, but we can make it specific to balancing trigger
	for _, df := range data.System.FileSystems {
		if !strings.HasPrefix(df.Filesystem, "/dev/") && !strings.Contains(df.MountPoint, "redpanda") {
			continue
		}
		usePctStr := strings.TrimSuffix(df.UsePercent, "%")
		usePct, _ := strconv.Atoi(usePctStr)
		
		if usePct >= 80 {
			results = append(results, CheckResult{
				Category:      "Cluster Balance",
				Name:          fmt.Sprintf("Disk Balancing Trigger (%s)", df.MountPoint),
				Description:   "Disk usage > 80% triggers partition balancing",
				CurrentValue:  df.UsePercent,
				ExpectedValue: "< 80%",
				Status:        SeverityInfo,
				Remediation:   "Cluster will attempt to rebalance away from this node.",
			})
		}
	}

	return results
}

func checkStalls(s store.Store) []CheckResult {
	var results []CheckResult
	
	// Search for "Reactor stalled" pattern
	entries, _, err := s.GetLogs(&models.LogFilter{
		Search: "Reactor stalled",
		Limit:  1,
	})
	if err == nil && len(entries) > 0 {
		results = append(results, CheckResult{
			Category:      "Reactor Health",
			Name:          "Reactor Stalls in Logs",
			Description:   "Log lines indicating the reactor loop was blocked",
			CurrentValue:  "Found",
			ExpectedValue: "None",
			Status:        SeverityWarning,
			Remediation:   "Check for long-running tasks or high contention.",
		})
	}

	return results
}

func checkHardware(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// 1. Check for OOM Events in Syslog
	if len(data.System.Syslog.OOMEvents) > 0 {
		results = append(results, CheckResult{
			Category:      "System Stability",
			Name:          "OOM Killer Invoked",
			Description:   "The kernel terminated processes to reclaim memory.",
			CurrentValue:  fmt.Sprintf("%d Events", len(data.System.Syslog.OOMEvents)),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "System is under-provisioned for memory. Investigate 'Extended Mem' in System page or reduce workload.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "System Stability",
			Name:          "OOM Killer Invoked",
			Description:   "The kernel terminated processes to reclaim memory.",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	// 2. Check RAID Status
	for _, array := range data.System.MDStat.Arrays {
		if strings.Contains(array.Status, "_") { // e.g. [_U]
			results = append(results, CheckResult{
				Category:      "Hardware Integrity",
				Name:          fmt.Sprintf("RAID %s Status", array.Name),
				Description:   "Software RAID Array Status",
				CurrentValue:  array.Status,
				ExpectedValue: "[UU...]",
				Status:        SeverityCritical,
				Remediation:   fmt.Sprintf("Array %s is degraded. Replace failed drive immediately.", array.Name),
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Hardware Integrity",
				Name:          fmt.Sprintf("RAID %s Status", array.Name),
				Description:   "Software RAID Array Status",
				CurrentValue:  "Healthy",
				ExpectedValue: "Healthy",
				Status:        SeverityPass,
			})
		}
	}

	// 3. Virtualization Warning
	isVirt := false
	if strings.Contains(strings.ToLower(data.System.DMI.Manufacturer), "vmware") || 
	   strings.Contains(strings.ToLower(data.System.DMI.Product), "vmware") {
		isVirt = true
	}
	
	if isVirt {
		results = append(results, CheckResult{
			Category:      "Platform",
			Name:          "Virtualization Detected",
			Description:   "Running on VMware/Virtual Platform",
			CurrentValue:  data.System.DMI.Manufacturer,
			ExpectedValue: "Bare Metal",
			Status:        SeverityWarning,
			Remediation:   "Ensure disk latency is low (<10ms). Avoid 'EagerZeroedThick' if possible? Actually, prefer EagerZeroedThick. Check IO Wait.",
		})
	}

	return results
}

func checkK8sEvents(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check Pod Statuses for OOMKilled or Errors
	for _, pod := range data.K8sStore.Pods {
		for _, status := range pod.Status.ContainerStatuses {
			if terminated, ok := status.LastState["terminated"]; ok {
				switch terminated.Reason {
				case "OOMKilled":
					results = append(results, CheckResult{
						Category:      "Kubernetes Events",
						Name:          fmt.Sprintf("Pod %s OOMKilled", pod.Metadata.Name),
						Description:   "Pod terminated due to Out Of Memory",
						CurrentValue:  fmt.Sprintf("Exit Code: %d, Finished At: %s", terminated.ExitCode, terminated.FinishedAt),
						ExpectedValue: "Running",
						Status:        SeverityCritical,
						Remediation:   "Increase container memory limit or investigate memory leak.",
					})
				case "Error":
					results = append(results, CheckResult{
						Category:      "Kubernetes Events",
						Name:          fmt.Sprintf("Pod %s Crash", pod.Metadata.Name),
						Description:   "Pod terminated with Error",
						CurrentValue:  fmt.Sprintf("Exit Code: %d, Reason: %s", terminated.ExitCode, terminated.Reason),
						ExpectedValue: "Running",
						Status:        SeverityCritical,
						Remediation:   "Check logs around the time of termination.",
					})
				}
			}
		}
	}

	return results
}

func checkCrashDumps(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	if len(data.CoreDumps) > 0 {
		results = append(results, CheckResult{
			Category:      "Crash Evidence",
			Name:          "Core Dumps Found",
			Description:   "Presence of core dump files indicating a crash",
			CurrentValue:  fmt.Sprintf("%d files found: %s", len(data.CoreDumps), strings.Join(data.CoreDumps, ", ")),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "Inspect core dumps with GDB or provide to support. A crash has occurred.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Crash Evidence",
			Name:          "Core Dumps Found",
			Description:   "Presence of core dump files indicating a crash",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	return results
}

func checkNetwork(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for high TIME_WAIT connections
	// This "proves" the need for tcp_tw_reuse
	timeWaitCount := data.System.ConnSummary.ByState["TIME-WAIT"] // Key from ss output is usually TIME-WAIT
	
	// ss output state format: ESTAB, TIME-WAIT, etc.
	// We need to match the exact string from parser. Let's assume standard ss output.
	// If the map key is empty/different, this check does nothing safe.
	
	// Also check common variations just in case
	if val, ok := data.System.ConnSummary.ByState["TIME_WAIT"]; ok {
		timeWaitCount += val
	}

	if timeWaitCount > 10000 {
		results = append(results, CheckResult{
			Category:      "Network Evidence",
			Name:          "TIME_WAIT Sockets",
			Description:   "Count of sockets in TIME_WAIT state",
			CurrentValue:  fmt.Sprintf("%d", timeWaitCount),
			ExpectedValue: "< 10000",
			Status:        SeverityWarning,
			Remediation:   "High TIME_WAIT count confirms the need for net.ipv4.tcp_tw_reuse = 1",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Network Evidence",
			Name:          "TIME_WAIT Sockets",
			Description:   "Count of sockets in TIME_WAIT state",
			CurrentValue:  fmt.Sprintf("%d", timeWaitCount),
			ExpectedValue: "< 10000",
			Status:        SeverityPass,
		})
	}

	return results
}

func checkSysctl(sysctl map[string]string) []CheckResult {
	var results []CheckResult

	// Helper to check integer values
	checkInt := func(key string, minVal int, severity Severity, desc, remediation string) {
		valStr, exists := sysctl[key]
		if !exists {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  "Not Found",
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        SeverityWarning,
				Remediation:   remediation,
			})
			return
		}

		val, err := strconv.Atoi(valStr)
		if err != nil {
			return // Skip if not int
		}

		if val < minVal {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  valStr,
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        severity,
				Remediation:   remediation,
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  valStr,
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        SeverityPass,
			})
		}
	}

	checkInt("fs.aio-max-nr", 1048576, SeverityCritical, "Maximum number of concurrent async I/O requests", "Increase fs.aio-max-nr in /etc/sysctl.conf")
	checkInt("net.core.somaxconn", 4096, SeverityWarning, "Max socket listen backlog", "Increase net.core.somaxconn to handle burst connections")
	checkInt("net.core.netdev_max_backlog", 2500, SeverityWarning, "Max packets queued on input interface", "Increase net.core.netdev_max_backlog")
	checkInt("net.ipv4.tcp_max_syn_backlog", 4096, SeverityWarning, "Max TCP SYN backlog", "Increase net.ipv4.tcp_max_syn_backlog to prevent dropped connections during bursts")

	// TCP Buffer Sizes (Read/Write)
	// format: "min default max"
	// We check the max value (3rd field)
	checkBuffer := func(key string, minMaxVal int, desc string) {
		valStr, exists := sysctl[key]
		if !exists {
			return
		}
		fields := strings.Fields(valStr)
		if len(fields) == 3 {
			maxVal, _ := strconv.Atoi(fields[2])
			if maxVal < minMaxVal {
				results = append(results, CheckResult{
					Category:      "OS Tuning",
					Name:          key,
					Description:   desc,
					CurrentValue:  valStr,
					ExpectedValue: fmt.Sprintf("Max >= %d", minMaxVal),
					Status:        SeverityInfo,
					Remediation:   fmt.Sprintf("Increase max %s to %d for high throughput (Tiered Storage/Recovery)", key, minMaxVal),
				})
			}
		}
	}

	// 16MB minimum for max TCP buffer
	checkBuffer("net.ipv4.tcp_rmem", 16777216, "TCP Read Memory (min default max)")
	checkBuffer("net.ipv4.tcp_wmem", 16777216, "TCP Write Memory (min default max)")

	// UDP/Core Buffers (for Gossip)
	checkInt("net.core.rmem_max", 2097152, SeverityInfo, "Max OS receive buffer size (affects UDP/Gossip)", "Increase net.core.rmem_max to 2MB+ to prevent gossip packet loss")
	checkInt("net.core.wmem_max", 2097152, SeverityInfo, "Max OS send buffer size (affects UDP/Gossip)", "Increase net.core.wmem_max to 2MB+")

	// TCP Timestamps should be enabled for safe TCP TIME-WAIT reuse
	checkInt("net.ipv4.tcp_timestamps", 1, SeverityInfo, "TCP Timestamps (required for tcp_tw_reuse safety)", "Ensure net.ipv4.tcp_timestamps is 1. If disabled, tcp_tw_reuse can cause data corruption under NAT.")

	// TCP Window Scaling (Critical for throughput)
	checkInt("net.ipv4.tcp_window_scaling", 1, SeverityCritical, "TCP Window Scaling", "Enable tcp_window_scaling to allow TCP window sizes > 64KB")

	// TCP SACK (Critical for loss recovery)
	checkInt("net.ipv4.tcp_sack", 1, SeverityWarning, "TCP Selective Acknowledgments", "Enable tcp_sack to improve throughput in lossy networks")

	// Ephemeral Port Range
	if valStr, exists := sysctl["net.ipv4.ip_local_port_range"]; exists {
		fields := strings.Fields(valStr)
		if len(fields) == 2 {
			minPort, _ := strconv.Atoi(fields[0])
			maxPort, _ := strconv.Atoi(fields[1])
			count := maxPort - minPort
			if count < 28000 {
				results = append(results, CheckResult{
					Category:      "OS Tuning",
					Name:          "net.ipv4.ip_local_port_range",
					Description:   "Range of ephemeral ports for outgoing connections",
					CurrentValue:  valStr,
					ExpectedValue: "> 28000 ports",
					Status:        SeverityWarning,
					Remediation:   "Widen ip_local_port_range (e.g., '1024 65535') to prevent port exhaustion",
				})
			}
		}
	}

	// TCP Slow Start after idle should be 0
	if valStr, exists := sysctl["net.ipv4.tcp_slow_start_after_idle"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val != 0 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_slow_start_after_idle",
				Description:   "TCP slow start after idle",
				CurrentValue:  valStr,
				ExpectedValue: "0",
				Status:        SeverityWarning,
				Remediation:   "Set net.ipv4.tcp_slow_start_after_idle to 0 to improve latency for bursty traffic",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_slow_start_after_idle",
				Description:   "TCP slow start after idle",
				CurrentValue:  valStr,
				ExpectedValue: "0",
				Status:        SeverityPass,
			})
		}
	}

	// TCP TW Reuse should be 1 (enabled)
	if valStr, exists := sysctl["net.ipv4.tcp_tw_reuse"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val != 1 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_tw_reuse",
				Description:   "Allow reuse of TIME-WAIT sockets",
				CurrentValue:  valStr,
				ExpectedValue: "1",
				Status:        SeverityInfo,
				Remediation:   "Enable net.ipv4.tcp_tw_reuse to efficiently reuse connections",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_tw_reuse",
				Description:   "Allow reuse of TIME-WAIT sockets",
				CurrentValue:  valStr,
				ExpectedValue: "1",
				Status:        SeverityPass,
			})
		}
	}
	
	// Swappiness should be 0 or 1
	if valStr, exists := sysctl["vm.swappiness"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val > 1 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "vm.swappiness",
				Description:   "Tendency to swap memory to disk",
				CurrentValue:  valStr,
				ExpectedValue: "<= 1",
				Status:        SeverityWarning,
				Remediation:   "Set vm.swappiness to 0 or 1 to prevent latency spikes",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "vm.swappiness",
				Description:   "Tendency to swap memory to disk",
				CurrentValue:  valStr,
				ExpectedValue: "<= 1",
				Status:        SeverityPass,
			})
		}
	}

	return results
}

func checkBottlenecks(s store.Store) []CheckResult {
	var results []CheckResult

	// Metrics required for performance analysis
	metricNames := []string{
		"vectorized_reactor_utilization",
		"vectorized_kafka_handler_latency_microseconds_count",
		"vectorized_io_queue_queue_length",
		"vectorized_storage_log_written_bytes",
		"vectorized_storage_log_batches_written",
		"vectorized_scheduler_runtime_ms",
		"vectorized_kafka_handler_sent_bytes_total",
		"vectorized_kafka_handler_received_bytes_total",
		"vectorized_io_queue_total_delay_sec",
		"vectorized_io_queue_total_operations",
		"vectorized_io_queue_total_exec_sec",
		"vectorized_batch_cache_hits",
		"vectorized_batch_cache_accesses",
		"vectorized_stall_detector_reported",
		"vectorized_scheduler_time_spent_on_task_quota_violations_ms",
		"vectorized_network_bytes_received",
		"vectorized_network_bytes_sent",
		"vectorized_host_snmp_tcp_established",
		"vectorized_raft_leadership_changes",
		"vectorized_raft_leader_for",
		"vectorized_raft_recovery_partitions_to_recover",
		"vectorized_raft_replicate_ack_all_requests",
		"vectorized_raft_replicate_ack_leader_requests",
	}

	// Use a wide time range to get all metrics
	now := time.Now()
	startTime := now.Add(-30 * 24 * time.Hour)
	endTime := now.Add(24 * time.Hour)

	metricsBundle := &models.MetricsBundle{
		Files: make(map[string][]models.PrometheusMetric),
	}
	
	// Create a "virtual" file to hold all fetched metrics
	var allMetrics []models.PrometheusMetric

	for _, name := range metricNames {
		metrics, err := s.GetMetrics(name, nil, startTime, endTime, 10000, 0)
		if err == nil {
			for _, m := range metrics {
				if m != nil {
					allMetrics = append(allMetrics, *m)
				}
			}
		}
	}
	metricsBundle.Files["store_metrics"] = allMetrics

	// Run Analysis
	report := analysis.AnalyzePerformance(metricsBundle)

	// Convert Report to CheckResults

	// CPU
	if report.IsCPUBound {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "CPU Saturation",
			Description:   "System is CPU bound (High Reactor Utilization)",
			CurrentValue:  "Critical",
			ExpectedValue: "Healthy",
			Status:        SeverityCritical,
			Remediation:   "Scale up CPU resources or optimize workload.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "CPU Saturation",
			Description:   "System is CPU bound (High Reactor Utilization)",
			CurrentValue:  "Healthy",
			ExpectedValue: "Healthy",
			Status:        SeverityPass,
		})
	}

	// Disk
	if report.IsDiskBound {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "Disk Saturation",
			Description:   "System is Disk bound (High IO Queue Length)",
			CurrentValue:  "Critical",
			ExpectedValue: "Healthy",
			Status:        SeverityWarning,
			Remediation:   "Check disk IOPS/throughput. Workload may require faster disks.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "Disk Saturation",
			Description:   "System is Disk bound (High IO Queue Length)",
			CurrentValue:  "Healthy",
			ExpectedValue: "Healthy",
			Status:        SeverityPass,
		})
	}

	// Workload Insights
	results = append(results, CheckResult{
		Category:      "Workload Characterization",
		Name:          "Workload Type",
		Description:   "Dominant request type (Produce vs Fetch)",
		CurrentValue:  report.WorkloadType,
		ExpectedValue: "N/A",
		Status:        SeverityInfo,
		Remediation:   strings.Join(report.Recommendations, " "),
	})

	// Other Observations
	for _, obs := range report.Observations {
		// Heuristic to detect "Small batch" observation for warning
		severity := SeverityInfo
		if strings.Contains(obs, "Small average batch size") {
			severity = SeverityWarning
		} else if strings.Contains(obs, "Critical:") {
			severity = SeverityCritical
		} else if strings.Contains(obs, "Warning:") {
			severity = SeverityWarning
		}

		results = append(results, CheckResult{
			Category:      "Performance Insights",
			Name:          "Observation",
			Description:   "Derived insight from metrics",
			CurrentValue:  obs,
			ExpectedValue: "N/A",
			Status:        severity,
		})
	}

	return results
}


func checkRedpandaConfig(config map[string]interface{}) []CheckResult {
	var results []CheckResult

	// Helper to check boolean string values (parser returns strings for YAML values)
	isTrue := func(key string) bool {
		if val, ok := config[key]; ok {
			if s, ok := val.(string); ok {
				return strings.ToLower(s) == "true"
			}
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return false
	}

	// Helper to check if a key exists and is explicitly false
	isFalse := func(key string) bool {
		if val, ok := config[key]; ok {
			if s, ok := val.(string); ok {
				return strings.ToLower(s) == "false"
			}
			if b, ok := val.(bool); ok {
				return !b
			}
		}
		return false
	}

	// Check developer_mode
	if isTrue("redpanda.developer_mode") {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "developer_mode",
			Description:   "Developer mode bypasses many checks",
			CurrentValue:  "true",
			ExpectedValue: "false",
			Status:        SeverityWarning,
			Remediation:   "Disable developer_mode for production clusters",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "developer_mode",
			Description:   "Developer mode",
			CurrentValue:  "false",
			ExpectedValue: "false",
			Status:        SeverityPass,
		})
	}

	// Check rpk.overprovisioned
	if isTrue("rpk.overprovisioned") {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "rpk.overprovisioned",
			Description:   "Disables thread pinning (Overprovisioned mode)",
			CurrentValue:  "true",
			ExpectedValue: "false",
			Status:        SeverityInfo,
			Remediation:   "Overprovisioned mode disables thread pinning. It is functionally well tested but performance-wise not well tested. Disable for production if possible, unless working around specific kernel bugs.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "rpk.overprovisioned",
			Description:   "Disables thread pinning (Overprovisioned mode)",
			CurrentValue:  "false",
			ExpectedValue: "false",
			Status:        SeverityPass,
		})
	}

	// Check rpk.tune_aio_events
	// We unconditionally recommend running the tuner. If this is explicitly false, it's a warning.
	if isFalse("rpk.tune_aio_events") {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "rpk.tune_aio_events",
			Description:   "Tuner AIO Events configuration",
			CurrentValue:  "false",
			ExpectedValue: "true",
			Status:        SeverityWarning,
			Remediation:   "Enable rpk tuning or run `rpk redpanda tune all` to optimize I/O settings.",
		})
	} else if isTrue("rpk.tune_aio_events") {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "rpk.tune_aio_events",
			Description:   "Tuner AIO Events configuration",
			CurrentValue:  "true",
			ExpectedValue: "true",
			Status:        SeverityPass,
		})
	}

	return results
}

func checkDisks(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for high disk usage and Filesystem Type
	for _, df := range data.System.FileSystems {
		// Filter for likely data partitions (mounted on /var/lib/redpanda or similar, or just large ones)
		// We'll check all that look like physical disks (starting with /dev/) or likely data mounts
		if !strings.HasPrefix(df.Filesystem, "/dev/") && !strings.Contains(df.MountPoint, "redpanda") {
			continue
		}

		// Check Filesystem Type (XFS preferred)
		if df.Type != "xfs" {
			results = append(results, CheckResult{
				Category:      "Disk Configuration",
				Name:          df.MountPoint + " Filesystem",
				Description:   "Filesystem type for data directory",
				CurrentValue:  df.Type,
				ExpectedValue: "xfs",
				Status:        SeverityWarning,
				Remediation:   "Redpanda is optimized for XFS. Consider reformatting with XFS for better performance.",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Disk Configuration",
				Name:          df.MountPoint + " Filesystem",
				Description:   "Filesystem type for data directory",
				CurrentValue:  df.Type,
				ExpectedValue: "xfs",
				Status:        SeverityPass,
			})
		}
		
		usePctStr := strings.TrimSuffix(df.UsePercent, "%")
		usePct, _ := strconv.Atoi(usePctStr)

		if usePct > 85 {
			severity := SeverityWarning
			if usePct > 95 {
				severity = SeverityCritical
			}
			results = append(results, CheckResult{
				Category:      "Disk Usage",
				Name:          df.MountPoint,
				Description:   "Disk space usage",
				CurrentValue:  df.UsePercent,
				ExpectedValue: "< 85%",
				Status:        severity,
				Remediation:   "Free up space or expand disk capacity",
			})
		}
	}

	return results
}

func checkResources(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for under-replicated partitions
	urp := data.HealthOverview.UnderReplicatedCount
	if urp > 0 {
		results = append(results, CheckResult{
			Category:      "Cluster Health",
			Name:          "Under Replicated Partitions",
			Description:   "Partitions not fully replicated",
			CurrentValue:  fmt.Sprintf("%d", urp),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "Check for offline nodes or network partitions. Investigate logs for replication errors.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Cluster Health",
			Name:          "Under Replicated Partitions",
			Description:   "Partitions not fully replicated",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	return results
}
