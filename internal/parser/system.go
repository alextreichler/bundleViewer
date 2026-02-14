package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
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

// ParseCmdLine parses Redpanda process command line, handling various bundle formats
func ParseCmdLine(bundlePath string) (string, error) {
	// 1. Try common locations for the Redpanda process command line
	candidates := []string{
		filepath.Join(bundlePath, "proc", "cmdline"),      // Generic capture
		filepath.Join(bundlePath, "redpanda.cmdline"),     // Direct capture
	}

	// Also look for any proc/<number>/cmdline files
	matches, _ := filepath.Glob(filepath.Join(bundlePath, "proc", "*", "cmdline"))
	candidates = append(candidates, matches...)

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Linux /proc/PID/cmdline uses null bytes (\0) as separators.
		// Replace them with spaces so regex/searches work.
		s := string(data)
		s = strings.ReplaceAll(s, "\x00", " ")
		s = strings.TrimSpace(s)

		// Check if this looks like a Redpanda command line (heuristic)
		if strings.Contains(s, "redpanda") {
			return s, nil
		}
	}

	// 2. Fallback: Try to find it in the first few lines of the log
	logPath := filepath.Join(bundlePath, "redpanda.log")
	if f, err := os.Open(logPath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; i < 200 && scanner.Scan(); i++ {
			line := scanner.Text()
			// Redpanda often logs: "command line: [ ... ]"
			if strings.Contains(line, "command line:") {
				return line, nil
			}
		}
	}

	return "", fmt.Errorf("redpanda command line not found")
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
// ParseTop parses the top.txt file to extract load average and top processes
func ParseTop(bundlePath string) (models.LoadAvg, []models.Process, error) {
	filePath := filepath.Join(bundlePath, "utils", "top.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return models.LoadAvg{}, nil, err
	}
	defer func() { _ = file.Close() }()

	var load models.LoadAvg
	var processes []models.Process
	scanner := bufio.NewScanner(file)
	
	headerFound := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// 1. Parse Load Average (usually first line)
		if strings.Contains(line, "load average:") {
			idx := strings.Index(line, "load average:")
			loadStr := line[idx+13:]
			parts := strings.Split(loadStr, ",")
			if len(parts) >= 3 {
				load.OneMin, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				load.FiveMin, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				load.FifteenMin, _ = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			}
			continue
		}

		// 2. Identify Process Header
		// PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
		if strings.Contains(line, "PID") && strings.Contains(line, "USER") && strings.Contains(line, "%CPU") {
			headerFound = true
			continue
		}

		// 3. Parse Process Rows
		if headerFound {
			fields := strings.Fields(line)
			if len(fields) >= 12 {
				// Column mapping varies slightly by top version but usually:
				// PID(0), USER(1), ..., %CPU(8), %MEM(9), ..., COMMAND(11)
				cpu, _ := strconv.ParseFloat(fields[8], 64)
				mem, _ := strconv.ParseFloat(fields[9], 64)
				
				processes = append(processes, models.Process{
					PID:     fields[0],
					User:    fields[1],
					CPU:     cpu,
					Memory:  mem,
					Command: strings.Join(fields[11:], " "),
				})
			}
		}
	}
	return load, processes, nil
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

