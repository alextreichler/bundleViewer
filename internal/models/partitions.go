package models

// PartitionInfo holds information about a single partition.
type PartitionInfo struct {
	ID       int   `json:"id"`
	Leader   int   `json:"leader"`
	Replicas []int `json:"replicas"`
}

// PartitionLeader represents the leader of a single partition.
type PartitionLeader struct {
	Ns          string `json:"ns"`
	Topic       string `json:"topic"`
	PartitionID int    `json:"partition_id"`
	Leader      int    `json:"leader"`
}

// ClusterPartition represents a single partition in the cluster.
type ClusterPartition struct {
	Ns          string `json:"ns"`
	Topic       string `json:"topic"`
	PartitionID int    `json:"partition_id"`
	Replicas    []struct {
		NodeID int `json:"node_id"`
		Core   int `json:"core"`
	} `json:"replicas"`
	Disabled bool `json:"disabled"`
	Leader   int  `json:"leader"` // Add Leader field
}
