package models

// Broker represents a broker in the cluster, derived from kafka.json or brokers.json
type Broker struct {
	NodeID int     `json:"node_id"`
	Port   int     `json:"port"`
	Host   string  `json:"host"`
	Rack   *string `json:"rack"` // Use pointer for optional field
}

// ClusterConfigEntry represents a single configuration entry from cluster_config.json
type ClusterConfigEntry struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	DocLink string      `json:"doc_link"`
}

type HealthOverview struct {
	IsHealthy            bool  `json:"is_healthy"`
	UnderReplicatedCount int   `json:"under_replicated_count"`
	LeaderlessCount      int   `json:"leaderless_count"`
	NodesDown            []int `json:"nodes_down"`
	AllNodes             []int `json:"all_nodes"`
}
