package dal

import (
	"context"
	"database/sql"
	"time"
)

type Logger interface {
	Trace(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)
}

type NoopLogger struct{}

func (n NoopLogger) Trace(msg string, keysAndValues ...any) {}
func (n NoopLogger) Debug(msg string, keysAndValues ...any) {}
func (n NoopLogger) Info(msg string, keysAndValues ...any)  {}
func (n NoopLogger) Warn(msg string, keysAndValues ...any)  {}
func (n NoopLogger) Error(msg string, keysAndValues ...any) {}
func (n NoopLogger) Fatal(msg string, keysAndValues ...any) {}

var noopLogger = NoopLogger{}

func NoopLoggerInstance() Logger {
	return noopLogger
}

type BaseDB struct {
	db  *sql.DB
	log Logger
}

func NewBaseDB(db *sql.DB, log Logger) *BaseDB {
	if log == nil {
		log = noopLogger
	}
	return &BaseDB{db: db, log: log}
}

func (b *BaseDB) SetLogger(log Logger) {
	if log == nil {
		log = noopLogger
	}
	b.log = log
}

func (b *BaseDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	b.log.Debug("query exec", "query", query, "args", args)
	result, err := b.db.ExecContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.log.Error("query exec error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.log.Debug("query exec done", "query", query, "duration", elapsed)
	return result, nil
}

func (b *BaseDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	b.log.Debug("query", "query", query, "args", args)
	rows, err := b.db.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)
	if err != nil {
		b.log.Error("query error", "query", query, "error", err, "duration", elapsed)
		return nil, err
	}
	b.log.Debug("query done", "query", query, "duration", elapsed)
	return rows, nil
}

func (b *BaseDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	b.log.Debug("query_row", "query", query, "args", args)
	return b.db.QueryRowContext(ctx, query, args...)
}

func (b *BaseDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	b.log.Debug("begin_tx")
	tx, err := b.db.BeginTx(ctx, opts)
	if err != nil {
		b.log.Error("begin_tx error", "error", err)
		return nil, err
	}
	return tx, nil
}

func (b *BaseDB) Close() error {
	b.log.Debug("close")
	return b.db.Close()
}

func (b *BaseDB) DB() *sql.DB {
	return b.db
}
