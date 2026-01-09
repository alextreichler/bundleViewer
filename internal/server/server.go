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
	cpuProfilesTemplate *template.Template
	logsOnly            bool
}

func (s *Server) SetDataDir(path string) {
	s.dataDir = path
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func New(bundlePath string, cachedData *cache.CachedData, s store.Store, logger *slog.Logger, logsOnly bool, persist bool, progress *models.ProgressTracker) (http.Handler, error) { // Accept both
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

	cpuProfilesTemplate, err := template.New("cpu_profiles.tmpl").Funcs(funcMap).ParseFS(ui.HTMLFiles, "html/cpu_profiles.tmpl", "html/nav.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu_profiles template: %w", err)
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
		segmentsTemplate:    segmentsTemplate,
		segmentViewTemplate: segmentViewTemplate,
		cpuProfilesTemplate: cpuProfilesTemplate,
		logsOnly:            logsOnly,
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
	mux.HandleFunc("/cpu", server.cpuProfilesHandler)
	mux.HandleFunc("/api/cpu/profiles", server.apiCpuProfilesHandler)
	mux.HandleFunc("/api/cpu/details", server.apiCpuProfileDetailsHandler)
	mux.HandleFunc("/api/cpu/binary-status", server.apiCpuBinaryStatusHandler)
	mux.HandleFunc("/api/cpu/download", server.apiCpuDownloadHandler)

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

