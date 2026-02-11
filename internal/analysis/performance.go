package analysis

import (
	"fmt"

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
}

// AnalyzePerformance evaluates the metrics bundle against known performance heuristics
func AnalyzePerformance(metrics *models.MetricsBundle) PerformanceReport {
	report := PerformanceReport{
		Recommendations: []string{},
		Observations:    []string{},
	}

	if metrics == nil {
		return report
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

	// 3. Small Fetch Detection
	if fetchCount > 0 {
		avgFetchSize := sentBytes / fetchCount
		if avgFetchSize < 500 { // < 500 bytes is very small (likely empty or just headers)
			report.Observations = append(report.Observations, fmt.Sprintf("Small Fetch Responses detected: Average %.0f bytes. Inefficient polling.", avgFetchSize))
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "High CPU due to excessive fetch polling. Increase 'fetch_reads_debounce_timeout'.")
			}
		}
	}

	// 4. Small Batch Detection (Produce)
	if produceBatches > 0 {
		avgBatchSize := writtenBytes / produceBatches
		report.Observations = append(report.Observations, fmt.Sprintf("Average Batch Size: %.2f KB", avgBatchSize/1024))
		if avgBatchSize < 10*1024 { // < 10KB
			report.Observations = append(report.Observations, "Small average batch size detected (< 10KB). High fixed CPU cost.")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "CPU overhead from small batches. Increase 'linger.ms' on clients.")
			}
		}
	}

	// 5. Disk Analysis (Queue + Latency)
	avgDiskLatency := 0.0
	if ioOps > 0 {
		avgDiskLatency = (ioDelaySec + ioExecSec) / ioOps
	}

	if maxDiskQueue > 50 {
		report.Observations = append(report.Observations, fmt.Sprintf("High Disk Queue Length: %.0f.", maxDiskQueue))
		if avgDiskLatency > 0.020 { // > 20ms
			report.IsDiskBound = true
			report.Observations = append(report.Observations, fmt.Sprintf("Critical: Disk Saturation. High Queue + High Latency (%.2f ms).", avgDiskLatency*1000))
			report.Recommendations = append(report.Recommendations, "Disk cannot keep up. Check IOPS limits or hardware.")
		} else {
			report.Observations = append(report.Observations, "Disk queue is high but latency is low. CPU might be the bottleneck (polling less frequently).")
		}
	}

	// 6. Cache Efficiency
	if cacheAccesses > 0 {
		hitRatio := cacheHits / cacheAccesses
		if hitRatio < 0.8 && report.WorkloadType == "Fetch-Heavy" {
			report.Observations = append(report.Observations, fmt.Sprintf("Low Batch Cache Hit Ratio: %.2f. Reads are hitting disk.", hitRatio))
			if report.IsDiskBound {
				report.Recommendations = append(report.Recommendations, "Disk load due to cache misses. Consumers are reading historical data.")
			}
		}
	}

	// 7. Stall Detection
	if reportedStalls > 0 {
		report.Observations = append(report.Observations, fmt.Sprintf("Reactor Stalls Detected: %.0f reported long stalls (>25ms).", reportedStalls))
		if report.IsCPUBound {
			report.Recommendations = append(report.Recommendations, "Stalls indicate tasks are hogging the reactor. Check for 'Reactor stalled' logs.")
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
		if fanoutRatio > 3.0 {
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

	return report
}
