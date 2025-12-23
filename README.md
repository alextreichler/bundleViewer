# BundleViewer

BundleViewer is a specialized offline analysis tool designed for **Redpanda diagnostic bundles**. It parses, indexes, and visualizes the complex contents of a bundle (logs, metrics, configuration, and metadata).

**You take on all risks when using this. Use at your own risk!**


## 🚀 Key Features

*   **Hybrid Architecture:** 
    *   **In-Memory:** Critical metadata (Kubernetes resources, Kafka topic configurations, disk usage) is parsed and held in memory for instant access.
    *   **SQLite-Backed:** High-volume data (logs, Prometheus metrics) is ingested into an embedded SQLite database.
*   **Advanced Log Analysis:**
    *   **Full-Text Search:** Utilizes SQLite's **FTS5** extension for lightning-fast global search across gigabytes of log data.
    *   **Pattern Fingerprinting:** Automatically clusters similar log messages to identify recurring errors and anomalies without drowning in noise.
*   **Interactive UI:** A lightweight, responsive web interface built with **Go Templates** and **HTMX**. No heavy frontend build step required.
*   **Rich Visualizations:**
    *   **Timeline View:** Correlate events across nodes.
    *   **Skew Analysis:** Detect uneven data distribution across partitions and disks.
    *   **Resource Usage:** Visualize disk and CPU metrics over time.

## 🛠️ Prerequisites

*   **Go:** Version 1.25 or higher.
*   **Task:** [Taskfile](https://taskfile.dev/) is used for build automation (optional, but recommended).
*   **SQLite:** The project uses `modernc.org/sqlite` (pure Go), so no CGO or external SQLite installation is strictly required for the build, but the underlying system must support the build targets.

## 📦 Installation

### 1. One-Line Installation (Recommended)
You can install the latest version of BundleViewer directly to `/usr/local/bin` using the following command:

```bash
curl -fsSL https://raw.githubusercontent.com/alextreichler/bundleViewer/main/install.sh | sh
```

*This script auto-detects your OS (macOS/Linux) and architecture (amd64/arm64) and fetches the correct binary.*

### 2. Manual Installation (From Source)

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/alextreichler/bundleViewer.git
    cd bundleViewer
    ```

2.  **Build the binary:**
    Using Task (recommended):
    ```bash
    task build
    ```
    Or using standard Go commands:
    ```bash
    go build -o bundleViewer cmd/webapp/main.go
    ```

## 🖥️ Usage

Run the tool by pointing it to a directory containing an unzipped Redpanda debug bundle, or simply run it without arguments to use the interactive setup wizard.

### 1. Interactive Mode (Recommended)
Simply run the command without arguments. BundleViewer will automatically open your default browser, where you can paste the path to your bundle:

```bash
bundleViewer
```

### 2. Direct Mode
Provide the path directly via the CLI:

```bash
bundleViewer /path/to/extracted/bundle
```

### Using Task
If you are developing locally:
```bash
task run -- /path/to/extracted/bundle
```

### Options
*   `-port <int>`: Port to listen on (default: `7575`).
*   `-host <string>`: Host to bind to (default: `127.0.0.1` for security).
*   `-persist`: Keep the SQLite database (`bundle.db`) after the server stops. By default, the database is cleaned up on start to ensure a fresh state for new bundles.
*   `-logs-only`: Only process and display logs (skips metrics and system parsing for faster loading).

## 🏗️ Architecture

The project follows a standard Go project layout:

*   `cmd/webapp/`: Entry point (`main.go`). Handles flag parsing and server initialization.
*   `internal/parser/`: specialized logic for parsing Redpanda-specific files.
*   `internal/store/`: Database layer. Manages the SQLite connection, schema migrations, and high-performance bulk insertion.
*   `internal/analysis/`: Business logic for log fingerprinting and heuristic analysis.
*   `internal/server/`: HTTP handlers and routing logic.
*   `ui/`: Contains HTML templates (`.tmpl`) and static assets (CSS, JS). These are embedded into the binary for single-file distribution.

