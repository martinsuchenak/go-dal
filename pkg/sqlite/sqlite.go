// Package sqlite provides a SQLite driver wrapper for the go-dal database abstraction layer.
package sqlite

import (
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that SQLiteDB implements DBInterface.
var _ dal.DBInterface = (*SQLiteDB)(nil)

// SQLiteDB wraps a *sql.DB with SQLite-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type SQLiteDB struct {
	*dal.BaseDB
}

// NewSQLiteDB creates a new SQLiteDB. Pass nil for log to disable logging.
func NewSQLiteDB(db *sql.DB, log dal.Logger) *SQLiteDB {
	return &SQLiteDB{BaseDB: dal.NewBaseDB(db, log)}
}

// NewQueryBuilder returns a QueryBuilder configured for SQLite ("?" placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
