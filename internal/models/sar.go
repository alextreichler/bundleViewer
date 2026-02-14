package models

import "time"

type SarEntry struct {
	Timestamp time.Time `json:"timestamp"`
	
	// CPU Stats (-u)
	CPUUser   float64   `json:"cpu_user"`
	CPUSystem float64   `json:"cpu_system"`
	CPUIdle   float64   `json:"cpu_idle"`
	CPUIowait float64   `json:"cpu_iowait"`

	// Memory Stats (-r)
	MemUsedPct float64  `json:"mem_used_pct"`

	// Network DEV Stats (-n DEV)
	NetDev []SarNetDevEntry `json:"net_dev"`

	// Network TCP Stats (-n TCP, ETCP)
	TcpActiveOpens  float64 `json:"tcp_active_opens"`
	TcpPassiveOpens float64 `json:"tcp_passive_opens"`
	TcpInSegs       float64 `json:"tcp_in_segs"`
	TcpOutSegs      float64 `json:"tcp_out_segs"`
	TcpRetransRate  float64 `json:"tcp_retrans_rate"` // retrans/s
	TcpInErrs       float64 `json:"tcp_in_errs"`

	// Derived metrics for UI
	PeakThroughputMBps float64 `json:"peak_throughput_mbps"`
}

type SarNetDevEntry struct {
	IFace   string  `json:"iface"`
	RxPck   float64 `json:"rx_pck"`
	TxPck   float64 `json:"tx_pck"`
	RxkB    float64 `json:"rx_kb"`
	TxkB    float64 `json:"tx_kb"`
	RxMcst  float64 `json:"rx_mcst"`
	IfUtil  float64 `json:"if_util"`
}

type SarData struct {
	Entries []SarEntry
}