// ParseLSBLK parses utils/lsblk.txt
func ParseLSBLK(bundlePath string) ([]models.BlockDevice, error) {
	filePath := filepath.Join(bundlePath, "utils", "lsblk.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var devices []models.BlockDevice
	scanner := bufio.NewScanner(file)

	// Check if it's JSON (sometimes captured with lsblk -J)
	firstLine := ""
	if scanner.Scan() {
		firstLine = strings.TrimSpace(scanner.Text())
	}
	if strings.HasPrefix(firstLine, "{") {
		// Try parsing as JSON
		var raw struct {
			BlockDevices []models.BlockDevice `json:"blockdevices"`
		}
		// Need to read the rest of the file
		var fullContent strings.Builder
		fullContent.WriteString(firstLine)
		for scanner.Scan() {
			fullContent.WriteString(scanner.Text())
		}
		if err := json.Unmarshal([]byte(fullContent.String()), &raw); err == nil {
			return raw.BlockDevices, nil
		}
	}

	// Fallback: Parse as tree/table
	// NAME  KNAME  TYPE  SIZE  MOUNTPOINT  MODEL  SCHEDULER  RO
	// sda   sda    disk  20G   /           ...    mq-deadline 0
	
	// We need to handle the tree structure (└─, ├─)
	// For now, let's do a simple flat parse and try to detect parent/child by indentation
	
	// Re-open/reset scanner if needed, or just process what we have
	// Since we already scanned firstLine, we need to process it
	
	line := firstLine
	for {
		if line != "" && !strings.HasPrefix(line, "NAME") {
			trimmed := strings.TrimLeft(line, " |-└─├─")
			fields := strings.Fields(trimmed)
			if len(fields) >= 4 {
				dev := models.BlockDevice{
					Name:  fields[0],
					Type:  fields[2],
					Size:  fields[3],
				}
				// Try to find mountpoint if present
				// This is heuristic because columns vary
				for _, f := range fields {
					if strings.HasPrefix(f, "/") {
						dev.MountPoint = f
						break
					}
				}
				devices = append(devices, dev)
			}
		}
		if !scanner.Scan() {
			break
		}
		line = scanner.Text()
	}

	return devices, nil
}

// ParseSS parses ss.txt for relevant network connections, including TCP metrics
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
	if !scanner.Scan() {
		return nil, nil
	}

	var currentConn *models.NetworkConnection

	for scanner.Scan() {
		line := scanner.Text()
		
		// If line starts with whitespace, it's likely a continuation of the previous connection
		if (strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ")) && currentConn != nil {
			parseTCPInfo(line, currentConn)
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// New connection
		netid := fields[0]
		if netid != "tcp" && netid != "udp" && !strings.HasPrefix(netid, "ESTAB") {
			// Some ss versions don't show Netid as first column if filtered
			// If it looks like a state, assume TCP
			if isState(netid) {
				netid = "tcp"
			} else {
				continue
			}
		}

		// Save previous
		if currentConn != nil {
			connections = append(connections, *currentConn)
		}

		// Handle cases where Netid is missing
		offset := 0
		if fields[0] == "tcp" || fields[0] == "udp" {
			offset = 1
		}

		currentConn = &models.NetworkConnection{
			Netid:     netid,
			State:     fields[0+offset],
			RecvQ:     fields[1+offset],
			SendQ:     fields[2+offset],
			LocalAddr: fields[3+offset],
			PeerAddr:  fields[4+offset],
		}
	}
	
	if currentConn != nil {
		connections = append(connections, *currentConn)
	}

	return connections, nil
}

func isState(s string) bool {
	states := []string{"ESTAB", "LISTEN", "TIME-WAIT", "CLOSE-WAIT", "FIN-WAIT-1", "FIN-WAIT-2", "SYN-SENT", "SYN-RECV"}
	for _, st := range states {
		if s == st {
			return true
		}
	}
	return false
}

func parseTCPInfo(line string, conn *models.NetworkConnection) {
	if conn.TCPInfo == nil {
		conn.TCPInfo = &models.TCPInfo{}
	}

	// Look for users:(...)
	if idx := strings.Index(line, "users:(("); idx != -1 {
		endIdx := strings.Index(line[idx:], "))")
		if endIdx != -1 {
			conn.Process = line[idx+7 : idx+endIdx]
		}
	}

	// Parse metrics: retrans:0/1 rto:204 cwnd:10 rtt:0.01/0.02
	fields := strings.Fields(line)
	for _, f := range fields {
		if strings.HasPrefix(f, "retrans:") {
			parts := strings.Split(strings.TrimPrefix(f, "retrans:"), "/")
			if len(parts) >= 2 {
				conn.TCPInfo.Retrans, _ = strconv.Atoi(parts[0])
				conn.TCPInfo.RetransTotal, _ = strconv.Atoi(parts[1])
			}
		} else if strings.HasPrefix(f, "rto:") {
			conn.TCPInfo.RTO, _ = strconv.ParseFloat(strings.TrimPrefix(f, "rto:"), 64)
		} else if strings.HasPrefix(f, "cwnd:") {
			conn.TCPInfo.Cwnd, _ = strconv.Atoi(strings.TrimPrefix(f, "cwnd:"))
		} else if strings.HasPrefix(f, "ssthresh:") {
			conn.TCPInfo.Ssthresh, _ = strconv.Atoi(strings.TrimPrefix(f, "ssthresh:"))
		} else if strings.HasPrefix(f, "rtt:") {
			parts := strings.Split(strings.TrimPrefix(f, "rtt:"), "/")
			if len(parts) >= 2 {
				conn.TCPInfo.Rtt, _ = strconv.ParseFloat(parts[0], 64)
				conn.TCPInfo.RttVar, _ = strconv.ParseFloat(parts[1], 64)
			}
		} else if strings.HasPrefix(f, "mss:") {
			conn.TCPInfo.MSS, _ = strconv.Atoi(strings.TrimPrefix(f, "mss:"))
		}
	}
}

// ParseTHP parses /sys/kernel/mm/transparent_hugepage/enabled
func ParseTHP(bundlePath string) (string, error) {
	// Try multiple common paths inside the bundle
	paths := []string{
		filepath.Join(bundlePath, "sys", "kernel", "mm", "transparent_hugepage", "enabled"),
		filepath.Join(bundlePath, "sys", "kernel", "mm", "redhat_transparent_hugepage", "enabled"), // RHEL/CentOS
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return strings.TrimSpace(string(data)), nil
		}
	}
	return "", os.ErrNotExist
}

