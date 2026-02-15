package analysis

import (
	"fmt"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

func AnalyzeReplication(
	p models.ClusterPartition,
	debugMap map[string]map[int]models.PartitionDebug,
	kafkaLeader int,
) *models.PartitionInsight {
	// 1. Discrepancy Check (authoritative Kafka Metadata vs Redpanda internal health)
	if p.LeaderID != kafkaLeader && kafkaLeader != -2 { // -2 is our "not found" sentinel
		return &models.PartitionInsight{
			Summary:     fmt.Sprintf("LEADERSHIP DISCREPANCY: Redpanda internal says leader is %d, but Kafka metadata says %d. This node likely has stale health information.", p.LeaderID, kafkaLeader),
			Severity:    "warn",
			Remediation: "If Redpanda reports this as leaderless but Kafka shows a leader, the reporting node may be stale. Check if the node is under heavy load or has reactor stalls.",
		}
	}

	// 2. Detect Leaderless Partitions (Critical)
	if p.LeaderID == -1 {
		insight := &models.PartitionInsight{
			Summary:  "Partition is LEADERLESS. No node is currently acting as the leader.",
			Severity: "error",
		}

		// Try to find clues in debug info
		if tMap, ok := debugMap[p.Topic]; ok {
			if d, ok := tMap[p.PartitionID]; ok {
				if len(d.Replicas) < len(p.Replicas) {
					insight.Summary += fmt.Sprintf(" Only %d/%d replicas reported debug info.", len(d.Replicas), len(p.Replicas))
				}

				// Check if any node thinks it's a leader but isn't recognized
				for _, r := range d.Replicas {
					if r.RaftState.IsLeader || r.RaftState.IsElectedLeader {
						insight.Summary += fmt.Sprintf(" Node %d thinks it is the leader, but the cluster disagrees.", r.RaftState.NodeID)
					}
					
					// Check for abnormal uninitialized indices (MinInt64-like)
					for _, f := range r.RaftState.Followers {
						if f.LastFlushedLogIndex < -1000000 {
							insight.Summary += fmt.Sprintf(" Abnormal log index (%d) detected for follower %d.", f.LastFlushedLogIndex, f.ID)
						}
					}
				}
			}
		}

		if insight.Remediation == "" {
			ntp := fmt.Sprintf("%s/%s/%d", p.Ns, p.Topic, p.PartitionID)
			insight.Remediation = fmt.Sprintf("View leadership history: <button onclick=\"showRaftTimeline('%s')\" class='button' style='padding:2px 8px; font-size:0.7rem;'>📜 Raft History</button>", ntp)
		}
		return insight
	}

	// 3. Detect Reconfiguration (Joint Consensus)
	if strings.Contains(strings.ToLower(p.Status), "reconfig") {
		return &models.PartitionInsight{
			Summary:     "Partition is undergoing RECONFIGURATION (Joint Consensus). replicas are being moved or changed.",
			Severity:    "info",
			Remediation: "This is a normal part of partition movement. If it is stuck, check for disk space issues or network connectivity between the old and new replica sets.",
		}
	}

	if p.Status != "under_replicated" {
		return nil
	}

	insight := &models.PartitionInsight{
		Severity: "warn",
	}

	// Check debug info for specific lag
	if tMap, ok := debugMap[p.Topic]; ok {
		if d, ok := tMap[p.PartitionID]; ok {
			var maxOffset int64
			var minOffset int64 = -1
			var laggingNode int = -1

			for _, r := range d.Replicas {
				if r.CommittedOffset > maxOffset {
					maxOffset = r.CommittedOffset
				}
				if minOffset == -1 || r.CommittedOffset < minOffset {
					minOffset = r.CommittedOffset
					laggingNode = r.RaftState.NodeID
				}
			}

			if maxOffset > minOffset && laggingNode != -1 {
				lag := maxOffset - minOffset
				insight.Summary = fmt.Sprintf("Node %d is lagging by %d offsets.", laggingNode, lag)

				// remediation
				var otherReplicas []string
				for _, r := range p.Replicas {
					if r.NodeID != laggingNode {
						otherReplicas = append(otherReplicas, fmt.Sprintf("%d", r.NodeID))
					}
				}
				if len(otherReplicas) > 0 {
					insight.Remediation = fmt.Sprintf("rpk cluster partitions move -p %s/%d:%s", p.Topic, p.PartitionID, strings.Join(otherReplicas, ","))
				}
			} else {
				// Check heartbeats
				for _, r := range d.Replicas {
					for _, f := range r.RaftState.Followers {
						if f.MsSinceLastHeartbeat > 10000 { // 10s
							insight.Summary = fmt.Sprintf("Node %d hasn't sent a heartbeat for %ds.", f.ID, f.MsSinceLastHeartbeat/1000)
							insight.Severity = "error"
							break
						}
					}
				}
			}
		}
	}

	if insight.Summary == "" {
		insight.Summary = "Partition is under-replicated. Replicas are catching up."
	}

	return insight
}
