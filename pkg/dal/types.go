package dal

import (
	"context"
	"database/sql"
	"errors"
)

// DBInterface defines the common operations for database interaction.
// All driver wrappers (MySQL, PostgreSQL, SQLite, MSSQL) implement this interface.
type DBInterface interface {
	// Exec executes a query without returning any rows.
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// Query executes a query that returns rows.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow executes a query that is expected to return at most one row.
	// Note: duration and errors are not logged because execution is deferred until Scan().
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	// BeginTx starts a transaction and returns a logged Tx wrapper.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error)
	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error
	// Close closes the underlying database connection.
	Close() error
	// Select starts a SELECT query pre-wired to this connection.
	Select(columns ...string) *SelectQuery
	// Insert starts an INSERT query pre-wired to this connection.
	Insert(table string) *InsertQuery
	// Update starts an UPDATE query pre-wired to this connection.
	Update(table string) *UpdateQuery
	// Delete starts a DELETE query pre-wired to this connection.
	Delete(table string) *DeleteQuery
	// NewQueryBuilder returns a standalone QueryBuilder using this connection's dialect.
	NewQueryBuilder() *QueryBuilder
	// Dialect returns the dialect used by this connection.
	Dialect() Dialect
	// WithTx executes fn within a transaction, auto-committing on success and
	// auto-rolling back on error or panic.
	WithTx(ctx context.Context, opts *sql.TxOptions, fn func(*Tx) error) error
}

var (
	// ErrEmptyTable is returned by Build when no table name is set.
	ErrEmptyTable = errors.New("dal: table name is required")
	// ErrEmptyColumns is returned by BuildInsert/BuildUpdate when no columns are set.
	ErrEmptyColumns = errors.New("dal: at least one column is required")
	// ErrEmptyInValues is returned when In() is called with no values.
	ErrEmptyInValues = errors.New("dal: In() requires at least one value")
	// ErrReturningNotSupported is returned when Returning() is used with a dialect that does not support it.
	ErrReturningNotSupported = errors.New("dal: RETURNING is not supported by this dialect")
	// ErrBatchRowLength is returned when a batch insert row has a different column count.
	ErrBatchRowLength = errors.New("dal: batch row has incorrect number of values")
	// ErrTooManyInValues is returned when In() exceeds the maximum allowed values.
	ErrTooManyInValues = errors.New("dal: In() exceeds maximum of 1000 values")
)

// MaxInValues is the upper limit for IN-clause expansion.
const MaxInValues = 1000

// InValues is a slice of values that should be expanded into individual
// placeholders when used with IN clauses. Create it with dal.In().
type InValues []interface{}

// clauseConnector determines how a whereClause is joined with its predecessor.
type clauseConnector int

const (
	andConnector clauseConnector = iota
	orConnector
	groupConnector
	groupOrConnector
)

// whereClause represents a single WHERE or HAVING condition.
type whereClause struct {
	condition string
	args      []interface{}
	connector clauseConnector
	children  []whereClause
}

// SelectQuery builds a SELECT statement using a fluent API.
type SelectQuery struct {
	table    string
	columns  []string
	distinct bool
	joins    []string
	wheres   []whereClause
	groupBy  []string
	having   []whereClause
	orderBy  []string
	limit    *int64
	offset   *int64
	dialect  Dialect
	db       DBInterface
}

// InsertQuery builds an INSERT statement using a fluent API.
// Supports both single-row (via Set) and multi-row (via Columns + Values) inserts.
type InsertQuery struct {
	table     string
	keys      []string
	values    []interface{}
	rows      [][]interface{}
	returning []string
	dialect   Dialect
	db        DBInterface
}

// UpdateQuery builds an UPDATE statement using a fluent API.
type UpdateQuery struct {
	table     string
	keys      []string
	values    []interface{}
	wheres    []whereClause
	returning []string
	dialect   Dialect
	db        DBInterface
}

// DeleteQuery builds a DELETE statement using a fluent API.
type DeleteQuery struct {
	table     string
	wheres    []whereClause
	returning []string
	dialect   Dialect
	db        DBInterface
}
