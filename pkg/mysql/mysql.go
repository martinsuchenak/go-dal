// Package mysql provides a MySQL driver wrapper for the xdal database abstraction layer.
package mysql

import (
	"database/sql"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

// Compile-time check that MySQLDB implements DBInterface.
var _ xdal.DBInterface = (*MySQLDB)(nil)

// MySQLDB wraps a *sql.DB with MySQL-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type MySQLDB struct {
	*xdal.BaseDB
}

// NewMySQLDB creates a new MySQLDB. Pass nil for log to disable logging.
func NewMySQLDB(db *sql.DB, log xdal.Logger) *MySQLDB {
	return &MySQLDB{BaseDB: xdal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for MySQL ("?" placeholders).
func NewQueryBuilder() *xdal.QueryBuilder {
	return xdal.NewQueryBuilder(NewDialect())
}
