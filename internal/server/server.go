package server

import (
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

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

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		
		// If the handler sets Content-Encoding later, we might have an issue.
		// But for our simple app, this should be fine.
		// A more robust version would check the content type before starting gzip.
		w.Header().Set("Content-Encoding", "gzip")
		
		next.ServeHTTP(gzw, r)
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

type BasePageData struct {
	Active          string
	Sessions        map[string]*BundleSession
	ActivePath      string
	LogsOnly        bool
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	UpdateCommand   string
	RaftRecovery    RaftRecoveryMetrics
}

type RaftRecoveryMetrics struct {
	OffsetsPending      float64
	PartitionsActive    float64
	PartitionsToRecover float64
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
	consensusTemplate   *template.Template
	segmentsTemplate    *template.Template
	segmentViewTemplate *template.Template
	cpuProfilesTemplate *template.Template
	notebookTemplate    *template.Template
	logsOnly            bool
	authToken           string
	shutdownTimeout     time.Duration
	inactivityTimeout   time.Duration
	lastHeartbeat       int64 // Unix timestamp
	currentVersion      string
	latestVersion       string
	updateAvailable     bool
}

func (s *Server) SetDataDir(path string) {
	s.dataDir = path
}

func (s *Server) newBasePageData(active string) BasePageData {
	var raftRecovery RaftRecoveryMetrics
	if s.store != nil {
		startTime, endTime := time.Time{}, time.Time{}
		offsets, _ := s.store.GetMetrics("vectorized_raft_recovery_offsets_pending", nil, startTime, endTime, 0, 0)
		for _, m := range offsets {
			raftRecovery.OffsetsPending += m.Value
		}
		activeParts, _ := s.store.GetMetrics("vectorized_raft_recovery_partitions_active", nil, startTime, endTime, 0, 0)
		for _, m := range activeParts {
			raftRecovery.PartitionsActive += m.Value
		}
		toRecover, _ := s.store.GetMetrics("vectorized_raft_recovery_partitions_to_recover", nil, startTime, endTime, 0, 0)
		for _, m := range toRecover {
			raftRecovery.PartitionsToRecover += m.Value
		}
	}

	return BasePageData{
		Active:          active,
		Sessions:        s.sessions,
		ActivePath:      s.activePath,
		LogsOnly:        s.logsOnly,
		CurrentVersion:  s.currentVersion,
		LatestVersion:   s.latestVersion,
		UpdateAvailable: s.updateAvailable,
		UpdateCommand:   "curl -sSL https://raw.githubusercontent.com/alextreichler/bundleViewer/main/install.sh | bash",
		RaftRecovery:    raftRecovery,
	}
}

// AuthMiddleware checks for a valid auth token if configured
func (s *Server) heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt64(&s.lastHeartbeat, time.Now().Unix())
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) startTimeoutWatcher() {
	if s.shutdownTimeout == 0 && s.inactivityTimeout == 0 {
		return
	}

	startTime := time.Now()
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()

			// Check hard shutdown timeout
			if s.shutdownTimeout > 0 && now.Sub(startTime) > s.shutdownTimeout {
				s.logger.Info("Hard shutdown timeout reached. Shutting down...", "timeout", s.shutdownTimeout)
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
				return
			}

			// Check inactivity timeout
			if s.inactivityTimeout > 0 {
				last := atomic.LoadInt64(&s.lastHeartbeat)
				if now.Unix()-last > int64(s.inactivityTimeout.Seconds()) {
					s.logger.Info("Inactivity timeout reached. Shutting down...", "timeout", s.inactivityTimeout)
					syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
					return
				}
			}
		}
	}()
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for token in query param
		token := r.URL.Query().Get("token")
		if token == s.authToken {
			// Valid token in URL, set cookie and redirect to same page without token
			http.SetCookie(w, &http.Cookie{
				Name:     "bv_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600 * 24, // 1 day
			})
			
			// Remove token from query string and redirect
			q := r.URL.Query()
			q.Del("token")
			r.URL.RawQuery = q.Encode()
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}

		// Check for token in cookie
		cookie, err := r.Cookie("bv_token")
		if err == nil && cookie.Value == s.authToken {
			next.ServeHTTP(w, r)
			return
		}

		// Allow static files to be served? 
		// For true security, we should probably block everything except maybe a login page.
		// But since we want "Launch -> Open", we'll just block everything with a 401.
		if strings.HasPrefix(r.URL.Path, "/static/") {
			// Optimization: allow static files so login page looks good (if we had one)
			// But for now let's just be strict.
		}

		http.Error(w, "Unauthorized: Valid token required via ?token=... or cookie", http.StatusUnauthorized)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func New(bundlePath string, cachedData *cache.CachedData, s store.Store, logger *slog.Logger, logsOnly bool, persist bool, progress *models.ProgressTracker, currentVersion, latestVersion, authToken string, shutdownTimeout, inactivityTimeout time.Duration) (http.Handler, error) { // Accept both
	funcMap := template.FuncMap{
		"renderData": func(key string, data interface{}) template.HTML {
			return renderDataLazy(key, "", data, 0, 0)
		},
		"lower": func(v interface{}) string {
			return strings.ToLower(fmt.Sprintf("%v", v))
		},
		"float64": func(v interface{}) float64 {
			switch i := v.(type) {
			case int:
				return float64(i)
			case int64:
				return float64(i)
			case float64:
				return i
			default:
				return 0.0
			}
		},
		"toFloat64": func(v interface{}) float64 {
			switch i := v.(type) {
			case int:
				return float64(i)
			case int64:
				return float64(i)
			case float64:
				return i
			default:
				return 0.0
			}
		},
		"formatBytes": func(b interface{}) template.HTML {
			var val float64
			switch v := b.(type) {
			case int:
				val = float64(v)
			case int64:
				val = float64(v)
			case float64:
				val = v
			}
			return template.HTML(formatBytes(val))
		},
		"formatBytesFloat": func(b float64) template.HTML {
			return template.HTML(formatBytes(b))
		},
		"sub": func(a, b interface{}) int64 {
			var va, vb int64
			switch v := a.(type) {
			case int: va = int64(v)
			case int64: va = v
			case float64: va = int64(v)
			}
			switch v := b.(type) {
			case int: vb = int64(v)
			case int64: vb = v
			case float64: vb = int64(v)
			}
			return va - vb
		},

		"mul": func(a, b interface{}) float64 {
			var va, vb float64
			switch v := a.(type) {
			case int: va = float64(v)
			case int64: va = float64(v)
			case float64: va = v
			}
			switch v := b.(type) {
			case int: vb = float64(v)
			case int64: vb = float64(v)
			case float64: vb = v
			}
			return va * vb
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
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02T15:04:05")
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
		"addf64": func(a, b interface{}) float64 {
			var va, vb float64
			switch v := a.(type) {
			case int: va = float64(v)
			case int64: va = float64(v)
			case float64: va = v
			}
			switch v := b.(type) {
			case int: vb = float64(v)
			case int64: vb = float64(v)
			case float64: vb = v
			}
			return va + vb
		},
		"add64": func(a, b interface{}) int64 {
			var va, vb int64
			switch v := a.(type) {
			case int: va = int64(v)
			case int64: va = v
			case float64: va = int64(v)
			}
			switch v := b.(type) {
			case int: vb = int64(v)
			case int64: vb = v
			case float64: vb = int64(v)
			}
			return va + vb
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
		"abs": func(a int64) int64 {
			if a < 0 {
				return -a
			}
			return a
		},
		"div": func(a, b interface{}) float64 {
			var va, vb float64
			switch v := a.(type) {
			case int: va = float64(v)
			case int64: va = float64(v)
			case float64: va = v
			}
			switch v := b.(type) {
			case int: vb = float64(v)
			case int64: vb = float64(v)
			case float64: vb = v
			}
			if vb == 0 {
				return 0.0 // Avoid division by zero
			}
			return va / vb
		},
		"percent": func(a, b int64) float64 {
			return percent(float64(a), float64(b))
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
		"list": func(values ...interface{}) []interface{} {
			return values
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

	partitionsTemplate, err := template.New("partitions.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/partitions.tmpl", "html/nav.tmpl", "html/modal.tmpl")
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

	logsTemplate, err := template.New("logs.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/logs.tmpl", "html/nav.tmpl", "html/modal.tmpl")
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

	consensusTemplate, err := template.New("consensus.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/consensus.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse consensus template: %w", err)
	}

	segmentsTemplate, err := template.New("segments.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/segments.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse segments template: %w", err)
	}

	segmentViewTemplate, err := template.New("segment_view.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/base.tmpl", "html/nav.tmpl", "html/segment_view.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse segment_view template: %w", err)
	}

	cpuProfilesTemplate, err := template.New("cpu_profiles.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/cpu_profiles.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu_profiles template: %w", err)
	}

	notebookTemplate, err := template.New("notebook.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/notebook.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse notebook template: %w", err)
	}

	server := &Server{
		sessions:            make(map[string]*BundleSession),
		activePath:          bundlePath,
		progress:            progress,
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
		consensusTemplate:   consensusTemplate,
		segmentsTemplate:    segmentsTemplate,
		segmentViewTemplate: segmentViewTemplate,
		cpuProfilesTemplate: cpuProfilesTemplate,
		notebookTemplate:    notebookTemplate,
		logsOnly:            logsOnly,
		authToken:           authToken,
		shutdownTimeout:     shutdownTimeout,
		inactivityTimeout:   inactivityTimeout,
		lastHeartbeat:       time.Now().Unix(),
		currentVersion:      currentVersion,
		latestVersion:       latestVersion,
		updateAvailable:     latestVersion != "" && latestVersion != currentVersion,
	}

	if server.progress == nil {
		server.progress = models.NewProgressTracker(22)
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
	mux.HandleFunc("/api/logs/pin", server.apiPinLogHandler)
	mux.HandleFunc("/api/logs/unpin", server.apiUnpinLogHandler)
	mux.HandleFunc("/api/raft/timeline", server.apiRaftTimelineHandler)
	mux.HandleFunc("/api/heartbeat", server.heartbeatHandler)
	mux.HandleFunc("/api/full-logs-page", server.apiFullLogsPageHandler) // New handler for lazy loading logs page
	mux.HandleFunc("/notebook", server.notebookHandler)
	mux.HandleFunc("/api/notebook/note", server.apiUpdatePinNoteHandler)
	mux.HandleFunc("/api/notebook/export", server.apiNotebookExportHandler)
	mux.HandleFunc("/api/metrics/list", server.handleMetricsList)
	mux.HandleFunc("/api/metrics/data", server.handleMetricsData)
	mux.HandleFunc("/api/full-metrics", server.apiFullMetricsHandler) // New handler for lazy loading metrics page
	mux.HandleFunc("/api/partition-leaders", server.handlePartitionLeadersAPI)
	mux.HandleFunc("/api/timeline/data", server.apiTimelineDataHandler)
	mux.HandleFunc("/api/timeline/aggregated", server.apiTimelineAggregatedHandler)
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/system", server.systemHandler)
	mux.HandleFunc("/consensus", server.consensusHandler)
	mux.HandleFunc("/timeline", server.timelineHandler)
	mux.HandleFunc("/skew", server.skewHandler)
	mux.HandleFunc("/search", server.searchHandler)
	mux.HandleFunc("/diagnostics", server.diagnosticsHandler)
	mux.HandleFunc("/segments", server.storageHandler) // Renamed from /storage
	// mux.HandleFunc("/segments", server.segmentsHandler) // Removed undefined handler
	mux.HandleFunc("/segments/view", server.segmentViewHandler)
	mux.HandleFunc("/cpu", server.cpuProfilesHandler)
	mux.HandleFunc("/api/cpu/profiles", server.apiCpuProfilesHandler)
	mux.HandleFunc("/api/cpu/details", server.apiCpuProfileDetailsHandler)
	mux.HandleFunc("/api/cpu/binary-status", server.apiCpuBinaryStatusHandler)
	mux.HandleFunc("/api/cpu/download", server.apiCpuDownloadHandler)

	// Wrap the mux with middlewares
	// Order: Auth -> Gzip -> Security -> Mux
	var handler http.Handler = mux
	handler = SecurityHeadersMiddleware(handler)
	handler = GzipMiddleware(handler)
	handler = server.AuthMiddleware(handler)

	server.handler = handler

	server.startTimeoutWatcher()

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

