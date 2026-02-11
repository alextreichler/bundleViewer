package models

import "encoding/json"

// KafkaMetadata represents the top-level structure of kafka.json.
type KafkaMetadata []struct {
	Name     string          `json:"Name"`
	Response json.RawMessage `json:"Response"` // Use RawMessage to handle varied response types
}

// KafkaMetadataResponse represents the structure of the "Response" field when Name is "metadata".
type KafkaMetadataResponse struct {
	Cluster    string `json:"Cluster"`
	Controller int    `json:"Controller"`
	Brokers    []struct {
		NodeID int     `json:"NodeID"`
		Port   int     `json:"Port"`
		Host   string  `json:"Host"`
		Rack   *string `json:"Rack"`
	} `json:"Brokers"`
	Topics map[string]struct {
		Topic                string   `json:"Topic"`
		ID                   string   `json:"ID"`
		IsInternal           bool     `json:"IsInternal"`
		AuthorizedOperations []string `json:"AuthorizedOperations"`
		Partitions           map[string]struct {
			Topic           string      `json:"Topic"`
			Partition       int         `json:"Partition"`
			Leader          int         `json:"Leader"`
			LeaderEpoch     int         `json:"LeaderEpoch"`
			Replicas        []int       `json:"Replicas"`
			ISR             []int       `json:"ISR"`
			OfflineReplicas interface{} `json:"OfflineReplicas"`
			Err             interface{} `json:"Err"`
		} `json:"Partitions"`
		Err interface{} `json:"Err"`
	} `json:"Topics"`
	AuthorizedOperations []string `json:"AuthorizedOperations"`
}

type ConsumerGroup struct {
	Group                string               `json:"Group"`
	Coordinator          GroupCoordinator     `json:"Coordinator"`
	State                string               `json:"State"`
	ProtocolType         string               `json:"ProtocolType"`
	Protocol             string               `json:"Protocol"`
	Members              []GroupMember        `json:"Members"`
	AuthorizedOperations []string             `json:"AuthorizedOperations"`
	Err                  interface{}          `json:"Err"`
	Offsets              []GroupTopicOffset   `json:"-"` // Not in the "groups" response, joined later
	TotalLag             int64                `json:"-"`
}

type GroupCoordinator struct {
	NodeID int     `json:"NodeID"`
	Port   int     `json:"Port"`
	Host   string  `json:"Host"`
	Rack   *string `json:"Rack"`
}

type GroupMember struct {
	MemberID   string                     `json:"MemberID"`
	InstanceID *string                    `json:"InstanceID"`
	ClientID   string                     `json:"ClientID"`
	ClientHost string                     `json:"ClientHost"`
	Join       interface{}                `json:"Join"`
	Assigned   map[string][]int           `json:"Assigned"` // Topic -> Partitions
}

type GroupTopicOffset struct {
	Topic      string
	Partitions map[int]*GroupPartitionOffset
	TopicLag   int64
}

type GroupPartitionOffset struct {
	Topic       string      `json:"Topic"`
	Partition   int         `json:"Partition"`
	At          int64       `json:"At"`
	LeaderEpoch int         `json:"LeaderEpoch"`
	Metadata    string      `json:"Metadata"`
	Err         interface{} `json:"Err"`
	Lag         int64       `json:"-"` // Calculated: HighWatermark - At
}

type HighWatermark struct {
	Topic       string      `json:"Topic"`
	Partition   int         `json:"Partition"`
	Timestamp   int64       `json:"Timestamp"`
	Offset      int64       `json:"Offset"`
	LeaderEpoch int         `json:"LeaderEpoch"`
	Err         interface{} `json:"Err"`
}


