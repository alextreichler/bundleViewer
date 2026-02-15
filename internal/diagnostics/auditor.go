package diagnostics

import (
	"fmt"
	"regexp"
	"sort"
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
	results = append(results, checkBottlenecks(data, s)...)

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

	// 13. Log Error Patterns
	results = append(results, checkLogErrors(s)...)

	// 14. Transparent Hugepages
	results = append(results, checkTransparentHugePages(data)...)

	// 15. K8s QoS
	results = append(results, checkK8sQoS(data)...)

	// 16. IRQ Balance
	results = append(results, checkIRQBalance(data)...)

	// 17. Data Partition on Root
	results = append(results, checkDataDirPartition(data)...)

	// 18. Public IP / External Haircut Detection
	results = append(results, checkPublicIPs(data)...)

	// 19. Consumer Parallelism Analysis
	results = append(results, checkConsumerParallelism(data)...)

	// 20. Shard 0 Administrative Overload
	results = append(results, checkShard0Overload(s)...)

	// 21. Connection Flapping (Client Churn)
	results = append(results, checkConnectionFlapping(s)...)

	// 22. Metadata Request Overload
	results = append(results, checkMetadataOverload(s)...)

	// 23. K8s CPU Steal Time (Overprovisioning)
	results = append(results, checkCPUStealTime(data, s)...)

	// 24. K8s Shard Optimization (N-1 Tax)
	results = append(results, checkK8sShardOptimization(data)...)

	// 25. RHEL-8 Sluggish Scheduler (sched_wakeup_granularity_ns)
	results = append(results, checkSluggishScheduler(data)...)

	// 26. IO Configuration (iotune)
	results = append(results, checkIOConfiguration(data)...)

	// 27. AIO Context Saturation
	results = append(results, checkAIOSaturation(s)...)

	// 28. SoftIRQ Imbalance (Network/IO steering)
	results = append(results, checkSoftIRQImbalance(data)...)

	// 29. AWS GP3 Performance Cliff
	results = append(results, checkGP3PerformanceCliff(data, s)...)

	// 30. Self-Test Performance Thresholds
	results = append(results, checkSelfTestPerformance(data)...)

	// 31. Network Concurrency (AIO Control Blocks)
	results = append(results, checkNetworkConcurrencyLimit(data, s)...)

	// 32. Socket-level Health (ss)
	results = append(results, checkSocketHealth(data)...)

	// 33. Advanced Metrics-based Integrity and Pressure Checks
	// Get some samples to check for errors/reclaims
	integrityBundle := &models.MetricsBundle{Files: map[string][]models.PrometheusMetric{"integrity": {}}}
	if metrics, err := s.GetMetrics("vectorized_storage_log_batch_parse_errors", nil, time.Time{}, time.Time{}, 0, 0); err == nil {
		for _, m := range metrics { integrityBundle.Files["integrity"] = append(integrityBundle.Files["integrity"], *m) }
	}
	if metrics, err := s.GetMetrics("vectorized_storage_log_batch_write_errors", nil, time.Time{}, time.Time{}, 0, 0); err == nil {
		for _, m := range metrics { integrityBundle.Files["integrity"] = append(integrityBundle.Files["integrity"], *m) }
	}
	if metrics, err := s.GetMetrics("vectorized_storage_log_corrupted_compaction_indices", nil, time.Time{}, time.Time{}, 0, 0); err == nil {
		for _, m := range metrics { integrityBundle.Files["integrity"] = append(integrityBundle.Files["integrity"], *m) }
	}
	results = append(results, checkIntegrityMetrics(integrityBundle)...)

	pressureBundle := &models.MetricsBundle{Files: map[string][]models.PrometheusMetric{"pressure": {}}}
	if metrics, err := s.GetMetrics("vectorized_memory_reclaims_operations", nil, time.Time{}, time.Time{}, 0, 0); err == nil {
		for _, m := range metrics { pressureBundle.Files["pressure"] = append(pressureBundle.Files["pressure"], *m) }
	}
	results = append(results, checkMemoryReclaimPressure(pressureBundle)...)

	// 34. RPC Telemetry and Stability Checks
	results = append(results, checkRPCTimeouts(s)...)

	return DiagnosticsReport{Results: results}
}

