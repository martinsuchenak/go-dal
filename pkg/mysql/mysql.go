// Package mysql provides a MySQL driver wrapper for the go-dal database abstraction layer.
package mysql

import (
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that MySQLDB implements DBInterface.
var _ dal.DBInterface = (*MySQLDB)(nil)

// MySQLDB wraps a *sql.DB with MySQL-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type MySQLDB struct {
	*dal.BaseDB
}

// NewMySQLDB creates a new MySQLDB. Pass nil for log to disable logging.
func NewMySQLDB(db *sql.DB, log dal.Logger) *MySQLDB {
	return &MySQLDB{BaseDB: dal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for MySQL ("?" placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
