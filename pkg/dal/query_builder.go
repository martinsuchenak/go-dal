package dal

// QueryBuilder creates SQL queries with the configured Dialect.
type QueryBuilder struct {
	dialect Dialect
}

// NewQueryBuilder returns a QueryBuilder using the given Dialect for SQL generation.
func NewQueryBuilder(dialect Dialect) *QueryBuilder {
	return &QueryBuilder{dialect: dialect}
}

// Select starts a SELECT query with the given columns. Use no columns for SELECT *.
func (qb *QueryBuilder) Select(columns ...string) *SelectQuery {
	return &SelectQuery{columns: columns, dialect: qb.dialect}
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

// From sets the table to select from.
func (q *SelectQuery) From(table string) *SelectQuery {
	q.table = table
	return q
}

// Join adds a JOIN clause (e.g., "INNER JOIN orders ON users.id = orders.user_id").
func (q *SelectQuery) Join(clause string) *SelectQuery {
	q.joins = append(q.joins, clause)
	return q
}

// Where adds a WHERE condition. Use "?" as placeholder for parameterized values.
// Multiple Where calls are combined with AND.
func (q *SelectQuery) Where(condition string, args ...interface{}) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

// GroupBy adds a GROUP BY clause with the specified columns.
func (q *SelectQuery) GroupBy(columns ...string) *SelectQuery {
	q.groupBy = append(q.groupBy, columns...)
	return q
}

// Having adds a HAVING condition. Use "?" as placeholder for parameterized values.
// Multiple Having calls are combined with AND.
func (q *SelectQuery) Having(condition string, args ...interface{}) *SelectQuery {
	q.having = append(q.having, whereClause{condition: condition, args: args})
	return q
}

// OrderBy adds an ORDER BY clause with the specified columns (e.g., "name ASC", "id DESC").
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
func (q *SelectQuery) Build() (string, []interface{}) {
	return q.dialect.BuildSelect(q)
}

// Set adds a column-value pair to the INSERT statement.
func (q *InsertQuery) Set(key string, value interface{}) *InsertQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

// Build constructs the INSERT SQL string and returns it along with the ordered argument slice.
// Returns an empty string if no columns have been set.
func (q *InsertQuery) Build() (string, []interface{}) {
	return q.dialect.BuildInsert(q)
}

// Set adds a column-value pair to the UPDATE statement.
func (q *UpdateQuery) Set(key string, value interface{}) *UpdateQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

// Where adds a WHERE condition. Use "?" as placeholder for parameterized values.
// Multiple Where calls are combined with AND.
func (q *UpdateQuery) Where(condition string, args ...interface{}) *UpdateQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

// Build constructs the UPDATE SQL string and returns it along with the ordered argument slice.
// Returns an empty string if no columns have been set.
func (q *UpdateQuery) Build() (string, []interface{}) {
	return q.dialect.BuildUpdate(q)
}

// Where adds a WHERE condition. Use "?" as placeholder for parameterized values.
// Multiple Where calls are combined with AND.
func (q *DeleteQuery) Where(condition string, args ...interface{}) *DeleteQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

// Build constructs the DELETE SQL string and returns it along with the ordered argument slice.
func (q *DeleteQuery) Build() (string, []interface{}) {
	return q.dialect.BuildDelete(q)
}
