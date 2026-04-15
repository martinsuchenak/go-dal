// Package postgres provides a PostgreSQL driver wrapper for the xdal database abstraction layer.
package postgres

import (
	"database/sql"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

// Compile-time check that PostgresDB implements DBInterface.
var _ xdal.DBInterface = (*PostgresDB)(nil)

// PostgresDB wraps a *sql.DB with PostgreSQL-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type PostgresDB struct {
	*xdal.BaseDB
}

// NewPostgresDB creates a new PostgresDB. Pass nil for log to disable logging.
func NewPostgresDB(db *sql.DB, log xdal.Logger) *PostgresDB {
	return &PostgresDB{BaseDB: xdal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for PostgreSQL ("$1, $2, ..." placeholders).
func NewQueryBuilder() *xdal.QueryBuilder {
	return xdal.NewQueryBuilder(NewDialect())
}
