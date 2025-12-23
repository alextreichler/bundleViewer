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
