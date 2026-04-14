// Package postgres provides a PostgreSQL driver wrapper for the go-dal database abstraction layer.
package postgres

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that PostgresDB implements DBInterface.
var _ dal.DBInterface = (*PostgresDB)(nil)

// PostgresDB wraps a *sql.DB with PostgreSQL-specific query building and optional logging.
type PostgresDB struct {
	*dal.BaseDB
}

// NewPostgresDB creates a new PostgresDB. An optional Logger can be provided for query logging.
func NewPostgresDB(db *sql.DB, log ...dal.Logger) *PostgresDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &PostgresDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (p *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.BaseDB.Exec(ctx, query, args...)
}

func (p *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.BaseDB.Query(ctx, query, args...)
}

func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.BaseDB.QueryRow(ctx, query, args...)
}

func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.BaseDB.BeginTx(ctx, opts)
}

func (p *PostgresDB) Close() error {
	return p.BaseDB.Close()
}

// NewQueryBuilder returns a QueryBuilder configured for PostgreSQL ("$1, $2, ..." placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
