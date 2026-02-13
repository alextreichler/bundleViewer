package analysis

import (
	"fmt"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

func AnalyzeReplication(
	p models.ClusterPartition,
	debugMap map[string]map[int]models.PartitionDebug,
) *models.PartitionInsight {
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
