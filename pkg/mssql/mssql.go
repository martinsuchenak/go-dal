// Package mssql provides a SQL Server driver wrapper for the xdal database abstraction layer.
package mssql

import (
	"database/sql"

	"github.com/martinsuchenak/xdal/pkg/xdal"
)

// Compile-time check that MSSQLDB implements DBInterface.
var _ xdal.DBInterface = (*MSSQLDB)(nil)

// MSSQLDB wraps a *sql.DB with SQL Server-specific query building and optional logging.
// All DBInterface methods are promoted from the embedded BaseDB.
type MSSQLDB struct {
	*xdal.BaseDB
}

// NewMSSQLDB creates a new MSSQLDB. Pass nil for log to disable logging.
func NewMSSQLDB(db *sql.DB, log xdal.Logger) *MSSQLDB {
	return &MSSQLDB{BaseDB: xdal.NewBaseDB(db, NewDialect(), log)}
}

// NewQueryBuilder returns a QueryBuilder configured for SQL Server ("@p1, @p2, ..." placeholders).
func NewQueryBuilder() *xdal.QueryBuilder {
	return xdal.NewQueryBuilder(NewDialect())
}
