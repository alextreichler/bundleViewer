package analysis

import (
	"fmt"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// PerformanceReport holds the result of the performance analysis
type PerformanceReport struct {
	IsCPUBound       bool
	IsDiskBound      bool
	IsNetworkBound   bool // Hard to determine from snapshot, but we can look for errors
	WorkloadType     string // "Produce-Heavy", "Fetch-Heavy", "Balanced", "Idle"
	Recommendations  []string
	Observations     []string
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
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_reactor_utilization" {
				if metric.Value > maxReactorUtil {
					maxReactorUtil = metric.Value
				}
			}
		}
	}

	if maxReactorUtil > 95 {
		report.IsCPUBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("High Reactor Utilization detected: %.2f%%. The system is likely CPU bound.", maxReactorUtil))
	}

	// 2. Workload Classification (Produce vs Fetch)
	// We need 'vectorized_kafka_handler_latency_microseconds_count' which is a counter. 
	// In a snapshot, we can't see rate, but we can see TOTAL counts to see what the dominant historical workload has been.
	var produceCount, fetchCount float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_kafka_handler_latency_microseconds_count" {
				handler := metric.Labels["handler"]
				if handler == "produce" {
					produceCount += metric.Value
				} else if handler == "fetch" {
					fetchCount += metric.Value
				}
			}
		}
	}

	if produceCount+fetchCount > 0 {
		ratio := fetchCount / (produceCount + fetchCount)
		if ratio > 0.7 {
			report.WorkloadType = "Fetch-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Fetch-heavy (>70% of requests).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Consider increasing 'fetch_reads_debounce_timeout' to reduce CPU load from excessive fetching.")
			}
		} else if ratio < 0.3 {
			report.WorkloadType = "Produce-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Produce-heavy (>70% of requests).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Consider increasing 'linger.ms' on clients to improve batching.")
			}
		} else {
			report.WorkloadType = "Balanced"
		}
	} else {
		report.WorkloadType = "Idle"
	}

	// 3. Disk Analysis
	// Check for IO Queue delays (Snapshot of counter is useless, but Queue Length is a Gauge)
	maxQueueLength := 0.0
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_io_queue_queue_length" {
				if metric.Value > maxQueueLength {
					maxQueueLength = metric.Value
				}
			}
		}
	}

	if maxQueueLength > 50 {
		report.IsDiskBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("High Disk Queue Length detected: %.0f. Disk may be saturating.", maxQueueLength))
	}

	// 4. Batch Size Efficiency (Heuristic based on totals)
	// vectorized_storage_log_written_bytes / vectorized_storage_log_batches_written
	var totalBytes, totalBatches float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_storage_log_written_bytes" {
				totalBytes += metric.Value
			}
			if metric.Name == "vectorized_storage_log_batches_written" {
				totalBatches += metric.Value
			}
		}
	}

	if totalBatches > 0 {
		avgBatchSize := totalBytes / totalBatches
		report.Observations = append(report.Observations, fmt.Sprintf("Average Batch Size: %.2f KB", avgBatchSize/1024))
		if avgBatchSize < 10*1024 { // < 10KB
			report.Observations = append(report.Observations, "Small average batch size detected (< 10KB).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Small batches increase CPU overhead. Increase batch size on producers.")
			}
		}
	}

	return report
}
