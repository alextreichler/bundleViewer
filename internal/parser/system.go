package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// ParseMemInfo parses /proc/meminfo
func ParseMemInfo(bundlePath string) (models.MemInfo, error) {
	filePath := filepath.Join(bundlePath, "proc", "meminfo")
	file, err := os.Open(filePath)
	if err != nil {
		return models.MemInfo{}, err
	}
	defer func() { _ = file.Close() }()

	info := models.MemInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		valStr := fields[1]
		val, _ := strconv.ParseInt(valStr, 10, 64)
		// MemInfo values are usually in kB
		valBytes := val * 1024

		switch key {
		case "AnonHugePages":
			info.AnonHugePages = valBytes
		case "Shmem":
			info.Shmem = valBytes
		case "Slab":
			info.Slab = valBytes
		case "PageTables":
			info.PageTables = valBytes
		case "Dirty":
			info.Dirty = valBytes
		case "Writeback":
			info.Writeback = valBytes
		case "Mapped":
			info.Mapped = valBytes
		case "Active":
			info.Active = valBytes
		case "Inactive":
			info.Inactive = valBytes
		}
	}
	return info, nil
}

// ParseVMStat parses /proc/vmstat
func ParseVMStat(bundlePath string) (models.VMStat, error) {
	filePath := filepath.Join(bundlePath, "proc", "vmstat")
	file, err := os.Open(filePath)
	if err != nil {
		return models.VMStat{}, err
	}
	defer func() { _ = file.Close() }()

	stat := models.VMStat{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := fields[0]
		valStr := fields[1]
		val, _ := strconv.ParseUint(valStr, 10, 64)

		switch key {
		case "ctxt":
			stat.ContextSwitches = val
		case "processes":
			stat.ProcessesForked = val
		case "procs_running":
			stat.ProcsRunning = val
		case "procs_blocked":
			stat.ProcsBlocked = val
		}
	}
	return stat, nil
}

// ParseVMStatTimeSeries parses utils/vmstat.txt
func ParseVMStatTimeSeries(bundlePath string) (models.VMStatSample, error) {
	filePath := filepath.Join(bundlePath, "utils", "vmstat.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.VMStatSample{}, err
	}
	defer func() { _ = file.Close() }()

	sample := models.VMStatSample{}
	scanner := bufio.NewScanner(file)
	
	// Skip headers
	// procs -----------------------memory...
	// r  b   swpd   free...
	scanner.Scan()
	scanner.Scan()
	
	var sumR, sumB, sumWa float64
	var count int
	
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		// r b swpd free buff cache si so bi bo in cs us sy id wa st
		// 0 1 2    3    4    5     6  7  8  9  10 11 12 13 14 15 16
		
		if len(fields) >= 16 {
			r, _ := strconv.Atoi(fields[0])
			b, _ := strconv.Atoi(fields[1])
			wa, _ := strconv.Atoi(fields[15])
			
			sumR += float64(r)
			sumB += float64(b)
			sumWa += float64(wa)
			
			if r > sample.MaxRunnable { sample.MaxRunnable = r }
			if b > sample.MaxBlocked { sample.MaxBlocked = b }
			if wa > sample.MaxIOWait { sample.MaxIOWait = wa }
			
			count++
		}
	}
	
	if count > 0 {
		sample.Samples = count
		sample.AvgRunnable = sumR / float64(count)
		sample.AvgBlocked = sumB / float64(count)
		sample.AvgIOWait = sumWa / float64(count)
	}
	
	return sample, nil
}

// ParseLSPCI parses utils/lspci.txt
func ParseLSPCI(bundlePath string) (models.LSPCIInfo, error) {
	filePath := filepath.Join(bundlePath, "utils", "lspci.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.LSPCIInfo{}, err
	}
	defer func() { _ = file.Close() }()

	info := models.LSPCIInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		info.Devices = append(info.Devices, scanner.Text())
	}
	return info, nil
}

// ParseDig parses utils/dig.txt
func ParseDig(bundlePath string) (models.DigInfo, error) {
	filePath := filepath.Join(bundlePath, "utils", "dig.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.DigInfo{}, err
	}
	defer func() { _ = file.Close() }()

	info := models.DigInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		
		// ;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 62227
		if strings.Contains(line, "status:") {
			parts := strings.Split(line, "status:")
			if len(parts) > 1 {
				statusPart := strings.Split(parts[1], ",")
				info.Status = strings.TrimSpace(statusPart[0])
			}
		}
		
		// ;; SERVER: 127.0.0.53#53(127.0.0.53)
		// Or common dig output might not have SERVER line if it failed or format differs.
		// The example showed:
		// .                       900     IN      NS      oleibdd01.itg.ti.com.
		// We can try to infer something or just look for 'status: NOERROR' which is key.
	}
	return info, nil
}

