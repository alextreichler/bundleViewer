package analysis

import (
	"math"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NodeSkew struct {
	NodeID         int
	PartitionCount int
	TopicCount     int
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

type BalancingAnalysis struct {
	Config     BalancingConfig
	DiskConfig DiskBalancingConfig
	Nodes      []NodeBalancingScore
}

type SkewAnalysis struct {
	TotalPartitions int
	TotalNodes      int
	MeanPartitions  float64
	StdDev          float64
	NodeSkews       []NodeSkew
	TopicSkews      []TopicSkew
	Balancing       BalancingAnalysis
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
) BalancingAnalysis {
	config := BalancingConfig{
		PartitionsPerShard:      getInt(clusterConfig, "topic_partitions_per_shard", 5000),
		PartitionsReserveShard0: getInt(clusterConfig, "topic_partitions_reserve_shard0", 0),
		TopicAware:              clusterConfig["partition_autobalancing_mode"] == "topic_aware",
		RackAwarenessEnabled:    getBool(clusterConfig, "enable_rack_awareness", false),
	}

	var nodeScores []NodeBalancingScore

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
	}

	diskConfig := DiskBalancingConfig{
		SpaceManagementEnabled:                    getBool(clusterConfig, "space_management_enabled", true),
		DiskReservationPercent:                    getInt(clusterConfig, "disk_reservation_percent", 20),                // Defaulting to 20%
		RetentionLocalTargetCapacityPercent:       getInt(clusterConfig, "retention_local_target_capacity_percent", 80), // Defaulting to 80%
		RetentionLocalTargetCapacityBytes:         getInt64(clusterConfig, "retention_local_target_capacity_bytes", 0),
		PartitionAutobalancingMaxDiskUsagePercent: getInt(clusterConfig, "partition_autobalancing_max_disk_usage_percent", 80), // Defaulting to 80%
		StorageSpaceAlertFreeThresholdPercent:     getInt(clusterConfig, "storage_space_alert_free_threshold_percent", 15),     // Defaulting to 15%
	}

	return BalancingAnalysis{
		Config:     config,
		DiskConfig: diskConfig,
		Nodes:      nodeScores,
	}
}

func AnalyzePartitionSkew(partitions []models.ClusterPartition, brokers []interface{}, clusterConfig map[string]interface{}) SkewAnalysis {
	// 1. Count Partitions and Topics per Node
	nodePartitionCounts := make(map[int]int)
	// topic -> node -> count
	nodeTopicCounts := make(map[string]map[int]int)
	allNodes := make(map[int]bool)

	// Initialize all known nodes with 0
	for _, b := range brokers {
		if bMap, ok := b.(map[string]interface{}); ok {
			if id, ok := bMap["node_id"].(float64); ok {
				allNodes[int(id)] = true
				nodePartitionCounts[int(id)] = 0
			}
		}
	}

	for _, p := range partitions {
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
	for nodeID, count := range nodePartitionCounts {
		deviation := 0.0
		if stdDev > 0 {
			deviation = (float64(count) - mean) / stdDev
		}

		skewPct := 0.0
		if mean > 0 {
			skewPct = ((float64(count) - mean) / mean) * 100
		}

		isSkewed := math.Abs(skewPct) > 20.0

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
			Deviation:      deviation,
			SkewPercentage: skewPct,
			IsSkewed:       isSkewed,
		})
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
					PartitionCount:   totalReplicas,
					NodeDistribution: nodeMap,
					MaxSkew:          maxSkew,
					IsSkewed:         isSkewed,
				})
			}
		}
	}

	balancingAnalysis := AnalyzeBalancing(brokers, nodePartitionCounts, nodeTopicCounts, clusterConfig)

	return SkewAnalysis{
		TotalPartitions: totalPartitions,
		TotalNodes:      totalNodes,
		MeanPartitions:  mean,
		StdDev:          stdDev,
		NodeSkews:       nodeSkews,
		TopicSkews:      topicSkews,
		Balancing:       balancingAnalysis,
	}
}
