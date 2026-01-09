================================================
FILE: README.md
================================================
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




================================================
FILE: go.mod
================================================
module github.com/alextreichler/bundleViewer

go 1.25.2

require modernc.org/sqlite v1.40.1

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sys v0.36.0 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)



================================================
FILE: go.sum
================================================
github.com/dustin/go-humanize v1.0.1 h1:GzkhY7T5VNhEkwH0PVJgjz+fX1rhBrR7pRT3mDkpeCY=
github.com/dustin/go-humanize v1.0.1/go.mod h1:Mu1zIs6XwVuF/gI1OepvI0qD18qycQx+mFykh5fBlto=
github.com/google/pprof v0.0.0-20250317173921-a4b03ec1a45e h1:ijClszYn+mADRFY17kjQEVQ1XRhq2/JR1M3sGqeJoxs=
github.com/google/pprof v0.0.0-20250317173921-a4b03ec1a45e/go.mod h1:boTsfXsheKC2y+lKOCMpSfarhxDeIzfZG1jqGcPl3cA=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
github.com/mattn/go-isatty v0.0.20 h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=
github.com/mattn/go-isatty v0.0.20/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
github.com/ncruces/go-strftime v0.1.9 h1:bY0MQC28UADQmHmaF5dgpLmImcShSi2kHU9XLdhx/f4=
github.com/ncruces/go-strftime v0.1.9/go.mod h1:Fwc5htZGVVkseilnfgOVb9mKy6w1naJmn9CehxcKcls=
github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec h1:W09IVJc94icq4NjY3clb7Lk8O1qJ8BdBEF8z0ibU0rE=
github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec/go.mod h1:qqbHyh8v60DhA7CoWK5oRCqLrMHRGoxYCSS9EjAz6Eo=
golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b h1:M2rDM6z3Fhozi9O7NWsxAkg/yqS/lQJ6PmkyIV3YP+o=
golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b/go.mod h1:3//PLf8L/X+8b4vuAfHzxeRUl04Adcb341+IGKfnqS8=
golang.org/x/mod v0.27.0 h1:kb+q2PyFnEADO2IEF935ehFUXlWiNjJWtRNgBLSfbxQ=
golang.org/x/mod v0.27.0/go.mod h1:rWI627Fq0DEoudcK+MBkNkCe0EetEaDSwJJkCcjpazc=
golang.org/x/sync v0.16.0 h1:ycBJEhp9p4vXvUZNszeOq0kGTPghopOL8q0fq3vstxw=
golang.org/x/sync v0.16.0/go.mod h1:1dzgHSNfp02xaA81J2MS99Qcpr2w7fw1gpm99rleRqA=
golang.org/x/sys v0.6.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.36.0 h1:KVRy2GtZBrk1cBYA7MKu5bEZFxQk4NIDV6RLVcC8o0k=
golang.org/x/sys v0.36.0/go.mod h1:OgkHotnGiDImocRcuBABYBEXf8A9a87e/uXjp9XT3ks=
golang.org/x/tools v0.36.0 h1:kWS0uv/zsvHEle1LbV5LE8QujrxB3wfQyxHfhOk0Qkg=
golang.org/x/tools v0.36.0/go.mod h1:WBDiHKJK8YgLHlcQPYQzNCkUxUypCaa5ZegCVutKm+s=
modernc.org/cc/v4 v4.26.5 h1:xM3bX7Mve6G8K8b+T11ReenJOT+BmVqQj0FY5T4+5Y4=
modernc.org/cc/v4 v4.26.5/go.mod h1:uVtb5OGqUKpoLWhqwNQo/8LwvoiEBLvZXIQ/SmO6mL0=
modernc.org/ccgo/v4 v4.28.1 h1:wPKYn5EC/mYTqBO373jKjvX2n+3+aK7+sICCv4Fjy1A=
modernc.org/ccgo/v4 v4.28.1/go.mod h1:uD+4RnfrVgE6ec9NGguUNdhqzNIeeomeXf6CL0GTE5Q=
modernc.org/fileutil v1.3.40 h1:ZGMswMNc9JOCrcrakF1HrvmergNLAmxOPjizirpfqBA=
modernc.org/fileutil v1.3.40/go.mod h1:HxmghZSZVAz/LXcMNwZPA/DRrQZEVP9VX0V4LQGQFOc=
modernc.org/gc/v2 v2.6.5 h1:nyqdV8q46KvTpZlsw66kWqwXRHdjIlJOhG6kxiV/9xI=
modernc.org/gc/v2 v2.6.5/go.mod h1:YgIahr1ypgfe7chRuJi2gD7DBQiKSLMPgBQe9oIiito=
modernc.org/goabi0 v0.2.0 h1:HvEowk7LxcPd0eq6mVOAEMai46V+i7Jrj13t4AzuNks=
modernc.org/goabi0 v0.2.0/go.mod h1:CEFRnnJhKvWT1c1JTI3Avm+tgOWbkOu5oPA8eH8LnMI=
modernc.org/libc v1.66.10 h1:yZkb3YeLx4oynyR+iUsXsybsX4Ubx7MQlSYEw4yj59A=
modernc.org/libc v1.66.10/go.mod h1:8vGSEwvoUoltr4dlywvHqjtAqHBaw0j1jI7iFBTAr2I=
modernc.org/mathutil v1.7.1 h1:GCZVGXdaN8gTqB1Mf/usp1Y/hSqgI2vAGGP4jZMCxOU=
modernc.org/mathutil v1.7.1/go.mod h1:4p5IwJITfppl0G4sUEDtCr4DthTaT47/N3aT6MhfgJg=
modernc.org/memory v1.11.0 h1:o4QC8aMQzmcwCK3t3Ux/ZHmwFPzE6hf2Y5LbkRs+hbI=
modernc.org/memory v1.11.0/go.mod h1:/JP4VbVC+K5sU2wZi9bHoq2MAkCnrt2r98UGeSK7Mjw=
modernc.org/opt v0.1.4 h1:2kNGMRiUjrp4LcaPuLY2PzUfqM/w9N23quVwhKt5Qm8=
modernc.org/opt v0.1.4/go.mod h1:03fq9lsNfvkYSfxrfUhZCWPk1lm4cq4N+Bh//bEtgns=
modernc.org/sortutil v1.2.1 h1:+xyoGf15mM3NMlPDnFqrteY07klSFxLElE2PVuWIJ7w=
modernc.org/sortutil v1.2.1/go.mod h1:7ZI3a3REbai7gzCLcotuw9AC4VZVpYMjDzETGsSMqJE=
modernc.org/sqlite v1.40.1 h1:VfuXcxcUWWKRBuP8+BR9L7VnmusMgBNNnBYGEe9w/iY=
modernc.org/sqlite v1.40.1/go.mod h1:9fjQZ0mB1LLP0GYrp39oOJXx/I2sxEnZtzCmEQIKvGE=
modernc.org/strutil v1.2.1 h1:UneZBkQA+DX2Rp35KcM69cSsNES9ly8mQWD71HKlOA0=
modernc.org/strutil v1.2.1/go.mod h1:EHkiggD70koQxjVdSBM3JKM7k6L0FbGE5eymy9i3B9A=
modernc.org/token v1.1.0 h1:Xl7Ap9dKaEs5kLoOQeQmPWevfnk/DM5qcLcYlA8ys6Y=
modernc.org/token v1.1.0/go.mod h1:UGzOrNV1mAFSEB63lOFHIpNRUVMvYTc6yu1SMY/XTDM=



================================================
FILE: install.sh
================================================
#!/bin/bash
set -e

# --- Configuration ---
REPO="alextreichler/bundleViewer"
BINARY_NAME="bundleViewer"
INSTALL_DIR="/usr/local/bin"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}>>> Starting installation for ${BINARY_NAME}...${NC}"

# --- 1. Detect Architecture & OS ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to the naming convention used in your Taskfile/Releases
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) 
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1 
        ;;
esac

# Construct the expected asset name (Must match what you upload to GitHub Releases)
# Based on your Taskfile: bundleViewer-linux-amd64 or bundleViewer-darwin-arm64
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"

echo -e "Detected Platform: ${GREEN}${OS}/${ARCH}${NC}"

# --- 2. Determine Download URL ---
# We use the 'latest' release endpoint.
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"

# --- 3. Download ---
echo -e "Downloading from: ${BLUE}${DOWNLOAD_URL}${NC}"

# Create a temporary file
TMP_FILE=$(mktemp)

if curl --fail -L --progress-bar "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo -e "${GREEN}Download successful.${NC}"
else
    echo -e "${RED}Error: Download failed.${NC}"
    echo "Check if a release exists for '${ASSET_NAME}' at https://github.com/${REPO}/releases"
    rm -f "$TMP_FILE"
    exit 1
fi

# --- 4. Install ---
echo "Installing to ${INSTALL_DIR}..."

# Check if we have write access to INSTALL_DIR, otherwise use sudo
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Sudo permissions required to write to ${INSTALL_DIR}"
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
fi

# --- 5. Verify ---
if command -v "$BINARY_NAME" >/dev/null; then
    INSTALLED_PATH=$(which "$BINARY_NAME")
    echo -e "${GREEN}Success! ${BINARY_NAME} installed to ${INSTALLED_PATH}${NC}"
    echo -e "Run '${BLUE}${BINARY_NAME} --help${NC}' to get started."
else
    echo -e "${RED}Error: Installation appeared to succeed, but binary not found in PATH.${NC}"
    echo "Please ensure ${INSTALL_DIR} is in your PATH."
fi


================================================
FILE: Taskfile.yml
================================================
version: '3'

tasks:
  build:
    desc: "Builds the bundleViewer application for Linux and macOS"
    deps:
      - build:linux
      - build:macos

  build:linux:
    desc: "Builds the bundleViewer application for Linux (amd64)"
    cmds:
      - GOOS=linux GOARCH=amd64 GOEXPERIMENT=greenteagc GOEXPERIMENT=jsonv2 go build -o dist/bundleViewer-linux-amd64 cmd/webapp/main.go
    silent: false

  build:macos:
    desc: "Builds the bundleViewer application for macOS (arm64)"
    cmds:
      - GOOS=darwin GOARCH=arm64 GOEXPERIMENT=greenteagc GOEXPERIMENT=jsonv2 go build -o dist/bundleViewer-darwin-arm64 cmd/webapp/main.go
    silent: false

  run:
    desc: "Runs the bundleViewer application. Usage: task run -- -bundlePath /path/to/bundle"
    deps:
      - db:clean
    cmds:
      - GOEXPERIMENT=greenteagc GOEXPERIMENT=jsonv2 ./bundleViewer {{.CLI_ARGS}}
    silent: false

  clean:
    desc: "Removes the compiled bundleViewer binaries and dist directory"
    cmds:
      - rm -rf dist bundleViewer bundle.db
    silent: false

  lint:
    desc: "Runs golangci-lint on the codebase"
    cmds:
      - golangci-lint run
    silent: false

  deps:
    desc: "Runs go mod tidy and vendor"
    cmds:
      - go mod tidy
      - go mod vendor
    silent: false

  release:
    desc: "Creates a GitHub release and uploads binaries. Usage: task release TAG=v1.0.0"
    deps:
      - build
    cmds:
      - |
        if [ -z "{{.TAG}}" ]; then
          echo "Error: TAG variable is required. Usage: task release TAG=v1.0.0"
          exit 1
        fi
      - echo "Creating release {{.TAG}}..."
      - gh release create {{.TAG}} dist/bundleViewer-linux-amd64 dist/bundleViewer-darwin-arm64 --title "{{.TAG}}" --notes "Release {{.TAG}}"
    silent: false

  db:clean:
    desc: "Removes the bundle.db file"
    cmds:
      - rm -f bundle.db
    silent: true

  default:
    cmds:
      - task: build



================================================
FILE: cmd/webapp/main.go
================================================
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/server"
	"github.com/alextreichler/bundleViewer/internal/store"
)

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		slog.Warn("Failed to open browser automatically", "error", err, "url", url)
	}
}

func main() {
	port := flag.Int("port", 7575, "Port to listen on")
	host := flag.String("host", "127.0.0.1", "Host to bind to (default: 127.0.0.1 for security)")
	persist := flag.Bool("persist", false, "Persist the database between runs (default: clean on start)")
	logsOnly := flag.Bool("logs-only", false, "Only process and display logs")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <bundle-directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	
	flag.Parse()

	// Initialize the logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up application data directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	dataDir := filepath.Join(cacheDir, "bundleViewer")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Error("Failed to create data directory", "path", dataDir, "error", err)
		os.Exit(1)
	}
	logger.Info("Using data directory", "path", dataDir)

	// Manually check for flags in trailing arguments (standard flag package stops at first non-flag)
	args := flag.Args()
	var bundlePath string
	
	for _, arg := range args {
		if arg == "-persist" || arg == "--persist" {
			*persist = true
		} else if arg == "-logs-only" || arg == "--logs-only" {
			*logsOnly = true
		} else if bundlePath == "" {
			bundlePath = arg
		}
	}

	var sqliteStore *store.SQLiteStore
	var cachedData *cache.CachedData

	if bundlePath != "" {
		// Determine if we should clean the DB (default true, unless -persist is set)
		cleanDB := !*persist

		// Initialize the SQLite store
		dbPath := filepath.Join(dataDir, "bundle.db")
		sqliteStore, err = store.NewSQLiteStore(dbPath, bundlePath, cleanDB)
		if err != nil {
			logger.Error("Failed to initialize SQLite store", "error", err)
			os.Exit(1)
		}

		// Initialize the cache with a temporary tracker
		initialProgress := models.NewProgressTracker(21)
		cachedData, err = cache.New(bundlePath, sqliteStore, *logsOnly, initialProgress)
		if err != nil {
			logger.Error("Failed to create cache", "error", err)
			os.Exit(1)
		}
		initialProgress.Finish()
	} else {
		logger.Info("Starting in setup mode. Please visit the web UI to configure the bundle path.")
	}

	var storeInterface store.Store
	if sqliteStore != nil {
		storeInterface = sqliteStore
	}

	srv, err := server.New(bundlePath, cachedData, storeInterface, logger, *logsOnly, *persist) // Pass logger, logsOnly, persist
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// Cast srv to the concrete type to access Shutdown
	srvInstance, ok := srv.(*server.Server)
	if ok && srvInstance != nil {
		srvInstance.SetDataDir(dataDir)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	logger.Info("Server listening on", "address", addr)

	// Automatically open browser
	go func() {
		// Wait a brief moment for server to start
		time.Sleep(200 * time.Millisecond)
		// We always open localhost in the browser for user convenience
		url := fmt.Sprintf("http://localhost:%d", *port)
		openBrowser(url)
	}()

	// Signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(addr, srv); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until signal
	<-c
	logger.Info("Shutting down...")

	// Cleanup
	if ok && srvInstance != nil {
		srvInstance.Shutdown()
	} else if sqliteStore != nil {
		sqliteStore.Close()
	}

	// Wipe the entire data directory on exit if persist is false
	if !*persist {
		logger.Info("Cleaning up database files...", "path", dataDir)
		if err := os.RemoveAll(dataDir); err != nil {
			logger.Warn("Failed to clean up data directory on exit", "path", dataDir, "error", err)
		}
	}
}



================================================
FILE: internal/analysis/log_analyzer.go
================================================
package analysis

import (
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// GenerateFingerprint creates a generic signature for a log message
// Optimized single-pass implementation
func GenerateFingerprint(message string) string {
	var b strings.Builder
	b.Grow(len(message))

	n := len(message)
	for i := 0; i < n; i++ {
		c := message[i]

		// 1. Quoted Strings: "..." or '...'
		if c == '"' || c == '\'' {
			quote := c
			b.WriteString("<STR>")
			i++
			for i < n && message[i] != quote {
				i++
			}
			continue
		}

		// 2. Numbers and Hex/UUIDs
		// If we hit a digit, we check the "word"
		if isDigit(c) {
			// Look ahead to see if this is a hex string/UUID or just a number
			hasLetter := false
			
			// Scan until non-alphanumeric
			for i < n {
				curr := message[i]
				if isDigit(curr) {
					// ok
				} else if (curr >= 'a' && curr <= 'z') || (curr >= 'A' && curr <= 'Z') || curr == '-' {
					hasLetter = true
				} else if curr == '.' {
					// 123.456 (float) or 192.168.1.1 (IP)
				} else {
					// End of word (space, bracket, etc)
					break
				}
				i++
			}
			
			// Back up one since outer loop increments
			i--
			
			if hasLetter {
				// It was mixed alphanum (UUID, Hex, "2nd", IP?)
				b.WriteString("<ID>")
			} else {
				// Pure digits/dots
				b.WriteString("<NUM>")
			}
			continue
		}

		b.WriteByte(c)
	}
	return b.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// AnalyzeLogs groups logs by fingerprint
func AnalyzeLogs(logs []*models.LogEntry) []models.LogPattern {
	groups := make(map[string]*models.LogPattern)

	for _, entry := range logs {
		// Include Level and Component in signature to separate similar errors from different sources/levels
		// Or strictly use message? Usually grouping by message pattern is enough, but Level distinction is useful.
		// Let's fingerprint the message only, but store key as Level+Fingerprint

		fp := GenerateFingerprint(entry.Message)
		key := entry.Level + "|" + fp

		if _, exists := groups[key]; !exists {
			groups[key] = &models.LogPattern{
				Signature:   fp,
				Count:       0,
				Level:       entry.Level,
				SampleEntry: *entry,
				FirstSeen:   entry.Timestamp,
				LastSeen:    entry.Timestamp,
			}
		}

		g := groups[key]
		g.Count++
		if entry.Timestamp.Before(g.FirstSeen) {
			g.FirstSeen = entry.Timestamp
		}
		if entry.Timestamp.After(g.LastSeen) {
			g.LastSeen = entry.Timestamp
		}
	}

	// Convert map to slice
	var patterns []models.LogPattern
	for _, p := range groups {
		patterns = append(patterns, *p)
	}

	// Sort by Count (Desc)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}



================================================
FILE: internal/analysis/performance.go
================================================
package analysis

import (
	"fmt"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// PerformanceReport holds the result of the performance analysis
type PerformanceReport struct {
	IsCPUBound       bool
	IsDiskBound      bool
	IsNetworkBound   bool // Hard to determine from snapshot, but we can look for errors
	WorkloadType     string // "Produce-Heavy", "Fetch-Heavy", "Balanced", "Idle"
	Recommendations  []string
	Observations     []string
}

// AnalyzePerformance evaluates the metrics bundle against known performance heuristics
func AnalyzePerformance(metrics *models.MetricsBundle) PerformanceReport {
	report := PerformanceReport{
		Recommendations: []string{},
		Observations:    []string{},
	}

	if metrics == nil {
		return report
	}

	// 1. CPU Analysis
	// Check reactor utilization (Gauge)
	maxReactorUtil := 0.0
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_reactor_utilization" {
				if metric.Value > maxReactorUtil {
					maxReactorUtil = metric.Value
				}
			}
		}
	}

	if maxReactorUtil > 95 {
		report.IsCPUBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("High Reactor Utilization detected: %.2f%%. The system is likely CPU bound.", maxReactorUtil))
	}

	// 2. Workload Classification (Produce vs Fetch)
	// We need 'vectorized_kafka_handler_latency_microseconds_count' which is a counter. 
	// In a snapshot, we can't see rate, but we can see TOTAL counts to see what the dominant historical workload has been.
	var produceCount, fetchCount float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_kafka_handler_latency_microseconds_count" {
				handler := metric.Labels["handler"]
				if handler == "produce" {
					produceCount += metric.Value
				} else if handler == "fetch" {
					fetchCount += metric.Value
				}
			}
		}
	}

	if produceCount+fetchCount > 0 {
		ratio := fetchCount / (produceCount + fetchCount)
		if ratio > 0.7 {
			report.WorkloadType = "Fetch-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Fetch-heavy (>70% of requests).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Consider increasing 'fetch_reads_debounce_timeout' to reduce CPU load from excessive fetching.")
			}
		} else if ratio < 0.3 {
			report.WorkloadType = "Produce-Heavy"
			report.Observations = append(report.Observations, "Workload is predominantly Produce-heavy (>70% of requests).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Consider increasing 'linger.ms' on clients to improve batching.")
			}
		} else {
			report.WorkloadType = "Balanced"
		}
	} else {
		report.WorkloadType = "Idle"
	}

	// 3. Disk Analysis
	// Check for IO Queue delays (Snapshot of counter is useless, but Queue Length is a Gauge)
	maxQueueLength := 0.0
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_io_queue_queue_length" {
				if metric.Value > maxQueueLength {
					maxQueueLength = metric.Value
				}
			}
		}
	}

	if maxQueueLength > 50 {
		report.IsDiskBound = true
		report.Observations = append(report.Observations, fmt.Sprintf("High Disk Queue Length detected: %.0f. Disk may be saturating.", maxQueueLength))
	}

	// 4. Batch Size Efficiency (Heuristic based on totals)
	// vectorized_storage_log_written_bytes / vectorized_storage_log_batches_written
	var totalBytes, totalBatches float64
	for _, m := range metrics.Files {
		for _, metric := range m {
			if metric.Name == "vectorized_storage_log_written_bytes" {
				totalBytes += metric.Value
			}
			if metric.Name == "vectorized_storage_log_batches_written" {
				totalBatches += metric.Value
			}
		}
	}

	if totalBatches > 0 {
		avgBatchSize := totalBytes / totalBatches
		report.Observations = append(report.Observations, fmt.Sprintf("Average Batch Size: %.2f KB", avgBatchSize/1024))
		if avgBatchSize < 10*1024 { // < 10KB
			report.Observations = append(report.Observations, "Small average batch size detected (< 10KB).")
			if report.IsCPUBound {
				report.Recommendations = append(report.Recommendations, "Small batches increase CPU overhead. Increase batch size on producers.")
			}
		}
	}

	return report
}



================================================
FILE: internal/analysis/skew_analyzer.go
================================================
package analysis

import (
	"math"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type NodeSkew struct {
	NodeID         int
	PartitionCount int
	TopicCount     int
	Deviation      float64 // Standard deviations from mean
	SkewPercentage float64 // % difference from mean
	IsSkewed       bool
}

type TopicSkew struct {
	TopicName        string
	PartitionCount   int
	NodeDistribution map[int]int
	MaxSkew          float64 // Max deviation found among nodes for this topic
	IsSkewed         bool
}

type BalancingConfig struct {
	PartitionsPerShard      int  `json:"partitions_per_shard"`
	PartitionsReserveShard0 int  `json:"partitions_reserve_shard0"`
	TopicAware              bool `json:"topic_aware"`
	RackAwarenessEnabled    bool `json:"rack_awareness_enabled"`
}

type NodeBalancingScore struct {
	NodeID      int
	MaxCapacity int
	Score       float64
	CPU         int
}

type DiskBalancingConfig struct {
	SpaceManagementEnabled                    bool
	DiskReservationPercent                    int
	RetentionLocalTargetCapacityPercent       int
	RetentionLocalTargetCapacityBytes         int64
	PartitionAutobalancingMaxDiskUsagePercent int
	StorageSpaceAlertFreeThresholdPercent     int
}

type BalancingAnalysis struct {
	Config     BalancingConfig
	DiskConfig DiskBalancingConfig
	Nodes      []NodeBalancingScore
}

type SkewAnalysis struct {
	TotalPartitions int
	TotalNodes      int
	MeanPartitions  float64
	StdDev          float64
	NodeSkews       []NodeSkew
	TopicSkews      []TopicSkew
	Balancing       BalancingAnalysis
}

func getInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func getBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getInt64(config map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := config[key].(float64); ok {
		return int64(val)
	}
	if val, ok := config[key].(int); ok {
		return int64(val)
	}
	return defaultValue
}

func AnalyzeBalancing(
	brokers []interface{},
	nodePartitionCounts map[int]int,
	nodeTopicCounts map[string]map[int]int,
	clusterConfig map[string]interface{},
) BalancingAnalysis {
	config := BalancingConfig{
		PartitionsPerShard:      getInt(clusterConfig, "topic_partitions_per_shard", 5000),
		PartitionsReserveShard0: getInt(clusterConfig, "topic_partitions_reserve_shard0", 0),
		TopicAware:              clusterConfig["partition_autobalancing_mode"] == "topic_aware",
		RackAwarenessEnabled:    getBool(clusterConfig, "enable_rack_awareness", false),
	}

	var nodeScores []NodeBalancingScore

	for _, b := range brokers {
		brokerMap, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		nodeID := int(brokerMap["node_id"].(float64))
		cores := int(brokerMap["num_cores"].(float64))

		maxCapacity := (cores * config.PartitionsPerShard) - config.PartitionsReserveShard0
		if maxCapacity <= 0 {
			maxCapacity = 1 // Avoid division by zero
		}

		var score float64
		if config.TopicAware {
			// Count unique topics on the node
			topicCount := 0
			for _, nodeMap := range nodeTopicCounts {
				if _, exists := nodeMap[nodeID]; exists {
					topicCount++
				}
			}
			score = float64(topicCount) / float64(maxCapacity)
		} else {
			score = float64(nodePartitionCounts[nodeID]) / float64(maxCapacity)
		}

		nodeScores = append(nodeScores, NodeBalancingScore{
			NodeID:      nodeID,
			MaxCapacity: maxCapacity,
			Score:       score,
			CPU:         cores,
		})
	}

	diskConfig := DiskBalancingConfig{
		SpaceManagementEnabled:                    getBool(clusterConfig, "space_management_enabled", true),
		DiskReservationPercent:                    getInt(clusterConfig, "disk_reservation_percent", 20),                // Defaulting to 20%
		RetentionLocalTargetCapacityPercent:       getInt(clusterConfig, "retention_local_target_capacity_percent", 80), // Defaulting to 80%
		RetentionLocalTargetCapacityBytes:         getInt64(clusterConfig, "retention_local_target_capacity_bytes", 0),
		PartitionAutobalancingMaxDiskUsagePercent: getInt(clusterConfig, "partition_autobalancing_max_disk_usage_percent", 80), // Defaulting to 80%
		StorageSpaceAlertFreeThresholdPercent:     getInt(clusterConfig, "storage_space_alert_free_threshold_percent", 15),     // Defaulting to 15%
	}

	return BalancingAnalysis{
		Config:     config,
		DiskConfig: diskConfig,
		Nodes:      nodeScores,
	}
}

func AnalyzePartitionSkew(partitions []models.ClusterPartition, brokers []interface{}, clusterConfig map[string]interface{}) SkewAnalysis {
	// 1. Count Partitions and Topics per Node
	nodePartitionCounts := make(map[int]int)
	// topic -> node -> count
	nodeTopicCounts := make(map[string]map[int]int)
	allNodes := make(map[int]bool)

	// Initialize all known nodes with 0
	for _, b := range brokers {
		if bMap, ok := b.(map[string]interface{}); ok {
			if id, ok := bMap["node_id"].(float64); ok {
				allNodes[int(id)] = true
				nodePartitionCounts[int(id)] = 0
			}
		}
	}

	for _, p := range partitions {
		for _, r := range p.Replicas {
			nodePartitionCounts[r.NodeID]++
			allNodes[r.NodeID] = true

			if _, ok := nodeTopicCounts[p.Topic]; !ok {
				nodeTopicCounts[p.Topic] = make(map[int]int)
			}
			nodeTopicCounts[p.Topic][r.NodeID]++
		}
	}

	totalNodes := len(allNodes)
	if totalNodes == 0 {
		return SkewAnalysis{}
	}

	totalPartitions := 0
	for _, count := range nodePartitionCounts {
		totalPartitions += count
	}

	mean := float64(totalPartitions) / float64(totalNodes)

	// Calculate Variance and StdDev
	var sumSquares float64
	for _, count := range nodePartitionCounts {
		diff := float64(count) - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(totalNodes)
	stdDev := math.Sqrt(variance)

	// Calculate Node Skews
	var nodeSkews []NodeSkew
	for nodeID, count := range nodePartitionCounts {
		deviation := 0.0
		if stdDev > 0 {
			deviation = (float64(count) - mean) / stdDev
		}

		skewPct := 0.0
		if mean > 0 {
			skewPct = ((float64(count) - mean) / mean) * 100
		}

		isSkewed := math.Abs(skewPct) > 20.0

		topicCount := 0
		for _, nodeMap := range nodeTopicCounts {
			if _, exists := nodeMap[nodeID]; exists {
				topicCount++
			}
		}

		nodeSkews = append(nodeSkews, NodeSkew{
			NodeID:         nodeID,
			PartitionCount: count,
			TopicCount:     topicCount,
			Deviation:      deviation,
			SkewPercentage: skewPct,
			IsSkewed:       isSkewed,
		})
	}

	// Calculate Topic Skews
	var topicSkews []TopicSkew
	for topic, nodeMap := range nodeTopicCounts {
		totalReplicas := 0
		maxReplicas := 0
		minReplicas := math.MaxInt32
		
		for nodeID := range allNodes {
			count := nodeMap[nodeID]
			totalReplicas += count
			if count > maxReplicas {
				maxReplicas = count
			}
			if count < minReplicas {
				minReplicas = count
			}
		}

		if totalNodes > 0 {
			meanReplicas := float64(totalReplicas) / float64(totalNodes)
			
			maxSkew := 0.0
			if meanReplicas > 0 {
				maxSkew = (float64(maxReplicas) - meanReplicas) / meanReplicas * 100
			}

			// Flag if unbalanced (max-min > 1) AND significant skew (> 20%)
			isSkewed := (maxReplicas - minReplicas) > 1 && maxSkew > 20.0

			if isSkewed {
				topicSkews = append(topicSkews, TopicSkew{
					TopicName:        topic,
					PartitionCount:   totalReplicas,
					NodeDistribution: nodeMap,
					MaxSkew:          maxSkew,
					IsSkewed:         isSkewed,
				})
			}
		}
	}

	balancingAnalysis := AnalyzeBalancing(brokers, nodePartitionCounts, nodeTopicCounts, clusterConfig)

	return SkewAnalysis{
		TotalPartitions: totalPartitions,
		TotalNodes:      totalNodes,
		MeanPartitions:  mean,
		StdDev:          stdDev,
		NodeSkews:       nodeSkews,
		TopicSkews:      topicSkews,
		Balancing:       balancingAnalysis,
	}
}



================================================
FILE: internal/cache/cache.go
================================================
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
	"github.com/alextreichler/bundleViewer/internal/store"
)

type CachedData struct {
	GroupedFiles      map[string][]parser.ParsedFile
	DataDiskFiles     []string
	CacheDiskFiles    []string
	Partitions        []models.ClusterPartition
	Leaders           []models.PartitionLeader
	KafkaMetadata     models.KafkaMetadataResponse
	TopicConfigs      models.TopicConfigsResponse
	ConsumerGroups    map[string]models.ConsumerGroup
	HealthOverview    models.HealthOverview
	ResourceUsage     models.ResourceUsage
	DataDiskStats     models.DiskStats
	CacheDiskStats    models.DiskStats
	DuEntries         []models.DuEntry
	K8sStore          models.K8sStore
	RedpandaDataDir   string
	RedpandaConfig    map[string]interface{}
	MetricDefinitions map[string]string
	System            models.SystemState
	SarData           models.SarData
	TopicDetails      map[string]models.TopicConfig
	Store             store.Store
	CoreDumps         []string
	LogsOnly          bool
}

func New(bundlePath string, s store.Store, logsOnly bool, p *models.ProgressTracker) (*CachedData, error) {
	adminPath := bundlePath + "/admin"
	utilsPath := bundlePath + "/utils"
	procPath := bundlePath + "/proc"

	dataDiskPattern := filepath.Join(bundlePath, "admin", "disk_stat_data_*.json")
	dataDiskFiles, _ := filepath.Glob(dataDiskPattern)

	cacheDiskPattern := filepath.Join(bundlePath, "admin", "disk_stat_cache_*.json")
	cacheDiskFiles, _ := filepath.Glob(cacheDiskPattern)

	var coreDumps []string
	if matches, err := filepath.Glob(filepath.Join(bundlePath, "core.*")); err == nil {
		for _, m := range matches {
			coreDumps = append(coreDumps, filepath.Base(m))
		}
	}

	var wg sync.WaitGroup
	var adminFiles, utilsFiles, procFiles, rootFiles []parser.ParsedFile
	var clusterConfigData []models.ClusterConfigEntry
	var clusterConfigError error
	var partitions []models.ClusterPartition
	var leaders []models.PartitionLeader
	var kafkaMetadataData models.KafkaMetadataResponse
	var topicConfigsData models.TopicConfigsResponse
	var consumerGroupsData map[string]models.ConsumerGroup
	var healthOverviewData models.HealthOverview
	var resourceUsageData models.ResourceUsage
	var dataDiskStatsData models.DiskStats
	var cacheDiskStatsData models.DiskStats
	var duEntriesData []models.DuEntry
	var k8sStoreData models.K8sStore
	var redpandaDataDir string
	var systemData models.SystemState
	var sarData models.SarData
	var metricDefinitionsData map[string]string

	wg.Add(19)

	go func() {
		defer wg.Done()
		p.Update(0, "Parsing SAR data...")
		sarData, _ = parser.ParseSar(bundlePath, logsOnly)
		p.Update(1, "SAR data parsed")
	}()

	go func() {
		defer wg.Done()
		p.Update(0, "Parsing system state...")
		df, _ := parser.ParseDF(bundlePath)
		free, _ := parser.ParseFree(bundlePath)
		top, _ := parser.ParseTop(bundlePath)
		sysctl, _ := parser.ParseSysctl(bundlePath)
		ntp, _ := parser.ParseNTP(bundlePath)
		ips, _ := parser.ParseIP(bundlePath)
		conns, _ := parser.ParseSS(bundlePath)
		uname, _ := parser.ParseUnameInfo(bundlePath)
		memInfo, _ := parser.ParseMemInfo(bundlePath)
		vmStat, _ := parser.ParseVMStat(bundlePath)
		cpuInfo, _ := parser.ParseCPUInfo(bundlePath)
		mdStat, _ := parser.ParseMDStat(bundlePath)
		cmdLine, _ := parser.ParseCmdLine(bundlePath)
		dmi, _ := parser.ParseDMI(bundlePath)
		dig, _ := parser.ParseDig(bundlePath)
		syslog, _ := parser.ParseSyslog(bundlePath)
		vmStatTS, _ := parser.ParseVMStatTimeSeries(bundlePath)
		lspci, _ := parser.ParseLSPCI(bundlePath)

		connSummary := models.ConnectionSummary{
			Total:   len(conns),
			ByState: make(map[string]int),
			ByPort:  make(map[string]int),
		}
		for _, c := range conns {
			connSummary.ByState[c.State]++
			if idx := strings.LastIndex(c.LocalAddr, ":"); idx != -1 {
				port := c.LocalAddr[idx+1:]
				connSummary.ByPort[port]++
			}
		}
		for port, count := range connSummary.ByPort {
			connSummary.SortedByPort = append(connSummary.SortedByPort, models.PortCount{Port: port, Count: count})
		}
		sort.Slice(connSummary.SortedByPort, func(i, j int) bool {
			return connSummary.SortedByPort[i].Count > connSummary.SortedByPort[j].Count
		})

		procPattern := filepath.Join(bundlePath, "proc", "cpuinfo-processor-*")
		coreFiles, _ := filepath.Glob(procPattern)
		coreCount := len(coreFiles)

		systemData = models.SystemState{
			FileSystems: df, Memory: free, MemInfo: memInfo, Load: top,
			Uname: uname, DMI: dmi, Dig: dig, Syslog: syslog,
			VMStatAnalysis: vmStatTS, LSPCI: lspci, CPU: cpuInfo, Sysctl: sysctl,
			NTP: ntp, Interfaces: ips, Connections: conns, ConnSummary: connSummary,
			VMStat: vmStat, MDStat: mdStat, CmdLine: cmdLine, CoreCount: coreCount,
		}
		p.Update(1, "System state parsed")
	}()

	go func() {
		defer wg.Done()
		p.Update(0, "Parsing Prometheus metrics...")
		patterns := []string{
			filepath.Join(bundlePath, "metrics", "*", "*_metrics.txt"),
			filepath.Join(bundlePath, "metrics", "*_metrics.txt"),
		}
		var metricsFiles []string
		for _, ptrn := range patterns {
			matches, _ := filepath.Glob(ptrn)
			if len(matches) > 0 {
				metricsFiles = matches
				break
			}
		}
		metricDefinitionsData = make(map[string]string)
		if len(metricsFiles) > 0 {
			if defs, err := parser.ParseMetricDefinitions(metricsFiles[0]); err == nil {
				for k, v := range defs {
					metricDefinitionsData[k] = v
				}
			}
			for _, file := range metricsFiles {
				metricTimestamp := time.Now()
				if fi, err := os.Stat(file); err == nil {
					metricTimestamp = fi.ModTime()
				}
				parser.ParsePrometheusMetrics(file, nil, s, metricTimestamp)
			}
		}
		p.Update(1, "Prometheus metrics parsed")
	}()

	go func() {
		defer wg.Done()
		p.Update(0, "Ingesting logs into SQLite...")
		hasLogs, _ := s.HasLogs()
		if !hasLogs {
			// Optimization: Drop indexes for faster bulk insert
			if sqlite, ok := s.(*store.SQLiteStore); ok {
				_ = sqlite.DropIndexes()
			}
			
			parser.ParseLogs(bundlePath, s, logsOnly, p)
			
			// Restore indexes and rebuild FTS
			if sqlite, ok := s.(*store.SQLiteStore); ok {
				p.SetStatus("Restoring indexes (this may take a minute)...")
				_ = sqlite.RestoreIndexes()
			}
		}
		p.Update(1, "Logs ingested")
	}()

	go func() {
		defer wg.Done()
		redpandaDataDir, _ = parser.ParseRedpandaDataDirectory(bundlePath)
		p.Update(1, "Data directory identified")
	}()

	go func() {
		defer wg.Done()
		redpandaConfig, _ := parser.ParseRedpandaConfig(bundlePath)
		_ = redpandaConfig // Unused locally but could be assigned to CachedData
		p.Update(1, "Configuration parsed")
	}()

	go func() {
		defer wg.Done()
		p.Update(0, "Parsing Kubernetes resources...")
		k8sStoreData, _ = parser.ParseK8sResources(bundlePath, logsOnly, s)
		p.Update(1, "Kubernetes resources parsed")
	}()

	go func() {
		defer wg.Done()
		partitions, _ = parser.ParseClusterPartitions(bundlePath)
		p.Update(1, "Partition metadata parsed")
	}()

	go func() {
		defer wg.Done()
		leaders, _ = parser.ParsePartitionLeaders(bundlePath, logsOnly)
		p.Update(1, "Leader metadata parsed")
	}()

	go func() {
		defer wg.Done()
		p.Update(0, "Parsing Kafka metadata, configs, and groups...")
		fullData, _ := parser.ParseKafkaJSON(bundlePath)
		kafkaMetadataData = fullData.Metadata
		topicConfigsData = fullData.TopicConfigs
		consumerGroupsData = fullData.ConsumerGroups
		p.Update(1, "Kafka data parsed")
	}()

	go func() {
		defer wg.Done()
		healthOverviewData, _ = parser.ParseHealthOverview(bundlePath)
		p.Update(1, "Health overview parsed")
	}()

	go func() {
		defer wg.Done()
		resourceUsageFile := filepath.Join(bundlePath, "resource-usage.json")
		if data, err := os.ReadFile(resourceUsageFile); err == nil {
			json.Unmarshal(data, &resourceUsageData)
		}
		p.Update(1, "Resource usage parsed")
	}()

	go func() {
		defer wg.Done()
		if len(dataDiskFiles) > 0 {
			if data, err := os.ReadFile(dataDiskFiles[0]); err == nil {
				json.Unmarshal(data, &dataDiskStatsData)
			}
		}
		p.Update(1, "Data disk stats parsed")
	}()

	go func() {
		defer wg.Done()
		if len(cacheDiskFiles) > 0 {
			if data, err := os.ReadFile(cacheDiskFiles[0]); err == nil {
				json.Unmarshal(data, &cacheDiskStatsData)
			}
		}
		p.Update(1, "Cache disk stats parsed")
	}()

	go func() {
		defer wg.Done()
		duEntriesData, _ = parser.ParseDuOutput(bundlePath, logsOnly)
		p.Update(1, "Disk usage (DU) parsed")
	}()

	go func() {
		defer wg.Done()
		allAdminFiles, _ := parser.ParseAllFiles(adminPath, []string{".json"})
		for _, file := range allAdminFiles {
			if file.FileName == "cluster_config.json" {
				clusterConfigData, clusterConfigError = parser.ParseClusterConfig(bundlePath)
			} else {
				adminFiles = append(adminFiles, file)
			}
		}
		p.Update(1, "Admin directory parsed")
	}()

	go func() {
		defer wg.Done()
		utilsFiles, _ = parser.ParseAllFiles(utilsPath, []string{".txt"})
		p.Update(1, "Utils directory parsed")
	}()

	go func() {
		defer wg.Done()
		allProcFiles, _ := parser.ParseAllFiles(procPath, []string{""})
		for _, file := range allProcFiles {
			if !strings.HasPrefix(file.FileName, "cpuinfo-processor-") {
				procFiles = append(procFiles, file)
			}
		}
		p.Update(1, "Proc directory parsed")
	}()

	go func() {
		defer wg.Done()
		rootFilesToParse := []string{"errors.txt", "kafka.json", "redpanda.yaml", "resource-usage.json"}
		rootFiles, _ = parser.ParseSpecificFiles(bundlePath, rootFilesToParse)
		p.Update(1, "Root files parsed")
	}()

	wg.Wait()

	p.SetStatus("Optimizing database and building indexes...")
	s.Optimize()
	p.Update(1, "Database optimized")

	sort.Slice(adminFiles, func(i, j int) bool { return adminFiles[i].FileName < adminFiles[j].FileName })
	sort.Slice(utilsFiles, func(i, j int) bool { return utilsFiles[i].FileName < utilsFiles[j].FileName })
	sort.Slice(procFiles, func(i, j int) bool { return procFiles[i].FileName < procFiles[j].FileName })
	sort.Slice(rootFiles, func(i, j int) bool { return rootFiles[i].FileName < rootFiles[j].FileName })

	groupedFiles := map[string][]parser.ParsedFile{
		"Root": rootFiles, "Admin": adminFiles, "Proc": procFiles, "Utils": utilsFiles,
	}
	if clusterConfigError != nil {
		groupedFiles["Admin"] = append(groupedFiles["Admin"], parser.ParsedFile{FileName: "cluster_config.json", Error: clusterConfigError})
	} else if len(clusterConfigData) > 0 {
		groupedFiles["Admin"] = append(groupedFiles["Admin"], parser.ParsedFile{FileName: "cluster_config.json", Data: clusterConfigData})
	}

	topicDetails := make(map[string]models.TopicConfig)
	for _, tc := range topicConfigsData {
		topicDetails[tc.Name] = tc
	}

	redpandaConfig, _ := parser.ParseRedpandaConfig(bundlePath)

	return &CachedData{
		GroupedFiles: groupedFiles, DataDiskFiles: dataDiskFiles, CacheDiskFiles: cacheDiskFiles,
		Partitions: partitions, Leaders: leaders, KafkaMetadata: kafkaMetadataData,
		TopicConfigs: topicConfigsData, ConsumerGroups: consumerGroupsData, HealthOverview: healthOverviewData, ResourceUsage: resourceUsageData,
		DataDiskStats: dataDiskStatsData, CacheDiskStats: cacheDiskStatsData, DuEntries: duEntriesData,
		K8sStore: k8sStoreData, RedpandaDataDir: redpandaDataDir, RedpandaConfig: redpandaConfig,
		MetricDefinitions: metricDefinitionsData, System: systemData, SarData: sarData,
		TopicDetails: topicDetails, Store: s, CoreDumps: coreDumps, LogsOnly: logsOnly,
	}, nil
}



================================================
FILE: internal/diagnostics/auditor.go
================================================
package diagnostics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityWarning  Severity = "Warning"
	SeverityInfo     Severity = "Info"
	SeverityPass     Severity = "Pass"
)

type CheckResult struct {
	Category      string   `json:"category"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	CurrentValue  string   `json:"current_value"`
	ExpectedValue string   `json:"expected_value"`
	Status        Severity `json:"status"`
	Remediation   string   `json:"remediation"`
}

type DiagnosticsReport struct {
	Results []CheckResult
}

// Audit performs all diagnostic checks on the cached data
func Audit(data *cache.CachedData, s store.Store) DiagnosticsReport {
	var results []CheckResult

	// 1. OS Tuning Checks (sysctl)
	results = append(results, checkSysctl(data.System.Sysctl)...)

	// 2. Redpanda Configuration Checks
	results = append(results, checkRedpandaConfig(data.RedpandaConfig)...)

	// 3. Disk Checks
	results = append(results, checkDisks(data)...)

	// 4. Resource Checks
	results = append(results, checkResources(data)...)

	// 5. Network State Checks
	results = append(results, checkNetwork(data)...)

	// 6. Bottleneck Checks
	results = append(results, checkBottlenecks(s)...)

	// 7. Crash Dump Checks (Critical)
	results = append(results, checkCrashDumps(data)...)

	// 8. K8s Pod Lifecycle (OOM/Crash)
	results = append(results, checkK8sEvents(data)...)

	// 9. Hardware & System Integrity (RAID, Virt, OOMs)
	results = append(results, checkHardware(data)...)

	return DiagnosticsReport{Results: results}
}

func checkHardware(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// 1. Check for OOM Events in Syslog
	if len(data.System.Syslog.OOMEvents) > 0 {
		results = append(results, CheckResult{
			Category:      "System Stability",
			Name:          "OOM Killer Invoked",
			Description:   "The kernel terminated processes to reclaim memory.",
			CurrentValue:  fmt.Sprintf("%d Events", len(data.System.Syslog.OOMEvents)),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "System is under-provisioned for memory. Investigate 'Extended Mem' in System page or reduce workload.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "System Stability",
			Name:          "OOM Killer Invoked",
			Description:   "The kernel terminated processes to reclaim memory.",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	// 2. Check RAID Status
	for _, array := range data.System.MDStat.Arrays {
		if strings.Contains(array.Status, "_") { // e.g. [_U]
			results = append(results, CheckResult{
				Category:      "Hardware Integrity",
				Name:          fmt.Sprintf("RAID %s Status", array.Name),
				Description:   "Software RAID Array Status",
				CurrentValue:  array.Status,
				ExpectedValue: "[UU...]",
				Status:        SeverityCritical,
				Remediation:   fmt.Sprintf("Array %s is degraded. Replace failed drive immediately.", array.Name),
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Hardware Integrity",
				Name:          fmt.Sprintf("RAID %s Status", array.Name),
				Description:   "Software RAID Array Status",
				CurrentValue:  "Healthy",
				ExpectedValue: "Healthy",
				Status:        SeverityPass,
			})
		}
	}

	// 3. Virtualization Warning
	isVirt := false
	if strings.Contains(strings.ToLower(data.System.DMI.Manufacturer), "vmware") || 
	   strings.Contains(strings.ToLower(data.System.DMI.Product), "vmware") {
		isVirt = true
	}
	
	if isVirt {
		results = append(results, CheckResult{
			Category:      "Platform",
			Name:          "Virtualization Detected",
			Description:   "Running on VMware/Virtual Platform",
			CurrentValue:  data.System.DMI.Manufacturer,
			ExpectedValue: "Bare Metal",
			Status:        SeverityWarning,
			Remediation:   "Ensure disk latency is low (<10ms). Avoid 'EagerZeroedThick' if possible? Actually, prefer EagerZeroedThick. Check IO Wait.",
		})
	}

	return results
}

func checkK8sEvents(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check Pod Statuses for OOMKilled or Errors
	for _, pod := range data.K8sStore.Pods {
		for _, status := range pod.Status.ContainerStatuses {
			if terminated, ok := status.LastState["terminated"]; ok {
				switch terminated.Reason {
				case "OOMKilled":
					results = append(results, CheckResult{
						Category:      "Kubernetes Events",
						Name:          fmt.Sprintf("Pod %s OOMKilled", pod.Metadata.Name),
						Description:   "Pod terminated due to Out Of Memory",
						CurrentValue:  fmt.Sprintf("Exit Code: %d, Finished At: %s", terminated.ExitCode, terminated.FinishedAt),
						ExpectedValue: "Running",
						Status:        SeverityCritical,
						Remediation:   "Increase container memory limit or investigate memory leak.",
					})
				case "Error":
					results = append(results, CheckResult{
						Category:      "Kubernetes Events",
						Name:          fmt.Sprintf("Pod %s Crash", pod.Metadata.Name),
						Description:   "Pod terminated with Error",
						CurrentValue:  fmt.Sprintf("Exit Code: %d, Reason: %s", terminated.ExitCode, terminated.Reason),
						ExpectedValue: "Running",
						Status:        SeverityCritical,
						Remediation:   "Check logs around the time of termination.",
					})
				}
			}
		}
	}

	return results
}

func checkCrashDumps(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	if len(data.CoreDumps) > 0 {
		results = append(results, CheckResult{
			Category:      "Crash Evidence",
			Name:          "Core Dumps Found",
			Description:   "Presence of core dump files indicating a crash",
			CurrentValue:  fmt.Sprintf("%d files found: %s", len(data.CoreDumps), strings.Join(data.CoreDumps, ", ")),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "Inspect core dumps with GDB or provide to support. A crash has occurred.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Crash Evidence",
			Name:          "Core Dumps Found",
			Description:   "Presence of core dump files indicating a crash",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	return results
}

func checkNetwork(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for high TIME_WAIT connections
	// This "proves" the need for tcp_tw_reuse
	timeWaitCount := data.System.ConnSummary.ByState["TIME-WAIT"] // Key from ss output is usually TIME-WAIT
	
	// ss output state format: ESTAB, TIME-WAIT, etc.
	// We need to match the exact string from parser. Let's assume standard ss output.
	// If the map key is empty/different, this check does nothing safe.
	
	// Also check common variations just in case
	if val, ok := data.System.ConnSummary.ByState["TIME_WAIT"]; ok {
		timeWaitCount += val
	}

	if timeWaitCount > 10000 {
		results = append(results, CheckResult{
			Category:      "Network Evidence",
			Name:          "TIME_WAIT Sockets",
			Description:   "Count of sockets in TIME_WAIT state",
			CurrentValue:  fmt.Sprintf("%d", timeWaitCount),
			ExpectedValue: "< 10000",
			Status:        SeverityWarning,
			Remediation:   "High TIME_WAIT count confirms the need for net.ipv4.tcp_tw_reuse = 1",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Network Evidence",
			Name:          "TIME_WAIT Sockets",
			Description:   "Count of sockets in TIME_WAIT state",
			CurrentValue:  fmt.Sprintf("%d", timeWaitCount),
			ExpectedValue: "< 10000",
			Status:        SeverityPass,
		})
	}

	return results
}

func checkSysctl(sysctl map[string]string) []CheckResult {
	var results []CheckResult

	// Helper to check integer values
	checkInt := func(key string, minVal int, severity Severity, desc, remediation string) {
		valStr, exists := sysctl[key]
		if !exists {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  "Not Found",
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        SeverityWarning,
				Remediation:   remediation,
			})
			return
		}

		val, err := strconv.Atoi(valStr)
		if err != nil {
			return // Skip if not int
		}

		if val < minVal {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  valStr,
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        severity,
				Remediation:   remediation,
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          key,
				Description:   desc,
				CurrentValue:  valStr,
				ExpectedValue: fmt.Sprintf(">= %d", minVal),
				Status:        SeverityPass,
			})
		}
	}

	checkInt("fs.aio-max-nr", 1048576, SeverityCritical, "Maximum number of concurrent async I/O requests", "Increase fs.aio-max-nr in /etc/sysctl.conf")
	checkInt("net.core.somaxconn", 4096, SeverityWarning, "Max socket listen backlog", "Increase net.core.somaxconn to handle burst connections")
	checkInt("net.core.netdev_max_backlog", 2500, SeverityWarning, "Max packets queued on input interface", "Increase net.core.netdev_max_backlog")
	checkInt("net.ipv4.tcp_max_syn_backlog", 4096, SeverityWarning, "Max TCP SYN backlog", "Increase net.ipv4.tcp_max_syn_backlog to prevent dropped connections during bursts")

	// TCP Buffer Sizes (Read/Write)
	// format: "min default max"
	// We check the max value (3rd field)
	checkBuffer := func(key string, minMaxVal int, desc string) {
		valStr, exists := sysctl[key]
		if !exists {
			return
		}
		fields := strings.Fields(valStr)
		if len(fields) == 3 {
			maxVal, _ := strconv.Atoi(fields[2])
			if maxVal < minMaxVal {
				results = append(results, CheckResult{
					Category:      "OS Tuning",
					Name:          key,
					Description:   desc,
					CurrentValue:  valStr,
					ExpectedValue: fmt.Sprintf("Max >= %d", minMaxVal),
					Status:        SeverityInfo,
					Remediation:   fmt.Sprintf("Increase max %s to %d for high throughput (Tiered Storage/Recovery)", key, minMaxVal),
				})
			}
		}
	}

	// 16MB minimum for max TCP buffer
	checkBuffer("net.ipv4.tcp_rmem", 16777216, "TCP Read Memory (min default max)")
	checkBuffer("net.ipv4.tcp_wmem", 16777216, "TCP Write Memory (min default max)")

	// UDP/Core Buffers (for Gossip)
	checkInt("net.core.rmem_max", 2097152, SeverityInfo, "Max OS receive buffer size (affects UDP/Gossip)", "Increase net.core.rmem_max to 2MB+ to prevent gossip packet loss")
	checkInt("net.core.wmem_max", 2097152, SeverityInfo, "Max OS send buffer size (affects UDP/Gossip)", "Increase net.core.wmem_max to 2MB+")

	// TCP Timestamps should be enabled for safe TCP TIME-WAIT reuse
	checkInt("net.ipv4.tcp_timestamps", 1, SeverityInfo, "TCP Timestamps (required for tcp_tw_reuse safety)", "Ensure net.ipv4.tcp_timestamps is 1. If disabled, tcp_tw_reuse can cause data corruption under NAT.")

	// TCP Window Scaling (Critical for throughput)
	checkInt("net.ipv4.tcp_window_scaling", 1, SeverityCritical, "TCP Window Scaling", "Enable tcp_window_scaling to allow TCP window sizes > 64KB")

	// TCP SACK (Critical for loss recovery)
	checkInt("net.ipv4.tcp_sack", 1, SeverityWarning, "TCP Selective Acknowledgments", "Enable tcp_sack to improve throughput in lossy networks")

	// Ephemeral Port Range
	if valStr, exists := sysctl["net.ipv4.ip_local_port_range"]; exists {
		fields := strings.Fields(valStr)
		if len(fields) == 2 {
			minPort, _ := strconv.Atoi(fields[0])
			maxPort, _ := strconv.Atoi(fields[1])
			count := maxPort - minPort
			if count < 28000 {
				results = append(results, CheckResult{
					Category:      "OS Tuning",
					Name:          "net.ipv4.ip_local_port_range",
					Description:   "Range of ephemeral ports for outgoing connections",
					CurrentValue:  valStr,
					ExpectedValue: "> 28000 ports",
					Status:        SeverityWarning,
					Remediation:   "Widen ip_local_port_range (e.g., '1024 65535') to prevent port exhaustion",
				})
			}
		}
	}

	// TCP Slow Start after idle should be 0
	if valStr, exists := sysctl["net.ipv4.tcp_slow_start_after_idle"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val != 0 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_slow_start_after_idle",
				Description:   "TCP slow start after idle",
				CurrentValue:  valStr,
				ExpectedValue: "0",
				Status:        SeverityWarning,
				Remediation:   "Set net.ipv4.tcp_slow_start_after_idle to 0 to improve latency for bursty traffic",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_slow_start_after_idle",
				Description:   "TCP slow start after idle",
				CurrentValue:  valStr,
				ExpectedValue: "0",
				Status:        SeverityPass,
			})
		}
	}

	// TCP TW Reuse should be 1 (enabled)
	if valStr, exists := sysctl["net.ipv4.tcp_tw_reuse"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val != 1 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_tw_reuse",
				Description:   "Allow reuse of TIME-WAIT sockets",
				CurrentValue:  valStr,
				ExpectedValue: "1",
				Status:        SeverityInfo,
				Remediation:   "Enable net.ipv4.tcp_tw_reuse to efficiently reuse connections",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "net.ipv4.tcp_tw_reuse",
				Description:   "Allow reuse of TIME-WAIT sockets",
				CurrentValue:  valStr,
				ExpectedValue: "1",
				Status:        SeverityPass,
			})
		}
	}
	
	// Swappiness should be 0 or 1
	if valStr, exists := sysctl["vm.swappiness"]; exists {
		val, _ := strconv.Atoi(valStr)
		if val > 1 {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "vm.swappiness",
				Description:   "Tendency to swap memory to disk",
				CurrentValue:  valStr,
				ExpectedValue: "<= 1",
				Status:        SeverityWarning,
				Remediation:   "Set vm.swappiness to 0 or 1 to prevent latency spikes",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "OS Tuning",
				Name:          "vm.swappiness",
				Description:   "Tendency to swap memory to disk",
				CurrentValue:  valStr,
				ExpectedValue: "<= 1",
				Status:        SeverityPass,
			})
		}
	}

	return results
}

func checkBottlenecks(s store.Store) []CheckResult {
	var results []CheckResult

	// Metrics required for performance analysis
	metricNames := []string{
		"vectorized_reactor_utilization",
		"vectorized_kafka_handler_latency_microseconds_count",
		"vectorized_io_queue_queue_length",
		"vectorized_storage_log_written_bytes",
		"vectorized_storage_log_batches_written",
	}

	// Use a wide time range to get all metrics
	now := time.Now()
	startTime := now.Add(-30 * 24 * time.Hour)
	endTime := now.Add(24 * time.Hour)

	metricsBundle := &models.MetricsBundle{
		Files: make(map[string][]models.PrometheusMetric),
	}
	
	// Create a "virtual" file to hold all fetched metrics
	var allMetrics []models.PrometheusMetric

	for _, name := range metricNames {
		metrics, err := s.GetMetrics(name, nil, startTime, endTime, 10000, 0)
		if err == nil {
			for _, m := range metrics {
				if m != nil {
					allMetrics = append(allMetrics, *m)
				}
			}
		}
	}
	metricsBundle.Files["store_metrics"] = allMetrics

	// Run Analysis
	report := analysis.AnalyzePerformance(metricsBundle)

	// Convert Report to CheckResults

	// CPU
	if report.IsCPUBound {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "CPU Saturation",
			Description:   "System is CPU bound (High Reactor Utilization)",
			CurrentValue:  "Critical",
			ExpectedValue: "Healthy",
			Status:        SeverityCritical,
			Remediation:   "Scale up CPU resources or optimize workload.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "CPU Saturation",
			Description:   "System is CPU bound (High Reactor Utilization)",
			CurrentValue:  "Healthy",
			ExpectedValue: "Healthy",
			Status:        SeverityPass,
		})
	}

	// Disk
	if report.IsDiskBound {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "Disk Saturation",
			Description:   "System is Disk bound (High IO Queue Length)",
			CurrentValue:  "Critical",
			ExpectedValue: "Healthy",
			Status:        SeverityWarning,
			Remediation:   "Check disk IOPS/throughput. Workload may require faster disks.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Performance Bottleneck",
			Name:          "Disk Saturation",
			Description:   "System is Disk bound (High IO Queue Length)",
			CurrentValue:  "Healthy",
			ExpectedValue: "Healthy",
			Status:        SeverityPass,
		})
	}

	// Workload Insights
	results = append(results, CheckResult{
		Category:      "Workload Characterization",
		Name:          "Workload Type",
		Description:   "Dominant request type (Produce vs Fetch)",
		CurrentValue:  report.WorkloadType,
		ExpectedValue: "N/A",
		Status:        SeverityInfo,
		Remediation:   strings.Join(report.Recommendations, " "),
	})

	// Other Observations
	for _, obs := range report.Observations {
		// Heuristic to detect "Small batch" observation for warning
		severity := SeverityInfo
		if strings.Contains(obs, "Small average batch size") {
			severity = SeverityWarning
		}

		results = append(results, CheckResult{
			Category:      "Performance Insights",
			Name:          "Observation",
			Description:   "Derived insight from metrics",
			CurrentValue:  obs,
			ExpectedValue: "N/A",
			Status:        severity,
		})
	}

	return results
}


func checkRedpandaConfig(config map[string]interface{}) []CheckResult {
	var results []CheckResult

	// Check if production mode is enabled (often implied by certain settings, but we can check specific keys)
	// For now, let's check basic sanity

	if val, ok := config["redpanda.developer_mode"]; ok {
		if isDev, _ := val.(bool); isDev {
			results = append(results, CheckResult{
				Category:      "Redpanda Config",
				Name:          "developer_mode",
				Description:   "Developer mode bypasses many checks",
				CurrentValue:  "true",
				ExpectedValue: "false",
				Status:        SeverityWarning,
				Remediation:   "Disable developer_mode for production clusters",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Redpanda Config",
				Name:          "developer_mode",
				Description:   "Developer mode",
				CurrentValue:  "false",
				ExpectedValue: "false",
				Status:        SeverityPass,
			})
		}
	}

	return results
}

func checkDisks(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for high disk usage and Filesystem Type
	for _, df := range data.System.FileSystems {
		// Filter for likely data partitions (mounted on /var/lib/redpanda or similar, or just large ones)
		// We'll check all that look like physical disks (starting with /dev/) or likely data mounts
		if !strings.HasPrefix(df.Filesystem, "/dev/") && !strings.Contains(df.MountPoint, "redpanda") {
			continue
		}

		// Check Filesystem Type (XFS preferred)
		if df.Type != "xfs" {
			results = append(results, CheckResult{
				Category:      "Disk Configuration",
				Name:          df.MountPoint + " Filesystem",
				Description:   "Filesystem type for data directory",
				CurrentValue:  df.Type,
				ExpectedValue: "xfs",
				Status:        SeverityWarning,
				Remediation:   "Redpanda is optimized for XFS. Consider reformatting with XFS for better performance.",
			})
		} else {
			results = append(results, CheckResult{
				Category:      "Disk Configuration",
				Name:          df.MountPoint + " Filesystem",
				Description:   "Filesystem type for data directory",
				CurrentValue:  df.Type,
				ExpectedValue: "xfs",
				Status:        SeverityPass,
			})
		}
		
		usePctStr := strings.TrimSuffix(df.UsePercent, "%")
		usePct, _ := strconv.Atoi(usePctStr)

		if usePct > 85 {
			severity := SeverityWarning
			if usePct > 95 {
				severity = SeverityCritical
			}
			results = append(results, CheckResult{
				Category:      "Disk Usage",
				Name:          df.MountPoint,
				Description:   "Disk space usage",
				CurrentValue:  df.UsePercent,
				ExpectedValue: "< 85%",
				Status:        severity,
				Remediation:   "Free up space or expand disk capacity",
			})
		}
	}

	return results
}

func checkResources(data *cache.CachedData) []CheckResult {
	var results []CheckResult

	// Check for under-replicated partitions
	urp := data.HealthOverview.UnderReplicatedCount
	if urp > 0 {
		results = append(results, CheckResult{
			Category:      "Cluster Health",
			Name:          "Under Replicated Partitions",
			Description:   "Partitions not fully replicated",
			CurrentValue:  fmt.Sprintf("%d", urp),
			ExpectedValue: "0",
			Status:        SeverityCritical,
			Remediation:   "Check for offline nodes or network partitions. Investigate logs for replication errors.",
		})
	} else {
		results = append(results, CheckResult{
			Category:      "Cluster Health",
			Name:          "Under Replicated Partitions",
			Description:   "Partitions not fully replicated",
			CurrentValue:  "0",
			ExpectedValue: "0",
			Status:        SeverityPass,
		})
	}

	return results
}



================================================
FILE: internal/models/cluster.go
================================================
package models

// Broker represents a broker in the cluster, derived from kafka.json or brokers.json
type Broker struct {
	NodeID int     `json:"node_id"`
	Port   int     `json:"port"`
	Host   string  `json:"host"`
	Rack   *string `json:"rack"` // Use pointer for optional field
}

// ClusterConfigEntry represents a single configuration entry from cluster_config.json
type ClusterConfigEntry struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	DocLink string      `json:"doc_link"`
}

type HealthOverview struct {
	IsHealthy            bool  `json:"is_healthy"`
	UnderReplicatedCount int   `json:"under_replicated_count"`
	LeaderlessCount      int   `json:"leaderless_count"`
	NodesDown            []int `json:"nodes_down"`
	AllNodes             []int `json:"all_nodes"`
}



================================================
FILE: internal/models/disk.go
================================================
package models

// DuEntry represents a single line from the du.txt output.
type DuEntry struct {
	Size int64
	Path string
}

// DiskStats represents the data from disk_stat_data/cache.json
type DiskStats struct {
	FreeBytes  int64 `json:"free_bytes"`
	TotalBytes int64 `json:"total_bytes"`
}

// ResourceUsage represents the data from resource-usage.json
type ResourceUsage struct {
	CPUPercentage float64 `json:"cpuPercentage"`
	FreeMemoryMB  float64 `json:"freeMemoryMB"`
	FreeSpaceMB   float64 `json:"freeSpaceMB"`
}



================================================
FILE: internal/models/k8s.go
================================================
package models

import "time"

type K8sList struct {
	Kind  string        `json:"kind"`
	Items []K8sResource `json:"items"`
}

type K8sResource struct {
	Metadata       K8sMetadata       `json:"metadata"`
	Status         K8sStatus         `json:"status,omitempty"`
	Spec           K8sSpec           `json:"spec,omitempty"`
	InvolvedObject K8sInvolvedObject `json:"involvedObject,omitempty"`
	Reason         string            `json:"reason,omitempty"`
	Message        string            `json:"message,omitempty"`
	Source         K8sSource         `json:"source,omitempty"`
	FirstTimestamp time.Time         `json:"firstTimestamp,omitempty"`
	LastTimestamp  time.Time         `json:"lastTimestamp,omitempty"`
	Count          int               `json:"count,omitempty"`
	Type           string            `json:"type,omitempty"`
	Data           map[string]string `json:"data,omitempty"`
}

type K8sSpec struct {
	NodeName         string                 `json:"nodeName,omitempty"`
	Containers       []K8sContainer         `json:"containers,omitempty"`
	StorageClassName string                 `json:"storageClassName,omitempty"`
	AccessModes      []string               `json:"accessModes,omitempty"`
	Resources        K8sResourceRequirements `json:"resources,omitempty"`
	Type             string                 `json:"type,omitempty"`      // Service type
	ClusterIP        string                 `json:"clusterIP,omitempty"` // Service IP
	Ports            []K8sServicePort       `json:"ports,omitempty"`     // Service ports
	Replicas         int                    `json:"replicas,omitempty"`  // Controller replicas
}

type K8sContainer struct {
	Name           string                  `json:"name"`
	Image          string                  `json:"image"`
	Resources      K8sResourceRequirements `json:"resources,omitempty"`
	LivenessProbe  *K8sProbe               `json:"livenessProbe,omitempty"`
	ReadinessProbe *K8sProbe               `json:"readinessProbe,omitempty"`
	StartupProbe   *K8sProbe               `json:"startupProbe,omitempty"`
}

type K8sProbe struct {
	InitialDelaySeconds int            `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds      int            `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       int            `json:"periodSeconds,omitempty"`
	SuccessThreshold    int            `json:"successThreshold,omitempty"`
	FailureThreshold    int            `json:"failureThreshold,omitempty"`
	HTTPGet             *K8sHTTPGet    `json:"httpGet,omitempty"`
	Exec                *K8sExecAction `json:"exec,omitempty"`
	TCPSocket           *K8sTCPSocket  `json:"tcpSocket,omitempty"`
}

type K8sHTTPGet struct {
	Path string      `json:"path,omitempty"`
	Port interface{} `json:"port"`
}

type K8sExecAction struct {
	Command []string `json:"command,omitempty"`
}

type K8sTCPSocket struct {
	Port interface{} `json:"port"`
}

type K8sResourceRequirements struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}

type K8sServicePort struct {
	Name       string      `json:"name,omitempty"`
	Port       int         `json:"port"`
	TargetPort interface{} `json:"targetPort,omitempty"`
	Protocol   string      `json:"protocol,omitempty"`
}

type K8sStatus struct {
	Phase             string               `json:"phase,omitempty"`
	PodIP             string               `json:"podIP,omitempty"`
	ContainerStatuses []K8sContainerStatus `json:"containerStatuses,omitempty"`
	Conditions        []K8sCondition       `json:"conditions,omitempty"`
	Capacity          map[string]string    `json:"capacity,omitempty"`    // PVC/Node Capacity
	Allocatable       map[string]string    `json:"allocatable,omitempty"` // Node Allocatable
	AccessModes       []string             `json:"accessModes,omitempty"`
	Replicas          int                  `json:"replicas,omitempty"`      // Controller current replicas
	ReadyReplicas     int                  `json:"readyReplicas,omitempty"` // Controller ready replicas
}

type K8sMetadata struct {
	Name              string               `json:"name"`
	Namespace         string               `json:"namespace"`
	CreationTimestamp time.Time            `json:"creationTimestamp"`
	Labels            map[string]string    `json:"labels"`
	OwnerReferences   []K8sOwnerReference  `json:"ownerReferences,omitempty"`
}

type K8sOwnerReference struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type K8sContainerStatus struct {
	Name         string                    `json:"name"`
	State        map[string]K8sStateDetail `json:"state"`
	LastState    map[string]K8sStateDetail `json:"lastState"`
	Ready        bool                      `json:"ready"`
	RestartCount int                       `json:"restartCount"`
	Image        string                    `json:"image"`
}

type K8sStateDetail struct {
	Reason     string    `json:"reason,omitempty"`
	Message    string    `json:"message,omitempty"`
	StartedAt  time.Time `json:"startedAt,omitempty"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	ExitCode   int       `json:"exitCode,omitempty"`
}

type K8sCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type K8sInvolvedObject struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type K8sSource struct {
	Component string `json:"component"`
	Host      string `json:"host,omitempty"`
}

type K8sStore struct {
	Pods                   []K8sResource
	Services               []K8sResource
	ConfigMaps             []K8sResource
	Endpoints              []K8sResource
	Events                 []K8sResource
	LimitRanges            []K8sResource
	PersistentVolumeClaims []K8sResource
	ReplicationControllers []K8sResource
	ResourceQuotas         []K8sResource
	ServiceAccounts        []K8sResource
	StatefulSets           []K8sResource
	Deployments            []K8sResource
	Nodes                  []K8sResource
}



================================================
FILE: internal/models/kafka.go
================================================
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





================================================
FILE: internal/models/logs.go
================================================
package models

import "time"

type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Node       string    `json:"node"`
	Shard      string    `json:"shard"`
	Component  string    `json:"component"`
	Message    string    `json:"message"`
	Raw        string    `json:"raw"` // Keep raw for full context if needed
	LineNumber int       `json:"lineNumber"`
	FilePath   string    `json:"filePath"`
	Fingerprint string   `json:"fingerprint"`
	Snippet     string   `json:"snippet"` // FTS snippet with highlighting
}

type LogFilter struct {
	Level     []string
	Node      []string
	Component []string
	Search    string
	Ignore    string
	StartTime string
	EndTime   string
	Sort      string // "asc" or "desc"
	Limit     int
	Offset    int
}

type LogPattern struct {
	Signature   string    `json:"signature"`
	Count       int       `json:"count"`
	Level       string    `json:"level"`
	SampleEntry LogEntry  `json:"sample_entry"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}



================================================
FILE: internal/models/metrics.go
================================================
package models

import "time"

// PrometheusMetric represents a single metric sample
type PrometheusMetric struct {
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricsBundle holds parsed metrics from a file or set of files
type MetricsBundle struct {
	// Map of filename -> metrics
	Files map[string][]PrometheusMetric
}



================================================
FILE: internal/models/partitions.go
================================================
package models

// PartitionInfo holds information about a single partition.
type PartitionInfo struct {
	ID       int   `json:"id"`
	Leader   int   `json:"leader"`
	Replicas []int `json:"replicas"`
}

// PartitionLeader represents the leader of a single partition.
type PartitionLeader struct {
	Ns          string `json:"ns"`
	Topic       string `json:"topic"`
	PartitionID int    `json:"partition_id"`
	Leader      int    `json:"leader"`
}

// ClusterPartition represents a single partition in the cluster.
type ClusterPartition struct {
	Ns          string `json:"ns"`
	Topic       string `json:"topic"`
	PartitionID int    `json:"partition_id"`
	Replicas    []struct {
		NodeID int `json:"node_id"`
		Core   int `json:"core"`
	} `json:"replicas"`
	Disabled bool `json:"disabled"`
	Leader   int  `json:"leader"` // Add Leader field
}



================================================
FILE: internal/models/progress.go
================================================
package models

import (
	"sync"
)

type ProgressTracker struct {
	mu       sync.RWMutex
	Current  int
	Total    int
	Status   string
	Finished bool
}

func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		Total:  total,
		Status: "Initializing...",
	}
}

func (p *ProgressTracker) Update(increment int, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current += increment
	if status != "" {
		p.Status = status
	}
	if p.Current > p.Total {
		p.Current = p.Total
	}
}

func (p *ProgressTracker) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
}

func (p *ProgressTracker) Get() (int, string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	percentage := 0
	if p.Total > 0 {
		percentage = (p.Current * 100) / p.Total
	}
	return percentage, p.Status, p.Finished
}

func (p *ProgressTracker) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current = p.Total
	p.Finished = true
	p.Status = "Complete"
}



================================================
FILE: internal/models/sar.go
================================================
package models

import "time"

type SarEntry struct {
	Timestamp time.Time `json:"timestamp"`
	CPUUser   float64   `json:"cpu_user"`
	CPUSystem float64   `json:"cpu_system"`
	CPUIdle   float64   `json:"cpu_idle"`
	MemUsed   float64   `json:"mem_used"` // Percent
}

type SarData struct {
	Entries []SarEntry
}



================================================
FILE: internal/models/segment.go
================================================
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



================================================
FILE: internal/models/system.go
================================================
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



================================================
FILE: internal/models/topic_configs.go
================================================
package models

type TopicConfigsResponse []TopicConfig

type TopicConfig struct {
	Name       string        `json:"Name"`
	Err        interface{}   `json:"Err"`
	ErrMessage string        `json:"ErrMessage"`
	Configs    []ConfigEntry `json:"Configs"`
}

type ConfigEntry struct {
	Key       string          `json:"Key"`
	Value     string          `json:"Value"`
	Source    string          `json:"Source"`
	Sensitive bool            `json:"Sensitive"`
	Synonyms  []ConfigSynonym `json:"Synonyms"`
}

type ConfigSynonym struct {
	Key    string `json:"Key"`
	Value  string `json:"Value"`
	Source string `json:"Source"`
}



================================================
FILE: internal/parser/parser.go
================================================
package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

// parseSizeToBytes converts a human-readable size string (e.g., "1.2M", "3G", "500K") to bytes.
func parseSizeToBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "0" {
		return 0, nil
	}

	var multiplier int64 = 1

	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	} else if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "T")
	} else if strings.HasSuffix(sizeStr, "B") { // Handle explicit 'B' for bytes
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	parsedValue, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size value '%s': %w", sizeStr, err)
	}

	return int64(parsedValue * float64(multiplier)), nil
}

// tryParseStandardLog attempts to parse a log line using fast string operations.
// Format: LEVEL  YYYY-MM-DD HH:MM:SS,mmm [shard X] component - message
func tryParseStandardLog(line string) (level, ts, shard, comp, msg string, ok bool) {
	if len(line) < 30 {
		return "", "", "", "", "", false
	}

	// 1. Level (INFO, WARN, ERROR, DEBUG, TRACE)
	idx := strings.IndexByte(line, ' ')
	if idx < 4 || idx > 5 {
		return "", "", "", "", "", false
	}
	level = line[:idx]
	switch level {
	case "INFO", "WARN", "ERROR", "DEBUG", "TRACE":
		// Valid level
	default:
		return "", "", "", "", "", false
	}

	// 2. Timestamp
	rest := line[idx:]
	// Skip spaces
	tsIdx := 0
	for tsIdx < len(rest) && rest[tsIdx] == ' ' {
		tsIdx++
	}
	rest = rest[tsIdx:]
	if len(rest) < 23 || rest[4] != '-' || rest[7] != '-' {
		return "", "", "", "", "", false
	}
	ts = rest[:23]
	rest = rest[23:]

	// 3. Shard (optional)
	rest = strings.TrimLeft(rest, " ")
	if strings.HasPrefix(rest, "[shard") {
		endIdx := strings.IndexByte(rest, ']')
		if endIdx != -1 {
			shard = strings.TrimSpace(rest[6:endIdx])
			rest = rest[endIdx+1:]
		}
	}

	// 4. Component and Message
	rest = strings.TrimLeft(rest, " ")
	hyphenIdx := strings.Index(rest, " - ")
	if hyphenIdx != -1 {
		comp = strings.TrimSpace(rest[:hyphenIdx])
		msg = strings.TrimSpace(rest[hyphenIdx+3:])
	} else {
		// Fallback to first space
		spaceIdx := strings.IndexByte(rest, ' ')
		if spaceIdx != -1 {
			comp = strings.TrimSpace(rest[:spaceIdx])
			msg = strings.TrimSpace(rest[spaceIdx+1:])
		} else {
			comp = strings.TrimSpace(rest)
		}
	}

	return level, ts, shard, comp, msg, true
}

// ParseLogs reads and parses all redpanda log files and stores them in the provided store.
func ParseLogs(bundlePath string, s store.Store, logsOnly bool, p *models.ProgressTracker) error {
	var allLogFiles []string

	if logsOnly {
		// In Logs Only mode, assume any file in the root is potentially a log file
		entries, err := os.ReadDir(bundlePath)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					// Skip likely non-log system files if obvious, but user requested "any file"
					// We'll trust the user but maybe skip dotfiles like .DS_Store
					if strings.HasPrefix(entry.Name(), ".") {
						continue
					}
					allLogFiles = append(allLogFiles, filepath.Join(bundlePath, entry.Name()))
				}
			}
		}
	} else {
		// In Full Bundle mode, be strict to avoid parsing config/status text files as logs
		mainLogFile := filepath.Join(bundlePath, "redpanda.log")
		if _, err := os.Stat(mainLogFile); err == nil {
			allLogFiles = append(allLogFiles, mainLogFile)
		}
	}

	// 2. Recursively find *.log and *.txt files in logs/
	dirsToWalk := []string{
		filepath.Join(bundlePath, "logs"),
	}

	for _, dir := range dirsToWalk {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				slog.Error("Error walking directory", "path", path, "error", err)
				return nil // Don't stop the walk on error for one file
			}
			if d.IsDir() {
				return nil // Skip directories
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".log" || ext == ".txt" {
				allLogFiles = append(allLogFiles, path)
			}
			return nil
		})
		if err != nil {
			slog.Error("Error during filepath.WalkDir", "path", dir, "error", err)
		}
	}

	// Use a map to store unique file paths and avoid processing the same file multiple times
	uniqueFiles := make(map[string]struct{})
	var distinctFiles []string
	for _, file := range allLogFiles {
		if _, exists := uniqueFiles[file]; !exists {
			uniqueFiles[file] = struct{}{}
			distinctFiles = append(distinctFiles, file)
		}
	}

	// Channel for sending batches to the DB writer
	// Buffer size allow parsers to proceed while DB is writing
	logBatchChan := make(chan []*models.LogEntry, 20)
	
	// Error channel to collect errors from goroutines
	errChan := make(chan error, len(distinctFiles)+1)

	var parserWg sync.WaitGroup
	var writerWg sync.WaitGroup

	slog.Info("Starting log parsing", "file_count", len(distinctFiles))

	// Heartbeat for UI liveliness
	heartbeatDone := make(chan struct{})
	if p != nil {
		go func() {
			ticker := time.NewTicker(20 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-heartbeatDone:
					return
				case <-ticker.C:
					p.SetStatus("Still processing logs... (not stuck)")
				}
			}
		}()
	}

	// Start DB Writer Goroutine
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for batch := range logBatchChan {
			if err := s.BulkInsertLogs(batch); err != nil {
				slog.Error("Error bulk inserting logs", "error", err)
				// Don't block if errChan is full
				select {
				case errChan <- fmt.Errorf("db write error: %w", err):
				default:
					slog.Warn("Error channel full, dropping error", "err", err)
				}
				return // Stop writing on error to avoid deadlock
			}
		}
	}()

	// Compile regex for log parsing
	logRegex := regexp.MustCompile(`([A-Z]+)\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}[.,]\d{3})\s+(?:\[shard\s+([^\]]+)\]\s+)?([^\s-]+)(?:\s+-\s+)?(.*)$`)
	k8sLogRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)\s+(.*)$`)
	levelKeyValRegex := regexp.MustCompile(`level="?([A-Za-z0-9]+)"?`)
	layout := "2006-01-02 15:04:05,000"

	const batchSize = 10000 // Increased batch size for better throughput

	// Limit concurrent file parsing to number of CPUs to avoid OOM
	sem := make(chan struct{}, runtime.NumCPU())

	for _, file := range distinctFiles {
		parserWg.Add(1)
		// Acquire semaphore
		sem <- struct{}{}
		go func(filePath string) {
			defer parserWg.Done()
			defer func() { <-sem }() // Release semaphore
			if p != nil {
				p.SetStatus(fmt.Sprintf("Parsing logs: %s", filepath.Base(filePath)))
			}
			slog.Info("Parsing file", "path", filePath)

			nodeName := ""
			baseName := filepath.Base(filePath)
			if baseName == "redpanda.log" {
				nodeName = "redpanda"
			} else {
				nodeName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
			}

			f, err := os.Open(filePath)
			if err != nil {
				slog.Error("Error opening log file", "path", filePath, "error", err)
				errChan <- fmt.Errorf("failed to open %s: %w", filePath, err)
				return
			}
			defer func() { _ = f.Close() }()

			// Pre-allocate batch slice to reduce re-allocations
			batch := make([]*models.LogEntry, 0, batchSize)
			
			scanner := bufio.NewScanner(f)
			const maxCapacity = 5 * 1024 * 1024 // 5MB
			buf := make([]byte, maxCapacity)
			scanner.Buffer(buf, maxCapacity)

			var currentEntry *models.LogEntry
			lineNumber := 0

			for scanner.Scan() {
				lineNumber++
				line := scanner.Text()

				if len(line) > 50000 {
					continue // Skip massive lines
				}

				// Try matching with regex

				// 1. FAST PATH: Try parsing standard Redpanda log format without regex
				fLevel, fTs, fShard, fComp, fMsg, ok := tryParseStandardLog(line)
				if ok {
					if currentEntry != nil {
						batch = append(batch, currentEntry)
						if len(batch) >= batchSize {
							logBatchChan <- batch
							batch = make([]*models.LogEntry, 0, batchSize)
						}
					}

					// Normalize timestamp
					fTs = strings.Replace(fTs, ".", ",", 1)
					timestamp, _ := time.Parse(layout, fTs)

					// Clean up shard info
					if idx := strings.Index(fShard, ":"); idx != -1 {
						fShard = fShard[:idx]
					}

					currentEntry = &models.LogEntry{
						Timestamp:  timestamp,
						Level:      fLevel,
						Node:       nodeName,
						Shard:      fShard,
						Component:  fComp,
						Message:    fMsg,
						Raw:        line,
						LineNumber: lineNumber,
						FilePath:   filePath,
					}
					continue
				}

				// 2. SLOW PATH: Try matching with regex (handles syslog, k8s, etc.)
				matches := logRegex.FindStringSubmatch(line)
				
				if matches != nil {
					if currentEntry != nil {
						batch = append(batch, currentEntry)
						if len(batch) >= batchSize {
							// Send copy of batch to avoid race conditions if we reused the slice backing array immediately
							// actually, since we do batch[:0], the backing array is reused.
							// The consumer (DB writer) reads the slice. If producer overwrites it concurrently, bad things happen.
							// So we must SEND A COPY or create a NEW slice.
							// Creating a new slice for the next batch is safer/easier.
							logBatchChan <- batch
							batch = make([]*models.LogEntry, 0, batchSize)
						}
					}

					level := matches[1]
					timestampStr := matches[2]
					shardInfo := strings.TrimSpace(matches[3])
					component := strings.TrimSpace(matches[4])
					message := strings.TrimSpace(matches[5])

					timestampStr = strings.Replace(timestampStr, ".", ",", 1)
					timestamp, _ := time.Parse(layout, timestampStr)

					if idx := strings.Index(shardInfo, ":"); idx != -1 {
						shardInfo = shardInfo[:idx]
					}
					shardInfo = strings.TrimSpace(shardInfo)

					currentEntry = &models.LogEntry{
						Timestamp:  timestamp,
						Level:      level,
						Node:       nodeName,
						Shard:      shardInfo,
						Component:  component,
						Message:    message,
						Raw:        line,
						LineNumber: lineNumber,
						FilePath:   filePath,
					}
				} else {
					k8sMatches := k8sLogRegex.FindStringSubmatch(line)
					if k8sMatches != nil {
						if currentEntry != nil {
							batch = append(batch, currentEntry)
							if len(batch) >= batchSize {
								logBatchChan <- batch
								batch = make([]*models.LogEntry, 0, batchSize)
							}
						}

						tsStr := k8sMatches[1]
						rest := k8sMatches[2]
						timestamp, _ := time.Parse(time.RFC3339, tsStr)

						level := "INFO"
						lvlMatch := levelKeyValRegex.FindStringSubmatch(rest)
						if lvlMatch != nil {
							level = strings.ToUpper(lvlMatch[1])
						}

						currentEntry = &models.LogEntry{
							Timestamp:  timestamp,
							Level:      level,
							Node:       nodeName,
							Component:  "k8s-sidecar",
							Message:    rest,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
						continue
					}

					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
						var jsonLog map[string]interface{}
						if json.Unmarshal([]byte(trimmed), &jsonLog) == nil {
							lvl, _ := jsonLog["level"].(string)
							if lvl == "" { lvl = "INFO" }
							lvl = strings.ToUpper(lvl)

							tsStr, _ := jsonLog["ts"].(string)
							if tsStr == "" { tsStr, _ = jsonLog["timestamp"].(string) }
							ts, err := time.Parse(time.RFC3339, tsStr)
							if err != nil {
								ts, _ = time.Parse("2006-01-02T15:04:05.999Z07:00", tsStr)
							}

							msg, _ := jsonLog["msg"].(string)
							if msg == "" { msg, _ = jsonLog["message"].(string) }
							if msg == "" { msg = trimmed }

							comp, _ := jsonLog["logger"].(string)
							if comp == "" { comp, _ = jsonLog["component"].(string) }
							if comp == "" { comp = "sidecar" }

							if currentEntry != nil {
								batch = append(batch, currentEntry)
							}
							
							currentEntry = &models.LogEntry{
								Timestamp:  ts,
								Level:      lvl,
								Node:       nodeName,
								Component:  comp,
								Message:    msg,
								Raw:        line,
								LineNumber: lineNumber,
								FilePath:   filePath,
							}
							continue
						}
					}

					if currentEntry != nil {
						currentEntry.Message += "\n" + line
						currentEntry.Raw += "\n" + line
					} else {
						currentEntry = &models.LogEntry{
							Timestamp:  time.Time{},
							Level:      "INFO",
							Node:       nodeName,
							Component:  "stdout",
							Message:    line,
							Raw:        line,
							LineNumber: lineNumber,
							FilePath:   filePath,
						}
					}
				}
			}

			if currentEntry != nil {
				batch = append(batch, currentEntry)
			}

			if len(batch) > 0 {
				logBatchChan <- batch
			}

			if err := scanner.Err(); err != nil {
				slog.Error("Error scanning log file", "path", filePath, "error", err)
				errChan <- fmt.Errorf("error reading %s: %w", filePath, err)
			}
		}(file)
	}

	// Wait for all parsers to finish
	parserWg.Wait()
	// Then close the channel to signal DB writer to stop
	close(logBatchChan)
	// Wait for DB writer to finish
	writerWg.Wait()
	
	close(errChan)
	
	// Stop heartbeat
	if p != nil {
		close(heartbeatDone)
	}

	// Return first error if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseDuOutput reads and parses the du.txt file.
func ParseDuOutput(bundlePath string, logsOnly bool) ([]models.DuEntry, error) {
	filePath := filepath.Join(bundlePath, "utils", "du.txt")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []models.DuEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "du:") { // Skip empty lines and du error messages
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			if !logsOnly {
				slog.Warn("Skipping malformed du.txt line", "line", line)
			}
			continue
		}

		size, err := parseSizeToBytes(parts[0])
		if err != nil {
			if !logsOnly {
				slog.Warn("Failed to parse size for du.txt entry", "line", line, "error", err)
			}
			continue
		}

		path := parts[1]
		entries = append(entries, models.DuEntry{Size: size, Path: path})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading du.txt: %w", err)
	}

	return entries, nil
}

// KafkaFullData encapsulates all data parsed from kafka.json.
type KafkaFullData struct {
	Metadata       models.KafkaMetadataResponse
	TopicConfigs   models.TopicConfigsResponse
	ConsumerGroups map[string]models.ConsumerGroup
}

// ParseKafkaJSON reads and parses the entire kafka.json file.
func ParseKafkaJSON(bundlePath string) (KafkaFullData, error) {
	filePath := filepath.Join(bundlePath, "kafka.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return KafkaFullData{}, err
	}

	var raw []struct {
		Name     string          `json:"Name"`
		Response json.RawMessage `json:"Response"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return KafkaFullData{}, fmt.Errorf("failed to parse kafka.json: %w", err)
	}

	fullData := KafkaFullData{
		ConsumerGroups: make(map[string]models.ConsumerGroup),
	}

	// Map to temporarily hold group commits before joining with groups
	groupCommits := make(map[string]map[string]map[int]*models.GroupPartitionOffset)
	highWatermarks := make(map[string]map[int]int64)

	for _, item := range raw {
		switch {
		case item.Name == "metadata":
			var metadata models.KafkaMetadataResponse
			if err := json.Unmarshal(item.Response, &metadata); err == nil {
				fullData.Metadata = metadata
				slog.Debug("Parsed metadata section from kafka.json")
			}
		case item.Name == "topic_configs":
			var topicConfigs models.TopicConfigsResponse
			if err := json.Unmarshal(item.Response, &topicConfigs); err == nil {
				fullData.TopicConfigs = topicConfigs
				slog.Debug("Parsed topic_configs section from kafka.json")
			}
		case item.Name == "groups":
			var groups map[string]models.ConsumerGroup
			if err := json.Unmarshal(item.Response, &groups); err == nil {
				for k, v := range groups {
					fullData.ConsumerGroups[k] = v
				}
				slog.Debug("Parsed groups section from kafka.json", "count", len(groups))
			} else {
				slog.Error("Failed to unmarshal groups section", "error", err)
			}
		case item.Name == "high_watermarks":
			var hwm map[string]map[int]models.HighWatermark
			if err := json.Unmarshal(item.Response, &hwm); err == nil {
				for topic, partitions := range hwm {
					if highWatermarks[topic] == nil {
						highWatermarks[topic] = make(map[int]int64)
					}
					for partID, hw := range partitions {
						highWatermarks[topic][partID] = hw.Offset
					}
				}
				slog.Debug("Parsed high_watermarks section from kafka.json")
			}
		case strings.HasPrefix(item.Name, "group_commits_"):
			groupID := strings.TrimPrefix(item.Name, "group_commits_")
			var commits map[string]map[int]*models.GroupPartitionOffset
			if err := json.Unmarshal(item.Response, &commits); err == nil {
				groupCommits[groupID] = commits
				slog.Debug("Parsed group_commits section from kafka.json", "groupID", groupID)
			}
		}
	}

	// Join commits with groups and calculate lag
	for groupID, commits := range groupCommits {
		group, ok := fullData.ConsumerGroups[groupID]
		if !ok {
			// Some groups might have commits but not be in the "groups" list (e.g. inactive)
			group = models.ConsumerGroup{Group: groupID, State: "Unknown"}
		}

		var totalGroupLag int64
		for topic, partitions := range commits {
			var topicLag int64
			for _, part := range partitions {
				if hwmTopic, ok := highWatermarks[topic]; ok {
					if hwmOffset, ok := hwmTopic[part.Partition]; ok {
						part.Lag = hwmOffset - part.At
						if part.Lag < 0 {
							part.Lag = 0 // Offset could be ahead of HWM in some race conditions or if HWM is stale
						}
						topicLag += part.Lag
					}
				}
			}
			group.Offsets = append(group.Offsets, models.GroupTopicOffset{
				Topic:      topic,
				Partitions: partitions,
				TopicLag:   topicLag,
			})
			totalGroupLag += topicLag
		}
		group.TotalLag = totalGroupLag
		fullData.ConsumerGroups[groupID] = group
	}

	return fullData, nil
}

// ParseKafkaMetadata reads and parses the kafka.json file.
func ParseKafkaMetadata(bundlePath string) (models.KafkaMetadataResponse, error) {
	fd, err := ParseKafkaJSON(bundlePath)
	if err != nil {
		return models.KafkaMetadataResponse{}, err
	}
	return fd.Metadata, nil
}

// ParseTopicConfigs reads and parses the topic_configs section from kafka.json.
func ParseTopicConfigs(bundlePath string) (models.TopicConfigsResponse, error) {
	fd, err := ParseKafkaJSON(bundlePath)
	if err != nil {
		return nil, err
	}
	return fd.TopicConfigs, nil
}


// ParseHealthOverview reads and parses the health_overview.json file.
func ParseHealthOverview(bundlePath string) (models.HealthOverview, error) {
	filePath := filepath.Join(bundlePath, "admin", "health_overview.json") // Assuming it's in admin folder
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.HealthOverview{}, err
	}

	var healthOverview models.HealthOverview
	err = json.Unmarshal(data, &healthOverview)
	if err != nil {
		slog.Error("Error unmarshaling health_overview.json", "error", err)
		return models.HealthOverview{}, err
	}
	return healthOverview, nil
}

// ParsePartitionLeaders reads and parses a partition_leader_table_...json file.
func ParsePartitionLeaders(bundlePath string, logsOnly bool) ([]models.PartitionLeader, error) {
	pattern := filepath.Join(bundlePath, "admin", "partition_leader_table_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		if !logsOnly {
			slog.Warn("No partition_leader_table file found.")
		}
		return []models.PartitionLeader{}, nil
	}

	// Try to parse each file until one succeeds
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			if !logsOnly {
				slog.Warn("Failed to read", "file", file, "error", err)
			}
			continue
		}

		var leaders []models.PartitionLeader
		if err := json.Unmarshal(data, &leaders); err == nil && len(leaders) > 0 {
			return leaders, nil
		} else {
			if !logsOnly {
				slog.Warn("Failed to parse", "file", file, "error", err)
			}
		}
	}

	return nil, fmt.Errorf("failed to parse any partition_leader_table file")
}

// ParseClusterPartitions reads and parses the cluster_partitions.json file.
func ParseClusterPartitions(bundlePath string) ([]models.ClusterPartition, error) {
	filePath := filepath.Join(bundlePath, "admin", "cluster_partitions.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var partitions []models.ClusterPartition
	err = json.Unmarshal(data, &partitions)
	if err != nil {
		return nil, err
	}

	return partitions, nil
}

// ParsedFile holds the filename and the parsed data of a file.
type ParsedFile struct {
	FileName string
	Data     interface{}
	Error    error
}

// ParseSpecificFiles reads and parses a list of specific files.
func ParseSpecificFiles(baseDir string, filenames []string) ([]ParsedFile, error) {
	var parsedFiles []ParsedFile

	for _, filename := range filenames {
		filePath := filepath.Join(baseDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			parsedFiles = append(parsedFiles, ParsedFile{
				FileName: filename,
				Error:    fmt.Errorf("file not found"),
			})
			continue
		}

		var parsedData interface{}
		if strings.HasSuffix(filename, ".json") {
			err = json.Unmarshal(data, &parsedData)
			if err != nil {
				// If JSON unmarshaling fails, treat it as plain text but log an error
				parsedFiles = append(parsedFiles, ParsedFile{
					FileName: filename,
					Data:     string(data),
					Error:    fmt.Errorf("failed to parse JSON: %w", err),
				})
				continue
			}
		} else {
			// Treat as plain text for other extensions (including .txt, .yaml, etc.)
			parsedData = string(data)
		}

		parsedFiles = append(parsedFiles, ParsedFile{
			FileName: filename,
			Data:     parsedData,
		})
	}

	return parsedFiles, nil
}

// ParseAllFiles reads and parses all files with specified extensions in a directory.
func ParseAllFiles(dir string, fileExtensions []string) ([]ParsedFile, error) {
	var parsedFiles []ParsedFile

	for _, ext := range fileExtensions {
		pattern := filepath.Join(dir, "*"+ext)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				parsedFiles = append(parsedFiles, ParsedFile{
					FileName: filepath.Base(file),
					Error:    fmt.Errorf("file not found"),
				})
				continue
			}

			var parsedData interface{}
			if strings.HasSuffix(file, ".json") {
				err = json.Unmarshal(data, &parsedData)
				if err != nil {
					// If JSON unmarshaling fails, treat it as plain text but log an error
					parsedFiles = append(parsedFiles, ParsedFile{
						FileName: filepath.Base(file),
						Data:     string(data),
						Error:    fmt.Errorf("failed to parse JSON: %w", err),
					})
					continue
				}
			} else {
				// Treat as plain text for other extensions
				parsedData = string(data)
			}

			parsedFiles = append(parsedFiles, ParsedFile{
				FileName: filepath.Base(file),
				Data:     parsedData,
			})
		}
	}

	return parsedFiles, nil
}

// ParseUnameInfo reads and parses the uname.txt file to extract system information.
func ParseUnameInfo(bundlePath string) (models.UnameInfo, error) {
	filePath := filepath.Join(bundlePath, "utils", "uname.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.UnameInfo{}, err
	}

	// uname -a output is typically:
	// Linux hostname 5.15.0-100-generic #110-Ubuntu SMP ... x86_64 ...
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		info := models.UnameInfo{
			KernelName:    fields[0],
			Hostname:      fields[1],
			KernelRelease: fields[2],
		}
		
		// Try to find architecture (usually last or second to last, e.g. x86_64)
		// And OS (GNU/Linux)
		for i, field := range fields {
			if field == "x86_64" || field == "aarch64" || field == "arm64" {
				info.Machine = field
			}
			if field == "GNU/Linux" {
				info.OperatingSystem = field
			}
			// Version often starts with # and contains build info, spreading across multiple fields
			if strings.HasPrefix(field, "#") && i > 2 {
				// Reconstruct version string roughly until we hit arch or end
				// This is heuristic as uname output varies by OS
				end := len(fields)
				for j := i; j < len(fields); j++ {
					if fields[j] == "x86_64" || fields[j] == "aarch64" || fields[j] == "arm64" {
						end = j
						break
					}
				}
				info.KernelVersion = strings.Join(fields[i:end], " ")
			}
		}
		
		return info, nil
	}

	return models.UnameInfo{}, fmt.Errorf("could not parse uname.txt: insufficient fields")
}

// ParseClusterConfig reads and parses the cluster_config.json file, generating documentation links.
func ParseClusterConfig(bundlePath string) ([]models.ClusterConfigEntry, error) {
	filePath := filepath.Join(bundlePath, "admin", "cluster_config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rawConfig map[string]interface{}
	err = json.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster_config.json: %w", err)
	}

	var configEntries []models.ClusterConfigEntry
	for key, value := range rawConfig {
		docLink := fmt.Sprintf("https://docs.redpanda.com/current/reference/properties/cluster-properties/#%s", key)
		configEntries = append(configEntries, models.ClusterConfigEntry{
			Key:     key,
			Value:   value,
			DocLink: docLink,
		})
	}

	return configEntries, nil
}

// ParseK8sResources reads and parses all supported Kubernetes JSON files in the k8s directory.
func ParseK8sResources(bundlePath string, logsOnly bool, s store.Store) (models.K8sStore, error) {
	k8sDir := filepath.Join(bundlePath, "k8s")
	k8sStore := models.K8sStore{}
	var wg sync.WaitGroup

	// Helper to parse a list file
	parseList := func(filename string, target *[]models.K8sResource) {
		defer wg.Done()
		path := filepath.Join(k8sDir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				if !logsOnly {
					slog.Warn("K8s resource file not found", "path", path)
				}
			} else {
				if !logsOnly {
					slog.Warn("Error reading K8s resource file", "path", path, "error", err)
				}
			}
			return
		}
		var list models.K8sList
		if err := json.Unmarshal(data, &list); err != nil {
			if !logsOnly {
				slog.Error("Error parsing", "filename", filename, "error", err)
			}
			return
		}
		*target = list.Items
	}

	files := []struct {
		Name   string
		Target *[]models.K8sResource
	}{
		{"pods.json", &k8sStore.Pods},
		{"services.json", &k8sStore.Services},
		{"configmaps.json", &k8sStore.ConfigMaps},
		{"endpoints.json", &k8sStore.Endpoints},
		{"events.json", &k8sStore.Events},
		{"limitranges.json", &k8sStore.LimitRanges},
		{"persistentvolumeclaims.json", &k8sStore.PersistentVolumeClaims},
		{"replicationcontrollers.json", &k8sStore.ReplicationControllers},
		{"resourcequotas.json", &k8sStore.ResourceQuotas},
		{"serviceaccounts.json", &k8sStore.ServiceAccounts},
		{"statefulsets.json", &k8sStore.StatefulSets},
		{"deployments.json", &k8sStore.Deployments},
		{"nodes.json", &k8sStore.Nodes},
	}

	wg.Add(len(files))
	for _, f := range files {
		go parseList(f.Name, f.Target)
	}

	wg.Wait()

	// Insert Events into SQLite
	if len(k8sStore.Events) > 0 {
		if err := s.BulkInsertK8sEvents(k8sStore.Events); err != nil {
			if !logsOnly {
				slog.Error("Failed to insert K8s events into store", "error", err)
			}
		}
	}

	return k8sStore, nil
}

// ParseRedpandaDataDirectory reads redpanda.yaml and extracts the data directory.
func ParseRedpandaDataDirectory(bundlePath string) (string, error) {
	filePath := filepath.Join(bundlePath, "redpanda.yaml")
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Look for data_directory: or directory:
		if strings.Contains(line, "data_directory:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				// Remove potential inline comments
				if idx := strings.Index(val, "#"); idx != -1 {
					val = strings.TrimSpace(val[:idx])
				}
				return val, nil
			}
		}
	}

	return "", fmt.Errorf("data_directory not found in redpanda.yaml")
}

// ParseRedpandaConfig reads redpanda.yaml and extracts key-value pairs.
// It performs a simple line-by-line parse to avoid external YAML dependencies.
// It handles:
// - Top level keys (redpanda:)
// - Indented keys (  data_directory: ...)
// - Inline comments
func ParseRedpandaConfig(bundlePath string) (map[string]interface{}, error) {
	filePath := filepath.Join(bundlePath, "redpanda.yaml")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	config := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	// Track context for indentation
	var parentKeys []string
	var lastIndentation int

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines or full comments
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Calculate indentation
		indentation := len(line) - len(strings.TrimLeft(line, " "))

		// Split key and value
		parts := strings.SplitN(trimmedLine, ":", 2)
		key := strings.TrimSpace(parts[0])

		if len(parts) < 2 {
			continue
		}

		value := strings.TrimSpace(parts[1])
		// Remove inline comments
		if idx := strings.Index(value, " #"); idx != -1 {
			value = strings.TrimSpace(value[:idx])
		}

		// Update context based on indentation
		if indentation == 0 {
			parentKeys = []string{}
		} else if indentation > lastIndentation {
			// Going deeper
		} else if indentation < lastIndentation {
			// Going back up - this is a rough approximation
			// A true parser would need a stack of indentations
			// For simple redpanda.yaml this often suffices to just reset if we drop back
			// But let's try to be slightly smarter: 2 spaces per level is standard
			level := indentation / 2
			if level < len(parentKeys) {
				parentKeys = parentKeys[:level]
			}
		}

		lastIndentation = indentation

		// Construct full key
		fullKey := key
		if len(parentKeys) > 0 {
			fullKey = strings.Join(parentKeys, ".") + "." + key
		}

		// If value is empty, it's likely a parent key
		if value == "" {
			parentKeys = append(parentKeys, key)
		} else {
			config[fullKey] = value
		}
	}

	return config, nil
}

// ParseMetricDefinitions extracts metric names and help text from a Prometheus format file.
func ParseMetricDefinitions(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	definitions := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// Regex for HELP line: # HELP metric_name Description text
	re := regexp.MustCompile(`^# HELP ([a-zA-Z0-9_:]+) (.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# HELP") {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				definitions[matches[1]] = matches[2]
			}
		}
	}

	return definitions, scanner.Err()
}

// ParsePrometheusMetrics parses a file containing Prometheus text-format metrics
// and stores them in the provided store.
func ParsePrometheusMetrics(filePath string, allowedMetrics map[string]struct{}, s store.Store, metricTimestamp time.Time) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var batch []*models.PrometheusMetric
	const batchSize = 1000 // Batch size for bulk inserts
	scanner := bufio.NewScanner(file)

	// regex for parsing: name{label="value",...} value
	// This is a simplified regex and might not cover all edge cases of Prom format (e.g. escaping), but sufficient for Redpanda metrics.
	// Groups: 1=Name, 2=Labels content (optional), 3=Value
	re := regexp.MustCompile(`^([a-zA-Z0-9_:]+)(?:\{([^}]+)\})?\s+([0-9eE.+\-]+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		name := matches[1]
		if allowedMetrics != nil {
			if _, ok := allowedMetrics[name]; !ok {
				continue
			}
		}

		labelsStr := matches[2]
		valueStr := matches[3]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		labels := make(map[string]string)
		if labelsStr != "" {
			// naive split by comma - assumes no commas in values.
			// Redpanda labels usually don't have commas.
			pairs := strings.Split(labelsStr, ",")
			for _, p := range pairs {
				kv := strings.SplitN(p, "=", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
					labels[key] = val
				}
			}
		}

		batch = append(batch, &models.PrometheusMetric{
			Name:   name,
			Labels: labels,
			Value:  value,
		})

		if len(batch) >= batchSize {
			if err := s.BulkInsertMetrics(batch, metricTimestamp); err != nil {
				return fmt.Errorf("failed to bulk insert metrics: %w", err)
			}
			batch = batch[:0] // Clear batch
		}
	}

	if len(batch) > 0 { // Insert any remaining metrics
		if err := s.BulkInsertMetrics(batch, metricTimestamp); err != nil {
			return fmt.Errorf("failed to bulk insert remaining metrics: %w", err)
		}
	}

	return scanner.Err()
}



================================================
FILE: internal/parser/sar.go
================================================
package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// ParseSar parses a sar text file (often in utils/sar or just sar output files)
func ParseSar(bundlePath string, logsOnly bool) (models.SarData, error) {
	// Try to find a file that looks like sar output
	// Often named 'sar', 'sar.txt', or inside 'sysstat' dir
	candidates := []string{
		filepath.Join(bundlePath, "utils", "sar"),
		filepath.Join(bundlePath, "utils", "sar.txt"),
		filepath.Join(bundlePath, "sar.txt"),
	}

	var filePath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			filePath = c
			break
		}
	}

	if filePath == "" {
		return models.SarData{}, nil // Not found
	}

	file, err := os.Open(filePath)
	if err != nil {
		return models.SarData{}, err
	}
	defer func() { _ = file.Close() }()

	var data models.SarData
	scanner := bufio.NewScanner(file)
	
	// Very basic parser for "sar -u" output
	// Format: HH:MM:SS     CPU     %user     %nice   %system   %iowait    %steal     %idle
	// 12:00:01     all      0.10      0.00      0.20      0.00      0.00     99.70

	// We need to know the date to construct full timestamp, but sar output usually only has time.
	// We'll infer date from bundle timestamp or just use a dummy date for graphing (as time is relative).
	// Let's assume today's date or 0000-01-01.

	baseDate := time.Now().Truncate(24 * time.Hour) // Use current date for relative display

	headerFound := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		if !headerFound {
			if fields[1] == "CPU" && fields[2] == "%user" {
				headerFound = true
			}
			continue
		}

		// Parse Data Row
		// Time CPU %user %nice %system %iowait %steal %idle
		// 0    1   2     3     4       5       6      7
		if len(fields) >= 8 && (fields[1] == "all" || fields[1] == "CPU") { // "all" cores or header repetition
			// Parse Time
			t, err := time.Parse("15:04:05", fields[0])
			if err != nil {
				// Try AM/PM format
				t, err = time.Parse("03:04:05 PM", fields[0]+" "+fields[1]) // Logic gets complex with AM/PM fields splitting
				if err != nil {
					// Simplified: Just skip if fail
					continue
				}
			}
			
			// Combine with base date
			fullTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), t.Hour(), t.Minute(), t.Second(), 0, baseDate.Location())

			user, _ := strconv.ParseFloat(fields[2], 64)
			system, _ := strconv.ParseFloat(fields[4], 64)
			idle, _ := strconv.ParseFloat(fields[7], 64)

			data.Entries = append(data.Entries, models.SarEntry{
				Timestamp: fullTime,
				CPUUser:   user,
				CPUSystem: system,
				CPUIdle:   idle,
			})
		}
	}

	return data, nil
}



================================================
FILE: internal/parser/segment.go
================================================
package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// ParseLogSegment reads a Redpanda/Kafka .log file and extracts batch headers
func ParseLogSegment(filePath string) (models.SegmentInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return models.SegmentInfo{}, err
	}
	defer func() { _ = file.Close() }()

	stat, _ := file.Stat()
	info := models.SegmentInfo{
		FilePath:  filePath,
		SizeBytes: stat.Size(),
	}

	// 61 bytes header
	headerBuf := make([]byte, 61)

	for {
		// Read header
		_, err := io.ReadFull(file, headerBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Unexpected EOF or error
			info.Error = fmt.Errorf("read error at offset %d: %w", 0, err) 
			break
		}

		batch := models.RecordBatchHeader{}
		
		// Redpanda format is Little Endian
		// <IiqbIhiqqqhii
		
		batch.HeaderCRC = binary.LittleEndian.Uint32(headerBuf[0:4])
		batch.BatchLength = int32(binary.LittleEndian.Uint32(headerBuf[4:8]))
		batch.BaseOffset = int64(binary.LittleEndian.Uint64(headerBuf[8:16]))
		batch.Type = int8(headerBuf[16])
		batch.CRC = int32(binary.LittleEndian.Uint32(headerBuf[17:21]))
		batch.Attributes = int16(binary.LittleEndian.Uint16(headerBuf[21:23]))
		batch.LastOffsetDelta = int32(binary.LittleEndian.Uint32(headerBuf[23:27]))
		batch.FirstTimestamp = int64(binary.LittleEndian.Uint64(headerBuf[27:35]))
		batch.MaxTimestamp = int64(binary.LittleEndian.Uint64(headerBuf[35:43]))
		batch.ProducerID = int64(binary.LittleEndian.Uint64(headerBuf[43:51]))
		batch.ProducerEpoch = int16(binary.LittleEndian.Uint16(headerBuf[51:53]))
		batch.BaseSequence = int32(binary.LittleEndian.Uint32(headerBuf[53:57]))
		batch.RecordCount = int32(binary.LittleEndian.Uint32(headerBuf[57:61]))

		// Sanity Checks
		if batch.BatchLength < 0 || batch.BatchLength > 100*1024*1024 { // Cap at 100MB
			// Check for zero-padding (truncation point)
			isZero := true
			// We check the first few bytes we read (HeaderCRC + BatchLength)
			// HeaderCRC is at headerBuf[0:4], BatchLength at headerBuf[4:8]
			for i := 0; i < 8; i++ {
				if headerBuf[i] != 0 {
					isZero = false
					break
				}
			}
			if isZero {
				break // Stop cleanly on zero padding
			}
			info.Error = fmt.Errorf("invalid batch length: %d", batch.BatchLength)
			break
		}
		
		// BatchLength includes header? Yes.
		if batch.BatchLength < 61 {
             // If length is 0, we already caught it above? 
             // batch.BatchLength is int32. If it's 0, it's < 61.
             // But we need to distinguish 0 (padding) from 1-60 (corruption).
             if batch.BatchLength == 0 {
                 break // Treat 0 as clean EOF (padding)
             }
             info.Error = fmt.Errorf("invalid batch length (smaller than header): %d", batch.BatchLength)
             break
        }
		
		// Derived
		batch.MaxOffset = batch.BaseOffset + int64(batch.RecordCount) - 1 // Redpanda logic: base + count - 1
		if batch.MaxTimestamp > 0 {
			batch.Timestamp = time.UnixMilli(batch.MaxTimestamp)
		}

		// Compression (lower 3 bits of attributes)
		compressionCode := batch.Attributes & 0x07
		switch compressionCode {
		case 0:
			batch.CompressionType = "None"
		case 1:
			batch.CompressionType = "Gzip"
		case 2:
			batch.CompressionType = "Snappy"
		case 3:
			batch.CompressionType = "LZ4"
		case 4:
			batch.CompressionType = "Zstd"
		default:
			batch.CompressionType = fmt.Sprintf("Unknown(%d)", compressionCode)
		}
		
		// Batch Type Mapping
		switch batch.Type {
		case 1: batch.TypeString = "Raft Data"
		case 2: batch.TypeString = "Raft Config"
		case 3: batch.TypeString = "Controller"
		case 4: batch.TypeString = "KVStore"
		case 5: batch.TypeString = "Checkpoint"
		case 6: batch.TypeString = "Topic Mgmt"
		default: batch.TypeString = fmt.Sprintf("%d", batch.Type)
		}

		info.Batches = append(info.Batches, batch)

		// Skip payload
		// BatchLength includes header size?
		// In Redpanda storage.py: records_size = header.batch_size - HEADER_SIZE
		// So BatchLength includes the 61 bytes of header.
		
		payloadSize := int64(batch.BatchLength) - 61
		if payloadSize < 0 {
			info.Error = fmt.Errorf("invalid batch length (smaller than header): %d", batch.BatchLength)
			break
		}

		if payloadSize > 0 {
			// If it's a Raft Config batch and not compressed, try to parse details
			if batch.Type == 2 && batch.CompressionType == "None" {
				payload := make([]byte, payloadSize)
				if _, err := io.ReadFull(file, payload); err != nil {
					info.Error = fmt.Errorf("payload read error: %w", err)
					break
				}
				
				// Attempt to parse records
				r := bytes.NewReader(payload)
				// Records are just concatenated? No, they are usually in a sequence?
				// Redpanda/Kafka RecordBatch payload IS the sequence of Records.
				
				// Read Record Count times
				for i := 0; i < int(batch.RecordCount); i++ {
					// Read Length (Varint)
					length, err := readVarint(r)
					if err != nil { break }
					
					// We should really consume 'length' bytes here to stay in sync
					// But we are parsing inside... let's just track start/end if possible
					// For now, assume our heuristic parsing of fields matches the length
					
					// Read Attributes (int8)
					_, err = r.ReadByte()
					if err != nil { break }
					
					// TimestampDelta (Varint)
					_, err = readVarint(r)
					if err != nil { break }
					
					// OffsetDelta (Varint)
					_, err = readVarint(r)
					if err != nil { break }
					
					// KeyLength (Varint)
					keyLen, err := readVarint(r)
					if err != nil { break }
					if keyLen > 0 {
						// Skip key
						r.Seek(keyLen, io.SeekCurrent)
					}
					
					// ValueLength (Varint)
					valLen, err := readVarint(r)
					if err != nil { break }
					
					if valLen > 0 {
						valBytes := make([]byte, valLen)
						if _, err := io.ReadFull(r, valBytes); err == nil {
							// Parse Raft Config from Value
							// Simple heuristic: Search for "voters" pattern or just 4-byte integers that look like node IDs (0, 1, 2...)
							// This is brittle but better than nothing.
							// Assuming version < 5 or simple vector structure:
							// [VectorSize(int32)][NodeID(int32)][Revision(int64)]...
							
							// Let's create a minimal config object
							cfg := &models.RaftConfig{}
							
							// Scan the bytes for sequences that look like node IDs
							// Node IDs are usually small (0-100). Revisions are huge.
							// If we find a small int followed by a large int (revision), it's likely a vnode.
							
							if len(valBytes) > 4 {
								// buf := bytes.NewReader(valBytes)
								// Try to read version byte
								// ver, _ := buf.ReadByte()
								
								// Since format is complex, let's just use a very loose heuristic
								// Iterate 4-byte chunks
								for j := 0; j < len(valBytes)-12; j++ {
									// Read int32
									id := int32(binary.LittleEndian.Uint32(valBytes[j:j+4]))
									// Read int64 (next 8 bytes)
									// rev := int64(binary.LittleEndian.Uint64(valBytes[j+4:j+12]))
									
									if id >= 0 && id < 1000 { // Reasonable Node ID range
										// Heuristic: Check if already added to avoid dupes/noise
										found := false
										for _, existing := range cfg.Voters {
											if existing == id { found = true; break }
										}
										if !found {
											cfg.Voters = append(cfg.Voters, id)
										}
									}
								}
							}
							if len(cfg.Voters) > 0 {
								// Assign to the LAST batch in list (which is the current one)
								info.Batches[len(info.Batches)-1].RaftConfig = cfg
							}
						}
					}
					
					// Important: Ensure we skipped exactly 'length' bytes for this record?
					// Ideally yes, but our manual field reading + skips *should* sum to length.
					// If not, we drift.
					// For robust parsing, we should have read 'length' bytes into a buffer first.
					_ = length
				}
				
			} else {
				if _, err := file.Seek(payloadSize, io.SeekCurrent); err != nil {
					info.Error = fmt.Errorf("seek error: %w", err)
					break
				}
			}
		}
	}

	return info, nil
}

// readVarint reads a signed variable-length integer (ZigZag encoded)
// But Kafka/Redpanda often use unsigned varints for lengths? 
// Actually Kafka "Varint" is signed ZigZag. "UVarint" is unsigned.
// The Record format uses Varints for lengths.
func readVarint(r io.ByteReader) (int64, error) {
	var x uint64
	var s uint
	for i := 0; i < 9; i++ { // Max 9 bytes for 64-bit
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if b < 0x80 {
			x |= uint64(b) << s
			// ZigZag decode
			return int64((x >> 1) ^ -(x & 1)), nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, fmt.Errorf("varint overflow")
}

// parseRaftConfig parses the Raft Configuration from a Record value
func parseRaftConfig(value []byte) (*models.RaftConfig, error) {
	// Simplified parser based on Redpanda model.py
	// This is a heuristic attempt.
	if len(value) < 1 {
		return nil, fmt.Errorf("empty value")
	}
	
	// Value format: 
	// version (int8)
	// if version < 5: vector of brokers...
	// if version >= 6: envelope...
	
	// Let's assume simplest case or try to find pattern.
	// Vector of VNodes: [Size(int32), VNode1, VNode2...]
	// VNode: ID(int32), Revision(int64)
	
	// Given the complexity, let's just look for a sequence of (NodeID, Revision) which look like small integers + large integers.
	// Or, if version < 5:
	// current_config: vector<vnode>
	// prev_config: vector<vnode>
	
	// Let's try to just dump the Node IDs found.
	config := &models.RaftConfig{}
	
	// Brute force scan for Node IDs? No, that's unreliable.
	// We really need a binary reader.
	
	// Since I can't easily implement the full parser in one go without 'bytes.Reader',
	// I will just return nil for now to setup the plumbing, 
	// but I really need to implement 'parseRecords' first.
	return config, nil
}

// ParseSnapshot reads a Redpanda snapshot file and extracts the header
func ParseSnapshot(filePath string) (models.SegmentInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return models.SegmentInfo{}, err
	}
	defer func() { _ = file.Close() }()

	stat, _ := file.Stat()
	info := models.SegmentInfo{
		FilePath:  filePath,
		SizeBytes: stat.Size(),
	}

	// Read fixed header fields
	// HeaderCRC(4) + DataCRC(4) + Version(1) + Size(4) + Index(8) + Term(8) = 29 bytes
	headerBuf := make([]byte, 29)
	if _, err := io.ReadFull(file, headerBuf); err != nil {
		info.Error = fmt.Errorf("failed to read snapshot header: %w", err)
		return info, nil
	}

	snap := &models.SnapshotHeader{
		HeaderCRC:         binary.LittleEndian.Uint32(headerBuf[0:4]),
		HeaderDataCRC:     binary.LittleEndian.Uint32(headerBuf[4:8]),
		HeaderVersion:     int8(headerBuf[8]),
		MetadataSize:      binary.LittleEndian.Uint32(headerBuf[9:13]),
		LastIncludedIndex: int64(binary.LittleEndian.Uint64(headerBuf[13:21])),
		LastIncludedTerm:  int64(binary.LittleEndian.Uint64(headerBuf[21:29])),
	}

	info.Snapshot = snap
	return info, nil
}



================================================
FILE: internal/parser/system.go
================================================
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

		// Example: /dev/mapper/vgbi--redpanda05-root ext4 380288728 8436056 352461800 3% /

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



================================================
FILE: internal/server/handlers_api.go
================================================
package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
)

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	var path string
	var err error

	switch runtime.GOOS {
	case "darwin":
		// macOS: Use AppleScript to open folder picker and return POSIX path
		script := `POSIX path of (choose folder with prompt "Select Redpanda Bundle Directory:")`
		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err == nil {
			path = strings.TrimSpace(string(output))
		}
	case "windows":
		// Windows: Use PowerShell to open folder picker
		script := `
		Add-Type -AssemblyName System.Windows.Forms
		$FolderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog
		$FolderBrowser.Description = "Select Redpanda Bundle Directory"
		$result = $FolderBrowser.ShowDialog()
		if ($result -eq "OK") {
			Write-Host $FolderBrowser.SelectedPath
		}
		`
		cmd := exec.Command("powershell", "-Command", script)
		output, err := cmd.Output()
		if err == nil {
			path = strings.TrimSpace(string(output))
		}
	default:
		http.Error(w, "Folder picker not supported on this platform. Please enter path manually.", http.StatusNotImplemented)
		return
	}

	if err != nil {
		// User likely cancelled or something went wrong
		s.logger.Debug("Folder picker closed or failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"path": ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": path})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	bundlePath := r.FormValue("bundlePath")
	logsOnly := r.FormValue("logsOnly") == "on"

	if bundlePath == "" {
		http.Error(w, "Bundle path is required", http.StatusBadRequest)
		return
	}

	// 1. Initialize Progress
	s.progress = models.NewProgressTracker(21)

	// 1. Generate unique DB name for this bundle
	bundleName := filepath.Base(bundlePath)
	sanitizedName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, bundleName)
	dbName := fmt.Sprintf("bundle_%s.db", sanitizedName)
	dbPath := filepath.Join(s.dataDir, dbName)

	cleanDB := !s.persist
	if cleanDB {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("Failed to remove old database", "db", dbPath, "error", err)
		}
	}

	sqliteStore, err := store.NewSQLiteStore(dbPath, bundlePath, cleanDB)
	if err != nil {
		s.logger.Error("Failed to initialize SQLite store", "error", err)
		http.Error(w, "Failed to initialize database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Initialize Cache with progress tracker
	cachedData, err := cache.New(bundlePath, sqliteStore, logsOnly, s.progress)
	if err != nil {
		s.logger.Error("Failed to create cache", "error", err)
		http.Error(w, "Failed to analyze bundle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Create and Store Session
	session := &BundleSession{
		Path:         bundlePath,
		Name:         bundleName,
		CachedData:   cachedData,
		Store:        sqliteStore,
		NodeHostname: getNodeHostname(bundlePath, cachedData),
		LogsOnly:     logsOnly,
	}
	s.sessions[bundlePath] = session

	// 4. Update Active Server State
	s.activePath = bundlePath
	s.bundlePath = session.Path
	s.logsOnly = session.LogsOnly
	s.store = session.Store
	s.cachedData = session.CachedData
	s.nodeHostname = session.NodeHostname

	s.progress.Finish()

	// 5. Signal success and redirect via HTMX
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSwitchBundle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	session, exists := s.sessions[path]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update active fields
	s.activePath = path
	s.bundlePath = session.Path
	s.logsOnly = session.LogsOnly
	s.store = session.Store
	s.cachedData = session.CachedData
	s.nodeHostname = session.NodeHostname

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handlePartitionLeadersAPI(w http.ResponseWriter, r *http.Request) {
	leaders := s.cachedData.Leaders

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(leaders); err != nil {
		s.logger.Error("Error encoding partition leaders", "error", err)
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) fullDataHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("fileName")
	jsonPath := r.URL.Query().Get("jsonPath")

	if fileName == "" || jsonPath == "" {
		http.Error(w, "fileName and jsonPath parameters required", http.StatusBadRequest)
		return
	}

	var targetData interface{}
	found := false
	for _, group := range s.cachedData.GroupedFiles {
		for _, file := range group {
			if file.FileName == fileName {
				targetData = file.Data
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		http.Error(w, "File not found in cache", http.StatusNotFound)
		return
	}

	currentData := traverseJSONPath(targetData, jsonPath)
	if currentData == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(currentData); err != nil {
		s.logger.Error("Error encoding full data", "error", err)
	}
}

func (s *Server) progressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	for {
		// Check if connection is still alive
		select {
		case <-r.Context().Done():
			return // Connection closed by client
		default:
			percentage, status, finished := s.progress.Get()
			if _, err := fmt.Fprintf(w, "data: {\"progress\": %d, \"status\": \"%s\", \"finished\": %v}\n\n", percentage, status, finished); err != nil {
				return
			}
			flusher.Flush()
			if finished {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (s *Server) topicDetailsHandler(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	if topicName == "" {
		http.Error(w, "topic parameter required", http.StatusBadRequest)
		return
	}

	partitions := s.cachedData.Partitions
	leaders := s.cachedData.Leaders
	leaderMap := make(map[string]int, len(leaders))
	for _, l := range leaders {
		key := fmt.Sprintf("%s-%d", l.Topic, l.PartitionID)
		leaderMap[key] = l.Leader
	}

	var topicPartitions []models.PartitionInfo
	for _, p := range partitions {
		if p.Topic == topicName {
			key := fmt.Sprintf("%s-%d", p.Topic, p.PartitionID)
			var replicas []int
			for _, r := range p.Replicas {
				replicas = append(replicas, r.NodeID)
			}
			topicPartitions = append(topicPartitions, models.PartitionInfo{
				ID:       p.PartitionID,
				Leader:   leaderMap[key],
				Replicas: replicas,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(topicPartitions); err != nil {
		s.logger.Error("Error encoding topic details", "error", err)
	}
}

// Lazy loading handler for large datasets
func (s *Server) lazyLoadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("fileName")
	jsonPath := r.URL.Query().Get("jsonPath")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	if fileName == "" || jsonPath == "" {
		http.Error(w, "fileName and jsonPath parameters required", http.StatusBadRequest)
		return
	}

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > maxChunkSize {
		limit = maxChunkSize
	}

	// Set timeout for slow operations
	ctx := r.Context()
	done := make(chan bool)
	var targetData interface{}
	var found bool

	go func() {
		// Find the file in cached data
		for _, group := range s.cachedData.GroupedFiles {
			for _, file := range group {
				if file.FileName == fileName {
					targetData = file.Data
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		done <- true
	}()

	select {
	case <-done:
		// Continue processing
	case <-ctx.Done():
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	}

	if !found {
		http.Error(w, "File not found in cache", http.StatusNotFound)
		return
	}

	// Traverse the JSON path
	currentData := traverseJSONPath(targetData, jsonPath)
	if currentData == nil {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	// Extract the requested slice
	var result interface{}
	var total int

	switch v := currentData.(type) {
	case []interface{}:
		total = len(v)
		end := offset + limit
		if end > total {
			end = total
		}
		if offset < total {
			result = v[offset:end]
		} else {
			result = []interface{}{}
		}
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		total = len(keys)
		end := offset + limit
		if end > total {
			end = total
		}
		if offset < total {
			resultMap := make(map[string]interface{})
			for i := offset; i < end; i++ {
				resultMap[keys[i]] = v[keys[i]]
			}
			result = resultMap
		} else {
			result = map[string]interface{}{}
		}
	default:
		result = currentData
		// total = 1 (unused)
	}

	// Render HTML with timeout protection
	htmlChan := make(chan template.HTML, 1)
	go func() {
		htmlChan <- renderDataLazy(fileName, jsonPath, result, 0, offset)
	}()

	var html template.HTML
	select {
	case html = <-htmlChan:
		// Got the HTML
	case <-time.After(10 * time.Second):
		html = template.HTML("<tr><td colspan='2'>Rendering timeout - data too complex</td></tr>")
		s.logger.Warn("Rendering timeout", "fileName", fileName, "jsonPath", jsonPath)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, string(html)); err != nil {
		s.logger.Error("Failed to write lazy load response", "error", err)
	}
}



================================================
FILE: internal/server/handlers_diagnostics.go
================================================
package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/diagnostics"
	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) diagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	report := diagnostics.Audit(s.cachedData, s.store)

	type DiagnosticsPageData struct {
		NodeHostname string
		Report       diagnostics.DiagnosticsReport
		Syslog       models.SyslogAnalysis
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := DiagnosticsPageData{
		NodeHostname: s.nodeHostname,
		Report:       report,
		Syslog:       s.cachedData.System.Syslog,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.diagnosticsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing diagnostics template", "error", err)
		http.Error(w, "Failed to execute diagnostics template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write diagnostics response", "error", err)
	}
}



================================================
FILE: internal/server/handlers_logs.go
================================================
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
)

func (s *Server) logsHandler(w http.ResponseWriter, r *http.Request) {
	// Preserving query params for the HTMX request
	query := r.URL.Query().Encode()
	targetURL := "/api/full-logs-page"
	if query != "" {
		targetURL += "?" + query
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Cluster Logs</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    <div id="main-content" hx-get="%s" hx-trigger="load" hx-select="body > *" hx-swap="innerHTML">
        <nav>
            <a href="/">Home</a>
            <a href="/logs" class="active">Logs</a>
            <div class="theme-switch-wrapper">
                <div class="theme-selection-container">
                    <button class="theme-toggle-btn">Themes</button>
                    <div class="theme-menu">
                        <button class="theme-option" data-theme="light">Light</button>
                        <button class="theme-option" data-theme="dark">Dark</button>
                        <button class="theme-option" data-theme="ultra-dark">Ultra Dark</button>
                        <button class="theme-option" data-theme="retro">Retro</button>
                        <button class="theme-option" data-theme="compact">Compact</button>
                        <button class="theme-option" data-theme="powershell">Powershell</button>
                    </div>
                </div>
            </div>
        </nav>
        <div class="container" style="height: 70vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%%;">
                <h2 style="margin-bottom: 1rem;">Loading Logs...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Filtering and retrieving logs from the bundle. This may take a moment for large datasets.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0%% { transform: rotate(0deg); } 100%% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
</body>
</html>`, targetURL)
}

func (s *Server) apiFullLogsPageHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query()

	search := query.Get("search")
	ignore := query.Get("ignore")

	// Merge "ignore" into "search" for the unified UI
	if ignore != "" {
		if search != "" {
			search = fmt.Sprintf("(%s) NOT (%s)", search, ignore)
		} else {
			search = fmt.Sprintf("NOT (%s)", ignore)
		}
		// Clear ignore so it isn't applied twice if the store logic were to change,
		// and so the template doesn't see it if we were to use it.
		ignore = ""
	}

	filter := models.LogFilter{
		Search:    search,
		Ignore:    ignore,         // Should be empty now
		Level:     query["level"], // Get all "level" parameters as a slice
		Node:      query["node"],
		Component: query["component"],
		StartTime: query.Get("startTime"),
		EndTime:   query.Get("endTime"),
		Sort:      query.Get("sort"),
	}

	// Calculate global min and max times for the UI date pickers from the store
	minTime, maxTime, err := s.store.GetMinMaxLogTime()
	if err != nil {
		s.logger.Error("Error getting min/max log times", "error", err)
		http.Error(w, "Failed to retrieve log time range", http.StatusInternalServerError)
		return
	}
	var minTimeStr, maxTimeStr string
	if !minTime.IsZero() && !maxTime.IsZero() {
		minTimeStr = minTime.UTC().Format("2006-01-02T15:04:05")
		maxTimeStr = maxTime.UTC().Format("2006-01-02T15:04:05")
	}

	// Collect unique nodes and components for dropdowns from the store
	nodes, err := s.store.GetDistinctLogNodes()
	if err != nil {
		s.logger.Error("Error getting distinct log nodes", "error", err)
		http.Error(w, "Failed to retrieve distinct log nodes", http.StatusInternalServerError)
		return
	}
	components, err := s.store.GetDistinctLogComponents()
	if err != nil {
		s.logger.Error("Error getting distinct log components", "error", err)
		http.Error(w, "Failed to retrieve distinct log components", http.StatusInternalServerError)
		return
	}

	type LogsPageData struct {
		Filter     models.LogFilter
		Nodes      []string
		Components []string
		MinTime    string
		MaxTime    string
		LogsOnly   bool
		Sessions   map[string]*BundleSession
		ActivePath string
	}

	data := LogsPageData{
		Filter:     filter,
		Nodes:      nodes,
		Components: components,
		MinTime:    minTimeStr,
		MaxTime:    maxTimeStr,
		LogsOnly:   s.logsOnly,
		Sessions:   s.sessions,
		ActivePath: s.activePath,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.logsTemplate.Execute(buf, data)
	if err != nil {
		s.logger.Error("Error executing logs template", "error", err)
		http.Error(w, "Failed to execute logs template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write logs response", "error", err)
	}
}

// apiLogsHandler serves filtered and paginated logs as JSON for infinite scrolling
func (s *Server) apiLogsHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	query := r.URL.Query()

	filter := models.LogFilter{
		Search:    query.Get("search"),
		Ignore:    query.Get("ignore"),
		Level:     query["level"], // Get all "level" parameters as a slice
		Node:      query["node"],
		Component: query["component"],
		StartTime: query.Get("startTime"),
		EndTime:   query.Get("endTime"),
		Sort:      query.Get("sort"),
	}

	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 1000 { // Enforce a reasonable limit
		limit = 100
	}

	filter.Limit = limit
	filter.Offset = offset

	pagedLogs, totalFilteredLogs, err := s.store.GetLogs(&filter)
	if err != nil {
		s.logger.Error("Error getting filtered logs from store", "error", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	type apiLogsResponse struct {
		Logs    []models.LogEntry `json:"logs"`
		Total   int               `json:"total"`
		Offset  int               `json:"offset"`
		Limit   int               `json:"limit"`
		Count   int               `json:"count"`
		HasMore bool              `json:"hasMore"`
	}

	// Convert []*models.LogEntry to []models.LogEntry
	logsForResponse := make([]models.LogEntry, len(pagedLogs))
	for i, logEntry := range pagedLogs {
		logsForResponse[i] = *logEntry
	}

	response := apiLogsResponse{
		Logs:    logsForResponse,
		Total:   totalFilteredLogs,
		Offset:  offset,
		Limit:   limit,
		Count:   len(logsForResponse),
		HasMore: (offset + len(logsForResponse)) < totalFilteredLogs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode logs response", "error", err)
	}
}

// logAnalysisHandler renders the initial page with a loading state
func (s *Server) logAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	// Render the full page with "Loading..." state
	err := s.logAnalysisTemplate.Execute(buf, map[string]interface{}{
		"Partial":    false,
		"LogsOnly":   s.logsOnly,
		"Sessions":   s.sessions,
		"ActivePath": s.activePath,
	})
	if err != nil {
		s.logger.Error("Error executing log analysis template", "error", err)
		http.Error(w, "Failed to execute log analysis template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write log analysis response", "error", err)
	}
}

// apiLogAnalysisDataHandler performs the heavy log analysis and returns the HTML partial
func (s *Server) apiLogAnalysisDataHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	patterns, totalLogs, err := s.store.GetLogPatterns()
	if err != nil {
		s.logger.Error("Error getting log patterns from store", "error", err)
		http.Error(w, "Failed to retrieve log patterns", http.StatusInternalServerError)
		return
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	err = s.logAnalysisTemplate.Execute(buf, map[string]interface{}{
		"Partial":       true,
		"Patterns":      patterns,
		"TotalLogs":     totalLogs,
		"TotalPatterns": len(patterns),
		"Sessions":      s.sessions,
		"ActivePath":    s.activePath,
		"LogsOnly":      s.logsOnly,
	})
	if err != nil {
		s.logger.Error("Error executing log analysis template (partial)", "error", err)
		http.Error(w, "Failed to execute log analysis template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write log analysis data response", "error", err)
	}
}
func (s *Server) apiLogContextHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	requestedPath := query.Get("filePath")
	lineNumberStr := query.Get("lineNumber")
	contextSizeStr := query.Get("contextSize")

	if requestedPath == "" || lineNumberStr == "" {
		http.Error(w, "Missing required parameters: filePath and lineNumber", http.StatusBadRequest)
		return
	}

	// Security: Validate that the file is within the bundle directory
	// Clean the requested path to resolve .. and symlinks roughly
	cleanRequested := filepath.Clean(requestedPath)
	
	// Determine the absolute path of the bundle for comparison
	absBundlePath, err := filepath.Abs(s.bundlePath)
	if err != nil {
		s.logger.Error("Failed to resolve absolute bundle path", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// The requestedPath might be absolute or relative. 
	// If it's absolute, check prefix. If relative, join with bundle path first.
	var absRequested string
	if filepath.IsAbs(cleanRequested) {
		absRequested = cleanRequested
	} else {
		absRequested = filepath.Join(absBundlePath, cleanRequested)
	}

	// Final Clean
	absRequested = filepath.Clean(absRequested)

	// Check if the resolved path starts with the bundle directory
	// We use the volume separator logic to be robust across OSs
	if !strings.HasPrefix(absRequested, absBundlePath) {
		s.logger.Warn("Security: Attempted path traversal", "requested", requestedPath, "resolved", absRequested, "base", absBundlePath)
		http.Error(w, "Forbidden: File access denied", http.StatusForbidden)
		return
	}

	lineNumber, err := strconv.Atoi(lineNumberStr)
	if err != nil {
		http.Error(w, "Invalid lineNumber", http.StatusBadRequest)
		return
	}

	contextSize := 10 // Default context size
	if contextSizeStr != "" {
		contextSize, err = strconv.Atoi(contextSizeStr)
		if err != nil {
			http.Error(w, "Invalid contextSize", http.StatusBadRequest)
			return
		}
	}

	file, err := os.Open(absRequested)
	if err != nil {
		s.logger.Error("Error opening file for context", "path", absRequested, "error", err)
		http.Error(w, "Failed to open log file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("Error scanning file for context", "error", err)
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}

	start := lineNumber - contextSize - 1
	if start < 0 {
		start = 0
	}

	end := lineNumber + contextSize
	if end > len(lines) {
		end = len(lines)
	}

	contextLines := lines[start:end]

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contextLines); err != nil {
		s.logger.Error("Failed to encode log context", "error", err)
	}
}



================================================
FILE: internal/server/handlers_metrics.go
================================================
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type ShardCPUMetric struct {
	ShardID string
	Usage   float64 // Percentage (0-100) based on average over uptime
}

type IOMetric struct {
	ShardID  string
	Class    string
	ReadOps  float64
	WriteOps float64
}

type KafkaLatencyStats struct {
	RequestType string
	AvgLatency  float64 // Seconds
	Count       float64
}

type NetworkMetrics struct {
	KafkaLatencies    []KafkaLatencyStats
	ActiveConnections float64
	RPCErrors         float64
}

type MetricsPageData struct {
	NodeHostname  string
	ResourceUsage models.ResourceUsage
	ShardCPU      []ShardCPUMetric
	IOMetrics     []IOMetric
	SarData       models.SarData
	CoreCount     int
	Network       NetworkMetrics
	MetricNames   []string // Added for interactive explorer
	Sessions      map[string]*BundleSession
	ActivePath    string
	LogsOnly      bool
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Render a lightweight loading page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Metrics Overview</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
    <script src="/static/chart.js"></script>
</head>
<body>
    <div id="main-content" hx-get="/api/full-metrics" hx-trigger="load" hx-select="body > *" hx-swap="innerHTML">
        <nav>
            <a href="/">Home</a>
            <a href="/metrics" class="active">Metrics</a>
            <div class="theme-switch-wrapper">
                <div class="theme-selection-container">
                    <button class="theme-toggle-btn">Themes</button>
                    <div class="theme-menu">
                        <button class="theme-option" data-theme="light">Light</button>
                        <button class="theme-option" data-theme="dark">Dark</button>
                        <button class="theme-option" data-theme="ultra-dark">Ultra Dark</button>
                        <button class="theme-option" data-theme="retro">Retro</button>
                        <button class="theme-option" data-theme="compact">Compact</button>
                        <button class="theme-option" data-theme="powershell">Powershell</button>
                    </div>
                </div>
            </div>
        </nav>
        <div class="container" style="height: 70vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%%;">
                <h2 style="margin-bottom: 1rem;">Loading Metrics...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Aggregating cluster-wide metrics and preparing visualizations. This may take a moment.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0%% { transform: rotate(0deg); } 100%% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
</body>
</html>`)
}

func (s *Server) apiFullMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	// 1. Get Resource Usage
	resourceUsage := s.cachedData.ResourceUsage

	// 2. Get Metric Names for Explorer
	metricNames, err := s.store.GetDistinctMetricNames()
	if err != nil {
		s.logger.Warn("Failed to get distinct metric names", "error", err)
	}

	// 3. Process Prometheus Metrics from the store (Existing logic)
	var shardCPU []ShardCPUMetric
	var ioMetrics []IOMetric
	var uptime float64

	// Determine time range for metrics
	startTime := time.Time{} // Zero time means no start filter
	endTime := time.Time{}   // Zero time means no end filter
	limit := 0               // 0 means no limit (all results)
	offset := 0

	// Check for direct utilization metric
	useDirectUtilization := false
	shardCPUMap := make(map[string]float64)

	utilizationMetrics, err := s.store.GetMetrics("vectorized_reactor_utilization", nil, startTime, endTime, limit, offset)
	if err != nil {
		s.logger.Warn("Failed to get vectorized_reactor_utilization metrics", "error", err)
	} else if len(utilizationMetrics) > 0 {
		useDirectUtilization = true
		for _, m := range utilizationMetrics {
			shardID := m.Labels["shard"]
			if shardID != "" {
				val := m.Value
				// If value is <= 1.0, treat as ratio and convert to percentage
				if val <= 1.0 {
					val *= 100
				}
				shardCPUMap[shardID] = val
			} else {
				s.logger.Debug("vectorized_reactor_utilization metric missing shard label", "labels", m.Labels)
			}
		}
	}

	// If direct utilization is not available, calculate from busy seconds
	if !useDirectUtilization {
		busyMetrics, err := s.store.GetMetrics("vectorized_reactor_busy_seconds_total", nil, startTime, endTime, 0, 0)
		if err != nil {
			s.logger.Warn("Failed to get vectorized_reactor_busy_seconds_total metrics", "error", err)
		} else {
			uptimeMetrics, err := s.store.GetMetrics("redpanda_application_uptime_seconds_total", nil, startTime, endTime, 1, 0)
			if err != nil {
				s.logger.Warn("Failed to get uptime metrics", "error", err)
			} else if len(uptimeMetrics) > 0 {
				uptime = uptimeMetrics[0].Value
				if uptime > 0 {
					for _, m := range busyMetrics {
						shardID := m.Labels["shard"]
						if shardID != "" {
							usage := (m.Value / uptime) * 100
							shardCPUMap[shardID] = usage
						}
					}
				}
			}
		}
	}

	// Convert map to slice for template
	for shardID, usage := range shardCPUMap {
		shardCPU = append(shardCPU, ShardCPUMetric{
			ShardID: shardID,
			Usage:   usage,
		})
	}

	// Sort by shard ID
	sort.Slice(shardCPU, func(i, j int) bool {
		iNum, _ := strconv.Atoi(shardCPU[i].ShardID)
		jNum, _ := strconv.Atoi(shardCPU[j].ShardID)
		return iNum < jNum
	})

	// 4. Process IO Metrics
	ioMap := make(map[string]map[string]*IOMetric) // shard -> class -> metric

	readMetrics, err := s.store.GetMetrics("vectorized_io_queue_total_read_ops", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get read ops metrics", "error", err)
	}
	writeMetrics, err := s.store.GetMetrics("vectorized_io_queue_total_write_ops", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get write ops metrics", "error", err)
	}

	for _, m := range readMetrics {
		shardID := m.Labels["shard"]
		class := m.Labels["class"]
		if shardID != "" && class != "" {
			if ioMap[shardID] == nil {
				ioMap[shardID] = make(map[string]*IOMetric)
			}
			if ioMap[shardID][class] == nil {
				ioMap[shardID][class] = &IOMetric{
					ShardID: shardID,
					Class:   class,
				}
			}
			ioMap[shardID][class].ReadOps = m.Value
		}
	}

	for _, m := range writeMetrics {
		shardID := m.Labels["shard"]
		class := m.Labels["class"]
		if shardID != "" && class != "" {
			if ioMap[shardID] == nil {
				ioMap[shardID] = make(map[string]*IOMetric)
			}
			if ioMap[shardID][class] == nil {
				ioMap[shardID][class] = &IOMetric{
					ShardID: shardID,
					Class:   class,
				}
			}
			ioMap[shardID][class].WriteOps = m.Value
		}
	}

	// Convert to slice
	for _, classes := range ioMap {
		for _, metric := range classes {
			ioMetrics = append(ioMetrics, *metric)
		}
	}

	// Sort by shard and class
	sort.Slice(ioMetrics, func(i, j int) bool {
		if ioMetrics[i].ShardID != ioMetrics[j].ShardID {
			iNum, _ := strconv.Atoi(ioMetrics[i].ShardID)
			jNum, _ := strconv.Atoi(ioMetrics[j].ShardID)
			return iNum < jNum
		}
		return ioMetrics[i].Class < ioMetrics[j].Class
	})

	// 5. Network & Kafka Metrics
	var networkMetrics NetworkMetrics

	// Kafka Latency (Sum / Count)
	kafkaSum, err := s.store.GetMetrics("redpanda_kafka_request_latency_seconds_sum", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get kafka latency sum", "error", err)
	}
	kafkaCount, err := s.store.GetMetrics("redpanda_kafka_request_latency_seconds_count", nil, startTime, endTime, 0, 0)
	if err != nil {
		s.logger.Warn("Failed to get kafka latency count", "error", err)
	}

	type latPair struct {
		sum   float64
		count float64
	}
	latMap := make(map[string]*latPair)

	for _, m := range kafkaSum {
		reqType := m.Labels["request"]
		if reqType == "" {
			continue
		}
		if _, ok := latMap[reqType]; !ok {
			latMap[reqType] = &latPair{}
		}
		latMap[reqType].sum += m.Value
	}
	for _, m := range kafkaCount {
		reqType := m.Labels["request"]
		if reqType == "" {
			continue
		}
		if _, ok := latMap[reqType]; !ok {
			latMap[reqType] = &latPair{}
		}
		latMap[reqType].count += m.Value
	}

	for reqType, pair := range latMap {
		if pair.count > 0 {
			networkMetrics.KafkaLatencies = append(networkMetrics.KafkaLatencies, KafkaLatencyStats{
				RequestType: reqType,
				AvgLatency:  pair.sum / pair.count,
				Count:       pair.count,
			})
		}
	}
	// Sort by latency desc
	sort.Slice(networkMetrics.KafkaLatencies, func(i, j int) bool {
		return networkMetrics.KafkaLatencies[i].AvgLatency > networkMetrics.KafkaLatencies[j].AvgLatency
	})

	// Active Connections (Sum across shards)
	connMetrics, err := s.store.GetMetrics("redpanda_rpc_active_connections", nil, startTime, endTime, 0, 0)
	if err == nil {
		for _, m := range connMetrics {
			networkMetrics.ActiveConnections += m.Value
		}
	}

	// RPC Errors (Sum across shards)
	rpcErrMetrics, err := s.store.GetMetrics("redpanda_rpc_request_errors_total", nil, startTime, endTime, 0, 0)
	if err == nil {
		for _, m := range rpcErrMetrics {
			networkMetrics.RPCErrors += m.Value
		}
	}

	// 5. Build page data
	pageData := MetricsPageData{
		NodeHostname:  s.nodeHostname,
		ResourceUsage: resourceUsage,
		ShardCPU:      shardCPU,
		IOMetrics:     ioMetrics,
		SarData:       s.cachedData.SarData,
		CoreCount:     s.cachedData.System.CoreCount,
		Network:       networkMetrics,
		MetricNames:   metricNames,
		Sessions:      s.sessions,
		ActivePath:    s.activePath,
		LogsOnly:      s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.metricsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing metrics template", "error", err)
		http.Error(w, "Failed to execute metrics template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write metrics response", "error", err)
	}
}

// API Handler for Metric List
func (s *Server) handleMetricsList(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	names, err := s.store.GetDistinctMetricNames()
	if err != nil {
		s.logger.Error("Failed to get metric names", "error", err)
		http.Error(w, "Failed to get metric names", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

// API Handler for Metric Data (Chart.js format)
func (s *Server) handleMetricsData(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		http.Error(w, "Missing metric name", http.StatusBadRequest)
		return
	}

	// Fetch all data for this metric (no time range for now, fetch all)
	metrics, err := s.store.GetMetrics(metricName, nil, time.Time{}, time.Time{}, 0, 0)
	if err != nil {
		s.logger.Error("Failed to fetch metric data", "metric", metricName, "error", err)
		http.Error(w, "Failed to fetch metric data", http.StatusInternalServerError)
		return
	}

	// Group by labels
	type SeriesPoint struct {
		X string  `json:"x"` // Timestamp ISO
		Y float64 `json:"y"`
	}
	type Series struct {
		Label string        `json:"label"`
		Data  []SeriesPoint `json:"data"`
	}

	seriesMap := make(map[string]*Series)

	for _, m := range metrics {
		// Construct label string (e.g. "shard=0, class=kafka")
		var keys []string
		for k := range m.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var labelParts []string
		for _, k := range keys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, m.Labels[k]))
		}
		label := strings.Join(labelParts, ", ")
		if label == "" {
			label = "value"
		}

		if _, ok := seriesMap[label]; !ok {
			seriesMap[label] = &Series{
				Label: label,
				Data:  []SeriesPoint{},
			}
		}

		seriesMap[label].Data = append(seriesMap[label].Data, SeriesPoint{
			X: m.Timestamp.Format(time.RFC3339),
			Y: m.Value,
		})
	}

	// Convert map to slice
	var result []Series
	for _, series := range seriesMap {
		// Sort data points by time
		sort.Slice(series.Data, func(i, j int) bool {
			return series.Data[i].X < series.Data[j].X
		})
		result = append(result, *series)
	}

	// Sort series by label
	sort.Slice(result, func(i, j int) bool {
		return result[i].Label < result[j].Label
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type TopicDiskInfo struct {
	Name       string
	TotalSize  int64
	Partitions map[string]int64
}

func (s *Server) diskOverviewHandler(w http.ResponseWriter, r *http.Request) {
	resourceUsage := s.cachedData.ResourceUsage
	dataDiskStats := s.cachedData.DataDiskStats
	cacheDiskStats := s.cachedData.CacheDiskStats
	duEntries := s.cachedData.DuEntries

	topicDiskUsage := make(map[string]int64)
	partitionDiskUsage := make(map[string]map[string]int64)

	kafkaDataPathPrefix := "/usrdata/redpanda/data/kafka/"
	if s.cachedData.RedpandaDataDir != "" {
		kafkaDataPathPrefix = filepath.Join(s.cachedData.RedpandaDataDir, "kafka")
		if !strings.HasSuffix(kafkaDataPathPrefix, "/") {
			kafkaDataPathPrefix += "/"
		}
	}

	// Auto-detect prefix if duEntries don't match the expected prefix
	hasMatch := false
	if len(duEntries) > 0 {
		for _, entry := range duEntries {
			if strings.HasPrefix(entry.Path, kafkaDataPathPrefix) {
				hasMatch = true
				break
			}
		}

		if !hasMatch {
			// Try to find a common pattern
			for _, entry := range duEntries {
				if idx := strings.Index(entry.Path, "/kafka/"); idx != -1 {
					possiblePrefix := entry.Path[:idx+7]
					kafkaDataPathPrefix = possiblePrefix
					break
				}
			}
		}
	}

	for _, entry := range duEntries {
		relativePath := ""
		if strings.HasPrefix(entry.Path, kafkaDataPathPrefix) {
			relativePath = strings.TrimPrefix(entry.Path, kafkaDataPathPrefix)
		} else if idx := strings.Index(entry.Path, "/data/kafka/"); idx != -1 {
			relativePath = entry.Path[idx+12:]
		}

		if relativePath != "" {
			parts := strings.Split(relativePath, "/")

			if len(parts) >= 1 {
				topicName := parts[0]
				topicDiskUsage[topicName] += entry.Size

				if len(parts) >= 2 {
					partitionID := parts[1]
					if _, ok := partitionDiskUsage[topicName]; !ok {
						partitionDiskUsage[topicName] = make(map[string]int64)
					}
					partitionDiskUsage[topicName][partitionID] += entry.Size
				}
			}
		}
	}

	var topicsDiskInfo []TopicDiskInfo
	for topicName, totalSize := range topicDiskUsage {
		topicsDiskInfo = append(topicsDiskInfo, TopicDiskInfo{
			Name:       topicName,
			TotalSize:  totalSize,
			Partitions: partitionDiskUsage[topicName],
		})
	}
	sort.Slice(topicsDiskInfo, func(i, j int) bool {
		return topicsDiskInfo[i].Name < topicsDiskInfo[j].Name
	})

	type DiskPageData struct {
		ResourceUsage  models.ResourceUsage
		DataDiskStats  models.DiskStats
		CacheDiskStats models.DiskStats
		TopicsDiskInfo []TopicDiskInfo
		NodeHostname   string
		Sessions       map[string]*BundleSession
		ActivePath     string
		LogsOnly       bool
	}

	pageData := DiskPageData{
		ResourceUsage:  resourceUsage,
		DataDiskStats:  dataDiskStats,
		CacheDiskStats: cacheDiskStats,
		TopicsDiskInfo: topicsDiskInfo,
		NodeHostname:   s.nodeHostname,
		Sessions:       s.sessions,
		ActivePath:     s.activePath,
		LogsOnly:       s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.diskTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute disk template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write disk overview response", "error", err)
	}
}



================================================
FILE: internal/server/handlers_pages.go
================================================
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

type RackData struct {
	Count   int
	NodeIDs []int
}

type HomePageData struct {
	GroupedFiles              map[string][]parser.ParsedFile
	NodeHostname              string
	TotalBrokers              int
	Version                   string
	IsHealthy                 bool
	UnderReplicatedPartitions int
	LeaderlessPartitions      int
	NodesInMaintenanceMode    int
	MaintenanceModeNodeIDs    []int
	NodesDown                 int
	RackAwarenessEnabled      bool
	RackInfo                  map[string]RackData
	StartupTime               time.Time
	Sessions                  map[string]*BundleSession
	ActivePath                string
	LogsOnly                  bool
}

func (s *Server) setupHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	data := map[string]interface{}{
		"CanCancel": s.bundlePath != "",
	}

	if err := s.setupTemplate.Execute(buf, data); err != nil {
		http.Error(w, "Failed to execute setup template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write setup response", "error", err)
	}
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		buf := builderPool.Get().(*strings.Builder)
		buf.Reset()
		defer builderPool.Put(buf)
		data := map[string]interface{}{
			"CanCancel": false,
		}
		if err := s.setupTemplate.Execute(buf, data); err != nil {
			http.Error(w, "Failed to execute setup template", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, buf.String())
		return
	}

	if s.logsOnly {
		http.Redirect(w, r, "/logs", http.StatusFound)
		return
	}

	pageData := s.buildHomePageData()
	pageData.Sessions = s.sessions
	pageData.ActivePath = s.activePath

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.homeTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write home response", "error", err)
	}
}

func (s *Server) partitionsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	pageData := s.buildPartitionsPageData()

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.partitionsTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute partitions template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write partitions response", "error", err)
	}
}

func (s *Server) kafkaHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	kafkaMetadata := s.cachedData.KafkaMetadata
	var rawJSON string
	for _, file := range s.cachedData.GroupedFiles["Root"] {
		if file.FileName == "kafka.json" {
			if data, ok := file.Data.(string); ok {
				rawJSON = data
			} else {
				jsonBytes, err := json.MarshalIndent(file.Data, "", "  ")
				if err == nil {
					rawJSON = string(jsonBytes)
				}
			}
			break
		}
	}

	topicConfigsMap := make(map[string]models.TopicConfig)
	for _, tc := range s.cachedData.TopicConfigs {
		topicConfigsMap[tc.Name] = tc
	}

	type KafkaPageData struct {
		Metadata       models.KafkaMetadataResponse
		RawJSON        string
		TopicConfigs   map[string]models.TopicConfig
		ConsumerGroups map[string]models.ConsumerGroup
		Sessions       map[string]*BundleSession
		ActivePath     string
		LogsOnly       bool
	}

	pageData := KafkaPageData{
		Metadata:       kafkaMetadata,
		RawJSON:        rawJSON,
		TopicConfigs:   topicConfigsMap,
		ConsumerGroups: s.cachedData.ConsumerGroups,
		Sessions:       s.sessions,
		ActivePath:     s.activePath,
		LogsOnly:       s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.kafkaTemplate.Execute(buf, pageData)
	if err != nil {
		http.Error(w, "Failed to execute kafka template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write kafka response", "error", err)
	}
}

func (s *Server) groupsHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type GroupsPageData struct {
		ConsumerGroups map[string]models.ConsumerGroup
		Sessions       map[string]*BundleSession
		ActivePath     string
		LogsOnly       bool
	}

	pageData := GroupsPageData{
		ConsumerGroups: s.cachedData.ConsumerGroups,
		Sessions:       s.sessions,
		ActivePath:     s.activePath,
		LogsOnly:       s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	err := s.groupsTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing groups template", "error", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write groups response", "error", err)
	}
}

func (s *Server) k8sHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type K8sPageData struct {
		NodeHostname string
		Store        models.K8sStore
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := K8sPageData{
		NodeHostname: s.nodeHostname,
		Store:        s.cachedData.K8sStore,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.k8sTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing k8s template", "error", err)
		http.Error(w, "Failed to execute k8s template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write k8s response", "error", err)
	}
}

func (s *Server) systemHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	type SystemPageData struct {
		NodeHostname string
		System       models.SystemState
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := SystemPageData{
		NodeHostname: s.nodeHostname,
		System:       s.cachedData.System,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err := s.systemTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing system template", "error", err)
		http.Error(w, "Failed to execute system template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write system response", "error", err)
	}
}

// Helpers

type TopicPartitions struct {
	Name       string
	Partitions []models.PartitionInfo
}

type PartitionsPageData struct {
	TotalPartitions   int
	PartitionsPerNode map[int]int
	LeadersPerNode    map[int]int
	Topics            map[string]*TopicPartitions
	NodeHostname      string
	Sessions          map[string]*BundleSession
	ActivePath        string
	LogsOnly          bool
}

func (s *Server) buildHomePageData() HomePageData {
	totalBrokers := len(s.cachedData.KafkaMetadata.Brokers)
	if totalBrokers == 0 {
		totalBrokers = len(s.cachedData.HealthOverview.AllNodes)
	}
	isHealthy := s.cachedData.HealthOverview.IsHealthy
	underReplicatedPartitions := s.cachedData.HealthOverview.UnderReplicatedCount
	leaderlessPartitions := s.cachedData.HealthOverview.LeaderlessCount

	versionSet := make(map[string]struct{})
	versions := []string{}
	nodesInMaintenanceMode := 0
	maintenanceModeNodeIDs := []int{}
	rackInfo := make(map[string]RackData)

	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "brokers.json" {
			if brokers, ok := file.Data.([]interface{}); ok {
				for _, brokerData := range brokers {
					if broker, ok := brokerData.(map[string]interface{}); ok {
						nodeID := -1
						if id, ok := broker["node_id"].(float64); ok {
							nodeID = int(id)
						}

						if v, ok := broker["version"].(string); ok {
							if _, exists := versionSet[v]; !exists {
								versionSet[v] = struct{}{}
								versions = append(versions, v)
							}
						}

						inMaintenance := false
						if ms, ok := broker["maintenance_status"].(map[string]interface{}); ok {
							for _, v := range ms {
								if status, ok := v.(bool); ok && status {
									inMaintenance = true
									break
								}
							}
						}
						// Fallback to draining if maintenance_status didn't trigger it
						if !inMaintenance {
							if draining, ok := broker["draining"].(bool); ok && draining {
								inMaintenance = true
							}
						}

						if inMaintenance {
							nodesInMaintenanceMode++
							if nodeID != -1 {
								maintenanceModeNodeIDs = append(maintenanceModeNodeIDs, nodeID)
							}
						}

						if rack, ok := broker["rack"].(string); ok {
							rackData := rackInfo[rack]
							rackData.Count++
							if nodeID != -1 {
								rackData.NodeIDs = append(rackData.NodeIDs, nodeID)
							}
							rackInfo[rack] = rackData
						}
					}
				}
			}
			break
		}
	}

	versionStr := "N/A"
	if len(versions) > 0 {
		versionStr = strings.Join(versions, ", ")
	}

	rackAwarenessEnabled := false
	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "cluster_config.json" {
			if configs, ok := file.Data.([]models.ClusterConfigEntry); ok {
				for _, config := range configs {
					if config.Key == "enable_rack_awareness" {
						if enabled, ok := config.Value.(bool); ok {
							rackAwarenessEnabled = enabled
						}
						break
					}
				}
			}
			break
		}
	}

	nodesDown := len(s.cachedData.HealthOverview.NodesDown)

	return HomePageData{
		GroupedFiles:              s.cachedData.GroupedFiles,
		NodeHostname:              s.nodeHostname,
		TotalBrokers:              totalBrokers,
		Version:                   versionStr,
		IsHealthy:                 isHealthy,
		UnderReplicatedPartitions: underReplicatedPartitions,
		LeaderlessPartitions:      leaderlessPartitions,
		NodesInMaintenanceMode:    nodesInMaintenanceMode,
		MaintenanceModeNodeIDs:    maintenanceModeNodeIDs,
		NodesDown:                 nodesDown,
		RackAwarenessEnabled:      rackAwarenessEnabled,
		RackInfo:                  rackInfo,
		StartupTime:               time.Now(),
		Sessions:                  s.sessions,
		ActivePath:                s.activePath,
		LogsOnly:                  s.logsOnly,
	}
}

func (s *Server) buildPartitionsPageData() PartitionsPageData {
	partitions := s.cachedData.Partitions
	leaders := s.cachedData.Leaders

	leaderMap := make(map[string]int)
	for _, l := range leaders {
		key := fmt.Sprintf("%s-%d", l.Topic, l.PartitionID)
		leaderMap[key] = l.Leader
	}

	partitionsPerNode := make(map[int]int)
	leadersPerNode := make(map[int]int)
	topics := make(map[string]*TopicPartitions, len(partitions)/10)

	for _, p := range partitions {
		if _, ok := topics[p.Topic]; !ok {
			topics[p.Topic] = &TopicPartitions{
				Name:       p.Topic,
				Partitions: []models.PartitionInfo{},
			}
		}

		replicas := []int{}
		for _, r := range p.Replicas {
			replicas = append(replicas, r.NodeID)
			partitionsPerNode[r.NodeID]++
		}

		key := fmt.Sprintf("%s-%d", p.Topic, p.PartitionID)
		leader := leaderMap[key]
		if leader != -1 {
			leadersPerNode[leader]++
		}

		topics[p.Topic].Partitions = append(topics[p.Topic].Partitions, models.PartitionInfo{
			ID:       p.PartitionID,
			Replicas: replicas,
			Leader:   leader,
		})
	}

	return PartitionsPageData{
		TotalPartitions:   len(partitions),
		PartitionsPerNode: partitionsPerNode,
		LeadersPerNode:    leadersPerNode,
		Topics:            topics,
		NodeHostname:      s.nodeHostname,
		Sessions:          s.sessions,
		ActivePath:        s.activePath,
		LogsOnly:          s.logsOnly,
	}
}


================================================
FILE: internal/server/handlers_search.go
================================================
package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type SearchResult struct {
	Category string
	Title    string
	Link     string
	Preview  template.HTML // Allow HTML for highlighting
}

type GlobalSearchPageData struct {
	Query        string
	Results      []SearchResult
	NodeHostname string
	Sessions     map[string]*BundleSession
	ActivePath   string
	LogsOnly     bool
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" || s.store == nil {
		if err := s.searchTemplate.Execute(w, GlobalSearchPageData{
			NodeHostname: s.nodeHostname,
			Sessions:     s.sessions,
			ActivePath:   s.activePath,
			LogsOnly:     s.logsOnly,
		}); err != nil {
			s.logger.Error("Failed to execute search template", "error", err)
		}
		return
	}

	var results []SearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	lowercaseQuery := strings.ToLower(query)

	// 1. Search Logs (Limit to top 50 matches)
	wg.Add(1)
	go func() {
		defer wg.Done()
		filter := models.LogFilter{
			Search: query, // Store handles FTS
			Limit:  50,
		}
		logs, _, err := s.store.GetLogs(&filter)
		if err == nil && len(logs) > 0 {
			mu.Lock()
			for _, log := range logs {
				var preview template.HTML
				if log.Snippet != "" {
					// Use DB-provided snippet with highlighting
					preview = template.HTML(log.Snippet)
				} else {
					// Fallback to manual highlighting
					preview = highlightMatch(log.Message, query)
				}

				results = append(results, SearchResult{
					Category: "Logs",
					Title:    fmt.Sprintf("%s [%s]", log.Timestamp.Format("15:04:05"), log.Level),
					Link:     fmt.Sprintf("/logs?search=%s", query), // Link to logs page with filter
					Preview:  preview,
				})
			}
			mu.Unlock()
		}
	}()

	// 2. Search In-Memory Files (Configs, Proc, Utils)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for group, files := range s.cachedData.GroupedFiles {
			for _, file := range files {
				contentStr := ""
				// Handle string or []byte data
				if str, ok := file.Data.(string); ok {
					contentStr = str
				} else if bytes, ok := file.Data.([]byte); ok {
					contentStr = string(bytes)
				} else if lines, ok := file.Data.([]string); ok {
					contentStr = strings.Join(lines, "\n")
				} else if file.Data != nil {
					// Fallback to JSON representation for structured data
				b, _ := json.Marshal(file.Data)
					contentStr = string(b)
				}

				if contentStr != "" {
					// Check for match
					if idx := strings.Index(strings.ToLower(contentStr), lowercaseQuery); idx != -1 {
						// Extract matching context
						start := idx - 50
						if start < 0 {
							start = 0
						}
						end := idx + len(query) + 50
						if end > len(contentStr) {
							end = len(contentStr)
						}
						
						snippet := contentStr[start:end]
						preview := highlightMatch(snippet, query)

						mu.Lock()
						results = append(results, SearchResult{
							Category: fmt.Sprintf("File: %s", group),
							Title:    file.FileName,
							Link:     "/", // TODO: Deep link to home page accordion?
							Preview:  preview,
						})
						mu.Unlock()
					}
				}
			}
		}
	}()

	// 3. Search K8s Resources
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Helper to search generic map/structs could be useful here, but let's do manual checks for now
		// Pods
		for _, pod := range s.cachedData.K8sStore.Pods {
			if strings.Contains(strings.ToLower(pod.Metadata.Name), lowercaseQuery) {
				mu.Lock()
				results = append(results, SearchResult{
					Category: "Kubernetes",
					Title:    fmt.Sprintf("Pod: %s", pod.Metadata.Name),
					Link:     "/k8s",
					Preview:  template.HTML(fmt.Sprintf("Found Pod named <b>%s</b>", pod.Metadata.Name)),
				})
				mu.Unlock()
			}
		}
		// Services
		for _, svc := range s.cachedData.K8sStore.Services {
			if strings.Contains(strings.ToLower(svc.Metadata.Name), lowercaseQuery) {
				mu.Lock()
				results = append(results, SearchResult{
					Category: "Kubernetes",
					Title:    fmt.Sprintf("Service: %s", svc.Metadata.Name),
					Link:     "/k8s",
					Preview:  template.HTML(fmt.Sprintf("Found Service named <b>%s</b>", svc.Metadata.Name)),
				})
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	pageData := GlobalSearchPageData{
		Query:        query,
		Results:      results,
		NodeHostname: s.nodeHostname,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	if err := s.searchTemplate.Execute(buf, pageData); err != nil {
		s.logger.Error("Failed to execute search template with results", "error", err)
		http.Error(w, "Failed to render search results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write search response", "error", err)
	}
}

func highlightMatch(text, query string) template.HTML {
	// Simple replace ignoring case is hard without regex
	// For now, simple case-insensitive replace
	// Note: Strings.Replace is case-sensitive.
	
	// Quick hack for case-insensitive highlighting
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	
	var sb strings.Builder
	lastIdx := 0
	
	for {
		idx := strings.Index(lowerText[lastIdx:], lowerQuery)
		if idx == -1 {
			sb.WriteString(template.HTMLEscapeString(text[lastIdx:]))
			break
		}
		
		idx += lastIdx
		sb.WriteString(template.HTMLEscapeString(text[lastIdx:idx]))
		sb.WriteString("<mark>")
		sb.WriteString(template.HTMLEscapeString(text[idx : idx+len(query)]))
		sb.WriteString("</mark>")
		lastIdx = idx + len(query)
	}
	
	return template.HTML(sb.String())
}



================================================
FILE: internal/server/handlers_segments.go
================================================
package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

func (s *Server) segmentsHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Find all .log files in controller-logs recursively
	var segmentFiles []string
	searchDir := filepath.Join(s.bundlePath, "controller-logs")
	
	// Check if directory exists
	if _, err := os.Stat(searchDir); err == nil {
		err := filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !d.IsDir() {
				if strings.HasSuffix(d.Name(), ".log") || d.Name() == "snapshot" {
					rel, _ := filepath.Rel(s.bundlePath, path)
					segmentFiles = append(segmentFiles, rel)
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Warn("Error walking controller-logs", "error", err)
		}
	} else {
		// Fallback: Try finding any .log file if controller-logs doesn't exist
		// This is useful for bundles that might just have logs at root
		filepath.WalkDir(s.bundlePath, func(path string, d os.DirEntry, err error) error {
			if err != nil { return nil }
			if !d.IsDir() {
				if strings.HasSuffix(d.Name(), ".log") || d.Name() == "snapshot" {
					// Avoid adding thousands of data segment logs if we are scanning root
					// Just add a few as a fallback or restrict to specific patterns
					if len(segmentFiles) < 50 {
						rel, _ := filepath.Rel(s.bundlePath, path)
						segmentFiles = append(segmentFiles, rel)
					}
				}
			}
			return nil
		})
	}

	type SegmentsPageData struct {
		NodeHostname string
		Files        []string
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	data := SegmentsPageData{
		NodeHostname: s.nodeHostname,
		Files:        segmentFiles,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	s.segmentsTemplate.Execute(buf, data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}

func (s *Server) segmentViewHandler(w http.ResponseWriter, r *http.Request) {
	if s.bundlePath == "" {
		http.Error(w, "No bundle loaded", http.StatusServiceUnavailable)
		return
	}

	relPath := r.URL.Query().Get("file")
	if relPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	// Security: Prevent breaking out of bundle
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.bundlePath, cleanPath)
	
	var info models.SegmentInfo
	var err error

	if filepath.Base(cleanPath) == "snapshot" {
		info, err = parser.ParseSnapshot(fullPath)
	} else {
		info, err = parser.ParseLogSegment(fullPath)
	}

	if err != nil && info.Batches == nil && info.Snapshot == nil { // Partial results ok
		http.Error(w, fmt.Sprintf("Failed to parse segment: %v", err), http.StatusInternalServerError)
		return
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	s.segmentViewTemplate.Execute(buf, info) // We'll need a separate template or partial for this
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, buf.String())
}



================================================
FILE: internal/server/handlers_skew.go
================================================
package server

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/analysis"
)

func (s *Server) skewHandler(w http.ResponseWriter, r *http.Request) {
	if s.cachedData == nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var brokers []interface{}
	var clusterConfig map[string]interface{}

	for _, file := range s.cachedData.GroupedFiles["Admin"] {
		if file.FileName == "brokers.json" {
			if bList, ok := file.Data.([]interface{}); ok {
				brokers = bList
			}
		}
		if file.FileName == "cluster_config.json" {
			if config, ok := file.Data.(map[string]interface{}); ok {
				clusterConfig = config
			}
		}
	}

	skewData := analysis.AnalyzePartitionSkew(s.cachedData.Partitions, brokers, clusterConfig)

	// Sort NodeSkews and balancing nodes by ID for consistent ordering
	sort.Slice(skewData.NodeSkews, func(i, j int) bool {
		return skewData.NodeSkews[i].NodeID < skewData.NodeSkews[j].NodeID
	})
	sort.Slice(skewData.Balancing.Nodes, func(i, j int) bool {
		return skewData.Balancing.Nodes[i].NodeID < skewData.Balancing.Nodes[j].NodeID
	})

	type SkewPageData struct {
		NodeHostname string
		Analysis     analysis.SkewAnalysis
		Sessions     map[string]*BundleSession
		ActivePath   string
		LogsOnly     bool
	}

	pageData := SkewPageData{
		NodeHostname: s.nodeHostname,
		Analysis:     skewData,
		Sessions:     s.sessions,
		ActivePath:   s.activePath,
		LogsOnly:     s.logsOnly,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(16384) // Increased buffer size for larger page

	err := s.skewTemplate.Execute(buf, pageData)
	if err != nil {
		s.logger.Error("Error executing skew template", "error", err)
		http.Error(w, "Failed to execute skew template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write skew response", "error", err)
	}
}



================================================
FILE: internal/server/handlers_timeline.go
================================================
package server

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/timeline"
)

func (s *Server) timelineHandler(w http.ResponseWriter, r *http.Request) {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	// Render the full page with "Loading..." state
	err := s.timelineTemplate.Execute(buf, map[string]interface{}{
		"NodeHostname": s.nodeHostname,
		"Partial":      false,
		"Sessions":     s.sessions,
		"ActivePath":   s.activePath,
		"LogsOnly":     s.logsOnly,
	})
	if err != nil {
		s.logger.Error("Error executing timeline template", "error", err)
		http.Error(w, "Failed to execute timeline template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write timeline response", "error", err)
	}
}

type AggregatedTimelinePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) apiTimelineAggregatedHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		respondJSON(w, []AggregatedTimelinePoint{})
		return
	}

	// Parse bucket size (default 2m)
	bucketStr := r.URL.Query().Get("bucket")
	bucketSize := 2 * time.Minute
	if bucketStr != "" {
		if d, err := time.ParseDuration(bucketStr); err == nil {
			bucketSize = d
		}
	}

	// 1. Get Log Counts (Errors/Warns)
	logCounts, err := s.store.GetLogCountsByTime(bucketSize)
	if err != nil {
		s.logger.Error("Error getting log counts for timeline", "error", err)
		http.Error(w, "Failed to get log counts", http.StatusInternalServerError)
		return
	}

	// 2. Get K8s Event Counts (Warnings)
	k8sCounts, err := s.store.GetK8sEventCountsByTime(bucketSize)
	if err != nil {
		s.logger.Warn("Error getting k8s event counts for timeline", "error", err)
		// Non-fatal
	}

	// Merge counts
	merged := make(map[time.Time]int)
	for t, c := range logCounts {
		merged[t] += c
	}
	for t, c := range k8sCounts {
		merged[t] += c
	}

	// Convert to sorted slice
	var result []AggregatedTimelinePoint
	for t, c := range merged {
		result = append(result, AggregatedTimelinePoint{Timestamp: t, Count: c})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	respondJSON(w, result)
}

func (s *Server) apiTimelineDataHandler(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		return
	}

	// Prepare source data
	// For now, we mainly use logs. We can expand to K8s events later if we parse them into a usable format.

	// Retrieve logs from the store (filtered for performance)
	// We only want ERROR and WARN logs for the timeline, and we limit the count to prevent OOM
	filter := models.LogFilter{
		Level: []string{"ERROR", "WARN"},
		Limit: 10000, 
	}
	allLogsPtr, _, err := s.store.GetLogs(&filter)
	if err != nil {
		s.logger.Error("Error getting all logs for timeline", "error", err)
		http.Error(w, "Failed to retrieve logs for timeline", http.StatusInternalServerError)
		return
	}

	// Use pointers directly
	sourceData := &models.TimelineSourceData{
		Logs:      allLogsPtr,
		K8sPods:   s.cachedData.K8sStore.Pods,
		K8sEvents: s.cachedData.K8sStore.Events,
	}

	events := timeline.GenerateTimeline(sourceData)

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	buf.Grow(8192)

	err = s.timelineTemplate.Execute(buf, map[string]interface{}{
		"NodeHostname": s.nodeHostname,
		"Events":       events,
		"Partial":      true,
		"Sessions":     s.sessions,
		"ActivePath":   s.activePath,
		"LogsOnly":     s.logsOnly,
	})
	if err != nil {
		s.logger.Error("Error executing timeline template (partial)", "error", err)
		http.Error(w, "Failed to execute timeline template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write timeline data response", "error", err)
	}
}



================================================
FILE: internal/server/helpers.go
================================================
package server

import (
	"encoding/json"
	"fmt"
	"html/template" // Added for logging Glob errors
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/alextreichler/bundleViewer/internal/cache" // Re-add cache import
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func getNodeHostname(bundlePath string, cachedData *cache.CachedData) string { // Changed to cachedData
	if bundlePath == "" || cachedData == nil {
		return "N/A"
	}
	unameInfo, err := parser.ParseUnameInfo(bundlePath)
	if err != nil {
		if cachedData != nil && len(cachedData.DataDiskFiles) > 0 {
			fileName := filepath.Base(cachedData.DataDiskFiles[0])
			parts := strings.Split(fileName, "_")
			if len(parts) >= 4 {
				return strings.TrimSuffix(parts[3], ".json")
			}
		}
		return "N/A"
	}
	return unameInfo.Hostname
}

func formatMilliseconds(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.2f ms", ms)
	}
	seconds := ms / 1000
	if seconds < 60 {
		return fmt.Sprintf("%.2f s", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		remainingSeconds := seconds - float64(int(minutes)*60)
		return fmt.Sprintf("%dm %.2fs", int(minutes), remainingSeconds)
	}
	hours := minutes / 60
	remainingMinutes := minutes - float64(int(hours)*60)
	return fmt.Sprintf("%dh %.2fm", int(hours), remainingMinutes)
}

func formatBytes(b float64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
		PB
		EB
	)

	switch {
	case b >= EB:
		return fmt.Sprintf("%.2f EB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2f PB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2f KB", b/KB)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// Helper function to traverse JSON path
func traverseJSONPath(data interface{}, jsonPath string) interface{} {
	if jsonPath == "" {
		return data
	}

	pathParts := strings.Split(jsonPath, ".")
	currentData := data

	for _, part := range pathParts {
		if part == "" {
			continue
		}

		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			idxStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil
			}
			if arr, ok := currentData.([]interface{}); ok && idx >= 0 && idx < len(arr) {
				currentData = arr[idx]
			} else {
				return nil
			}
		} else {
			if m, ok := currentData.(map[string]interface{}); ok {
				if val, exists := m[part]; exists {
					currentData = val
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	return currentData
}

// renderDataLazy renders data with lazy loading support for large structures
// startOffset is the index offset of the current data chunk within the original parent collection
func renderDataLazy(fileName string, jsonPath string, data interface{}, currentDepth int, startOffset int) template.HTML {
	var buf strings.Builder
	buf.Grow(1024)

	if currentDepth > maxInitialRenderDepth {
		switch data.(type) {
		case map[string]interface{}:
			return template.HTML("{...}")
		case []interface{}:
			return template.HTML("[...]")
		default:
			// Sanitize
			s := fmt.Sprintf("%v", data)
			return template.HTML(template.HTMLEscapeString(s))
		}
	}

	switch v := data.(type) {
	case []models.ClusterConfigEntry:
		buf.WriteString("<table><thead><tr><th>Config (key)</th><th>Value</th><th>Doc Link</th></tr></thead><tbody>")
		for _, entry := range v {
			buf.WriteString("<tr>")
			buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(entry.Key)))
			value := entry.Value
			if f, ok := value.(float64); ok {
				if strings.Contains(entry.Key, "bytes") || strings.Contains(entry.Key, "size") {
					buf.WriteString(fmt.Sprintf("<td>%s (%.0f B)</td>", formatBytes(f), f))
				} else if strings.Contains(entry.Key, "ms") {
					buf.WriteString(fmt.Sprintf("<td>%s (%.0f ms)</td>", formatMilliseconds(f), f))
				} else {
					s := strconv.FormatFloat(f, 'f', -1, 64)
					buf.WriteString(fmt.Sprintf("<td>%s</td>", s))
				}
			} else {
				buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(fmt.Sprintf("%v", value))))
			}
			buf.WriteString(fmt.Sprintf("<td><a href=\"%s\" target=\"_blank\">doc</a></td>", template.HTMLEscapeString(entry.DocLink)))
			buf.WriteString("</tr>")
		}
		buf.WriteString("</tbody></table>")

	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteString("<table class=\"sortable-table\"><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody>")
		count := 0
		for _, k := range keys {
			if count >= maxInitialRenderItems {
				// HTMX Lazy Load Button
				// This replaces the current row with new rows (outerHTML)
				// We target 'this' (the button), but actually we want to append rows.
				// Table structure makes this tricky. We put the button in a TR.
				// If we swap outerHTML of the TR, we replace the button-row with the new data-rows.
				// AND the new data-rows might include a NEW button-row at the end.
				// URL params: offset = startOffset + count.
				nextOffset := startOffset + count
				url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
					template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

				buf.WriteString(fmt.Sprintf(`<tr><td colspan="2"><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest tr" hx-swap="outerHTML">Load %d more items...</button></td></tr>`,
					url, len(v)-count))
				break
			}
			buf.WriteString("<tr>")
			buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(k)))
			buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, fmt.Sprintf("%s.%s", jsonPath, k), v[k], currentDepth+1, 0)))
			buf.WriteString("</tr>")
			count++
		}
		buf.WriteString("</tbody></table>")

	case []interface{}:
		if len(v) == 0 {
			return template.HTML("[]")
		}

		// Try to render as table if elements are objects
		var processedSlice []map[string]interface{}
		for _, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				processedSlice = append(processedSlice, obj)
			} else {
				jsonBytes, err := json.Marshal(item)
				if err == nil {
					var objMap map[string]interface{}
					if json.Unmarshal(jsonBytes, &objMap) == nil {
						processedSlice = append(processedSlice, objMap)
					} else {
						processedSlice = nil
						break
					}
				} else {
					processedSlice = nil
					break
				}
			}
		}

		if len(processedSlice) > 0 {
			headersMap := make(map[string]struct{})
			for _, item := range processedSlice {
				for k := range item {
					headersMap[k] = struct{}{}
				}
			}

			var headers []string
			var otherHeaders []string

			if _, ok := headersMap["Key"]; ok {
				headers = append(headers, "Key")
			}
			if _, ok := headersMap["Value"]; ok {
				headers = append(headers, "Value")
			}
			if _, ok := headersMap["DocLink"]; ok {
				headers = append(headers, "DocLink")
			}

			for h := range headersMap {
				isPredefined := false
				switch h {
				case "Key", "Value", "DocLink":
					isPredefined = true
				}
				if !isPredefined {
					otherHeaders = append(otherHeaders, h)
				}
			}
			sort.Strings(otherHeaders)
			headers = append(headers, otherHeaders...)

			buf.WriteString("<table class=\"sortable-table\"><thead><tr>")
			for _, h := range headers {
				displayHeader := template.HTMLEscapeString(h)
				switch h {
				case "Key":
					displayHeader = "Config (key)"
				case "DocLink":
					displayHeader = "Doc Link"
				}
				buf.WriteString(fmt.Sprintf("<th>%s</th>", displayHeader))
			}
			buf.WriteString("</tr></thead><tbody>")

			for i, item := range processedSlice {
				if i >= maxInitialRenderItems {
					nextOffset := startOffset + i
					url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
						template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

					buf.WriteString(fmt.Sprintf(`<tr><td colspan="%d"><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest tr" hx-swap="outerHTML">Load %d more items...</button></td></tr>`,
						len(headers), url, len(processedSlice)-i))
					break
				}

				buf.WriteString("<tr>")
				for _, h := range headers {
					val := item[h]
					// Correctly calculate the absolute index for the child path
					absoluteIndex := startOffset + i
					childPath := fmt.Sprintf("%s[%d].%s", jsonPath, absoluteIndex, h)

					if h == "DocLink" {
						if s, ok := val.(string); ok && strings.HasPrefix(s, "http") {
							buf.WriteString(fmt.Sprintf("<td><a href=\"%s\" target=\"_blank\">doc</a></td>", template.HTMLEscapeString(s)))
						} else {
							buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, childPath, val, currentDepth+1, 0)))
						}
					} else {
						buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, childPath, val, currentDepth+1, 0)))
					}
				}
				buf.WriteString("</tr>")
			}
			buf.WriteString("</tbody></table>")
		} else {
			// Simple array
			buf.WriteString("<ul>")
			count := 0
			for i, item := range v {
				if count >= maxInitialRenderItems {
					nextOffset := startOffset + count
					url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
						template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

					buf.WriteString(fmt.Sprintf(`<li><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest li" hx-swap="outerHTML">Load %d more items...</button></li>`,
						url, len(v)-count))
					break
				}
				// Correctly calculate the absolute index for the child path
				absoluteIndex := startOffset + i
				childPath := fmt.Sprintf("%s[%d]", jsonPath, absoluteIndex)

				buf.WriteString(fmt.Sprintf("<li>%s</li>", renderDataLazy(fileName, childPath, item, currentDepth+1, 0)))
				count++
			}
			buf.WriteString("</ul>")
		}

	default:
		if s, ok := data.(string); ok {
			buf.WriteString("<pre><code>")
			// Truncate very long strings
			if len(s) > 10000 {
				buf.WriteString(template.HTMLEscapeString(s[:10000]))
				buf.WriteString("\n... (truncated)")
			} else {
				buf.WriteString(template.HTMLEscapeString(s))
			}
			buf.WriteString("</code></pre>")
			return template.HTML(buf.String())
		}

		if f, ok := data.(float64); ok {
			if fileName != "" && (strings.HasSuffix(fileName, "_bytes") || strings.HasSuffix(fileName, "_size")) {
				return template.HTML(fmt.Sprintf("%s (%.0f B)", formatBytes(f), f))
			}
			s := strconv.FormatFloat(f, 'f', -1, 64)
			return template.HTML(s)
		}
		s := fmt.Sprintf("%v", data)
		return template.HTML(template.HTMLEscapeString(s))
	}

	return template.HTML(buf.String())
}



================================================
FILE: internal/server/server.go
================================================
package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/alextreichler/bundleViewer/internal/cache"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/store"
	"github.com/alextreichler/bundleViewer/ui"
)

// SecurityHeadersMiddleware adds security-related headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// XSS Protection (legacy but still useful for some browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Clickjacking protection
		w.Header().Set("X-Frame-Options", "DENY")
		// Strict Transport Security (HSTS) - 1 year
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Content Security Policy (Basic version)
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

const (
	maxInitialRenderDepth = 2   // Only render 2 levels deep initially
	maxInitialRenderItems = 25  // Reduced from 50
	maxChunkSize          = 100 // For lazy loading
)

type BundleSession struct {
	Path         string
	Name         string
	CachedData   *cache.CachedData
	Store        store.Store
	NodeHostname string
	LogsOnly     bool
}

type Server struct {
	sessions            map[string]*BundleSession
	activePath          string
	progress            *models.ProgressTracker
	handler             http.Handler // Internal mux + middleware
	dataDir             string       // Directory for SQLite databases
	bundlePath          string
	cachedData          *cache.CachedData
	store               store.Store
	logger              *slog.Logger
	nodeHostname        string
	persist             bool
	setupTemplate       *template.Template
	homeTemplate        *template.Template
	partitionsTemplate  *template.Template
	diskTemplate        *template.Template
	kafkaTemplate       *template.Template
	k8sTemplate         *template.Template
	groupsTemplate      *template.Template
	logsTemplate        *template.Template
	logAnalysisTemplate *template.Template
	metricsTemplate     *template.Template
	systemTemplate      *template.Template
	timelineTemplate    *template.Template
	skewTemplate        *template.Template
	searchTemplate      *template.Template
	diagnosticsTemplate *template.Template
	segmentsTemplate    *template.Template
	segmentViewTemplate *template.Template
	logsOnly            bool
}

func (s *Server) SetDataDir(path string) {
	s.dataDir = path
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func New(bundlePath string, cachedData *cache.CachedData, s store.Store, logger *slog.Logger, logsOnly bool, persist bool) (http.Handler, error) { // Accept both
	funcMap := template.FuncMap{
		"renderData": func(key string, data interface{}) template.HTML {
			return renderDataLazy(key, "", data, 0, 0)
		},
		"lower": func(v interface{}) string {
			return strings.ToLower(fmt.Sprintf("%v", v))
		},
		"float64": func(v int) float64 {
			return float64(v)
		},
		"formatBytes": func(b int64) template.HTML {
			return template.HTML(formatBytes(float64(b)))
		},
		"formatBytesFloat": func(b float64) template.HTML {
			return template.HTML(formatBytes(b))
		},
		"sub": func(a, b int64) int64 {
			return a - b
		},

		"mul": func(a, b float64) float64 {
			return a * b
		},
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"containsString": func(slice []string, s string) bool {
			for _, item := range slice {
				if item == s {
					return true
				}
			}
			return false
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"sumRestarts": func(statuses []models.K8sContainerStatus) int {
			sum := 0
			for _, s := range statuses {
				sum += s.RestartCount
			}
			return sum
		},
		"add": func(a, b int) int {
			return a + b
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0.0 // Avoid division by zero
			}
			return a / b
		},
		"pct": func(val, max int) int {
			if max == 0 {
				return 0
			}
			return int((float64(val) / float64(max)) * 100)
		},
		"slice": func(s string, start, end int) string {
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start > end {
				return ""
			}
			return s[start:end]
		},
		"toInt64": func(v interface{}) int64 {
			switch i := v.(type) {
			case int:
				return int64(i)
			case int64:
				return i
			case float64:
				return int64(i)
			default:
				logger.Warn("toInt64 received unsupported type, returning 0", "type", fmt.Sprintf("%T", v))
				return 0
			}
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	homeTemplate, err := template.New("home.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/home.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse home template: %w", err)
	}

	setupTemplate, err := template.New("setup.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/setup.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse setup template: %w", err)
	}

	partitionsTemplate, err := template.New("partitions.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/partitions.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse partitions template: %w", err)
	}

	diskTemplate, err := template.New("disk.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/disk.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse disk template: %w", err)
	}

	kafkaTemplate, err := template.New("kafka.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/kafka.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse kafka template: %w", err)
	}

	k8sTemplate, err := template.New("k8s.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/k8s.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse k8s template: %w", err)
	}

	groupsTemplate, err := template.New("groups.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/groups.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse groups template: %w", err)
	}

	logsTemplate, err := template.New("logs.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/logs.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse logs template: %w", err)
	}

	logAnalysisTemplate, err := template.New("log_analysis.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/log_analysis.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse log_analysis template: %w", err)
	}

	metricsTemplate, err := template.New("metrics.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/metrics.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics template: %w", err)
	}

	systemTemplate, err := template.New("system.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/system.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse system template: %w", err)
	}

	timelineTemplate, err := template.New("timeline.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/timeline.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse timeline template: %w", err)
	}

	skewTemplate, err := template.New("skew.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/skew.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse skew template: %w", err)
	}

	searchTemplate, err := template.New("search.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/search.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse search template: %w", err)
	}

	diagnosticsTemplate, err := template.New("diagnostics.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/diagnostics.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse diagnostics template: %w", err)
	}

	segmentsTemplate, err := template.New("segments.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/segments.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse segments template: %w", err)
	}

	segmentViewTemplate, err := template.New("segment_view.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/segment_view.tmpl") // No nav needed here? Actually it's an HTMX partial, usually. But if it's standalone... wait, segment_view is likely a partial.
	if err != nil {
		return nil, fmt.Errorf("failed to parse segment_view template: %w", err)
	}

	server := &Server{
		sessions:            make(map[string]*BundleSession),
		activePath:          bundlePath,
		progress:            models.NewProgressTracker(21),
		bundlePath:          bundlePath,
		cachedData:          cachedData,
		store:               s,
		logger:              logger,
		nodeHostname:        getNodeHostname(bundlePath, cachedData),
		persist:             persist,
		setupTemplate:       setupTemplate,
		homeTemplate:        homeTemplate,
		partitionsTemplate:  partitionsTemplate,
		diskTemplate:        diskTemplate,
		kafkaTemplate:       kafkaTemplate,
		k8sTemplate:         k8sTemplate,
		groupsTemplate:      groupsTemplate,
		logsTemplate:        logsTemplate,
		logAnalysisTemplate: logAnalysisTemplate,
		metricsTemplate:     metricsTemplate,
		systemTemplate:      systemTemplate,
		timelineTemplate:    timelineTemplate,
		skewTemplate:        skewTemplate,
		searchTemplate:      searchTemplate,
		diagnosticsTemplate: diagnosticsTemplate,
		segmentsTemplate:    segmentsTemplate,
		segmentViewTemplate: segmentViewTemplate,
		logsOnly:            logsOnly,
	}

	if bundlePath != "" {
		server.sessions[bundlePath] = &BundleSession{
			Path:         bundlePath,
			Name:         filepath.Base(bundlePath),
			CachedData:   cachedData,
			Store:        s,
			NodeHostname: server.nodeHostname,
			LogsOnly:     logsOnly,
		}
	}

	mux := http.NewServeMux()

	mux.Handle("/static/", http.FileServer(http.FS(ui.StaticFiles)))

	mux.HandleFunc("/", server.homeHandler)
	mux.HandleFunc("/setup", server.setupHandler)
	mux.HandleFunc("/api/setup", server.handleSetup)
	mux.HandleFunc("/api/switch-bundle", server.handleSwitchBundle)
	mux.HandleFunc("/api/browse", server.handleBrowse)
	mux.HandleFunc("/partitions", server.partitionsHandler)
	mux.HandleFunc("/disk", server.diskOverviewHandler)
	mux.HandleFunc("/kafka", server.kafkaHandler)
	mux.HandleFunc("/k8s", server.k8sHandler)
	mux.HandleFunc("/groups", server.groupsHandler)
	mux.HandleFunc("/logs", server.logsHandler)
	mux.HandleFunc("/logs/analysis", server.logAnalysisHandler)
	mux.HandleFunc("/api/topic-details", server.topicDetailsHandler)
	mux.HandleFunc("/api/load-progress", server.progressHandler)
	mux.HandleFunc("/api/full-data", server.fullDataHandler)
	mux.HandleFunc("/api/lazy-load", server.lazyLoadHandler)
	mux.HandleFunc("/api/logs", server.apiLogsHandler)
	mux.HandleFunc("/api/logs/analysis", server.apiLogAnalysisDataHandler)
	mux.HandleFunc("/api/logs/context", server.apiLogContextHandler)
	mux.HandleFunc("/api/full-logs-page", server.apiFullLogsPageHandler) // New handler for lazy loading logs page
	mux.HandleFunc("/api/metrics/list", server.handleMetricsList)
	mux.HandleFunc("/api/metrics/data", server.handleMetricsData)
	mux.HandleFunc("/api/full-metrics", server.apiFullMetricsHandler) // New handler for lazy loading metrics page
	mux.HandleFunc("/api/partition-leaders", server.handlePartitionLeadersAPI)
	mux.HandleFunc("/api/timeline/data", server.apiTimelineDataHandler)
	mux.HandleFunc("/api/timeline/aggregated", server.apiTimelineAggregatedHandler)
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/system", server.systemHandler)
	mux.HandleFunc("/timeline", server.timelineHandler)
	mux.HandleFunc("/skew", server.skewHandler)
	mux.HandleFunc("/search", server.searchHandler)
	mux.HandleFunc("/diagnostics", server.diagnosticsHandler)
	mux.HandleFunc("/segments", server.segmentsHandler)
	mux.HandleFunc("/segments/view", server.segmentViewHandler)

	// Wrap the mux with middlewares
	// Order: Security -> Mux
	var handler http.Handler = mux
	handler = SecurityHeadersMiddleware(handler)

	server.handler = handler

	return server, nil
}

func (s *Server) Shutdown() {
	s.logger.Info("Shutting down server and closing all bundle sessions...")
	for path, session := range s.sessions {
		if session.Store != nil {
			if err := session.Store.Close(); err != nil {
				s.logger.Warn("Failed to close store for bundle", "path", path, "error", err)
			}
		}
	}
}




================================================
FILE: internal/store/sqlite_schema.go
================================================
package store

const LogSchema = `
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    level TEXT NOT NULL,
    node TEXT NOT NULL,
    shard TEXT NOT NULL,
    component TEXT NOT NULL,
    message TEXT NOT NULL,
    raw TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    fingerprint TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_node ON logs(node);
CREATE INDEX IF NOT EXISTS idx_logs_component ON logs(component);
CREATE INDEX IF NOT EXISTS idx_logs_fingerprint ON logs(fingerprint);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp_level ON logs(timestamp, level);

CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts USING fts5(
    message,
    content='logs',
    content_rowid='id',
    tokenize='trigram'
);

-- Triggers to keep FTS index in sync
CREATE TRIGGER IF NOT EXISTS logs_ai AFTER INSERT ON logs BEGIN
  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
END;
CREATE TRIGGER IF NOT EXISTS logs_ad AFTER DELETE ON logs BEGIN
  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
END;
CREATE TRIGGER IF NOT EXISTS logs_au AFTER UPDATE ON logs BEGIN
  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
END;

-- This table is required by fts5 and stores configuration data
CREATE TABLE IF NOT EXISTS logs_fts_config(k PRIMARY KEY, v) WITHOUT ROWID;
`

const MetricSchema = `
CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    labels TEXT NOT NULL, -- Stored as JSON string
    value REAL NOT NULL,
    timestamp INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
`

const K8sEventSchema = `
CREATE TABLE IF NOT EXISTS k8s_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    reason TEXT NOT NULL,
    message TEXT NOT NULL,
    count INTEGER NOT NULL,
    first_timestamp INTEGER,
    last_timestamp INTEGER NOT NULL,
    involved_object_kind TEXT NOT NULL,
    involved_object_name TEXT NOT NULL,
    involved_object_namespace TEXT NOT NULL,
    source_component TEXT NOT NULL,
    source_host TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_k8s_events_last_timestamp ON k8s_events(last_timestamp);
CREATE INDEX IF NOT EXISTS idx_k8s_events_type ON k8s_events(type);
`

const SchemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);
`

const MetadataTable = `
CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY,
    value TEXT
);
`



================================================
FILE: internal/store/sqlite_store.go
================================================
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/analysis"
	"github.com/alextreichler/bundleViewer/internal/models"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewSQLiteStore(dbPath string, bundlePath string, clean bool) (*SQLiteStore, error) {
	// Check if we need to wipe the DB (if clean flag set, bundle path changed, or DB invalid)
	if clean || shouldWipeDB(dbPath, bundlePath) {
		slog.Info("Creating fresh database", "db", dbPath, "reason", getWipeReason(clean, dbPath, bundlePath))
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to remove old database: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=10000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Performance Optimizations
	pragmas := []string{
		"PRAGMA journal_mode=OFF;",      // No rollback journal (fastest, risky)
		"PRAGMA synchronous=OFF;",       // Don't wait for disk flush
		"PRAGMA temp_store=MEMORY;",     // Temp tables in RAM
		"PRAGMA cache_size=-500000;",    // ~500MB cache
		"PRAGMA mmap_size=8000000000;",  // Memory map up to 8GB
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			slog.Warn("Failed to set pragma", "pragma", p, "error", err)
		}
	}

	// FORCE single connection to avoid SQLITE_BUSY
	db.SetMaxOpenConns(1)

	slog.Info("SQLite Store Initialized", "ptr", fmt.Sprintf("%p", db))
	store := &SQLiteStore{db: db}

	if err := store.InitSchema(bundlePath); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func getWipeReason(clean bool, dbPath, bundlePath string) string {
	if clean {
		return "clean flag set"
	}
	if shouldWipeDB(dbPath, bundlePath) {
		return "bundle path changed or db invalid"
	}
	return "unknown"
}

func shouldWipeDB(dbPath, currentBundlePath string) bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false // File doesn't exist, generic create
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return true // Corrupt? Wipe.
	}
	defer func() { _ = db.Close() }()

	// Check Version
	var version int
	err = db.QueryRow("SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil || version < 5 {
		return true // Old version or no version table? Wipe.
	}

	var storedPath string
	err = db.QueryRow("SELECT value FROM metadata WHERE key = 'bundle_path'").Scan(&storedPath)
	if err != nil {
		return true // No metadata? Wipe.
	}

	return storedPath != currentBundlePath
}

func (s *SQLiteStore) DropIndexes() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("Dropping indexes for bulk ingestion...")
	statements := []string{
		"DROP TABLE IF EXISTS logs_fts",
		"DROP TRIGGER IF EXISTS logs_ai",
		"DROP TRIGGER IF EXISTS logs_ad",
		"DROP TRIGGER IF EXISTS logs_au",
		"DROP INDEX IF EXISTS idx_logs_timestamp",
		"DROP INDEX IF EXISTS idx_logs_level",
		"DROP INDEX IF EXISTS idx_logs_node",
		"DROP INDEX IF EXISTS idx_logs_component",
		"DROP INDEX IF EXISTS idx_logs_fingerprint",
		"DROP INDEX IF EXISTS idx_logs_timestamp_level",
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to drop index/trigger: %w", err)
		}
	}
	return nil
}

func (s *SQLiteStore) RestoreIndexes() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("Restoring indexes...")
	
	// 1. Re-create FTS
	createFTS := `CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts USING fts5(
		message,
		content='logs',
		content_rowid='id',
		tokenize='trigram'
	)`
	if _, err := s.db.Exec(createFTS); err != nil {
		return fmt.Errorf("failed to create logs_fts: %w", err)
	}

	// 2. Re-create Triggers
	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS logs_ai AFTER INSERT ON logs BEGIN
		  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS logs_ad AFTER DELETE ON logs BEGIN
		  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS logs_au AFTER UPDATE ON logs BEGIN
		  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
		  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
		END;`,
	}
	for _, t := range triggers {
		if _, err := s.db.Exec(t); err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}

	// 3. Rebuild FTS Index
	slog.Info("Rebuilding FTS index...")
	if _, err := s.db.Exec("INSERT INTO logs_fts(logs_fts) VALUES('rebuild')"); err != nil {
		return fmt.Errorf("failed to rebuild logs_fts: %w", err)
	}

	// 4. Re-create Standard Indexes
	slog.Info("Rebuilding standard indexes...")
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)",
		"CREATE INDEX IF NOT EXISTS idx_logs_node ON logs(node)",
		"CREATE INDEX IF NOT EXISTS idx_logs_component ON logs(component)",
		"CREATE INDEX IF NOT EXISTS idx_logs_fingerprint ON logs(fingerprint)",
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp_level ON logs(timestamp, level)",
	}
	for _, idx := range indexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStore) InitSchema(bundlePath string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for schema initialization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Create Schema Version and Metadata tables first
	if _, err := tx.Exec(SchemaVersionTable); err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}
	if _, err := tx.Exec(MetadataTable); err != nil {
		return fmt.Errorf("failed to create metadata table: %w", err)
	}

	// 2. Check/Set Bundle Path
	var storedPath string
	err = tx.QueryRow("SELECT value FROM metadata WHERE key = 'bundle_path'").Scan(&storedPath)
	if err == sql.ErrNoRows {
		if _, err := tx.Exec("INSERT INTO metadata (key, value) VALUES ('bundle_path', ?)", bundlePath); err != nil {
			return fmt.Errorf("failed to set bundle path: %w", err)
		}
	}

	// 3. Get Current Version
	var currentVersion int
	err = tx.QueryRow("SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// 4. Run Migrations
	// Version 0 -> 1: Base Schema (includes FTS table from LogSchema)
	if currentVersion < 1 {
		if _, err := tx.Exec(LogSchema); err != nil {
			return fmt.Errorf("failed to create logs table: %w", err)
		}
		if _, err := tx.Exec(MetricSchema); err != nil {
			return fmt.Errorf("failed to create metrics table: %w", err)
		}
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (1)"); err != nil {
			return fmt.Errorf("failed to update schema version to 1: %w", err)
		}
		currentVersion = 1
	}

	// Version 1 -> 2: Add Fingerprint (if not already present)
	if currentVersion < 2 {
		// Check if fingerprint column exists
		var hasFingerprint bool
		row := tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('logs') WHERE name='fingerprint'")
		if err := row.Scan(&hasFingerprint); err == nil && !hasFingerprint {
			_, err := tx.Exec("ALTER TABLE logs ADD COLUMN fingerprint TEXT DEFAULT ''")
			if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("failed to add fingerprint column: %w", err)
			}
			_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_logs_fingerprint ON logs(fingerprint)")
			if err != nil {
				return fmt.Errorf("failed to create fingerprint index: %w", err)
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (2)"); err != nil {
			return fmt.Errorf("failed to update schema version to 2: %w", err)
		}
		currentVersion = 2
	}

	// Version 2 -> 3: Add FTS Triggers and Rebuild Index
	if currentVersion < 3 {
		triggers := []string{
			`CREATE TRIGGER IF NOT EXISTS logs_ai AFTER INSERT ON logs BEGIN
  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
END;`,
			`CREATE TRIGGER IF NOT EXISTS logs_ad AFTER DELETE ON logs BEGIN
  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
END;`,
			`CREATE TRIGGER IF NOT EXISTS logs_au AFTER UPDATE ON logs BEGIN
  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
END;`,
		}

		for _, trigger := range triggers {
			if _, err := tx.Exec(trigger); err != nil {
				return fmt.Errorf("failed to create FTS trigger: %w", err)
			}
		}

		// Rebuild the FTS index to ensure existing data is indexed
		if _, err := tx.Exec("INSERT INTO logs_fts(logs_fts) VALUES('rebuild')"); err != nil {
			return fmt.Errorf("failed to rebuild logs_fts index: %w", err)
		}

		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (3)"); err != nil {
			return fmt.Errorf("failed to update schema version to 3: %w", err)
		}
		currentVersion = 3
	}

	// Version 3 -> 4: Add K8s Events
	if currentVersion < 4 {
		if _, err := tx.Exec(K8sEventSchema); err != nil {
			return fmt.Errorf("failed to create k8s_events table: %w", err)
		}
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (4)"); err != nil {
			return fmt.Errorf("failed to update schema version to 4: %w", err)
		}
		currentVersion = 4
	}

	// Version 4 -> 5: Use Integer Timestamps (Unix Microseconds)
	if currentVersion < 5 {
		// Version 5 is handled by shouldWipeDB which forces a fresh DB.
		// We just record the version here.
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (5)"); err != nil {
			return fmt.Errorf("failed to update schema version to 5: %w", err)
		}
		currentVersion = 5
	}

	// Version 5 -> 6: Optimize FTS to use External Content (Content='logs')
	if currentVersion < 6 {
		// 1. Drop old FTS table and triggers
		if _, err := tx.Exec("DROP TABLE IF EXISTS logs_fts"); err != nil {
			return fmt.Errorf("failed to drop old logs_fts: %w", err)
		}
		if _, err := tx.Exec("DROP TRIGGER IF EXISTS logs_ai"); err != nil {
			return fmt.Errorf("failed to drop logs_ai: %w", err)
		}
		if _, err := tx.Exec("DROP TRIGGER IF EXISTS logs_ad"); err != nil {
			return fmt.Errorf("failed to drop logs_ad: %w", err)
		}
		if _, err := tx.Exec("DROP TRIGGER IF EXISTS logs_au"); err != nil {
			return fmt.Errorf("failed to drop logs_au: %w", err)
		}

		// 2. Re-create FTS table as External Content
		// Note: We duplicate the definition from LogSchema here to ensure this migration is self-contained
		createFTS := `CREATE VIRTUAL TABLE logs_fts USING fts5(
			message,
			content='logs',
			content_rowid='id',
			tokenize='trigram'
		)`
		if _, err := tx.Exec(createFTS); err != nil {
			return fmt.Errorf("failed to recreate logs_fts: %w", err)
		}

		// 3. Re-create Triggers
		triggers := []string{
			`CREATE TRIGGER logs_ai AFTER INSERT ON logs BEGIN
			  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
			END;`,
			`CREATE TRIGGER logs_ad AFTER DELETE ON logs BEGIN
			  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
			END;`,
			`CREATE TRIGGER logs_au AFTER UPDATE ON logs BEGIN
			  INSERT INTO logs_fts(logs_fts, rowid, message) VALUES('delete', old.id, old.message);
			  INSERT INTO logs_fts(rowid, message) VALUES (new.id, new.message);
			END;`,
		}
		for _, trigger := range triggers {
			if _, err := tx.Exec(trigger); err != nil {
				return fmt.Errorf("failed to recreate FTS trigger: %w", err)
			}
		}

		// 4. Rebuild Index
		if _, err := tx.Exec("INSERT INTO logs_fts(logs_fts) VALUES('rebuild')"); err != nil {
			return fmt.Errorf("failed to rebuild logs_fts index: %w", err)
		}

		// 5. Update Version
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (6)"); err != nil {
			return fmt.Errorf("failed to update schema version to 6: %w", err)
		}
		currentVersion = 6
	}

	return tx.Commit()
}

func (s *SQLiteStore) HasLogs() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM logs LIMIT 1").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check for logs: %w", err)
	}
	return count > 0, nil
}

func (s *SQLiteStore) BulkInsertK8sEvents(events []models.K8sResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for k8s events: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO k8s_events (type, reason, message, count, first_timestamp, last_timestamp, involved_object_kind, involved_object_name, involved_object_namespace, source_component, source_host)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for k8s events: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, e := range events {
		var firstTS, lastTS int64
		if !e.FirstTimestamp.IsZero() {
			firstTS = e.FirstTimestamp.UnixMicro()
		}
		if !e.LastTimestamp.IsZero() {
			lastTS = e.LastTimestamp.UnixMicro()
		}

		_, err = stmt.Exec(e.Type, e.Reason, e.Message, e.Count, firstTS, lastTS, e.InvolvedObject.Kind, e.InvolvedObject.Name, e.InvolvedObject.Namespace, e.Source.Component, e.Source.Host)
		if err != nil {
			return fmt.Errorf("failed to insert k8s event: %w", err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetLogCountsByTime(bucketSize time.Duration) (map[time.Time]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketSeconds := int64(bucketSize.Seconds())
	if bucketSeconds <= 0 {
		bucketSeconds = 60 // Default 1 min
	}

	query := `
		SELECT 
			((timestamp / 1000000) / ?) * ? as bucket, 
			COUNT(*) 
		FROM logs 
		WHERE level IN ('ERROR', 'WARN') 
		GROUP BY 1
		ORDER BY 1 ASC
	`

	rows, err := s.db.Query(query, bucketSeconds, bucketSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query log counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[time.Time]int)
	for rows.Next() {
		var ts int64
		var count int
		if err := rows.Scan(&ts, &count); err != nil {
			return nil, err
		}
		counts[time.Unix(ts, 0)] = count
	}
	return counts, nil
}

func (s *SQLiteStore) GetK8sEventCountsByTime(bucketSize time.Duration) (map[time.Time]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketSeconds := int64(bucketSize.Seconds())
	if bucketSeconds <= 0 {
		bucketSeconds = 60
	}

	query := `
		SELECT 
			((last_timestamp / 1000000) / ?) * ? as bucket, 
			SUM(count) 
		FROM k8s_events 
		WHERE type = 'Warning'
		GROUP BY 1
		ORDER BY 1 ASC
	`

	rows, err := s.db.Query(query, bucketSeconds, bucketSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query k8s event counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[time.Time]int)
	for rows.Next() {
		var ts int64
		var count int
		if err := rows.Scan(&ts, &count); err != nil {
			return nil, err
		}
		counts[time.Unix(ts, 0)] = count
	}
	return counts, nil
}

func (s *SQLiteStore) InsertLog(logEntry *models.LogEntry) error {
	logEntry.Fingerprint = analysis.GenerateFingerprint(logEntry.Message)

	s.mu.Lock()
	defer s.mu.Unlock()

	var ts int64
	if !logEntry.Timestamp.IsZero() {
		ts = logEntry.Timestamp.UnixMicro()
	}

	_, err := s.db.Exec(`
		INSERT INTO logs (timestamp, level, node, shard, component, message, raw, line_number, file_path, fingerprint)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ts, logEntry.Level, logEntry.Node, logEntry.Shard, logEntry.Component, logEntry.Message, logEntry.Raw, logEntry.LineNumber, logEntry.FilePath, logEntry.Fingerprint)
	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}
	return nil
}

func (s *SQLiteStore) BulkInsertLogs(logEntries []*models.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for bulk log insert: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO logs (timestamp, level, node, shard, component, message, raw, line_number, file_path, fingerprint)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for bulk log insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, logEntry := range logEntries {
		logEntry.Fingerprint = analysis.GenerateFingerprint(logEntry.Message)
		var ts int64
		if !logEntry.Timestamp.IsZero() {
			ts = logEntry.Timestamp.UnixMicro()
		}
		_, err := stmt.Exec(ts, logEntry.Level, logEntry.Node, logEntry.Shard, logEntry.Component, logEntry.Message, logEntry.Raw, logEntry.LineNumber, logEntry.FilePath, logEntry.Fingerprint)
		if err != nil {
			return fmt.Errorf("failed to insert log entry during bulk operation: %w", err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) InsertMetric(metric *models.PrometheusMetric, timestamp time.Time) error {
	labelsJSON, err := json.Marshal(metric.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels to JSON: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var ts int64
	if !timestamp.IsZero() {
		ts = timestamp.UnixMicro()
	}

	_, err = s.db.Exec(`
		INSERT INTO metrics (name, labels, value, timestamp)
		VALUES (?, ?, ?, ?)
	`, metric.Name, labelsJSON, metric.Value, ts)
	if err != nil {
		return fmt.Errorf("failed to insert metric entry: %w", err)
	}
	return nil
}

func (s *SQLiteStore) BulkInsertMetrics(metrics []*models.PrometheusMetric, timestamp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for bulk metric insert: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO metrics (name, labels, value, timestamp)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for bulk metric insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	var ts int64
	if !timestamp.IsZero() {
		ts = timestamp.UnixMicro()
	}

	for _, metric := range metrics {
		labelsJSON, err := json.Marshal(metric.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels to JSON during bulk operation: %w", err)
		}
		_, err = stmt.Exec(metric.Name, labelsJSON, metric.Value, ts)
		if err != nil {
			return fmt.Errorf("failed to insert metric entry during bulk operation: %w", err)
		}
	}

	return tx.Commit()
}

// buildFTSQuery parses a user query string and converts it into a safe FTS5 syntax,
// preserving operators (AND, OR, NOT, (, )) and quoting other terms.
// It also supports field:value syntax for specific columns.
func buildFTSQuery(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	operators := map[string]bool{
		"AND": true, "OR": true, "NOT": true, "(": true, ")": true,
	}

	allowedFields := map[string]bool{
		"message":   true,
		"raw":       true,
		"node":      true,
		"component": true,
		"level":     true,
	}

	var tokens []string
	var currentToken strings.Builder
	inQuote := false

	for _, r := range input {
		if r == '"' {
			inQuote = !inQuote
			currentToken.WriteRune(r)
			continue
		}

		if inQuote {
			currentToken.WriteRune(r)
			continue
		}

		// Handle parentheses and whitespace outside quotes
		if r == '(' || r == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(r))
			continue
		}

		if r == ' ' || r == '\t' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		} else {
			currentToken.WriteRune(r)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	// Process tokens
	var processed []string
	for _, token := range tokens {
		upper := strings.ToUpper(token)
		if operators[upper] {
			processed = append(processed, upper)
			continue
		}

		// Check for field:value syntax
		// We look for the first colon
		colonIdx := strings.Index(token, ":")
		if colonIdx > 0 {
			field := token[:colonIdx]
			value := token[colonIdx+1:]
			
			// Check if valid field (case-insensitive match logic, but schema is lowercase)
			// Assuming user types "Level:ERROR", we want to match "level"
			if allowedFields[strings.ToLower(field)] {
				// Handle value quoting
				// If value is already quoted e.g. "foo bar", we strip quotes to re-escape safely
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
					value = value[1 : len(value)-1]
				}
				
				// Escape double quotes in value
				safeValue := strings.ReplaceAll(value, "\"", "\"\"")
				
				// Reconstruct as field:"value"
				// Note: FTS5 column filters are unquoted field names usually
				processed = append(processed, fmt.Sprintf("%s:\"%s\"", strings.ToLower(field), safeValue))
				continue
			}
		}

		// Standard term processing
		if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") && len(token) >= 2 {
			// Already quoted
			inner := token[1 : len(token)-1]
			safeInner := strings.ReplaceAll(inner, "\"", "\"\"")
			processed = append(processed, fmt.Sprintf("\"%s\"", safeInner))
		} else {
			// Term: wrap in quotes and escape
			safeTerm := strings.ReplaceAll(token, "\"", "\"\"")
			processed = append(processed, fmt.Sprintf("\"%s\"", safeTerm))
		}
	}

	return strings.Join(processed, " ")
}

func (s *SQLiteStore) GetLogs(filter *models.LogFilter) ([]*models.LogEntry, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		queryBuilder strings.Builder
		countBuilder strings.Builder
		args         []interface{}
		countArgs    []interface{}
		whereClauses []string
	)

	// Combine Search and Ignore into a single FTS query if either is present
	ftsSearch := buildFTSQuery(filter.Search)
	ftsIgnore := buildFTSQuery(filter.Ignore)

	// If the search query starts with "NOT", treat it as an ignore filter
	// This allows "NOT error" to work as expected (implicitly "everything NOT error")
	// because FTS5 "NOT" is a binary operator and cannot start a query.
	if strings.HasPrefix(ftsSearch, "NOT ") && ftsIgnore == "" {
		ftsIgnore = strings.TrimPrefix(ftsSearch, "NOT ")
		ftsSearch = ""
	}

	var finalFTSQuery string
	if ftsSearch != "" {
		if ftsIgnore != "" {
			finalFTSQuery = fmt.Sprintf("(%s) NOT (%s)", ftsSearch, ftsIgnore)
		} else {
			finalFTSQuery = ftsSearch
		}
	}

	if finalFTSQuery != "" {
		queryBuilder.WriteString("SELECT T1.id, T1.timestamp, T1.level, T1.node, T1.shard, T1.component, T1.message, T1.raw, T1.line_number, T1.file_path, snippet(logs_fts, 0, '<mark>', '</mark>', '...', 64) as snippet FROM logs AS T1 JOIN logs_fts AS T2 ON T1.id = T2.rowid")
		countBuilder.WriteString("SELECT COUNT(T1.id) FROM logs AS T1 JOIN logs_fts AS T2 ON T1.id = T2.rowid")

		whereClauses = append(whereClauses, "T2.logs_fts MATCH ?")

		args = append(args, finalFTSQuery)
		countArgs = append(countArgs, finalFTSQuery)
	} else {
		queryBuilder.WriteString("SELECT id, timestamp, level, node, shard, component, message, raw, line_number, file_path, '' as snippet FROM logs")
		countBuilder.WriteString("SELECT COUNT(*) FROM logs")

		if ftsIgnore != "" {
			whereClauses = append(whereClauses, "id NOT IN (SELECT rowid FROM logs_fts WHERE logs_fts MATCH ?)")
			args = append(args, ftsIgnore)
			countArgs = append(countArgs, ftsIgnore)
		}
	}

	if len(filter.Level) > 0 {
		placeholders := make([]string, len(filter.Level))
		for i := range filter.Level {
			placeholders[i] = "?"
			args = append(args, filter.Level[i])
			countArgs = append(countArgs, filter.Level[i])
		}
		if finalFTSQuery != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("T1.level IN (%s)", strings.Join(placeholders, ",")))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("level IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if len(filter.Node) > 0 {
		placeholders := make([]string, len(filter.Node))
		for i := range filter.Node {
			placeholders[i] = "?"
			args = append(args, filter.Node[i])
			countArgs = append(countArgs, filter.Node[i])
		}
		if finalFTSQuery != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("T1.node IN (%s)", strings.Join(placeholders, ",")))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("node IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if len(filter.Component) > 0 {
		placeholders := make([]string, len(filter.Component))
		for i := range filter.Component {
			placeholders[i] = "?"
			args = append(args, filter.Component[i])
			countArgs = append(countArgs, filter.Component[i])
		}
		if finalFTSQuery != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("T1.component IN (%s)", strings.Join(placeholders, ",")))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("component IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if filter.StartTime != "" {
		t, timeParseErr := time.Parse("2006-01-02T15:04:05", filter.StartTime)
		if timeParseErr != nil {
			t, timeParseErr = time.Parse("2006-01-02T15:04", filter.StartTime)
		}

		if timeParseErr == nil {
			ts := t.UnixMicro()
			if finalFTSQuery != "" {
				whereClauses = append(whereClauses, "T1.timestamp >= ?")
			} else {
				whereClauses = append(whereClauses, "timestamp >= ?")
			}
			args = append(args, ts)
			countArgs = append(countArgs, ts)
		} else {
			slog.Warn("failed to parse start time", "start_time", filter.StartTime, "error", timeParseErr)
		}
	}
	if filter.EndTime != "" {
		t, timeParseErr := time.Parse("2006-01-02T15:04:05", filter.EndTime)
		if timeParseErr != nil {
			t, timeParseErr = time.Parse("2006-01-02T15:04", filter.EndTime)
		}

		if timeParseErr == nil {
			ts := t.UnixMicro()
			if finalFTSQuery != "" {
				whereClauses = append(whereClauses, "T1.timestamp <= ?")
			} else {
				whereClauses = append(whereClauses, "timestamp <= ?")
			}
			args = append(args, ts)
			countArgs = append(countArgs, ts)
		} else {
			slog.Warn("failed to parse end time", "end_time", filter.EndTime, "error", timeParseErr)
		}
	}

	if len(whereClauses) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(whereClauses, " AND "))
		countBuilder.WriteString(" WHERE ")
		countBuilder.WriteString(strings.Join(whereClauses, " AND "))
	}

	// Get total count
	var totalCount int
	err := s.db.QueryRow(countBuilder.String(), countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count of filtered logs: %w", err)
	}

	// Add ORDER BY with proper table prefix
	sortDir := "DESC"
	if strings.ToLower(filter.Sort) == "asc" {
		sortDir = "ASC"
	}

	if finalFTSQuery != "" {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY T1.timestamp %s", sortDir))
	} else {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY timestamp %s", sortDir))
	}

	if filter.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", filter.Limit))
		if filter.Offset > 0 {
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET %d", filter.Offset))
		}
	}

	rows, err := s.db.Query(queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logEntries []*models.LogEntry

	for rows.Next() {
		var logEntry models.LogEntry
		var ts int64
		var id int64
		err = rows.Scan(&id, &ts, &logEntry.Level, &logEntry.Node, &logEntry.Shard, &logEntry.Component, &logEntry.Message, &logEntry.Raw, &logEntry.LineNumber, &logEntry.FilePath, &logEntry.Snippet)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log entry: %w", err)
		}
		if ts != 0 {
			logEntry.Timestamp = time.UnixMicro(ts).UTC()
		}
		logEntries = append(logEntries, &logEntry)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during rows iteration for logs: %w", err)
	}

	return logEntries, totalCount, nil
}

func (s *SQLiteStore) GetMetrics(name string, labels map[string]string, startTime, endTime time.Time, limit, offset int) ([]*models.PrometheusMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		queryBuilder strings.Builder
		args         []interface{}
		whereClauses []string
	)

	queryBuilder.WriteString("SELECT name, labels, value, timestamp FROM metrics")

	whereClauses = append(whereClauses, "name = ?")
	args = append(args, name)

	if len(labels) > 0 {
		for k, v := range labels {
			whereClauses = append(whereClauses, fmt.Sprintf("json_extract(labels, '$.%s') = ?", k))
			args = append(args, v)
		}
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp BETWEEN ? AND ?")
		args = append(args, startTime.UnixMicro(), endTime.UnixMicro())
	} else if !startTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, startTime.UnixMicro())
	} else if !endTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, endTime.UnixMicro())
	}

	if len(whereClauses) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(whereClauses, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY timestamp ASC")

	if limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", limit))
		if offset > 0 {
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET %d", offset))
		}
	}

	rows, err := s.db.Query(queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var metrics []*models.PrometheusMetric
	for rows.Next() {
		var metric models.PrometheusMetric
		var labelsJSON string
		var ts int64
		err := rows.Scan(&metric.Name, &labelsJSON, &metric.Value, &ts)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric entry: %w", err)
		}

		err = json.Unmarshal([]byte(labelsJSON), &metric.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels JSON: %w", err)
		}
		if ts != 0 {
			metric.Timestamp = time.UnixMicro(ts).UTC()
		}
		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for metrics: %w", err)
	}

	return metrics, nil
}

func (s *SQLiteStore) GetLogPatterns() ([]models.LogPattern, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT
			COUNT(*) as count,
			level,
			fingerprint,
			MIN(timestamp) as first_seen,
			MAX(timestamp) as last_seen,
			(SELECT message FROM logs l2 WHERE l2.fingerprint = logs.fingerprint LIMIT 1) as sample_message
		FROM logs
		GROUP BY level, fingerprint
		ORDER BY count DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query log patterns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var patterns []models.LogPattern
	var totalLogs int

	for rows.Next() {
		var p models.LogPattern
		var sampleMessage string
		var firstSeen, lastSeen int64

		err := rows.Scan(&p.Count, &p.Level, &p.Signature, &firstSeen, &lastSeen, &sampleMessage)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log pattern: %w", err)
		}

		if firstSeen != 0 {
			p.FirstSeen = time.UnixMicro(firstSeen).UTC()
		}
		if lastSeen != 0 {
			p.LastSeen = time.UnixMicro(lastSeen).UTC()
		}
		p.SampleEntry = models.LogEntry{Message: sampleMessage}
		patterns = append(patterns, p)
		totalLogs += p.Count
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during rows iteration for patterns: %w", err)
	}

	return patterns, totalLogs, nil
}

func (s *SQLiteStore) Optimize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("Running database optimizations...")

	// 1. Optimize FTS5 index (merges index segments)
	if _, err := s.db.Exec("INSERT INTO logs_fts(logs_fts) VALUES('optimize')"); err != nil {
		slog.Warn("Failed to optimize FTS5 index", "error", err)
	}

	// 2. Run ANALYZE to update query planner statistics
	if _, err := s.db.Exec("ANALYZE"); err != nil {
		slog.Warn("Failed to run ANALYZE", "error", err)
	}

	// 3. Run VACUUM to reclaim space and defragment the database
	// Note: VACUUM can be slow on very large databases, but since this is a one-time
	// post-ingestion task, it's highly beneficial.
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to run VACUUM: %w", err)
	}

	slog.Info("Database optimizations complete")
	return nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) GetMinMaxLogTime() (minTime, maxTime time.Time, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nullMinStr, nullMaxStr sql.NullInt64
	row := s.db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM logs")
	err = row.Scan(&nullMinStr, &nullMaxStr)

	if err == sql.ErrNoRows {
		return time.Time{}, time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to get min/max log times: %w", err)
	}

	if nullMinStr.Valid && nullMinStr.Int64 != 0 {
		minTime = time.UnixMicro(nullMinStr.Int64).UTC()
	}
	if nullMaxStr.Valid && nullMaxStr.Int64 != 0 {
		maxTime = time.UnixMicro(nullMaxStr.Int64).UTC()
	}
	return minTime, maxTime, nil
}

func (s *SQLiteStore) GetDistinctLogNodes() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT DISTINCT node FROM logs ORDER BY node ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct log nodes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var nodes []string
	for rows.Next() {
		var node string
		if err := rows.Scan(&node); err != nil {
			return nil, fmt.Errorf("failed to scan distinct log node: %w", err)
		}
		nodes = append(nodes, node)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for nodes: %w", err)
	}
	
	return nodes, nil
}

func (s *SQLiteStore) GetDistinctLogComponents() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT DISTINCT component FROM logs ORDER BY component ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct log components: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var components []string
	for rows.Next() {
		var component string
		if err := rows.Scan(&component); err != nil {
			return nil, fmt.Errorf("failed to scan distinct log component: %w", err)
		}
		components = append(components, component)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for components: %w", err)
	}
	
	return components, nil
}

func (s *SQLiteStore) GetDistinctMetricNames() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT DISTINCT name FROM metrics ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct metric names: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan metric name: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}



================================================
FILE: internal/store/store.go
================================================
package store

import (
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// Store defines the interface for data storage operations.
type Store interface {
	HasLogs() (bool, error)
	InsertLog(log *models.LogEntry) error
	BulkInsertLogs(logs []*models.LogEntry) error
	GetLogs(filter *models.LogFilter) ([]*models.LogEntry, int, error)
	GetLogPatterns() ([]models.LogPattern, int, error)
	GetMinMaxLogTime() (minTime, maxTime time.Time, err error)
	GetDistinctLogNodes() ([]string, error)
	GetDistinctLogComponents() ([]string, error)
	GetDistinctMetricNames() ([]string, error)

	BulkInsertK8sEvents(events []models.K8sResource) error
	
	GetLogCountsByTime(bucketSize time.Duration) (map[time.Time]int, error)
	GetK8sEventCountsByTime(bucketSize time.Duration) (map[time.Time]int, error)

	InsertMetric(metric *models.PrometheusMetric, timestamp time.Time) error
	BulkInsertMetrics(metrics []*models.PrometheusMetric, timestamp time.Time) error
	GetMetrics(name string, labels map[string]string, startTime, endTime time.Time, limit, offset int) ([]*models.PrometheusMetric, error)

	Optimize() error
	Close() error
}



================================================
FILE: internal/timeline/timeline.go
================================================
package timeline

import (
	"sort"
	"strconv"
	"time"

	"github.com/alextreichler/bundleViewer/internal/models"
)

// EventType defines the source/type of the event
type EventType string

const (
	EventTypeLog     EventType = "Log"
	EventTypeK8s     EventType = "K8s"
	EventTypeRestart EventType = "Restart"
)

// TimelineEvent represents a single point on the timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        EventType `json:"type"`
	Source      string    `json:"source"` // e.g. "Node 1", "Pod redpanda-0"
	Description string    `json:"description"`
	Level       string    `json:"level"` // "Normal", "Warning", "Critical"
	Details     string    `json:"details,omitempty"`
}

// GenerateTimeline creates a merged timeline from various sources
func GenerateTimeline(data *models.TimelineSourceData) []TimelineEvent {
	var events []TimelineEvent

	// 1. Process Logs (Only Critical/Error/Warning)
	for _, log := range data.Logs {
		if log.Level == "ERROR" || log.Level == "WARN" {
			severity := "Warning"
			if log.Level == "ERROR" {
				severity = "Critical"
			}

			events = append(events, TimelineEvent{
				Timestamp:   log.Timestamp,
				Type:        EventTypeLog,
				Source:      log.Node,
				Description: log.Message,
				Level:       severity,
				Details:     log.Raw,
			})
		}
	}

	// 2. Process K8s Events (Resource Lifecycle & System Events)
	for _, event := range data.K8sEvents {
		level := "Normal"
		if event.Type == "Warning" {
			level = "Warning"
		}

		events = append(events, TimelineEvent{
			Timestamp:   event.LastTimestamp,
			Type:        EventTypeK8s,
			Source:      event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
			Description: event.Reason + ": " + event.Message,
			Level:       level,
			Details:     "Source: " + event.Source.Component + "\nCount: " + strconv.Itoa(event.Count),
		})
	}

	// 3. Process Pod Lifecycle (Derived from Pod Status)
	for _, pod := range data.K8sPods {
		for _, status := range pod.Status.ContainerStatuses {
			// Current Running State
			if running, ok := status.State["running"]; ok && !running.StartedAt.IsZero() {
				events = append(events, TimelineEvent{
					Timestamp:   running.StartedAt,
					Type:        EventTypeK8s,
					Source:      pod.Metadata.Name,
					Description: "Pod Started (Running)",
					Level:       "Normal",
					Details:     "Container: " + status.Name,
				})
			}

			// Last Termination State (Crash/Restart)
			if terminated, ok := status.LastState["terminated"]; ok && !terminated.FinishedAt.IsZero() {
				level := "Warning"
				if terminated.Reason == "OOMKilled" || terminated.Reason == "Error" || terminated.ExitCode != 0 {
					level = "Critical"
				}

				events = append(events, TimelineEvent{
					Timestamp:   terminated.FinishedAt,
					Type:        EventTypeRestart,
					Source:      pod.Metadata.Name,
					Description: "Pod Terminated: " + terminated.Reason,
					Level:       level,
					Details:     "Exit Code: " + strconv.Itoa(terminated.ExitCode) + "\nContainer: " + status.Name,
				})
			}
		}
	}

	// Sort by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return events
}



================================================
FILE: k8s/configmaps.json
================================================
{"kind":"ConfigMapList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"kube-root-ca.crt","namespace":"redpanda","uid":"c8518e26-9892-46e3-a804-66881c9d40dc","resourceVersion":"740581239","creationTimestamp":"2023-12-12T20:28:21Z","annotations":{"kubernetes.io/description":"Contains a CA bundle that can be used to verify the kube-apiserver when using internal endpoints such as the internal service IP or kubernetes.default.svc. No other usage is guaranteed across distributions of Kubernetes clusters."},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:28:21Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{".":{},"f:ca.crt":{}},"f:metadata":{"f:annotations":{".":{},"f:kubernetes.io/description":{}}}}}]},"data":{"ca.crt":"-----BEGIN CERTIFICATE-----\nMIIDBTCCAe2gAwIBAgIIBnp+czoXQL0wDQYJKoZIhvcNAQELBQAwFTETMBEGA1UE\nAxMKa3ViZXJuZXRlczAeFw0yMzEyMDcxOTEwMjZaFw0zMzEyMDQxOTE1MjZaMBUx\nEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQDl9RrFRCqw89Drvt2sCQUwWz30Al5Um0QjuQNDLqWh8dkicUrt1cKHavRS\nWE9dcWyy+jq1XO3usxw6iYgSUx971sxlsj+we+JjAYiarCt0KaEGkb2m/D/e+FjF\nYjY31ZM9/7lNSIzAu2Qm2VLsv1reNh1tHB5p5V9OAcPDxoYKKn7IfI0An0vk7vB2\nSb8W6vLVxsnsKLqgAxRmWBgLIkjjNJ3qCuaG3KPH+tF4LH2ZQwI8QI1nwXr8t7Vh\nv6QMYng39FHNqNCqVmECBUiCEAfkDWhu7S8wl8wZgyPZYSRnkrzIHpQI9tj+HCzL\nfXbysmsty0oZbMR5ViKF1KZiiE3VAgMBAAGjWTBXMA4GA1UdDwEB/wQEAwICpDAP\nBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTJvPl5Y9SW5+pz9bEppGM5m1t/yzAV\nBgNVHREEDjAMggprdWJlcm5ldGVzMA0GCSqGSIb3DQEBCwUAA4IBAQBxD+gWQx7V\nuua8HvR8b7YamDjglJR0DZtiMZJslGF+4uwuIHjKmGqw/Mrp9sVV3gC0G6oIJgJg\noShK0ble1SC3NQRwIfn7T0KZCnIAuUEGRJ9TAiWumEpi/P/dplHiC4OeCz0YaRHu\nuFL0ijWKRat2y3TJmR/01ZLIfRIvI9MsaNr+4gV54tbEHd8f80Ny7JcsjKog0VPC\nGR4+i5OOmxmLUIVs/kosX1CiSbDSuxPyt/Lov8DtNJkAlomwwatzJHo/ERrLx72y\nkEhu8n5j/7eqDIbHYyXF39AmB5KTJJ+tmJ46R5Iaz4ftVu/jB3vGSfR7Gv12nFcM\nOCkKDx53agTC\n-----END CERTIFICATE-----\n"}},{"metadata":{"name":"redpanda-console","namespace":"redpanda","uid":"c5d139b5-71fb-4d95-bc53-057770ef2074","resourceVersion":"740672781","creationTimestamp":"2024-06-06T21:45:14Z","labels":{"app.kubernetes.io/instance":"redpanda-console","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"console","app.kubernetes.io/version":"v2.7.0","helm.sh/chart":"console-0.7.29"},"annotations":{"meta.helm.sh/release-name":"redpanda-console","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-22T23:52:26Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{".":{},"f:config.yaml":{}},"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:app.kubernetes.io/version":{},"f:helm.sh/chart":{}}}}}]},"data":{"config.yaml":"# from .Values.console.config\nconnect: {}\nkafka:\n  brokers:\n  - redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local:9093\n  - redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local:9093\n  - redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local:9093\n  sasl:\n    enabled: false\n  schemaRegistry:\n    enabled: false\n    tls:\n      enabled: false\n    urls:\n    - http://redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local:8081\n    - http://redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local:8081\n    - http://redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local:8081\n  tls:\n    enabled: false\nredpanda:\n  adminApi:\n    enabled: true\n    urls:\n    - http://redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - http://redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - http://redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.:9644\n"}},{"metadata":{"name":"redpanda-prod","namespace":"redpanda","uid":"beab07ae-4839-4005-98c3-644a71db8f99","resourceVersion":"1829722703","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","helm.sh/chart":"redpanda-5.10.2"},"annotations":{"meta.helm.sh/release-name":"redpanda-prod","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-09T21:23:21Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{},"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{}}}}},{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-10-16T21:43:52Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{"f:bootstrap.yaml":{},"f:redpanda.yaml":{}},"f:metadata":{"f:labels":{"f:helm.sh/chart":{}}}}}]},"data":{"bootstrap.yaml":"admin_api_require_auth: false\naudit_enabled: false\ncloud_storage_backend: aws\ncloud_storage_bucket: fgp-redpanda-trading-prod\ncloud_storage_cache_size: 5368709120\ncloud_storage_credentials_source: sts\ncloud_storage_enable_remote_read: false\ncloud_storage_enable_remote_write: false\ncloud_storage_enabled: true\ncloud_storage_region: us-east-2\ncompacted_log_segment_size: 67108864\ndefault_topic_replications: 3\nenable_rack_awareness: true\nenable_sasl: true\nkafka_batch_max_bytes: 15728640\nkafka_connection_rate_limit: 1000\nkafka_enable_authorization: true\nlog_segment_size_max: 268435456\nlog_segment_size_min: 16777216\nmax_compacted_log_segment_size: 536870912\npartition_autobalancing_mode: continuous\nstorage_min_free_bytes: 5368709120\nsuperusers:\n- kubernetes-controller","redpanda.yaml":"config_file: /etc/redpanda/redpanda.yaml\npandaproxy:\n  pandaproxy_api:\n  - address: 0.0.0.0\n    authentication_method: http_basic\n    name: internal\n    port: 8082\n  - address: 0.0.0.0\n    authentication_method: http_basic\n    name: default\n    port: 8083\n  pandaproxy_api_tls: null\npandaproxy_client:\n  brokers:\n  - address: redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\nredpanda:\n  admin:\n  - address: 0.0.0.0\n    name: internal\n    port: 9644\n  - address: 0.0.0.0\n    name: default\n    port: 9645\n  admin_api_tls: null\n  crash_loop_limit: 5\n  empty_seed_starts_cluster: false\n  kafka_api:\n  - address: 0.0.0.0\n    authentication_method: none\n    name: internal\n    port: 9093\n  - address: 0.0.0.0\n    authentication_method: sasl\n    name: default\n    port: 9094\n  kafka_api_tls: null\n  rpc_server:\n    address: 0.0.0.0\n    port: 33145\n  seed_servers:\n  - host:\n      address: redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.\n      port: 33145\n  - host:\n      address: redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.\n      port: 33145\n  - host:\n      address: redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.\n      port: 33145\n  - host:\n      address: redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.\n      port: 33145\n  - host:\n      address: redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.\n      port: 33145\nrpk:\n  additional_start_flags:\n  - --default-log-level=info\n  - --memory=18432M\n  - --reserve-memory=1024M\n  - --smp=4\n  admin_api:\n    addresses:\n    - redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.:9644\n    - redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.:9644\n    tls: null\n  enable_memory_locking: false\n  kafka_api:\n    brokers:\n    - redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.:9093\n    - redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.:9093\n    - redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.:9093\n    - redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.:9093\n    - redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.:9093\n    tls: null\n  overprovisioned: false\n  schema_registry:\n    addresses:\n    - redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.:8081\n    - redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.:8081\n    - redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.:8081\n    - redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.:8081\n    - redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.:8081\n    tls: null\n  tune_aio_events: true\nschema_registry:\n  schema_registry_api:\n  - address: 0.0.0.0\n    authentication_method: http_basic\n    name: internal\n    port: 8081\n  - address: 0.0.0.0\n    authentication_method: http_basic\n    name: default\n    port: 8084\n  schema_registry_api_tls: null\nschema_registry_client:\n  brokers:\n  - address: redpanda-prod-0.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-1.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-2.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-3.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093\n  - address: redpanda-prod-4.redpanda-prod.redpanda.svc.cluster.local.\n    port: 9093"}},{"metadata":{"name":"redpanda-prod-rpk","namespace":"redpanda","uid":"9f4ec8f1-d87d-4d45-8a16-cffc0d538500","resourceVersion":"1093374120","creationTimestamp":"2024-01-02T22:36:38Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","helm.sh/chart":"redpanda-5.10.2"},"annotations":{"meta.helm.sh/release-name":"redpanda-prod","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-09T21:23:21Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{},"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{}}}}},{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-06-05T23:14:47Z","fieldsType":"FieldsV1","fieldsV1":{"f:data":{"f:profile":{}},"f:metadata":{"f:labels":{"f:helm.sh/chart":{}}}}}]},"data":{"profile":"admin_api:\n  addresses:\n  - redpanda-prod-0.prod.internalfgp.com:31644\n  - redpanda-prod-1.prod.internalfgp.com:31644\n  - redpanda-prod-2.prod.internalfgp.com:31644\n  - redpanda-prod-3.prod.internalfgp.com:31644\n  - redpanda-prod-4.prod.internalfgp.com:31644\n  tls: null\nkafka_api:\n  brokers:\n  - redpanda-prod-0.prod.internalfgp.com:31092\n  - redpanda-prod-1.prod.internalfgp.com:31092\n  - redpanda-prod-2.prod.internalfgp.com:31092\n  - redpanda-prod-3.prod.internalfgp.com:31092\n  - redpanda-prod-4.prod.internalfgp.com:31092\n  tls: null\nname: default\nschema_registry:\n  addresses:\n  - redpanda-prod-0.prod.internalfgp.com:30081\n  - redpanda-prod-1.prod.internalfgp.com:30081\n  - redpanda-prod-2.prod.internalfgp.com:30081\n  - redpanda-prod-3.prod.internalfgp.com:30081\n  - redpanda-prod-4.prod.internalfgp.com:30081\n  tls: null"}}]}



================================================
FILE: k8s/endpoints.json
================================================
{"kind":"EndpointsList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"redpanda-console","namespace":"redpanda","uid":"fb3ca088-6bf1-4bd3-8306-daa3fb5203d0","resourceVersion":"1965470907","creationTimestamp":"2024-06-06T21:45:14Z","labels":{"app.kubernetes.io/instance":"redpanda-console","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"console","app.kubernetes.io/version":"v2.7.0","endpoints.kubernetes.io/managed-by":"endpoint-controller","helm.sh/chart":"console-0.7.29"},"annotations":{"endpoints.kubernetes.io/last-change-trigger-time":"2025-11-04T14:35:18Z"},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-11-04T14:35:18Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:endpoints.kubernetes.io/last-change-trigger-time":{}},"f:labels":{".":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:app.kubernetes.io/version":{},"f:endpoints.kubernetes.io/managed-by":{},"f:helm.sh/chart":{}}},"f:subsets":{}}}]},"subsets":[{"addresses":[{"ip":"10.50.174.119","nodeName":"ip-10-50-175-121.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-console-849c45569f-l4j2t","uid":"8e051bbf-6e47-49b0-860a-7065c57b15f1"}}],"ports":[{"name":"http","port":8080,"protocol":"TCP"}]}]},{"metadata":{"name":"redpanda-prod","namespace":"redpanda","uid":"df0fcbf8-0058-4c0c-b8d9-e52e5b337a63","resourceVersion":"1829747608","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","endpoints.kubernetes.io/managed-by":"endpoint-controller","helm.sh/chart":"redpanda-5.10.2","monitoring.redpanda.com/enabled":"false","service.kubernetes.io/headless":""},"annotations":{"endpoints.kubernetes.io/last-change-trigger-time":"2025-10-16T21:47:59Z"},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-10-16T21:48:06Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:endpoints.kubernetes.io/last-change-trigger-time":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:endpoints.kubernetes.io/managed-by":{},"f:helm.sh/chart":{},"f:monitoring.redpanda.com/enabled":{},"f:service.kubernetes.io/headless":{}}},"f:subsets":{}}}]},"subsets":[{"addresses":[{"ip":"10.50.160.85","hostname":"redpanda-prod-2","nodeName":"ip-10-50-161-245.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-2","uid":"e285d9fd-0612-4a47-973e-72cd8ff6a21e"}},{"ip":"10.50.167.165","hostname":"redpanda-prod-4","nodeName":"ip-10-50-165-44.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-4","uid":"111b1c43-e7ee-4df1-b52c-57b355aa13e0"}},{"ip":"10.50.206.101","hostname":"redpanda-prod-1","nodeName":"ip-10-50-215-28.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-1","uid":"d509b389-40ac-4b91-8198-77921ee159b4"}},{"ip":"10.50.251.212","hostname":"redpanda-prod-0","nodeName":"ip-10-50-236-103.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-0","uid":"389570a6-acc6-4c0c-912e-c7827f7ac170"}},{"ip":"10.50.255.37","hostname":"redpanda-prod-3","nodeName":"ip-10-50-255-165.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-3","uid":"bf354fdf-707f-4d15-9925-649a0dac8bc4"}}],"ports":[{"name":"http","port":8082,"protocol":"TCP"},{"name":"rpc","port":33145,"protocol":"TCP"},{"name":"kafka","port":9093,"protocol":"TCP"},{"name":"schemaregistry","port":8081,"protocol":"TCP"},{"name":"admin","port":9644,"protocol":"TCP"}]}]},{"metadata":{"name":"redpanda-prod-external","namespace":"redpanda","uid":"fdcab346-edbe-482d-bd4d-25dbcf993670","resourceVersion":"1829747607","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","endpoints.kubernetes.io/managed-by":"endpoint-controller","helm.sh/chart":"redpanda-5.10.2"},"annotations":{"endpoints.kubernetes.io/last-change-trigger-time":"2025-10-16T21:47:59Z"},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-10-16T21:48:06Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:endpoints.kubernetes.io/last-change-trigger-time":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:endpoints.kubernetes.io/managed-by":{},"f:helm.sh/chart":{}}},"f:subsets":{}}}]},"subsets":[{"addresses":[{"ip":"10.50.160.85","nodeName":"ip-10-50-161-245.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-2","uid":"e285d9fd-0612-4a47-973e-72cd8ff6a21e"}},{"ip":"10.50.167.165","nodeName":"ip-10-50-165-44.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-4","uid":"111b1c43-e7ee-4df1-b52c-57b355aa13e0"}},{"ip":"10.50.206.101","nodeName":"ip-10-50-215-28.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-1","uid":"d509b389-40ac-4b91-8198-77921ee159b4"}},{"ip":"10.50.251.212","nodeName":"ip-10-50-236-103.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-0","uid":"389570a6-acc6-4c0c-912e-c7827f7ac170"}},{"ip":"10.50.255.37","nodeName":"ip-10-50-255-165.us-east-2.compute.internal","targetRef":{"kind":"Pod","namespace":"redpanda","name":"redpanda-prod-3","uid":"bf354fdf-707f-4d15-9925-649a0dac8bc4"}}],"ports":[{"name":"schema-default","port":8084,"protocol":"TCP"},{"name":"http-default","port":8083,"protocol":"TCP"},{"name":"admin-default","port":9645,"protocol":"TCP"},{"name":"kafka-default","port":9094,"protocol":"TCP"}]}]}]}



================================================
FILE: k8s/events.json
================================================
{"kind":"EventList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[]}



================================================
FILE: k8s/limitranges.json
================================================
{"kind":"LimitRangeList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"namespace-contraints-redpanda","namespace":"redpanda","uid":"f6de641e-be8a-4bc7-acc0-f876af1e9f0f","resourceVersion":"94769173","creationTimestamp":"2024-04-04T20:42:19Z","labels":{"app.kubernetes.io/managed-by":"Helm"},"annotations":{"meta.helm.sh/release-name":"namespace-cfg","meta.helm.sh/release-namespace":"default"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-04-04T20:42:19Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/managed-by":{}}},"f:spec":{"f:limits":{}}}}]},"spec":{"limits":[{"type":"Container","default":{"cpu":"256m","memory":"128Mi"},"defaultRequest":{"cpu":"256m","memory":"128Mi"}}]}}]}



================================================
FILE: k8s/persistentvolumeclaims.json
================================================
{"kind":"PersistentVolumeClaimList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"datadir-redpanda-prod-0","namespace":"redpanda","uid":"f69c9c71-6bde-49b2-830b-cf4143a7d6f6","resourceVersion":"1106196940","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"annotations":{"ebs.csi.aws.com/volumeType":"gp3","pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes","volume.beta.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/selected-node":"ip-10-50-34-231.us-east-2.compute.internal","volume.kubernetes.io/storage-provisioner":"ebs.csi.aws.com"},"finalizers":["kubernetes.io/pvc-protection"],"managedFields":[{"manager":"kube-scheduler","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:31Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:volume.kubernetes.io/selected-node":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{},"f:volume.beta.kubernetes.io/storage-provisioner":{},"f:volume.kubernetes.io/storage-provisioner":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{}},"f:storageClassName":{},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{},"f:phase":{}}},"subresource":"status"},{"manager":"kubelet","operation":"Update","apiVersion":"v1","time":"2025-01-03T00:01:27Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:capacity":{"f:storage":{}}}},"subresource":"status"},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-09T20:53:51Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:ebs.csi.aws.com/volumeType":{}}},"f:spec":{"f:resources":{"f:requests":{"f:storage":{}}}}}}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1000Gi"}},"volumeName":"pvc-f69c9c71-6bde-49b2-830b-cf4143a7d6f6","storageClassName":"gp2-xfs","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"2000Gi"}}},{"metadata":{"name":"datadir-redpanda-prod-1","namespace":"redpanda","uid":"da7a2553-fd11-4225-b34e-45487ffffa2f","resourceVersion":"1106255465","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"annotations":{"ebs.csi.aws.com/volumeType":"gp3","pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes","volume.beta.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/selected-node":"ip-10-50-20-203.us-east-2.compute.internal","volume.kubernetes.io/storage-provisioner":"ebs.csi.aws.com"},"finalizers":["kubernetes.io/pvc-protection"],"managedFields":[{"manager":"kube-scheduler","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:31Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:volume.kubernetes.io/selected-node":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{},"f:volume.beta.kubernetes.io/storage-provisioner":{},"f:volume.kubernetes.io/storage-provisioner":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{}},"f:storageClassName":{},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{},"f:phase":{}}},"subresource":"status"},{"manager":"kubelet","operation":"Update","apiVersion":"v1","time":"2025-01-02T23:59:41Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:capacity":{"f:storage":{}}}},"subresource":"status"},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-09T21:19:41Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:ebs.csi.aws.com/volumeType":{}}},"f:spec":{"f:resources":{"f:requests":{"f:storage":{}}}}}}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1000Gi"}},"volumeName":"pvc-da7a2553-fd11-4225-b34e-45487ffffa2f","storageClassName":"gp2-xfs","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"2000Gi"}}},{"metadata":{"name":"datadir-redpanda-prod-2","namespace":"redpanda","uid":"318e9ea7-52f5-4428-ad8e-0e16148dd5d3","resourceVersion":"1107086531","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"annotations":{"ebs.csi.aws.com/volumeType":"gp3","pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes","volume.beta.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/selected-node":"ip-10-50-5-206.us-east-2.compute.internal","volume.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/storage-resizer":"ebs.csi.aws.com"},"finalizers":["kubernetes.io/pvc-protection"],"managedFields":[{"manager":"kube-scheduler","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:31Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:volume.kubernetes.io/selected-node":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2023-12-12T20:32:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{},"f:volume.beta.kubernetes.io/storage-provisioner":{},"f:volume.kubernetes.io/storage-provisioner":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{}},"f:storageClassName":{},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2024-07-01T21:31:46Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:volume.kubernetes.io/storage-resizer":{}}},"f:status":{"f:accessModes":{},"f:capacity":{},"f:phase":{}}},"subresource":"status"},{"manager":"kubelet","operation":"Update","apiVersion":"v1","time":"2025-01-03T00:00:14Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:capacity":{"f:storage":{}}}},"subresource":"status"},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-10T02:54:07Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:ebs.csi.aws.com/volumeType":{}}},"f:spec":{"f:resources":{"f:requests":{"f:storage":{}}}}}}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1000Gi"}},"volumeName":"pvc-318e9ea7-52f5-4428-ad8e-0e16148dd5d3","storageClassName":"gp2-xfs","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"2000Gi"}}},{"metadata":{"name":"datadir-redpanda-prod-3","namespace":"redpanda","uid":"124c386f-c60e-4e47-8d79-0d7b04d3c01e","resourceVersion":"1106895771","creationTimestamp":"2025-03-31T22:00:38Z","labels":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"annotations":{"ebs.csi.aws.com/volumeType":"gp3","pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes","volume.beta.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/selected-node":"ip-10-50-41-246.us-east-2.compute.internal","volume.kubernetes.io/storage-provisioner":"ebs.csi.aws.com"},"finalizers":["kubernetes.io/pvc-protection"],"managedFields":[{"manager":"kube-scheduler","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:25Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:volume.kubernetes.io/selected-node":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:27Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{},"f:volume.beta.kubernetes.io/storage-provisioner":{},"f:volume.kubernetes.io/storage-provisioner":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{".":{},"f:storage":{}}},"f:storageClassName":{},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:27Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-10T01:27:40Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:ebs.csi.aws.com/volumeType":{}}}}}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2000Gi"}},"volumeName":"pvc-124c386f-c60e-4e47-8d79-0d7b04d3c01e","storageClassName":"gp2-xfs","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"2000Gi"}}},{"metadata":{"name":"datadir-redpanda-prod-4","namespace":"redpanda","uid":"6662ff51-3646-4055-8093-be9d8fc100b7","resourceVersion":"1108278794","creationTimestamp":"2025-03-31T22:00:38Z","labels":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"annotations":{"ebs.csi.aws.com/volumeType":"gp3","pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes","volume.beta.kubernetes.io/storage-provisioner":"ebs.csi.aws.com","volume.kubernetes.io/selected-node":"ip-10-50-9-237.us-east-2.compute.internal","volume.kubernetes.io/storage-provisioner":"ebs.csi.aws.com"},"finalizers":["kubernetes.io/pvc-protection"],"managedFields":[{"manager":"kube-scheduler","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:16Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:volume.kubernetes.io/selected-node":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:33Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{},"f:volume.beta.kubernetes.io/storage-provisioner":{},"f:volume.kubernetes.io/storage-provisioner":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{".":{},"f:storage":{}}},"f:storageClassName":{},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-03-31T22:01:33Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-10T11:29:52Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:ebs.csi.aws.com/volumeType":{}}}}}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2000Gi"}},"volumeName":"pvc-6662ff51-3646-4055-8093-be9d8fc100b7","storageClassName":"gp2-xfs","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"2000Gi"}}}]}



================================================
FILE: k8s/replicationcontrollers.json
================================================
{"kind":"ReplicationControllerList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[]}



================================================
FILE: k8s/resourcequotas.json
================================================
{"kind":"ResourceQuotaList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"primary-quota-redpanda","namespace":"redpanda","uid":"d8f6cbe0-1b44-4aba-9a64-da8e533cb8c5","resourceVersion":"1829752095","creationTimestamp":"2024-04-04T20:42:19Z","labels":{"app.kubernetes.io/managed-by":"Helm"},"annotations":{"meta.helm.sh/release-name":"namespace-cfg","meta.helm.sh/release-namespace":"default"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-02T21:18:18Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/managed-by":{}}},"f:spec":{"f:hard":{}}}},{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-02-13T16:59:44Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:hard":{"f:limits.cpu":{},"f:limits.memory":{},"f:requests.cpu":{},"f:requests.memory":{}}}}},{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2025-10-16T21:48:51Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:hard":{".":{},"f:limits.cpu":{},"f:limits.memory":{},"f:requests.cpu":{},"f:requests.memory":{}},"f:used":{".":{},"f:limits.cpu":{},"f:limits.memory":{},"f:requests.cpu":{},"f:requests.memory":{}}}},"subresource":"status"}]},"spec":{"hard":{"limits.cpu":"100","limits.memory":"400Gi","requests.cpu":"100","requests.memory":"400Gi"}},"status":{"hard":{"limits.cpu":"100","limits.memory":"400Gi","requests.cpu":"100","requests.memory":"400Gi"},"used":{"limits.cpu":"47280m","limits.memory":"197504Mi","requests.cpu":"34480m","requests.memory":"197504Mi"}}}]}



================================================
FILE: k8s/serviceaccounts.json
================================================
{"kind":"ServiceAccountList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"default","namespace":"redpanda","uid":"f8aaf6c3-d460-4c19-bb81-a25542fddac7","resourceVersion":"1442543","creationTimestamp":"2023-12-12T20:28:21Z"}},{"metadata":{"name":"redpanda-console","namespace":"redpanda","uid":"7b800b51-cde7-42f3-aedd-dc568558107c","resourceVersion":"448654066","creationTimestamp":"2024-06-06T21:45:14Z","labels":{"app.kubernetes.io/instance":"redpanda-console","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"console","app.kubernetes.io/version":"v2.7.0","helm.sh/chart":"console-0.7.29"},"annotations":{"meta.helm.sh/release-name":"redpanda-console","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-22T23:52:26Z","fieldsType":"FieldsV1","fieldsV1":{"f:automountServiceAccountToken":{},"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:app.kubernetes.io/version":{},"f:helm.sh/chart":{}}}}}]},"automountServiceAccountToken":true},{"metadata":{"name":"redpanda-prod","namespace":"redpanda","uid":"1069e509-4314-4a76-a201-37cc37929058","resourceVersion":"1176165519","creationTimestamp":"2025-06-05T23:14:47Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","helm.sh/chart":"redpanda-5.10.2"},"annotations":{"eks.amazonaws.com/role-arn":"arn:aws:iam::482091280812:role/redpanda-trading-prod","meta.helm.sh/release-name":"redpanda-prod","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-06-05T23:14:47Z","fieldsType":"FieldsV1","fieldsV1":{"f:automountServiceAccountToken":{},"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:helm.sh/chart":{}}}}},{"manager":"kubectl-edit","operation":"Update","apiVersion":"v1","time":"2025-06-28T15:00:39Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:eks.amazonaws.com/role-arn":{}}}}}]},"automountServiceAccountToken":false},{"metadata":{"name":"redpanda-prod-archiver","namespace":"redpanda","uid":"d9a3c242-14ca-49fe-a22a-23fce4c8f6d4","resourceVersion":"211281562","creationTimestamp":"2024-06-20T20:46:17Z","labels":{"app.kubernetes.io/managed-by":"Helm"},"annotations":{"eks.amazonaws.com/role-arn":"arn:aws:iam::482091280812:role/redpanda-trading-prod","meta.helm.sh/release-name":"kafka-s3-archiver","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-06-20T20:46:17Z","fieldsType":"FieldsV1","fieldsV1":{"f:automountServiceAccountToken":{},"f:metadata":{"f:annotations":{".":{},"f:eks.amazonaws.com/role-arn":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/managed-by":{}}}}}]},"automountServiceAccountToken":true}]}



================================================
FILE: k8s/services.json
================================================
{"kind":"ServiceList","apiVersion":"v1","metadata":{"resourceVersion":"1988937923"},"items":[{"metadata":{"name":"redpanda-console","namespace":"redpanda","uid":"3975b7a8-4cf7-4a94-b2fd-0461ac922248","resourceVersion":"448654069","creationTimestamp":"2024-06-06T21:45:14Z","labels":{"app.kubernetes.io/instance":"redpanda-console","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"console","app.kubernetes.io/version":"v2.7.0","helm.sh/chart":"console-0.7.29"},"annotations":{"meta.helm.sh/release-name":"redpanda-console","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-22T23:52:26Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:app.kubernetes.io/version":{},"f:helm.sh/chart":{}}},"f:spec":{"f:internalTrafficPolicy":{},"f:ports":{".":{},"k:{\"port\":8080,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}}},"f:selector":{},"f:sessionAffinity":{},"f:type":{}}}}]},"spec":{"ports":[{"name":"http","protocol":"TCP","port":8080,"targetPort":8080}],"selector":{"app.kubernetes.io/instance":"redpanda-console","app.kubernetes.io/name":"console"},"clusterIP":"172.20.117.31","clusterIPs":["172.20.117.31"],"type":"ClusterIP","sessionAffinity":"None","ipFamilies":["IPv4"],"ipFamilyPolicy":"SingleStack","internalTrafficPolicy":"Cluster"},"status":{"loadBalancer":{}}},{"metadata":{"name":"redpanda-prod","namespace":"redpanda","uid":"83a10a28-611b-41b1-bf37-3cff221b7603","resourceVersion":"1093374141","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","helm.sh/chart":"redpanda-5.10.2","monitoring.redpanda.com/enabled":"false"},"annotations":{"external-dns.alpha.kubernetes.io/endpoints-type":"HostIP","external-dns.alpha.kubernetes.io/hostname":"prod.internalfgp.com","meta.helm.sh/release-name":"redpanda-prod","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-09T21:23:21Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:external-dns.alpha.kubernetes.io/endpoints-type":{},"f:external-dns.alpha.kubernetes.io/hostname":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{},"f:monitoring.redpanda.com/enabled":{}}},"f:spec":{"f:clusterIP":{},"f:internalTrafficPolicy":{},"f:ports":{".":{},"k:{\"port\":8081,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":8082,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":9093,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":9644,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":33145,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:port":{},"f:protocol":{},"f:targetPort":{}}},"f:publishNotReadyAddresses":{},"f:selector":{},"f:sessionAffinity":{},"f:type":{}}}},{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-06-05T23:14:47Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:helm.sh/chart":{}}}}}]},"spec":{"ports":[{"name":"admin","protocol":"TCP","port":9644,"targetPort":9644},{"name":"http","protocol":"TCP","port":8082,"targetPort":8082},{"name":"kafka","protocol":"TCP","port":9093,"targetPort":9093},{"name":"rpc","protocol":"TCP","port":33145,"targetPort":33145},{"name":"schemaregistry","protocol":"TCP","port":8081,"targetPort":8081}],"selector":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"clusterIP":"None","clusterIPs":["None"],"type":"ClusterIP","sessionAffinity":"None","publishNotReadyAddresses":true,"ipFamilies":["IPv4"],"ipFamilyPolicy":"SingleStack","internalTrafficPolicy":"Cluster"},"status":{"loadBalancer":{}}},{"metadata":{"name":"redpanda-prod-external","namespace":"redpanda","uid":"43ffb74d-7216-4b91-b90e-4426edbd3330","resourceVersion":"1093374138","creationTimestamp":"2023-12-12T20:32:31Z","labels":{"app.kubernetes.io/component":"redpanda","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"redpanda","helm.sh/chart":"redpanda-5.10.2"},"annotations":{"meta.helm.sh/release-name":"redpanda-prod","meta.helm.sh/release-namespace":"redpanda"},"managedFields":[{"manager":"terraform-provider-helm_v2.11.0_x5","operation":"Update","apiVersion":"v1","time":"2024-10-09T21:23:21Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:meta.helm.sh/release-name":{},"f:meta.helm.sh/release-namespace":{}},"f:labels":{".":{},"f:app.kubernetes.io/component":{},"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/managed-by":{},"f:app.kubernetes.io/name":{}}},"f:spec":{"f:externalTrafficPolicy":{},"f:internalTrafficPolicy":{},"f:ports":{".":{},"k:{\"port\":8083,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:nodePort":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":8084,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:nodePort":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":9094,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:nodePort":{},"f:port":{},"f:protocol":{},"f:targetPort":{}},"k:{\"port\":9645,\"protocol\":\"TCP\"}":{".":{},"f:name":{},"f:nodePort":{},"f:port":{},"f:protocol":{},"f:targetPort":{}}},"f:publishNotReadyAddresses":{},"f:selector":{},"f:sessionAffinity":{},"f:type":{}}}},{"manager":"terraform-provider-helm_v2.15.0_x5","operation":"Update","apiVersion":"v1","time":"2025-06-05T23:14:47Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:labels":{"f:helm.sh/chart":{}}}}}]},"spec":{"ports":[{"name":"admin-default","protocol":"TCP","port":9645,"targetPort":9645,"nodePort":31644},{"name":"kafka-default","protocol":"TCP","port":9094,"targetPort":9094,"nodePort":31092},{"name":"http-default","protocol":"TCP","port":8083,"targetPort":8083,"nodePort":30082},{"name":"schema-default","protocol":"TCP","port":8084,"targetPort":8084,"nodePort":30081}],"selector":{"app.kubernetes.io/component":"redpanda-statefulset","app.kubernetes.io/instance":"redpanda-prod","app.kubernetes.io/name":"redpanda"},"clusterIP":"172.20.28.112","clusterIPs":["172.20.28.112"],"type":"NodePort","sessionAffinity":"None","externalTrafficPolicy":"Local","publishNotReadyAddresses":true,"ipFamilies":["IPv4"],"ipFamilyPolicy":"SingleStack","internalTrafficPolicy":"Cluster"},"status":{"loadBalancer":{}}}]}



================================================
FILE: ui/ui.go
================================================
package ui

import "embed"

//go:embed all:html
var HTMLFiles embed.FS

//go:embed all:static
var StaticFiles embed.FS



================================================
FILE: ui/html/diagnostics.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Diagnostics - Bundle Viewer</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
	<script src="/static/htmx.min.js"></script>
    <style>
        .status-pass { color: var(--success-color, #28a745); font-weight: bold; }
        .status-warning { color: var(--warning-color, #ffc107); font-weight: bold; }
        .status-critical { color: var(--danger-color, #dc3545); font-weight: bold; }
        .status-info { color: var(--info-color, #17a2b8); font-weight: bold; }
        
        .diag-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        .diag-table th, .diag-table td {
            border: 1px solid var(--border-color);
            padding: 10px;
            text-align: left;
        }
        .diag-table th {
            background-color: var(--header-bg-color);
        }
    </style>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "Diagnostics" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}

	<h2>Node: {{.NodeHostname}} - Diagnostics Report</h2>

    <p>This report runs automated checks against best-practice configurations and system health indicators.</p>

    {{if .Syslog.OOMEvents}}
    <div style="margin-bottom: 20px; border: 1px solid #dc3545; border-radius: 8px; padding: 20px; background-color: rgba(220, 53, 69, 0.05);">
        <h3 style="color: #dc3545; margin-top: 0;">⚠️ Critical: OOM Killer Events Detected</h3>
        <p>The system Out-Of-Memory (OOM) killer has terminated processes. This indicates severe memory pressure.</p>
        <table class="diag-table">
            <thead>
                <tr>
                    <th>Kernel Time</th>
                    <th>Invoked By</th>
                    <th>Message</th>
                </tr>
            </thead>
            <tbody>
                {{range .Syslog.OOMEvents}}
                <tr>
                    <td style="font-family: monospace;">{{.Timestamp}}</td>
                    <td>{{.ProcessName}}</td>
                    <td>{{.Message}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}

    <table class="diag-table">
        <thead>
            <tr>
                <th>Status</th>
                <th>Category</th>
                <th>Check</th>
                <th>Current Value</th>
                <th>Expected</th>
                <th>Recommendation</th>
            </tr>
        </thead>
        <tbody>
            {{range .Report.Results}}
            <tr>
                <td class="status-{{.Status | lower}}">{{.Status}}</td>
                <td>{{.Category}}</td>
                <td title="{{.Description}}">{{.Name}}</td>
                <td>{{.CurrentValue}}</td>
                <td>{{.ExpectedValue}}</td>
                <td>{{.Remediation}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

</body>
</html>



================================================
FILE: ui/html/disk.tmpl
================================================
<!DOCTYPE html>
<html>
<head>
	<title>Disk Overview</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
</head>
<body>
	<div id="loading-overlay">
		<div id="loading-box">
			<h2>Loading...</h2>
			<div id="progress-container">
				<div id="progress-bar">0%</div>
			</div>
		</div>
	</div>
    {{template "nav" (dict "Active" "Disk" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
	<h1>Disk Overview</h1>
	<h2>Node: {{.NodeHostname}}</h2>

	<h2>Overall Disk Usage</h2>
	<table>
		<tbody>
			<tr>
				<th>Total Data Disk Space</th>
				<td>{{.DataDiskStats.TotalBytes | formatBytes}}</td>
			</tr>
			<tr>
				<th>Free Data Disk Space</th>
				<td>{{.DataDiskStats.FreeBytes | formatBytes}}</td>
			</tr>
			<tr>
				<th>Used Data Disk Space</th>
				<td>{{ (sub .DataDiskStats.TotalBytes .DataDiskStats.FreeBytes) | formatBytes}}</td>
			</tr>
			<tr>
				<th>Overall Free Space (from resource-usage.json)</th>
				<td>{{ (mul .ResourceUsage.FreeSpaceMB 1048576.0) | formatBytesFloat}}</td>
			</tr>
		</tbody>
	</table>

	<h2>Disk Usage per Topic</h2>
	<table>
		<thead>
			<tr>
				<th>Topic Name</th>
				<th>Total Size</th>
			</tr>
		</thead>
		<tbody>
			{{range .TopicsDiskInfo}}
			<tr>
				<td><a href="#" onclick="toggleVisibility('{{.Name}}-details'); return false;">{{.Name}}</a></td>
				<td>{{.TotalSize | formatBytes}}</td>
			</tr>
			<tr id="{{.Name}}-details" style="display:none;">
				<td colspan="2">
					<table>
						<thead>
							<tr>
								<th>Partition ID</th>
								<th>Size</th>
							</tr>
						</thead>
						<tbody>
							{{range $partitionID, $size := .Partitions}}
							<tr>
								<td>{{$partitionID}}</td>
								<td>{{$size | formatBytes}}</td>
							</tr>
							{{end}}
						</tbody>
					</table>
				</td>
			</tr>
			{{end}}
		</tbody>
	</table>

	<script src="/static/common.js"></script>
	<script src="/static/progress.js"></script>
</body>
</html>



================================================
FILE: ui/html/groups.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Consumer Groups</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
</head>
<body>
    {{template "nav" (dict "Active" "Groups" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    <div class="container">
        <h1>Consumer Groups</h1>

        <div class="card" style="margin-bottom: 1.5rem; padding: 1rem;">
            <div style="display: flex; gap: 1rem; align-items: center;">
                <div style="flex-grow: 1;">
                    <input type="text" id="groupSearch" placeholder="Search by Group ID or Topic name..." 
                           onkeyup="filterGroups()" 
                           style="width: 100%; padding: 0.75rem; font-size: 1rem;">
                </div>
                <div id="searchStats" style="color: var(--text-color-muted); font-size: 0.875rem; min-width: 120px; text-align: right;">
                    <!-- Filled by JS -->
                </div>
            </div>
        </div>
        
        {{if .ConsumerGroups}}
            <table class="data-table" id="groupsTable">
                <thead>
                    <tr>
                        <th onclick="sortTable(0)" style="cursor: pointer;">Group ID <span class="sort-icon">↕</span></th>
                        <th onclick="sortTable(1)" style="cursor: pointer;">State <span class="sort-icon">↕</span></th>
                        <th onclick="sortTable(2)" style="cursor: pointer;">Protocol <span class="sort-icon">↕</span></th>
                        <th onclick="sortTable(3)" style="cursor: pointer;">Coordinator <span class="sort-icon">↕</span></th>
                        <th onclick="sortTable(4)" style="cursor: pointer;">Members <span class="sort-icon">↕</span></th>
                        <th onclick="sortTable(5)" style="cursor: pointer;">Total Lag <span class="sort-icon">↕</span></th>
                        <th>Details</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .ConsumerGroups}}
                    <tr class="group-row" 
                        data-group="{{lower .Group}}" 
                        data-topics="{{range .Offsets}}{{lower .Topic}} {{end}}">
                        <td style="font-family: var(--font-mono); font-weight: 600;">{{.Group}}</td>
                        <td><span class="badge">{{.State}}</span></td>
                        <td>{{.Protocol}} <small style="color: var(--text-color-muted)">({{.ProtocolType}})</small></td>
                        <td>Node {{.Coordinator.NodeID}} <br><small style="color: var(--text-color-muted)">{{.Coordinator.Host}}</small></td>
                        <td>{{len .Members}}</td>
                        <td data-sort="{{.TotalLag}}">
                            {{if gt .TotalLag 0}}
                                <span style="color: var(--log-level-WARN); font-weight: 600;">{{.TotalLag}}</span>
                            {{else}}
                                <span style="color: var(--text-color-muted);">0</span>
                            {{end}}
                        </td>
                        <td>
                            <button class="button-sm" onclick="toggleVisibility('details-{{.Group}}', this)">View Details</button>
                        </td>
                    </tr>
                    <tr id="details-{{.Group}}" class="details-row content" style="display: none;">
                        <td colspan="7" style="background-color: var(--bg-color); padding: 1.5rem;">
                            <div class="grid-2">
                                <div>
                                    <h4>Members & Assignments</h4>
                                    {{if .Members}}
                                    <table class="nested-table">
                                        <thead>
                                            <tr>
                                                <th>Client ID</th>
                                                <th>Host</th>
                                                <th>Assignments</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {{range .Members}}
                                            <tr>
                                                <td title="{{.MemberID}}">{{.ClientID}}</td>
                                                <td>{{.ClientHost}}</td>
                                                <td>
                                                    {{if .Assigned}}
                                                        {{range $topic, $partitions := .Assigned}}
                                                            <div style="font-size: 0.75rem; margin-bottom: 0.25rem;">
                                                                <strong>{{$topic}}</strong>: [{{range $i, $p := $partitions}}{{if $i}}, {{end}}{{$p}}{{end}}]
                                                            </div>
                                                        {{end}}
                                                    {{else}}
                                                        <span style="color: var(--text-color-muted); font-size: 0.75rem;">None</span>
                                                    {{end}}
                                                </td>
                                            </tr>
                                            {{end}}
                                        </tbody>
                                    </table>
                                    {{else}}
                                    <p>No active members.</p>
                                    {{end}}
                                </div>
                                <div>
                                    <h4>Committed Offsets & Lag</h4>
                                    {{if .Offsets}}
                                        {{range .Offsets}}
                                        <details style="margin-bottom: 0.5rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--card-bg);">
                                            <summary style="padding: 0.5rem; cursor: pointer; font-weight: 600; display: flex; justify-content: space-between;">
                                                <span>Topic: {{.Topic}} ({{len .Partitions}} partitions)</span>
                                                {{if gt .TopicLag 0}}
                                                    <span style="color: var(--log-level-WARN); padding-right: 1rem;">Lag: {{.TopicLag}}</span>
                                                {{end}}
                                            </summary>
                                            <div style="padding: 0.5rem;">
                                                <table class="nested-table" style="margin-bottom: 0;">
                                                    <thead>
                                                        <tr>
                                                            <th>Partition</th>
                                                            <th>Offset</th>
                                                            <th>Lag</th>
                                                            <th>Error</th>
                                                        </tr>
                                                    </thead>
                                                    <tbody>
                                                        {{range .Partitions}}
                                                        <tr>
                                                            <td>{{.Partition}}</td>
                                                            <td>{{.At}}</td>
                                                            <td>
                                                                {{if gt .Lag 0}}
                                                                    <span style="color: var(--log-level-WARN); font-weight: 600;">{{.Lag}}</span>
                                                                {{else}}
                                                                    0
                                                                {{end}}
                                                            </td>
                                                            <td>{{if .Err}}{{.Err}}{{else}}-{{end}}</td>
                                                        </tr>
                                                        {{end}}
                                                    </tbody>
                                                </table>
                                            </div>
                                        </details>
                                        {{end}}
                                    {{else}}
                                    <p>No committed offsets found.</p>
                                    {{end}}
                                </div>
                            </div>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        {{else}}
            <div class="card">
                <p>No consumer group data found in kafka.json.</p>
            </div>
        {{end}}
    </div>

    <script src="/static/common.js"></script>
    <style>
        .button-sm {
            padding: 0.25rem 0.5rem;
            font-size: 0.75rem;
        }
        .data-table {
            width: 100%;
            border-collapse: collapse;
        }
        .data-table th {
            text-align: left;
            padding: 0.75rem;
            background-color: var(--bg-color);
            border-bottom: 2px solid var(--border-color);
        }
        .data-table td {
            padding: 0.75rem;
            border-bottom: 1px solid var(--border-color);
        }
        .sort-icon {
            margin-left: 0.25rem;
            color: var(--text-color-muted);
            font-size: 0.8rem;
        }
    </style>
    <script>
        let sortDirections = [true, true, true, true, true, true]; // Track sort direction for each column

        function sortTable(colIndex) {
            const table = document.getElementById("groupsTable");
            const tbody = table.querySelector("tbody");
            const rows = Array.from(tbody.querySelectorAll(".group-row"));
            const direction = sortDirections[colIndex] ? 1 : -1;
            
            rows.sort((a, b) => {
                let aVal = a.cells[colIndex].innerText.toLowerCase().trim();
                let bVal = b.cells[colIndex].innerText.toLowerCase().trim();

                // Special handling for numeric columns (Members, Total Lag)
                if (colIndex === 4 || colIndex === 5) {
                    // Try to get numeric value from data-sort attribute if available
                    const aSort = a.cells[colIndex].getAttribute('data-sort');
                    const bSort = b.cells[colIndex].getAttribute('data-sort');
                    
                    aVal = parseFloat(aSort || aVal) || 0;
                    bVal = parseFloat(bSort || bVal) || 0;
                }

                if (aVal < bVal) return -direction;
                if (aVal > bVal) return direction;
                return 0;
            });

            // Re-append rows in sorted order, ensuring each details row stays after its group row
            rows.forEach(row => {
                tbody.appendChild(row);
                const groupID = row.cells[0].textContent.trim();
                const detailsRow = document.getElementById('details-' + groupID);
                if (detailsRow) {
                    tbody.appendChild(detailsRow);
                }
            });

            // Toggle direction for next click
            sortDirections[colIndex] = !sortDirections[colIndex];

            // Update UI Icons (Optional but nice)
            const headers = table.querySelectorAll("th");
            headers.forEach((th, idx) => {
                const icon = th.querySelector(".sort-icon");
                if (icon) {
                    if (idx === colIndex) {
                        icon.textContent = sortDirections[colIndex] ? "↑" : "↓";
                    } else {
                        icon.textContent = "↕";
                    }
                }
            });
        }

        function filterGroups() {
            const input = document.getElementById('groupSearch');
            const filter = input.value.toLowerCase();
            const rows = document.querySelectorAll('.group-row');
            let visibleCount = 0;

            rows.forEach(row => {
                const groupID = row.getAttribute('data-group');
                const topics = row.getAttribute('data-topics');
                const detailsRow = document.getElementById('details-' + row.cells[0].textContent.trim());

                if (groupID.includes(filter) || topics.includes(filter)) {
                    row.style.display = "";
                    visibleCount++;
                } else {
                    row.style.display = "none";
                    if (detailsRow) detailsRow.style.display = "none"; // Hide expanded details if group is filtered out
                }
            });

            const stats = document.getElementById('searchStats');
            stats.textContent = `Showing ${visibleCount} of ${rows.length}`;
        }

        // Initialize stats
        document.addEventListener('DOMContentLoaded', () => {
            const rows = document.querySelectorAll('.group-row');
            document.getElementById('searchStats').textContent = `Showing ${rows.length} of ${rows.length}`;
        });
    </script>
</body>
</html>



================================================
FILE: ui/html/home.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Bundle Details</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
	<script src="/static/tables.js"></script>
	<script src="/static/htmx.min.js"></script>
</head>
<body>
	<div id="loading-overlay">
		<div id="loading-box">
			<h2>Loading...</h2>
			<div id="progress-container">
				<div id="progress-bar">0%</div>
			</div>
		</div>
	</div>

    {{template "nav" (dict "Active" "Home" "Sessions" .Sessions "ActivePath" .ActivePath)}}

	<div class="container">
		<div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 2rem; flex-wrap: wrap; gap: 1rem;">
			<div>
				<h1>Bundle Overview</h1>
				<h2 style="margin-top: 0; font-size: 1.25rem; color: var(--text-color-muted);">Node: <span style="color: var(--text-color);">{{.NodeHostname}}</span></h2>
				<p style="color: var(--text-color-muted); font-size: 0.875rem;">Page generated at: {{.StartupTime}}</p>
			</div>
			
			<form action="/search" method="GET" style="display: flex; gap: 0.5rem; width: 100%; max-width: 400px; align-items: center;">
				<input type="text" name="q" placeholder="Global Search..." style="flex-grow: 1;">
				<button type="submit">Search</button>
			</form>
		</div>

		<div class="grid-2">
			<div class="card">
				<h3>Cluster Status</h3>
				<table style="margin-bottom: 0; border: none;">
					<tbody>
						<tr>
							<td style="border: none; padding-left: 0;">Total Brokers</td>
							<td style="border: none; text-align: right; font-weight: 600;">{{.TotalBrokers}}</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Version</td>
							<td style="border: none; text-align: right; font-family: var(--font-mono);">{{.Version}}</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Rack Awareness</td>
							<td style="border: none; text-align: right;">
								{{if .RackAwarenessEnabled}}
									<span class="badge" style="background-color: #dcfce7; color: #166534;">Enabled</span>
								{{else}}
									<span class="badge">Disabled</span>
								{{end}}
							</td>
						</tr>
					</tbody>
				</table>
			</div>

			<div class="card">
				<h3>Health Check</h3>
				<table style="margin-bottom: 0; border: none;">
					<tbody>
						<tr>
							<td style="border: none; padding-left: 0;">Cluster Health</td>
							<td style="border: none; text-align: right;">
								{{if .IsHealthy}}
									<span class="badge" style="background-color: #dcfce7; color: #166534;">Healthy</span>
								{{else}}
									<span class="badge" style="background-color: #fee2e2; color: #991b1b;">Unhealthy</span>
								{{end}}
							</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Under Replicated Partitions</td>
							<td style="border: none; text-align: right;">
								{{if gt .UnderReplicatedPartitions 0}}
									<span class="badge" style="background-color: #fee2e2; color: #991b1b;">{{.UnderReplicatedPartitions}}</span>
								{{else}}
									<span style="color: var(--text-color-muted);">0</span>
								{{end}}
							</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Leaderless Partitions</td>
							<td style="border: none; text-align: right;">
								{{if gt .LeaderlessPartitions 0}}
									<span class="badge" style="background-color: #fee2e2; color: #991b1b;">{{.LeaderlessPartitions}}</span>
								{{else}}
									<span style="color: var(--text-color-muted);">0</span>
								{{end}}
							</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Nodes in Maintenance</td>
							<td style="border: none; text-align: right;">
								{{if gt .NodesInMaintenanceMode 0}}
									<span class="badge" style="background-color: #fef08a; color: #854d0e;">{{.NodesInMaintenanceMode}}</span>
									<div style="font-size: 0.75rem; color: var(--text-color-muted); margin-top: 4px;">
										IDs: {{range $i, $id := .MaintenanceModeNodeIDs}}{{if $i}}, {{end}}{{$id}}{{end}}
									</div>
								{{else}}
									<span style="color: var(--text-color-muted);">0</span>
								{{end}}
							</td>
						</tr>
						<tr>
							<td style="border: none; padding-left: 0;">Nodes Down</td>
							<td style="border: none; text-align: right;">
								{{if gt .NodesDown 0}}
									<span class="badge" style="background-color: #fee2e2; color: #991b1b;">{{.NodesDown}}</span>
								{{else}}
									<span style="color: var(--text-color-muted);">0</span>
								{{end}}
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>

		{{if .RackAwarenessEnabled}}
		<div class="card">
			<h3>Rack Information</h3>
			<table>
				<thead>
					<tr>
						<th>Rack Name</th>
						<th>Broker Count</th>
						<th>Node IDs</th>
					</tr>
				</thead>
				<tbody>
					{{range $rack, $data := .RackInfo}}
					<tr>
						<td>{{$rack}}</td>
						<td>{{$data.Count}}</td>
						<td>{{range $i, $id := $data.NodeIDs}}{{if $i}}, {{end}}{{$id}}{{end}}</td>
					</tr>
					{{end}}
				</tbody>
			</table>
		</div>
		{{end}}

		{{range $groupName, $files := .GroupedFiles}}
		<h2 style="border-bottom: 1px solid var(--border-color); padding-bottom: 0.5rem; margin-bottom: 1.5rem; margin-top: 2rem;">{{$groupName}}</h2>
		
		<div class="file-list">
			{{range $files}}
			<details class="file-list-item">
				<summary class="file-list-summary">
					<span class="file-name">📄 {{.FileName}}</span>
					<span style="font-size: 0.75rem; color: var(--text-color-muted);">Click to expand</span>
				</summary>
				<div class="file-list-content">
					{{if .Error}}
						<p style="color: #ef4444;">Error: {{.Error}}</p>
					{{else if eq .FileName "kafka.json"}}
						<p><a href="/kafka">View Kafka Metadata &rarr;</a></p>
					{{else}}
						{{renderData .FileName .Data}}
					{{end}}
				</div>
			</details>
			{{end}}
		</div>
		{{end}}
	</div>

	<a href="#top" id="back-to-top" title="Back to top">&uarr;</a>

	<script src="/static/common.js"></script>
	<script src="/static/progress.js"></script>
</body>
</html>



================================================
FILE: ui/html/k8s.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Kubernetes Overview</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
	<script src="/static/tables.js"></script>
	<script src="/static/htmx.min.js"></script>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "K8s" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
	<h1>Kubernetes Resources</h1>
    <p>Node: {{.NodeHostname}}</p>

    {{if not (or .Store.Pods .Store.Services .Store.ConfigMaps .Store.Endpoints .Store.Events .Store.LimitRanges .Store.PersistentVolumeClaims .Store.ReplicationControllers .Store.ResourceQuotas .Store.ServiceAccounts)}}
        <p>No Kubernetes resources found.</p>
    {{end}}

	{{if .Store.Pods}}
	<h2>Pods</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Node</th><th>Phase</th><th>Restarts</th><th>Containers & Resources</th><th>Labels</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.Pods}}
		<tr id="pod-{{.Metadata.Name}}">
			<td>
                {{.Metadata.Name}}
                {{range .Metadata.OwnerReferences}}
                    <div style="font-size: 0.8em; color: var(--text-color-light);">Owner: {{.Kind}}/{{.Name}}</div>
                {{end}}
            </td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Spec.NodeName}}</td>
			<td>
                {{.Status.Phase}}
                {{range .Status.Conditions}}
					{{if ne .Status "True"}}
						<div class="condition-failed" title="{{.Type}}">{{.Type}}</div>
					{{end}}
				{{end}}
            </td>
			<td>{{sumRestarts .Status.ContainerStatuses}}</td>
			<td>
				{{range .Spec.Containers}}
					<div style="margin-bottom: 8px; border-bottom: 1px solid var(--border-color); padding-bottom: 4px;">
						<strong>{{.Name}}</strong>
                        <div style="font-size: 0.85em; margin-bottom: 4px;">
                            {{if .Resources.Limits}}
                                <span title="Limits">L: {{range $k, $v := .Resources.Limits}}{{$k}}={{$v}} {{end}}</span>
                            {{end}}
                            {{if .Resources.Requests}}
                                <br><span title="Requests">R: {{range $k, $v := .Resources.Requests}}{{$k}}={{$v}} {{end}}</span>
                            {{end}}
                        </div>
                        <div style="font-size: 0.8em; color: var(--text-color-light);">
                            {{if .LivenessProbe}}
                                <div><strong>Liveness:</strong> {{template "probe" .LivenessProbe}}</div>
                            {{end}}
                            {{if .ReadinessProbe}}
                                <div><strong>Readiness:</strong> {{template "probe" .ReadinessProbe}}</div>
                            {{end}}
                            {{if .StartupProbe}}
                                <div><strong>Startup:</strong> {{template "probe" .StartupProbe}}</div>
                            {{end}}
                        </div>
					</div>
				{{end}}
			</td>
			<td>
				{{range $k, $v := .Metadata.Labels}}
					<div class="badge" title="{{$k}}: {{$v}}">{{$k}}</div>
				{{end}}
			</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{define "probe"}}
        {{if .HTTPGet}}HTTP:{{.HTTPGet.Path}}:{{.HTTPGet.Port}}{{end}}
        {{if .Exec}}Exec:{{range .Exec.Command}}{{.}} {{end}}{{end}}
        {{if .TCPSocket}}TCP:{{.TCPSocket.Port}}{{end}}
        (delay={{.InitialDelaySeconds}}s, timeout={{.TimeoutSeconds}}s, period={{.PeriodSeconds}}s, fail={{.FailureThreshold}})
    {{end}}

	{{if .Store.Events}}
	<h2>Events</h2>
	<table class="sortable-table">
		<thead><tr><th>Type</th><th>Reason</th><th>Object</th><th>Message</th><th>Count</th><th>Last Seen</th></tr></thead>
		<tbody>
		{{range .Store.Events}}
		<tr class="{{if eq .Type "Warning"}}event-warning{{else}}event-normal{{end}}">
			<td>{{.Type}}</td>
			<td>{{.Reason}}</td>
			<td>{{.InvolvedObject.Kind}}/{{.InvolvedObject.Name}}</td>
			<td>{{.Message}}</td>
			<td>{{.Count}}</td>
			<td>{{.LastTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

	{{if .Store.Services}}
	<h2>Services</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Type</th><th>ClusterIP</th><th>Ports</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.Services}}
		<tr id="service-{{.Metadata.Name}}">
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
            <td>{{.Spec.Type}}</td>
            <td>{{.Spec.ClusterIP}}</td>
            <td>
                {{range .Spec.Ports}}
                    <div>{{.Port}}:{{.TargetPort}} ({{.Protocol}})</div>
                {{end}}
            </td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{if .Store.PersistentVolumeClaims}}
	<h2>Persistent Volume Claims</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Status</th><th>StorageClass</th><th>Capacity</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.PersistentVolumeClaims}}
		<tr id="pvc-{{.Metadata.Name}}">
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
            <td>{{.Status.Phase}}</td>
            <td>{{.Spec.StorageClassName}}</td>
            <td>{{range $k, $v := .Status.Capacity}}{{$v}}{{end}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{if .Store.LimitRanges}}
	<h2>Limit Ranges</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.LimitRanges}}
		<tr>
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{if .Store.ResourceQuotas}}
	<h2>Resource Quotas</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.ResourceQuotas}}
		<tr>
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{if .Store.ReplicationControllers}}
	<h2>Replication Controllers</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.ReplicationControllers}}
		<tr>
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

    {{if .Store.ServiceAccounts}}
	<h2>Service Accounts</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.ServiceAccounts}}
		<tr>
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

	{{if .Store.Nodes}}
	<h2>Nodes</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Conditions</th><th>Capacity</th><th>Allocatable</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.Nodes}}
		<tr id="node-{{.Metadata.Name}}">
			<td>{{.Metadata.Name}}</td>
			<td>
				{{range .Status.Conditions}}
					<div class="{{if eq .Status "True"}}{{if eq .Type "Ready"}}state-running{{else}}condition-failed{{end}}{{else}}{{if eq .Type "Ready"}}condition-failed{{else}}state-running{{end}}{{end}}">
						{{.Type}}: {{.Status}}
					</div>
				{{end}}
			</td>
			<td>{{range $k, $v := .Status.Capacity}}<div>{{$k}}: {{$v}}</div>{{end}}</td>
			<td>{{range $k, $v := .Status.Allocatable}}<div>{{$k}}: {{$v}}</div>{{end}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

	{{if .Store.StatefulSets}}
	<h2>StatefulSets</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Replicas (Ready/Desired)</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.StatefulSets}}
		<tr id="sts-{{.Metadata.Name}}">
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Status.ReadyReplicas}} / {{.Spec.Replicas}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

	{{if .Store.Deployments}}
	<h2>Deployments</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Replicas (Ready/Desired)</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.Deployments}}
		<tr id="deploy-{{.Metadata.Name}}">
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>{{.Status.ReadyReplicas}} / {{.Spec.Replicas}}</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}

	{{if .Store.ConfigMaps}}
	<h2>ConfigMaps</h2>
	<table class="sortable-table">
		<thead><tr><th>Name</th><th>Namespace</th><th>Data</th><th>Creation Time</th></tr></thead>
		<tbody>
		{{range .Store.ConfigMaps}}
		<tr>
			<td>{{.Metadata.Name}}</td>
			<td>{{.Metadata.Namespace}}</td>
			<td>
				{{range $k, $v := .Data}}
					<div style="margin-bottom: 10px;">
						<strong>{{$k}}</strong>:
						<pre style="font-size: 0.8em; background: var(--bg-color); padding: 5px; max-height: 200px; overflow: auto; border: 1px solid var(--border-color); white-space: pre-wrap; word-break: break-all;">{{$v}}</pre>
					</div>
				{{end}}
			</td>
			<td>{{.Metadata.CreationTimestamp}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>
	{{end}}
</body>
</html>



================================================
FILE: ui/html/kafka.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Kafka Metadata</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
</head>
<body>
    		
    		</div>
    	    {{template "nav" (dict "Active" "Kafka" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    		    <h1>Kafka Metadata</h1>    <button onclick="toggleRawJson()">View Raw JSON</button>
    		    <pre id="raw-json" style="display: none;"><code>{{if .RawJSON}}{{.RawJSON}}{{else}}Raw JSON data not available.{{end}}</code></pre>
    <p><strong>Cluster:</strong> {{.Metadata.Cluster}}</p>
    <p><strong>Controller:</strong> {{.Metadata.Controller}}</p>
    <p><strong>Brokers:</strong> {{len .Metadata.Brokers}}</p>
    <table>
        <thead>
            <tr>
                <th>Node ID</th>
                <th>Host</th>
                <th>Port</th>
                <th>Rack</th>
            </tr>
        </thead>
        <tbody>
            {{range .Metadata.Brokers}}
            <tr>
                <td>{{.NodeID}}</td>
                <td>{{.Host}}</td>
                <td>{{.Port}}</td>
                <td>{{if .Rack}}{{.Rack}}{{else}}N/A{{end}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <h2>Topics</h2>
    {{range .Metadata.Topics}}
    <div class="topic-section">
        <h3>{{.Topic}}</h3>
        <p><strong>ID:</strong> {{.ID}}</p>
        <p><strong>Internal:</strong> {{.IsInternal}}</p>
        <p><strong>Authorized Operations:</strong> {{.AuthorizedOperations}}</p>
        <p><strong>Error:</strong> {{.Err}}</p>

        <button onclick="toggleVisibility('partitions-{{.Topic}}', this)">Partitions</button>
        <div id="partitions-{{.Topic}}" class="content" style="display: none;">
            <table class="nested-table">
                <thead>
                    <tr>
                        <th>Partition</th>
                        <th>Leader</th>
                        <th>Leader Epoch</th>
                        <th>Replicas</th>
                        <th>ISR</th>
                        <th>Offline Replicas</th>
                        <th>Error</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Partitions}}
                    <tr>
                        <td>{{.Partition}}</td>
                        <td>{{.Leader}}</td>
                        <td>{{.LeaderEpoch}}</td>
                        <td>{{.Replicas}}</td>
                        <td>{{.ISR}}</td>
                        <td>{{.OfflineReplicas}}</td>
                        <td>{{.Err}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        {{$topicConfig := index $.TopicConfigs .Topic}}
        {{if $topicConfig}}
            <button onclick="toggleVisibility('config-{{.Topic}}', this)">Config</button>
            <div id="config-{{.Topic}}" class="content" style="display: none;">
                            <table class="nested-table">
                                <thead>
                                    <tr>
                                        <th>Key</th>
                                        <th>Value</th>
                                        <th>Docs</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{range $topicConfig.Configs}}
                                    <tr>
                                        <td>
                                            {{.Key}}
                                            {{if .Synonyms}}
                                                ({{range $i, $s := .Synonyms}}{{if $i}}, {{end}}{{$s.Key}}{{end}})
                                            {{end}}
                                        </td>
                                        <td>{{.Value}}</td>
                                                                <td>
                                                                    {{$keyForUrl := .Key}}
                                                                    {{$isTopic := (contains .Key ".")}}
                                                                    {{if .Synonyms}}
                                                                        {{$synonymKey := (index .Synonyms 0).Key}}
                                                                        {{$keyForUrl = $synonymKey}}
                                                                        {{if (contains $synonymKey ".")}}
                                                                            {{$isTopic = true}}
                                                                        {{else if (contains $synonymKey "_")}}
                                                                            {{$isTopic = false}}
                                                                        {{end}}
                                                                    {{end}}
                                                                    {{if $isTopic}}
                                                                        <a href="https://docs.redpanda.com/current/reference/properties/topic-properties/#{{replace $keyForUrl "." ""}}" target="_blank">doc</a>
                                                                    {{else}}
                                                                        <a href="https://docs.redpanda.com/current/reference/properties/cluster-properties/#{{$keyForUrl}}" target="_blank">doc</a>
                                                                    {{end}}
                                                                </td>                                    </tr>
                                    {{end}}
                                </tbody>
                            </table>            </div>
        {{end}}
    </div>
    {{end}}

    <a href="#top" id="back-to-top" title="Back to top">&uarr;</a>

    <script src="/static/common.js"></script>
    <script>
        function toggleRawJson() {
            var el = document.getElementById("raw-json");
            if (el.style.display === "none") {
                el.style.display = "block";
            } else {
                el.style.display = "none";
            }
        }
    </script>
</body>
</html>



================================================
FILE: ui/html/log_analysis.tmpl
================================================
{{if not .Partial}}
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Log Analysis</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/tables.js"></script>
    <script src="/static/htmx.min.js"></script>
    <style>
        .pattern-row {
            cursor: pointer;
        }
        .pattern-row:hover {
            background-color: var(--row-hover-color);
        }
        .signature-cell {
            font-family: monospace;
            font-size: 0.9em;
            word-break: break-all;
        }
        .level-badge {
            padding: 2px 6px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 0.8em;
        }
        .level-ERROR { background-color: #ffcccc; color: #cc0000; }
        .level-WARN { background-color: #fff4cc; color: #996600; }
        .level-INFO { background-color: #ccffcc; color: #006600; }
        .level-DEBUG { background-color: #eeeeee; color: #666666; }
        
        [data-theme="dark"] .level-ERROR, [data-theme="ultra-dark"] .level-ERROR { background-color: #5c0000; color: #ff9999; }
        [data-theme="dark"] .level-WARN, [data-theme="ultra-dark"] .level-WARN { background-color: #5c4200; color: #ffcc00; }
        [data-theme="dark"] .level-INFO, [data-theme="ultra-dark"] .level-INFO { background-color: #003300; color: #66ff66; }
        [data-theme="dark"] .level-DEBUG, [data-theme="ultra-dark"] .level-DEBUG { background-color: #333333; color: #aaaaaa; }
    </style>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "Logs" "LogsOnly" .LogsOnly "Sessions" .Sessions "ActivePath" .ActivePath)}}
    
    <h1>Log Analysis</h1>
    <div id="analysis-content" hx-get="/api/logs/analysis" hx-trigger="load" hx-swap="outerHTML">
        <div class="container" style="height: 60vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%;">
                <h2 style="margin-bottom: 1rem;">Analyzing Logs...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Fingerprinting log messages and identifying recurring patterns.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
    <script src="/static/common.js"></script>
    <script src="/static/log_analysis_view.js"></script>
</body>
</html>
{{else}}
    <p><a href="/logs">&larr; Back to Raw Logs</a></p>
    <p>Grouping <strong>{{.TotalLogs}}</strong> log entries into <strong>{{.TotalPatterns}}</strong> unique patterns.</p>

    <table class="sortable-table">
        <thead>
            <tr>
                <th data-type="number">Count</th>
                <th>Level</th>
                <th>Signature (Fingerprint)</th>
                <th>First Seen</th>
                <th>Last Seen</th>
            </tr>
        </thead>
        <tbody>
            {{range .Patterns}}
            <tr class="pattern-row">
                <td>{{.Count}}</td>
                <td><span class="level-badge level-{{.Level}}">{{.Level}}</span></td>
                <td class="signature-cell" title="{{.SampleEntry.Message}}">{{.Signature}}</td>
                <td>{{.FirstSeen.Format "15:04:05"}}</td>
                <td>{{.LastSeen.Format "15:04:05"}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
{{end}}



================================================
FILE: ui/html/logs.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Cluster Logs</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
	<script src="/static/tables.js"></script>
	<script src="/static/htmx.min.js"></script>
	<script src="/static/spinner.js"></script>
	<style>
		.log-filters {
			display: flex;
			gap: 15px;
			margin-bottom: 20px;
			align-items: center;
			flex-wrap: wrap;
			background-color: var(--card-bg);
			padding: 15px;
			border-radius: 8px;
			box-shadow: 0 2px 5px rgba(0,0,0,0.2);
			position: sticky;
			top: 10px;
			z-index: 100;
		}
		
		.log-filters input, .log-filters select {
			padding: 8px;
			border: 1px solid #ddd;
			border-radius: 4px;
			background-color: var(--bg-color);
			color: var(--text-color);
		}
		
		.log-table {
			font-family: 'Menlo', 'Monaco', 'Consolas', 'Courier New', monospace;
			font-size: 16px; /* Increased font size */
			width: 100%;
			border-collapse: collapse;
		}
		
		.log-table th {
			background-color: var(--th-bg);
			color: var(--text-color);
			text-align: left;
			padding: 12px;
			border-bottom: 2px solid #ddd;
			position: sticky;
			top: 195px; /* Offset to appear below sticky filters. Adjusted for approximate height. */
			z-index: 90;
		}
		
		.log-table td {
			padding: 8px 12px;
			border-bottom: 1px solid #eee;
			vertical-align: top;
			line-height: 1.5;
		}
		
		mark {
			background-color: #ffd700;
			color: black;
			border-radius: 2px;
			padding: 0 2px;
		}

		.log-level-ERROR { color: #d32f2f; font-weight: bold; }
		.log-level-WARN { color: #fbc02d; font-weight: bold; }
		.log-level-INFO { color: #388e3c; }
		.log-level-DEBUG { color: #1976d2; }
		.log-level-TRACE { color: #757575; }
		
		.pagination {
			display: flex;
			justify-content: center;
			gap: 10px;
			margin-top: 20px;
			margin-bottom: 50px;
		}
		
		.pagination button {
			padding: 8px 16px;
			cursor: pointer;
			background-color: var(--link-color);
			color: white;
			border: none;
			border-radius: 4px;
		}
		
		.pagination button:disabled {
			background-color: #ccc;
			cursor: not-allowed;
		}

		/* Dark mode adjustments */
		[data-theme="dark"] .log-filters {
			background-color: var(--card-bg);
			border: 1px solid #444;
		}
		[data-theme="dark"] .log-table td {
			border-bottom: 1px solid #444;
		}
		[data-theme="dark"] .log-level-ERROR { color: #ff6b6b; }
		[data-theme="dark"] .log-level-WARN { color: #ffd93d; }
		[data-theme="dark"] .log-level-INFO { color: #69db7c; }
		[data-theme="dark"] .log-level-DEBUG { color: #64b5f6; }
		[data-theme="dark"] .log-level-TRACE { color: #9e9e9e; }
		[data-theme="dark"] mark { background-color: #5a5200; color: #e0e0e0; }
	</style>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "Logs" "LogsOnly" .LogsOnly "Sessions" .Sessions "ActivePath" .ActivePath)}}

	<div class="log-filters">
		<!-- Section 1: Time Range & Facet Filters -->
		<div class="filter-row primary-controls">
			<div class="time-range-group">
				<div class="input-group-vertical">
					<label>From <span class="time-helper">(Earliest: {{.MinTime}})</span></label>
					<input type="datetime-local" id="start-time" value="{{.Filter.StartTime}}" min="{{.MinTime}}" max="{{.MaxTime}}" step="1">
				</div>
				<div class="input-group-vertical">
					<label>To <span class="time-helper">(Latest: {{.MaxTime}})</span></label>
					<input type="datetime-local" id="end-time" value="{{.Filter.EndTime}}" min="{{.MinTime}}" max="{{.MaxTime}}" step="1">
				</div>
			</div>

			<div class="facet-filters">
				<select id="level-select" multiple>
					<option value="ERROR" {{if containsString .Filter.Level "ERROR"}}selected{{end}}>ERROR</option>
					<option value="WARN" {{if containsString .Filter.Level "WARN"}}selected{{end}}>WARN</option>
					<option value="INFO" {{if containsString .Filter.Level "INFO"}}selected{{end}}>INFO</option>
					<option value="DEBUG" {{if containsString .Filter.Level "DEBUG"}}selected{{end}}>DEBUG</option>
					<option value="TRACE" {{if containsString .Filter.Level "TRACE"}}selected{{end}}>TRACE</option>
				</select>
				
				<select id="node-select" multiple>
					<option value="">All Nodes</option>
					{{range .Nodes}}
						<option value="{{.}}" {{if containsString $.Filter.Node .}}selected{{end}}>{{.}}</option>
					{{end}}
				</select>
				
				<select id="component-select" multiple>
					<option value="">All Components</option>
					{{range .Components}}
						<option value="{{.}}" {{if containsString $.Filter.Component .}}selected{{end}}>{{.}}</option>
					{{end}}
				</select>

				<select id="sort-select" style="min-width: 140px;">
					<option value="desc" {{if eq .Filter.Sort "desc"}}selected{{end}}>Newest First</option>
					<option value="asc" {{if eq .Filter.Sort "asc"}}selected{{end}}>Oldest First</option>
				</select>
			</div>
		</div>

		<!-- Section 2: Search & Actions -->
		<div class="filter-row secondary-controls">
			<div class="search-container">
				<input type="text" id="query-input" placeholder="Search logs... (e.g. 'error NOT timeout' or 'pod-name')" value="{{.Filter.Search}}">
				<div class="help-container">
					<span class="help-icon">?</span>
					<div class="help-tooltip">
						<strong>Search Syntax:</strong><br>
						• <code>word</code> : matches exact word<br>
						• <code>"phrase"</code> : matches exact phrase<br>
						• <code>AND</code>, <code>OR</code>, <code>NOT</code> : boolean operators<br>
						• <code>( ... )</code> : grouping<br>
						<br>
						<strong>Examples:</strong><br>
						• <code>NOT timeout</code> (exclude matches)<br>
						• <code>error NOT timeout</code><br>
						• <code>(failed OR crash) AND "system disk"</code>
					</div>
				</div>
			</div>

			<div class="action-buttons">
				<button onclick="applyFilters()" class="primary-btn">Filter</button>
				<button onclick="clearFilters()" class="secondary-btn">Clear</button>
				<div class="analyze-wrapper">
					<button id="analyze-logs-button" type="button" class="analyze-btn">Analyze</button>
					<span id="analyze-loading-indicator">Analyzing...</span>
				</div>
			</div>
		</div>
	</div>

    <style>
        /* Enhanced Filter Layout Styles */
        .log-filters {
            display: flex;
            flex-direction: column;
            gap: 20px;
            padding: 20px;
            background-color: var(--card-bg);
            border-radius: 8px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            border: 1px solid var(--border-color);
        }

        .filter-row {
            display: flex;
            gap: 20px;
            align-items: flex-end; /* Align to bottom for better visual flow with labels */
        }

        .primary-controls {
            flex-wrap: wrap;
        }

        .search-container {
            flex-grow: 1;
            position: relative;
            min-width: 300px;
        }

        #query-input {
            width: 100%;
            padding: 10px 35px 10px 12px;
            font-size: 15px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
            background-color: var(--bg-color);
            transition: border-color 0.2s, box-shadow 0.2s;
        }
        
        #query-input:focus {
            border-color: var(--primary-color);
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
            outline: none;
        }

        .help-container {
            position: absolute;
            right: 12px;
            top: 0;
            bottom: 0;
            margin: auto;
            height: 20px;
            display: flex;
            align-items: center;
            color: var(--text-color-muted);
            cursor: help;
            z-index: 5;
        }

        .help-icon {
            font-weight: 700;
            border: 1px solid var(--text-color-muted);
            border-radius: 50%;
            width: 18px;
            height: 18px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            background-color: var(--bg-color);
        }

        .time-range-group {
            display: flex;
            gap: 15px;
            padding-right: 20px;
            border-right: 1px solid var(--border-color);
        }

        .input-group-vertical {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }

        .input-group-vertical label {
            font-size: 12px;
            font-weight: 600;
            color: var(--text-color);
        }

        .time-helper {
            font-weight: 400;
            color: var(--text-color-muted);
            font-size: 11px;
            margin-left: 4px;
        }

        .input-group-vertical input[type="datetime-local"] {
            border: 1px solid var(--border-color);
            background: var(--bg-color);
            font-size: 13px;
            padding: 6px 8px;
            color: var(--text-color);
            border-radius: 4px;
            width: 210px; /* Fixed width for better alignment */
        }

        .secondary-controls {
            align-items: center; /* Center vertically for search bar row */
            padding-top: 15px;
            border-top: 1px solid var(--border-color);
        }

        .facet-filters {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
            align-items: flex-end; /* Align bottoms */
            flex-grow: 1;
        }

        .action-buttons {
            display: flex;
            gap: 10px;
            align-items: center;
            flex-shrink: 0;
        }

        .primary-btn {
            background-color: var(--primary-color);
            color: white;
            border: none;
            padding: 8px 20px;
            border-radius: 6px;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s;
            height: 38px; /* Match input height roughly */
        }
        
        .primary-btn:hover { background-color: var(--primary-hover); }

        .secondary-btn {
            background-color: transparent;
            color: var(--text-color);
            border: 1px solid var(--border-color);
            padding: 8px 20px;
            border-radius: 6px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
            height: 38px;
        }

        .secondary-btn:hover {
            background-color: var(--bg-color);
            border-color: var(--text-color-muted);
        }

        .analyze-btn {
            background-color: #7c3aed; /* Violet 600 */
            color: white;
            border: none;
            padding: 8px 20px;
            border-radius: 6px;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s;
            height: 38px;
        }

        .analyze-btn:hover { background-color: #6d28d9; } /* Violet 700 */

        .analyze-wrapper {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        #analyze-loading-indicator {
            display: none;
            font-size: 13px;
            color: var(--text-color-muted);
        }

        /* Responsive adjustments */
        @media (max-width: 1200px) {
            .time-range-group { border-right: none; padding-right: 0; width: 100%; justify-content: space-between; }
            .input-group-vertical input[type="datetime-local"] { width: 100%; }
            .filter-row.primary-controls { flex-direction: column; align-items: stretch; }
        }

        @media (max-width: 768px) {
            .filter-row.secondary-controls { flex-direction: column; align-items: stretch; }
            .action-buttons { justify-content: flex-end; }
        }
    </style>

	<div style="margin-top: 20px; margin-bottom: 10px; text-align: right; color: var(--text-color-muted); font-size: 0.875rem; font-weight: 500;">
		<span id="log-count-display">
			<!-- Log count will be updated by JavaScript -->
		</span>
	</div>

	<table class="log-table">
		<thead>
			<tr>
				<th style="width: 150px;">Timestamp</th>
				<th style="width: 60px;">Level</th>
				<th style="width: 100px;">Node</th>
				<th style="width: 80px;">Shard</th>
				<th style="width: 120px;">Component</th>
				<th>Message</th>
				<th style="width: 80px;">Actions</th>
			</tr>
		</thead>
		<tbody>
			<!-- Logs will be dynamically loaded here by JavaScript -->
		</tbody>
	</table>

	<div id="loading-indicator" style="text-align: center; padding: 20px;">
		<!-- Loading text will be dynamically added by JavaScript -->
	</div>

	<style>
		.help-tooltip {
			visibility: hidden;
			width: 300px;
			background-color: var(--card-bg);
			color: var(--text-color);
			text-align: left;
			border-radius: 6px;
			padding: 10px;
			position: absolute;
			z-index: 100; /* Increased z-index */
			top: 130%;
			right: 0;
			box-shadow: 0 4px 12px rgba(0,0,0,0.15);
			border: 1px solid var(--border-color);
			font-size: 0.9em;
			opacity: 0;
			transition: opacity 0.2s;
		}
		.help-container:hover .help-tooltip {
			visibility: visible;
			opacity: 1;
		}
		.help-tooltip code {
			background-color: var(--bg-color);
			padding: 2px 4px;
			border-radius: 3px;
			font-family: monospace;
		}
	</style>

	<script src="/static/common.js"></script>
	<script src="/static/logs_view.js"></script>
</body>
</html>
<style>
	/* Modal styles */
	.modal {
		display: none; /* Hidden by default */
		position: fixed; /* Stay in place */
		z-index: 1001; /* Sit on top of everything (filters are 100) */
		left: 0;
		top: 0;
		width: 100%;
		height: 100%;
		overflow: auto; /* Enable scroll if needed */
		background-color: rgba(0,0,0,0.6); /* Black w/ opacity */
	}

	.modal-content {
		background-color: var(--card-bg);
		margin: 5% auto;
		padding: 20px;
		border: 1px solid #888;
		width: 80%;
		max-width: 1200px;
		border-radius: 8px;
		box-shadow: 0 4px 8px rgba(0,0,0,0.2);
		color: var(--text-color);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-bottom: 10px;
		border-bottom: 1px solid #ddd;
	}

	.modal-header h2 {
		margin: 0;
	}

	.close-button {
		color: #aaa;
		font-size: 28px;
		font-weight: bold;
		cursor: pointer;
	}

	.close-button:hover,
	.close-button:focus {
		color: var(--text-color);
	}

	.modal-body {
		padding-top: 15px;
		font-family: 'Menlo', 'Monaco', 'Consolas', 'Courier New', monospace;
		font-size: 14px;
		line-height: 1.6;
		white-space: pre-wrap; /* Wrap long lines */
		background-color: var(--bg-color);
		border-radius: 4px;
		max-height: 70vh;
		overflow-y: auto;
	}

	.log-context-line {
		border-bottom: 2px solid var(--border-color);
		padding: 8px 10px;
	}

	.log-context-line.even-line {
		background-color: var(--bg-color);
	}

	.log-context-line.odd-line {
		background-color: var(--table-alt-row-bg);
	}
	
	:root {
		--highlight-bg: #ffd700; /* Gold for light mode */
	}

	[data-theme="dark"] {
		--highlight-bg: #5a5200; /* Darker, less vibrant yellow for dark mode */
	}

	[data-theme="ultra-dark"] {
		--highlight-bg: #403e00; /* Even darker for ultra-dark, good contrast */
	}
	
	[data-theme="retro"] {
		--highlight-bg: #4a148c; /* A deep violet for retro highlight, providing good contrast with neon text */
		--card-bg: #2d1b4e;
	}
</style>

<div id="log-context-modal" class="modal">
	<div class="modal-content">
		<div class="modal-header">
			<h2>Log Context</h2>
			<span class="close-button" onclick="closeLogContext()">&times;</span>
		</div>
		<div id="log-context-body" class="modal-body">
			<!-- Context will be loaded here -->
		</div>
	</div>
</div>



================================================
FILE: ui/html/metrics.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Metrics Overview</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    <script src="/static/chart.js"></script>
    <script src="/static/metrics_view.js"></script>
    <style>
        .metric-card {
            background-color: var(--header-bg-color);
            border: 1px solid var(--border-color);
            border-radius: var(--radius);
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            box-shadow: var(--shadow-sm);
        }
        .metric-value {
            font-size: 2rem;
            font-weight: 700;
            color: var(--primary-color);
            line-height: 1;
            margin-bottom: 0.25rem;
        }
        .metric-label {
            color: var(--text-color-muted);
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
    </style>
    {{template "nav" (dict "Active" "Metrics" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    <h1>Metrics Overview</h1>
    <h2>Node: {{.NodeHostname}}</h2>

    <div class="metric-card">
        <h3>Metrics Explorer</h3>
        <div style="margin-bottom: 15px; display: flex; gap: 10px;">
            <select id="metric-selector" style="flex-grow: 1; padding: 8px; font-size: 14px;">
                <option value="">Select a metric to visualize...</option>
                {{range .MetricNames}}
                <option value="{{.}}">{{.}}</option>
                {{end}}
            </select>
            <div style="display: flex; align-items: center; gap: 5px;">
                <input type="checkbox" id="rate-toggle" style="width: auto;">
                <label for="rate-toggle" style="font-size: 0.9em; white-space: nowrap;">As Rate (/s)</label>
            </div>
            <button onclick="loadMetricData()" style="padding: 8px 16px; background-color: var(--link-color); color: var(--bg-color); font-weight: bold;">Plot</button>
        </div>
        <div style="height: 400px; position: relative;">
            <canvas id="explorerChart"></canvas>
            <div id="explorer-loading" style="display: none; position: absolute; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.1); display: flex; justify-content: center; align-items: center; z-index: 10;">
                Loading...
            </div>
        </div>
    </div>

    <div class="grid-2">
        <div class="card">
            <h3 style="margin-bottom: 1.5rem; display: flex; align-items: center; gap: 0.5rem;">
                <span style="font-size: 1.25rem;">📊</span> System Resources
            </h3>
            <div class="grid-4">
                <div>
                    <div class="metric-label">Normalized CPU</div>
                    <div class="metric-value">
                        {{if gt .CoreCount 0}}
                            {{printf "%.2f" (div .ResourceUsage.CPUPercentage (float64 .CoreCount))}}%
                            <div style="font-size: 0.75rem; color: var(--text-color-muted); font-weight: normal; margin-top: -4px;">
                                {{.CoreCount}} Cores
                            </div>
                        {{else}}
                            {{printf "%.2f" .ResourceUsage.CPUPercentage}}%
                        {{end}}
                    </div>
                </div>
                <div>
                    <div class="metric-label">Free Memory</div>
                    <div class="metric-value" style="font-size: 1.5rem;">{{formatBytesFloat (mul .ResourceUsage.FreeMemoryMB 1048576.0)}}</div>
                </div>
                <div>
                    <div class="metric-label">Free Disk</div>
                    <div class="metric-value" style="font-size: 1.5rem;">{{formatBytesFloat (mul .ResourceUsage.FreeSpaceMB 1048576.0)}}</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h3 style="margin-bottom: 1.5rem; display: flex; align-items: center; gap: 0.5rem;">
                <span style="font-size: 1.25rem;">🌐</span> Network & RPC
            </h3>
            <div class="grid-4">
                <div>
                    <div class="metric-label">Active RPC</div>
                    <div class="metric-value">{{printf "%.0f" .Network.ActiveConnections}}</div>
                </div>
                <div>
                    <div class="metric-label">RPC Errors</div>
                    <div class="metric-value" style="{{if gt .Network.RPCErrors 0.0}}color: #ef4444;{{end}}">
                        {{printf "%.0f" .Network.RPCErrors}}
                    </div>
                </div>
            </div>
            
            {{if .Network.KafkaLatencies}}
            <div style="margin-top: 1.5rem;">
                <div class="metric-label" style="margin-bottom: 0.5rem;">Kafka Latency (Lifetime Avg)</div>
                <table style="font-size: 0.85rem; margin-bottom: 0; border: none;">
                    <thead>
                        <tr>
                            <th style="border: none; background: transparent; padding: 4px 8px;">Type</th>
                            <th style="border: none; background: transparent; padding: 4px 8px;">Latency</th>
                            <th style="border: none; background: transparent; padding: 4px 8px;">Count</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Network.KafkaLatencies}}
                        <tr>
                            <td style="padding: 4px 8px; border: none;">{{.RequestType}}</td>
                            <td style="padding: 4px 8px; border: none; font-family: var(--font-mono);">{{printf "%.2f" (mul .AvgLatency 1000.0)}}ms</td>
                            <td style="padding: 4px 8px; border: none;">{{printf "%.0f" .Count}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{end}}
        </div>
    </div>

    {{if .SarData.Entries}}
    <div class="metric-card">
        <h3>Historical CPU Usage (SAR)</h3>
        <div style="height: 300px;">
            <canvas id="sarCpuChart"></canvas>
        </div>
        <script>
            (function() {
                const labels = [
                    {{range .SarData.Entries}}"{{.Timestamp.Format "15:04:05"}}",{{end}}
                ];
                const userCpu = [{{range .SarData.Entries}}{{.CPUUser}},{{end}}];
                const sysCpu = [{{range .SarData.Entries}}{{.CPUSystem}},{{end}}];
                const idleCpu = [{{range .SarData.Entries}}{{.CPUIdle}},{{end}}];

                function tryInit() {
                    if (window.initSarChart) {
                        window.initSarChart(labels, userCpu, sysCpu, idleCpu);
                    } else {
                        setTimeout(tryInit, 50);
                    }
                }
                tryInit();
            })();
        </script>
    </div>
    {{end}}

    {{if .ShardCPU}}
    <div class="metric-card">
        <h3>CPU Usage by Shard (Core)</h3>
        <p><em>CPU utilization per shard (0-100%).</em></p>
        <div style="height: 300px;">
            <canvas id="shardCpuChart"></canvas>
        </div>
        <script>
            (function() {
                const labels = [
                    {{range .ShardCPU}}"Shard {{.ShardID}}",{{end}}
                ];
                const data = [
                    {{range .ShardCPU}}{{.Usage}},{{end}}
                ];

                function tryInit() {
                    if (window.initShardCpuChart) {
                        window.initShardCpuChart(labels, data);
                    } else {
                        setTimeout(tryInit, 50);
                    }
                }
                tryInit();
            })();
        </script>
    </div>
    {{end}}

    {{if .IOMetrics}}
    <div class="metric-card">
        <h3>IO Operations</h3>
        <table class="sortable-table">
            <thead>
                <tr>
                    <th>Shard</th>
                    <th>IO Class</th>
                    <th>Read Ops</th>
                    <th>Write Ops</th>
                </tr>
            </thead>
            <tbody>
                {{range .IOMetrics}}
                <tr>
                    <td>{{.ShardID}}</td>
                    <td>{{.Class}}</td>
                    <td>{{printf "%.0f" .ReadOps}}</td>
                    <td>{{printf "%.0f" .WriteOps}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}

</body>
</html>


================================================
FILE: ui/html/nav.tmpl
================================================
{{define "nav"}}
<nav>
    <div class="nav-links">
        {{if not .LogsOnly}}
        <a href="/" class="{{if eq .Active "Home"}}active{{end}}">Home</a>
        <a href="/partitions" class="{{if eq .Active "Partitions"}}active{{end}}">Partitions</a>
        <a href="/disk" class="{{if eq .Active "Disk"}}active{{end}}">Disk</a>
        <a href="/kafka" class="{{if eq .Active "Kafka"}}active{{end}}">Kafka</a>
        <a href="/groups" class="{{if eq .Active "Groups"}}active{{end}}">Groups</a>
        <a href="/k8s" class="{{if eq .Active "K8s"}}active{{end}}">K8s</a>
        <span class="nav-sep"></span>
        {{end}}

        <a href="/logs" class="{{if eq .Active "Logs"}}active{{end}}">Logs</a>
        {{if not .LogsOnly}}
        <a href="/metrics" class="{{if eq .Active "Metrics"}}active{{end}}">Metrics</a>
        <a href="/system" class="{{if eq .Active "System"}}active{{end}}">System</a>
        <a href="/segments" class="{{if eq .Active "Segments"}}active{{end}}">Segments</a>
        <span class="nav-sep"></span>
        
        <a href="/timeline" class="{{if eq .Active "Timeline"}}active{{end}}">Timeline</a>
        <a href="/skew" class="{{if eq .Active "Skew"}}active{{end}}">Skew Analysis</a>
        <a href="/diagnostics" class="{{if eq .Active "Diagnostics"}}active{{end}}">Diagnostics</a>
        {{end}}
    </div>

    {{$activeSession := ""}}
    {{if .Sessions}}
        {{$activeSession = index .Sessions .ActivePath}}
    {{end}}
    {{if $activeSession}}
    <div class="nav-info-bar">
        <span>Viewing Bundle:</span>
        <span class="nav-bundle-name">{{$activeSession.Name}}</span>
    </div>
    {{end}}

    <div class="theme-switch-wrapper" style="display: flex; gap: 10px; align-items: center;">
        {{if .Sessions}}
        <div class="theme-selection-container">
            <button class="bundle-select-btn">
                {{if gt (len .Sessions) 1}}
                    Switch Bundles
                {{else}}
                    Analyze new bundle
                {{end}}
            </button>
            <div class="bundle-select-menu">
                {{range $path, $session := .Sessions}}
                <button class="bundle-select-option" onclick="window.location.href='/api/switch-bundle?path={{$path}}'">{{$session.Name}}</button>
                {{end}}
                <hr style="border: 0; border-top: 1px solid var(--border-color); margin: 4px 0;">
                <button class="bundle-select-option" onclick="window.location.href='/setup'">+ Analyze New Bundle</button>
            </div>
        </div>
        {{else}}
        <a href="/setup" class="button" style="padding: 4px 12px; font-size: 0.75rem; background-color: var(--card-bg); color: var(--text-color); border: 1px solid var(--border-color);">Analyze New Bundle</a>
        {{end}}

        <div class="theme-selection-container">
            <button class="theme-select-btn">Themes</button>
            <div class="theme-select-menu">
                <button class="theme-select-option" data-theme="light">Light</button>
                <button class="theme-select-option" data-theme="dark">Dark</button>
                <button class="theme-select-option" data-theme="ultra-dark">Ultra Dark</button>
                <button class="theme-select-option" data-theme="retro">Retro</button>
                <button class="theme-select-option" data-theme="compact">Compact</button>
                <button class="theme-select-option" data-theme="powershell">Powershell</button>
            </div>
        </div>
    </div>
</nav>
{{end}}



================================================
FILE: ui/html/partitions.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Partitions Overview</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/tables.js"></script>
	<script src="/static/htmx.min.js"></script>
</head>
<body>
    							
    							</div>    			    {{template "nav" (dict "Active" "Partitions" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    				<h1>Partitions Overview</h1>
    <h2>Summary</h2>
    <table>
        <tbody>
            <tr>
                <th>Total Partitions</th>
                <td>{{.TotalPartitions}}</td>
            </tr>
        </tbody>
    </table>
    <h2>Partitions per Node</h2>
    <table id="partitions-table">
        <thead>
            <tr>
                <th class="sortable">Node ID</th>
                <th class="sortable">Partition Count</th>
            </tr>
        </thead>
        <tbody>
            {{range $nodeID, $count := .PartitionsPerNode}}
            <tr>
                <td>{{$nodeID}}</td>
                <td>{{$count}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <h2>Leaders per Node</h2>
    <table id="leaders-table">
        <thead>
            <tr>
                <th class="sortable">Node ID</th>
                <th class="sortable">Leader Count</th>
            </tr>
        </thead>
        <tbody>
            {{range $nodeID, $count := .LeadersPerNode}}
            <tr>
                <td>{{$nodeID}}</td>
                <td>{{$count}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    <h2>Topics</h2>
    <table id="topics-table">
        <thead>
            <tr>
                <th class="sortable">Topic Name</th>
                <th class="sortable">Partition Count</th>
                <th class="sortable">Replication Factor</th>
            </tr>
        </thead>
        <tbody>
            {{range .Topics}}
            <tr onclick="toggleTopicDetails('{{.Name}}')">
                <td><a href="#">{{.Name}}</a></td>
                <td>{{len .Partitions}}</td>
                <td>{{if .Partitions}}{{len (index .Partitions 0).Replicas}}{{end}}</td>
            </tr>
            <tr id="{{.Name}}-details" style="display:none;">
                <td colspan="3">
                    {{/* This will be populated by JavaScript */}}
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
    <script src="/static/common.js"></script>
    <script src="/static/partitions_view.js"></script>
</body>
</html>



================================================
FILE: ui/html/search.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Global Search Results</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="/static/theme.js"></script>
    <style>
        .search-result {
            margin-bottom: 20px;
            padding: 15px;
            background-color: var(--header-bg-color);
            border: 1px solid var(--border-color);
            border-radius: 6px;
        }
        .search-category {
            font-size: 0.85em;
            text-transform: uppercase;
            color: var(--text-color-light);
            margin-bottom: 5px;
        }
        .search-title {
            font-size: 1.1em;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .search-title a {
            text-decoration: none;
            color: var(--link-color);
        }
        .search-title a:hover {
            text-decoration: underline;
        }
        .search-preview {
            font-family: monospace;
            font-size: 0.9em;
            white-space: pre-wrap;
            color: var(--text-color);
            background-color: var(--bg-color);
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
        }
        mark {
            background-color: yellow;
            color: black;
        }
        [data-theme="dark"] mark, [data-theme="ultra-dark"] mark {
            background-color: #b8860b;
            color: white;
        }
    </style>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
	
    <h1>Global Search Results for "{{.Query}}"</h1>

    <div style="margin-bottom: 20px;">
        <form action="/search" method="GET" style="display: flex; gap: 10px; max-width: 600px;">
            <input type="text" name="q" placeholder="Global Search..." value="{{.Query}}" style="flex-grow: 1; padding: 8px; font-size: 14px; border: 1px solid var(--border-color); border-radius: 4px; background-color: var(--bg-color); color: var(--text-color);">
            <button type="submit" style="padding: 8px 16px; font-size: 14px; background-color: var(--link-color); color: var(--bg-color); border: none; border-radius: 4px; cursor: pointer;">Search</button>
        </form>
    </div>

    {{if .Results}}
        <p>Found {{len .Results}} results.</p>
        {{range .Results}}
        <div class="search-result">
            <div class="search-category">{{.Category}}</div>
            <div class="search-title"><a href="{{.Link}}">{{.Title}}</a></div>
            {{if .Preview}}
            <div class="search-preview">{{.Preview}}</div>
            {{end}}
        </div>
        {{end}}
    {{else}}
        <p>No results found.</p>
    {{end}}

</body>
</html>



================================================
FILE: ui/html/segment_view.tmpl
================================================
<h3>File: {{.FilePath}}</h3>
<p><strong>Size:</strong> {{formatBytes .SizeBytes}}</p>
{{if .Error}}
<div class="alert alert-danger">
    <strong>Error:</strong> {{.Error}}
</div>
{{end}}

{{if .Snapshot}}
<div class="card" style="margin-bottom: 20px;">
    <h3>Raft Snapshot Header</h3>
    <div class="stat-row">
        <span class="stat-label">Last Included Index</span>
        <span class="stat-value">{{.Snapshot.LastIncludedIndex}}</span>
    </div>
    <div class="stat-row">
        <span class="stat-label">Last Included Term</span>
        <span class="stat-value">{{.Snapshot.LastIncludedTerm}}</span>
    </div>
    <div class="stat-row">
        <span class="stat-label">Metadata Size</span>
        <span class="stat-value">{{.Snapshot.MetadataSize}} bytes</span>
    </div>
    <div class="stat-row">
        <span class="stat-label">Version</span>
        <span class="stat-value">{{.Snapshot.HeaderVersion}}</span>
    </div>
    <div class="stat-row">
        <span class="stat-label">Header CRC</span>
        <span class="stat-value" style="font-family: monospace;">{{printf "%x" .Snapshot.HeaderCRC}}</span>
    </div>
</div>
{{else}}
<p>Found {{len .Batches}} batches.</p>

<table class="sortable-table">
    <thead>
        <tr>
            <th>Offset (Base/Max)</th>
            <th>Timestamp (Max)</th>
            <th>Type</th>
            <th>Batch Size</th>
            <th>Records</th>
            <th>Compress</th>
            <th>CRC</th>
        </tr>
    </thead>
    <tbody>
        {{range .Batches}}
        <tr>
            <td style="font-family: monospace;">{{.BaseOffset}} / {{.MaxOffset}}</td>
            <td>{{.Timestamp.Format "2006-01-02 15:04:05.000"}}</td>
            <td>
                {{.TypeString}}
                {{if .RaftConfig}}
                <div style="font-size: 0.8em; color: var(--text-color-light); margin-top: 4px;">
                    <strong>Nodes:</strong> {{range .RaftConfig.Voters}}{{.}}, {{end}}
                </div>
                {{end}}
            </td>
            <td>{{.BatchLength}}</td>
            <td>{{.RecordCount}}</td>
            <td>{{.CompressionType}}</td>
            <td style="font-family: monospace;">{{printf "%x" .CRC}}</td>
        </tr>
        {{end}}
    </tbody>
</table>
{{end}}



================================================
FILE: ui/html/segments.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Log Segments</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "Segments" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    <h1>Log Segments Inspector</h1>
    <p>Select a binary log segment (.log) to inspect its Record Batch headers.</p>

    <div style="display: grid; grid-template-columns: 300px 1fr; gap: 20px;">
        <div class="card" style="height: 80vh; overflow-y: auto;">
            <h3>Available Segments</h3>
            <ul style="list-style: none; padding: 0;">
                {{range .Files}}
                <li style="margin-bottom: 5px;">
                    <button class="link-btn" 
                            hx-get="/segments/view?file={{.}}" 
                            hx-target="#segment-detail" 
                            hx-indicator="#loading"
                            style="text-align: left; width: 100%; border: 1px solid var(--border-color); background: none; color: var(--link-color); padding: 8px; cursor: pointer;">
                        {{.}}
                    </button>
                </li>
                {{else}}
                <li>No .log files found in controller-logs/</li>
                {{end}}
            </ul>
        </div>

        <div class="card" style="height: 80vh; overflow-y: auto; padding: 20px;">
            <div id="loading" class="htmx-indicator">Loading segment data...</div>
            <div id="segment-detail">
                <p>Select a file to view details.</p>
            </div>
        </div>
    </div>
</body>
</html>



================================================
FILE: ui/html/setup.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>BundleViewer - Setup</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    <div class="theme-switch-wrapper">
        <div class="theme-selection-container">
            <button class="theme-select-btn">Themes</button>
            <div class="theme-select-menu">
                <button class="theme-select-option" data-theme="light">Light</button>
                <button class="theme-select-option" data-theme="dark">Dark</button>
                <button class="theme-select-option" data-theme="ultra-dark">Ultra Dark</button>
                <button class="theme-select-option" data-theme="retro">Retro</button>
                <button class="theme-select-option" data-theme="compact">Compact</button>
                <button class="theme-select-option" data-theme="powershell">Powershell</button>
            </div>
        </div>
    </div>

    <div id="loading-overlay" style="display: none;">
		<div id="loading-box">
			<h2 id="loading-title">Initializing...</h2>
			<div id="progress-container">
				<div id="progress-bar">0%</div>
			</div>
            <p id="loading-status" style="margin-top: 1rem; color: var(--text-color-muted); font-size: 0.875rem;">Parsing bundle files...</p>
		</div>
	</div>

    <div class="container" style="height: 100vh; display: flex; align-items: center; justify-content: center;">
        <div class="card" style="max-width: 500px; width: 100%;">
            <div style="text-align: center; margin-bottom: 2rem;">
                <h1 style="margin-bottom: 0.5rem;">Welcome to BundleViewer</h1>
                <p style="color: var(--text-color-muted);">Select a Redpanda diagnostic bundle to analyze.</p>
            </div>

            <form hx-post="/api/setup" hx-target="body" hx-indicator="#loading-overlay" onsubmit="startProgress()">
                <div style="margin-bottom: 1.5rem;">
                    <label style="display: block; margin-bottom: 0.5rem; font-weight: 600;">Bundle Directory Path</label>
                    <div style="display: flex; gap: 0.5rem;">
                        <input type="text" id="bundlePath" name="bundlePath" placeholder="/path/to/extracted/bundle" required style="flex-grow: 1;">
                        <button type="button" class="button" style="background-color: var(--card-bg); color: var(--text-color); border: 1px solid var(--border-color); white-space: nowrap;" onclick="browseFolders()">Browse...</button>
                    </div>
                    <p style="font-size: 0.75rem; color: var(--text-color-muted); margin-top: 0.5rem;">Provide the absolute path to the directory containing the unzipped bundle files.</p>
                </div>

                <div style="margin-bottom: 2rem; display: flex; align-items: center; gap: 0.75rem;">
                    <input type="checkbox" id="logsOnly" name="logsOnly" style="width: auto;">
                    <label for="logsOnly" style="font-weight: 500; cursor: pointer;">Logs Only Mode</label>
                </div>

                <div style="display: flex; gap: 1rem;">
                    <button type="submit" style="flex-grow: 2; padding: 0.75rem; font-size: 1rem;">Start Analysis</button>
                    {{if .CanCancel}}
                    <button type="button" class="button" style="flex-grow: 1; padding: 0.75rem; font-size: 1rem; background-color: var(--card-bg); color: var(--text-color); border: 1px solid var(--border-color);" onclick="window.location.href='/'">Cancel</button>
                    {{end}}
                </div>
            </form>
        </div>
    </div>

    <script>
        async function browseFolders() {
            try {
                const response = await fetch('/api/browse');
                const data = await response.json();
                if (data.path) {
                    document.getElementById('bundlePath').value = data.path;
                }
            } catch (err) {
                console.error("Failed to open folder picker:", err);
                alert("Native folder picker is only supported when running the server locally.");
            }
        }

        function startProgress() {
            const overlay = document.getElementById('loading-overlay');
            overlay.style.display = 'flex';
            
            const progressBar = document.getElementById('progress-bar');
            const statusText = document.getElementById('loading-status');
            const eventSource = new EventSource('/api/load-progress');

            eventSource.onmessage = function(event) {
                const data = JSON.parse(event.data);
                const progress = data.progress;
                const status = data.status;
                const finished = data.finished;

                progressBar.style.width = progress + '%';
                progressBar.textContent = progress + '%';
                if (status) {
                    statusText.textContent = status;
                }

                if (finished) {
                    eventSource.close();
                }
            };

            eventSource.onerror = function(err) {
                console.error("EventSource failed:", err);
                eventSource.close();
            };
        }
    </script>
</body>
</html>



================================================
FILE: ui/html/skew.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Partition Skew Analysis</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/tables.js"></script>
    <style>
        .skew-card {
            background-color: var(--header-bg-color);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }

        .skew-positive { color: #d32f2f; font-weight: bold; } /* Overloaded */
        .skew-negative { color: #1976d2; font-weight: bold; } /* Underloaded */
        .skew-normal { color: var(--text-color); }

        .stat-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .stat-box {
            background-color: var(--bg-color);
            padding: 15px;
            border-radius: 6px;
            border: 1px solid var(--border-color);
            text-align: center;
        }

        .stat-value {
            font-size: 1.8em;
            font-weight: bold;
            color: var(--link-color);
        }

        .stat-label {
            font-size: 0.9em;
            color: var(--text-color-light);
            margin-top: 5px;
        }

        .bar-container {
            display: flex;
            align-items: center;
            height: 20px;
            background-color: var(--border-color);
            border-radius: 4px;
            overflow: hidden;
            width: 100px;
        }

        .bar-fill {
            height: 100%;
            background-color: var(--link-color);
        }

        .bar-fill.skewed {
            background-color: #fbc02d;
        }

        .bar-fill.critically-skewed {
            background-color: #d32f2f;
        }

        .config-grid {
            display: grid;
            grid-template-columns: 250px 1fr;
            gap: 10px;
        }

        .config-label {
            font-weight: bold;
            color: var(--text-color-light);
        }
    </style>
</head>
<body>
	
	</div>
    {{template "nav" (dict "Active" "Skew" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
	<h1>Skew Analysis</h1>

    <div class="skew-card">
        <h3>Cluster Summary</h3>
        <div class="stat-grid">
            <div class="stat-box">
                <div class="stat-value">{{.Analysis.TotalPartitions}}</div>
                <div class="stat-label">Total Partitions</div>
            </div>
            <div class="stat-box">
                <div class="stat-value">{{.Analysis.TotalNodes}}</div>
                <div class="stat-label">Total Nodes</div>
            </div>
            <div class="stat-box">
                <div class="stat-value">{{printf "%.1f" .Analysis.MeanPartitions}}</div>
                <div class="stat-label">Mean Partitions/Node</div>
            </div>
            <div class="stat-box">
                <div class="stat-value">{{printf "%.2f" .Analysis.StdDev}}</div>
                <div class="stat-label">Standard Deviation</div>
            </div>
        </div>
    </div>

    <div class="skew-card">
        <h3>Node Balance</h3>
        <p>Analyzing deviation from the mean partition count. >20% deviation is highlighted.</p>

        <table class="sortable-table">
            <thead>
                <tr>
                    <th>Node ID</th>
                    <th>Partition Count</th>
                    <th>Topic Count</th>
                    <th>Deviation from Mean</th>
                    <th>Skew %</th>
                    <th>Status</th>
                    <th>Balancing Score</th>
                </tr>
            </thead>
            <tbody>
                {{range $i, $nodeSkew := .Analysis.NodeSkews}}
                <tr>
                    <td>{{$nodeSkew.NodeID}}</td>
                    <td>{{$nodeSkew.PartitionCount}}</td>
                    <td>{{$nodeSkew.TopicCount}}</td>
                    <td>
                        {{if gt $nodeSkew.Deviation 0.0}}+{{end}}{{printf "%.2f" $nodeSkew.Deviation}} σ
                    </td>
                    <td class="{{if gt $nodeSkew.SkewPercentage 20.0}}skew-positive{{else if lt $nodeSkew.SkewPercentage -20.0}}skew-negative{{else}}skew-normal{{end}}">
                        {{if gt $nodeSkew.SkewPercentage 0.0}}+{{end}}{{printf "%.1f" $nodeSkew.SkewPercentage}}%
                    </td>
                    <td>
                        {{if $nodeSkew.IsSkewed}}
                            {{if gt $nodeSkew.SkewPercentage 0.0}}Overloaded{{else}}Underloaded{{end}}
                        {{else}}
                            Balanced
                        {{end}}
                    </td>
                    <td>
                        {{ with (index $.Analysis.Balancing.Nodes $i) }}
                            {{printf "%.8f" .Score}}
                        {{ end }}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    {{if .Analysis.TopicSkews}}
    <div class="skew-card">
        <h3>Top Skewed Topics</h3>
        <p>Topics with uneven replica distribution across the cluster.</p>
        <table class="sortable-table">
            <thead>
                <tr>
                    <th>Topic Name</th>
                    <th>Total Replicas</th>
                    <th>Max Skew %</th>
                    <th>Node Distribution (Node: Count)</th>
                </tr>
            </thead>
            <tbody>
                {{range .Analysis.TopicSkews}}
                <tr>
                    <td>{{.TopicName}}</td>
                    <td>{{.PartitionCount}}</td>
                    <td class="skew-positive">+{{printf "%.1f" .MaxSkew}}%</td>
                    <td>
                        {{range $nodeID, $count := .NodeDistribution}}
                            {{if gt $count 0}}
                                <span class="node-badge">
                                    Node {{$nodeID}}: <b>{{$count}}</b>
                                </span>
                            {{end}}
                        {{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}

    <div class="skew-card">
        <h3>Partition Balancing Configuration</h3>
        <div class="config-grid">
            <div class="config-label">Topic Aware Balancing:</div>
            <div>{{.Analysis.Balancing.Config.TopicAware}}</div>

            <div class="config-label">Rack Awareness Enabled:</div>
            <div>{{.Analysis.Balancing.Config.RackAwarenessEnabled}}</div>

            <div class="config-label">Partitions Per Shard:</div>
            <div>{{.Analysis.Balancing.Config.PartitionsPerShard}}</div>

            <div class="config-label">Reserved Partitions (Shard 0):</div>
            <div>{{.Analysis.Balancing.Config.PartitionsReserveShard0}}</div>
        </div>
        <p style="margin-top: 15px;">A lower balancing score is better. A move is only triggered if a destination node has a lower score than the source.</p>
    </div>

    <div class="skew-card">
        <h3>Node Balancing Scores</h3>
         <table class="sortable-table">
            <thead>
                <tr>
                    <th>Node ID</th>
                    <th>CPU Cores</th>
                    <th>Max Capacity</th>
                    <th>Score</th>
                </tr>
            </thead>
            <tbody>
                {{range .Analysis.Balancing.Nodes}}
                <tr>
                    <td>{{.NodeID}}</td>
                    <td>{{.CPU}}</td>
                    <td>{{.MaxCapacity}}</td>
                    <td>{{printf "%.8f" .Score}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <div class="skew-card">
        <h3>Disk Space Management Analysis</h3>
        <p>This section explains how Redpanda manages disk space and when it triggers rebalancing based on disk usage.</p>

        <h4>Configuration Settings</h4>
        <div class="config-grid">
            <div class="config-label">Space Management Enabled:</div>
            <div>{{.Analysis.Balancing.DiskConfig.SpaceManagementEnabled}}</div>

            <div class="config-label">Disk Reservation Percent:</div>
            <div>{{.Analysis.Balancing.DiskConfig.DiskReservationPercent}}%</div>

            <div class="config-label">Retention Local Target Capacity Percent:</div>
            <div>{{.Analysis.Balancing.DiskConfig.RetentionLocalTargetCapacityPercent}}%</div>

            <div class="config-label">Retention Local Target Capacity Bytes:</div>
            <div>{{if gt .Analysis.Balancing.DiskConfig.RetentionLocalTargetCapacityBytes 0}}{{.Analysis.Balancing.DiskConfig.RetentionLocalTargetCapacityBytes | formatBytes}}{{- else}}Disabled{{- end}}</div>

            <div class="config-label">Partition Autobalancing Max Disk Usage Percent (Soft Max):</div>
            <div>{{.Analysis.Balancing.DiskConfig.PartitionAutobalancingMaxDiskUsagePercent}}%</div>

            <div class="config-label">Storage Space Alert Free Threshold Percent (Hard Max):</div>
            <div>{{.Analysis.Balancing.DiskConfig.StorageSpaceAlertFreeThresholdPercent}}%</div>
        </div>

        <p style="margin-top: 15px;">
            Rebalancing also considers a 'hard max' disk usage, defined by <code>storage_space_alert_free_threshold_percent</code>. The balancer will not move partitions to a node if its disk usage would exceed <code>100% - {{.Analysis.Balancing.DiskConfig.StorageSpaceAlertFreeThresholdPercent}}%</code> of the total disk.
        </p>
    </div>
</body>
</html>



================================================
FILE: ui/html/system.tmpl
================================================
<!DOCTYPE html>
<html lang="en">
<head>
    <title>System Status</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <style>
        .system-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .card {
            background-color: var(--header-bg-color);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 20px;
        }
        .stat-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 5px;
        }
        .stat-row:last-child {
            border-bottom: none;
        }
        .stat-label {
            font-weight: bold;
            color: var(--text-color);
        }
        .stat-value {
            font-family: monospace;
        }
        .progress-bar {
            background-color: var(--border-color);
            height: 10px;
            border-radius: 5px;
            overflow: hidden;
            margin-top: 5px;
        }
        .progress-fill {
            background-color: var(--link-color);
            height: 100%;
        }
        .warning { background-color: #ff9800; }
        .danger { background-color: #f44336; }
    </style>
</head>
<body>
    					
    					</div>    	    {{template "nav" (dict "Active" "System" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    	    <h1>System Status: {{.NodeHostname}}</h1>    <div class="system-grid">
        <!-- System Info -->
        <div class="card">
            <h3>System Info</h3>
            {{if .System.Uname.KernelName}}
            <div class="stat-row">
                <span class="stat-label">Kernel</span>
                <span class="stat-value">{{.System.Uname.KernelName}} {{.System.Uname.KernelRelease}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Arch</span>
                <span class="stat-value">{{.System.Uname.Machine}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">OS</span>
                <span class="stat-value">{{.System.Uname.OperatingSystem}}</span>
            </div>
            {{if .System.DMI.Manufacturer}}
            <div class="stat-row">
                <span class="stat-label">Manufacturer</span>
                <span class="stat-value">{{.System.DMI.Manufacturer}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Product</span>
                <span class="stat-value">{{.System.DMI.Product}}</span>
            </div>
            {{end}}
            {{if .System.Dig.Status}}
             <div class="stat-row">
                <span class="stat-label">DNS Status</span>
                <span class="stat-value" style="{{if ne .System.Dig.Status "NOERROR"}}color: red;{{end}}">{{.System.Dig.Status}}</span>
            </div>
            {{end}}
            {{if .System.CPU.ModelName}}
            <h4 style="margin-top: 15px; margin-bottom: 5px;">CPU <span style="font-weight: normal; font-size: 0.7em; color: var(--text-color-light);">(Seastar/Vectorized Support)</span></h4>
            <div class="stat-row">
                <span class="stat-label">Model</span>
                <span class="stat-value" title="{{.System.CPU.ModelName}}">{{slice .System.CPU.ModelName 0 30}}...</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">MHz</span>
                <span class="stat-value">{{.System.CPU.Mhz}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Cache</span>
                <span class="stat-value">{{.System.CPU.CacheSize}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Flags</span>
                <span class="stat-value" style="font-size: 0.8em;" title="Look for: avx, avx2, sse4_2 (Critical for CRC32C)">{{range .System.CPU.Flags}}{{.}} {{end}}</span>
            </div>
            {{end}}
            {{if .System.CmdLine}}
            <h4 style="margin-top: 15px; margin-bottom: 5px;">Boot Params <span style="font-weight: normal; font-size: 0.7em; color: var(--text-color-light);">(Latency Tuning)</span></h4>
            <div style="font-family: monospace; font-size: 0.8em; word-break: break-all; max-height: 60px; overflow-y: auto; color: var(--text-color-light);" title="Look for: isolcpus, hugepages, max_cstate=0">
                {{.System.CmdLine}}
            </div>
            {{end}}
            {{else}}
            <p>System info not available.</p>
            {{end}}
        </div>

        <!-- RAID Status -->
        {{if .System.MDStat.Arrays}}
        <div class="card">
            <h3>Software RAID (MD)</h3>
            <p style="font-size: 0.9em; margin-bottom: 10px;"><em>Degraded arrays cause high latency.</em></p>
            <p><strong>Personalities:</strong> {{range .System.MDStat.Personalities}}{{.}} {{end}}</p>
            {{range .System.MDStat.Arrays}}
            <div style="border-top: 1px solid var(--border-color); padding-top: 10px; margin-top: 10px;">
                <div class="stat-row">
                    <span class="stat-label">{{.Name}}</span>
                    <span class="stat-value">{{.State}}</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Status</span>
                    <span class="stat-value" style="{{if contains .Status "_"}}color: red; font-weight: bold;{{end}}">{{.Status}}</span>
                </div>
                 <div class="stat-row">
                    <span class="stat-label">Size</span>
                    <span class="stat-value">{{formatBytes .Blocks}} (Blocks)</span>
                </div>
                <p style="font-size: 0.8em; color: var(--text-color-light);">{{range .Devices}}{{.}} {{end}}</p>
            </div>
            {{end}}
        </div>
        {{end}}
        
        <!-- VM Stats -->
        <div class="card">
            <h3>VM Statistics</h3>
             {{if .System.VMStat.ContextSwitches}}
            <div class="stat-row">
                <span class="stat-label">Context Switches</span>
                <span class="stat-value">{{.System.VMStat.ContextSwitches}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Forks</span>
                <span class="stat-value">{{.System.VMStat.ProcessesForked}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Procs Running</span>
                <span class="stat-value">{{.System.VMStat.ProcsRunning}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label" title="Processes waiting for I/O. > 0 indicates disk saturation.">Procs Blocked</span>
                <span class="stat-value" {{if gt .System.VMStat.ProcsBlocked 0}}style="color: red; font-weight: bold;"{{end}}>{{.System.VMStat.ProcsBlocked}}</span>
            </div>
            {{else}}
            <p>VM statistics not available.</p>
            {{end}}
        </div>

        <!-- Live Load (VMStat TS) -->
        {{if .System.VMStatAnalysis.Samples}}
        <div class="card">
            <h3>Live System Load <span style="font-weight: normal; font-size: 0.7em; color: var(--text-color-light);">
(During Collection)</span></h3>
            <p style="font-size: 0.8em; color: var(--text-color-light); margin-bottom: 10px; line-height: 1.4;">
                <strong>Run Queue (r):</strong> Should be &lt; Cores. High = CPU Saturation.<br>
                <strong>Blocked (b):</strong> Should be near 0. High = Disk Saturation.<br>
                <strong>IO Wait (wa):</strong> Should be low. High = Disk Bottleneck.
            </p>
            <div class="stat-row">
                <span class="stat-label">Run Queue (r)</span>
                <span class="stat-value">Avg: {{printf "%.1f" .System.VMStatAnalysis.AvgRunnable}} / Max: {{.System.VMStatAnalysis.MaxRunnable}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Blocked (b)</span>
                <span class="stat-value" {{if gt .System.VMStatAnalysis.AvgBlocked 1.0}}style="color: red; font-weight: bold;"{{end}}>Avg: {{printf "%.1f" .System.VMStatAnalysis.AvgBlocked}} / Max: {{.System.VMStatAnalysis.MaxBlocked}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">IO Wait (wa)</span>
                <span class="stat-value" {{if gt .System.VMStatAnalysis.AvgIOWait 10.0}}style="color: red; font-weight: bold;"{{end}}>Avg: {{printf "%.1f" .System.VMStatAnalysis.AvgIOWait}}% / Max: {{.System.VMStatAnalysis.MaxIOWait}}%</span>
            </div>
        </div>
        {{end}}

        <!-- Load Average -->
        <div class="card">
            <h3>Load Average</h3>
            {{if .System.CoreCount}}
                <p style="font-size: 0.9em; color: var(--text-color-light); margin-bottom: 10px;">
                    <strong>CPU Cores: {{.System.CoreCount}}</strong><br>
                    Load > {{.System.CoreCount}} indicates saturation.
                </p>
            {{end}}
            <div class="stat-row">
                <span class="stat-label">1 min</span>
                <span class="stat-value" {{if and .System.CoreCount (gt .System.Load.OneMin (float64 .System.CoreCount))}}style="color: red; font-weight: bold;"{{end}}>{{.System.Load.OneMin}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">5 min</span>
                <span class="stat-value" {{if and .System.CoreCount (gt .System.Load.FiveMin (float64 .System.CoreCount))}}style="color: red; font-weight: bold;"{{end}}>{{.System.Load.FiveMin}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">15 min</span>
                <span class="stat-value" {{if and .System.CoreCount (gt .System.Load.FifteenMin (float64 .System.CoreCount))}}style="color: red; font-weight: bold;"{{end}}>{{.System.Load.FifteenMin}}</span>
            </div>
        </div>
        </div>

        <!-- Memory -->
        <div class="card">
            <h3>Memory</h3>
            <div class="stat-row">
                <span class="stat-label">Total</span>
                <span class="stat-value">{{formatBytes .System.Memory.Total}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Used</span>
                <span class="stat-value">{{formatBytes .System.Memory.Used}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Free</span>
                <span class="stat-value">{{formatBytes .System.Memory.Free}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Available</span>
                <span class="stat-value">{{formatBytes .System.Memory.Available}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Swap Used</span>
                <span class="stat-value">{{formatBytes .System.Memory.SwapUsed}}</span>
            </div>
            
            {{if .System.MemInfo.AnonHugePages}}
            <h4 style="margin-top: 15px; margin-bottom: 5px;">Extended Mem</h4>
            <div class="stat-row">
                <span class="stat-label">AnonHugePages</span>
                <span class="stat-value">{{formatBytes .System.MemInfo.AnonHugePages}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Slab</span>
                <span class="stat-value">{{formatBytes .System.MemInfo.Slab}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Dirty</span>
                <span class="stat-value">{{formatBytes .System.MemInfo.Dirty}}</span>
            </div>
            {{end}}
        </div>
        
        <!-- VM Stats -->
        <div class="card">
            <h3>VM Statistics</h3>
             {{if .System.VMStat.ContextSwitches}}
            <div class="stat-row">
                <span class="stat-label">Context Switches</span>
                <span class="stat-value">{{.System.VMStat.ContextSwitches}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Forks</span>
                <span class="stat-value">{{.System.VMStat.ProcessesForked}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Procs Running</span>
                <span class="stat-value">{{.System.VMStat.ProcsRunning}}</span>
            </div>
            <div class="stat-row">
                <span class="stat-label">Procs Blocked</span>
                <span class="stat-value" {{if gt .System.VMStat.ProcsBlocked 0}}style="color: red; font-weight: bold;"{{end}}>{{.System.VMStat.ProcsBlocked}}</span>
            </div>
            {{else}}
            <p>VM statistics not available.</p>
            {{end}}
        </div>

    </div>

    <!-- Disk Usage -->
    <div class="card">
        <h3>File Systems</h3>
        <table class="sortable-table">
            <thead>
                <tr>
                    <th>Mount</th>
                    <th>Filesystem</th>
                    <th>Type</th>
                    <th>Size</th>
                    <th>Used</th>
                    <th>Avail</th>
                    <th>Use%</th>
                </tr>
            </thead>
            <tbody>
                {{range .System.FileSystems}}
                <tr>
                    <td>{{.MountPoint}}</td>
                    <td>{{.Filesystem}}</td>
                    <td>{{.Type}}</td>
                    <td>{{formatBytes .Total}}</td>
                    <td>{{formatBytes .Used}}</td>
                    <td>{{formatBytes .Available}}</td>
                    <td>
                        {{.UsePercent}}
                        {{if contains .UsePercent "%"}}
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: {{.UsePercent}};"></div>
                        </div>
                        {{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <!-- Active Connections -->
    {{if .System.Connections}}
    <div class="card" style="margin-top: 20px;">
        <h3>Active TCP/UDP Connections</h3>
        
        <div style="display: flex; gap: 20px; flex-wrap: wrap; margin-bottom: 20px;">
            <div style="flex: 1; min-width: 200px;">
                <h4>By State</h4>
                <table class="sortable-table">
                    <thead><tr><th>State</th><th>Count</th></tr></thead>
                    <tbody>
                        {{range $state, $count := .System.ConnSummary.ByState}}
                        <tr><td>{{$state}}</td><td>{{$count}}</td></tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            <div style="flex: 1; min-width: 200px;">
                <h4>Top Ports (Local)</h4>
                <div style="max-height: 200px; overflow-y: auto;">
                    <table class="sortable-table">
                        <thead><tr><th>Port</th><th>Count</th></tr></thead>
                        <tbody>
                            {{range .System.ConnSummary.SortedByPort}}
                            <tr><td>{{.Port}}</td><td>{{.Count}}</td></tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
            <div style="flex: 1; min-width: 150px; display: flex; align-items: center; justify-content: center; flex-direction: column;">
                 <span style="font-size: 3em; font-weight: bold; color: var(--link-color);">{{.System.ConnSummary.Total}}</span>
                 <span style="text-transform: uppercase; letter-spacing: 1px; color: var(--text-color);">Total Connections</span>
            </div>
        </div>

        <p><em>Filtered to exclude most localhost traffic unless relevant to Redpanda ports.</em></p>
        <div style="max-height: 500px; overflow-y: auto;">
            <table class="sortable-table">
                <thead>
                    <tr>
                        <th>Proto</th>
                        <th>State</th>
                        <th>Local Address</th>
                        <th>Peer Address</th>
                        <th>Recv-Q</th>
                        <th>Send-Q</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .System.Connections}}
                    <tr>
                        <td>{{.Netid}}</td>
                        <td>{{.State}}</td>
                        <td>{{.LocalAddr}}</td>
                        <td>{{.PeerAddr}}</td>
                        <td>{{.RecvQ}}</td>
                        <td>{{.SendQ}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>
    {{end}}

    <!-- Kernel Config -->
    <div class="card" style="margin-top: 20px;">
        <h3>Kernel Configuration (Sysctl)</h3>
        <table class="sortable-table">
            <thead>
                <tr>
                    <th>Key</th>
                    <th>Value</th>
                </tr>
            </thead>
            <tbody>
                {{range $k, $v := .System.Sysctl}}
                <tr>
                    <td>{{$k}}</td>
                    <td>{{$v}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

</body>
</html>



================================================
FILE: ui/html/timeline.tmpl
================================================
{{if not .Partial}}
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Event Timeline</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/theme.js"></script>
    <script src="/static/htmx.min.js"></script>
    <script src="/static/chart.js"></script>
    <style>
        .timeline-container {
            position: relative;
            max-width: 1000px;
            margin: 0 auto;
            padding: 20px 0;
        }

        /* The vertical line */
        .timeline-container::after {
            content: '';
            position: absolute;
            width: 4px;
            background-color: var(--border-color);
            top: 0;
            bottom: 0;
            left: 50px; /* Adjust for timestamp width */
            margin-left: -2px;
        }

        .timeline-event {
            padding: 10px 40px;
            position: relative;
            background-color: inherit;
            width: 100%;
            box-sizing: border-box;
            display: flex;
        }

        .timeline-timestamp {
            width: 100px; /* Fixed width for timestamp */
            text-align: right;
            padding-right: 20px;
            font-size: 0.85em;
            color: var(--text-color-light);
            flex-shrink: 0;
        }

        .timeline-content {
            padding: 15px;
            background-color: var(--header-bg-color);
            border: 1px solid var(--border-color);
            border-radius: 6px;
            flex-grow: 1;
            position: relative;
        }

        /* The circle on the line */
        .timeline-content::after {
            content: '';
            position: absolute;
            width: 16px;
            height: 16px;
            right: 100%;
            background-color: var(--header-bg-color);
            border: 4px solid var(--link-color);
            top: 15px;
            border-radius: 50%;
            z-index: 1;
            margin-right: 12px; /* Distance from box */
            transform: translateX(50%); /* Center on line */
            left: -32px; /* Position relative to content box to hit the line at 50px */
        }
        
        /* Adjust circle position based on timestamp width (100px) and line position (50px) */
        /* Actually simpler: line is at 50px absolute. 
           Timestamp is 0-100px.
           Content starts at ~100px + padding.
           Circle needs to be at 50px absolute.
        */
        
        .event-Critical .timeline-content {
            border-left: 5px solid #d32f2f;
        }
        .event-Critical .timeline-content::after {
            border-color: #d32f2f;
        }

        .event-Warning .timeline-content {
            border-left: 5px solid #fbc02d;
        }
        .event-Warning .timeline-content::after {
            border-color: #fbc02d;
        }

        .event-source {
            font-weight: bold;
            font-size: 0.9em;
            margin-bottom: 5px;
            color: var(--link-color);
        }
        
        .event-desc {
            font-family: monospace;
            white-space: pre-wrap;
            word-break: break-all;
        }
    </style>
</head>
<body>
    					
    					</div>    	    {{template "nav" (dict "Active" "Timeline" "Sessions" .Sessions "ActivePath" .ActivePath "LogsOnly" .LogsOnly)}}
    		<h1>Event Timeline</h1>    <p>Chronological view of critical errors, warnings, and system events.</p>

    <div id="timeline-content" hx-get="/api/timeline/data" hx-trigger="load" hx-swap="outerHTML">
        <div class="container" style="height: 60vh; display: flex; flex-direction: column; align-items: center; justify-content: center;">
            <div class="card" style="text-align: center; max-width: 400px; width: 100%;">
                <h2 style="margin-bottom: 1rem;">Building Timeline...</h2>
                <p style="color: var(--text-color-muted); margin-bottom: 2rem;">Correlating logs and system events to build a chronological view.</p>
                <div class="spinner" style="border: 4px solid var(--border-color); width: 48px; height: 48px; border-radius: 50%; border-left-color: var(--primary-color); animation: spin 1s linear infinite; display: inline-block;"></div>
                <style>
                    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
                </style>
            </div>
        </div>
    </div>
    <script src="/static/common.js"></script>
    <script src="/static/timeline_view.js"></script>
</body>
</html>
{{else}}
    <div style="max-width: 1000px; margin: 0 auto;">
        <div class="chart-container" style="position: relative; height: 150px; width: 100%; margin-bottom: 20px; background: var(--header-bg-color); border: 1px solid var(--border-color); border-radius: 6px; padding: 10px;">
            <canvas id="timelineChart"></canvas>
        </div>
    </div>
    
    <div class="timeline-container">
        {{range .Events}}
        <div class="timeline-event event-{{.Level}}">
            <div class="timeline-timestamp">
                {{.Timestamp.Format "15:04:05"}}<br>
                <span style="font-size: 0.8em;">{{.Timestamp.Format "Jan 02"}}</span>
            </div>
            <div class="timeline-content">
                <div class="event-source">{{.Type}}: {{.Source}}</div>
                <div class="event-desc">{{.Description}}</div>
            </div>
        </div>
        {{end}}
    </div>
{{end}}



================================================
FILE: ui/static/common.js
================================================
// Common UI functions

// Toggle Visibility Function
function toggleVisibility(id, linkElement) {
    var el = document.getElementById(id);
    if (!el) return;

    var icon = null;
    if (linkElement) {
        icon = linkElement.querySelector('.toggle-icon');
    }

    // Determine appropriate display type
    var targetDisplay = 'block';
    if (el.tagName === 'TR') {
        targetDisplay = 'table-row';
    } else if (el.tagName === 'TD' || el.tagName === 'TH') {
        targetDisplay = 'table-cell';
    }

    if (el.style.display === 'none') {
        el.style.display = targetDisplay;
        if (icon) icon.classList.add('expanded');
        if (linkElement) linkElement.classList.add('expanded');
    } else {
        el.style.display = 'none';
        if (icon) icon.classList.remove('expanded');
        if (linkElement) linkElement.classList.remove('expanded');
    }
}

// Back to top button logic
document.addEventListener("DOMContentLoaded", function() {
    var mybutton = document.getElementById("back-to-top");
    if (mybutton) {
        window.onscroll = function() {scrollFunction()};
        mybutton.addEventListener("click", function() {
            document.body.scrollTop = 0; // For Safari
            document.documentElement.scrollTop = 0; // For Chrome, Firefox, IE and Opera
        });
    }

    function scrollFunction() {
        if (!mybutton) return;
        if (document.body.scrollTop > 20 || document.documentElement.scrollTop > 20) {
            mybutton.style.display = "block";
        } else {
            mybutton.style.display = "none";
        }
    }
});

// Custom Multi-Select Dropdown Helper
class CustomMultiSelect {
    constructor(selectElement, placeholder = "Select options...") {
        this.selectElement = selectElement;
        this.placeholder = placeholder;
        this.selectedValues = new Set();
        
        // Initialize from existing select
        Array.from(this.selectElement.selectedOptions).forEach(opt => {
            this.selectedValues.add(opt.value);
        });

        // Hide original select
        this.selectElement.style.display = 'none';

        // Create UI elements
        this.container = document.createElement('div');
        this.container.className = 'custom-multiselect-container';
        
        this.button = document.createElement('div');
        this.button.className = 'custom-multiselect-button';
        this.updateButtonText();
        
        this.dropdown = document.createElement('div');
        this.dropdown.className = 'custom-multiselect-dropdown';
        
        // Build options
        const options = Array.from(this.selectElement.options);
        
        options.forEach(opt => {
            if (opt.value === "") return; // Skip placeholder/empty value options if any

            const optionDiv = document.createElement('div');
            optionDiv.className = 'custom-multiselect-option';
            
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = opt.value;
            checkbox.checked = this.selectedValues.has(opt.value);
            
            const label = document.createElement('span');
            label.textContent = opt.text;
            
            optionDiv.appendChild(checkbox);
            optionDiv.appendChild(label);
            
            // Toggle logic
            optionDiv.addEventListener('click', (e) => {
                // Prevent bubbling to button click if clicked directly
                if (e.target !== checkbox) {
                    checkbox.checked = !checkbox.checked;
                }
                
                if (checkbox.checked) {
                    this.selectedValues.add(opt.value);
                    opt.selected = true;
                } else {
                    this.selectedValues.delete(opt.value);
                    opt.selected = false;
                }
                this.updateButtonText();
                
                // Trigger change event on original select so other listeners know
                this.selectElement.dispatchEvent(new Event('change'));
            });

            this.dropdown.appendChild(optionDiv);
        });

        this.container.appendChild(this.button);
        this.container.appendChild(this.dropdown);
        
        // Insert after original select
        this.selectElement.parentNode.insertBefore(this.container, this.selectElement.nextSibling);

        // Event Listeners
        this.button.addEventListener('click', (e) => {
            // Only toggle if we didn't click a pill remove button
            if (!e.target.classList.contains('multiselect-pill-remove')) {
                e.stopPropagation();
                this.toggleDropdown();
            }
        });

        document.addEventListener('click', (e) => {
            if (!this.container.contains(e.target)) {
                this.closeDropdown();
            }
        });
    }

    toggleDropdown() {
        this.container.classList.toggle('open');
    }

    closeDropdown() {
        this.container.classList.remove('open');
    }

    deselect(value) {
        this.selectedValues.delete(value);
        
        // Update original select
        Array.from(this.selectElement.options).forEach(opt => {
             if (opt.value === value) opt.selected = false;
        });

        // Update Checkbox in dropdown
        const checkbox = this.dropdown.querySelector(`input[value="${CSS.escape(value)}"]`);
        if (checkbox) checkbox.checked = false;

        this.updateButtonText();
        this.selectElement.dispatchEvent(new Event('change'));
    }

    updateButtonText() {
        this.button.innerHTML = ''; // Clear content

        const selectedOptions = Array.from(this.selectElement.options)
            .filter(opt => this.selectedValues.has(opt.value));

        if (selectedOptions.length === 0) {
            this.button.textContent = this.placeholder;
            this.button.classList.add('placeholder');
            return;
        }

        this.button.classList.remove('placeholder');

        // Logic: Show up to 2 pills, then +N
        const maxDisplay = 2;
        const displayItems = selectedOptions.slice(0, maxDisplay);
        const remaining = selectedOptions.length - maxDisplay;

        displayItems.forEach(opt => {
            const pill = document.createElement('span');
            pill.className = 'multiselect-pill';
            
            const text = document.createElement('span');
            text.textContent = opt.text;
            text.style.maxWidth = '100px';
            text.style.overflow = 'hidden';
            text.style.textOverflow = 'ellipsis';
            
            const removeBtn = document.createElement('span');
            removeBtn.className = 'multiselect-pill-remove';
            removeBtn.innerHTML = '&times;';
            removeBtn.title = "Remove";
            removeBtn.onclick = (e) => {
                e.stopPropagation(); // Stop bubbling to button click
                this.deselect(opt.value);
            };

            pill.appendChild(text);
            pill.appendChild(removeBtn);
            this.button.appendChild(pill);
        });

        if (remaining > 0) {
            const morePill = document.createElement('span');
            morePill.className = 'multiselect-pill more-pill';
            morePill.textContent = `+${remaining}`;
            morePill.title = "Show all selected items";
            this.button.appendChild(morePill);
        }
    }
}



================================================
FILE: ui/static/log_analysis_view.js
================================================
document.body.addEventListener('htmx:responseError', function(evt) {
    const content = document.getElementById("analysis-content");
    if (content) {
        content.innerHTML = 
            '<div style="color: red; text-align: center; padding: 20px;">' +
            '<h3>Error Loading Analysis</h3>' +
            '<p>Server returned status: ' + evt.detail.xhr.status + '</p>' +
            '<p>Check server logs for details.</p>' +
            '</div>';
    }
});

document.body.addEventListener('htmx:sendError', function(evt) {
    const content = document.getElementById("analysis-content");
    if (content) {
        content.innerHTML = 
            '<div style="color: red; text-align: center; padding: 20px;">' +
            '<h3>Connection Error</h3>' +
            '<p>Failed to connect to server. Is it running?</p>' +
            '</div>';
    }
});



================================================
FILE: ui/static/logs_view.js
================================================
// ui/static/logs_view.js

if (typeof window.logsViewInitialized === 'undefined') {
    window.logsViewInitialized = false;
}

// Global State
window.allLogs = [];
window.isFetching = false;
window.hasMoreLogsServer = true;
window.serverOffset = 0;
window.serverLimit = 1000; // Fetch larger chunks for virtual scroll
window.rowHeight = 85; // Accommodate ~3 lines of wrapped text
window.lastScrollY = -1;
window.renderFrameRequestId = null;

// DOM Elements
window.logTableBody = null;
window.logTable = null;
window.loadingIndicator = null;
window.filterState = {};

if (!window.logsViewInitialized) {
    window.initLogsView = function() {
        window.logTableBody = document.querySelector('.log-table tbody');
        window.logTable = document.querySelector('.log-table');
        window.loadingIndicator = document.getElementById('loading-indicator');
        
        // Initial Fetch
        window.applyFilters(true);

        // Scroll Event Listener (Throttled via RequestAnimationFrame)
        window.addEventListener('scroll', () => {
            if (!window.renderFrameRequestId) {
                window.renderFrameRequestId = requestAnimationFrame(window.renderVirtualLogs);
            }
        });

        // Resize Listener
        window.addEventListener('resize', () => {
            window.rowHeight = window.estimateRowHeight();
            window.renderVirtualLogs();
        });

        // Infinite Scroll Trigger (Approaching bottom)
        const infiniteScrollObserver = new IntersectionObserver((entries) => {
            if (entries[0].isIntersecting && window.hasMoreLogsServer && !window.isFetching) {
                window.fetchLogs();
            }
        }, { rootMargin: '200px' });
        
        if (window.loadingIndicator) {
            infiniteScrollObserver.observe(window.loadingIndicator);
        }

        // Search Input
        const queryInput = document.getElementById('query-input');
        if (queryInput) {
            queryInput.addEventListener('keypress', function (e) {
                if (e.key === 'Enter') window.applyFilters();
            });
        }

        // Initialize Custom Multi-Selects
        const levelSelect = document.getElementById('level-select');
        if (levelSelect) {
            new CustomMultiSelect(levelSelect, "All Levels");
        }
        
        const componentSelect = document.getElementById('component-select');
        if (componentSelect) {
            new CustomMultiSelect(componentSelect, "All Components");
        }

        const nodeSelect = document.getElementById('node-select');
        if (nodeSelect) {
            new CustomMultiSelect(nodeSelect, "All Nodes");
        }

        // Analyze Logs Button
        const analyzeBtn = document.getElementById('analyze-logs-button');
        if (analyzeBtn) {
            analyzeBtn.addEventListener('click', function() {
                const indicator = document.getElementById('analyze-loading-indicator');
                if (indicator) indicator.style.display = 'inline';
                window.location.href = '/logs/analysis';
            });
        }

        // Modal interactions
        window.addEventListener('click', function(event) {
            const modal = document.getElementById('log-context-modal');
            if (modal && event.target === modal) {
                modal.style.display = 'none';
            }
        });

        window.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                const modal = document.getElementById('log-context-modal');
                if (modal && modal.style.display === 'block') {
                    modal.style.display = 'none';
                }
            }
        });
    };

    window.estimateRowHeight = function() {
        const firstRow = window.logTableBody ? window.logTableBody.querySelector('tr.log-row') : null;
        return firstRow ? firstRow.offsetHeight : (window.rowHeight || 45);
    };

    window.buildApiUrl = function(offset, limit) {
        const params = new URLSearchParams();
        for (const [key, value] of Object.entries(window.filterState)) {
            if (Array.isArray(value)) {
                value.forEach(v => params.append(key, v));
            } else if (value) {
                params.set(key, value);
            }
        }
        params.set('offset', offset);
        params.set('limit', limit);
        return '/api/logs?' + params.toString();
    };

    window.applyFilters = function(isInitial = false) {
        // Capture filter state
        const levelSelect = document.getElementById('level-select');
        const componentSelect = document.getElementById('component-select');
        const nodeSelect = document.getElementById('node-select');
        window.filterState = {
            search: document.getElementById('query-input')?.value || '',
            level: levelSelect ? Array.from(levelSelect.selectedOptions).map(o => o.value) : [],
            node: nodeSelect ? Array.from(nodeSelect.selectedOptions).map(o => o.value) : [],
            component: componentSelect ? Array.from(componentSelect.selectedOptions).map(o => o.value) : [],
            sort: document.getElementById('sort-select')?.value || 'desc',
            startTime: document.getElementById('start-time')?.value || '',
            endTime: document.getElementById('end-time')?.value || ''
        };

        // Reset State
        window.allLogs = [];
        window.serverOffset = 0;
        window.hasMoreLogsServer = true;
        window.lastScrollY = -1;
        
        if (window.logTableBody) window.logTableBody.innerHTML = '';
        window.scrollTo(0, 0);

        window.fetchLogs();
    };

    window.fetchLogs = async function() {
        if (window.isFetching || !window.hasMoreLogsServer) return;
        window.isFetching = true;
        if (window.loadingIndicator) window.loadingIndicator.textContent = 'Loading logs...';

        const url = window.buildApiUrl(window.serverOffset, window.serverLimit);

        try {
            const response = await fetch(url);
            const data = await response.json();

            if (data.logs && data.logs.length > 0) {
                // Pre-process logs (escape HTML, linkify) for faster rendering later
                const processed = data.logs.map(log => {
                    const escapedMsg = escapeHtml(log.message);
                    return {
                        ...log,
                        escapedMsg: escapedMsg,
                        linkedMsg: linkifyMessage(escapedMsg),
                        hasTrace: log.message.includes('\n')
                    };
                });

                window.allLogs.push(...processed);
                window.serverOffset += data.logs.length;
                window.hasMoreLogsServer = data.logs.length === window.serverLimit;
                
                // Update Total Count UI
                const displaySpan = document.getElementById('log-count-display');
                if (displaySpan) {
                    displaySpan.textContent = `Loaded ${window.allLogs.length} of ${data.total} matching logs`;
                }

                // If first load, measure row height after render
                const isFirstLoad = window.allLogs.length === data.logs.length;
                window.renderVirtualLogs(true);
                
                if (isFirstLoad) {
                    setTimeout(() => {
                        window.rowHeight = window.estimateRowHeight();
                        window.renderVirtualLogs(true); // Re-render with correct height
                    }, 50);
                }

            } else {
                window.hasMoreLogsServer = false;
                if (window.allLogs.length === 0 && window.logTableBody) {
                    window.logTableBody.innerHTML = '<tr><td colspan="7" style="text-align: center; padding: 20px;">No logs found.</td></tr>';
                }
            }
        } catch (error) {
            console.error("Fetch error:", error);
            if (window.loadingIndicator) window.loadingIndicator.textContent = 'Error loading logs.';
        } finally {
            window.isFetching = false;
            window.renderFrameRequestId = null;
            if (window.loadingIndicator) {
                 window.loadingIndicator.textContent = window.hasMoreLogsServer ? 'Scroll for more...' : 'End of logs.';
            }
        }
    };

    window.renderVirtualLogs = function(force = false) {
        window.renderFrameRequestId = null;

        if (!window.logTable || !window.logTableBody || window.allLogs.length === 0) return;

        const scrollY = window.scrollY;
        
        // Optimization: Don't re-render if scroll hasn't changed enough to shift rows
        // unless forced (e.g. new data arrived)
        if (!force && Math.abs(scrollY - window.lastScrollY) < (window.rowHeight / 2)) {
           return;
        }
        window.lastScrollY = scrollY;

        // Calculate Visible Window
        // We assume the table starts after the header. 
        // The header is sticky, but the *document* scroll position determines where we are in the "list".
        // The table's top position relative to the document:
        const tableTop = window.logTable.offsetTop + window.logTable.tHead.offsetHeight;
        const viewportHeight = window.innerHeight;
        
        // Relative scroll position within the data list
        const relativeScroll = Math.max(0, scrollY - tableTop);
        
        const buffer = 10; // Extra rows top/bottom
        const startIndex = Math.max(0, Math.floor(relativeScroll / window.rowHeight) - buffer);
        const endIndex = Math.min(window.allLogs.length, Math.ceil((relativeScroll + viewportHeight) / window.rowHeight) + buffer);

        // Calculate Padding
        const paddingTop = startIndex * window.rowHeight;
        const paddingBottom = (window.allLogs.length - endIndex) * window.rowHeight;

        // Render
        const fragment = document.createDocumentFragment();

        // Top Spacer
        if (paddingTop > 0) {
            const tr = document.createElement('tr');
            tr.style.height = `${paddingTop}px`;
            tr.style.border = 'none'; // Prevent borders on spacers
            // Needs a cell to be valid HTML, but we can hide it or make it empty
            const td = document.createElement('td');
            td.colSpan = 7;
            td.style.padding = 0;
            td.style.border = 'none';
            tr.appendChild(td);
            fragment.appendChild(tr);
        }

        // Data Rows
        const searchInput = document.getElementById('query-input');
        const searchTerms = getSearchTerms(searchInput ? searchInput.value : '');

        for (let i = startIndex; i < endIndex; i++) {
            const log = window.allLogs[i];
            const tr = document.createElement('tr');
            tr.className = 'log-row';
            tr.style.height = `${window.rowHeight}px`; // Enforce height
            
            // View Trace Button
            let actionBtn = '';
            if (log.hasTrace) {
                actionBtn = `<button class="trace-btn" onclick="openTraceModal(${i})" style="font-size: 0.8em; padding: 2px 6px; cursor: pointer;">View Trace</button>`;
            }

            // Message Display (Truncated if too long for one line, but we rely on CSS mostly)
            // We use a div inside the cell to enforce max-height/ellipsis if needed
            // Updated to allow 3 lines of wrapping
            const messageHtml = `<div style="max-height: ${window.rowHeight}px; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; white-space: normal; line-height: 1.4;">${log.linkedMsg}</div>`;

            tr.innerHTML = `
                <td style="white-space:nowrap;">${new Date(log.timestamp).toISOString().replace('T', ' ').substring(0, 23)}</td>
                <td class="log-level-${log.level}">${log.level}</td>
                <td>${log.node}</td>
                <td>${log.shard}</td>
                <td>${log.component}</td>
                <td style="max-width: 0; width: 40%; overflow: hidden;">${messageHtml}</td>
                <td>
                    <button onclick="showLogContext('${log.filePath}', ${log.lineNumber})" style="font-size: 0.8em;">Context</button>
                    ${actionBtn}
                </td>
            `;

            // Highlighting (Performance intensive, strictly limit to visible rows)
            if (searchTerms.length > 0) {
                const msgDiv = tr.cells[5].firstElementChild;
                if(msgDiv) highlightElement(msgDiv, searchTerms);
            }

            fragment.appendChild(tr);
        }

        // Bottom Spacer
        if (paddingBottom > 0) {
            const tr = document.createElement('tr');
            tr.style.height = `${paddingBottom}px`;
            tr.style.border = 'none';
            const td = document.createElement('td');
            td.colSpan = 7;
            td.style.padding = 0;
            td.style.border = 'none';
            tr.appendChild(td);
            fragment.appendChild(tr);
        }

        window.logTableBody.innerHTML = '';
        window.logTableBody.appendChild(fragment);
    };

    // --- Helper Functions ---
    
    window.openTraceModal = function(index) {
        const log = window.allLogs[index];
        if (!log) return;
        
        const modal = document.getElementById('log-context-modal'); // Reuse context modal
        const body = document.getElementById('log-context-body');
        const header = modal.querySelector('h2');
        
        if (header) header.textContent = 'Log Trace';
        if (body) {
            body.innerHTML = `
                <div style="margin-bottom: 10px; font-weight: bold;">${log.level} @ ${log.timestamp}</div>
                <div style="background: var(--table-alt-row-bg); padding: 10px; border-radius: 4px; overflow-x: auto;">
                    ${log.escapedMsg} 
                </div>
            `;
        }
        modal.style.display = 'block';
    };

    window.logsViewInitialized = true;
}

// Helpers (Keep existing ones mostly)
function escapeHtml(text) {
    if (!text) return '';
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}

function linkifyMessage(text) {
    // Basic regex for K8s resources - same as before
    const regex = /(pod|node|service|pvc|deployment|statefulset)[s]?([\/: ]\s*)([a-z0-9](?:[-a-z0-9]*[a-z0-9])?)/gi;
    return text.replace(regex, (match, type, sep, name) => {
        let idPrefix = type.toLowerCase();
        if (idPrefix.endsWith('s')) idPrefix = idPrefix.slice(0, -1);
        const map = { 'pod': 'pod-', 'node': 'node-', 'service': 'service-', 'pvc': 'pvc-', 'deployment': 'deploy-', 'statefulset': 'sts-' };
        const prefix = map[idPrefix] || (idPrefix + '-');
        return `${type}${sep}<a href="/k8s#${prefix}${name}" target="_blank" style="text-decoration: underline; color: var(--link-color);">${name}</a>`;
    });
}

function getSearchTerms(query) {
    if (!query) return [];
    const tokens = query.match(/"[^ vital ]+"|\S+/g) || [];
    const terms = [];
    const operators = new Set(['AND', 'OR', 'NOT', '(', ')']);
    tokens.forEach(token => {
        let term = token;
        if (term.startsWith('"') && term.endsWith('"') && term.length >= 2) term = term.substring(1, term.length - 1);
        if (operators.has(term.toUpperCase())) return;
        const colonIdx = term.indexOf(':');
        if (colonIdx > 0) {
            // Very basic field support for highlighting only
            const value = term.substring(colonIdx + 1);
            terms.push(value);
        } else {
            terms.push(term);
        }
    });
    return terms.filter(t => t.length > 0);
}

function highlightElement(element, terms) {
    if (!terms || terms.length === 0) return;
    try {
        const distinctTerms = [...new Set(terms)].sort((a, b) => b.length - a.length);
        if (distinctTerms.length === 0) return;
        
        // Simple regex escape
        const pattern = new RegExp(`(${distinctTerms.map(t => t.replace(/[.*+?^${}()|[\\]/g, '\\$&')).join('|')})`, 'gi');
        
        // We only highlight text nodes to preserve HTML (links)
        const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT, null, false);
        const nodes = [];
        while(walker.nextNode()) nodes.push(walker.currentNode);
        
        nodes.forEach(node => {
             if (node.parentNode.tagName === 'MARK' || node.parentNode.tagName === 'A') return; // Skip already highlighted or links
             const text = node.nodeValue;
             if (pattern.test(text)) {
                 const span = document.createElement('span');
                 span.innerHTML = escapeHtml(text).replace(pattern, '<mark>$1</mark>');
                 node.parentNode.replaceChild(span, node);
             }
        });
    } catch(e) { console.error("Highlight error", e); }
}

// Re-expose legacy functions for buttons
function clearFilters() {
    window.location.reload();
}

async function showLogContext(filePath, lineNumber) {
    const modal = document.getElementById('log-context-modal');
    const contextBody = document.getElementById('log-context-body');
    const header = modal.querySelector('h2');
    if(header) header.textContent = 'Log Context';
    
    if (!contextBody) return;
    contextBody.textContent = 'Loading...';
    if (modal) modal.style.display = 'block';
    
    try {
        const safeFilePath = encodeURIComponent(filePath);
        const url = `/api/logs/context?filePath=${safeFilePath}&lineNumber=${lineNumber}&contextSize=10`;
        const response = await fetch(url);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const contextLines = await response.json();
        contextBody.innerHTML = '';
        contextLines.forEach((line, index) => {
            const lineElem = document.createElement('div');
            lineElem.textContent = line;
            lineElem.className = 'log-context-line';
            if (index % 2 === 0) lineElem.classList.add('even-line'); else lineElem.classList.add('odd-line');
            if (index === 10) { lineElem.style.backgroundColor = 'var(--highlight-bg)'; lineElem.style.fontWeight = 'bold'; }
            contextBody.appendChild(lineElem);
        });
    } catch (error) {
        contextBody.textContent = `Error loading log context: ${error.message}`;
    }
}

function closeLogContext() {
    const modal = document.getElementById('log-context-modal');
    if (modal) modal.style.display = 'none';
}

// Init
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        if (window.initLogsView) window.initLogsView();
    });
} else if (typeof window.initLogsView === 'function') {
    window.initLogsView();
}


================================================
FILE: ui/static/metrics_view.js
================================================
// metrics_view.js - Global chart initialization functions

window.explorerChart = null;

window.loadMetricData = async function() {
    const selector = document.getElementById('metric-selector');
    const metricName = selector.value;
    const rateToggle = document.getElementById('rate-toggle');
    const asRate = rateToggle ? rateToggle.checked : false;

    if (!metricName) return;

    const loading = document.getElementById('explorer-loading');
    if (loading) loading.style.display = 'flex';

    try {
        const response = await fetch(`/api/metrics/data?metric=${encodeURIComponent(metricName)}`);
        const data = await response.json();

        window.renderExplorerChart(metricName, data, asRate);
    } catch (error) {
        console.error('Error fetching metric data:', error);
        alert('Failed to load metric data.');
    } finally {
        if (loading) loading.style.display = 'none';
    }
};

window.renderExplorerChart = function(title, seriesData, asRate) {
    const canvas = document.getElementById('explorerChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    
    const existingChart = Chart.getChart(canvas);
    if (existingChart) {
        existingChart.destroy();
    }

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || 
                       document.documentElement.getAttribute('data-theme') === 'ultra-dark' ||
                       document.documentElement.getAttribute('data-theme') === 'powershell';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    const colors = [
        '#36a2eb', '#ff6384', '#4bc0c0', '#ff9f40', '#9966ff', '#ffcd56', '#c9cbcf',
        '#00ff00', '#00ffff', '#ff00ff', '#ffff00'
    ];

    const datasets = seriesData.map((s, index) => {
        const color = colors[index % colors.length];
        
        let processedData = [];
        
        // We map data to {x, y} where x is a timestamp number (ms).
        // This is required for 'parsing: false' + 'decimation'.
        if (asRate && s.data.length > 1) {
            for (let i = 1; i < s.data.length; i++) {
                const p1 = s.data[i-1];
                const p2 = s.data[i];
                
                const t1 = new Date(p1.x).getTime();
                const t2 = new Date(p2.x).getTime();
                const deltaSeconds = (t2 - t1) / 1000;
                
                if (deltaSeconds > 0) {
                    const rate = (p2.y - p1.y) / deltaSeconds;
                    if (rate >= 0) {
                        processedData.push({
                            x: t2, 
                            y: rate
                        });
                    }
                }
            }
        } else {
            processedData = s.data.map(p => {
                return {
                    x: new Date(p.x).getTime(),
                    y: p.y
                };
            });
        }

        return {
            label: s.label + (asRate ? ' (rate/s)' : ''),
            data: processedData,
            borderColor: color,
            backgroundColor: color,
            borderWidth: 2,
            fill: false,
            pointRadius: 0, // Disable points for performance by default
            tension: 0.1
        };
    });

    new Chart(ctx, {
        type: 'line',
        data: {
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            parsing: false, // Critical for performance
            normalized: true, // Data is sorted
            animation: false, // Disable animation for large datasets
            interaction: {
                mode: 'nearest',
                axis: 'x',
                intersect: false
            },
            plugins: {
                decimation: {
                    enabled: true,
                    algorithm: 'lttb',
                    samples: 500, // Max pixels width-wise roughly
                    threshold: 500, // Only decimate if more than this many points
                },
                title: {
                    display: true,
                    text: title + (asRate ? ' (Rate)' : ''),
                    color: textColor
                },
                legend: {
                    labels: { color: textColor }
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                    callbacks: {
                        title: function(context) {
                            if (context.length > 0) {
                                const d = new Date(context[0].parsed.x);
                                return d.toISOString().substring(11, 19);
                            }
                            return '';
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'linear', // Use linear scale for numeric timestamps
                    grid: { color: gridColor },
                    ticks: { 
                        color: textColor,
                        callback: function(value) {
                            return new Date(value).toISOString().substring(11, 19);
                        }
                    },
                    title: { display: true, text: 'Time (HH:mm:ss)', color: textColor }
                },
                y: {
                    beginAtZero: true,
                    grid: { color: gridColor },
                    ticks: { color: textColor },
                    title: { display: true, text: asRate ? 'Rate / s' : 'Value', color: textColor }
                }
            }
        }
    });
};

window.initSarChart = function(labels, userCpu, sysCpu, idleCpu) {
    const canvas = document.getElementById('sarCpuChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: '% User',
                    data: userCpu,
                    borderColor: 'rgba(54, 162, 235, 1)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: '% System',
                    data: sysCpu,
                    borderColor: 'rgba(255, 99, 132, 1)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: '% Idle',
                    data: idleCpu,
                    borderColor: 'rgba(75, 192, 192, 1)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    fill: false,
                    hidden: true,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    title: { display: true, text: 'Percentage' },
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                },
                x: {
                    grid: { color: gridColor },
                    ticks: { color: textColor }
                }
            }
        }
    });
};

window.initShardCpuChart = function(labels, data) {
    const canvas = document.getElementById('shardCpuChart');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [{
                label: 'CPU Usage (%)',
                data: data,
                backgroundColor: 'rgba(54, 162, 235, 0.6)',
                borderColor: 'rgba(54, 162, 235, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    grid: {
                        color: gridColor
                    },
                    ticks: {
                        color: textColor
                    }
                },
                x: {
                    grid: {
                        color: gridColor
                    },
                    ticks: {
                        color: textColor
                    }
                }
            }
        }
    });
};



================================================
FILE: ui/static/partitions_view.js
================================================
async function toggleTopicDetails(topicName) {
    const detailRow = document.getElementById(topicName + '-details');
    
    if (detailRow.style.display === 'table-row') {
        detailRow.style.display = 'none';
        return;
    }

    // Check if already loaded
    if (detailRow.dataset.loaded === 'true') {
        detailRow.style.display = 'table-row';
        return;
    }

    // Show loading state
    detailRow.querySelector('td').innerHTML = '<p>Loading...</p>';
    detailRow.style.display = 'table-row';

    try {
        const response = await fetch(`/api/topic-details?topic=${encodeURIComponent(topicName)}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const partitions = await response.json();

        let html = '<table><thead><tr><th>Partition ID</th><th>Leader</th><th>Replicas</th></tr></thead><tbody>';
        for (const partition of partitions) {
            html += `<tr><td>${partition.id}</td><td>${partition.leader}</td><td>${partition.replicas.join(', ')}</td></tr>`;
        }
        html += '</tbody></table>';

        detailRow.querySelector('td').innerHTML = html;
        detailRow.dataset.loaded = 'true';
    } catch (error) {
        console.error('Error loading partitions:', error);
        detailRow.querySelector('td').innerHTML = '<p>Error loading partitions. See console for details.</p>';
    }
}

function makePartitionTableSortable(tableId) {
    const table = document.querySelector(tableId);
    if (!table) return;

    const headers = table.querySelectorAll('th.sortable');
    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            sortPartitionTable(table, index, header);
        });
    });
}

function sortPartitionTable(table, column, header) {
    const tbody = table.querySelector('tbody');
    
    // Get all rows - we need to identify main vs detail rows more carefully
    const allRows = Array.from(tbody.children).filter(el => el.tagName === 'TR');
    
    // Build row pairs: main rows with their associated detail rows
    const rowPairs = [];
    for (let i = 0; i < allRows.length; i++) {
        const row = allRows[i];
        
        // Skip if this row is a detail row (has an id ending in -details)
        if (row.id && row.id.endsWith('-details')) {
            continue;
        }
        
        // This is a main row
        const pair = {
            mainRow: row,
            detailRow: null
        };
        
        // Check if next row is the detail row
        const nextRow = allRows[i + 1];
        if (nextRow && nextRow.id && nextRow.id.endsWith('-details')) {
            pair.detailRow = nextRow;
        }
        
        rowPairs.push(pair);
    }
    
    const isAsc = header.classList.contains('asc');

    // Remove sort classes from all headers in this table only
    table.querySelectorAll('th.sortable').forEach(h => h.classList.remove('asc', 'desc'));

    // Add appropriate class to clicked header
    header.classList.toggle('asc', !isAsc);
    header.classList.toggle('desc', isAsc);

    rowPairs.sort((a, b) => {
        // Get the cell content, excluding any nested links
        const getCellText = (row) => {
            const cell = row.cells[column];
            if (!cell) return '';
            // If cell contains a link, get the link text, otherwise get cell text
            const link = cell.querySelector('a');
            return link ? link.textContent.trim() : cell.textContent.trim();
        };

        const aCell = getCellText(a.mainRow);
        const bCell = getCellText(b.mainRow);

        // Try to parse as number (remove any non-numeric characters except . and -)
        const aNum = parseFloat(aCell.replace(/[^0-9.-]/g, ''));
        const bNum = parseFloat(bCell.replace(/[^0-9.-]/g, ''));

        // If both are valid numbers, sort numerically
        if (!isNaN(aNum) && !isNaN(bNum)) {
            return isAsc ? bNum - aNum : aNum - bNum;
        }

        // Otherwise sort alphabetically
        return isAsc ? bCell.localeCompare(aCell) : aCell.localeCompare(bCell);
    });

    // Clear tbody
    while (tbody.firstChild) {
        tbody.removeChild(tbody.firstChild);
    }

    // Re-append sorted rows with their detail rows
    rowPairs.forEach(pair => {
        tbody.appendChild(pair.mainRow);
        if (pair.detailRow) {
            tbody.appendChild(pair.detailRow);
        }
    });
}

// Initialize after page load
document.addEventListener('DOMContentLoaded', () => {
    makePartitionTableSortable('#partitions-table');
    makePartitionTableSortable('#leaders-table');
    makePartitionTableSortable('#topics-table');
});



================================================
FILE: ui/static/progress.js
================================================
document.addEventListener("DOMContentLoaded", function() {
    // Progress Bar Logic
    const progressBar = document.getElementById('progress-bar');
    const loadingOverlay = document.getElementById('loading-overlay');
    
    if (progressBar && loadingOverlay) {
        const eventSource = new EventSource('/api/load-progress');

        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            const progress = data.progress;
            progressBar.style.width = progress + '%';
            progressBar.textContent = progress + '%';

            if (progress >= 100) {
                eventSource.close();
                setTimeout(() => {
                    loadingOverlay.style.display = 'none';
                }, 500);
            }
        };

        eventSource.onerror = function(err) {
            console.error("EventSource failed:", err);
            eventSource.close();
            loadingOverlay.style.display = 'none';
        };
    }
});


================================================
FILE: ui/static/spinner.js
================================================
// spinner.js
let spinnerCounter = 0; // To handle multiple potential spinners, though for this specific case it's one.

function showSpinner(element, spinnerClass = 'loading-spinner') {
    spinnerCounter++;
    if (!element) {
        console.warn("showSpinner called with null or undefined element.");
        return;
    }

    // Create spinner element
    const spinner = document.createElement('div');
    spinner.className = spinnerClass;
    spinner.id = 'active-spinner-' + spinnerCounter; // Unique ID

    // Basic spinner styling (can be enhanced with CSS)
    spinner.style.border = '4px solid #f3f3f3';
    spinner.style.borderTop = '4px solid #3498db';
    spinner.style.borderRadius = '50%';
    spinner.style.width = '12px';
    spinner.style.height = '12px';
    spinner.style.animation = 'spin 1s linear infinite';
    spinner.style.display = 'inline-block';
    spinner.style.marginLeft = '10px';
    spinner.style.verticalAlign = 'middle';

    // Add keyframes for spin animation if not already present
    if (!document.getElementById('spinner-keyframes')) {
        const styleSheet = document.createElement('style');
        styleSheet.id = 'spinner-keyframes';
        styleSheet.innerHTML = `
            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }
        `;
        document.head.appendChild(styleSheet);
    }
    
    // Hide the original content of the button/element and append the spinner
    element.originalContent = element.innerHTML; // Store original content
    element.innerHTML = '';
    element.appendChild(spinner);
    element.disabled = true; // Disable button while loading
}

function hideSpinner(element) {
    spinnerCounter--;
    if (!element) {
        console.warn("hideSpinner called with null or undefined element.");
        return;
    }

    const spinner = element.querySelector('[id^="active-spinner-"]');
    if (spinner) {
        spinner.remove();
        if (element.originalContent) {
            element.innerHTML = element.originalContent; // Restore original content
        }
    }
    element.disabled = false; // Re-enable button
}



================================================
FILE: ui/static/style.css
================================================
:root {
    /* Modern Light Theme (Default) */
    --primary-color: #2563eb; /* Inter Blue */
    --primary-hover: #1d4ed8;
    --bg-color: #f8fafc; /* Slate 50 */
    --card-bg: #ffffff;
    --text-color: #334155; /* Slate 700 */
    --text-color-muted: #64748b; /* Slate 500 */
    --border-color: #e2e8f0; /* Slate 200 */
    --header-bg-color: #ffffff;
    --link-color: #2563eb;
    --link-hover: #1e40af;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    
    --radius: 0.5rem;
    --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

html[data-theme="dark"] {
    --primary-color: #3b82f6;
    --primary-hover: #60a5fa;
    --bg-color: #0f172a; /* Slate 900 */
    --card-bg: #1e293b; /* Slate 800 */
    --text-color: #e2e8f0; /* Slate 200 */
    --text-color-muted: #94a3b8; /* Slate 400 */
    --border-color: #334155; /* Slate 700 */
    --header-bg-color: #1e293b;
    --link-color: #3b82f6;
    --link-hover: #60a5fa;
}

html[data-theme="ultra-dark"] {
    --bg-color: #000000;
    --card-bg: #0a0a0a;
    --text-color: #ededed;
    --border-color: #262626;
    --header-bg-color: #0a0a0a;
    --link-color: #3b82f6;
}

/* Retro Theme */
html[data-theme="retro"] {
    --bg-color: #1a0b2e;
    --card-bg: #2d1b4e;
    --text-color: #00ffff;
    --border-color: #ff00ff;
    --header-bg-color: #2d1b4e;
    --link-color: #00e5ff;
    --primary-color: #ff00ff;
    --text-color-muted: #00cccc;
    --radius: 0; /* Retro usually means sharp corners */
}
html[data-theme="retro"] body {
    font-family: 'Courier New', Courier, monospace;
}
html[data-theme="retro"] .card, 
html[data-theme="retro"] nav, 
html[data-theme="retro"] th,
html[data-theme="retro"] button,
html[data-theme="retro"] input,
html[data-theme="retro"] select {
    border: 1px solid var(--border-color);
    box-shadow: 4px 4px 0px var(--border-color);
    border-radius: 0;
}

html[data-theme="retro"] .file-list-summary:hover {
    background-color: #3d2b5e; /* Darker purple, distinct from magenta border/primary */
}

html[data-theme="retro"] .badge {
    background-color: #ff00ff;
    color: #1a0b2e;
    border: 1px solid #00ffff;
}

html[data-theme="retro"] nav a {
    color: #00cccc; /* Brighter than muted, but distinct from active */
}

html[data-theme="retro"] nav a:hover {
    color: #00ffff;
}

/* PowerShell Theme */
html[data-theme="powershell"] {
    --bg-color: #012456; /* PowerShell Dark Blue */
    --card-bg: #012456;
    --text-color: #eeedf0; /* Light Gray/White */
    --text-color-muted: #cccccc;
    --border-color: #cccccc;
    --header-bg-color: #012456;
    --link-color: #ffff00; /* Yellow */
    --primary-color: #cccccc; /* Button border color essentially */
    --primary-hover: #ffffff;
    --radius: 0;
}

html[data-theme="powershell"] body {
    font-family: 'Consolas', 'Lucida Console', monospace;
}

html[data-theme="powershell"] table, 
html[data-theme="powershell"] th, 
html[data-theme="powershell"] td, 
html[data-theme="powershell"] pre, 
html[data-theme="powershell"] button, 
html[data-theme="powershell"] input, 
html[data-theme="powershell"] select,
html[data-theme="powershell"] nav,
html[data-theme="powershell"] .card {
    border-color: #cccccc;
    background-color: #012456;
    color: #eeedf0;
    font-family: 'Consolas', 'Lucida Console', monospace;
    box-shadow: none;
}

html[data-theme="powershell"] nav a:hover {
    background-color: #003366;
    text-decoration: none;
    color: #ffff00;
}

html[data-theme="powershell"] tr:nth-child(even) {
    background-color: #012456;
}
html[data-theme="powershell"] tr:hover {
    background-color: #003366;
}

html[data-theme="powershell"] .file-list-summary:hover {
    background-color: #003366;
}

html[data-theme="powershell"] .log-level-ERROR { color: #ff3333; font-weight: bold; }
html[data-theme="powershell"] .log-level-WARN { color: #ffff00; font-weight: bold; }
html[data-theme="powershell"] .log-level-INFO { color: #eeedf0; }
html[data-theme="powershell"] .log-level-DEBUG { color: #00ffff; }
html[data-theme="powershell"] .log-level-TRACE { color: #aaaaaa; }

/* Compact Theme */
html[data-theme="compact"] {
    font-size: 13px;
    --radius: 4px;
}
html[data-theme="compact"] .card, 
html[data-theme="compact"] table {
    margin-bottom: 0.75rem;
    padding: 1rem;
}
html[data-theme="compact"] th, 
html[data-theme="compact"] td {
    padding: 0.25rem 0.5rem;
}
html[data-theme="compact"] h1 { font-size: 1.5rem; margin-bottom: 1rem; }
html[data-theme="compact"] h2 { font-size: 1.2rem; margin-top: 1.5rem; }

/* Base Reset & Typography */
* {
    box-sizing: border-box;
}

body {
    font-family: var(--font-sans);
    background-color: var(--bg-color);
    color: var(--text-color);
    margin: 0;
    line-height: 1.5;
    -webkit-font-smoothing: antialiased;
}

h1, h2, h3, h4, h5, h6 {
    margin-top: 0;
    font-weight: 600;
    line-height: 1.25;
    color: var(--text-color);
}

h1 { font-size: 1.875rem; margin-bottom: 1.5rem; }
h2 { font-size: 1.5rem; margin-bottom: 1rem; margin-top: 2rem; border-bottom: none; }
h3 { font-size: 1.25rem; margin-bottom: 0.75rem; }

a {
    color: var(--link-color);
    text-decoration: none;
    transition: color 0.2s;
}

a:hover {
    color: var(--link-hover);
    text-decoration: underline;
}

code, pre {
    font-family: var(--font-mono);
    font-size: 0.875em;
}

/* Layout */
.container {
    max-width: 1280px;
    margin: 0 auto;
    padding: 0 1rem 2rem 1rem;
}

/* Components: Navbar */
nav {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    background-color: var(--header-bg-color);
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 2rem;
    box-shadow: var(--shadow-sm);
    position: sticky;
    top: 0;
    z-index: 1000;
}

nav a {
    color: var(--text-color-muted);
    font-weight: 500;
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius);
    transition: all 0.2s;
}

nav a:hover {
    color: var(--primary-color);
    background-color: var(--bg-color);
    text-decoration: none;
}

nav a.active {
    color: var(--primary-color);
    background-color: color-mix(in srgb, var(--primary-color) 10%, transparent);
    font-weight: 600;
}

.nav-sep {
    color: var(--border-color);
    margin: 0 0.25rem;
}

/* Components: Card */
.card, .metric-card {
    background-color: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    box-shadow: var(--shadow-sm);
}

/* Components: Buttons & Inputs */
button, .button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    line-height: 1.25rem;
    color: #ffffff;
    background-color: var(--primary-color);
    border: 1px solid transparent;
    border-radius: var(--radius);
    cursor: pointer;
    transition: all 0.2s;
    text-decoration: none;
}

button:hover, .button:hover {
    background-color: var(--primary-hover);
    text-decoration: none;
}

/* General Input Styling - Default to width: 100% for block layouts */
input[type="text"], 
input[type="datetime-local"], 
select {
    display: block;
    width: 100%;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    color: var(--text-color);
    background-color: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    transition: border-color 0.2s, box-shadow 0.2s;
}

/* Scoped Input Styling for specific contexts where auto width is better */
.log-filters input[type="text"], 
.log-filters input[type="datetime-local"], 
.log-filters select,
form[action="/search"] input[type="text"] {
    width: auto; /* Allow auto width in filters */
    min-width: 150px;
}

input:focus, select:focus {
    outline: none;
    border-color: var(--primary-color);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--primary-color) 20%, transparent);
}

/* Components: Tables */
table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    margin-bottom: 1.5rem;
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    overflow: hidden;
}

th, td {
    padding: 0.75rem 1rem;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
}

th {
    background-color: var(--bg-color);
    color: var(--text-color);
    font-weight: 600;
    font-size: 0.875rem;
    white-space: nowrap;
}

th:last-child, td:last-child {
    border-right: none;
}

tr:last-child td {
    border-bottom: none;
}

tr:hover td {
    background-color: color-mix(in srgb, var(--primary-color) 5%, transparent);
}

/* Utility: Badges */
.badge {
    display: inline-flex;
    align-items: center;
    padding: 0.125rem 0.5rem;
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1;
    border-radius: 9999px;
    background-color: var(--border-color);
    color: var(--text-color);
}

/* Log Analysis Specifics */
.log-level-ERROR { color: #ef4444; }
.log-level-WARN { color: #f59e0b; }
.log-level-INFO { color: #10b981; }
.log-level-DEBUG { color: #3b82f6; }

mark {
    background-color: #fef08a; /* Yellow 200 */
    color: #854d0e; /* Yellow 800 */
    padding: 0 0.125rem;
    border-radius: 0.125rem;
}

[data-theme="dark"] mark {
    background-color: #854d0e;
    color: #fef08a;
}

/* Theme Switcher - Top Right Hover Dropdown */
.theme-switch-wrapper {
    position: absolute;
    top: 0.75rem;
    right: 1rem;
    z-index: 2000;
}

.theme-selection-container {
    position: relative;
    display: inline-block;
    padding-bottom: 10px; /* Small buffer to keep hover active */
}

.theme-toggle-btn, .theme-select-btn, .bundle-select-btn {
    background-color: var(--card-bg);
    color: var(--text-color);
    border: 1px solid var(--border-color);
    padding: 0.5rem 1rem;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    box-shadow: var(--shadow-sm);
}

.theme-menu, .theme-select-menu, .bundle-select-menu {
    position: absolute;
    top: 100%;
    right: 0;
    background-color: var(--card-bg);
    min-width: 160px;
    box-shadow: var(--shadow-md);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    display: none;
    z-index: 2001;
    overflow: hidden;
    /* Bridge the gap between button and menu so hover doesn't flicker */
    padding-top: 0; 
}

/* Invisible bridge to prevent menu from closing when moving mouse from button to menu */
.theme-menu::before, .theme-select-menu::before, .bundle-select-menu::before {
    content: '';
    position: absolute;
    top: -20px; /* Extends above the menu */
    left: 0;
    right: 0;
    height: 20px;
    background: transparent;
}

.theme-selection-container:hover .theme-menu, 
.theme-selection-container:hover .theme-select-menu,
.theme-selection-container:hover .bundle-select-menu {
    display: block;
}

.theme-option, .theme-select-option, .bundle-select-option {
    color: var(--text-color);
    padding: 0.75rem 1rem;
    text-decoration: none;
    display: block;
    width: 100%;
    text-align: left;
    border: none;
    background: none;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color 0.2s;
}

.theme-option:hover, .theme-select-option:hover, .bundle-select-option:hover {
    background-color: var(--bg-color);
    color: var(--primary-color);
}

.theme-option.active-theme, .theme-select-option.active-theme, .bundle-select-option.active-theme {
    font-weight: 600;
    background-color: color-mix(in srgb, var(--primary-color) 10%, transparent);
    color: var(--primary-color);
}

/* Adjust Nav to make room for theme switcher and second line */
nav {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    padding: 0.75rem 1rem;
    padding-right: 320px; /* Increased space for both Switch and Themes */
    gap: 0.5rem;
}

.nav-links {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
}

.nav-info-bar {
    font-size: 0.75rem;
    color: var(--text-color-muted);
    font-weight: 500;
    padding-left: 0.75rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.nav-bundle-name {
    color: var(--primary-color);
    font-family: var(--font-mono);
    font-weight: 600;
}


/* Grid Layouts */
.grid-2 { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem; }
.grid-4 { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; }

/* Timeline Specifics */
.timeline-container::after {
    background-color: var(--border-color);
}
.timeline-content {
    background-color: var(--card-bg);
    border: 1px solid var(--border-color);
    box-shadow: var(--shadow-sm);
}

/* Loading Overlay */
#loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    backdrop-filter: blur(4px);
    background-color: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
}
#loading-box {
    background-color: var(--card-bg);
    box-shadow: var(--shadow-md);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    padding: 2rem;
    border-radius: var(--radius);
    text-align: center;
    max-width: 400px;
    width: 90%;
}

/* File List Styles (for Home Page) */
.file-list {
    list-style: none;
    padding: 0;
    margin: 0;
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    background-color: var(--card-bg);
    overflow: hidden;
}

.file-list-item {
    border-bottom: 1px solid var(--border-color);
}

.file-list-item:last-child {
    border-bottom: none;
}

.file-list-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    cursor: pointer;
    background-color: var(--bg-color);
    font-weight: 500;
    transition: background-color 0.2s;
}

.file-list-summary:hover {
    background-color: var(--border-color);
}

.file-list-summary::-webkit-details-marker {
    display: none;
}

.file-list-content {
    padding: 1rem;
    border-top: 1px solid var(--border-color);
    background-color: var(--card-bg);
}

.file-name {
    font-family: var(--font-mono);
    color: var(--primary-color);
}

/* Custom Multi-Select Styles */
.custom-multiselect-container {
    position: relative;
    display: block;
    width: 200px; /* Adjust as needed */
    font-size: 0.875rem;
}

.custom-multiselect-button {
    background-color: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    padding: 2px 28px 2px 6px; /* Right padding for arrow */
    cursor: pointer;
    min-height: 38px;
    height: auto; /* Allow growth if really needed, but pills limit it */
    position: relative;
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    gap: 4px;
    overflow: hidden;
}

.custom-multiselect-button::after {
    content: '';
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    border-left: 4px solid transparent;
    border-right: 4px solid transparent;
    border-top: 4px solid var(--text-color-muted);
}

.custom-multiselect-button.placeholder {
    color: var(--text-color-muted);
    padding: 0.5rem 0.75rem; /* Standard padding for text */
}

.multiselect-pill {
    background-color: var(--border-color); /* Fallback */
    border-radius: 4px;
    padding: 2px 6px;
    font-size: 12px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    max-width: 150px;
    color: var(--text-color);
    border: 1px solid transparent;
}

/* Theme specifics for pills */
html[data-theme="light"] .multiselect-pill {
    background-color: #e2e8f0;
    border: 1px solid #cbd5e1;
}

html[data-theme="dark"] .multiselect-pill {
    background-color: #334155;
    border: 1px solid #475569;
}

.multiselect-pill-remove {
    cursor: pointer;
    font-weight: 700;
    color: var(--text-color-muted);
    font-size: 14px;
    line-height: 1;
    display: flex;
    align-items: center;
}

.multiselect-pill-remove:hover {
    color: #ef4444; /* Red */
}

.more-pill {
    background-color: transparent !important;
    border: 1px dashed var(--border-color) !important;
    color: var(--text-color-muted);
    padding: 2px 6px;
}

.custom-multiselect-dropdown {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: var(--radius);
    box-shadow: var(--shadow-md);
    z-index: 1002;
    max-height: 250px;
    overflow-y: auto;
    margin-top: 4px;
}

.custom-multiselect-container.open .custom-multiselect-dropdown {
    display: block;
}

.custom-multiselect-option {
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: background-color 0.1s;
}

.custom-multiselect-option:hover {
    background-color: var(--bg-color);
}

.custom-multiselect-option input[type="checkbox"] {
    width: auto;
    margin: 0;
    cursor: pointer;
}



================================================
FILE: ui/static/tables.js
================================================
// static/tables.js

document.addEventListener('DOMContentLoaded', function() {
    initTables();
});

function initTables() {
    const tables = document.querySelectorAll('table.sortable-table');
    tables.forEach(table => {
        makeTableSortable(table);
    });
}

function makeTableSortable(table) {
    const headers = table.querySelectorAll('th');
    headers.forEach((header, index) => {
        header.style.cursor = 'pointer';
        header.addEventListener('click', () => {
            const tbody = table.querySelector('tbody');
            const rows = Array.from(tbody.querySelectorAll('tr'));
            const isAscending = header.getAttribute('data-order') === 'asc';
            
            // Reset other headers
            headers.forEach(h => {
                h.removeAttribute('data-order');
                h.classList.remove('sorted-asc', 'sorted-desc');
            });

            // Set current header
            header.setAttribute('data-order', isAscending ? 'desc' : 'asc');
            header.classList.add(isAscending ? 'sorted-desc' : 'sorted-asc');

            rows.sort((a, b) => {
                const cellA = a.cells[index].innerText.trim();
                const cellB = b.cells[index].innerText.trim();
                
                // Try numeric sort
                const numA = parseFloat(cellA);
                const numB = parseFloat(cellB);
                
                if (!isNaN(numA) && !isNaN(numB)) {
                    return isAscending ? numB - numA : numA - numB;
                }
                
                // Date sort (simple heuristic)
                const dateA = new Date(cellA);
                const dateB = new Date(cellB);
                if (!isNaN(dateA) && !isNaN(dateB)) {
                    return isAscending ? dateB - dateA : dateA - dateB;
                }

                // String sort
                return isAscending ? cellB.localeCompare(cellA) : cellA.localeCompare(cellB);
            });

            rows.forEach(row => tbody.appendChild(row));
        });
    });
}




================================================
FILE: ui/static/theme.js
================================================
// static/theme.js
(function() {
    let currentTheme = localStorage.getItem('theme');
    // Handle legacy values
    if (currentTheme === 'dark-mode') currentTheme = 'dark';
    if (currentTheme === 'light-mode') currentTheme = 'light';

    const validThemes = ['light', 'dark', 'ultra-dark', 'compact', 'retro', 'powershell'];
    if (!validThemes.includes(currentTheme)) {
        currentTheme = 'light';
    }

    document.documentElement.setAttribute('data-theme', currentTheme);
})();

function initThemeToggle() {
    const displayBtn = document.querySelector('.theme-select-btn');
    if (!displayBtn) return;

    const themeOptions = document.querySelectorAll('.theme-select-option');

    function updateActiveTheme(theme) {
        // Update button text
        const themeName = theme.charAt(0).toUpperCase() + theme.slice(1).replace('-', ' ');
        displayBtn.textContent = `Themes: ${themeName}`;

        // Update active class on options
        themeOptions.forEach(opt => {
            if (opt.getAttribute('data-theme') === theme) {
                opt.classList.add('active-theme');
            } else {
                opt.classList.remove('active-theme');
            }
        });
    }

    // Initialize
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    updateActiveTheme(currentTheme);

    // Add listeners
    themeOptions.forEach(opt => {
        // Remove existing listener to prevent duplicates if called multiple times
        opt.onclick = null; 
        opt.addEventListener('click', function() {
            const selectedTheme = this.getAttribute('data-theme');
            document.documentElement.setAttribute('data-theme', selectedTheme);
            localStorage.setItem('theme', selectedTheme);
            updateActiveTheme(selectedTheme);
        });
    });
}

// Initial load trigger
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        initThemeToggle();
        // Add HTMX listener here to ensure body exists
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initThemeToggle();
        });
    });
} else {
    initThemeToggle();
    // Body definitely exists here
    if (!window.htmxThemeListenerAdded) {
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            initThemeToggle();
        });
        window.htmxThemeListenerAdded = true;
    }
}



================================================
FILE: ui/static/timeline_view.js
================================================
document.body.addEventListener('htmx:afterSettle', function(evt) {
    if (document.getElementById('timelineChart')) {
        loadTimelineChart();
    }
});

async function loadTimelineChart() {
    const canvas = document.getElementById('timelineChart');
    if (!canvas) return;

    try {
        const response = await fetch('/api/timeline/aggregated?bucket=2m');
        if (!response.ok) throw new Error('Failed to fetch aggregated data');
        
        const data = await response.json();
        renderTimelineChart(canvas, data);
    } catch (e) {
        console.error("Chart load error:", e);
        canvas.parentNode.innerHTML = '<div style="text-align:center; padding: 20px; color: var(--text-color-muted);">Failed to load timeline chart.</div>';
    }
}

function renderTimelineChart(canvas, data) {
    const ctx = canvas.getContext('2d');
    
    // Destroy existing chart if present to prevent "Canvas is already in use" error
    const existingChart = Chart.getChart(canvas);
    if (existingChart) {
        existingChart.destroy();
    }
    
    // Theme colors
    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark' || 
                       document.documentElement.getAttribute('data-theme') === 'ultra-dark';
    const gridColor = isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)';
    const textColor = isDarkMode ? '#ddd' : '#666';
    const barColor = isDarkMode ? 'rgba(255, 99, 132, 0.9)' : 'rgba(255, 99, 132, 0.8)';

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.map(d => new Date(d.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})),
            datasets: [{
                label: 'Errors & Warnings',
                data: data.map(d => d.count),
                backgroundColor: barColor,
                borderColor: isDarkMode ? 'rgba(255, 99, 132, 1)' : 'rgba(255, 99, 132, 0.8)',
                borderWidth: 1,
                barPercentage: 0.9,      // Restore small gap between bars
                categoryPercentage: 0.9, // Restore small gap between categories
                minBarLength: 8,         // Keep 1-event visibility
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                title: { display: true, text: 'Event Frequency (per 2 minutes)', color: textColor }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: { color: gridColor },
                    ticks: { color: textColor, precision: 0 }
                },
                x: {
                    grid: { display: false },
                    ticks: { 
                        color: textColor
                    }
                }
            },
            onClick: (e, elements) => {
                if (elements.length > 0) {
                    const idx = elements[0].index;
                    const point = data[idx];
                    // On click, navigate to logs view with this time range
                    const start = new Date(point.timestamp);
                    const end = new Date(start.getTime() + 120000); // +2 mins
                    
                    // Format for input type="datetime-local" (YYYY-MM-DDTHH:MM)
                    const fmt = d => d.toISOString().substring(0, 16);
                    window.location.href = `/logs?level=ERROR&level=WARN&startTime=${fmt(start)}&endTime=${fmt(end)}`;
                }
            }
        }
    });
}


================================================
FILE: ui/static/virtual_scroll.js
================================================
// ui/static/virtual_scroll.js
document.addEventListener('DOMContentLoaded', async () => {
    const tableContainer = document.getElementById('partition-leaders-container');
    const tbody = document.getElementById('partition-leaders-tbody');

    if (!tableContainer || !tbody) {
        console.warn("Virtual scroll elements not found.");
        return;
    }

    // Helper to handle 0 values correctly (so they don't turn into dashes)
    const safeGet = (val, fallback = '-') => (val !== undefined && val !== null) ? val : fallback;

    // 1. Fetch Data
    let data;
    try {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center;">Loading partition leaders...</td></tr>';
        
        const response = await fetch('/api/partition-leaders');
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        data = await response.json();
        
        // Debug: Check this in console if columns are still empty
        if(data.length > 0) console.log("First row data:", data[0]);

    } catch (error) {
        console.error("Failed to load partition leaders:", error);
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: red;">Failed to load partition leaders.</td></tr>';
        return;
    }

    if (!data || data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center;">No partition leaders found.</td></tr>';
        return;
    }

    // 2. Setup Virtual Scroll Constants
    const rowHeight = 35; // Set to match your CSS row height
    let visibleRows = Math.ceil(tableContainer.clientHeight / rowHeight);
    const bufferRows = 5;

    // Reset styles that might conflict
    tbody.style.display = '';
    tbody.style.position = '';
    
    let lastStartIndex = -1;

    const renderRows = () => {
        const scrollTop = tableContainer.scrollTop;
        
        // Calculate the window of data to show
        const startIndex = Math.max(0, Math.floor(scrollTop / rowHeight) - bufferRows);
        const endIndex = Math.min(data.length, startIndex + visibleRows + (2 * bufferRows));

        if (startIndex === lastStartIndex) return;
        lastStartIndex = startIndex;

        // Calculate size of invisible top and bottom areas
        const paddingTop = startIndex * rowHeight;
        const paddingBottom = (data.length - endIndex) * rowHeight;

        tbody.innerHTML = '';
        const fragment = document.createDocumentFragment();

        // A. Top Spacer Row (Pushes content down)
        if (paddingTop > 0) {
            const topRow = document.createElement('tr');
            topRow.style.height = `${paddingTop}px`;
            // Empty cell is required to maintain table structure
            const cell = document.createElement('td');
            cell.colSpan = 7; 
            cell.style.padding = 0;
            cell.style.border = 'none';
            topRow.appendChild(cell);
            fragment.appendChild(topRow);
        }

        // B. Render Visible Data Rows
        for (let i = startIndex; i < endIndex; i++) {
            const leader = data[i];
            const row = document.createElement('tr');
            row.style.height = `${rowHeight}px`;

            // Data Mapping with Fallbacks
            // Tries multiple property names to find the right data
            const cells = [
                safeGet(leader.ns || leader.namespace),
                safeGet(leader.topic),
                safeGet(leader.partition_id || leader.partition),
                safeGet(leader.leader || leader.leader_id),
                
                // Term: Tries 'term' OR 'last_stable_leader_term'
                safeGet(leader.term ?? leader.last_stable_leader_term),
                
                // Rev: Tries 'revision' OR 'rev' OR 'partition_revision'
                safeGet(leader.revision ?? leader.rev ?? leader.partition_revision),
                
                // Prev Leader: Tries 'previous_leader' OR 'previous_leader_id'
                safeGet(leader.previous_leader ?? leader.previous_leader_id)
            ];

            cells.forEach(content => {
                const cell = document.createElement('td');
                cell.textContent = content;
                
                // Styling to prevent cell blowout
                cell.style.whiteSpace = 'nowrap';
                cell.style.overflow = 'hidden';
                cell.style.textOverflow = 'ellipsis';
                cell.style.padding = '8px'; // Adjust based on your CSS
                
                row.appendChild(cell);
            });

            fragment.appendChild(row);
        }

        // C. Bottom Spacer Row (Maintains scrollbar size)
        if (paddingBottom > 0) {
            const bottomRow = document.createElement('tr');
            bottomRow.style.height = `${paddingBottom}px`;
            const cell = document.createElement('td');
            cell.colSpan = 7;
            cell.style.padding = 0;
            cell.style.border = 'none';
            bottomRow.appendChild(cell);
            fragment.appendChild(bottomRow);
        }

        tbody.appendChild(fragment);
    };

    // Initial render
    renderRows();

    // Event Listeners
    tableContainer.addEventListener('scroll', () => {
        window.requestAnimationFrame(renderRows);
    });

    const resizeObserver = new ResizeObserver(() => {
        visibleRows = Math.ceil(tableContainer.clientHeight / rowHeight);
        lastStartIndex = -1; // Force re-render
        renderRows();
    });
    resizeObserver.observe(tableContainer);

    // Disable sorting headers (since virtual scroll breaks standard sorting)
    const table = tbody.closest('table');
    if (table) {
        table.classList.remove('sortable-table');
        const headers = table.querySelectorAll('th');
        headers.forEach(th => {
            th.removeAttribute('onclick');
            th.style.cursor = 'default';
            th.classList.remove('sorted-asc', 'sorted-desc');
        });
    }
});



