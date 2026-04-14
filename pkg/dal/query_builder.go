package dal

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	style PlaceholderStyle
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{style: QuestionMark}
}

func NewQueryBuilderWithStyle(style PlaceholderStyle) *QueryBuilder {
	return &QueryBuilder{style: style}
}

func (qb *QueryBuilder) Select(columns ...string) *SelectQuery {
	return &SelectQuery{columns: columns, style: qb.style}
}

func (qb *QueryBuilder) Insert(table string) *InsertQuery {
	return &InsertQuery{table: table, style: qb.style}
}

func (qb *QueryBuilder) Update(table string) *UpdateQuery {
	return &UpdateQuery{table: table, style: qb.style}
}

func (qb *QueryBuilder) Delete(table string) *DeleteQuery {
	return &DeleteQuery{table: table, style: qb.style}
}

func (s PlaceholderStyle) placeholder(idx int) string {
	switch s {
	case DollarNumber:
		return fmt.Sprintf("$%d", idx)
	case AtPNumber:
		return fmt.Sprintf("@p%d", idx)
	default:
		return "?"
	}
}

// replacePlaceholders replaces unquoted '?' characters in sql with numbered
// placeholders, starting at startIdx (1-based). It skips '?' inside single
// or double-quoted string literals and respects ”/"" as escaped quotes.
func (s PlaceholderStyle) replacePlaceholders(sql string, startIdx int) string {
	if s == QuestionMark {
		return sql
	}

	var b strings.Builder
	b.Grow(len(sql) + 16)
	idx := startIdx
	inSingle := false
	inDouble := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		switch {
		case inSingle:
			b.WriteByte(ch)
			if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					b.WriteByte(sql[i+1])
					i++
				} else {
					inSingle = false
				}
			}
		case inDouble:
			b.WriteByte(ch)
			if ch == '"' {
				if i+1 < len(sql) && sql[i+1] == '"' {
					b.WriteByte(sql[i+1])
					i++
				} else {
					inDouble = false
				}
			}
		case ch == '\'':
			b.WriteByte(ch)
			inSingle = true
		case ch == '"':
			b.WriteByte(ch)
			inDouble = true
		case ch == '?':
			b.WriteString(s.placeholder(idx))
			idx++
		default:
			b.WriteByte(ch)
		}
	}

	return b.String()
}

// countPlaceholders counts unquoted '?' characters in sql.
func countPlaceholders(sql string) int {
	count := 0
	inSingle := false
	inDouble := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch {
		case inSingle:
			if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++
				} else {
					inSingle = false
				}
			}
		case inDouble:
			if ch == '"' {
				if i+1 < len(sql) && sql[i+1] == '"' {
					i++
				} else {
					inDouble = false
				}
			}
		case ch == '\'':
			inSingle = true
		case ch == '"':
			inDouble = true
		case ch == '?':
			count++
		}
	}

	return count
}

// SelectQuery

func (q *SelectQuery) From(table string) *SelectQuery {
	q.table = table
	return q
}

func (q *SelectQuery) Join(clause string) *SelectQuery {
	q.joins = append(q.joins, clause)
	return q
}

func (q *SelectQuery) Where(condition string, args ...interface{}) *SelectQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

func (q *SelectQuery) GroupBy(columns ...string) *SelectQuery {
	q.groupBy = append(q.groupBy, columns...)
	return q
}

func (q *SelectQuery) Having(condition string, args ...interface{}) *SelectQuery {
	q.having = append(q.having, whereClause{condition: condition, args: args})
	return q
}

func (q *SelectQuery) OrderBy(columns ...string) *SelectQuery {
	q.orderBy = append(q.orderBy, columns...)
	return q
}

func (q *SelectQuery) Limit(limit int64) *SelectQuery {
	q.limit = &limit
	return q
}

func (q *SelectQuery) Offset(offset int64) *SelectQuery {
	q.offset = &offset
	return q
}

func (q *SelectQuery) Build() (string, []interface{}) {
	cols := "*"
	if len(q.columns) > 0 {
		cols = strings.Join(q.columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, q.table)
	var args []interface{}
	paramIdx := 1

	if len(q.joins) > 0 {
		query += " " + strings.Join(q.joins, " ")
	}

	if len(q.wheres) > 0 {
		parts := make([]string, len(q.wheres))
		for i, w := range q.wheres {
			replaced := q.style.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " WHERE " + strings.Join(parts, " AND ")
	}

	if len(q.groupBy) > 0 {
		query += " GROUP BY " + strings.Join(q.groupBy, ", ")
	}

	if len(q.having) > 0 {
		parts := make([]string, len(q.having))
		for i, w := range q.having {
			replaced := q.style.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " HAVING " + strings.Join(parts, " AND ")
	}

	if len(q.orderBy) > 0 {
		query += " ORDER BY " + strings.Join(q.orderBy, ", ")
	}

	if q.limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *q.limit)
	}

	if q.offset != nil {
		query += fmt.Sprintf(" OFFSET %d", *q.offset)
	}

	return query, args
}

// InsertQuery

func (q *InsertQuery) Set(key string, value interface{}) *InsertQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

func (q *InsertQuery) Build() (string, []interface{}) {
	if len(q.keys) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(q.keys))
	for i := range placeholders {
		placeholders[i] = q.style.placeholder(i + 1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(q.keys, ", "),
		strings.Join(placeholders, ", "))

	return query, q.values
}

// UpdateQuery

func (q *UpdateQuery) Set(key string, value interface{}) *UpdateQuery {
	q.keys = append(q.keys, key)
	q.values = append(q.values, value)
	return q
}

func (q *UpdateQuery) Where(condition string, args ...interface{}) *UpdateQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

func (q *UpdateQuery) Build() (string, []interface{}) {
	if len(q.keys) == 0 {
		return "", nil
	}

	setClauses := make([]string, len(q.keys))
	for i, key := range q.keys {
		setClauses[i] = fmt.Sprintf("%s = %s", key, q.style.placeholder(i+1))
	}

	query := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(setClauses, ", "))
	args := make([]interface{}, len(q.values))
	copy(args, q.values)
	paramIdx := len(q.keys) + 1

	if len(q.wheres) > 0 {
		parts := make([]string, len(q.wheres))
		for i, w := range q.wheres {
			replaced := q.style.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " WHERE " + strings.Join(parts, " AND ")
	}

	return query, args
}

// DeleteQuery

func (q *DeleteQuery) Where(condition string, args ...interface{}) *DeleteQuery {
	q.wheres = append(q.wheres, whereClause{condition: condition, args: args})
	return q
}

func (q *DeleteQuery) Build() (string, []interface{}) {
	query := fmt.Sprintf("DELETE FROM %s", q.table)
	var args []interface{}
	paramIdx := 1

	if len(q.wheres) > 0 {
		parts := make([]string, len(q.wheres))
		for i, w := range q.wheres {
			replaced := q.style.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " WHERE " + strings.Join(parts, " AND ")
	}

	return query, args
}
