package dal

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
)

// QueryBuilder creates SQL queries with the configured Dialect.
type QueryBuilder struct {
	dialect Dialect
}

// NewQueryBuilder returns a QueryBuilder using the given Dialect for SQL generation.
func NewQueryBuilder(dialect Dialect) *QueryBuilder {
	return &QueryBuilder{dialect: dialect}
}

// Dialect returns the dialect used by this query builder.
func (qb *QueryBuilder) Dialect() Dialect {
	return qb.dialect
}

// Select starts a SELECT query with the given columns.
func (qb *QueryBuilder) Select(columns ...string) *SelectQuery {
	return &SelectQuery{columns: columns, dialect: qb.dialect}
}

// SelectAll starts a SELECT * query.
func (qb *QueryBuilder) SelectAll() *SelectQuery {
	return &SelectQuery{dialect: qb.dialect}
}

// Insert starts an INSERT query for the specified table.
func (qb *QueryBuilder) Insert(table string) *InsertQuery {
	return &InsertQuery{table: table, dialect: qb.dialect}
}

// Update starts an UPDATE query for the specified table.
func (qb *QueryBuilder) Update(table string) *UpdateQuery {
	return &UpdateQuery{table: table, dialect: qb.dialect}
}

// Delete starts a DELETE query for the specified table.
func (qb *QueryBuilder) Delete(table string) *DeleteQuery {
	return &DeleteQuery{table: table, dialect: qb.dialect}
}

// In marks values for IN-clause expansion. When passed to Where, a single "?"
// is expanded to match the number of values:
//
//	qb.Where("id IN (?)", dal.In(1, 2, 3))
//	// Generates: WHERE id IN (?, ?, ?) with args [1, 2, 3]
//
// Returns an error if no values are provided or if the count exceeds MaxInValues.
func In(values ...interface{}) (InValues, error) {
	if len(values) == 0 {
		return nil, ErrEmptyInValues
	}
	if len(values) > MaxInValues {
		return nil, ErrTooManyInValues
	}
	return InValues(values), nil
}

// --- SelectQuery ---

// From sets the table to select from.
func (q *SelectQuery) From(table string) *SelectQuery {
	q.table = table
	return q
}

// Distinct adds the DISTINCT keyword to the SELECT query.
func (q *SelectQuery) Distinct() *SelectQuery {
	q.distinct = true
	return q
}

// Join adds a JOIN clause (e.g., "INNER JOIN orders ON users.id = orders.user_id").
// The clause string is included verbatim in the generated SQL — only pass trusted,
// hardcoded strings, never user input.
func (q *SelectQuery) Join(clause string) *SelectQuery {
	q.joins = append(q.joins, clause)
	return q
}

// Where adds a WHERE condition combined with AND.
// Use "?" as placeholder for parameterized values.
func (q *SelectQuery) Where(condition string, args ...interface{}) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: andConnector})
	return q
}

// OrWhere adds a WHERE condition combined with OR.
func (q *SelectQuery) OrWhere(condition string, args ...interface{}) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: orConnector})
	return q
}

// WhereGroup adds a parenthesized group of conditions combined with AND.
// Use the callback to add conditions within the group:
//
//	qb.WhereGroup(func(g *WhereGroup) {
//	    g.Where("a = ?", 1).OrWhere("b = ?", 2)
//	})
func (q *SelectQuery) WhereGroup(fn func(*WhereGroup)) *SelectQuery {
	g := &WhereGroup{}
	fn(g)
	q.wheres = append(q.wheres, whereClause{connector: groupConnector, children: g.clauses})
	return q
}

// OrWhereGroup adds a parenthesized group of conditions combined with OR.
func (q *SelectQuery) OrWhereGroup(fn func(*WhereGroup)) *SelectQuery {
	g := &WhereGroup{}
	fn(g)
	q.wheres = append(q.wheres, whereClause{connector: groupOrConnector, children: g.clauses})
	return q
}

// WhereIsNull adds "column IS NULL" condition.
func (q *SelectQuery) WhereIsNull(column string) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: fmt.Sprintf("%s IS NULL", q.dialect.QuoteIdentifier(column)), connector: andConnector})
	return q
}

// WhereIsNotNull adds "column IS NOT NULL" condition.
func (q *SelectQuery) WhereIsNotNull(column string) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: fmt.Sprintf("%s IS NOT NULL", q.dialect.QuoteIdentifier(column)), connector: andConnector})
	return q
}

// WhereBetween adds "column BETWEEN low AND high" condition.
func (q *SelectQuery) WhereBetween(column string, low, high interface{}) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{
		condition: fmt.Sprintf("%s BETWEEN ? AND ?", q.dialect.QuoteIdentifier(column)),
		args:      []interface{}{low, high},
		connector: andConnector,
	})
	return q
}

// GroupBy adds a GROUP BY clause with the specified columns.
func (q *SelectQuery) GroupBy(columns ...string) *SelectQuery {
	q.groupBy = append(q.groupBy, columns...)
	return q
}

