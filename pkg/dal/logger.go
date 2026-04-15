package dal

import (
	"context"
	"database/sql"
	"time"
)

// Logger defines a structured logging interface with six log levels.
// It is compatible with github.com/fortix/go-libs/logger.
type Logger interface {
	Trace(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)
}

// NoopLogger is a no-op Logger that discards all log output.
type NoopLogger struct{}

func (n NoopLogger) Trace(msg string, keysAndValues ...any) {}
func (n NoopLogger) Debug(msg string, keysAndValues ...any) {}
func (n NoopLogger) Info(msg string, keysAndValues ...any)  {}
func (n NoopLogger) Warn(msg string, keysAndValues ...any)  {}
func (n NoopLogger) Error(msg string, keysAndValues ...any) {}
func (n NoopLogger) Fatal(msg string, keysAndValues ...any) {}

var noopLogger = NoopLogger{}

// NoopLoggerInstance returns a shared NoopLogger instance.
func NoopLoggerInstance() Logger {
	return noopLogger
}

// DBExecutor is a common interface for Exec, Query, and QueryRow,
// satisfied by both BaseDB and Tx, enabling functions that accept either.
type DBExecutor interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// BaseDB wraps a *sql.DB with structured query logging.
// All database operations are logged at Debug level with query text, arguments,
// and duration. Errors are logged at Error level.
type BaseDB struct {
	db      *sql.DB
	log     Logger
	logArgs bool
}

// NewBaseDB creates a BaseDB wrapping the given *sql.DB. If log is nil, logging is disabled.
func NewBaseDB(db *sql.DB, log Logger) *BaseDB {
	if log == nil {
		log = noopLogger
	}
	return &BaseDB{db: db, log: log}
}

// SetLogger replaces the current logger. Pass nil to disable logging.
func (b *BaseDB) SetLogger(log Logger) {
	if log == nil {
		log = noopLogger
	}
	b.log = log
}

// SetLogArgs controls whether query arguments are included in log output.
// Defaults to false (args are redacted). Set to true to log actual argument values.
func (b *BaseDB) SetLogArgs(enabled bool) {
	b.logArgs = enabled
}

func (b *BaseDB) logArgsValue(args []interface{}) interface{} {
	if b.logArgs {
		return args
	}
	return "<redacted>"
}

// Exec executes a query without returning any rows, with logging.
func (b *BaseDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	b.log.Debug("query exec", "query", query, "args", b.logArgsValue(args))
	result, err := b.db.ExecContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.log.Error("query exec error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.log.Debug("query exec done", "query", query, "duration", elapsed)
	return result, nil
}

// Query executes a query that returns rows, with logging.
func (b *BaseDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	b.log.Debug("query", "query", query, "args", b.logArgsValue(args))
	rows, err := b.db.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.log.Error("query error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.log.Debug("query done", "query", query, "duration", elapsed)
	return rows, nil
}

// QueryRow executes a query expected to return at most one row, with logging.
// Note: duration and errors are not logged because *sql.Row defers execution
// until Scan() is called on the returned Row.
func (b *BaseDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	b.log.Debug("query_row", "query", query, "args", b.logArgsValue(args))
	return b.db.QueryRowContext(ctx, query, args...)
}

// Close closes the underlying database connection, with logging.
func (b *BaseDB) Close() error {
	b.log.Debug("close")
	return b.db.Close()
}

// Ping verifies the database connection is alive, with logging.
func (b *BaseDB) Ping(ctx context.Context) error {
	b.log.Debug("ping")
	return b.db.PingContext(ctx)
}

// DB returns the underlying *sql.DB for advanced use cases
// (connection pool configuration, raw queries, etc.).
func (b *BaseDB) DB() *sql.DB {
	return b.db
}

// WithTx executes fn within a transaction. If fn returns an error, the transaction
// is rolled back. If fn returns nil, the transaction is committed.
// This ensures proper cleanup even on panic.
func (b *BaseDB) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(*Tx) error) error {
	tx, err := b.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// Tx wraps a *sql.Tx with structured query logging, mirroring BaseDB's interface.
type Tx struct {
	tx      *sql.Tx
	log     Logger
	logArgs bool
}

// BeginTx starts a transaction and returns a Tx wrapper with logging.
func (b *BaseDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	b.log.Debug("begin_tx")
	tx, err := b.db.BeginTx(ctx, opts)
	if err != nil {
		b.log.Error("begin_tx error", "error", err)
		return nil, err
	}
	return &Tx{tx: tx, log: b.log, logArgs: b.logArgs}, nil
}

func (t *Tx) logArgsValue(args []interface{}) interface{} {
	if t.logArgs {
		return args
	}
	return "<redacted>"
}

// Exec executes a query within the transaction, with logging.
func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	t.log.Debug("tx exec", "query", query, "args", t.logArgsValue(args))
	result, err := t.tx.ExecContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		t.log.Error("tx exec error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	t.log.Debug("tx exec done", "query", query, "duration", elapsed)
	return result, nil
}

// Query executes a query within the transaction that returns rows, with logging.
func (t *Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	t.log.Debug("tx query", "query", query, "args", t.logArgsValue(args))
	rows, err := t.tx.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		t.log.Error("tx query error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	t.log.Debug("tx query done", "query", query, "duration", elapsed)
	return rows, nil
}

// QueryRow executes a query within the transaction expected to return at most one row.
// Note: duration and errors are not logged because *sql.Row defers execution until Scan().
func (t *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	t.log.Debug("tx query_row", "query", query, "args", t.logArgsValue(args))
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Commit commits the transaction, with logging.
func (t *Tx) Commit() error {
	t.log.Debug("tx commit")
	err := t.tx.Commit()
	if err != nil {
		t.log.Error("tx commit error", "error", err)
	}
	return err
}

// Rollback rolls back the transaction, with logging.
func (t *Tx) Rollback() error {
	t.log.Debug("tx rollback")
	err := t.tx.Rollback()
	if err != nil {
		t.log.Error("tx rollback error", "error", err)
	}
	return err
}
