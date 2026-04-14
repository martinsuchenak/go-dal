package dal

import (
	"fmt"
	"strings"
)

// LimitStyle determines how LIMIT/OFFSET is rendered in SELECT queries.
type LimitStyle int

const (
	// LimitOffsetStyle uses "LIMIT X OFFSET Y" syntax (MySQL, PostgreSQL, SQLite).
	LimitOffsetStyle LimitStyle = iota
	// FetchNextStyle uses "OFFSET X ROWS FETCH NEXT Y ROWS ONLY" syntax (SQL Server).
	FetchNextStyle
)

// Dialect abstracts database-specific SQL generation. Each supported database
// provides its own Dialect implementation.
type Dialect interface {
	BuildSelect(q *SelectQuery) (string, []interface{})
	BuildInsert(q *InsertQuery) (string, []interface{})
	BuildUpdate(q *UpdateQuery) (string, []interface{})
	BuildDelete(q *DeleteQuery) (string, []interface{})
}

// BaseDialect provides a common SQL generation implementation that covers
// MySQL, PostgreSQL, SQLite, and SQL Server. Embed and override methods for
// databases with additional quirks.
type BaseDialect struct {
	PlaceholderStyle PlaceholderStyle
	LimitStyle       LimitStyle
}

func (d *BaseDialect) placeholder(idx int) string {
	switch d.PlaceholderStyle {
	case DollarNumber:
		return fmt.Sprintf("$%d", idx)
	case AtPNumber:
		return fmt.Sprintf("@p%d", idx)
	default:
		return "?"
	}
}

func (d *BaseDialect) replacePlaceholders(sql string, startIdx int) string {
	if d.PlaceholderStyle == QuestionMark {
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
			b.WriteString(d.placeholder(idx))
			idx++
		default:
			b.WriteByte(ch)
		}
	}

	return b.String()
}

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

func (d *BaseDialect) BuildSelect(q *SelectQuery) (string, []interface{}) {
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
			replaced := d.replacePlaceholders(w.condition, paramIdx)
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
			replaced := d.replacePlaceholders(w.condition, paramIdx)
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

	if d.LimitStyle == FetchNextStyle && (q.limit != nil || q.offset != nil) {
		if len(q.orderBy) == 0 {
			query += " ORDER BY (SELECT NULL)"
		}
		if q.offset != nil {
			query += fmt.Sprintf(" OFFSET %d ROWS", *q.offset)
		} else {
			query += " OFFSET 0 ROWS"
		}
		if q.limit != nil {
			query += fmt.Sprintf(" FETCH NEXT %d ROWS ONLY", *q.limit)
		}
	} else {
		if q.limit != nil {
			query += fmt.Sprintf(" LIMIT %d", *q.limit)
		}
		if q.offset != nil {
			query += fmt.Sprintf(" OFFSET %d", *q.offset)
		}
	}

	return query, args
}

func (d *BaseDialect) BuildInsert(q *InsertQuery) (string, []interface{}) {
	if len(q.keys) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(q.keys))
	for i := range placeholders {
		placeholders[i] = d.placeholder(i + 1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		q.table,
		strings.Join(q.keys, ", "),
		strings.Join(placeholders, ", "))

	return query, q.values
}

func (d *BaseDialect) BuildUpdate(q *UpdateQuery) (string, []interface{}) {
	if len(q.keys) == 0 {
		return "", nil
	}

	setClauses := make([]string, len(q.keys))
	for i, key := range q.keys {
		setClauses[i] = fmt.Sprintf("%s = %s", key, d.placeholder(i+1))
	}

	query := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(setClauses, ", "))
	args := make([]interface{}, len(q.values))
	copy(args, q.values)
	paramIdx := len(q.keys) + 1

	if len(q.wheres) > 0 {
		parts := make([]string, len(q.wheres))
		for i, w := range q.wheres {
			replaced := d.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " WHERE " + strings.Join(parts, " AND ")
	}

	return query, args
}

func (d *BaseDialect) BuildDelete(q *DeleteQuery) (string, []interface{}) {
	query := fmt.Sprintf("DELETE FROM %s", q.table)
	var args []interface{}
	paramIdx := 1

	if len(q.wheres) > 0 {
		parts := make([]string, len(q.wheres))
		for i, w := range q.wheres {
			replaced := d.replacePlaceholders(w.condition, paramIdx)
			n := countPlaceholders(w.condition)
			paramIdx += n
			args = append(args, w.args...)
			parts[i] = replaced
		}
		query += " WHERE " + strings.Join(parts, " AND ")
	}

	return query, args
}