// ParseSyslog parses utils/syslog.txt for critical events (OOM)
func ParseSyslog(bundlePath string) (models.SyslogAnalysis, error) {
	filePath := filepath.Join(bundlePath, "utils", "syslog.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.SyslogAnalysis{}, err
	}
	defer func() { _ = file.Close() }()

	analysis := models.SyslogAnalysis{}
	scanner := bufio.NewScanner(file)

	// Regex for OOM
	// <4>[25850749.679985] prometheus invoked oom-killer: gfp_mask=...
	reOOM := regexp.MustCompile(`\[\s*(\d+\.\d+)\s*\]\s+(.*)\s+invoked oom-killer`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Check for OOM
		if strings.Contains(line, "oom-killer") {
			matches := reOOM.FindStringSubmatch(line)
			if len(matches) >= 3 {
				analysis.OOMEvents = append(analysis.OOMEvents, models.OOMEvent{
					Timestamp:   matches[1], // Kernel timestamp
					ProcessName: matches[2],
					Message:     line,
				})
			}
		}
	}
	return analysis, nil
}

// ParseDMI parses utils/dmidecode.txt
func ParseDMI(bundlePath string) (models.DMIInfo, error) {
	filePath := filepath.Join(bundlePath, "utils", "dmidecode.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.DMIInfo{}, err
	}
	defer func() { _ = file.Close() }()

	info := models.DMIInfo{}
	scanner := bufio.NewScanner(file)

	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		if !strings.HasPrefix(line, "\t") && trimmed != "" {
			currentSection = trimmed
			continue
		}

		if currentSection == "BIOS Information" {
			if strings.HasPrefix(trimmed, "Vendor:") {
				info.BIOSVendor = strings.TrimSpace(strings.TrimPrefix(trimmed, "Vendor:"))
			}
		} else if currentSection == "System Information" {
			if strings.HasPrefix(trimmed, "Manufacturer:") {
				info.Manufacturer = strings.TrimSpace(strings.TrimPrefix(trimmed, "Manufacturer:"))
			} else if strings.HasPrefix(trimmed, "Product Name:") {
				info.Product = strings.TrimSpace(strings.TrimPrefix(trimmed, "Product Name:"))
			}
		}
	}
	return info, nil
}

// ParseCPUInfo parses /proc/cpuinfo
func ParseCPUInfo(bundlePath string) (models.CPUInfo, error) {
	filePath := filepath.Join(bundlePath, "proc", "cpuinfo")
	file, err := os.Open(filePath)
	if err != nil {
		return models.CPUInfo{}, err
	}
	defer func() { _ = file.Close() }()

	info := models.CPUInfo{}
	scanner := bufio.NewScanner(file)

	// We only need to parse the first processor's info as they are typically identical
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if info.ModelName == "" {
				info.ModelName = val
			}
		case "vendor_id":
			if info.VendorID == "" {
				info.VendorID = val
			}
		case "cpu MHz":
			if info.Mhz == 0 {
				info.Mhz, _ = strconv.ParseFloat(val, 64)
			}
		case "cache size":
			if info.CacheSize == "" {
				info.CacheSize = val
			}
		case "flags":
			if len(info.Flags) == 0 {
				allFlags := strings.Fields(val)
				// Filter for interesting flags (avx, sse4_2, etc.)
				interesting := []string{"avx", "avx2", "sse4_2", "fma", "hypervisor"}
				for _, flag := range allFlags {
					for _, target := range interesting {
						if flag == target {
							info.Flags = append(info.Flags, flag)
						}
					}
				}
			}
		}
	}
	return info, nil
}

// ParseMDStat parses /proc/mdstat
func ParseMDStat(bundlePath string) (models.MDStat, error) {
	filePath := filepath.Join(bundlePath, "proc", "mdstat")
	file, err := os.Open(filePath)
	if err != nil {
		return models.MDStat{}, err
	}
	defer func() { _ = file.Close() }()

	stat := models.MDStat{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Personalities") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				fields := strings.Fields(parts[1])
				for _, f := range fields {
					stat.Personalities = append(stat.Personalities, strings.Trim(f, "[]"))
				}
			}
			continue
		}

		if strings.HasPrefix(line, "md") {
			// md0 : active raid1 sdb1[1] sda1[0]
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			rest := strings.TrimSpace(parts[1])
			
			// Split "active raid1 sdb1[1] sda1[0]"
			fields := strings.Fields(rest)
			if len(fields) >= 2 {
				state := fields[0] + " " + fields[1]
				devices := fields[2:]
				
				array := models.MDArray{
					Name:    name,
					State:   state,
					Devices: devices,
				}
				
				// Next line usually contains blocks and status
				// 12345 blocks ... [UU]
				if scanner.Scan() {
					nextLine := scanner.Text()
					nextFields := strings.Fields(nextLine)
					if len(nextFields) > 0 {
						if blocks, err := strconv.ParseInt(nextFields[0], 10, 64); err == nil {
							array.Blocks = blocks
						}
					}
					// Find status [UU] or [_U]
					if idx := strings.LastIndex(nextLine, "["); idx != -1 {
						if endIdx := strings.LastIndex(nextLine, "]"); endIdx > idx {
							array.Status = nextLine[idx : endIdx+1]
						}
					}
				}
				stat.Arrays = append(stat.Arrays, array)
			}
		}
	}
	return stat, nil
}

