package logutil

import (
	"regexp"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type LogInsight struct {
	Pattern     string
	Description string
	Severity    string // "Critical", "Warning", "Info"
	HelpLink    string // Optional link to docs
	Action      string // Recommended action
	regex       *regexp.Regexp
}

var CommonErrorKnowledgeBase = []LogInsight{
	{
		Pattern:     `vote_stm\.cc.*error code: 10`,
		Description: "Raft Vote Timeout. The node timed out while waiting for votes during an election.",
		Severity:    "Warning",
		Action:      "Check for network latency, CPU starvation, or disk stalls preventing the node from responding in time.",
	},
	{
		Pattern:     `raft::errc::timeout`,
		Description: "Raft Consensus Timeout. A request (append, vote, etc.) timed out.",
		Severity:    "Warning",
		Action:      "Check network connectivity and node load (CPU/Disk).",
	},
	{
		Pattern:     `bad_alloc`,
		Description: "Out of Memory (OOM). The Seastar memory allocator failed to allocate memory.",
		Severity:    "Critical",
		Action:      "Check if the node's memory is over-provisioned or if there is a memory leak. Verify cgroup limits.",
	},
	{
		Pattern:     `not_enough_free_space`,
		Description: "Disk Full. Redpanda cannot write data because the disk is nearly full.",
		Severity:    "Critical",
		Action:      "Increase disk space or adjust retention policies to delete older data.",
	},
	{
		Pattern:     `filesystem error`,
		Description: "Filesystem Error. The underlying OS reported a failure writing to or reading from disk.",
		Severity:    "Critical",
		Action:      "Check hardware health (SMART), dmesg, and disk mounting options.",
	},
	{
		Pattern:     `not_leader_for_partition`,
		Description: "Not Leader for Partition. A client tried to produce/consume from a node that is no longer the leader.",
		Severity:    "Info",
		Action:      "Normal during elections. If persistent, check if the client's metadata is refreshing properly.",
	},
	{
		Pattern:     `crc mismatch`,
		Description: "Data Corruption. A checksum mismatch was detected in a log segment.",
		Severity:    "Critical",
		Action:      "Identify the corrupted segment. You may need to delete it and let it recover from a replica, or restore from backup.",
	},
	{
		Pattern:     `segment_not_found`,
		Description: "Segment Missing. A requested log segment could not be found on disk.",
		Severity:    "Critical",
		Action:      "Check if retention deleted the file while it was being read, or if there is disk corruption.",
	},
	{
		Pattern:     `connection refused`,
		Description: "Connection Refused. Failed to establish a network connection to a peer or service.",
		Severity:    "Warning",
		Action:      "Verify that the target process is running and that the network/firewall allows the connection.",
	},
	{
		Pattern:     `broken pipe`,
		Description: "Broken Pipe. An existing connection was closed by the remote peer.",
		Severity:    "Warning",
		Action:      "Check for network instability or the peer node crashing/restarting.",
	},
	{
		Pattern:     `too many open files`,
		Description: "File Descriptor Exhaustion. The process has reached its limit of open files.",
		Severity:    "Critical",
		Action:      "Increase the 'ulimit -n' for the Redpanda process or check for a file descriptor leak.",
	},
	{
		Pattern:     `reactor_stalled`,
		Description: "Reactor Stalled. A task took too long before yielding to the scheduler.",
		Severity:    "Warning",
		Action:      "Contributes to tail latency. Large stalls (>1.5s) can trigger leader stepdowns and cluster instability. Investigate CPU-bound synchronous loops.",
	},
	{
		Pattern:     `authentication failed`,
		Description: "SASL Authentication Failed. A client failed to authenticate.",
		Severity:    "Warning",
		Action:      "Verify client credentials, SASL mechanism (SCRAM/PLAIN/GSSAPI), and user permissions.",
	},
	{
		Pattern:     `cloud_storage.*upload failed`,
		Description: "Tiered Storage Upload Failed. Failed to upload a segment to cloud storage (S3/GCS).",
		Severity:    "Warning",
		Action:      "Check cloud provider connectivity, permissions (IAM roles), and bucket configuration.",
	},
	{
		Pattern:     `cloud_storage.*download failed`,
		Description: "Tiered Storage Download Failed. Failed to retrieve a segment from cloud storage.",
		Severity:    "Warning",
		Action:      "Check network connectivity to the cloud provider and verify the object exists in the bucket.",
	},
	{
		Pattern:     `schema_registry.*error`,
		Description: "Schema Registry Error. An error occurred in the Schema Registry component.",
		Severity:    "Warning",
		Action:      "Check the schema registry logs for specific validation or storage errors.",
	},
	{
		Pattern:     `reconfiguration_stm.*moving partition`,
		Description: "Partition Movement. A partition is being relocated to different nodes.",
		Severity:    "Info",
		Action:      "Check if this was triggered by the partition balancer or a manual 'rpk cluster partitions move'.",
	},
	{
		Pattern:     `reconfiguration_stm.*finishing reconfiguration`,
		Description: "Partition Movement Finished. The relocation of a partition has completed successfully.",
		Severity:    "Info",
		Action:      "Normal lifecycle event.",
	},
	{
		Pattern:     `leadership_balancer.*transferring leadership`,
		Description: "Automatic Leadership Rebalance. The leadership balancer is moving a leader to improve cluster balance.",
		Severity:    "Info",
		Action:      "Normal background operation if 'enable_leader_balancer' is true.",
	},
	{
		Pattern:     `health_monitor.*detected node.*is down`,
		Description: "Node Down Detected. The cluster health monitor has flagged a node as unresponsive.",
		Severity:    "Critical",
		Action:      "Investigate the specific node for crashes, OOMs, or network isolation.",
	},
	{
		Pattern:     `became the leader term`,
		Description: "Raft Leadership Election. A node successfully became the leader of a partition.",
		Severity:    "Info",
		Action:      "Informational. Useful for tracking leadership stability.",
	},
	{
		Pattern:     `Stepping down as leader in term.*reasoning: leadership_transfer`,
		Description: "Manual Leadership Transfer. The leader is stepping down because a transfer was requested (e.g., by the automatic balancer or an admin).",
		Severity:    "Info",
		Action:      "Check if partition balancing is active or if a move was manually initiated.",
	},
	{
		Pattern:     `Stepping down as leader in term.*reasoning: heartbeats_majority`,
		Description: "Loss of Majority. The leader stepped down because it lost contact with a majority of its followers.",
		Severity:    "Warning",
		Action:      "Check for network partitions, high CPU load, or disk stalls causing heartbeats to fail. Consider increasing 'election_timeout_ms' if the environment is unstable.",
	},
	{
		Pattern:     `ghost_record_batch`,
		Description: "Ghost Batch Detected. Redpanda encountered a gap in the log and filled it with a placeholder to maintain continuity.",
		Severity:    "Info",
		Action:      "This is a normal part of the recovery process when gaps are detected. No action needed unless frequent.",
	},
	{
		Pattern:     `oversized_record_batch`,
		Description: "Oversized Batch. A client attempted to produce a batch larger than the configured max_message_bytes.",
		Severity:    "Warning",
		Action:      "Increase 'kafka_batch_max_bytes' on the cluster or reduce the batch size in the client producer settings.",
	},
	{
		Pattern:     `cluster_id_mismatch`,
		Description: "Cluster ID Mismatch. The node's data directory belongs to a different cluster.",
		Severity:    "Critical",
		Action:      "Verify that you have mounted the correct data volume. Do not force start without verifying, or data loss may occur.",
	},
	{
		Pattern:     `stored_wasm_binary`,
		Description: "Wasm Binary Stored. A new data transform (Wasm) binary has been deployed.",
		Severity:    "Info",
		Action:      "Informational event for Data Transforms.",
	},
	{
		Pattern:     `datalake_throttle_manager.*Translation status updated.*throttling may be applied`,
		Description: "Data Lake Translation Throttling (Iceberg Topics). The translation process (Kafka to Iceberg) is falling behind and the backlog exceeds the threshold.",
		Severity:    "Warning",
		Action:      "Check Iceberg destination health. Increase 'iceberg_throttle_backlog_size_ratio' if disk space allows. Check disk space usage.",
	},
	{
		Pattern:     `pandaproxy::schema_registry::build_file_with_refs`,
		Description: "Schema Registry Bottleneck. Heavy activity in schema normalization or reference building.",
		Severity:    "Warning",
		Action:      "Known issue in v25.1.1 (Fixed in v25.1.2). Occurs with large numbers of schema subjects.",
	},
	{
		Pattern:     `raft::state_machine_manager::apply`,
		Description: "Raft State Machine recovery loop.",
		Severity:    "Warning",
		Action:      "May indicate a known issue with state machine recovery policy. Check Redpanda version.",
	},
	{
		Pattern:     `ssl_accept.*failed`,
		Description: "SSL Handshake Failure / Churn.",
		Severity:    "Warning",
		Action:      "High frequency of SSL errors often indicates clients with repeat login attempts or missing connection pooling.",
	},
	{
		Pattern:     `unclean_abort_reconfiguration`,
		Description: "Unclean Partition Reconfiguration Abort. A partition movement was forcefully cancelled.",
		Severity:    "Critical",
		Action:      "DANGER: This operation involves data loss by truncating logs. Verify data consistency for the affected partition.",
	},
}

func init() {
	for i := range CommonErrorKnowledgeBase {
		CommonErrorKnowledgeBase[i].regex = regexp.MustCompile("(?i)" + CommonErrorKnowledgeBase[i].Pattern)
	}
}

// GetInsight finds the first matching insight for a message
func GetInsight(message string) *models.LogInsight {
	for _, insight := range CommonErrorKnowledgeBase {
		if insight.regex.MatchString(message) {
			return &models.LogInsight{
				Description: insight.Description,
				Severity:    insight.Severity,
				Action:      insight.Action,
			}
		}
	}
	return nil
}
