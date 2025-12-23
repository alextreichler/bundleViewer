package models

// FileSystemEntry represents a row from df command
type FileSystemEntry struct {
	Filesystem string
	Type       string
	Total      int64 // KB or bytes depending on parser, let's normalize to bytes
	Used       int64
	Available  int64
	UsePercent string
	MountPoint string
}

// NetworkInterface represents ip addr output
type NetworkInterface struct {
	Name      string
	Flags     string
	MTU       int
	Addresses []string // IPv4 and IPv6
}

// NTPStatus represents ntp.txt JSON content
type NTPStatus struct {
	Host            string  `json:"host"`
	RoundTripTimeMs float64 `json:"roundTripTimeMs"`
	RemoteTimeUTC   string  `json:"remoteTimeUTC"`
	LocalTimeUTC    string  `json:"localTimeUTC"`
	PrecisionMs     float64 `json:"precisionMs"`
	Offset          int64   `json:"offset"` // Can be nanoseconds or similar, usually
}

// NetworkConnection represents a line from ss.txt
type NetworkConnection struct {
	Netid     string
	State     string
	RecvQ     string
	SendQ     string
	LocalAddr string
	PeerAddr  string
	Process   string
}

// PortCount for sorting
type PortCount struct {
	Port  string
	Count int
}

// ConnectionSummary holds aggregated stats for network connections
type ConnectionSummary struct {
	Total        int
	ByState      map[string]int
	ByPort       map[string]int // Keep for internal use if needed
	SortedByPort []PortCount    // For sorted display in UI
}

// MemoryStats represents data from free command
type MemoryStats struct {
	Total     int64
	Used      int64
	Free      int64
	Shared    int64
	BuffCache int64
	Available int64
	SwapTotal int64
	SwapUsed  int64
	SwapFree  int64
}

// LoadAvg represents load average from top command
type LoadAvg struct {
	OneMin     float64
	FiveMin    float64
	FifteenMin float64
}

// UnameInfo represents system information from uname command
type UnameInfo struct {
	KernelName    string
	Hostname      string
	KernelRelease string
	KernelVersion string
	Machine       string
	OperatingSystem string
}

// DMIInfo represents hardware information from dmidecode
type DMIInfo struct {
	Manufacturer string
	Product      string
	BIOSVendor   string
}

// DigInfo represents DNS information from dig command
type DigInfo struct {
	Status string // NOERROR, SERVFAIL, etc.
	Server string // Nameserver that answered
}

// VMStatSample represents aggregated stats from utils/vmstat.txt
type VMStatSample struct {
	AvgRunnable float64
	MaxRunnable int
	AvgBlocked  float64
	MaxBlocked  int
	AvgIOWait   float64
	MaxIOWait   int
	Samples     int
}

// LSPCIInfo represents parsed utils/lspci.txt
type LSPCIInfo struct {
	Devices []string
}

// SyslogAnalysis represents findings from syslog
type SyslogAnalysis struct {
	OOMEvents []OOMEvent
}

type OOMEvent struct {
	Timestamp   string
	ProcessName string
	Message     string
}

// MemInfo represents parsed /proc/meminfo data
type MemInfo struct {
	AnonHugePages int64 // Bytes
	Shmem         int64 // Bytes
	Slab          int64 // Bytes
	PageTables    int64 // Bytes
	Dirty         int64 // Bytes
	Writeback     int64 // Bytes
	Mapped        int64 // Bytes
	Active        int64 // Bytes
	Inactive      int64 // Bytes
}

// VMStat represents parsed /proc/vmstat data
type VMStat struct {
	ContextSwitches uint64 // ctxt
	ProcessesForked uint64 // processes
	ProcsRunning    uint64 // procs_running
	ProcsBlocked    uint64 // procs_blocked
}

// CPUInfo represents parsed /proc/cpuinfo
type CPUInfo struct {
	ModelName string
	VendorID  string
	Mhz       float64
	CacheSize string
	Flags     []string // partial list of interesting flags
}

// MDStat represents parsed /proc/mdstat (Software RAID)
type MDStat struct {
	Personalities []string
	Arrays        []MDArray
}

type MDArray struct {
	Name    string
	State   string // e.g., "active raid1"
	Devices []string
	Blocks  int64
	Status  string // e.g., "[UU]", "[_U]" (degraded)
}

// SystemState aggregates all system level information
type SystemState struct {
	FileSystems []FileSystemEntry
	Memory      MemoryStats
	MemInfo     MemInfo // Detailed memory info from /proc/meminfo
	Load        LoadAvg
	Uname       UnameInfo
	DMI         DMIInfo // Hardware info
	Dig         DigInfo // DNS info
	Syslog      SyslogAnalysis // OOMs and critical errors
	VMStatAnalysis VMStatSample // From utils/vmstat.txt (Time series)
	LSPCI       LSPCIInfo
	CPU         CPUInfo // From /proc/cpuinfo
	Sysctl      map[string]string
	Interfaces  []NetworkInterface
	NTP         NTPStatus
	Connections []NetworkConnection
	ConnSummary ConnectionSummary
	VMStat      VMStat // From /proc/vmstat
	MDStat      MDStat // From /proc/mdstat
	CmdLine     string // From /proc/cmdline
	CoreCount   int
}

// TimelineSourceData holds all data needed to generate the timeline
type TimelineSourceData struct {
	Logs      []*LogEntry
	K8sEvents []K8sResource // Placeholder, we might need a specific Event struct
	K8sPods   []K8sResource // Added for Pod lifecycle analysis
}
