// Package dal provides a database abstraction layer with a fluent SQL query builder
// and structured logging support for MySQL, PostgreSQL, SQLite, and SQL Server.
package dal

import (
	"context"
	"database/sql"
)

// DBInterface defines the common operations for database interaction.
// All driver wrappers (MySQL, PostgreSQL, SQLite, MSSQL) implement this interface.
type DBInterface interface {
	// Exec executes a query without returning any rows.
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// Query executes a query that returns rows.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow executes a query that is expected to return at most one row.
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	// BeginTx starts a transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	// Close closes the underlying database connection.
	Close() error
}

// PlaceholderStyle represents the SQL parameter placeholder format used by a database driver.
type PlaceholderStyle int

const (
	// QuestionMark uses "?" placeholders (MySQL, SQLite).
	QuestionMark PlaceholderStyle = iota
	// DollarNumber uses "$1, $2, ..." placeholders (PostgreSQL).
	DollarNumber
	// AtPNumber uses "@p1, @p2, ..." placeholders (SQL Server).
	AtPNumber
)

// SelectQuery builds a SELECT statement using a fluent API.
type SelectQuery struct {
	table   string
	columns []string
	joins   []string
	wheres  []whereClause
	groupBy []string
	having  []whereClause
	orderBy []string
	limit   *int64
	offset  *int64
	dialect Dialect
}

// InsertQuery builds an INSERT statement using a fluent API.
type InsertQuery struct {
	table   string
	keys    []string
	values  []interface{}
	dialect Dialect
}

// UpdateQuery builds an UPDATE statement using a fluent API.
type UpdateQuery struct {
	table   string
	keys    []string
	values  []interface{}
	wheres  []whereClause
	dialect Dialect
}

// DeleteQuery builds a DELETE statement using a fluent API.
type DeleteQuery struct {
	table   string
	wheres  []whereClause
	dialect Dialect
}

type whereClause struct {
	condition string
	args      []interface{}
}
