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

const CpuProfileSchema = `
CREATE TABLE IF NOT EXISTS cpu_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node TEXT NOT NULL,
    shard_id INTEGER NOT NULL,
    scheduling_group TEXT NOT NULL,
    user_backtrace TEXT NOT NULL,
    occurrences INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cpu_profiles_node_shard ON cpu_profiles(node, shard_id);
CREATE INDEX IF NOT EXISTS idx_cpu_profiles_group ON cpu_profiles(scheduling_group);
`