// ParseInterrupts parses /proc/interrupts
func ParseInterrupts(bundlePath string) (models.Interrupts, error) {
	filePath := filepath.Join(bundlePath, "proc", "interrupts")
	file, err := os.Open(filePath)
	if err != nil {
		return models.Interrupts{}, err
	}
	defer func() { _ = file.Close() }()

	var interrupts models.Interrupts
	scanner := bufio.NewScanner(file)

	// First line is header:           CPU0       CPU1 ...
	if scanner.Scan() {
		header := scanner.Text()
		interrupts.CPUs = strings.Fields(header)
	}

	numCPUs := len(interrupts.CPUs)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < numCPUs+1 {
			continue
		}

		entry := models.IRQEntry{
			IRQID: strings.TrimSuffix(parts[0], ":"),
		}

		for i := 0; i < numCPUs; i++ {
			val, _ := strconv.ParseInt(parts[i+1], 10, 64)
			entry.CPUCounts = append(entry.CPUCounts, val)
		}

		// Remaining parts are controller and device
		if len(parts) > numCPUs+1 {
			rest := parts[numCPUs+1:]
			// Try to capture meaningful device name
			// Format varies: IO-APIC 2-edge timer
			// or: PCI-MSI 524288-edge nvme0q0
			if len(rest) > 0 {
				entry.Controller = rest[0]
				if len(rest) > 1 {
					entry.Device = strings.Join(rest[1:], " ")
				} else {
					entry.Device = rest[0] // Fallback
				}
			}
		}
		interrupts.Entries = append(interrupts.Entries, entry)
	}
	return interrupts, nil
}

// ParseSoftIRQs parses /proc/softirqs
func ParseSoftIRQs(bundlePath string) (models.Interrupts, error) {
	filePath := filepath.Join(bundlePath, "proc", "softirqs")
	file, err := os.Open(filePath)
	if err != nil {
		return models.Interrupts{}, err
	}
	defer func() { _ = file.Close() }()

	var interrupts models.Interrupts
	scanner := bufio.NewScanner(file)

	// First line is header:           CPU0       CPU1 ...
	if scanner.Scan() {
		header := scanner.Text()
		interrupts.CPUs = strings.Fields(header)
	}

	numCPUs := len(interrupts.CPUs)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < numCPUs+1 {
			continue
		}

		entry := models.IRQEntry{
			IRQID: strings.TrimSuffix(parts[0], ":"),
		}

		for i := 0; i < numCPUs; i++ {
			val, _ := strconv.ParseInt(parts[i+1], 10, 64)
			entry.CPUCounts = append(entry.CPUCounts, val)
		}
		
		interrupts.Entries = append(interrupts.Entries, entry)
	}
	return interrupts, nil
}

// ParseSelfTestResults finds and parses self_test_status_*.json files in the admin directory.
func ParseSelfTestResults(bundlePath string) ([]models.SelfTestNodeResult, error) {
	pattern := filepath.Join(bundlePath, "admin", "self_test_status_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob self-test files: %w", err)
	}

	var allResults []models.SelfTestNodeResult

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var nodeResults []models.SelfTestNodeResult
		// It could be a single result or a list of results
		if strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
			if err := json.Unmarshal(data, &nodeResults); err != nil {
				continue
			}
		} else {
			var singleResult models.SelfTestNodeResult
			if err := json.Unmarshal(data, &singleResult); err != nil {
				continue
			}
			nodeResults = append(nodeResults, singleResult)
		}
		
		allResults = append(allResults, nodeResults...)
	}

	return allResults, nil
}
