package models

import "time"

// RecordBatchHeader represents the Redpanda batch header
// https://github.com/redpanda-data/redpanda/blob/dev/src/v/model/record_batch.h
type RecordBatchHeader struct {
	HeaderCRC            uint32
	BatchLength          int32
	BaseOffset           int64
	Type                 int8
	CRC                  int32
	Attributes           int16
	LastOffsetDelta      int32
	FirstTimestamp       int64
	MaxTimestamp         int64
	ProducerID           int64
	ProducerEpoch        int16
	BaseSequence         int32
	RecordCount          int32
	
	// Derived/Helper fields
	CompressionType      string
	Timestamp            time.Time
	MaxOffset            int64
	TypeString           string // e.g. "Raft", "Controller"
	RaftConfig           *RaftConfig // Populated if Type == Raft Config
}

// RaftConfig represents the decoded payload of a Raft Configuration batch
type RaftConfig struct {
	Voters   []int32
	Learners []int32
}

// SnapshotHeader represents the Redpanda Raft Snapshot header
type SnapshotHeader struct {
	HeaderCRC         uint32
	HeaderDataCRC     uint32
	HeaderVersion     int8
	MetadataSize      uint32
	LastIncludedIndex int64
	LastIncludedTerm  int64
}

// SegmentInfo holds summary data for a log segment file
type SegmentInfo struct {
	FilePath    string
	SizeBytes   int64
	BaseOffset  int64 // From filename usually
	Batches     []RecordBatchHeader
	Snapshot    *SnapshotHeader // Populated if this is a snapshot file
	Error       error
}
