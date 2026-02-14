package analysis

import (
	"fmt"
	"sort"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NetworkHealthReport struct {
	MaxRetransRate      float64
	AvgRetransRate      float64
	MaxInterfaceUtil    float64
	PeakThroughputMBps  float64
	DetectedAnomalies   []string
	Recommendations     []string
}

// AnalyzeNetwork performs a deep dive into network performance using SAR data and metrics
func AnalyzeNetwork(sar models.SarData, metrics *models.MetricsBundle) *NetworkHealthReport {
	report := &NetworkHealthReport{
		DetectedAnomalies: []string{},
		Recommendations:   []string{},
	}

	if len(sar.Entries) == 0 {
		return nil
	}

	var totalRetrans, peakRetrans float64
	var peakUtil float64
	var peakTraffic float64 // in KB/s
	var totalInErrs float64

	// Sort entries by timestamp just in case
	entries := make([]models.SarEntry, len(sar.Entries))
	copy(entries, sar.Entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	for _, entry := range entries {
		// TCP Retransmissions (The most critical network metric for Raft)
		if entry.TcpRetransRate > peakRetrans {
			peakRetrans = entry.TcpRetransRate
		}
		totalRetrans += entry.TcpRetransRate
		totalInErrs += entry.TcpInErrs

		// Interface Utilization and Throughput
		for _, dev := range entry.NetDev {
			if dev.IFace == "lo" {
				continue // Skip loopback
			}
			if dev.IfUtil > peakUtil {
				peakUtil = dev.IfUtil
			}
			traffic := dev.RxkB + dev.TxkB
			if traffic > peakTraffic {
				peakTraffic = traffic
			}
		}
	}

	report.MaxRetransRate = peakRetrans
	report.AvgRetransRate = totalRetrans / float64(len(entries))
	report.MaxInterfaceUtil = peakUtil
	report.PeakThroughputMBps = peakTraffic / 1024.0

	// 1. Analyze Retransmissions
	// Heuristic: > 1.0 retrans/s is worth noting, > 5.0 is a problem
	if peakRetrans > 5.0 {
		report.DetectedAnomalies = append(report.DetectedAnomalies, fmt.Sprintf("Critical: High TCP Retransmission Rate (Peak: %.2f/s).", peakRetrans))
		report.Recommendations = append(report.Recommendations, "High retransmissions cause Raft timeouts and leadership churn. Check for cable issues, switch congestion, or VPC throttling.")
	} else if peakRetrans > 1.0 {
		report.DetectedAnomalies = append(report.DetectedAnomalies, fmt.Sprintf("Warning: Elevated TCP Retransmissions (Peak: %.2f/s).", peakRetrans))
	}

	// 2. Analyze Interface Saturation
	// Heuristic: > 80% utility is saturation
	if peakUtil > 80.0 {
		report.DetectedAnomalies = append(report.DetectedAnomalies, fmt.Sprintf("Critical: Network Interface Saturation (Peak: %.1f%%).", peakUtil))
		report.Recommendations = append(report.Recommendations, "The physical network link is saturated. Consider upgrading instance types or using 25Gbps+ networking.")
	} else if peakUtil > 50.0 {
		report.DetectedAnomalies = append(report.DetectedAnomalies, fmt.Sprintf("Warning: High Network Interface Utilization (Peak: %.1f%%).", peakUtil))
	}

	// 3. Analyze TCP Errors
	if totalInErrs > 0 {
		report.DetectedAnomalies = append(report.DetectedAnomalies, "TCP Input Errors detected. Packets are being dropped at the OS/Stack level.")
		report.Recommendations = append(report.Recommendations, "Check 'sysctl net.core.netdev_max_backlog' and 'net.ipv4.tcp_max_syn_backlog'. The kernel may be dropping packets under load.")
	}

	// 4. Correlate with CPU (System time)
	// High system time + high network traffic = Interrupt storm
	for _, entry := range entries {
		if entry.CPUSystem > 20.0 && (entry.TcpInSegs+entry.TcpOutSegs) > 50000 {
			report.DetectedAnomalies = append(report.DetectedAnomalies, "Potential Interrupt Storm: High kernel CPU usage correlated with high packet rates.")
			report.Recommendations = append(report.Recommendations, "Ensure NIC IRQs are balanced across cores. Check '/proc/interrupts' or enable 'irqbalance'.")
			break
		}
	}

	return report
}
