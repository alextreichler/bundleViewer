package analysis

import (
	"fmt"
	"math"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NodeSkew struct {
	NodeID         int
	PartitionCount int
	TopicCount     int
	CoreCount      int     // Number of CPU cores
	SMPCount       int     // Number of cores configured via --smp (0 if not set)
	Deviation      float64 // Standard deviations from mean
	SkewPercentage float64 // % difference from mean
	IsSkewed       bool
}

type TopicSkew struct {
	TopicName        string
	PartitionCount   int
	NodeDistribution map[int]int
	MaxSkew          float64 // Max deviation found among nodes for this topic
	IsSkewed         bool
}

type RackSkew struct {
	RackID           string
	NodeCount        int
	PartitionCount   int
	NodeDistribution map[int]int
	Skewed           bool
	Nodes            []int
}

type ConstraintCheck struct {
	NodeID          int
	IsValidTarget   bool
	BlockerReasons  []string
	DiskUsagePct    float64
	IsAlive         bool
	IsDraining      bool
}

type BalancingConfig struct {
	PartitionsPerShard      int  `json:"partitions_per_shard"`
	PartitionsReserveShard0 int  `json:"partitions_reserve_shard0"`
	TopicAware              bool `json:"topic_aware"`
	RackAwarenessEnabled    bool `json:"rack_awareness_enabled"`
}

type NodeBalancingScore struct {
	NodeID      int
	MaxCapacity int
	Score       float64
	CPU         int
}

type DiskBalancingConfig struct {
	SpaceManagementEnabled                    bool
	DiskReservationPercent                    int
	RetentionLocalTargetCapacityPercent       int
	RetentionLocalTargetCapacityBytes         int64
	PartitionAutobalancingMaxDiskUsagePercent int
	StorageSpaceAlertFreeThresholdPercent     int
}

type DiskAnalysis struct {
	TotalBytes            int64
	FreeBytes             int64
	UsedBytes             int64
	ReservedBytes         int64
	UsableBytes           int64
	TargetBytes           int64
	RebalanceThreshold    int64 // When rebalancing kicks in
	AlertThreshold        int64 // When space alerts trigger
	AvailableToRebalance  int64 // Headroom until Rebalance Threshold
	AvailableToAlert      int64 // Headroom until Alert Threshold (Hard Max)
	IsFull                bool
	IsRebalancingTriggered bool
}

type BalancingAnalysis struct {
	Config      BalancingConfig
	DiskConfig  DiskBalancingConfig
	Nodes       []NodeBalancingScore
	NodeDiskMap map[int]DiskAnalysis // Map NodeID -> DiskAnalysis
}

type HotTopic struct {
	TopicName  string
	BytesValue float64 // Could be bytes/sec or total bytes depending on input
	Deviation  float64 // Std Devs above mean
	IsHot      bool
}

type SkewAnalysis struct {
	TotalPartitions  int
	TotalNodes       int
	MeanPartitions   float64
	StdDev           float64
	NodeSkews        []NodeSkew
	TopicSkews       []TopicSkew
	RackSkews        []RackSkew
	HotTopics        []HotTopic // NEW: Throughput outliers
	ConstraintChecks []ConstraintCheck
	BalancerStatus   models.PartitionBalancerStatus
	Balancing        BalancingAnalysis
	IsClusterSkewed  bool
	ClusterSkewHint  string
}

func getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func getBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getInt64(config map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := config[key].(float64); ok {
		return int64(val)
	}
	if val, ok := config[key].(int); ok {
		return int64(val)
	}
	return defaultValue
}

func AnalyzeBalancing(
	brokers []interface{},
	nodePartitionCounts map[int]int,
	nodeTopicCounts map[string]map[int]int,
	clusterConfig map[string]interface{},
	diskStats map[int]models.DiskStats,
) BalancingAnalysis {
	config := BalancingConfig{
		PartitionsPerShard:      getInt(clusterConfig, "topic_partitions_per_shard", 5000),
		PartitionsReserveShard0: getInt(clusterConfig, "topic_partitions_reserve_shard0", 0),
		TopicAware:              clusterConfig["partition_autobalancing_mode"] == "topic_aware",
		RackAwarenessEnabled:    getBool(clusterConfig, "enable_rack_awareness", false),
	}

	diskConfig := DiskBalancingConfig{
		SpaceManagementEnabled:                    getBool(clusterConfig, "space_management_enabled", true),
		DiskReservationPercent:                    getInt(clusterConfig, "disk_reservation_percent", 20),                // Defaulting to 20%
		RetentionLocalTargetCapacityPercent:       getInt(clusterConfig, "retention_local_target_capacity_percent", 80), // Defaulting to 80%
		RetentionLocalTargetCapacityBytes:         getInt64(clusterConfig, "retention_local_target_capacity_bytes", 0),
		PartitionAutobalancingMaxDiskUsagePercent: getInt(clusterConfig, "partition_autobalancing_max_disk_usage_percent", 80), // Defaulting to 80%
		StorageSpaceAlertFreeThresholdPercent:     getInt(clusterConfig, "storage_space_alert_free_threshold_percent", 15),     // Defaulting to 15%
	}

	var nodeScores []NodeBalancingScore
	nodeDiskAnalysis := make(map[int]DiskAnalysis)

	for _, b := range brokers {
		brokerMap, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		nodeID := int(brokerMap["node_id"].(float64))
		cores := int(brokerMap["num_cores"].(float64))

		maxCapacity := (cores * config.PartitionsPerShard) - config.PartitionsReserveShard0
		if maxCapacity <= 0 {
			maxCapacity = 1 // Avoid division by zero
		}

		var score float64
		if config.TopicAware {
			// Count unique topics on the node
			topicCount := 0
			for _, nodeMap := range nodeTopicCounts {
				if _, exists := nodeMap[nodeID]; exists {
					topicCount++
				}
			}
			score = float64(topicCount) / float64(maxCapacity)
		} else {
			score = float64(nodePartitionCounts[nodeID]) / float64(maxCapacity)
		}

		nodeScores = append(nodeScores, NodeBalancingScore{
			NodeID:      nodeID,
			MaxCapacity: maxCapacity,
			Score:       score,
			CPU:         cores,
		})

		// Disk Analysis Logic
		if stats, hasStats := diskStats[nodeID]; hasStats {
			totalBytes := stats.TotalBytes
			freeBytes := stats.FreeBytes
			usedBytes := totalBytes - freeBytes
			
			var reservedBytes, usableBytes, targetBytes int64
			var rebalanceThreshold, alertThreshold int64
			
			if diskConfig.SpaceManagementEnabled {
				// 1. Calculate Reservation
				reservedBytes = int64(float64(totalBytes) * (float64(diskConfig.DiskReservationPercent) / 100.0))
				usableBytes = totalBytes - reservedBytes
				if usableBytes < 0 { usableBytes = 0 }

				// 2. Calculate Target Size (Retention Limit)
				targetPctBytes := int64(float64(usableBytes) * (float64(diskConfig.RetentionLocalTargetCapacityPercent) / 100.0))
				targetFixedBytes := diskConfig.RetentionLocalTargetCapacityBytes
				
				if targetPctBytes > 0 && targetFixedBytes == 0 {
					targetBytes = targetPctBytes
				} else if targetPctBytes == 0 && targetFixedBytes > 0 {
					targetBytes = targetFixedBytes
				} else if targetPctBytes > 0 && targetFixedBytes > 0 {
					if targetPctBytes < targetFixedBytes {
						targetBytes = targetPctBytes
					} else {
						targetBytes = targetFixedBytes
					}
				} else {
					targetBytes = 0 // Disabled
				}
			} else {
				// If management disabled, treat effectively whole disk as target (minus alerts)
				targetBytes = totalBytes
				usableBytes = totalBytes
			}

			// 3. Rebalance Threshold (partition_autobalancing_max_disk_usage_percent)
			// This applies to the *total* disk usually, or sometimes usable. 
			// Based on the notes: "80% of 90G = ~ 73GB" where 90G was the target.
			// It implies rebalance triggers at (MaxDiskUsage% * TargetSize).
			// Let's assume it checks against Target Bytes if enabled.
			if targetBytes > 0 {
				rebalanceThreshold = int64(float64(targetBytes) * (float64(diskConfig.PartitionAutobalancingMaxDiskUsagePercent) / 100.0))
			} else {
				// Fallback to percentage of total if target not set? 
				// Logic says soft_max_disk_usage_ratio = _max_disk_usage_percent() / 100.0
				rebalanceThreshold = int64(float64(totalBytes) * (float64(diskConfig.PartitionAutobalancingMaxDiskUsagePercent) / 100.0))
			}

			// 4. Alert Threshold (storage_space_alert_free_threshold_percent)
			// This is based on Total disk free space.
			alertThreshold = int64(float64(totalBytes) * (float64(diskConfig.StorageSpaceAlertFreeThresholdPercent) / 100.0))
			
			nodeDiskAnalysis[nodeID] = DiskAnalysis{
				TotalBytes:            totalBytes,
				FreeBytes:             freeBytes,
				UsedBytes:             usedBytes,
				ReservedBytes:         reservedBytes,
				UsableBytes:           usableBytes,
				TargetBytes:           targetBytes,
				RebalanceThreshold:    rebalanceThreshold,
				AlertThreshold:        alertThreshold,
				AvailableToRebalance:  rebalanceThreshold - usedBytes,
				AvailableToAlert:      (totalBytes - alertThreshold) - usedBytes,
				IsFull:                freeBytes < alertThreshold,
				IsRebalancingTriggered: usedBytes > rebalanceThreshold,
			}
		}
	}

	return BalancingAnalysis{
		Config:      config,
		DiskConfig:  diskConfig,
		Nodes:       nodeScores,
		NodeDiskMap: nodeDiskAnalysis,
	}
}

func AnalyzePartitionSkew(
	partitions []models.ClusterPartition, 
	brokers []interface{}, 
	clusterConfig map[string]interface{},
	diskStats map[int]models.DiskStats,
	balancerStatus models.PartitionBalancerStatus,
	topicThroughput map[string]float64, // NEW param
	smpCount int,
) SkewAnalysis {
	// 1. Count Partitions and Topics per Node
	nodePartitionCounts := make(map[int]int)
	topicPartitionCounts := make(map[string]int)
	// topic -> node -> count
	nodeTopicCounts := make(map[string]map[int]int)
	allNodes := make(map[int]bool)
	rackMap := make(map[string][]int) // RackID -> []NodeID
	nodeRacks := make(map[int]string) // NodeID -> RackID
	
	// Constraint Checking Info
	nodeIsAlive := make(map[int]bool)
	nodeIsDraining := make(map[int]bool)
	nodeCores := make(map[int]int)

	// Initialize all known nodes with 0
	for _, b := range brokers {
		if bMap, ok := b.(map[string]interface{}); ok {
			if id, ok := bMap["node_id"].(float64); ok {
				nodeID := int(id)
				allNodes[nodeID] = true
				nodePartitionCounts[nodeID] = 0
				
				// Rack Info
				rack := "default"
				if r, ok := bMap["rack"].(string); ok && r != "" {
					rack = r
				}
				rackMap[rack] = append(rackMap[rack], nodeID)
				nodeRacks[nodeID] = rack
				
				// Status Info
				if alive, ok := bMap["is_alive"].(bool); ok {
					nodeIsAlive[nodeID] = alive
				} else {
					nodeIsAlive[nodeID] = true // Assume alive if missing field
				}
				
				if status, ok := bMap["maintenance_status"].(map[string]interface{}); ok {
					if draining, ok := status["draining"].(bool); ok {
						nodeIsDraining[nodeID] = draining
					}
				}

				if cores, ok := bMap["num_cores"].(float64); ok {
					nodeCores[nodeID] = int(cores)
				}
			}
		}
	}

	for _, p := range partitions {
		topicPartitionCounts[p.Topic]++
		for _, r := range p.Replicas {
			nodePartitionCounts[r.NodeID]++
			allNodes[r.NodeID] = true

			if _, ok := nodeTopicCounts[p.Topic]; !ok {
				nodeTopicCounts[p.Topic] = make(map[int]int)
			}
			nodeTopicCounts[p.Topic][r.NodeID]++
		}
	}

	totalNodes := len(allNodes)
	if totalNodes == 0 {
		return SkewAnalysis{}
	}

	totalPartitions := 0
	for _, count := range nodePartitionCounts {
		totalPartitions += count
	}

	mean := float64(totalPartitions) / float64(totalNodes)

	// Calculate Variance and StdDev
	var sumSquares float64
	for _, count := range nodePartitionCounts {
		diff := float64(count) - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(totalNodes)
	stdDev := math.Sqrt(variance)

	// Calculate Node Skews
	var nodeSkews []NodeSkew
	isClusterSkewed := false
	
	for nodeID, count := range nodePartitionCounts {
		deviation := 0.0
		if stdDev > 0 {
			deviation = (float64(count) - mean) / stdDev
		}

		skewPct := 0.0
		if mean > 0 {
			skewPct = ((float64(count) - mean) / mean) * 100
		}

		// Skewed if deviation > 1.5 sigma or > 20% deviation
		isSkewed := math.Abs(deviation) > 1.5 || math.Abs(skewPct) > 20.0
		if isSkewed {
			isClusterSkewed = true
		}

		topicCount := 0
		for _, nodeMap := range nodeTopicCounts {
			if _, exists := nodeMap[nodeID]; exists {
				topicCount++
			}
		}

		nodeSkews = append(nodeSkews, NodeSkew{
			NodeID:         nodeID,
			PartitionCount: count,
			TopicCount:     topicCount,
			CoreCount:      nodeCores[nodeID],
			SMPCount:       smpCount,
			Deviation:      deviation,
			SkewPercentage: skewPct,
			IsSkewed:       isSkewed,
		})
	}

	// Analyze Rack Skew (Intra-Rack Balancing)
	var rackSkews []RackSkew
	for rackID, nodes := range rackMap {
		if len(nodes) < 2 {
			continue // No intra-rack balancing possible with 1 node
		}
		
		rackTotalPartitions := 0
		nodeDist := make(map[int]int)
		minCount := math.MaxInt32
		maxCount := 0
		
		for _, nid := range nodes {
			c := nodePartitionCounts[nid]
			rackTotalPartitions += c
			nodeDist[nid] = c
			if c < minCount { minCount = c }
			if c > maxCount { maxCount = c }
		}
		
		// Flag as skewed if max > 2x min AND absolute diff is significant (>100)
		isRackSkewed := (maxCount > minCount*2) && (maxCount-minCount > 100)
		
		rackSkews = append(rackSkews, RackSkew{
			RackID:           rackID,
			NodeCount:        len(nodes),
			PartitionCount:   rackTotalPartitions,
			NodeDistribution: nodeDist,
			Skewed:           isRackSkewed,
			Nodes:            nodes,
		})
		
		if isRackSkewed {
			isClusterSkewed = true
		}
	}
	
	// Analyze Constraints for Under-loaded nodes
	var constraintChecks []ConstraintCheck
	// Identify under-loaded nodes (e.g. < 80% of mean)
	underLoadedThreshold := mean * 0.8
	
	for nodeID, count := range nodePartitionCounts {
		if float64(count) < underLoadedThreshold {
			var reasons []string
			isValid := true
			
			// Check 1: Is Alive?
			if !nodeIsAlive[nodeID] {
				isValid = false
				reasons = append(reasons, "Node is not alive")
			}
			
			// Check 2: Is Draining?
			if nodeIsDraining[nodeID] {
				isValid = false
				reasons = append(reasons, "Node is draining (maintenance mode)")
			}
			
			// Check 3: Disk Space
			// If disk usage > 80% (default limit), it cannot accept more
			diskPct := 0.0
			if stats, ok := diskStats[nodeID]; ok && stats.TotalBytes > 0 {
				used := stats.TotalBytes - stats.FreeBytes
				diskPct = (float64(used) / float64(stats.TotalBytes)) * 100
				
				maxDiskPct := float64(getInt(clusterConfig, "partition_autobalancing_max_disk_usage_percent", 80))
								if diskPct >= maxDiskPct {
									isValid = false
									reasons = append(reasons, fmt.Sprintf("Disk usage (%.1f%%) exceeds limit (%.1f%%)", diskPct, maxDiskPct))
								}
							}
				
							constraintChecks = append(constraintChecks, ConstraintCheck{				NodeID:         nodeID,
				IsValidTarget:  isValid,
				BlockerReasons: reasons,
				DiskUsagePct:   diskPct,
				IsAlive:        nodeIsAlive[nodeID],
				IsDraining:     nodeIsDraining[nodeID],
			})
		}
	}

	// Calculate Topic Skews
	var topicSkews []TopicSkew
	for topic, nodeMap := range nodeTopicCounts {
		totalReplicas := 0
		maxReplicas := 0
		minReplicas := math.MaxInt32
		
		for nodeID := range allNodes {
			count := nodeMap[nodeID]
			totalReplicas += count
			if count > maxReplicas {
				maxReplicas = count
			}
			if count < minReplicas {
				minReplicas = count
			}
		}

		if totalNodes > 0 {
			meanReplicas := float64(totalReplicas) / float64(totalNodes)
			
			maxSkew := 0.0
			if meanReplicas > 0 {
				maxSkew = (float64(maxReplicas) - meanReplicas) / meanReplicas * 100
			}

			// Flag if unbalanced (max-min > 1) AND significant skew (> 20%)
			isSkewed := (maxReplicas - minReplicas) > 1 && maxSkew > 20.0

			if isSkewed {
				topicSkews = append(topicSkews, TopicSkew{
					TopicName:        topic,
					PartitionCount:   topicPartitionCounts[topic],
					NodeDistribution: nodeMap,
					MaxSkew:          maxSkew,
					IsSkewed:         isSkewed,
				})
			}
		}
	}

	balancingAnalysis := AnalyzeBalancing(brokers, nodePartitionCounts, nodeTopicCounts, clusterConfig, diskStats)
	
	// Analyze Hot Topics (Throughput)
	var hotTopics []HotTopic
	if len(topicThroughput) > 0 {
		var totalThroughput float64
		var count float64
		for _, v := range topicThroughput {
			totalThroughput += v
			count++
		}
		meanThroughput := totalThroughput / count

		var sumSquares float64
		for _, v := range topicThroughput {
			diff := v - meanThroughput
			sumSquares += diff * diff
		}
		stdDevThroughput := math.Sqrt(sumSquares / count)

		for topic, val := range topicThroughput {
			// Threshold: > 2 StdDevs above mean (and non-trivial amount)
			deviation := 0.0
			if stdDevThroughput > 0 {
				deviation = (val - meanThroughput) / stdDevThroughput
			}
			
			// We only care if it's significantly higher than average
			if deviation > 2.0 && val > 1024*1024 { // at least 1MB to avoid noise in empty clusters
				hotTopics = append(hotTopics, HotTopic{
					TopicName:  topic,
					BytesValue: val,
					Deviation:  deviation,
					IsHot:      true,
				})
			}
		}
	}

	// Generate Hint
	skewHint := ""
	if isClusterSkewed {
		skewHint = "Cluster is skewed. "
		if balancerStatus.Status == "ready" || balancerStatus.Status == "off" {
			skewHint += fmt.Sprintf("Balancer is '%s' but skew exists. ", balancerStatus.Status)
			
			// Check for Intra-Rack masking
			hasRackSkew := false
			for _, rs := range rackSkews {
				if rs.Skewed {
					hasRackSkew = true
					break
				}
			}
			
			if hasRackSkew {
				skewHint += "Likely cause: Intra-Rack imbalance masked by satisfied Rack Awareness constraints. Action: Trigger on-demand rebalance."
			} else {
				skewHint += "Action: Check constraints or trigger rebalance."
			}
		} else {
			skewHint += fmt.Sprintf("Balancer is '%s'.", balancerStatus.Status)
		}
	}

	return SkewAnalysis{
		TotalPartitions:  totalPartitions,
		TotalNodes:       totalNodes,
		MeanPartitions:   mean,
		StdDev:           stdDev,
		NodeSkews:        nodeSkews,
		TopicSkews:       topicSkews,
		RackSkews:        rackSkews,
		ConstraintChecks: constraintChecks,
		BalancerStatus:   balancerStatus,
		Balancing:        balancingAnalysis,
		IsClusterSkewed:  false,
		ClusterSkewHint:  skewHint,
	}
}