// Having adds a HAVING condition. Use "?" as placeholder for parameterized values.
func (q *SelectQuery) Having(condition string, args ...interface{}) *SelectQuery {
	q.having = append(q.having, whereClause{condition: condition, args: args, connector: andConnector})
	return q
}

// OrHaving adds a HAVING condition combined with OR.
func (q *SelectQuery) OrHaving(condition string, args ...interface{}) *SelectQuery {
	q.having = append(q.having, whereClause{condition: condition, args: args, connector: orConnector})
	return q
}

// OrderBy adds an ORDER BY clause with the specified columns (e.g., "name ASC", "id DESC").
// Only pass trusted, hardcoded strings — never user input.
func (q *SelectQuery) OrderBy(columns ...string) *SelectQuery {
	q.orderBy = append(q.orderBy, columns...)
	return q
}

// Limit sets the maximum number of rows to return.
func (q *SelectQuery) Limit(limit int64) *SelectQuery {
	q.limit = &limit
	return q
}

// Offset sets the number of rows to skip.
func (q *SelectQuery) Offset(offset int64) *SelectQuery {
	q.offset = &offset
	return q
}

// Build constructs the SELECT SQL string and returns it along with the ordered argument slice.
func (q *SelectQuery) Build() (string, []interface{}, error) {
	return q.dialect.BuildSelect(q)
}

// Query builds the SELECT and executes it, returning rows.
// Returns an error if the query was not created via a DB factory method.
func (q *SelectQuery) Query(ctx context.Context) (*sql.Rows, error) {
	if q.db == nil {
		return nil, fmt.Errorf("dal: query not bound to a database, use db.Select() instead of qb.Select()")
	}
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.Query(ctx, query, args...)
}

// QueryRow builds the SELECT and executes it, returning a single row.
// Returns an empty *sql.Row if not bound to a database.
func (q *SelectQuery) QueryRow(ctx context.Context) *sql.Row {
	if q.db == nil {
		return &sql.Row{}
	}
	query, args, err := q.Build()
	if err != nil {
		return &sql.Row{}
	}
	return q.db.QueryRow(ctx, query, args...)
}

// --- WhereGroup ---

// WhereGroup collects conditions for a parenthesized group.
type WhereGroup struct {
	clauses []whereClause
}

// Where adds a condition combined with AND within the group.
func (g *WhereGroup) Where(condition string, args ...interface{}) *WhereGroup {
	g.clauses = append(g.clauses, whereClause{condition: condition, args: args, connector: andConnector})
	return g
}

// OrWhere adds a condition combined with OR within the group.
func (g *WhereGroup) OrWhere(condition string, args ...interface{}) *WhereGroup {
	g.clauses = append(g.clauses, whereClause{condition: condition, args: args, connector: orConnector})
	return g
}

// --- InsertQuery ---

// Set adds a column-value pair for a single-row INSERT.
func (q *InsertQuery) Set(key string, value interface{}) *InsertQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

// SetMap adds all key-value pairs from a map for a single-row INSERT.
// Keys are sorted for deterministic SQL output.
func (q *InsertQuery) SetMap(m map[string]interface{}) *InsertQuery {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		q.keys = append(q.keys, k)
		q.values = append(q.values, m[k])
	}
	return q
}

// SetStruct adds exported struct fields as column-value pairs for a single-row INSERT.
// Field names are mapped to column names via struct tags:
//
//	type User struct {
//	    Name  string `db:"name"`
//	    Email string `db:"email"`
//	}
//
// Untagged fields use the Go field name in lowercase. Unexported fields and
// fields with `db:"-"` are skipped. Pointer fields with nil values are skipped.
func (q *InsertQuery) SetStruct(s interface{}) *InsertQuery {
	keys, vals := structFields(s)
	q.keys = append(q.keys, keys...)
	q.values = append(q.values, vals...)
	return q
}

// Columns sets the column names for a multi-row INSERT.
func (q *InsertQuery) Columns(columns ...string) *InsertQuery {
	q.keys = columns
	return q
}

// Values adds a row of values for a multi-row INSERT.
func (q *InsertQuery) Values(values ...interface{}) *InsertQuery {
	q.rows = append(q.rows, values)
	return q
}

// Returning adds a RETURNING clause (PostgreSQL, SQLite) or OUTPUT clause (MSSQL).
// Returns an error from Build() if the dialect does not support it.
func (q *InsertQuery) Returning(columns ...string) *InsertQuery {
	q.returning = columns
	return q
}

// Build constructs the INSERT SQL string and returns it along with the ordered argument slice.
// Returns an error if no columns have been set or if RETURNING is unsupported.
func (q *InsertQuery) Build() (string, []interface{}, error) {
	return q.dialect.BuildInsert(q)
}

// Exec builds the INSERT and executes it.
// Returns an error if the query was not created via a DB factory method.
func (q *InsertQuery) Exec(ctx context.Context) (sql.Result, error) {
	if q.db == nil {
		return nil, fmt.Errorf("dal: query not bound to a database, use db.Insert() instead of qb.Insert()")
	}
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.Exec(ctx, query, args...)
}

