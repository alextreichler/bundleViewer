# BundleViewer

BundleViewer is a specialized offline analysis tool designed for **Redpanda diagnostic bundles**. It parses, indexes, and visualizes the complex contents of a bundle (logs, metrics, configuration, and metadata) to accelerate root-cause analysis.

**You take on all risks when using this. Use at your own risk!**

## 🚀 Key Features

*   **Hybrid Architecture:** 
    *   **In-Memory:** Critical metadata (Kubernetes resources, Kafka topic configurations, disk usage) is parsed and held in memory for instant access.
    *   **SQLite-Backed:** High-volume data (logs, Prometheus metrics, CPU profiles) is ingested into an embedded SQLite database.
*   **Advanced Log Analysis:**
    *   **Full-Text Search:** Utilizes SQLite's **FTS5** extension for lightning-fast global search across gigabytes of log data.
    *   **Pattern Fingerprinting:** Automatically clusters similar log messages to identify recurring errors and anomalies without drowning in noise.
*   **Automated Diagnostics (Auditor):** Runs 50+ heuristic checks across:
    *   **OS Tuning:** Verifies `sysctl` settings (AIO, network backlogs, virtual memory).
    *   **Cluster Health:** Detects Under-Replicated Partitions (URP), leaderless partitions, and controller instability.
    *   **Performance Bottlenecks:** Identifies CPU, Disk, and Network saturation.
    *   **Infrastructure:** Checks for OOM events, RAID degradation, and Kubernetes pod restarts.
*   **Rich Visualizations:**
    *   **Timeline View:** Unified event stream correlating logs and K8s events across all nodes.
    *   **Skew Analysis:** Detects uneven distribution of partitions and topics across shards and disks.
    *   **CPU Profiling:** Visualizes reactor utilization and hot paths from internal Redpanda CPU profiles.
    *   **Metric Dashboards:** Interactive charts for reactor utilization, I/O queue depth, and throughput.

## 🛠️ Prerequisites

*   **Go:** Version 1.25 or higher.
*   **Task:** [Taskfile](https://taskfile.dev/) is used for build automation (optional, but recommended).
*   **SQLite:** The project uses `modernc.org/sqlite` (pure Go), so no CGO or external SQLite installation is required.

## 📦 Installation

### 1. One-Line Installation (Recommended)
Install the latest version directly to `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/alextreichler/bundleViewer/main/install.sh | sh
```

### 2. Manual Installation (From Source)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/alextreichler/bundleViewer.git
    cd bundleViewer
    ```

2.  **Build the binary:**
    ```bash
    task build
    ```
    *Or use `go build -o bundleViewer cmd/webapp/main.go`*

## 🖥️ Usage

### 1. Interactive Mode (Recommended)
Simply run the command without arguments. BundleViewer will open your browser to the setup page where you can paste the bundle path:

```bash
bundleViewer
```

### 2. Direct Mode
Provide the path directly via the CLI:

```bash
bundleViewer /path/to/extracted/bundle
```

### CLI Options
*   `-port <int>`: Port to listen on (default: `7575`).
*   `-host <string>`: Host to bind to (default: `127.0.0.1`).
*   `-persist`: Keep the SQLite database (`bundle.db`) after the server stops.
*   `-logs-only`: Only process and display logs (faster loading).

## 🛡️ Security Notice

**This tool is designed for local desktop use only.**

*   **Local File Access:** The tool has access to the filesystem of the user running it. It exposes features to browse and open directories.
*   **Network Exposure:** By default, it binds to `127.0.0.1` (localhost). **Do not** bind it to `0.0.0.0` or expose it to a public network, as this would allow unauthenticated remote access to your files.
*   **Input Validation:** While efforts are made to sanitize inputs, you should only open diagnostic bundles from trusted sources.

## 🏗️ Architecture

The project follows a standard Go project layout:

*   `internal/parser/`: Specialized logic for Redpanda logs, metrics, admin API responses, and Linux system files (`sar`, `vmstat`, `sysctl`).
*   `internal/store/`: Database layer. High-performance bulk insertion and FTS indexing using SQLite.
*   `internal/analysis/`: Logic for log fingerprinting, performance heuristics, and partition skew calculations.
*   `internal/diagnostics/`: The "Auditor" engine that runs automated health checks.
*   `internal/server/`: HTMX-powered web server providing a responsive, zero-build-step UI.
*   `ui/`: Embedded HTML templates and static assets.

