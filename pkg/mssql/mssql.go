// Package mssql provides a SQL Server driver wrapper for the go-dal database abstraction layer.
package mssql

import (
	"database/sql"

	"github.com/martinsuchenak/go-dal/pkg/dal"
)

// Compile-time check that MSSQLDB implements DBInterface.
var _ dal.DBInterface = (*MSSQLDB)(nil)

// MSSQLDB wraps a *sql.DB with SQL Server-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type MSSQLDB struct {
	*dal.BaseDB
}

// NewMSSQLDB creates a new MSSQLDB. Pass nil for log to disable logging.
func NewMSSQLDB(db *sql.DB, log dal.Logger) *MSSQLDB {
	return &MSSQLDB{BaseDB: dal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for SQL Server ("@p1, @p2, ..." placeholders).
func NewQueryBuilder() *dal.QueryBuilder {
	return dal.NewQueryBuilder(NewDialect())
}