// ParseCmdLine parses /proc/cmdline
func ParseCmdLine(bundlePath string) (string, error) {
	filePath := filepath.Join(bundlePath, "proc", "cmdline")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// ParseDF parses the df.txt file, falling back to /proc/mounts if missing.
func ParseDF(bundlePath string) ([]models.FileSystemEntry, error) {
	filePath := filepath.Join(bundlePath, "utils", "df.txt")
	file, err := os.Open(filePath)
	if err != nil {
		// Fallback to /proc/mounts
		return ParseMounts(bundlePath)
	}
	defer func() { _ = file.Close() }()

	var entries []models.FileSystemEntry
	scanner := bufio.NewScanner(file)

	// Skip header
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Filesystem Type 1K-blocks Used Available Use% Mounted on
		// If Type column is missing (some df versions), we need to be careful,
		// but the example showed it exists.
		// However, sometimes fields can be merged if names are long.
		// Assuming standard columns from the example provided.

		// Example: /dev/mapper/vgbi--redpanda-node-root ext4 380288728 8436056 352461800 3% /

		if len(fields) >= 7 {
			// fields[0] = Filesystem
			// fields[1] = Type
			// fields[2] = 1K-blocks
			// fields[3] = Used
			// fields[4] = Available
			// fields[5] = Use%
			// fields[6] = Mounted on (could be multiple if spaces? unlikely for mount points in server context usually)

			total, _ := strconv.ParseInt(fields[2], 10, 64)
			used, _ := strconv.ParseInt(fields[3], 10, 64)
			avail, _ := strconv.ParseInt(fields[4], 10, 64)

			entries = append(entries, models.FileSystemEntry{
				Filesystem: fields[0],
				Type:       fields[1],
				Total:      total * 1024,
				Used:       used * 1024,
				Available:  avail * 1024,
				UsePercent: fields[5],
				MountPoint: fields[6],
			})
		}
	}
	return entries, nil
}

// ParseMounts parses /proc/mounts as a fallback for missing df.txt
func ParseMounts(bundlePath string) ([]models.FileSystemEntry, error) {
	filePath := filepath.Join(bundlePath, "proc", "mounts")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []models.FileSystemEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Format: device mountpoint type options dump pass
		entries = append(entries, models.FileSystemEntry{
			Filesystem: fields[0],
			MountPoint: fields[1],
			Type:       fields[2],
			UsePercent: "-", // Data not available in mounts
		})
	}
	return entries, nil
}

// ParseFree parses the free.txt file
func ParseFree(bundlePath string) (models.MemoryStats, error) {
	filePath := filepath.Join(bundlePath, "utils", "free.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.MemoryStats{}, err
	}
	defer func() { _ = file.Close() }()

	stats := models.MemoryStats{}
	scanner := bufio.NewScanner(file)

	// Example:
	//                total        used        free      shared  buff/cache   available
	// Mem:      2113309740  1929583736   179052256        5532    15357256   183726004
	// Swap:        1003516           0     1003516

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "Mem:":
			// Mem: total used free shared buff/cache available
			if len(fields) >= 7 {
				stats.Total = parseKB(fields[1])
				stats.Used = parseKB(fields[2])
				stats.Free = parseKB(fields[3])
				stats.Shared = parseKB(fields[4])
				stats.BuffCache = parseKB(fields[5])
				stats.Available = parseKB(fields[6])
			}
		case "Swap:":
			// Swap: total used free
			if len(fields) >= 4 {
				stats.SwapTotal = parseKB(fields[1])
				stats.SwapUsed = parseKB(fields[2])
				stats.SwapFree = parseKB(fields[3])
			}
		}
	}
	return stats, nil
}

func parseKB(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val * 1024
}

