// Package postgres provides a PostgreSQL driver wrapper for the go-dal database abstraction layer.
package postgres

import (
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that PostgresDB implements DBInterface.
var _ dal.DBInterface = (*PostgresDB)(nil)

// PostgresDB wraps a *sql.DB with PostgreSQL-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type PostgresDB struct {
	*dal.BaseDB
}

// NewPostgresDB creates a new PostgresDB. Pass nil for log to disable logging.
func NewPostgresDB(db *sql.DB, log dal.Logger) *PostgresDB {
	return &PostgresDB{BaseDB: dal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for PostgreSQL ("$1, $2, ..." placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
