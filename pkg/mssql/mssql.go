// Package mssql provides a SQL Server driver wrapper for the go-dal database abstraction layer.
package mssql

import (
	"context"
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that MSSQLDB implements DBInterface.
var _ dal.DBInterface = (*MSSQLDB)(nil)

// MSSQLDB wraps a *sql.DB with SQL Server-specific query building and optional logging.
type MSSQLDB struct {
	*dal.BaseDB
}

// NewMSSQLDB creates a new MSSQLDB. An optional Logger can be provided for query logging.
func NewMSSQLDB(db *sql.DB, log ...dal.Logger) *MSSQLDB {
	var logger dal.Logger
	if len(log) > 0 {
		logger = log[0]
	}
	return &MSSQLDB{BaseDB: dal.NewBaseDB(db, logger)}
}

func (m *MSSQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.BaseDB.Exec(ctx, query, args...)
}

func (m *MSSQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.BaseDB.Query(ctx, query, args...)
}

func (m *MSSQLDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.BaseDB.QueryRow(ctx, query, args...)
}

func (m *MSSQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.BaseDB.BeginTx(ctx, opts)
}

func (m *MSSQLDB) Close() error {
	return m.BaseDB.Close()
}

// NewQueryBuilder returns a QueryBuilder configured for SQL Server ("@p1, @p2, ..." placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
