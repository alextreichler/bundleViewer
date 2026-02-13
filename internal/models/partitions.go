package models

// PartitionInfo holds information about a single partition.
type PartitionInfo struct {
	ID       int              `json:"id"`
	Leader   int              `json:"leader"`
	Replicas []int            `json:"replicas"`
	Status   string           `json:"status"`
	Insight  *PartitionInsight `json:"insight,omitempty"`
}

type PartitionInsight struct {
	Summary    string   `json:"summary"`
	Remediation string   `json:"remediation"`
	Severity   string   `json:"severity"` // "info", "warn", "error"
}

// PartitionLeader represents the leader of a single partition.
type PartitionLeader struct {
	Ns                   string `json:"ns"`
	Topic                string `json:"topic"`
	PartitionID          int    `json:"partition_id"`
	Leader               int    `json:"leader"`
	PreviousLeader       int    `json:"previous_leader"`
	LastStableLeaderTerm int    `json:"last_stable_leader_term"`
}

// ClusterPartition represents a single partition in the cluster.
type ClusterPartition struct {
	Ns           string    `json:"ns"`
	Topic        string    `json:"topic"`
	PartitionID  int       `json:"partition_id"`
	Status       string    `json:"status"`
	LeaderID     int       `json:"leader_id"`
	Replicas     []struct {
		NodeID int `json:"node_id"`
		Core   int `json:"core"`
	} `json:"replicas"`
	Revision     int       `json:"revision"`
}

// PartitionDebug represents the output of /v1/debug/partition/...
type PartitionDebug struct {
	Ns       string           `json:"ns"`
	Topic    string           `json:"topic"`
	ID       int              `json:"partition_id"`
	Replicas []ReplicaDebug   `json:"replicas"`
}

type ReplicaDebug struct {
	CommittedOffset int64 `json:"committed_offset"`
	RaftState       struct {
		NodeID    int `json:"node_id"`
		Followers []struct {
			ID                   int   `json:"id"`
			HeartbeatsFailed     int   `json:"heartbeats_failed"`
			MsSinceLastHeartbeat int64 `json:"ms_since_last_heartbeat"`
		} `json:"followers"`
	} `json:"raft_state"`
}

