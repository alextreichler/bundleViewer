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
	"github.com/alextreichler/bundleViewer/internal/downloader"
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
	"github.com/alextreichler/bundleViewer/internal/server"
	"github.com/alextreichler/bundleViewer/internal/store"
	"github.com/alextreichler/bundleViewer/internal/version"
)

var Version = "dev"

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
	envPort := 7575
	if p := os.Getenv("PORT"); p != "" {
		if val, err := fmt.Sscanf(p, "%d", &envPort); err == nil && val > 0 {
			// use envPort
		}
	}
	port := flag.Int("port", envPort, "Port to listen on (can also be set via PORT env var)")

	envHost := "127.0.0.1"
	if h := os.Getenv("HOST"); h != "" {
		envHost = h
	}
	host := flag.String("host", envHost, "Host to bind to (default: 127.0.0.1 for security, use 0.0.0.0 for containers) (can also be set via HOST env var)")
	
	// Support environment variables for cloud/container usage
	envFetchURL := os.Getenv("FETCH_URL")
	fetchURL := flag.String("fetch-url", envFetchURL, "URL to download and extract a bundle from (can also be set via FETCH_URL env var)")
	
	authToken := flag.String("auth-token", os.Getenv("AUTH_TOKEN"), "Secret token required to access the UI (can also be set via AUTH_TOKEN env var)")
	
	envShutdown := os.Getenv("SHUTDOWN_TIMEOUT")
	defaultShutdown := time.Duration(0)
	if envShutdown != "" {
		if d, err := time.ParseDuration(envShutdown); err == nil {
			defaultShutdown = d
		}
	}
	shutdownTimeout := flag.Duration("shutdown-timeout", defaultShutdown, "Hard limit on server lifetime (e.g., 20m). 0 means no limit. (can also be set via SHUTDOWN_TIMEOUT env var)")
	
	envInactivity := os.Getenv("INACTIVITY_TIMEOUT")
	defaultInactivity := time.Duration(0)
	if envInactivity != "" {
		if d, err := time.ParseDuration(envInactivity); err == nil {
			defaultInactivity = d
		}
	}
	inactivityTimeout := flag.Duration("inactivity-timeout", defaultInactivity, "Inactivity timeout (e.g., 5m). Shutdown if no UI heartbeats received. 0 means no limit. (can also be set via INACTIVITY_TIMEOUT env var)")
	
	envMemoryLimit := os.Getenv("MEMORY_LIMIT")
	memoryLimit := flag.String("memory-limit", envMemoryLimit, "Memory limit for the application (e.g., 2GB, 512MB). Adjusts SQLite cache and mmap. (can also be set via MEMORY_LIMIT env var)")

	envOneShot := os.Getenv("ONE_SHOT") == "true"
	oneShot := flag.Bool("one-shot", envOneShot, "One-shot mode: analyzes provided bundle and serves it in read-only mode. Disables bundle switching. (can also be set via ONE_SHOT=true env var)")

	persist := flag.Bool("persist", false, "Persist the database between runs (default: clean on start)")
	logsOnly := flag.Bool("logs-only", false, "Only process and display logs")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <bundle-directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	
	flag.Parse()

	if *versionFlag {
		fmt.Printf("bundleViewer version %s\n", Version)
		os.Exit(0)
	}

	// Initialize the logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Version check
	version.Current = Version
	latestVersionChan := make(chan string, 1)
	go func() {
		latest, err := version.CheckUpdate()
		if err != nil {
			logger.Debug("Failed to check for updates", "error", err)
			latestVersionChan <- ""
			return
		}
		if latest != "" {
			logger.Info("A new version of bundleViewer is available!", "latest", latest, "current", Version)
			logger.Info("Run this command to update:", "command", version.UpdateCommand())
			latestVersionChan <- latest
		} else {
			latestVersionChan <- ""
		}
	}()

	var latestVersion string
	select {
	case latestVersion = <-latestVersionChan:
	case <-time.After(1 * time.Second):
		logger.Debug("Version check timed out")
	}

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

	// Fetch bundle from URL if provided
	if *fetchURL != "" {
		extractedPath, err := downloader.FetchAndExtractBundle(*fetchURL, dataDir)
		if err != nil {
			logger.Error("Failed to fetch and extract bundle", "url", *fetchURL, "error", err)
			os.Exit(1)
		}
		// Standardize args so the rest of the logic uses the extracted path
		os.Args = append(os.Args[:1], extractedPath)
	}

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
	var initialProgress *models.ProgressTracker

	limitBytes, _ := parser.ParseSizeToBytes(*memoryLimit)

	if bundlePath != "" {
		// Determine if we should clean the DB (default true, unless -persist is set)
		cleanDB := !*persist

		// Initialize the SQLite store
		dbPath := filepath.Join(dataDir, "bundle.db")
		sqliteStore, err = store.NewSQLiteStore(dbPath, bundlePath, cleanDB, limitBytes)
		if err != nil {
			logger.Error("Failed to initialize SQLite store", "error", err)
			os.Exit(1)
		}

		// Initialize the cache with a temporary tracker
		initialProgress = models.NewProgressTracker(21)
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

	srv, err := server.New(bundlePath, cachedData, storeInterface, logger, *logsOnly, *persist, initialProgress, Version, latestVersion, *authToken, *shutdownTimeout, *inactivityTimeout, limitBytes, *oneShot) // Pass logger, logsOnly, persist
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
