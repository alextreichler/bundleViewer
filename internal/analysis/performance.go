package analysis

import (
	"fmt"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// PerformanceReport holds the result of the performance analysis
type PerformanceReport struct {
	IsCPUBound      bool
	IsDiskBound     bool
	IsNetworkBound  bool   // Hard to determine from snapshot, but we can look for errors
	WorkloadType    string // "Produce-Heavy", "Fetch-Heavy", "Balanced", "Idle"
	Recommendations []string
	Observations    []string
	NetworkReport   *NetworkHealthReport
}

// AnalyzePerformance evaluates the metrics bundle against known performance heuristics
func AnalyzePerformance(metrics *models.MetricsBundle, sar models.SarData, selfTest []models.SelfTestNodeResult) PerformanceReport {
	report := PerformanceReport{
		Recommendations: []string{},
		Observations:    []string{},
	}

	if metrics == nil {
		return report
	}

	// Network Deep Dive (if SAR data is available)
	report.NetworkReport = AnalyzeNetwork(sar, metrics)
	if report.NetworkReport != nil {
		for _, anomaly := range report.NetworkReport.DetectedAnomalies {
			report.Observations = append(report.Observations, anomaly)
		}
		for _, rec := range report.NetworkReport.Recommendations {
			report.Recommendations = append(report.Recommendations, rec)
		}
	}

	// 1. CPU Analysis
	// Check reactor utilization (Gauge)
	maxReactorUtil := 0.0
	var shardUtils []float64
	
	// Better CPU Analysis (Counters)
	var cpuBusySeconds, cpuBusyMs float64
	hasCpuCounters := false

	// Stall & Fallback Metrics
	var reactorStallsHistogram float64 // Just checking existence/count for now
	var threadedFallbacks float64
	
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_reactor_utilization" {
				shardUtils = append(shardUtils, metric.Value)
				if metric.Value > maxReactorUtil {
					maxReactorUtil = metric.Value
				}
			} else if metric.Name == "redpanda_cpu_busy_seconds_total" {
				// This is a counter, so we need to see if we have multiple samples to calc rate, 
				// but for a snapshot, a high raw value just means uptime. 
				// To properly use this in a static bundle viewer without time-series queries is hard.
				// However, if we have multiple samples in the bundle, we might be able to calculate it.
				// For now, let's just flag that we found it.
				cpuBusySeconds += metric.Value
				hasCpuCounters = true
			} else if metric.Name == "vectorized_reactor_cpu_busy_ms" {
				cpuBusyMs += metric.Value
				hasCpuCounters = true
			} else if metric.Name == "vectorized_reactor_stalls" {
				reactorStallsHistogram += metric.Value
			} else if metric.Name == "vectorized_reactor_threaded_fallbacks" {
				threadedFallbacks += metric.Value
			}
		}
	}

	// NOTE: Ideally we would calculate utilization from the busy counters (rate(busy)/rate(uptime)),
	// but in a single-snapshot bundle context, we rely on the pre-calculated gauge unless we have time-series data.
	// If the gauge is high, we can check the counters for context if needed.
	if hasCpuCounters {
		if cpuBusySeconds > 0 {
			report.Observations = append(report.Observations, fmt.Sprintf("System Info: Found high-resolution CPU busy counters (%.0fs total).", cpuBusySeconds))
		} else if cpuBusyMs > 0 {
			report.Observations = append(report.Observations, fmt.Sprintf("System Info: Found high-resolution CPU busy counters (%.0fms total).", cpuBusyMs))
		}
	}

	if maxReactorUtil >= 100 {
		report.IsCPUBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("Critical: Reactor Utilization at 100%%. Shard is likely falling behind. Note: Seastar reactors poll eagerly, so 100%% OS CPU is normal; this metric tracks *useful* work."))
	} else if maxReactorUtil > 95 {
		report.IsCPUBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("Warning: High Reactor Utilization detected: %.2f%%. Risk of saturation.", maxReactorUtil))
	}

	// Check for Shard Imbalance (Hot Shard)
	if len(shardUtils) > 1 {
		totalUtil := 0.0
		for _, u := range shardUtils {
			totalUtil += u
		}
		avgUtil := totalUtil / float64(len(shardUtils))
		
		// If max is significantly higher than average (e.g. 30% deviation), flag it
		// Only relevant if the load is somewhat significant (> 20%)
		if maxReactorUtil > 20 && (maxReactorUtil - avgUtil) > 30 {
			report.Observations = append(report.Observations, fmt.Sprintf("Shard Imbalance detected. Max shard: %.0f%%, Avg: %.0f%%. One core is working much harder than others.", maxReactorUtil, avgUtil))
			report.Recommendations = append(report.Recommendations, "Investigate hot shard/partition distribution. A single heavy partition may be pinning one core.")
		}
	}

	// 2. Workload Classification (Produce vs Fetch)
	// Preferred method: Scheduler Groups (vectorized_scheduler_runtime_ms)
	var produceRuntime, fetchRuntime float64
	hasSchedulerMetrics := false

	// Gather metrics first
	var produceCount, fetchCount float64
	var sentBytes, receivedBytes float64
	var produceBatches, writtenBytes float64

	// Disk Latency Metrics
	var maxDiskQueue float64
	var ioDelaySec, ioExecSec, ioOps float64

	// Cache Metrics
	var cacheHits, cacheAccesses float64

	// Stall Metrics
	var reportedStalls, taskQuotaViolationsMs float64

	// Network Metrics
	var networkRx, networkTx float64
	var archivalUploadRuntime float64
	var tcpConnections float64

	// Raft Metrics
	var leadershipChanges float64
	var partitionsToRecover float64
	var ackAllRequests, ackLeaderRequests float64

	// Memory Metrics
	var freeMem, totalMem, reclaimableMem float64
	var availableLowWater float64
	var hasLowWater bool
	var uptimeSeconds float64

	for _, m := range metrics.Files {
		for _, metric := range m {
			switch metric.Name {
			case "vectorized_scheduler_runtime_ms":
				hasSchedulerMetrics = true
				group := metric.Labels["group"]
				// "produce" group or "main" (pre-25.1) + raft groups
				if group == "produce" || group == "main" || group == "raft_send" || group == "raft_receive" {
					produceRuntime += metric.Value
				} else if group == "fetch" {
					fetchRuntime += metric.Value
				} else if group == "archival_upload" {
					archivalUploadRuntime += metric.Value
				}
			case "vectorized_memory_free_bytes", "redpanda_memory_free_memory":
				freeMem += metric.Value
			case "vectorized_memory_allocated_bytes", "redpanda_memory_allocated_memory":
				totalMem += (metric.Value + freeMem) // Rough total per shard
			case "vectorized_storage_batch_cache_reclaimable_bytes", "redpanda_storage_batch_cache_reclaimable_bytes":
				reclaimableMem += metric.Value
			case "redpanda_memory_available_memory_low_water_mark":
				if !hasLowWater || metric.Value < availableLowWater {
					availableLowWater = metric.Value
					hasLowWater = true
				}
			case "vectorized_reactor_uptime":
				if metric.Value > uptimeSeconds {
					uptimeSeconds = metric.Value
				}
			case "vectorized_kafka_handler_latency_microseconds_count":
				handler := metric.Labels["handler"]
				if handler == "produce" {
					produceCount += metric.Value
				} else if handler == "fetch" {
					fetchCount += metric.Value
				}
			case "vectorized_kafka_handler_sent_bytes_total":
				if metric.Labels["handler"] == "fetch" {
					sentBytes += metric.Value
				}
			case "vectorized_kafka_handler_received_bytes_total":
				if metric.Labels["handler"] == "produce" {
					receivedBytes += metric.Value
				}
			case "vectorized_storage_log_batches_written":
				produceBatches += metric.Value
			case "vectorized_storage_log_written_bytes":
				writtenBytes += metric.Value
			case "vectorized_io_queue_queue_length":
				if metric.Value > maxDiskQueue {
					maxDiskQueue = metric.Value
				}
			case "vectorized_io_queue_total_delay_sec":
				ioDelaySec += metric.Value
			case "vectorized_io_queue_total_exec_sec":
				ioExecSec += metric.Value
			case "vectorized_io_queue_total_operations":
				ioOps += metric.Value
			case "vectorized_batch_cache_hits":
				cacheHits += metric.Value
			case "vectorized_batch_cache_accesses":
				cacheAccesses += metric.Value
			case "vectorized_stall_detector_reported":
				reportedStalls += metric.Value
			case "vectorized_scheduler_time_spent_on_task_quota_violations_ms":
				taskQuotaViolationsMs += metric.Value
			case "vectorized_network_bytes_received":
				networkRx += metric.Value
			case "vectorized_network_bytes_sent":
				networkTx += metric.Value
			case "vectorized_host_snmp_tcp_established":
				// This is a Gauge, so adding them up over time is wrong. We want the MAX seen.
				if metric.Value > tcpConnections {
					tcpConnections = metric.Value
				}
			case "vectorized_raft_leadership_changes":
				leadershipChanges += metric.Value
			case "vectorized_raft_recovery_partitions_to_recover":
				// This is a Gauge, generally we want max or current. Summing over time isn't quite right but if it's > 0 it matters.
				// Let's take the max value seen.
				if metric.Value > partitionsToRecover {
					partitionsToRecover = metric.Value
				}
			case "vectorized_raft_replicate_ack_all_requests":
				ackAllRequests += metric.Value
			case "vectorized_raft_replicate_ack_leader_requests":
				ackLeaderRequests += metric.Value
			}
		}
	}

	// Classify Workload
	if hasSchedulerMetrics && (produceRuntime+fetchRuntime > 0) {
		ratio := fetchRuntime / (produceRuntime + fetchRuntime)
		if ratio > 0.7 {
			report.WorkloadType = "Fetch-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Fetch-heavy (based on scheduler runtime).")
		} else if ratio < 0.3 {
			report.WorkloadType = "Produce-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Produce-heavy (based on scheduler runtime).")
		} else {
			report.WorkloadType = "Balanced"
		}
	} else if produceCount+fetchCount > 0 {
		// Fallback to request counts
		ratio := fetchCount / (produceCount + fetchCount)
		if ratio > 0.7 {
			report.WorkloadType = "Fetch-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Fetch-heavy (based on request count).")
		} else if ratio < 0.3 {
			report.WorkloadType = "Produce-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Produce-heavy (based on request count).")
		} else {
			report.WorkloadType = "Balanced"
		}
	} else {
		report.WorkloadType = "Idle"
	}

	// 2.5 Reactor Spinning Detection (WorldQuant case)
	if maxReactorUtil >= 100 && hasSchedulerMetrics && (produceRuntime+fetchRuntime < 100) {
		report.Observations = append(report.Observations, "🚩 Potential Reactor Spinning Detected: Shard shows 100% utilization but very low scheduler runtime. This matches a known Seastar bug on certain OS/Environment combinations (e.g. RHEL8 + overprovisioned).")
	}

	// 3. Small Fetch Detection
	if fetchCount > 0 {
		avgFetchSize := sentBytes / fetchCount
		if avgFetchSize < 500 { // < 500 bytes is very small
			report.Observations = append(report.Observations, fmt.Sprintf("Warning: 'Empty' Fetch Polls detected. Average response is only %.0f bytes.", avgFetchSize))
			report.Recommendations = append(report.Recommendations, "Consumers are polling too aggressively. Increase 'fetch_reads_debounce_timeout' in Redpanda or 'fetch.min.bytes' on clients.")
		}
	}

	// 4. Small Batch Detection (Produce)
	if produceBatches > 0 {
		avgBatchSize := writtenBytes / produceBatches
		report.Observations = append(report.Observations, fmt.Sprintf("Average Batch Size: %.2f KB", avgBatchSize/1024))
		
		// Heuristic: O_DIRECT 4KB Tax
		if avgBatchSize < 4096 {
			efficiency := (avgBatchSize / 4096) * 100
			report.Observations = append(report.Observations, fmt.Sprintf("Warning: Disk IOPS Inflation. Batches are smaller than 4KB block size (Efficiency: %.1f%%). Every write is rounded up to 4KB by the OS.", efficiency))
			report.Recommendations = append(report.Recommendations, "Increase 'linger.ms' or 'batch.size' on your producers to improve disk efficiency.")
		} else if avgBatchSize < 10*1024 { // < 10KB
			report.Observations = append(report.Observations, "Small average batch size detected (< 10KB). High fixed CPU cost.")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "CPU overhead from small batches. Increase 'linger.ms' on clients.")
			}
		}
	}

	// 5. Disk Analysis (Queue + Latency + Seastar Throttling)
	avgDiskExec := 0.0
	avgSeastarDelay := 0.0
	if ioOps > 0 {
		avgDiskExec = (ioExecSec / ioOps) * 1000 // ms
		avgSeastarDelay = (ioDelaySec / ioOps) * 1000 // ms
	}

	if maxDiskQueue > 50 {
		report.Observations = append(report.Observations, fmt.Sprintf("High Disk Queue Length: %.0f.", maxDiskQueue))
		
		if avgDiskExec > 20.0 {
			report.IsDiskBound = true
			report.Observations = append(report.Observations, fmt.Sprintf("Critical: Slow Hardware/Volume. Disk execution time is high (%.2f ms/op).", avgDiskExec))
			report.Recommendations = append(report.Recommendations, "Physical disk latency is high. Check for EBS throttling, RAID issues, or slow hardware.")
		} else if avgSeastarDelay > avgDiskExec * 2 {
			report.Observations = append(report.Observations, fmt.Sprintf("Warning: Seastar Throttling. Seastar is holding IO in userspace for %.2f ms (vs %.2f ms disk time).", avgSeastarDelay, avgDiskExec))
			report.Recommendations = append(report.Recommendations, "The Seastar IO scheduler is limiting throughput. This usually means 'iotune' values are too conservative for the hardware.")
		}
	}

	// 6. Cache Efficiency
	if cacheAccesses > 0 {
		hitRatio := cacheHits / cacheAccesses
		if hitRatio < 0.8 && report.WorkloadType == "Fetch-Heavy" {
			report.Observations = append(report.Observations, fmt.Sprintf("Warning: Low Batch Cache Hit Ratio (%.1f%%). Readers are hitting disk.", hitRatio*100))
			report.Recommendations = append(report.Recommendations, "Consumers are reading historical data (stale reads). This stresses the disk and increases latency. Consider increasing 'batch_cache_max_memory'.")
		}
	}

	// 7. Stall Detection and Background CPU
	if hasSchedulerMetrics {
		// archivalUploadRuntime, compactionRuntime? (Need to add compaction to gathering)
		// For now, check if archival is significant
		if produceRuntime > 0 && archivalUploadRuntime > (produceRuntime * 0.5) {
			report.Observations = append(report.Observations, "Warning: Background Task Pressure. Tiered Storage uploads are consuming > 50% of the produce CPU time.")
		}
	}

	if reactorStallsHistogram > 0 {
		report.Observations = append(report.Observations, "Reactor Stall Histogram populated: Stalls > 1ms detected.")
	}
	
	if threadedFallbacks > 0 {
		report.Observations = append(report.Observations, fmt.Sprintf("Threaded Fallbacks: %.0f syscalls forced to fallback thread. May cause latency spikes not seen in disk queue metrics.", threadedFallbacks))
	}

	if taskQuotaViolationsMs > 1000 { // > 1 second total violation
		report.Observations = append(report.Observations, fmt.Sprintf("Task Quota Violations: %.0f ms spent in violation. Short stalls are frequent.", taskQuotaViolationsMs))
	}

	// 8. Network Analysis
	// Fanout Ratio
	if networkRx > 0 {
		fanoutRatio := networkTx / networkRx
		if fanoutRatio > 5.0 {
			report.IsNetworkBound = true
			report.Observations = append(report.Observations, fmt.Sprintf("Critical: Extreme Network Fan-out (%.1fx). You are sending 5x more data than you receive.", fanoutRatio))
			report.Recommendations = append(report.Recommendations, "High consumer fan-out stresses cloud network limits (the 'haircut'). Consider using private VPC endpoints or reducing consumer counts.")
		} else if fanoutRatio > 3.0 {
			report.Observations = append(report.Observations, fmt.Sprintf("High Network Fanout: %.1fx more TX than RX. High consumer load.", fanoutRatio))
			if report.IsNetworkBound {
				report.Recommendations = append(report.Recommendations, "High fanout stresses network. Ensure adequate bandwidth.")
			}
		}
	}

	// Tiered Storage Stress
	if archivalUploadRuntime > 0 {
		// Just noting it exists is useful context
		report.Observations = append(report.Observations, "Tiered Storage Uploads Active: Network is used for archival.")
		// If upload runtime is significant compared to produce, it's a factor
		if produceRuntime > 0 && archivalUploadRuntime > (produceRuntime * 0.5) {
			report.Observations = append(report.Observations, "Heavy Tiered Storage Activity: Uploads consuming significant scheduler time.")
			report.Recommendations = append(report.Recommendations, "Tiered Storage uploads compete with public consumers for network egress. Monitor cloud provider bandwidth limits.")
		}
	}

	// Connections
	if tcpConnections > 50000 {
		report.Observations = append(report.Observations, fmt.Sprintf("High Connection Count: %.0f TCP connections. Monitor fd limits.", tcpConnections))
	}

	// 9. Raft Analysis
	// Ack Level Analysis
	if ackAllRequests+ackLeaderRequests > 0 {
		ackAllRatio := ackAllRequests / (ackAllRequests + ackLeaderRequests)
		if ackAllRatio > 0.5 {
			report.Observations = append(report.Observations, "Heavy Durable Writes: >50% of produce requests use acks=all (fsync on majority). This increases latency.")
		} else {
			report.Observations = append(report.Observations, "High Performance Writes: Predominantly using acks=1/0 (leader ack/no ack).")
		}
	}

	// Leadership Stability
	// We don't have a rate here, just a total count since startup (usually). 
	// If it's very high relative to runtime it's bad, but we don't know runtime easily here.
	// But if it's > 1000 it's worth noting.
	if leadershipChanges > 1000 {
		report.Observations = append(report.Observations, fmt.Sprintf("Frequent Leadership Changes: %.0f changes detected. Check for network stability or overloaded nodes.", leadershipChanges))
	}

	// Recovery
	if partitionsToRecover > 0 {
		report.Observations = append(report.Observations, fmt.Sprintf("Cluster in Recovery: %0.f partitions need recovery. Background traffic will be higher.", partitionsToRecover))
	}

	// 10. Memory Analysis (Healthy Zero & Fragmentation)
	if totalMem > 0 {
		effectiveFree := freeMem + reclaimableMem
		freeRatio := effectiveFree / totalMem

		if freeRatio < 0.05 {
			report.Observations = append(report.Observations, fmt.Sprintf("Critical: Low Effective Memory. Only %.1f%% of memory is available (including reclaimable cache). Risk of OOM.", freeRatio*100))
			report.Recommendations = append(report.Recommendations, "Increase node memory or reduce shard count to give more head-room to each core.")
		} else if freeMem < (128 * 1024 * 1024) { // Less than 128MB raw free
			if reclaimableMem > (totalMem * 0.2) {
				report.Observations = append(report.Observations, "Healthy Memory State: Raw free memory is low, but >20% is reclaimable from the Batch Cache (Normal Redpanda behavior).")
			} else {
				report.Observations = append(report.Observations, "Warning: Low Free Memory. Redpanda is operating with very little memory head-room.")
			}
		}

		// Fragmentation heuristic
		const oneWeek = 7 * 24 * 60 * 60
		if uptimeSeconds > oneWeek && freeRatio > 0.2 && freeMem < (256 * 1024 * 1024) {
			report.Observations = append(report.Observations, "ℹ️ High Fragmentation Risk: This node has been up for >7 days. Seastar memory may be fragmented, which can cause large allocations to fail even with high reclaimable memory.")
			report.Recommendations = append(report.Recommendations, "If you see 'bad_alloc' errors despite high free memory, consider a rolling restart to defragment the heap.")
		}

		if hasLowWater && availableLowWater < (64 * 1024 * 1024) {
			report.Observations = append(report.Observations, "🚩 Historical Memory Pressure: The node's available memory low water mark reached near-zero. This confirms a past memory spike that almost caused an OOM.")
		}
	}

	// 11. Hardware Validation (The 4/30 Rule)
	for _, nr := range selfTest {
		for _, r := range nr.Results {
			if r.Type != "disk" {
				continue
			}
			
			// 4MB/s small block check
			if strings.Contains(strings.ToLower(r.Name), "4kb") {
				throughputMB := float64(r.Throughput) / (1024 * 1024)
				if throughputMB < 4.0 {
					report.Observations = append(report.Observations, fmt.Sprintf("❌ Unsupported Hardware: Small block write throughput (%.2f MB/s) is below the minimum threshold (4.0 MB/s).", throughputMB))
					report.Recommendations = append(report.Recommendations, "Physical disk performance is insufficient for Redpanda production support. Upgrade your storage volume.")
				}
			}
			
			// 30MB/s medium block check
			if strings.Contains(strings.ToLower(r.Name), "16kb") {
				throughputMB := float64(r.Throughput) / (1024 * 1024)
				if throughputMB < 30.0 {
					report.Observations = append(report.Observations, fmt.Sprintf("❌ Unsupported Hardware: Medium block write throughput (%.2f MB/s) is below the minimum threshold (30.0 MB/s).", throughputMB))
					report.Recommendations = append(report.Recommendations, "Physical disk performance is insufficient for Redpanda production support. Upgrade your storage volume.")
				}
			}
		}
	}

	// 12. Demand vs Capacity Analysis (The "Upstream" Check)
	if !report.IsCPUBound && !report.IsDiskBound && !report.IsNetworkBound {
		// If cluster is healthy and utilization is low, check if it's effectively idle
		if maxReactorUtil < 20.0 && produceCount > 0 {
			report.Observations = append(report.Observations, "Cluster Headroom: Redpanda is under-utilized (< 20% CPU). The current throughput appears to be limited by Application Demand, not the cluster.")
			report.Recommendations = append(report.Recommendations, "If throughput is lower than expected, check producer-side metrics. The bottleneck is likely upstream or in the client's serial send-loop.")
		}
	}

	return report
}