func checkNetworkConcurrencyLimit(data *cache.CachedData, s store.Store) []CheckResult {
	var results []CheckResult

	// 1. Determine the limit (default is 10000 if not specified)
	limit := 10000.0
	
	// We need to look in two places:
	// A) The literal process command line
	// B) The redpanda.yaml "additional_start_flags" list
	
	searchTargets := []string{data.System.CmdLine}
	for _, v := range data.RedpandaConfig {
		if s, ok := v.(string); ok {
			searchTargets = append(searchTargets, s)
		}
	}

	re := regexp.MustCompile(`--max-networking-io-control-blocks=(\d+)`)
	for _, target := range searchTargets {
		matches := re.FindStringSubmatch(target)
		if len(matches) > 1 {
			val, _ := strconv.ParseFloat(matches[1], 64)
			if val > 0 {
				limit = val
				break // Found it
			}
		}
	}

	// 2. Get active connections per shard
	metrics, err := s.GetMetrics("redpanda_rpc_active_connections", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil || len(metrics) == 0 {
		return results
	}

	maxConnsPerShard := 0.0
	worstShard := ""
	for _, m := range metrics {
		if m.Value > maxConnsPerShard {
			maxConnsPerShard = m.Value
			// Try various common shard labels
			if s, ok := m.Labels["shard"]; ok {
				worstShard = s
			} else if r, ok := m.Labels["reactor"]; ok {
				worstShard = r
			} else if c, ok := m.Labels["cpu"]; ok {
				worstShard = c
			}
		}
	}

	if maxConnsPerShard > 0 {
		utilization := (maxConnsPerShard / limit) * 100
		
		if utilization > 80.0 {
			status := SeverityWarning
			if utilization > 95.0 {
				status = SeverityCritical
			}

			results = append(results, CheckResult{
				Category:      "Network Performance",
				Name:          "Network AIO Concurrency",
				Description:   "Active connections are approaching the AIO control block limit",
				CurrentValue:  fmt.Sprintf("%.0f conns (Limit: %.0f, Shard %s)", maxConnsPerShard, limit, worstShard),
				ExpectedValue: "< 80% utilization",
				Status:        status,
				Remediation:   "Your network concurrency is near the hard limit. Increase '--max-networking-io-control-blocks' (e.g., to 50000) to prevent the network stack from stalling under high connection loads.",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Network Performance",
				Name:          "Network AIO Concurrency",
				Description:   "Active connections relative to AIO control block limit",
				CurrentValue:  fmt.Sprintf("%.1f%% utilized", utilization),
				ExpectedValue: "Healthy",
				Status:        SeverityPass,
			})
		}
	}

	return results
}

func checkSelfTestPerformance(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	if len(data.SelfTestResults) == 0 {
		return results
	}

	for _, nodeResult := range data.SelfTestResults {
		for _, test := range nodeResult.Results {
			if test.Type != "disk" {
				continue
			}

			// Small block write threshold: 4 MB/s
			if strings.Contains(test.Name, "4KB sequential write") && strings.Contains(test.Info, "dsync: true") {
				throughputMBps := float64(test.Throughput) / (1024 * 1024)
				if throughputMBps < 4.0 {
					results = append(results, CheckResult{
						Category:      "Hardware Performance",
						Name:          fmt.Sprintf("Disk Write Performance (Small Block) - Node %d", nodeResult.NodeID),
						Description:   "4KB sequential write throughput is below the minimum supported threshold",
						CurrentValue:  fmt.Sprintf("%.2f MB/s", throughputMBps),
						ExpectedValue: ">= 4.00 MB/s",
						Status:        SeverityCritical,
						Remediation:   "Your disk does not meet the minimum performance threshold for small block writes (dsync: true). This will cause high latency and low throughput. Check your storage subsystem (EBS type, RAID config, etc.).",
					})
				} else {
					results = append(results, CheckResult{
						Category:      "Hardware Performance",
						Name:          fmt.Sprintf("Disk Write Performance (Small Block) - Node %d", nodeResult.NodeID),
						Description:   "4KB sequential write throughput",
						CurrentValue:  fmt.Sprintf("%.2f MB/s", throughputMBps),
						ExpectedValue: ">= 4.00 MB/s",
						Status:        SeverityPass,
					})
				}
			}

			// Medium block write threshold: 30 MB/s
			if strings.Contains(test.Name, "16KB sequential") && strings.Contains(test.Info, "dsync: false") {
				throughputMBps := float64(test.Throughput) / (1024 * 1024)
				if throughputMBps < 30.0 {
					results = append(results, CheckResult{
						Category:      "Hardware Performance",
						Name:          fmt.Sprintf("Disk Write Performance (Medium Block) - Node %d", nodeResult.NodeID),
						Description:   "16KB sequential write throughput is below the minimum supported threshold",
						CurrentValue:  fmt.Sprintf("%.2f MB/s", throughputMBps),
						ExpectedValue: ">= 30.00 MB/s",
						Status:        SeverityCritical,
						Remediation:   "Your disk does not meet the minimum performance threshold for medium block writes. This indicates your disk throughput is insufficient for production Redpanda workloads.",
					})
				} else {
					results = append(results, CheckResult{
						Category:      "Hardware Performance",
						Name:          fmt.Sprintf("Disk Write Performance (Medium Block) - Node %d", nodeResult.NodeID),
						Description:   "16KB sequential write throughput",
						CurrentValue:  fmt.Sprintf("%.2f MB/s", throughputMBps),
						ExpectedValue: ">= 30.00 MB/s",
						Status:        SeverityPass,
					})
				}
			}
		}
	}

	return results
}

func checkGP3PerformanceCliff(data *cache.CachedData, s store.Store) []CheckResult {
	var results []CheckResult

	// Only relevant for AWS
	if data.System.Uname.CloudProvider != "AWS" {
		return results
	}

	// 1. Get configured limits from RedpandaConfig (iotune output)
	var limitIOPS float64
	if val, ok := data.RedpandaConfig["redpanda.write_iops"]; ok {
		if s, ok := val.(string); ok {
			limitIOPS, _ = strconv.ParseFloat(s, 64)
		}
	}
	if _, ok := data.RedpandaConfig["redpanda.write_bandwidth"]; ok {
		// (Bandwidth check could be added here in the future)
	}

	if limitIOPS == 0 {
		return results
	}

	// 2. Get current usage metrics
	metrics, err := s.GetMetrics("vectorized_io_queue_total_write_ops", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil || len(metrics) < 2 {
		return results
	}

	// Calculate current IOPS rate
	var totalCurrentIOPS float64
	// Group by shard and calculate rate
	shardMetrics := make(map[string][]models.PrometheusMetric)
	for _, m := range metrics {
		shard := m.Labels["shard"]
		shardMetrics[shard] = append(shardMetrics[shard], *m)
	}

	for _, ms := range shardMetrics {
		if len(ms) >= 2 {
			first, last := ms[0], ms[len(ms)-1]
			dur := last.Timestamp.Sub(first.Timestamp).Seconds()
			if dur > 0 {
				totalCurrentIOPS += (last.Value - first.Value) / dur
			}
		}
	}

	if totalCurrentIOPS > 0 {
		utilization := (totalCurrentIOPS / limitIOPS) * 100
		
		// The document says performance breaks down at 70% of IOPS
		if utilization > 70.0 {
			results = append(results, CheckResult{
				Category:      "Disk Performance",
				Name:          "AWS GP3 IOPS Cliff",
				Description:   "EBS gp3 volumes experience high tail latency above 70% utilization",
				CurrentValue:  fmt.Sprintf("%.1f%% IOPS Util", utilization),
				ExpectedValue: "< 70%",
				Status:        SeverityWarning,
				Remediation:   "AWS gp3 volumes often 'fall off a cliff' at 70% IOPS capacity. Increase volume IOPS or switch to io2/Block Express for consistent tail latencies.",
			})
		}
	}

	return results
}

func checkSoftIRQImbalance(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	if len(data.System.SoftIRQs.Entries) == 0 {
		return results
	}

	for _, entry := range data.System.SoftIRQs.Entries {
		if entry.IRQID != "NET_RX" && entry.IRQID != "BLOCK" {
			continue
		}

		// Calculate distribution
		var total int64
		var activeCores int
		for _, count := range entry.CPUCounts {
			total += count
			if count > (total / int64(len(entry.CPUCounts)) / 10) { // More than 10% of avg
				activeCores++
			}
		}

		if total < 1000000 {
			continue // Ignore low volume clusters
		}

		numCPUs := len(entry.CPUCounts)
		activePct := (float64(activeCores) / float64(numCPUs)) * 100

		// If < 10% of cores are doing 90% of the work on a large machine
		if activePct < 10.0 && numCPUs > 16 {
			results = append(results, CheckResult{
				Category:      "Kernel Performance",
				Name:          fmt.Sprintf("SoftIRQ Imbalance (%s)", entry.IRQID),
				Description:   fmt.Sprintf("%s interrupts are concentrated on only %d/%d cores", entry.IRQID, activeCores, numCPUs),
				CurrentValue:  fmt.Sprintf("%.1f%% core utilization", activePct),
				ExpectedValue: "> 25% core distribution",
				Status:        SeverityWarning,
				Remediation:   fmt.Sprintf("Kernel %s processing is bottlenecked on a few cores. This limits aggregate throughput. Check 'rpk redpanda tune network' or RPS/XPS settings.", entry.IRQID),
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Kernel Performance",
				Name:          fmt.Sprintf("SoftIRQ Imbalance (%s)", entry.IRQID),
				Description:   fmt.Sprintf("%s interrupt distribution across %d cores", entry.IRQID, numCPUs),
				CurrentValue:  fmt.Sprintf("%.1f%% core distribution", activePct),
				ExpectedValue: "Distributed",
				Status:        SeverityPass,
			})
		}
	}

	return results
}

func checkIOConfiguration(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for iotune properties in redpanda.yaml or specific keys
	// Common default values often seen when iotune hasn't run or is conservative
	isDefault := false
	hasIOConfig := false

	keys := []string{
		"redpanda.developer_mode",
		"redpanda.read_iops",
		"redpanda.write_iops",
		"redpanda.read_bandwidth",
		"redpanda.write_bandwidth",
	}

	for _, k := range keys {
		if _, ok := data.RedpandaConfig[k]; ok {
			hasIOConfig = true
		}
	}

	// Heuristic: Check for common "round" default numbers
	if val, ok := data.RedpandaConfig["redpanda.read_iops"]; ok {
		if s, ok := val.(string); ok && (s == "10000" || s == "1000") {
			isDefault = true
		}
	}

	if !hasIOConfig {
		results = append(results, CheckResult{
			Category:      "Disk Performance",
			Name:          "IO Properties Missing",
			Description:   "Disk IO performance properties (iotune) are not configured",
			CurrentValue:  "None",
			ExpectedValue: "Configured via 'rpk iotune'",
			Status:        SeverityWarning,
			Remediation:   "Run 'rpk redpanda tune all' or 'rpk iotune' to benchmark your disks. Without this, Redpanda uses conservative defaults that will limit throughput.",
		})
	} else if isDefault {
		results = append(results, CheckResult{
			Category:      "Disk Performance",
			Name:          "Default IO Properties",
			Description:   "Redpanda appears to be using default/conservative IO limits",
			CurrentValue:  "Default Values Found",
			ExpectedValue: "Hardware-specific values",
			Status:        SeverityInfo,
			Remediation:   "The configured IOPS/Bandwidth look like defaults. Ensure 'rpk iotune' was run on the production hardware.",
		})
	}

	// Check for Max Cost Function
	if val, ok := data.RedpandaConfig["redpanda.max_cost_function"]; ok {
		if s, ok := val.(string); ok && strings.ToLower(s) == "false" {
			results = append(results, CheckResult{
				Category:      "Disk Performance",
				Name:          "IO Cost Function",
				Description:   "Standard Seastar cost function is used instead of Max",
				CurrentValue:  "max_cost_function: false",
				ExpectedValue: "true",
				Status:        SeverityWarning,
				Remediation:   "Redpanda's 'max' cost function models disks better. Disabling it can lead to significant IO under-utilization.",
			})
		}
	}

	return results
}

func checkAIOSaturation(s store.Store) []CheckResult {
	var results []CheckResult

	// The hard limit is 1024 per shard for linux-aio
	metrics, err := s.GetMetrics("vectorized_io_queue_queue_length", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		return results
	}

	maxQueue := 0.0
	worstShard := ""
	for _, m := range metrics {
		if m.Value > maxQueue {
			maxQueue = m.Value
			worstShard = m.Labels["shard"]
		}
	}

	if maxQueue >= 1000.0 { // Near the 1024 limit
		status := SeverityWarning
		if maxQueue >= 1024.0 {
			status = SeverityCritical
		}

		results = append(results, CheckResult{
			Category:      "Disk Performance",
			Name:          "AIO Context Saturation",
			Description:   "IO requests are hitting the hard limit of the AIO sink",
			CurrentValue:  fmt.Sprintf("%.0f requests (Shard %s)", maxQueue, worstShard),
			ExpectedValue: "< 1024",
			Status:        status,
			Remediation:   "The shard is submitting IO faster than the kernel context can handle (1024 limit). This indicates an IO bottleneck in the kernel or extremely small IO sizes. Check for 'Reactor stalled' logs.",
		})
	}

	return results
}

func checkSluggishScheduler(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	valStr, exists := data.System.Sysctl["kernel.sched_wakeup_granularity_ns"]
	if !exists {
		return results
	}

	val, _ := strconv.Atoi(valStr)
	if val < 10000000 {
		return results // Granularity is healthy
	}

	isRHEL8 := data.System.Uname.Distro == "RHEL/CentOS 8"
	
	// Check Redpanda version for auto-mitigation (starting 25.1.5)
	version := data.RedpandaVersion
	hasAutoMitigation := false
	if version != "" {
		hasAutoMitigation = isVersionGreaterOrEqual(version, "25.1.5")
	}

	status := SeverityWarning
	if !hasAutoMitigation {
		status = SeverityCritical
	}

	desc := "Scheduler wakeup granularity is too high"
	if isRHEL8 {
		desc += " (known RHEL-8 'tuned' issue)"
	}

	remediation := "High granularity prevents Redpanda threads from yielding to kernel IO threads, causing extremely poor IOPS. "
	if isRHEL8 {
		remediation += "This is caused by the default 'tuned' profile on RHEL-8. "
	}
	remediation += "Action: Set 'kernel.sched_wakeup_granularity_ns' to 4000000 or use '--overprovisioned' mode."
	
	if hasAutoMitigation {
		remediation += " Note: Your Redpanda version has auto-mitigation, but the underlying OS setting is still high."
	}

	results = append(results, CheckResult{
		Category:      "OS Tuning",
		Name:          "kernel.sched_wakeup_granularity_ns",
		Description:   desc,
		CurrentValue:  fmt.Sprintf("%d ns", val),
		ExpectedValue: "< 10000000 ns",
		Status:        status,
		Remediation:   remediation,
	})

	return results
}

func isVersionGreaterOrEqual(current, target string) bool {
	// Simple semver comparison for major.minor.patch
	currParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	targetParts := strings.Split(target, ".")

	for i := 0; i < len(currParts) && i < len(targetParts); i++ {
		c, _ := strconv.Atoi(currParts[i])
		t, _ := strconv.Atoi(targetParts[i])
		if c > t {
			return true
		}
		if c < t {
			return false
		}
	}
	return len(currParts) >= len(targetParts)
}

func checkCPUStealTime(data *cache.CachedData, s store.Store) []CheckResult {
	var results []CheckResult

	// Get steal time metrics
	metrics, err := s.GetMetrics("vectorized_reactor_cpu_steal_time_ms", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil || len(metrics) < 2 {
		return results
	}

	// Group by shard to find the worst offender
	shardSteal := make(map[string][]models.PrometheusMetric)
	for _, m := range metrics {
		shard := m.Labels["shard"]
		shardSteal[shard] = append(shardSteal[shard], *m)
	}

	maxRate := 0.0
	worstShard := ""

	for shard, ms := range shardSteal {
		if len(ms) < 2 {
			continue
		}
		// Calculate rate between first and last sample in the bundle
		first := ms[0]
		last := ms[len(ms)-1]
		
		duration := last.Timestamp.Sub(first.Timestamp).Seconds()
		if duration > 0 {
			rate := (last.Value - first.Value) / duration
			if rate > maxRate {
				maxRate = rate
				worstShard = shard
			}
		}
	}

	maxRateMs := maxRate
	if maxRate > 1000.0 {
		// Values > 1000 in a per-second rate likely mean the metric is actually in microseconds
		maxRateMs = maxRate / 1000.0
	}

	if maxRateMs > 10.0 { // > 10ms per second stolen
		status := SeverityWarning
		if maxRateMs > 100.0 {
			status = SeverityCritical
		}

		remediation := "CPU Steal occurs when the host OS or hypervisor takes CPU cycles away from Redpanda. This causes unpredictable latency spikes and performance degradation.\n\n" +
			"Actionable Steps:\n" +
			"• Dedicated Nodes: Use Kubernetes taints/tolerations to ensure Redpanda has exclusive access to the node's CPUs.\n" +
			"• CPU Pinning: Set 'requests' equal to 'limits' to enable the 'Guaranteed' QoS class, which helps the Kubelet's CPU Manager isolate Redpanda.\n" +
			"• Overprovisioned Mode: If sharing nodes is necessary, enable '--overprovisioned' (or 'redpanda.overprovisioned: true'). This stops Redpanda from busy-spinning on the CPU, reducing contention with other processes.\n" +
			"• Resolve Noisy Neighbors: Investigate the busy processes listed below."
		
		// Try to identify noisy neighbors from process list
		neighborCounts := make(map[string]int)
		neighborCPUs := make(map[string]float64)
		for _, p := range data.System.Processes {
			if p.CPU > 5.0 { // > 5% CPU
				cmd := strings.ToLower(p.Command)
				// Filter out Redpanda and common kernel noise
				isRP := strings.Contains(cmd, "redpanda") || strings.Contains(cmd, "rp_") || strings.Contains(cmd, "reactor")
				isKernel := strings.HasPrefix(cmd, "[") && strings.HasSuffix(cmd, "]")
				
				if !isRP && !isKernel {
					neighborCounts[p.Command]++
					neighborCPUs[p.Command] += p.CPU
				}
			}
		}

		if len(neighborCounts) > 0 {
			type neighbor struct {
				name string
				cpu  float64
				text string
			}
			var list []neighbor
			for cmd, count := range neighborCounts {
				avgCPU := neighborCPUs[cmd] / float64(count)
				text := ""
				if count > 1 {
					text = fmt.Sprintf("• %s x%d (avg %.1f%% CPU)", cmd, count, avgCPU)
				} else {
					text = fmt.Sprintf("• %s (%.1f%% CPU)", cmd, avgCPU)
				}
				list = append(list, neighbor{cmd, neighborCPUs[cmd], text})
			}
			
			sort.Slice(list, func(i, j int) bool {
				return list[i].cpu > list[j].cpu
			})

			var busyNeighbors []string
			for _, n := range list {
				busyNeighbors = append(busyNeighbors, n.text)
			}
			remediation = fmt.Sprintf("%s\n\nDetected Noisy Neighbors:\n%s", remediation, strings.Join(busyNeighbors, "\n"))
		}

		results = append(results, CheckResult{
			Category:      "Kubernetes Performance",
			Name:          "High CPU Steal Time",
			Description:   "The OS is taking CPU cycles away from Redpanda shards",
			CurrentValue:  fmt.Sprintf("%.1f ms/sec (Shard %s)", maxRateMs, worstShard),
			ExpectedValue: "< 10 ms/sec",
			Status:        status,
			Remediation:   remediation,
		})
	}

	return results
}

func checkK8sShardOptimization(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	if !data.System.Uname.IsContainer || data.System.CoreCount == 0 {
		return results
	}

	cmdLine := strings.ToLower(data.System.CmdLine)
	isOverprovisioned := strings.Contains(cmdLine, "--overprovisioned")
	hostCores := data.System.CoreCount

	// Determine shard count from config or cmdline
	smpVal := 0
	if val, ok := data.RedpandaConfig["redpanda.smp"]; ok {
		if s, ok := val.(string); ok {
			smpVal, _ = strconv.Atoi(s)
		} else if i, ok := val.(int); ok {
			smpVal = i
		}
	}
	// Fallback to cmdline
	if smpVal == 0 && strings.Contains(cmdLine, "--smp") {
		parts := strings.Fields(cmdLine)
		for i, p := range parts {
			if strings.HasPrefix(p, "--smp") {
				val := ""
				if strings.Contains(p, "=") {
					val = strings.Split(p, "=")[1]
				} else if i+1 < len(parts) {
					val = parts[i+1]
				}
				smpVal, _ = strconv.Atoi(val)
				break
			}
		}
	}

	if smpVal == 0 {
		return results // Could not determine shard count
	}

	// 1. Audit "K8s CPU Tax" (N-1 logic)
	if smpVal >= hostCores && hostCores > 1 && !isOverprovisioned {
		results = append(results, CheckResult{
			Category:      "Kubernetes Performance",
			Name:          "K8s Shard Tax (N-1)",
			Description:   "Redpanda is using all host cores as shards in a container",
			CurrentValue:  fmt.Sprintf("%d shards on %d cores", smpVal, hostCores),
			ExpectedValue: fmt.Sprintf("%d shards (N-1)", hostCores-1),
			Status:        SeverityWarning,
			Remediation:   "In K8s, it is recommended to run Redpanda with one fewer shard than host cores (N-1) to leave room for the Kubelet and system processes. Action: Set '--smp' to N-1 or ensure '--overprovisioned' is used if cores must be shared.",
		})
	}

	// 2. Large Core Count Node Check
	if hostCores >= 32 && smpVal > 0 && smpVal < (hostCores/2) {
		if !isOverprovisioned {
			results = append(results, CheckResult{
				Category:      "Kubernetes Performance",
				Name:          "Large Core Node Interference",
				Description:   "Redpanda is using a small subset of a large-core host",
				CurrentValue:  fmt.Sprintf("%d Shards on %d Core Host", smpVal, hostCores),
				ExpectedValue: "--overprovisioned=true",
				Status:        SeverityWarning,
				Remediation:   "Redpanda can experience poor IOPS on Kubernetes nodes with large core counts if it only uses a small fraction of the cores. Action: Enable '--overprovisioned' mode to mitigate kernel scheduler/DIO interference.",
			})
		}
	}

	// 3. Busy Spinning in Shared Environment
	if !isOverprovisioned {
		results = append(results, CheckResult{
			Category:      "Kubernetes Performance",
			Name:          "Busy Spinning Mode",
			Description:   "Redpanda busy-spinning is enabled in a potentially shared environment",
			CurrentValue:  "spinning=on",
			ExpectedValue: "spinning=off (via --overprovisioned)",
			Status:        SeverityInfo,
			Remediation:   "Redpanda reactor threads busy-spin by default for lowest latency. In shared Kubernetes environments, this can cause contention. If throughput is low or CPU usage is high without useful work, consider enabling '--overprovisioned'.",
		})
	}

	return results
}

func checkConnectionFlapping(s store.Store) []CheckResult {
	var results []CheckResult

	// Search for high frequency of new connections in logs
	entries, _, err := s.GetLogs(&models.LogFilter{
		Search: "Accepted new connection",
		Limit:  100,
	})

	if err == nil && len(entries) >= 50 {
		// If we found 50+ connection logs, check the time spread
		first := entries[len(entries)-1].Timestamp
		last := entries[0].Timestamp
		duration := last.Sub(first)

		// If 50 connections happened in < 10 seconds
		if duration > 0 && duration < 10*time.Second {
			results = append(results, CheckResult{
				Category:      "Client Behavior",
				Name:          "High Connection Churn",
				Description:   "Clients are connecting/disconnecting at a high rate",
				CurrentValue:  fmt.Sprintf("%.1f conns/sec", float64(len(entries))/duration.Seconds()),
				ExpectedValue: "< 1 conn/sec",
				Status:        SeverityWarning,
				Remediation:   "High connection churn increases CPU usage and producer latency. Ensure clients are reusing producer/consumer instances instead of creating them per-message.",
			})
		}
	}

	return results
}

func checkMetadataOverload(s store.Store) []CheckResult {
	var results []CheckResult

	// Get Metadata vs Produce request counts
	metrics, err := s.GetMetrics("redpanda_kafka_request_latency_seconds_count", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		return results
	}

	var metadataCount, produceCount float64
	for _, m := range metrics {
		handler := m.Labels["request"]
		if handler == "metadata" {
			metadataCount += m.Value
		} else if handler == "produce" {
			produceCount += m.Value
		}
	}

	if produceCount > 1000 { // Only check if there's meaningful traffic
		ratio := metadataCount / produceCount
		if ratio > 0.1 { // > 10% ratio
			results = append(results, CheckResult{
				Category:      "Client Behavior",
				Name:          "Metadata Request Overload",
				Description:   "Producers are requesting metadata excessively relative to data",
				CurrentValue:  fmt.Sprintf("%.1f%% of produce rate", ratio*100),
				ExpectedValue: "< 1%",
				Status:        SeverityWarning,
				Remediation:   "Producers block application threads when fetching metadata. High metadata rates suggest client-side timeouts or cluster instability. Check 'metadata_dissemination_interval_ms'.",
			})
		}
	}

	return results
}

func checkConsumerParallelism(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	for groupName, group := range data.ConsumerGroups {
		if group.State == "Empty" || len(group.Members) == 0 {
			continue
		}

		// Check each topic this group is consuming
		topicsChecked := make(map[string]bool)
		for _, member := range group.Members {
			for topicName := range member.Assigned {
				if topicsChecked[topicName] {
					continue
				}
				topicsChecked[topicName] = true

				// Find partition count for this topic
				var partitionCount int
				if topic, ok := data.KafkaMetadata.Topics[topicName]; ok {
					partitionCount = len(topic.Partitions)
				}

				if partitionCount > 0 {
					memberCount := len(group.Members)
					if memberCount < partitionCount {
						results = append(results, CheckResult{
							Category:      "Consumer Efficiency",
							Name:          fmt.Sprintf("Under-provisioned Consumer (%s)", groupName),
							Description:   fmt.Sprintf("Group has fewer members (%d) than partitions (%d) in topic %s", memberCount, partitionCount, topicName),
							CurrentValue:  fmt.Sprintf("%d members", memberCount),
							ExpectedValue: fmt.Sprintf("%d members (min)", partitionCount),
							Status:        SeverityInfo,
							Remediation:   "Add more consumer instances to process partitions in parallel and increase throughput ceiling.",
						})
					} else if memberCount > partitionCount {
						results = append(results, CheckResult{
							Category:      "Consumer Efficiency",
							Name:          fmt.Sprintf("Over-provisioned Consumer (%s)", groupName),
							Description:   fmt.Sprintf("Group has more members (%d) than partitions (%d). Extra members are idle.", memberCount, partitionCount),
							CurrentValue:  fmt.Sprintf("%d members", memberCount),
							ExpectedValue: fmt.Sprintf("%d members", partitionCount),
							Status:        SeverityInfo,
							Remediation:   "Scale down consumer count or increase partition count to utilize all members.",
						})
					}
				}
			}
		}
	}

	return results
}

func checkShard0Overload(s store.Store) []CheckResult {
	var results []CheckResult

	// Use recent metrics to check shard utilization
	metrics, err := s.GetMetrics("vectorized_reactor_utilization", nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil || len(metrics) < 2 {
		return results
	}

	var shard0Util float64
	var otherSum float64
	var otherCount int

	for _, m := range metrics {
		shard := m.Labels["shard"]
		if shard == "0" {
			shard0Util = m.Value
		} else {
			otherSum += m.Value
			otherCount++
		}
	}

	if otherCount > 0 {
		avgOther := otherSum / float64(otherCount)
		// If shard 0 is 30% higher than average, flag it
		if shard0Util > (avgOther + 30.0) && shard0Util > 50.0 {
			results = append(results, CheckResult{
				Category:      "CPU Balance",
				Name:          "Shard 0 Overload",
				Description:   "Shard 0 is significantly busier than other shards",
				CurrentValue:  fmt.Sprintf("%.1f%% (Avg others: %.1f%%)", shard0Util, avgOther),
				ExpectedValue: "Balanced (< 30% delta)",
				Status:        SeverityWarning,
				Remediation:   "Shard 0 handles administrative tasks. If it is overloaded, cluster-wide latency and management operations (like rebalancing) may be delayed.",
			})
		}
	}

	return results
}

func checkPublicIPs(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Collect all unique advertised addresses
	publicAddrs := make(map[string]bool)
	
	// 1. Check from Config
	keys := []string{"advertised_kafka_api", "advertised_rpc_api", "advertised_pandaproxy_api"}
	for _, k := range keys {
		if val, ok := data.RedpandaConfig[k]; ok {
			if s, ok := val.(string); ok {
				// Handle lists or single values
				parts := strings.Split(s, ",")
				for _, p := range parts {
					addr := strings.TrimSpace(p)
					if idx := strings.LastIndex(addr, ":"); idx != -1 {
						addr = addr[:idx]
					}
					if isPublicIP(addr) {
						publicAddrs[addr] = true
					}
				}
			}
		}
	}

	// 2. Check from Metadata (What the cluster actually reports)
	for _, b := range data.KafkaMetadata.Brokers {
		if isPublicIP(b.Host) {
			publicAddrs[b.Host] = true
		}
	}

	if len(publicAddrs) > 0 {
		var addrList []string
		for addr := range publicAddrs {
			addrList = append(addrList, addr)
		}

		remediation := "Use private VPC IP addresses for internal communication. "
		if data.System.Uname.CloudProvider == "AWS" {
			remediation += "AWS imposes a 50% bandwidth 'haircut' on traffic using public IPs, even within the same VPC."
		} else if data.System.Uname.CloudProvider == "GCP" {
			remediation += "GCP limits public IP egress to 7 Gbps regardless of instance size."
		} else {
			remediation += "Using public IPs for cluster communication often incurs significant bandwidth penalties and latency on cloud providers."
		}

		results = append(results, CheckResult{
			Category:      "Network Configuration",
			Name:          "Public Advertised Addresses",
			Description:   "Redpanda is advertising public IP addresses for its APIs",
			CurrentValue:  strings.Join(addrList, ", "),
			ExpectedValue: "Private IPs (RFC1918)",
			Status:        SeverityWarning,
			Remediation:   remediation,
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Network Configuration",
			Name:          "Public Advertised Addresses",
			Description:   "Redpanda is advertising public IP addresses for its APIs",
			CurrentValue:  "None (Private IPs used)",
			ExpectedValue: "Private IPs (RFC1918)",
			Status:        SeverityPass,
		})
	}

	return results
}

func isPublicIP(ip string) bool {
	if ip == "" || ip == "localhost" || ip == "127.0.0.1" || ip == "::1" || strings.HasSuffix(ip, ".local") {
		return false
	}

	// If it's a hostname (contains dots but no numbers), it's hard to tell without DNS, 
	// but usually public if it's not a shortname. 
	// For simplicity, let's focus on literal IPs.
	
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false // Not an IPv4 literal
	}

	p1, _ := strconv.Atoi(parts[0])
	p2, _ := strconv.Atoi(parts[1])

	// RFC1918 Private Ranges
	if p1 == 10 {
		return false
	}
	if p1 == 172 && p2 >= 16 && p2 <= 31 {
		return false
	}
	if p1 == 192 && p2 == 168 {
		return false
	}
	
	// Carrier Grade NAT
	if p1 == 100 && p2 >= 64 && p2 <= 127 {
		return false
	}

	return true // It's a public IP
}

func checkLogErrors(s store.Store) []CheckResult {
	var results []CheckResult

	// 1. Partition Rebalance Failures
	entries, _, err := s.GetLogs(&models.LogFilter{
		Search: "no nodes are available to perform allocation",
		Limit:  1,
	})
	if err == nil && len(entries) > 0 {
		results = append(results, CheckResult{
			Category:      "Cluster Balance",
			Name:          "Rebalance Failure (Hard Constraints)",
			Description:   "Balancer cannot find a valid node for a partition due to hard constraints (disk full, rack awareness).",
			CurrentValue:  "Errors Found",
			ExpectedValue: "None",
			Status:        SeverityCritical,
			Remediation:   "Check disk usage (df -h), rack configuration, or add nodes. Search logs for 'No nodes are available'.",
		})
	}

	// 2. Generic Move Failures
	if len(results) == 0 { // Only check this if the specific one wasn't found to avoid noise
		entries, _, err = s.GetLogs(&models.LogFilter{
			Search: "attempt to move replica",
			Limit:  10, // Get a few to check for "failed"
		})
		if err == nil {
			foundFailure := false
			for _, e := range entries {
				if strings.Contains(e.Message, "failed") {
					foundFailure = true
					break
				}
			}
			if foundFailure {
				results = append(results, CheckResult{
					Category:      "Cluster Balance",
					Name:          "Partition Move Failed",
					Description:   "Attempts to move partition replicas are failing.",
					CurrentValue:  "Failures Found",
					ExpectedValue: "None",
					Status:        SeverityWarning,
					Remediation:   "Check logs for 'attempt to move replica' to see the reason (e.g., disk_full, node_down).",
				})
			}
		}
	}

	return results
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
	checkInt("fs.file-max", 500000, SeverityWarning, "Max open file descriptors (system-wide)", "Increase fs.file-max in /etc/sysctl.conf")
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

func checkBottlenecks(data *cache.CachedData, s store.Store) []CheckResult {
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
		"redpanda_cpu_busy_seconds_total",
		"vectorized_reactor_cpu_busy_ms",
		"vectorized_reactor_stalls",
		"vectorized_reactor_threaded_fallbacks",
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
	report := analysis.AnalyzePerformance(metricsBundle, data.SarData, data.SelfTestResults)

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

	// Check auto_create_topics_enabled
	if isTrue("auto_create_topics_enabled") {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "auto_create_topics_enabled",
			Description:   "Automatic topic creation",
			CurrentValue:  "true",
			ExpectedValue: "false",
			Status:        SeverityWarning,
			Remediation:   "Disable auto_create_topics_enabled in production to prevent accidental topic creation and partition count explosion.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Redpanda Config",
			Name:          "auto_create_topics_enabled",
			Description:   "Automatic topic creation",
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

	dataDir := data.RedpandaDataDir
	if dataDir == "" {
		dataDir = "/var/lib/redpanda/data"
	}
	if data.StorageAnalysis != nil && data.StorageAnalysis.EffectiveDataDir != "" {
		dataDir = data.StorageAnalysis.EffectiveDataDir
	}

	// Find best match FS for data directory
	var dataDirFS models.FileSystemEntry
	maxLen := -1
	for _, fs := range data.System.FileSystems {
		if strings.HasPrefix(dataDir, fs.MountPoint) {
			if len(fs.MountPoint) > maxLen {
				maxLen = len(fs.MountPoint)
				dataDirFS = fs
			}
		}
	}

	isNoisyMount := func(mount string) bool {
		noisyPrefixes := []string{"/etc/", "/dev/", "/proc/", "/sys/", "/run/secrets/", "/var/lib/kubelet/"}
		for _, prefix := range noisyPrefixes {
			if strings.HasPrefix(mount, prefix) {
				return true
			}
		}
		return false
	}

	// Check for high disk usage and Filesystem Type
	for _, df := range data.System.FileSystems {
		isDataFS := (maxLen != -1 && df.MountPoint == dataDirFS.MountPoint)
		isRedpandaMount := strings.Contains(df.MountPoint, "redpanda")
		isPhysicalDisk := strings.HasPrefix(df.Filesystem, "/dev/")

		// We only want XFS warnings for Redpanda data locations
		if isDataFS || isRedpandaMount {
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
		}

		// Usage check: check data FS, redpanda mounts, AND physical disks
		// BUT filter out small/irrelevant stuff that usually clutters df
		if isDataFS || isRedpandaMount || (isPhysicalDisk && !isNoisyMount(df.MountPoint)) {
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

func checkTransparentHugePages(data *cache.CachedData) []CheckResult {
	var results []CheckResult
	
	thp := data.System.TransparentHugePages
	if thp == "" {
		// Not captured or not available
		return results
	}
	
	// Format is typically "[always] madvise never" or "always [madvise] never"
	// We want "never" to be selected, i.e., "[never]" or "never" if it's just a word
	
	if strings.Contains(thp, "[never]") {
		results = append(results, CheckResult{
			Category:      "OS Tuning",
			Name:          "Transparent Hugepages",
			Description:   "Kernel feature that can cause latency spikes",
			CurrentValue:  "never",
			ExpectedValue: "never",
			Status:        SeverityPass,
		})
	} else {
		// Extract selected one
		selected := "unknown"
		parts := strings.Fields(thp)
		for _, p := range parts {
			if strings.HasPrefix(p, "[") && strings.HasSuffix(p, "]") {
				selected = strings.Trim(p, "[]")
				break
			}
		}
		
		results = append(results, CheckResult{
			Category:      "OS Tuning",
			Name:          "Transparent Hugepages",
			Description:   "Kernel feature that can cause latency spikes",
			CurrentValue:  selected,
			ExpectedValue: "never",
			Status:        SeverityWarning,
			Remediation:   "Disable THP: echo never > /sys/kernel/mm/transparent_hugepage/enabled",
		})
	}
	
	return results
}

func checkK8sQoS(data *cache.CachedData) []CheckResult {
	var results []CheckResult
	
	// Only run if we have K8s data
	if len(data.K8sStore.Pods) == 0 {
		return results
	}
	
	// Check Redpanda pods (label app=redpanda or similar, or just check all in our store since we filter on ingestion usually)
	for _, pod := range data.K8sStore.Pods {
		for _, container := range pod.Spec.Containers {
			// Check if requests == limits for CPU and Memory
			// Note: This relies on parsed resource units which might be strings like "1000m" or "1G"
			// String comparison works if they are identical strings, but "1G" != "1024M"
			// Implementing full resource unit parsing is complex. Let's do a basic string equality check first.
			
			cpuReq := container.Resources.Requests["cpu"]
			cpuLim := container.Resources.Limits["cpu"]
			memReq := container.Resources.Requests["memory"]
			memLim := container.Resources.Limits["memory"]
			
			if cpuReq != cpuLim || memReq != memLim {
				results = append(results, CheckResult{
					Category:      "Kubernetes QoS",
					Name:          fmt.Sprintf("Pod %s QoS", pod.Metadata.Name),
					Description:   "Guaranteed QoS requires requests == limits",
					CurrentValue:  fmt.Sprintf("CPU: %s/%s, Mem: %s/%s", cpuReq, cpuLim, memReq, memLim),
					ExpectedValue: "Requests == Limits",
					Status:        SeverityWarning,
					Remediation:   "Set resources.requests equal to resources.limits to ensure Guaranteed QoS and prevent throttling.",
				})
			}
		}
	}
	
	return results
}

func checkIRQBalance(data *cache.CachedData) []CheckResult {
	var results []CheckResult
	
	if len(data.System.Interrupts.Entries) == 0 {
		return results
	}
	
	// deviceGroup represents a set of IRQs belonging to a single physical device (e.g., nvme0)
	type deviceGroup struct {
		Name       string
		IRQs       []models.IRQEntry
		TotalCount int64
		CPUDist    map[int]int64 // CPU ID -> Total Interrupts for this device
	}
	
	groups := make(map[string]*deviceGroup)
	
	// Identify relevant interrupts (Network and Disk)
	keywords := []string{"nvme", "eth", "ens", "eno", "scsi", "virtio", "mlx"}
	
	// Helper to extract base device name
	getBaseName := func(s string) string {
		// nvme0q1 -> nvme0
		if strings.HasPrefix(s, "nvme") {
			if idx := strings.Index(s, "q"); idx != -1 {
				return s[:idx]
			}
		}
		// eth0-TxRx-0 -> eth0
		if idx := strings.Index(s, "-"); idx != -1 {
			return s[:idx]
		}
		return s
	}

	for _, entry := range data.System.Interrupts.Entries {
		isRelevant := false
		lowerDev := strings.ToLower(entry.Device)
		lowerCtrl := strings.ToLower(entry.Controller)
		
		for _, k := range keywords {
			if strings.Contains(lowerDev, k) || strings.Contains(lowerCtrl, k) {
				isRelevant = true
				break
			}
		}
		
		if !isRelevant {
			continue
		}
		
		baseName := getBaseName(entry.Device)
		if _, exists := groups[baseName]; !exists {
			groups[baseName] = &deviceGroup{
				Name:    baseName,
				CPUDist: make(map[int]int64),
			}
		}
		
		// Sum counts
		var entryTotal int64
		for i, count := range entry.CPUCounts {
			entryTotal += count
			groups[baseName].CPUDist[i] += count
		}
		
		groups[baseName].IRQs = append(groups[baseName].IRQs, entry)
		groups[baseName].TotalCount += entryTotal
	}
	
	// Analyze Groups
	for _, g := range groups {
		if g.TotalCount < 10000 {
			continue // Ignore low volume devices
		}
		
		// 1. Check if the device as a whole is pinned to a single CPU
		var maxCPU int
		var maxCount int64
		activeCPUs := 0
		
		for cpu, count := range g.CPUDist {
			if count > 0 {
				activeCPUs++
			}
			if count > maxCount {
				maxCount = count
				maxCPU = cpu
			}
		}
		
		// Logic:
		// If Multi-Queue (len(IRQs) > 1), we expect activeCPUs > 1 (Distributed)
		// We skip Single-Queue devices as pinning is often intentional or unavoidable.
		
		if len(g.IRQs) > 1 {
			// Multi-Queue Logic
			// If all traffic is on 1 CPU despite having multiple queues, that's bad.
			if activeCPUs == 1 {
				results = append(results, CheckResult{
					Category:      "Hardware & IRQ",
					Name:          fmt.Sprintf("MQ Imbalance (%s)", g.Name),
					Description:   fmt.Sprintf("Multi-Queue device %s has %d queues but is pinned to single CPU%d", g.Name, len(g.IRQs), maxCPU),
					CurrentValue:  "Pinned to 1 CPU",
					ExpectedValue: "Distributed across CPUs",
					Status:        SeverityWarning,
					Remediation:   "Check 'rpk redpanda tune cpu' or irqbalance settings. All queues are hitting one core.",
				})
			}
		}
	}
	
	return results
}

func checkDataDirPartition(data *cache.CachedData) []CheckResult {
	var results []CheckResult
	
	dataDir := data.RedpandaDataDir
	if dataDir == "" {
		dataDir = "/var/lib/redpanda/data" // Default guess if not parsed
	}

	// Use effective data directory (resolved from actual segment paths) if available
	// This handles symlinks where configured dir is on root but actual data is on a mount.
	if data.StorageAnalysis != nil && data.StorageAnalysis.EffectiveDataDir != "" {
		dataDir = data.StorageAnalysis.EffectiveDataDir
	}
	
	// Find mount point
	var bestMatch models.FileSystemEntry
	maxLen := -1
	
	for _, fs := range data.System.FileSystems {
		if strings.HasPrefix(dataDir, fs.MountPoint) {
			if len(fs.MountPoint) > maxLen {
				maxLen = len(fs.MountPoint)
				bestMatch = fs
			}
		}
	}
	
	if maxLen != -1 {
		if bestMatch.MountPoint == "/" {
			results = append(results, CheckResult{
				Category:      "Disk Configuration",
				Name:          "Data Directory on Root",
				Description:   "Redpanda data directory is on the root partition",
				CurrentValue:  fmt.Sprintf("%s on /", dataDir),
				ExpectedValue: "Separate Partition",
				Status:        SeverityWarning,
				Remediation:   "Move Redpanda data to a dedicated XFS partition to avoid OS contention.",
			})
		}
	}
	
	return results
}

func checkRPCTimeouts(s store.Store) []CheckResult {
	var results []CheckResult

	// 1. Check redpanda_node_status_rpcs_timed_out_total (Liveness checks)
	statusTimeouts, _ := s.GetMetrics("redpanda_node_status_rpcs_timed_out_total", nil, time.Time{}, time.Time{}, 0, 0)
	totalStatusTimeouts := 0.0
	for _, m := range statusTimeouts {
		totalStatusTimeouts += m.Value
	}

	if totalStatusTimeouts > 0 {
		results = append(results, CheckResult{
			Category:      "Consensus",
			Name:          "Node Status RPC Timeouts",
			Description:   "Timeouts detected in the 100ms internal liveness checks",
			CurrentValue:  fmt.Sprintf("%.0f timeouts", totalStatusTimeouts),
			ExpectedValue: "0 timeouts",
			Status:        SeverityWarning,
			Remediation:   "Node status timeouts are 'leading indicators' of metadata pressure. They often trigger hostile elections. Action: Increase 'node_status_interval' to 1000ms.",
		})
	}

	// 2. Check redpanda_rpc_client_requests_blocked_memory_total
	blockedMem, _ := s.GetMetrics("redpanda_rpc_client_requests_blocked_memory_total", nil, time.Time{}, time.Time{}, 0, 0)
	totalBlocked := 0.0
	for _, m := range blockedMem {
		totalBlocked += m.Value
	}

	if totalBlocked > 0 {
		results = append(results, CheckResult{
			Category:      "Performance",
			Name:          "RPC Memory Backpressure",
			Description:   "RPC requests were delayed because the node is out of memory for networking",
			CurrentValue:  fmt.Sprintf("%.0f requests blocked", totalBlocked),
			ExpectedValue: "0 blocked",
			Status:        SeverityWarning,
			Remediation:   "The node is hitting 'rpc_server_listen_backlog' or general memory limits. This causes severe latency. Action: Check if 'memory_limit' or 'rpc_max_service_mem_bytes' is too low.",
		})
	}

	return results
}

func checkIntegrityMetrics(metrics *models.MetricsBundle) []CheckResult {
	var results []CheckResult

	var parseErrors, writeErrors, corruptCompaction float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			switch metric.Name {
			case "vectorized_storage_log_batch_parse_errors":
				parseErrors += metric.Value
			case "vectorized_storage_log_batch_write_errors":
				writeErrors += metric.Value
			case "vectorized_storage_log_corrupted_compaction_indices":
				corruptCompaction += metric.Value
			}
		}
	}

	if parseErrors > 0 || writeErrors > 0 {
		results = append(results, CheckResult{
			Category:      "Data Integrity",
			Name:          "Log Storage Errors",
			Description:   "Redpanda encountered errors while parsing or writing log batches to disk",
			CurrentValue:  fmt.Sprintf("%.0f parse errors, %.0f write errors", parseErrors, writeErrors),
			ExpectedValue: "0 errors",
			Status:        SeverityCritical,
			Remediation:   "CRITICAL: Storage layer errors detected. This strongly indicates disk hardware failure, filesystem corruption (e.g. XFS errors), or OS-level IO timeout. Investigate 'dmesg' for 'I/O error' and Redpanda logs for 'segment_not_found', 'batch_parse_error', or 'checksum mismatch'.",
		})
	}

	if corruptCompaction > 0 {
		results = append(results, CheckResult{
			Category:      "Storage",
			Name:          "Compaction Index Corruption",
			Description:   "One or more log compaction indices are corrupted",
			CurrentValue:  fmt.Sprintf("%.0f indices corrupted", corruptCompaction),
			ExpectedValue: "0 corrupted",
			Status:        SeverityWarning,
			Remediation:   "Redpanda has to rebuild corrupted compaction indices on-the-fly, which consumes significant CPU (often visible as high runtime in the 'main' scheduler group) and disk IO. This usually occurs after abrupt shutdowns, SIGKILLs, or Kubernetes liveness probe timeouts.\n\nAction: Search the Logs for 'Rebuilding index file' to identify the affected segments. Note: Frequent rebuilding can be caused by pod restart loops.",
		})
	}

	return results
}

func checkMemoryReclaimPressure(metrics *models.MetricsBundle) []CheckResult {
	var results []CheckResult

	// vectorized_memory_reclaims_operations is a counter
	// Since we likely have only one or two snapshots, we look at the absolute value or existence
	var totalReclaims float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_memory_reclaims_operations" || metric.Name == "redpanda_memory_reclaims_operations" {
				totalReclaims += metric.Value
			}
		}
	}

	// In a single snapshot, if reclaims are > 1000, the node has been struggling for a while
	if totalReclaims > 10000 {
		results = append(results, CheckResult{
			Category:      "Performance",
			Name:          "High Memory Reclaim Rate",
			Description:   "Redpanda is frequently evicting data from its batch cache to satisfy memory requests",
			CurrentValue:  fmt.Sprintf("%.0f total reclaims", totalReclaims),
			ExpectedValue: "Low reclaim activity",
			Status:        SeverityWarning,
			Remediation:   "Frequent reclaims indicate the node is operating at the edge of its memory limit. This causes 'cache thrashing', where data is evicted then immediately re-read from disk, killing performance. Action: Increase node memory, decrease '--smp' (shards), or check for memory-heavy producers.",
		})
	}

	return results
}

func checkSocketHealth(data *cache.CachedData) []CheckResult {
	var results []CheckResult
	if len(data.System.Connections) == 0 {
		return results
	}

	highRetransCount := 0
	worstConn := ""
	maxRetrans := 0

	for _, conn := range data.System.Connections {
		if conn.TCPInfo != nil && conn.TCPInfo.RetransTotal > 50 {
			highRetransCount++
			if conn.TCPInfo.RetransTotal > maxRetrans {
				maxRetrans = conn.TCPInfo.RetransTotal
				worstConn = fmt.Sprintf("%s -> %s", conn.LocalAddr, conn.PeerAddr)
			}
		}
	}

	if highRetransCount > 0 {
		results = append(results, CheckResult{
			Category:      "Network Stack",
			Name:          "Socket Retransmissions",
			Description:   "Individual TCP connections are experiencing high retransmission counts",
			CurrentValue:  fmt.Sprintf("%d connections affected (Worst: %d retrans on %s)", highRetransCount, maxRetrans, worstConn),
			ExpectedValue: "0 or low retransmissions",
			Status:        SeverityWarning,
			Remediation:   "High per-socket retransmissions indicate packet loss on specific network paths. If many clients are affected, check the physical NIC or switch. If only one client is affected, investigate that client's network path.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Network Stack",
			Name:          "Socket Retransmissions",
			Description:   "Individual TCP connections health",
			CurrentValue:  "Healthy",
			ExpectedValue: "Healthy",
			Status:        SeverityPass,
		})
	}

	return results
}
