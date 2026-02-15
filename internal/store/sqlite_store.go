package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alextreichler/bundleViewer/internal/logutil"
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

	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=10000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Performance Optimizations
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",      // Better concurrency
		"PRAGMA synchronous=NORMAL;",    // Good balance of safety and speed
		"PRAGMA locking_mode=EXCLUSIVE;", // Exclusive access for speed
		"PRAGMA page_size=65536;",       // 64KB page size
		"PRAGMA temp_store=MEMORY;",     // Temp tables in RAM
		"PRAGMA cache_size=-500000;",    // ~500MB cache
		"PRAGMA mmap_size=30000000000;", // Memory map up to 30GB (if available)
		"PRAGMA count_changes=OFF;",     // Don't count changes
		"PRAGMA threads=4;",             // Parallel query execution
		"PRAGMA secure_delete=OFF;",     // Faster deletes
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			slog.Warn("Failed to set pragma", "pragma", p, "error", err)
		}
	}

	// Allow multiple readers in WAL mode
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

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
	if err != nil || version < 9 {
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
		message_lower,
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
		  INSERT INTO logs_fts(rowid, message_lower) VALUES (new.id, new.message_lower);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS logs_ad AFTER DELETE ON logs BEGIN
		  INSERT INTO logs_fts(logs_fts, rowid, message_lower) VALUES('delete', old.id, old.message_lower);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS logs_au AFTER UPDATE ON logs BEGIN
		  INSERT INTO logs_fts(logs_fts, rowid, message_lower) VALUES('delete', old.id, old.message_lower);
		  INSERT INTO logs_fts(rowid, message_lower) VALUES (new.id, new.message_lower);
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

	// Version 6 -> 7: Add CPU Profiles
	if currentVersion < 7 {
		if _, err := tx.Exec(CpuProfileSchema); err != nil {
			return fmt.Errorf("failed to create cpu_profiles table: %w", err)
		}
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (7)"); err != nil {
			return fmt.Errorf("failed to update schema version to 7: %w", err)
		}
		currentVersion = 7
	}

	// Version 7 -> 8: Add Pins table
	if currentVersion < 8 {
		if _, err := tx.Exec(PinsSchema); err != nil {
			return fmt.Errorf("failed to create pins table: %w", err)
		}
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (8)"); err != nil {
			return fmt.Errorf("failed to update schema version to 8: %w", err)
		}
		currentVersion = 8
	}

	// Version 8 -> 9: Case-insensitive search (handled by shouldWipeDB forcing fresh schema)
	if currentVersion < 9 {
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (9)"); err != nil {
			return fmt.Errorf("failed to update schema version to 9: %w", err)
		}
		currentVersion = 9
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
	logEntry.Fingerprint = logutil.GenerateFingerprint(logEntry.Message)

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
		logEntry.Fingerprint = logutil.GenerateFingerprint(logEntry.Message)
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
	var labelsJSON []byte
	var err error

	if metric.LabelsJSON != "" {
		labelsJSON = []byte(metric.LabelsJSON)
	} else {
		labelsJSON, err = json.Marshal(metric.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels to JSON: %w", err)
		}
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
		var labelsJSON []byte
		var err error

		if metric.LabelsJSON != "" {
			labelsJSON = []byte(metric.LabelsJSON)
		} else {
			labelsJSON, err = json.Marshal(metric.Labels)
			if err != nil {
				return fmt.Errorf("failed to marshal labels to JSON during bulk operation: %w", err)
			}
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

	// Only these fields are actually in the FTS table
	allowedFields := map[string]bool{
		"message":       true,
		"message_lower": true,
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
	parenCount := 0
	for _, token := range tokens {
		upper := strings.ToUpper(token)
		if operators[upper] {
			if upper == "(" {
				parenCount++
				processed = append(processed, upper)
			} else if upper == ")" {
				if parenCount <= 0 {
					continue // Skip unbalanced closing paren
				}
				// Avoid empty parentheses ()
				if len(processed) > 0 && processed[len(processed)-1] == "(" {
					processed = processed[:len(processed)-1]
					parenCount--
					continue
				}
				parenCount--
				processed = append(processed, upper)
			} else {
				// Binary operators (AND, OR, NOT)
				// Avoid consecutive operators or operator after (
				if len(processed) > 0 {
					last := processed[len(processed)-1]
					if last == "(" || operators[last] {
						continue 
					}
				} else {
					// Allow starting with NOT so GetLogs can detect it,
					// but skip other binary operators like AND, OR.
					if upper != "NOT" && upper != "(" {
						continue
					}
				}
				processed = append(processed, upper)
			}
			continue
		}

		// Check for field:value syntax
		colonIdx := strings.Index(token, ":")
		if colonIdx > 0 {
			field := token[:colonIdx]
			value := token[colonIdx+1:]
			
			fLower := strings.ToLower(field)
			if allowedFields[fLower] {
				// Normalize to message_lower
				if fLower == "message" {
					fLower = "message_lower"
				}
				
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
					value = value[1 : len(value)-1]
				}
				
				safeValue := strings.ReplaceAll(value, "\"", "\"\"")
				processed = append(processed, fmt.Sprintf("%s:\"%s\"", fLower, safeValue))
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

	// Close unbalanced parentheses
	for parenCount > 0 {
		processed = append(processed, ")")
		parenCount--
	}

	// Remove leading binary operators (AND, OR)
	for len(processed) > 0 {
		first := processed[0]
		if first == "AND" || first == "OR" {
			processed = processed[1:]
		} else {
			break
		}
	}

	// Remove trailing operators
	for len(processed) > 0 {
		last := processed[len(processed)-1]
		if last == "AND" || last == "OR" || last == "NOT" || last == "(" {
			processed = processed[:len(processed)-1]
		} else {
			break
		}
	}

	return strings.Join(processed, " ")
}

func (s *SQLiteStore) GetRaftEvents(ntp string) ([]models.RaftTimelineEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use FTS to find logs related to this NTP and Raft patterns
	// Patterns: "became the leader", "Stepping down", "error code: 10", "raft::errc::timeout"
	searchQuery := fmt.Sprintf("\"%s\" AND (leader OR stepdown OR timeout OR term)", ntp)
	ftsQuery := buildFTSQuery(searchQuery)

	query := `
		SELECT T1.timestamp, T1.node, T1.message 
		FROM logs AS T1 
		JOIN logs_fts AS T2 ON T1.id = T2.rowid 
		WHERE T2.logs_fts MATCH ? 
		ORDER BY T1.timestamp ASC`

	rows, err := s.db.Query(query, ftsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.RaftTimelineEvent
	
	// Regex for term extraction and reason extraction
	termRegex := regexp.MustCompile(`term (\d+)`)
	reasonRegex := regexp.MustCompile(`(?:reasoning: |\[)([\w_]+)(?:\]\s+Stepping down|$)`)
	ntpRegex := regexp.MustCompile(`([\w._-]+/[\w._-]+/\d+)`)

	for rows.Next() {
		var ts int64
		var node, msg string
		if err := rows.Scan(&ts, &node, &msg); err != nil {
			continue
		}

		// Extract NTP if present
		var ntpFound string
		if matches := ntpRegex.FindStringSubmatch(msg); len(matches) > 1 {
			ntpFound = matches[1]
		} else {
			ntpFound = ntp // Use the provided ntp if not found in message
		}

		event := models.RaftTimelineEvent{
			Timestamp:   time.UnixMicro(ts).UTC(),
			Node:        node,
			NTP:         ntpFound,
			Description: msg,
		}

		// Extract term if present
		if matches := termRegex.FindStringSubmatch(msg); len(matches) > 1 {
			event.Term, _ = strconv.Atoi(matches[1])
		}

		// Categorize event
		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "became the leader") {
			event.Type = "Elected"
		} else if strings.Contains(lowerMsg, "stepping down") {
			event.Type = "Stepdown"
			if matches := reasonRegex.FindStringSubmatch(msg); len(matches) > 1 {
				event.Reason = matches[1]
			}
		} else if strings.Contains(lowerMsg, "timeout") || strings.Contains(lowerMsg, "error code: 10") {
			event.Type = "Timeout"
		} else {
			// Skip logs that don't match our core interest even if FTS picked them up
			continue
		}

		events = append(events, event)
	}

	return events, nil
}

func (s *SQLiteStore) GetGlobalRaftEvents() ([]models.RaftTimelineEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Broad search for Raft and configuration changes across all NTPs
	searchQuery := "(leader OR stepdown OR timeout OR term OR reconfiguration OR moving OR \"is down\")"
	ftsQuery := buildFTSQuery(searchQuery)

	query := `
		SELECT T1.timestamp, T1.node, T1.message 
		FROM logs AS T1 
		JOIN logs_fts AS T2 ON T1.id = T2.rowid 
		WHERE T2.logs_fts MATCH ? 
		ORDER BY T1.timestamp DESC`

	rows, err := s.db.Query(query, ftsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.RaftTimelineEvent
	termRegex := regexp.MustCompile(`term (\d+)`)
	reasonRegex := regexp.MustCompile(`(?:reasoning: |\[)([\w_]+)(?:\]\s+Stepping down|$)`)
	ntpRegex := regexp.MustCompile(`([\w._-]+/[\w._-]+/\d+)`)

	for rows.Next() {
		var ts int64
		var node, msg string
		if err := rows.Scan(&ts, &node, &msg); err != nil {
			continue
		}

		event := models.RaftTimelineEvent{
			Timestamp:   time.UnixMicro(ts).UTC(),
			Node:        node,
			Description: msg,
		}

		if matches := termRegex.FindStringSubmatch(msg); len(matches) > 1 {
			event.Term, _ = strconv.Atoi(matches[1])
		}
		if matches := ntpRegex.FindStringSubmatch(msg); len(matches) > 1 {
			event.NTP = matches[1]
		}

		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "became the leader") {
			event.Type = "Elected"
		} else if strings.Contains(lowerMsg, "stepping down") {
			event.Type = "Stepdown"
			if matches := reasonRegex.FindStringSubmatch(msg); len(matches) > 1 {
				event.Reason = matches[1]
			}
		} else if strings.Contains(lowerMsg, "timeout") || strings.Contains(lowerMsg, "error code: 10") {
			event.Type = "Timeout"
		} else if strings.Contains(lowerMsg, "moving") || strings.Contains(lowerMsg, "reconfiguration") {
			event.Type = "Move"
		} else if strings.Contains(lowerMsg, "is down") {
			event.Type = "Health"
		} else {
			continue
		}

		events = append(events, event)
	}

	return events, nil
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
	ftsSearch := buildFTSQuery(strings.ToLower(filter.Search))
	ftsIgnore := buildFTSQuery(strings.ToLower(filter.Ignore))

	// If the ignore query starts with "NOT", strip it. It's an ignore filter already.
	if strings.HasPrefix(ftsIgnore, "NOT ") {
		ftsIgnore = strings.TrimPrefix(ftsIgnore, "NOT ")
	}

	// If the search query starts with "NOT", merge it into the ignore filter.
	// FTS5 doesn't support a leading NOT in a MATCH expression.
	if strings.HasPrefix(ftsSearch, "NOT ") {
		ignorePart := strings.TrimPrefix(ftsSearch, "NOT ")
		if ftsIgnore != "" {
			ftsIgnore = fmt.Sprintf("(%s) OR (%s)", ignorePart, ftsIgnore)
		} else {
			ftsIgnore = ignorePart
		}
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
		queryBuilder.WriteString("SELECT T1.id, T1.timestamp, T1.level, T1.node, T1.shard, T1.component, T1.message, T1.raw, T1.line_number, T1.file_path, snippet(logs_fts, 0, '<mark>', '</mark>', '...', 64) as snippet, (P.log_id IS NOT NULL) as is_pinned FROM logs AS T1 JOIN logs_fts AS T2 ON T1.id = T2.rowid LEFT JOIN pins AS P ON T1.id = P.log_id")
		countBuilder.WriteString("SELECT COUNT(T1.id) FROM logs AS T1 JOIN logs_fts AS T2 ON T1.id = T2.rowid")

		whereClauses = append(whereClauses, "T2.logs_fts MATCH ?")

		args = append(args, finalFTSQuery)
		countArgs = append(countArgs, finalFTSQuery)
	} else {
		queryBuilder.WriteString("SELECT L.id, L.timestamp, L.level, L.node, L.shard, L.component, L.message, L.raw, L.line_number, L.file_path, '' as snippet, (P.log_id IS NOT NULL) as is_pinned FROM logs AS L LEFT JOIN pins AS P ON L.id = P.log_id")
		countBuilder.WriteString("SELECT COUNT(*) FROM logs AS L")

		if ftsIgnore != "" {
			whereClauses = append(whereClauses, "L.id NOT IN (SELECT rowid FROM logs_fts WHERE logs_fts MATCH ?)")
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
			whereClauses = append(whereClauses, fmt.Sprintf("L.level IN (%s)", strings.Join(placeholders, ",")))
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
			whereClauses = append(whereClauses, fmt.Sprintf("L.node IN (%s)", strings.Join(placeholders, ",")))
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
			whereClauses = append(whereClauses, fmt.Sprintf("L.component IN (%s)", strings.Join(placeholders, ",")))
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
				whereClauses = append(whereClauses, "L.timestamp >= ?")
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
				whereClauses = append(whereClauses, "L.timestamp <= ?")
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
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY L.timestamp %s", sortDir))
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
		err = rows.Scan(&logEntry.ID, &ts, &logEntry.Level, &logEntry.Node, &logEntry.Shard, &logEntry.Component, &logEntry.Message, &logEntry.Raw, &logEntry.LineNumber, &logEntry.FilePath, &logEntry.Snippet, &logEntry.IsPinned)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan log entry: %w", err)
		}
		if ts != 0 {
			logEntry.Timestamp = time.UnixMicro(ts).UTC()
		}
		logEntry.Insight = logutil.GetInsight(logEntry.Message)
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
		p.Insight = logutil.GetInsight(sampleMessage)
		patterns = append(patterns, p)
		totalLogs += p.Count
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during rows iteration for patterns: %w", err)
	}

	return patterns, totalLogs, nil
}

func (s *SQLiteStore) Optimize(p *models.ProgressTracker) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("Running database optimizations...")
	if p != nil {
		p.SetStatus("Optimizing FTS5 index (merging segments)...")
	}

	// 1. Optimize FTS5 index (merges index segments)
	if _, err := s.db.Exec("INSERT INTO logs_fts(logs_fts) VALUES('optimize')"); err != nil {
		slog.Warn("Failed to optimize FTS5 index", "error", err)
	}

	if p != nil {
		p.SetStatus("Analyzing database statistics...")
	}

	// 2. Run ANALYZE to update query planner statistics
	if _, err := s.db.Exec("ANALYZE"); err != nil {
		slog.Warn("Failed to run ANALYZE", "error", err)
	}

	if p != nil {
		p.SetStatus("Vacuuming database (reclaiming space)...")
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

func (s *SQLiteStore) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Query(query, args...)
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
	
	func (s *SQLiteStore) BulkInsertCpuProfiles(profiles []models.CpuProfileEntry) error {
		s.mu.Lock()
		defer s.mu.Unlock()
	
		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
	
		stmt, err := tx.Prepare("INSERT INTO cpu_profiles (node, shard_id, scheduling_group, user_backtrace, occurrences) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer func() { _ = stmt.Close() }()
	
		for _, p := range profiles {
			if _, err := stmt.Exec(p.Node, p.ShardID, p.SchedulingGroup, p.UserBacktrace, p.Occurrences); err != nil {
				return fmt.Errorf("failed to insert cpu profile: %w", err)
			}
		}
	
		return tx.Commit()
	}
	
	func (s *SQLiteStore) GetCpuProfiles() ([]models.CpuProfileAggregate, error) {
		s.mu.RLock()
		defer s.mu.RUnlock()
	
		// Check if table exists (in case migration failed or running on old DB without clean)
		var exists int
		err := s.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='cpu_profiles'").Scan(&exists)
		if err != nil || exists == 0 {
			return []models.CpuProfileAggregate{}, nil
		}
	
		query := `
			SELECT node, shard_id, scheduling_group, SUM(occurrences) as total
			FROM cpu_profiles
			GROUP BY node, shard_id, scheduling_group
			ORDER BY node, shard_id, total DESC
		`
	
		rows, err := s.db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query cpu profiles: %w", err)
		}
			defer func() { _ = rows.Close() }()
		
			rawResults := make([]models.CpuProfileAggregate, 0)
			
			type key struct {
		
			node  string
			shard int
		}
		shardTotals := make(map[key]int)
	
		for rows.Next() {
			var r models.CpuProfileAggregate
			if err := rows.Scan(&r.Node, &r.ShardID, &r.SchedulingGroup, &r.TotalSamples); err != nil {
				return nil, err
			}
			rawResults = append(rawResults, r)
			k := key{r.Node, r.ShardID}
			shardTotals[k] += r.TotalSamples
		}
	
		for i := range rawResults {
			k := key{rawResults[i].Node, rawResults[i].ShardID}
			if total := shardTotals[k]; total > 0 {
				rawResults[i].Percentage = (float64(rawResults[i].TotalSamples) / float64(total)) * 100
			}
		}
	
			return rawResults, nil
		}
		
		func (s *SQLiteStore) GetCpuProfileDetails(node string, shardID int, group string) ([]models.CpuProfileDetail, error) {
			s.mu.RLock()
			defer s.mu.RUnlock()
		
			// 1. Get total samples for this group/shard to calculate percentage
			var total int
			err := s.db.QueryRow("SELECT SUM(occurrences) FROM cpu_profiles WHERE node = ? AND shard_id = ? AND scheduling_group = ?", node, shardID, group).Scan(&total)
			if err != nil {
				return nil, fmt.Errorf("failed to get total samples: %w", err)
			}
		
			if total == 0 {
				return []models.CpuProfileDetail{}, nil
			}
		
			// 2. Get top backtraces
			query := `
				SELECT user_backtrace, SUM(occurrences) as occ
				FROM cpu_profiles
				WHERE node = ? AND shard_id = ? AND scheduling_group = ?
				GROUP BY user_backtrace
				ORDER BY occ DESC
				LIMIT 50
			`
			rows, err := s.db.Query(query, node, shardID, group)
			if err != nil {
				return nil, fmt.Errorf("failed to query profile details: %w", err)
			}
			defer func() { _ = rows.Close() }()
		
			var details []models.CpuProfileDetail
			for rows.Next() {
				var d models.CpuProfileDetail
				if err := rows.Scan(&d.UserBacktrace, &d.Occurrences); err != nil {
					return nil, err
				}
				d.Percentage = (float64(d.Occurrences) / float64(total)) * 100
							details = append(details, d)
						}
					
						return details, nil
					}
				
				func (s *SQLiteStore) PinLog(logID int64) error {
					s.mu.Lock()
					defer s.mu.Unlock()
				
					_, err := s.db.Exec(`
						INSERT INTO pins (log_id, created_at)
						VALUES (?, ?)
						ON CONFLICT(log_id) DO NOTHING
					`, logID, time.Now().UnixMicro())
					return err
				}
				
				func (s *SQLiteStore) UnpinLog(logID int64) error {
					s.mu.Lock()
					defer s.mu.Unlock()
				
					_, err := s.db.Exec("DELETE FROM pins WHERE log_id = ?", logID)
					return err
				}
				
				func (s *SQLiteStore) UpdatePinNote(logID int64, note string) error {
					s.mu.Lock()
					defer s.mu.Unlock()
				
					_, err := s.db.Exec("UPDATE pins SET note = ? WHERE log_id = ?", note, logID)
					return err
				}
				
				func (s *SQLiteStore) GetPinnedLogs() ([]*models.PinnedLog, error) {
					s.mu.RLock()
					defer s.mu.RUnlock()
				
					query := `
						SELECT L.id, L.timestamp, L.level, L.node, L.shard, L.component, L.message, L.raw, L.line_number, L.file_path, P.note
						FROM logs L
						JOIN pins P ON L.id = P.log_id
						ORDER BY P.created_at ASC
					`
				
					rows, err := s.db.Query(query)
					if err != nil {
						return nil, fmt.Errorf("failed to query pinned logs: %w", err)
					}
					defer func() { _ = rows.Close() }()
				
					var pins []*models.PinnedLog
					for rows.Next() {
						var p models.PinnedLog
						var ts int64
						err := rows.Scan(&p.ID, &ts, &p.Level, &p.Node, &p.Shard, &p.Component, &p.Message, &p.Raw, &p.LineNumber, &p.FilePath, &p.Note)
						if err != nil {
							return nil, err
						}
						if ts != 0 {
							p.Timestamp = time.UnixMicro(ts).UTC()
						}
						p.IsPinned = true
						pins = append(pins, &p)
					}
					return pins, rows.Err()
				}
				
