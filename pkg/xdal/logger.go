package xdal

import (
	"context"
	"database/sql"
	"sync"
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
	dialect Dialect
	mu      sync.RWMutex
	log     Logger
	logArgs bool
}

// NewBaseDB creates a BaseDB wrapping the given *sql.DB. If log is nil, logging is disabled.
func NewBaseDB(db *sql.DB, dialect Dialect, log Logger) *BaseDB {
	if log == nil {
		log = noopLogger
	}
	return &BaseDB{db: db, dialect: dialect, log: log}
}

func (b *BaseDB) SetLogger(log Logger) {
	if log == nil {
		log = noopLogger
	}
	b.mu.Lock()
	b.log = log
	b.mu.Unlock()
}

func (b *BaseDB) SetLogArgs(enabled bool) {
	b.mu.Lock()
	b.logArgs = enabled
	b.mu.Unlock()
}

func (b *BaseDB) getLogger() Logger {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.log
}

func (b *BaseDB) getLogArgs() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.logArgs
}

func (b *BaseDB) logArgsValue(args []interface{}) interface{} {
	if b.getLogArgs() {
		return args
	}
	return "<redacted>"
}

func (b *BaseDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	b.getLogger().Debug("query exec", "query", query, "args", b.logArgsValue(args))
	result, err := b.db.ExecContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.getLogger().Error("query exec error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.getLogger().Debug("query exec done", "query", query, "duration", elapsed)
	return result, nil
}

func (b *BaseDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	b.getLogger().Debug("query", "query", query, "args", b.logArgsValue(args))
	rows, err := b.db.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.getLogger().Error("query error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.getLogger().Debug("query done", "query", query, "duration", elapsed)
	return rows, nil
}

func (b *BaseDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	b.getLogger().Debug("query_row", "query", query, "args", b.logArgsValue(args))
	return b.db.QueryRowContext(ctx, query, args...)
}

func (b *BaseDB) Close() error {
	b.getLogger().Debug("close")
	return b.db.Close()
}

func (b *BaseDB) Ping(ctx context.Context) error {
	b.getLogger().Debug("ping")
	return b.db.PingContext(ctx)
}

// DB returns the underlying *sql.DB for advanced use cases
// (connection pool configuration, raw queries, etc.).
func (b *BaseDB) DB() *sql.DB {
	return b.db
}

// Dialect returns the dialect used by this database connection.
func (b *BaseDB) Dialect() Dialect {
	return b.dialect
}

// Select starts a SELECT query pre-wired to this database connection.
// Call Query(ctx) or QueryRow(ctx) to execute directly, or Build() for the raw SQL.
func (b *BaseDB) Select(columns ...string) *SelectQuery {
	return &SelectQuery{columns: columns, dialect: b.dialect, db: b}
}

// Insert starts an INSERT query pre-wired to this database connection.
// Call Exec(ctx) to execute directly, or Build() for the raw SQL.
func (b *BaseDB) Insert(table string) *InsertQuery {
	return &InsertQuery{table: table, dialect: b.dialect, db: b}
}

// Update starts an UPDATE query pre-wired to this database connection.
// Call Exec(ctx) to execute directly, or Build() for the raw SQL.
func (b *BaseDB) Update(table string) *UpdateQuery {
	return &UpdateQuery{table: table, dialect: b.dialect, db: b}
}

// Delete starts a DELETE query pre-wired to this database connection.
// Call Exec(ctx) to execute directly, or Build() for the raw SQL.
func (b *BaseDB) Delete(table string) *DeleteQuery {
	return &DeleteQuery{table: table, dialect: b.dialect, db: b}
}

// NewQueryBuilder returns a standalone QueryBuilder using this connection's dialect.
// For direct execution, prefer the Select/Insert/Update/Delete factory methods instead.
func (b *BaseDB) NewQueryBuilder() *QueryBuilder {
	return NewQueryBuilder(b.dialect)
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

func (b *BaseDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	log := b.getLogger()
	log.Debug("begin_tx")
	tx, err := b.db.BeginTx(ctx, opts)
	if err != nil {
		log.Error("begin_tx error", "error", err)
		return nil, err
	}
	return &Tx{tx: tx, log: log, logArgs: b.getLogArgs()}, nil
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
