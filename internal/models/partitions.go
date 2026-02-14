package models

import "time"

// PartitionInfo holds information about a single partition.
type PartitionInfo struct {
	Ns          string            `json:"ns"`
	Topic       string            `json:"topic"`
	ID          int               `json:"id"`
	Leader      int               `json:"leader"`
	Term        int               `json:"term"`
	Replicas    []int             `json:"replicas"`
	Status      string            `json:"status"`
	RaftDetails []ReplicaRaftInfo `json:"raft_details,omitempty"`
	Insight     *PartitionInsight `json:"insight,omitempty"`
}

type ReplicaRaftInfo struct {
	NodeID   int  `json:"node_id"`
	IsLeader bool `json:"is_leader"`
	Term     int  `json:"term"`
}

type RaftTimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Node        string    `json:"node"`
	Term        int       `json:"term"`
	Type        string    `json:"type"` // "Elected", "Stepdown", "Timeout", "Move"
	NTP         string    `json:"ntp"`  // Namespace/Topic/Partition
	Reason      string    `json:"reason"`
	Description string    `json:"description"`
}

type PartitionInsight struct {
	Summary     string `json:"summary"`
	Remediation string `json:"remediation"`
	Severity    string `json:"severity"` // "info", "warn", "error"
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
	Ns          string `json:"ns"`
	Topic       string `json:"topic"`
	PartitionID int    `json:"partition_id"`
	Status      string `json:"status"`
	LeaderID    int    `json:"leader_id"`
	Replicas    []struct {
		NodeID int `json:"node_id"`
		Core   int `json:"core"`
	} `json:"replicas"`
	Revision int `json:"revision"`
}

// PartitionDebug represents the output of /v1/debug/partition/...
type PartitionDebug struct {
	Ns       string         `json:"ns"`
	Topic    string         `json:"topic"`
	ID       int            `json:"partition_id"`
	Replicas []ReplicaDebug `json:"replicas"`
}

type ReplicaDebug struct {
	CommittedOffset int64 `json:"committed_offset"`
	RaftState       struct {
		NodeID          int  `json:"node_id"`
		Term            int  `json:"term"`
		IsLeader        bool `json:"is_leader"`
		IsElectedLeader bool `json:"is_elected_leader"`
		Followers       []struct {
			ID                   int   `json:"id"`
			HeartbeatsFailed     int   `json:"heartbeats_failed"`
			MsSinceLastHeartbeat int64 `json:"ms_since_last_heartbeat"`
			LastFlushedLogIndex  int64 `json:"last_flushed_log_index"`
			LastDirtyLogIndex    int64 `json:"last_dirty_log_index"`
		} `json:"followers"`
	} `json:"raft_state"`
}