// QueryRow builds the INSERT and executes it, returning the row (for RETURNING clauses).
func (q *InsertQuery) QueryRow(ctx context.Context) *sql.Row {
	if q.db == nil {
		return &sql.Row{}
	}
	query, args, err := q.Build()
	if err != nil {
		return &sql.Row{}
	}
	return q.db.QueryRow(ctx, query, args...)
}

// --- UpdateQuery ---

// Set adds a column-value pair to the UPDATE statement.
func (q *UpdateQuery) Set(key string, value interface{}) *UpdateQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

// SetMap adds all key-value pairs from a map to the UPDATE statement.
// Keys are sorted for deterministic SQL output.
func (q *UpdateQuery) SetMap(m map[string]interface{}) *UpdateQuery {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		q.keys = append(q.keys, k)
		q.values = append(q.values, m[k])
	}
	return q
}

// SetStruct adds exported struct fields as column-value pairs to the UPDATE statement.
// See InsertQuery.SetStruct for tag conventions.
func (q *UpdateQuery) SetStruct(s interface{}) *UpdateQuery {
	keys, vals := structFields(s)
	q.keys = append(q.keys, keys...)
	q.values = append(q.values, vals...)
	return q
}

// Where adds a WHERE condition combined with AND.
func (q *UpdateQuery) Where(condition string, args ...interface{}) *UpdateQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: andConnector})
	return q
}

// OrWhere adds a WHERE condition combined with OR.
func (q *UpdateQuery) OrWhere(condition string, args ...interface{}) *UpdateQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: orConnector})
	return q
}

// Returning adds a RETURNING clause (PostgreSQL) or OUTPUT clause (MSSQL).
func (q *UpdateQuery) Returning(columns ...string) *UpdateQuery {
	q.returning = columns
	return q
}

// Build constructs the UPDATE SQL string and returns it along with the ordered argument slice.
// Returns an error if no columns have been set.
func (q *UpdateQuery) Build() (string, []interface{}, error) {
	return q.dialect.BuildUpdate(q)
}

// Exec builds the UPDATE and executes it.
// Returns an error if the query was not created via a DB factory method.
func (q *UpdateQuery) Exec(ctx context.Context) (sql.Result, error) {
	if q.db == nil {
		return nil, fmt.Errorf("dal: query not bound to a database, use db.Update() instead of qb.Update()")
	}
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.Exec(ctx, query, args...)
}

// QueryRow builds the UPDATE and executes it, returning the row (for RETURNING clauses).
func (q *UpdateQuery) QueryRow(ctx context.Context) *sql.Row {
	if q.db == nil {
		return &sql.Row{}
	}
	query, args, err := q.Build()
	if err != nil {
		return &sql.Row{}
	}
	return q.db.QueryRow(ctx, query, args...)
}

// --- DeleteQuery ---

// Where adds a WHERE condition combined with AND.
func (q *DeleteQuery) Where(condition string, args ...interface{}) *DeleteQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: andConnector})
	return q
}

// OrWhere adds a WHERE condition combined with OR.
func (q *DeleteQuery) OrWhere(condition string, args ...interface{}) *DeleteQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args, connector: orConnector})
	return q
}

// Returning adds a RETURNING clause (PostgreSQL) or OUTPUT clause (MSSQL).
func (q *DeleteQuery) Returning(columns ...string) *DeleteQuery {
	q.returning = columns
	return q
}

// Build constructs the DELETE SQL string and returns it along with the ordered argument slice.
func (q *DeleteQuery) Build() (string, []interface{}, error) {
	return q.dialect.BuildDelete(q)
}

// Exec builds the DELETE and executes it.
// Returns an error if the query was not created via a DB factory method.
func (q *DeleteQuery) Exec(ctx context.Context) (sql.Result, error) {
	if q.db == nil {
		return nil, fmt.Errorf("dal: query not bound to a database, use db.Delete() instead of qb.Delete()")
	}
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.Exec(ctx, query, args...)
}

// QueryRow builds the DELETE and executes it, returning the row (for RETURNING clauses).
func (q *DeleteQuery) QueryRow(ctx context.Context) *sql.Row {
	if q.db == nil {
		return &sql.Row{}
	}
	query, args, err := q.Build()
	if err != nil {
		return &sql.Row{}
	}
	return q.db.QueryRow(ctx, query, args...)
}

// --- Struct field reflection ---

func structFields(s interface{}) ([]string, []interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var keys []string
	var vals []interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("db")
		if tag == "-" {
			continue
		}

		colName := tag
		if colName == "" {
			colName = toSnakeCase(field.Name)
		}

		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		keys = append(keys, colName)
		vals = append(vals, fv.Interface())
	}

	return keys, vals
}

func toSnakeCase(s string) string {
	var result []byte
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(r+('a'-'A')))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}
