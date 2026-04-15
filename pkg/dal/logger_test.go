package dal

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

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
		_ = db.Close()
		t.Fatal(err)
	}
	return db, func() { _ = db.Close() }
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
	_ = rows.Close()

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
	defer func() { _ = tx.Rollback() }()

	var _ DBExecutor = tx
}

func TestTxQueryRow(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, _ = db.Exec("INSERT INTO test (id, name) VALUES (?, ?)", 1, "foo")

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	tx, err := bdb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	var name string
	err = tx.QueryRow(context.Background(), "SELECT name FROM test WHERE id = ?", 1).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}

	if e := log.find("debug", "tx query_row"); e == nil {
		t.Error("expected 'tx query_row' log entry")
	}
}

func TestTxRollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	tx, err := bdb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	if e := log.find("debug", "tx rollback"); e == nil {
		t.Error("expected 'tx rollback' log entry")
	}
}

func TestWithTxCommit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	err := bdb.WithTx(context.Background(), nil, func(tx *Tx) error {
		_, err := tx.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 1, "withtx_test")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if e := log.find("debug", "begin_tx"); e == nil {
		t.Error("expected 'begin_tx' log entry")
	}
	if e := log.find("debug", "tx commit"); e == nil {
		t.Error("expected 'tx commit' log entry")
	}

	var name string
	err = bdb.QueryRow(context.Background(), "SELECT name FROM test WHERE id = ?", 1).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}
	if name != "withtx_test" {
		t.Errorf("got %q, want 'withtx_test'", name)
	}
}

func TestWithTxRollbackOnError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)

	err := bdb.WithTx(context.Background(), nil, func(tx *Tx) error {
		return fmt.Errorf("intentional error")
	})
	if err == nil {
		t.Fatal("expected error from WithTx")
	}

	if e := log.find("debug", "tx rollback"); e == nil {
		t.Error("expected 'tx rollback' log entry on error")
	}
}

func TestSetLogArgsEnabled(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)
	bdb.SetLogArgs(true)

	_, _ = bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 42, "hello")

	e := log.find("debug", "query exec")
	if e == nil {
		t.Fatal("expected 'query exec' log entry")
	}
	for i := 0; i < len(e.kv)-1; i += 2 {
		if e.kv[i] == "args" {
			args, ok := e.kv[i+1].([]interface{})
			if !ok {
				t.Fatalf("expected []interface{}, got %T", e.kv[i+1])
			}
			if len(args) != 2 || args[0] != 42 || args[1] != "hello" {
				t.Errorf("got args %v, want [42 hello]", args)
			}
			return
		}
	}
	t.Error("expected 'args' key in log entry")
}

func TestSetLogArgsRedacted(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := &mockLogger{}
	bdb := NewBaseDB(db, log)
	bdb.SetLogArgs(false)

	_, _ = bdb.Exec(context.Background(), "INSERT INTO test (id, name) VALUES (?, ?)", 42, "hello")

	e := log.find("debug", "query exec")
	if e == nil {
		t.Fatal("expected 'query exec' log entry")
	}
	for i := 0; i < len(e.kv)-1; i += 2 {
		if e.kv[i] == "args" {
			if e.kv[i+1] != "<redacted>" {
				t.Errorf("got %v, want '<redacted>'", e.kv[i+1])
			}
			return
		}
	}
	t.Error("expected 'args' key in log entry")
}

func TestDBExecutorInterface(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	bdb := NewBaseDB(db, nil)

	var _ DBExecutor = bdb

	tx, err := bdb.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	var _ DBExecutor = tx
}