// ParseTop parses the top.txt file to extract load average
func ParseTop(bundlePath string) (models.LoadAvg, error) {
	filePath := filepath.Join(bundlePath, "utils", "top.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.LoadAvg{}, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		// top - 15:18:12 up 34 days, 20 min,  3 users,  load average: 66.85, 59.87, 60.92
		if idx := strings.Index(line, "load average:"); idx != -1 {
			loadStr := line[idx+13:]
			parts := strings.Split(loadStr, ",")
			if len(parts) >= 3 {
				one, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				five, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				fifteen, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
				return models.LoadAvg{
					OneMin:     one,
					FiveMin:    five,
					FifteenMin: fifteen,
				}, nil
			}
		}
	}
	return models.LoadAvg{}, nil
}

// ParseSysctl parses the sysctl.txt file
func ParseSysctl(bundlePath string) (map[string]string, error) {
	filePath := filepath.Join(bundlePath, "utils", "sysctl.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// regex: key = value
	re := regexp.MustCompile(`^([^=]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			config[strings.TrimSpace(matches[1])] = strings.TrimSpace(matches[2])
		}
	}
	return config, nil
}

// ParseNTP parses the ntp.txt file (JSON)
func ParseNTP(bundlePath string) (models.NTPStatus, error) {
	filePath := filepath.Join(bundlePath, "utils", "ntp.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.NTPStatus{}, err
	}

	var status models.NTPStatus
	// Some bundles might have multiple JSON objects or array?
	// The example showed a single object: {"host":"pool.ntp.org",...}
	if err := json.Unmarshal(data, &status); err != nil {
		return models.NTPStatus{}, err
	}
	return status, nil
}

// ParseIP parses the ip.txt file for network interfaces
func ParseIP(bundlePath string) ([]models.NetworkInterface, error) {
	filePath := filepath.Join(bundlePath, "utils", "ip.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var interfaces []models.NetworkInterface
	var currentInterface *models.NetworkInterface

	scanner := bufio.NewScanner(file)
	// Example:
	// 1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 ...
	//     link/loopback ...
	//     inet 127.0.0.1/8 scope host lo ...
	//     inet6 ::1/128 scope host ...

	reInterface := regexp.MustCompile(`^\d+: `)

	for scanner.Scan() {
		line := scanner.Text()

		// New interface start
		if reInterface.MatchString(line) {
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}
			parts := strings.SplitN(line, ": ", 3)
			if len(parts) >= 2 {
				name := parts[1]
				remaining := ""
				if len(parts) > 2 {
					remaining = parts[2]
				}

				// Extract MTU
				mtu := 0
				mtuIdx := strings.Index(remaining, "mtu ")
				if mtuIdx != -1 {
					mtuStr := strings.Fields(remaining[mtuIdx+4:])[0]
					mtu, _ = strconv.Atoi(mtuStr)
				}

				// Extract Flags (inside <>)
				flags := ""
				start := strings.Index(remaining, "<")
				end := strings.Index(remaining, ">")
				if start != -1 && end != -1 {
					flags = remaining[start+1 : end]
				}

				currentInterface = &models.NetworkInterface{
					Name:  name,
					Flags: flags,
					MTU:   mtu,
				}
			}
		} else if currentInterface != nil {
			// Parse addresses
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "inet ") {
				fields := strings.Fields(trimmed)
				if len(fields) >= 2 {
					currentInterface.Addresses = append(currentInterface.Addresses, fields[1])
				}
			} else if strings.HasPrefix(trimmed, "inet6 ") {
				fields := strings.Fields(trimmed)
				if len(fields) >= 2 {
					currentInterface.Addresses = append(currentInterface.Addresses, fields[1])
				}
			}
		}
	}
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces, nil
}

// ParseSS parses ss.txt for relevant network connections
func ParseSS(bundlePath string) ([]models.NetworkConnection, error) {
	filePath := filepath.Join(bundlePath, "utils", "ss.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var connections []models.NetworkConnection
	scanner := bufio.NewScanner(file)

	// Skip header
	// Netid  State      Recv-Q Send-Q Local Address:Port                 Peer Address:Port
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		netid := fields[0]

		// Filter: We care mostly about TCP and UDP
		if netid != "tcp" && netid != "udp" {
			continue
		}

		state := fields[1]
		recvQ := fields[2]
		sendQ := fields[3]
		local := fields[4]
		peer := ""
		if len(fields) > 5 {
			peer = fields[5]
		}

		// Further filtering:
		// Hide standard loopback unless specifically asked?
		// Actually loopback 127.0.0.1:9644 is common for admin.
		// Let's keep everything but maybe emphasize external IPs in UI.

		connections = append(connections, models.NetworkConnection{
			Netid:     netid,
			State:     state,
			RecvQ:     recvQ,
			SendQ:     sendQ,
			LocalAddr: local,
			PeerAddr:  peer,
		})
	}
	return connections, nil
}
