// Package sqlite provides a SQLite driver wrapper for the go-dal database abstraction layer.
package sqlite

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that SQLiteDB implements DBInterface.
var _ dal.DBInterface = (*SQLiteDB)(nil)

// SQLiteDB wraps a *sql.DB with SQLite-specific query building and optional logging.
type SQLiteDB struct {
	*dal.BaseDB
}

// NewSQLiteDB creates a new SQLiteDB. An optional Logger can be provided for query logging.
func NewSQLiteDB(db *sql.DB, log ...dal.Logger) *SQLiteDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &SQLiteDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (s *SQLiteDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.BaseDB.Exec(ctx, query, args...)
}

func (s *SQLiteDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.BaseDB.Query(ctx, query, args...)
}

func (s *SQLiteDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.BaseDB.QueryRow(ctx, query, args...)
}

func (s *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.BaseDB.BeginTx(ctx, opts)
}

func (s *SQLiteDB) Close() error {
	return s.BaseDB.Close()
}

// NewQueryBuilder returns a QueryBuilder configured for SQLite ("?" placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
