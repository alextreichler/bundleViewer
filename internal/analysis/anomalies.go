package analysis

import (
	"fmt"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// DetectAnomalies scans the cluster data for critical issues described in the Blue Book.
func DetectAnomalies(
	partitionDebug []models.PartitionDebug,
	kafkaMetadata models.KafkaMetadataResponse,
	healthOverview models.HealthOverview,
	patterns []models.LogPattern,
) []models.Anomaly {
	var anomalies []models.Anomaly

	// 1. Raft Index Anomaly (Ghost/Negative Indices)
	for _, pd := range partitionDebug {
		for _, r := range pd.Replicas {
			for _, f := range r.RaftState.Followers {
				// The troubleshooting guide mentions -9223372036854776000 as abnormal
				if f.LastFlushedLogIndex < -1000000 {
					anomalies = append(anomalies, models.Anomaly{
						Type:     "Raft Index",
						Severity: "Critical",
						Message:  fmt.Sprintf("Partition %s/%s/%d: Node %d reports abnormal log index (%d). Likely uninitialized or corrupt Raft state.", pd.Ns, pd.Topic, pd.ID, f.ID, f.LastFlushedLogIndex),
						Link:     fmt.Sprintf("/partitions?topic=%s", pd.Topic),
					})
				}
			}
		}
	}

	// 2. Leadership Discrepancy Check
	kafkaLeaderlessCount := 0
	for _, topic := range kafkaMetadata.Topics {
		for _, part := range topic.Partitions {
			if part.Leader == -1 {
				kafkaLeaderlessCount++
			}
		}
	}
	healthLeaderless := healthOverview.LeaderlessCount

	if healthLeaderless > 0 && kafkaLeaderlessCount == 0 {
		anomalies = append(anomalies, models.Anomaly{
			Type:     "Stale Health",
			Severity: "Warning",
			Message:  fmt.Sprintf("Global Discrepancy: Health monitor reports %d leaderless partitions, but Kafka metadata shows all have leaders. reporting node state is likely stale.", healthLeaderless),
			Link:     "/partitions",
		})
	}

	// 3. Log Pattern Frequency (Ghost Batches, Reactor Stalls)
	for _, p := range patterns {
		lowerMsg := strings.ToLower(p.SampleEntry.Message)
		if strings.Contains(lowerMsg, "ghost_record_batch") && p.Count > 10 {
			anomalies = append(anomalies, models.Anomaly{
				Type:     "Log Pattern",
				Severity: "Warning",
				Message:  fmt.Sprintf("High frequency of 'Ghost Record Batches' detected (%d occurrences). This indicates gaps in the log being recovered.", p.Count),
				Link:     "/logs?search=ghost_record_batch",
			})
		}
		if strings.Contains(lowerMsg, "reactor_stalled") && p.Count > 5 {
			anomalies = append(anomalies, models.Anomaly{
				Type:     "Performance",
				Severity: "Warning",
				Message:  fmt.Sprintf("Multiple 'Reactor Stalls' detected (%d occurrences). This causes high latency and can lead to Raft timeouts.", p.Count),
				Link:     "/logs?search=reactor_stalled",
			})
		}
	}

	return anomalies
}
