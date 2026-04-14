package dal

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

type logEntry struct {
	level string
	msg   string
	kv    []any
}

type mockLogger struct {
	entries []logEntry
}

func (m *mockLogger) Trace(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"trace", msg, kv})
}
func (m *mockLogger) Debug(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"debug", msg, kv})
}
func (m *mockLogger) Info(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"info", msg, kv})
}
func (m *mockLogger) Warn(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"warn", msg, kv})
}
func (m *mockLogger) Error(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"error", msg, kv})
}
func (m *mockLogger) Fatal(msg string, kv ...any) {
	m.entries = append(m.entries, logEntry{"fatal", msg, kv})
}

func (m *mockLogger) find(level, msg string) *logEntry {
	for _, e := range m.entries {
		if e.level == level && e.msg == msg {
			return &e
		}
	}
	return nil
}

func (m *mockLogger) hasKey(msg string, key string) bool {
	for _, e := range m.entries {
		if e.msg == msg {
			for i := 0; i < len(e.kv)-1; i += 2 {
				if e.kv[i] == key {
					return true
				}
			}
		}
	}
	return false
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db, func() { db.Close() }
}

func TestBaseDBExecLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	_, err := bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if e := log.find("debug", "query exec"); e == nil {
		t.Error("expected debug 'query exec' log entry")
	}
	if e := log.find("debug", "query exec done"); e == nil {
		t.Error("expected debug 'query exec done' log entry")
	}
	if !log.hasKey("query exec done", "duration") {
		t.Error("expected 'duration' key in 'query exec done'")
	}
}

func TestBaseDBQueryLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, _ = db.Exec("INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	rows, err := bdb.Query(context.Background(), "SELECT id, name FROM test WHERE id = ?", 1)
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()

	if e := log.find("debug", "query"); e == nil {
		t.Error("expected debug 'query' log entry")
	}
	if e := log.find("debug", "query done"); e == nil {
		t.Error("expected debug 'query done' log entry")
	}
}

func TestBaseDBQueryRowLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, _ = db.Exec("INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	var name string
	err := bdb.QueryRow(context.Background(), "SELECT name FROM test WHERE id = ?", 1).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}

	if e := log.find("debug", "query_row"); e == nil {
		t.Error("expected debug 'query_row' log entry")
	}
}

func TestBaseDBBeginTxLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	tx, err := bdb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	tx.Rollback()

	if e := log.find("debug", "begin_tx"); e == nil {
		t.Error("expected debug 'begin_tx' log entry")
	}
}

func TestBaseDBCloseLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	bdb.Close()
	cleanup = func() {}

	if e := log.find("debug", "close"); e == nil {
		t.Error("expected debug 'close' log entry")
	}
}

func TestBaseDBExecErrorLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	_, err := bdb.Exec(context.Background(), "INSERT INTO nonexistent (id) VALUES (?)", 1)
	if err == nil {
		t.Fatal("expected error from invalid SQL")
	}

	if e := log.find("error", "query exec error"); e == nil {
		t.Error("expected error 'query exec error' log entry")
	}
	if !log.hasKey("query exec error", "error") {
		t.Error("expected 'error' key in error log entry")
	}
}

func TestBaseDBQueryErrorLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	rows, err := bdb.Query(context.Background(), "SELECT * FROM nonexistent")
	if err == nil {
		rows.Close()
		t.Fatal("expected error from invalid SQL")
	}

	if e := log.find("error", "query error"); e == nil {
		t.Error("expected error 'query error' log entry")
	}
}

func TestNoopLoggerDoesNotPanic(t *testing.T) {
	log := NoopLoggerInstance()
	log.Trace("test", "key", "value")
	log.Debug("test", "key", "value")
	log.Info("test", "key", "value")
	log.Warn("test", "key", "value")
	log.Error("test", "key", "value")
	log.Fatal("test", "key", "value")
}

func TestNilLoggerDefaultsToNoop(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bdb := NewBaseDB(db, nil)
	_, err := bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetLogger(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bdb := NewBaseDB(db, nil)
	log := &mockLogger{}
	bdb.SetLogger(log)

	_, err := bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if len(log.entries) == 0 {
		t.Error("expected log entries after SetLogger")
	}
}

func TestSetLoggerNil(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bdb := NewBaseDB(db, &mockLogger{})
	bdb.SetLogger(nil)

	_, err := bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecLogsArgs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	_, _ = bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 42, "hello")

	e := log.find("debug", "query exec")
	if e == nil {
		t.Fatal("expected 'query exec' log entry")
	}
	found := false
	for i := 0; i < len(e.kv)-1; i += 2 {
		if e.kv[i] == "args" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'args' key in log entry")
	}
}

func TestDurationIsLogged(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	_, _ = bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")

	e := log.find("debug", "query exec done")
	if e == nil {
		t.Fatal("expected 'query exec done' log entry")
	}
	for i := 0; i < len(e.kv)-1; i += 2 {
		if e.kv[i] == "duration" {
			d, ok := e.kv[i+1].(time.Duration)
			if !ok {
				t.Errorf("expected duration to be time.Duration, got %T", e.kv[i+1])
			}
			if d < 0 {
				t.Error("expected non-negative duration")
			}
			return
		}
	}
	t.Error("expected 'duration' key in 'query exec done'")
}

func TestDBMethodReturnsUnderlyingDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bdb := NewBaseDB(db, nil)
	if bdb.DB() != db {
		t.Error("DB() should return the underlying *sql.DB")
	}
}
