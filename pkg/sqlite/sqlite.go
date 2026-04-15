// Package sqlite provides a SQLite driver wrapper for the xdal database abstraction layer.
package sqlite

import (
	"database/sql"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

// Compile-time check that SQLiteDB implements DBInterface.
var _ xdal.DBInterface = (*SQLiteDB)(nil)

// SQLiteDB wraps a *sql.DB with SQLite-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type SQLiteDB struct {
	*xdal.BaseDB
}

// NewSQLiteDB creates a new SQLiteDB. Pass nil for log to disable logging.
func NewSQLiteDB(db *sql.DB, log xdal.Logger) *SQLiteDB {
	return &SQLiteDB{BaseDB: xdal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for SQLite ("?" placeholders).
func NewQueryBuilder() *xdal.QueryBuilder {
	return xdal.NewQueryBuilder(NewDialect())
}
