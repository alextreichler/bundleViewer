# BundleViewer

BundleViewer is a specialized offline analysis tool designed for **Redpanda diagnostic bundles**. It parses, indexes, and visualizes the complex contents of a bundle (logs, metrics, configuration, and metadata) to accelerate root-cause analysis.

**You take on all risks when using this. Use at your own risk!**

## 🚀 Key Features

*   **Hybrid Accelerated Architecture:** 
    *   **Zig Sidecar:** Uses a high-performance **Zig engine with SIMD instructions** to accelerate log scanning, Prometheus metric parsing, and pattern fingerprinting.
    *   **SQLite-Backed:** High-volume data is ingested into an optimized SQLite database with **WAL mode** and **30GB+ memory-mapping** for near-instant searches.
*   **Advanced Log Analysis:**
    *   **Full-Text Search:** Lightning-fast global search across gigabytes of log data using SQLite's **FTS5** trigram tokenizer.
    *   **SIMD Fingerprinting:** Vectorized clustering of log messages to identify recurring errors and anomalies at memory speed.
*   **Expert Diagnostics:** Automatically detects complex failure modes based on real-world Redpanda internals:
    *   **Hardware Audit:** Validates disk throughput against the **"4/30" minimum threshold** (4MB/s small block, 30MB/s medium block).
    *   **Performance Bottlenecks:** Detects Seastar reactor spinning, Schema Registry reference building issues, and Raft state machine recovery loops.
    *   **Memory Intelligence:** Accounts for **Batch Cache reclaim** logic ("Healthy Zero") and flags heap fragmentation risks on long-running nodes.
    *   **Safety Audits:** Warns of data loss risks from **unclean reconfiguration aborts**.
*   **Rich Visualizations:**
    *   **Timeline View:** Unified event stream correlating logs and K8s events across all nodes.
    *   **CPU Profiling:** Visualizes reactor utilization with **Smart Insights** that automatically flag known code-level bottlenecks.
    *   **Metric Dashboards:** Interactive charts for reactor utilization, I/O queue depth, and throughput.

## 🛠️ Prerequisites

*   **Go:** Version 1.25 or higher.
*   **Zig:** Version 0.15.2 or higher (required for building the native accelerator).
*   **Task:** [Taskfile](https://taskfile.dev/) is used for orchestrating the hybrid Go/Zig build.

## 📦 Installation

### 1. One-Line Installation (Recommended)
Install the latest pre-compiled version directly to `/usr/local/bin`. The distributed binaries are statically linked and include the Zig accelerator pre-built:

```bash
curl -fsSL https://raw.githubusercontent.com/alextreichler/bundleViewer/main/install.sh | sh
```

### 2. Manual Installation (From Source)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/alextreichler/bundleViewer.git
    cd bundleViewer
    ```

2.  **Build the hybrid binary:**
    The project uses a specialized build pipeline that uses **Zig as a C compiler** to ensure static linking and easy cross-compilation.
    ```bash
    task build
    ```
    This will produce optimized binaries for both macOS and Linux in the `dist/` directory.

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

### ☁️ Cloud & Ad-hoc Usage
BundleViewer is optimized for ephemeral container environments (AWS EKS, Fargate, etc.). It supports:
*   **Automatic Fetching:** Download and extract bundles directly from S3/HTTPS.
*   **One-Shot Mode:** Immutable, stateless analysis for single-investigation pods.
*   **Auto-Cleanup:** Self-terminates after a period of inactivity to save costs.
*   **Resource Tuning:** Scales its memory footprint to fit your container limits.

**Note:** All cloud features are optional and default to off. Local usage remains exactly the same.

See the **[Cloud Deployment Guide](CLOUD_DEPLOYMENT.md)** for more details.

### CLI Options
*   `-port <int>`: Port to listen on (default: `7575`).
*   `-host <string>`: Host to bind to (default: `127.0.0.1`).
*   `-one-shot`: Enable stateless mode (disables bundle switching).
*   `-memory-limit <string>`: Cap memory usage (e.g. `2GB`).
*   `-persist`: Keep the SQLite database (`bundle.db`) after the server stops.
*   `-logs-only`: Only process and display logs (faster loading).

## 🛡️ Security Notice

**This tool is designed for local desktop use only.**

*   **Local File Access:** The tool has access to the filesystem of the user running it. It exposes features to browse and open directories.
*   **Network Exposure:** By default, it binds to `127.0.0.1` (localhost). **Do not** bind it to `0.0.0.0` or expose it to a public network, as this would allow unauthenticated remote access to your files.
*   **Input Validation:** While efforts are made to sanitize inputs, you should only open diagnostic bundles from trusted sources.

## 🏗️ Architecture

The project follows a hybrid Go/Zig project layout:

*   **`internal/parser/native/`**: High-performance Zig parsing engine using SIMD vectors.
*   **`internal/parser/`**: Go bridge for the Zig engine and standard parsers for Linux system files (`sar`, `vmstat`, `sysctl`).
*   **`internal/store/`**: Optimized database layer using WAL-mode SQLite for high concurrency.
*   **`internal/analysis/`**: Expert-system heuristics for performance, memory, and hardware validation.
*   **`internal/diagnostics/`**: The "Auditor" engine that runs automated health checks.
*   **`internal/server/`**: HTMX-powered web server providing a responsive, zero-build-step UI.

